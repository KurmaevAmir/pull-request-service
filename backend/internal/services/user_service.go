package services

import (
	"context"
	stderrs "errors"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/errors"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/repositories"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/validators"
)

type UserService interface {
	SetIsActive(ctx context.Context, in dtos.SetIsActiveRequest) (dtos.SetIsActiveResponse, error)
	GetReview(ctx context.Context, userID string) (dtos.GetReviewResponse, error)
}

type userService struct {
	users     repositories.UserRepository
	validator validators.UserValidator
}

func NewUserService(users repositories.UserRepository, validator validators.UserValidator) UserService {
	return &userService{
		users:     users,
		validator: validator,
	}
}

func (s *userService) SetIsActive(ctx context.Context, in dtos.SetIsActiveRequest) (dtos.SetIsActiveResponse, error) {
	if err := s.validator.ValidateSetIsActive(ctx, in); err != nil {
		return dtos.SetIsActiveResponse{}, err
	}

	updated, teamName, err := s.users.SetIsActive(ctx, in.UserID, in.IsActive)
	if err != nil {
		if stderrs.Is(err, repositories.ErrNotFound) {
			return dtos.SetIsActiveResponse{}, errors.New(errors.CodeNotFound, "resource not found")
		}
		return dtos.SetIsActiveResponse{}, errors.New(errors.CodeValidation, "invalid request")
	}

	return dtos.SetIsActiveResponse{
		User: dtos.User{
			UserID:   updated.UserID,
			Username: updated.Name,
			TeamName: teamName,
			IsActive: updated.IsActive,
		},
	}, nil
}

func (s *userService) GetReview(ctx context.Context, userID string) (dtos.GetReviewResponse, error) {
	if err := s.validator.ValidateUserID(ctx, userID); err != nil {
		return dtos.GetReviewResponse{}, err
	}

	prs, err := s.users.GetReviewPullRequests(ctx, userID)
	if err != nil {
		return dtos.GetReviewResponse{}, errors.New(errors.CodeValidation, "invalid request")
	}

	return dtos.GetReviewResponse{
		UserID:       userID,
		PullRequests: prs,
	}, nil
}
