package validators

import (
	"context"
	"strings"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/errors"
)

type UserValidator interface {
	ValidateSetIsActive(ctx context.Context, in dtos.SetIsActiveRequest) error
	ValidateUserID(ctx context.Context, userID string) error
}

type userValidator struct{}

func NewUserValidator() UserValidator {
	return &userValidator{}
}

func (v *userValidator) ValidateSetIsActive(_ context.Context, in dtos.SetIsActiveRequest) error {
	if strings.TrimSpace(in.UserID) == "" {
		return errors.New(errors.CodeValidation, "user_id required")
	}
	return nil
}

func (v *userValidator) ValidateUserID(_ context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return errors.New(errors.CodeValidation, "user_id required")
	}
	return nil
}
