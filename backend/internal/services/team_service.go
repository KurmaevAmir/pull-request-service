package services

import (
	"context"
	"errors"
	"strings"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
	derr "github.com/KurmaevAmir/pull-request-service/backend/internal/errors"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/models"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/repositories"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/validators"
)

type TeamService interface {
	AddTeam(ctx context.Context, in dtos.AddTeamRequest) (dtos.TeamResponse, error)
	GetTeam(ctx context.Context, name string) (dtos.TeamDTO, error)
	BulkDeactivate(ctx context.Context, req dtos.BulkDeactivateRequest) (*dtos.BulkDeactivateResponse, error)
}

type teamService struct {
	repo      repositories.TeamRepository
	userRepo  repositories.UserRepository
	prRepo    repositories.PRRepository
	validator validators.TeamValidator
}

func NewTeamService(
	repo repositories.TeamRepository,
	userRepo repositories.UserRepository,
	prRepo repositories.PRRepository,
	validator validators.TeamValidator) TeamService {
	return &teamService{
		repo:      repo,
		userRepo:  userRepo,
		prRepo:    prRepo,
		validator: validator,
	}
}

func (s *teamService) AddTeam(ctx context.Context, in dtos.AddTeamRequest) (dtos.TeamResponse, error) {
	if err := s.validator.ValidateAddTeam(ctx, in); err != nil {
		return dtos.TeamResponse{}, err
	}
	members := make([]models.User, 0, len(in.Members))
	for _, m := range in.Members {
		members = append(members, models.User{
			Name:     m.Username,
			IsActive: m.IsActive,
			UserID:   m.UserID,
		})
	}
	teamName := strings.TrimSpace(in.TeamName)
	if err := s.repo.CreateTeamWithMembers(ctx, teamName, members); err != nil {
		if errors.Is(err, repositories.ErrTeamExists) {
			return dtos.TeamResponse{}, derr.New(derr.CodeTeamExists, "team already exists")
		}
		return dtos.TeamResponse{}, err
	}
	return dtos.TeamResponse{Team: dtos.TeamDTO{TeamName: teamName, Members: in.Members}}, nil
}

func (s *teamService) GetTeam(ctx context.Context, name string) (dtos.TeamDTO, error) {
	teamName := strings.TrimSpace(name)

	t, users, err := s.repo.GetTeamWithMembers(ctx, teamName)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return dtos.TeamDTO{}, derr.New(derr.CodeNotFound, "resource not found")
		}
		return dtos.TeamDTO{}, err
	}

	outMembers := make([]dtos.TeamMemberDTO, 0, len(users))
	for _, u := range users {
		outMembers = append(outMembers, dtos.TeamMemberDTO{
			UserID:   u.UserID,
			Username: u.Name,
			IsActive: u.IsActive,
		})
	}

	return dtos.TeamDTO{
		TeamName: t.Name,
		Members:  outMembers,
	}, nil
}

func (s *teamService) BulkDeactivate(
	ctx context.Context,
	req dtos.BulkDeactivateRequest) (*dtos.BulkDeactivateResponse, error) {
	team, _, err := s.repo.GetTeamWithMembers(ctx, strings.TrimSpace(req.TeamName))
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, derr.New(derr.CodeNotFound, "team not found")
		}
		return nil, err
	}

	deactivatedIDs, err := s.userRepo.BulkDeactivate(ctx, team.ID, req.UserIDs)
	if err != nil {
		return nil, err
	}
	if len(deactivatedIDs) == 0 {
		return &dtos.BulkDeactivateResponse{
			DeactivatedUsers: req.UserIDs,
			ReassignedPRs:    []dtos.ReassignedPRSummary{},
		}, nil
	}

	prsWithReviewers, err := s.prRepo.GetOpenPRsWithReviewers(ctx, deactivatedIDs)
	if err != nil {
		return nil, err
	}

	activeCandidates, err := s.userRepo.GetActiveForReassignment(ctx, team.ID, deactivatedIDs)
	if err != nil {
		return nil, err
	}

	reassigned := s.reassignReviewers(ctx, prsWithReviewers, activeCandidates, deactivatedIDs)

	return &dtos.BulkDeactivateResponse{
		DeactivatedUsers: req.UserIDs,
		ReassignedPRs:    reassigned,
	}, nil
}

func (s *teamService) reassignReviewers(
	ctx context.Context,
	prsWithReviewers []dtos.PRWithReviewers,
	activeCandidates []models.User,
	deactivatedInternalIDs []int64,
) []dtos.ReassignedPRSummary {
	if len(activeCandidates) == 0 || len(deactivatedInternalIDs) == 0 {
		return []dtos.ReassignedPRSummary{}
	}

	deactivatedSet := make(map[int64]struct{}, len(deactivatedInternalIDs))
	for _, id := range deactivatedInternalIDs {
		deactivatedSet[id] = struct{}{}
	}

	reassigned := make([]dtos.ReassignedPRSummary, 0)
	candidateIdx := 0

	for _, prwr := range prsWithReviewers {
		replacements := make(map[string]string)

		for _, reviewer := range prwr.Reviewers {
			if _, needReplace := deactivatedSet[reviewer.ID]; !needReplace {
				continue
			}
			if len(activeCandidates) == 0 {
				break
			}

			slot, err := s.prRepo.GetReviewerSlot(ctx, prwr.PR.ID, reviewer.ID)
			if err != nil {
				continue
			}

			newReviewer := activeCandidates[candidateIdx]
			candidateIdx = (candidateIdx + 1) % len(activeCandidates)

			if err := s.prRepo.ReplaceReviewer(ctx, prwr.PR.ID, reviewer.ID, newReviewer.ID, slot); err != nil {
				continue
			}

			replacements[reviewer.UserID] = newReviewer.UserID
		}

		if len(replacements) > 0 {
			reassigned = append(reassigned, dtos.ReassignedPRSummary{
				PullRequestID: prwr.PR.PullRequestID,
				Replacements:  replacements,
			})
		}
	}

	return reassigned
}
