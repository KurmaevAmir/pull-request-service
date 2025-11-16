package validators

import (
	"context"
	stderrs "errors"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/errors"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/repositories"
)

type PRValidator interface {
	ValidateCreate(ctx context.Context, req dtos.CreatePRRequest) error
	ValidateMerge(ctx context.Context, req dtos.MergePRRequest) error
	ValidateReassign(ctx context.Context, req dtos.ReassignRequest) error
}

type prValidator struct {
	prRepo   repositories.PRRepository
	userRepo repositories.UserRepository
}

func NewPRValidator(pr repositories.PRRepository, user repositories.UserRepository) PRValidator {
	return &prValidator{prRepo: pr, userRepo: user}
}

func (v *prValidator) ValidateCreate(ctx context.Context, req dtos.CreatePRRequest) error {
	if req.PullRequestID == "" || req.Title == "" || req.Author == "" {
		return errors.New(errors.CodeValidation, "invalid request")
	}

	exists, err := v.prRepo.ExistsByPullRequestID(ctx, req.PullRequestID)
	if err != nil {
		return errors.New(errors.CodeInternal, "internal error")
	}
	if exists {
		return errors.New(errors.CodePRExists, "PR id already exists")
	}

	_, err = v.userRepo.GetByUserID(ctx, req.Author)
	if stderrs.Is(err, repositories.ErrNotFound) {
		return errors.New(errors.CodeNotFound, "resource not found")
	}
	if err != nil {
		return errors.New(errors.CodePRExists, "internal error")
	}

	return nil
}

func (v *prValidator) ValidateMerge(ctx context.Context, req dtos.MergePRRequest) error {
	if req.PullRequestID == "" {
		return errors.New(errors.CodeValidation, "invalid request")
	}
	return nil
}

func (v *prValidator) ValidateReassign(ctx context.Context, req dtos.ReassignRequest) error {
	if req.PullRequestID == "" || req.OldUserID == "" {
		return errors.New(errors.CodeValidation, "invalid request")
	}
	return nil
}
