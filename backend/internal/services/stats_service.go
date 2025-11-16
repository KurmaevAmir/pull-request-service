package services

import (
	"context"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/repositories"
)

type StatsService struct {
	repo *repositories.StatsRepository
}

func NewStatsService(repo *repositories.StatsRepository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) GetAssignmentStats(ctx context.Context) (*dtos.AssignmentStats, error) {
	reviewerStats, err := s.repo.GetReviewerStats(ctx)
	if err != nil {
		return nil, err
	}

	prStats, err := s.repo.GetPRStats(ctx)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.GetTotalAssignments(ctx)
	if err != nil {
		return nil, err
	}

	return &dtos.AssignmentStats{
		ReviewerStats:    reviewerStats,
		PRStats:          prStats,
		TotalAssignments: total,
	}, nil
}
