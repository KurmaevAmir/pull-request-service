package services

import (
	"context"
	stdrr "errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/errors"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/models"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/repositories"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/validators"
)

type PRService interface {
	Create(ctx context.Context, req dtos.CreatePRRequest) (dtos.PRResponse, error)
	Merge(ctx context.Context, req dtos.MergePRRequest) (dtos.PRResponse, error)
	Reassign(ctx context.Context, req dtos.ReassignRequest) (dtos.ReassignResponse, error)
}

type prService struct {
	prRepo    repositories.PRRepository
	userRepo  repositories.UserRepository
	validator validators.PRValidator
}

func NewPRService(
	pr repositories.PRRepository,
	user repositories.UserRepository,
	val validators.PRValidator) PRService {
	return &prService{prRepo: pr, userRepo: user, validator: val}
}

func (s *prService) Create(ctx context.Context, req dtos.CreatePRRequest) (dtos.PRResponse, error) {
	if err := s.validator.ValidateCreate(ctx, req); err != nil {
		return dtos.PRResponse{}, err
	}

	author, err := s.userRepo.GetByUserID(ctx, req.Author)
	if stdrr.Is(err, repositories.ErrNotFound) {
		return dtos.PRResponse{}, errors.New(errors.CodeNotFound, "resource not found")
	}

	pr := models.PullRequest{
		PullRequestID: req.PullRequestID,
		Title:         req.Title,
		AuthorUserID:  author.ID,
		Status:        models.PROpen,
	}
	created, err := s.prRepo.Create(ctx, pr)
	fmt.Println(created, err)
	if err != nil {
		return dtos.PRResponse{}, errors.New(errors.CodeInternal, "internal error")
	}

	teamMembers, err := s.userRepo.GetTeamMembers(ctx, author.TeamID)
	if err != nil {
		return dtos.PRResponse{}, errors.New(errors.CodeInternal, "internal error")
	}

	reviewerUserIDs := s.selectReviewers(teamMembers, req.Author, 2)

	if len(reviewerUserIDs) > 0 {
		var reviewerInternalIDs []int64
		for _, uid := range reviewerUserIDs {
			u, err := s.userRepo.GetByUserID(ctx, uid)
			if err != nil {
				return dtos.PRResponse{}, errors.New(errors.CodeInternal, "internal error")
			}
			reviewerInternalIDs = append(reviewerInternalIDs, u.ID)
		}

		if err := s.prRepo.AssignReviewers(ctx, created.ID, reviewerInternalIDs); err != nil {
			return dtos.PRResponse{}, errors.New(errors.CodeInternal, "internal error")
		}
	}

	created.Reviewers = reviewerUserIDs

	return dtos.PRResponse{
		PR: mapPRToDTO(created, author.UserID),
	}, nil
}

func (s *prService) Merge(ctx context.Context, req dtos.MergePRRequest) (dtos.PRResponse, error) {
	if err := s.validator.ValidateMerge(ctx, req); err != nil {
		return dtos.PRResponse{}, err
	}

	pr, err := s.prRepo.GetByPullRequestID(ctx, req.PullRequestID)
	if stdrr.Is(err, repositories.ErrNotFound) {
		return dtos.PRResponse{}, errors.New(errors.CodeNotFound, "resource not found")
	}
	if err != nil {
		return dtos.PRResponse{}, errors.New(errors.CodeInternal, "internal error")
	}

	author, err := s.getUserByInternalID(ctx, pr.AuthorUserID)
	if err != nil {
		return dtos.PRResponse{}, errors.New(errors.CodeInternal, "internal error")
	}

	if pr.Status == models.PRMerged {
		return dtos.PRResponse{PR: mapPRToDTO(pr, author.UserID)}, nil
	}

	now := time.Now()
	if err := s.prRepo.UpdateStatus(ctx, req.PullRequestID, models.PRMerged, &now); err != nil {
		return dtos.PRResponse{}, errors.New(errors.CodeInternal, "internal error")
	}

	pr.Status = models.PRMerged
	pr.UpdatedAt = &now

	return dtos.PRResponse{PR: mapPRToDTO(pr, author.UserID)}, nil
}

func (s *prService) Reassign(ctx context.Context, req dtos.ReassignRequest) (dtos.ReassignResponse, error) {
	if err := s.validator.ValidateReassign(ctx, req); err != nil {
		return dtos.ReassignResponse{}, err
	}

	pr, err := s.prRepo.GetByPullRequestID(ctx, req.PullRequestID)
	if stdrr.Is(err, repositories.ErrNotFound) {
		return dtos.ReassignResponse{}, errors.New(errors.CodeNotFound, "resource not found")
	}
	if err != nil {
		return dtos.ReassignResponse{}, errors.New(errors.CodeInternal, "internal error")
	}

	if pr.Status == models.PRMerged || !contains(pr.Reviewers, req.OldUserID) {
		return dtos.ReassignResponse{}, errors.New(errors.CodePRMerged, "cannot reassign on merged PR")
	}

	oldUser, err := s.userRepo.GetByUserID(ctx, req.OldUserID)
	if stdrr.Is(err, repositories.ErrNotFound) {
		return dtos.ReassignResponse{}, errors.New(errors.CodeNotFound, "resource not found")
	}
	if err != nil {
		return dtos.ReassignResponse{}, errors.New(errors.CodeInternal, "internal error")
	}

	teamMembers, err := s.userRepo.GetTeamMembers(ctx, oldUser.TeamID)
	if err != nil {
		return dtos.ReassignResponse{}, errors.New(errors.CodeInternal, "internal error")
	}

	author, err := s.getUserByInternalID(ctx, pr.AuthorUserID)
	if err != nil {
		return dtos.ReassignResponse{}, errors.New(errors.CodeInternal, "internal error")
	}

	exclude := append(pr.Reviewers, author.UserID)
	candidates := s.selectReviewers(teamMembers, exclude, 1)

	if len(candidates) == 0 {
		return dtos.ReassignResponse{}, errors.New(errors.CodeNoCandidate, "cannot reassign on merged PR")
	}

	newReviewerUserID := candidates[0]
	newReviewer, err := s.userRepo.GetByUserID(ctx, newReviewerUserID)
	if err != nil {
		return dtos.ReassignResponse{}, errors.New(errors.CodeInternal, "internal error")
	}

	slot, err := s.userRepo.GetReviewerSlot(ctx, pr.ID, oldUser.ID)
	if err != nil {
		return dtos.ReassignResponse{}, errors.New(errors.CodeInternal, "internal error")
	}

	if err := s.prRepo.RemoveReviewer(ctx, pr.ID, oldUser.ID); err != nil {
		return dtos.ReassignResponse{}, errors.New(errors.CodeInternal, "internal error")
	}
	if err := s.prRepo.AddReviewer(ctx, pr.ID, newReviewer.ID, slot); err != nil {
		return dtos.ReassignResponse{}, errors.New(errors.CodeInternal, "internal error")
	}

	for i, r := range pr.Reviewers {
		if r == req.OldUserID {
			pr.Reviewers[i] = newReviewerUserID
			break
		}
	}

	return dtos.ReassignResponse{
		PR:         mapPRToDTO(pr, author.UserID),
		ReplacedBy: newReviewerUserID,
	}, nil
}

func (s *prService) getUserByInternalID(ctx context.Context, internalID int64) (models.User, error) {
	return s.userRepo.GetByInternalID(ctx, internalID)
}

func (s *prService) selectReviewers(members []models.User, exclude interface{}, limit int) []string {
	var excludeSet map[string]bool
	switch v := exclude.(type) {
	case string:
		excludeSet = map[string]bool{v: true}
	case []string:
		excludeSet = make(map[string]bool, len(v))
		for _, e := range v {
			excludeSet[e] = true
		}
	}

	var candidates []string
	for _, m := range members {
		if m.IsActive && !excludeSet[m.UserID] {
			candidates = append(candidates, m.UserID)
		}
	}

	if len(candidates) == 0 {
		return nil
	}
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	return candidates
}

func mapPRToDTO(pr models.PullRequest, authorUserID string) dtos.PullRequestDTO {
	return dtos.PullRequestDTO{
		PullRequestID:     pr.PullRequestID,
		Title:             pr.Title,
		AuthorUserID:      authorUserID,
		Status:            pr.Status,
		AssignedReviewers: pr.Reviewers,
		CreatedAt:         pr.CreatedAt,
	}
}

func contains(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
