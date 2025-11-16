package validators

import (
	"context"
	"strconv"
	"strings"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/errors"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/models"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/repositories"
)

type TeamValidator interface {
	ValidateAddTeam(ctx context.Context, in dtos.AddTeamRequest) error
}

type DefaultTeamValidator struct {
	MaxMembers int
	repo       repositories.TeamRepository
}

func NewTeamValidator(repo repositories.TeamRepository) *DefaultTeamValidator {
	return &DefaultTeamValidator{MaxMembers: 200, repo: repo}
}

func (v *DefaultTeamValidator) ValidateAddTeam(ctx context.Context, in dtos.AddTeamRequest) error {
	name := strings.TrimSpace(in.TeamName)
	members := in.Members
	if name == "" {
		return errors.New(errors.CodeValidation, "team_name empty")
	}
	if len(members) == 0 {
		return errors.New(errors.CodeValidation, "members empty")
	}
	if len(members) > v.MaxMembers {
		return errors.New(errors.CodeValidation, "too many members")
	}

	seen := make(map[string]struct{}, len(in.Members))
	for i, m := range in.Members {
		if strings.TrimSpace(m.UserID) == "" {
			return errors.New(errors.CodeValidation, "members["+strconv.Itoa(i)+"].user_id empty")
		}
		if strings.TrimSpace(m.Username) == "" {
			return errors.New(errors.CodeValidation, "members["+strconv.Itoa(i)+"].username empty")
		}
		if _, ok := seen[m.UserID]; ok {
			return errors.New(errors.CodeValidation, "duplicate user_id "+m.UserID)
		}
		seen[m.UserID] = struct{}{}
	}

	modelMembers := make([]models.User, 0, len(in.Members))
	for _, m := range in.Members {
		modelMembers = append(modelMembers, models.User{
			UserID:   m.UserID,
			Name:     m.Username,
			IsActive: m.IsActive,
		})
	}
	exists, err := v.repo.ExistsTeamWithMembers(ctx, strings.TrimSpace(in.TeamName), modelMembers)
	if err != nil {
		return errors.New(errors.CodeValidation, "invalid data")
	}
	if exists {
		return errors.New(errors.CodeValidation, "team_name or one of user_id already exists")
	}

	return nil
}
