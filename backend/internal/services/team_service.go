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
}

type teamService struct {
	repo      repositories.TeamRepository
	validator validators.TeamValidator
}

func NewTeamService(repo repositories.TeamRepository, validator validators.TeamValidator) TeamService {
	return &teamService{repo: repo, validator: validator}
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
