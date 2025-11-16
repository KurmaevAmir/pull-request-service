package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
)

type StatsRepository struct {
	pool *pgxpool.Pool
}

func NewStatsRepository(pool *pgxpool.Pool) *StatsRepository {
	return &StatsRepository{pool: pool}
}

func (r *StatsRepository) GetReviewerStats(ctx context.Context) ([]dtos.ReviewerStats, error) {
	query := `
        SELECT 
            u.user_id,
            u.name,
            COUNT(pr.id) as assigned_count,
            COUNT(CASE WHEN pr.status = 'OPEN' THEN 1 END) as active_pr_count,
            COUNT(CASE WHEN pr.status = 'MERGED' THEN 1 END) as merged_pr_count
        FROM users u
        LEFT JOIN pr_reviews prr ON u.id = prr.reviewer_id
        LEFT JOIN pull_requests pr ON prr.pr_id = pr.id AND pr.deleted_at IS NULL
        WHERE u.deleted_at IS NULL
        GROUP BY u.id, u.user_id, u.name
        ORDER BY assigned_count DESC
    `
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []dtos.ReviewerStats
	for rows.Next() {
		var s dtos.ReviewerStats
		if err := rows.Scan(&s.UserID, &s.Username, &s.AssignedCount, &s.ActivePRCount, &s.MergedPRCount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func (r *StatsRepository) GetPRStats(ctx context.Context) ([]dtos.PRStats, error) {
	query := `
        SELECT 
            pr.pr_id,
            pr.title,
            COUNT(prr.reviewer_id) as reviewers_count,
            pr.status::text
        FROM pull_requests pr
        LEFT JOIN pr_reviews prr ON pr.id = prr.pr_id
        WHERE pr.deleted_at IS NULL
        GROUP BY pr.id, pr.pr_id, pr.title, pr.status
        ORDER BY reviewers_count DESC
    `
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []dtos.PRStats
	for rows.Next() {
		var s dtos.PRStats
		if err := rows.Scan(&s.PullRequestID, &s.Title, &s.ReviewersCount, &s.Status); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

func (r *StatsRepository) GetTotalAssignments(ctx context.Context) (int, error) {
	var total int
	query := `
        SELECT COUNT(*) 
        FROM pr_reviews prr
        JOIN pull_requests pr ON prr.pr_id = pr.id
        WHERE pr.deleted_at IS NULL
    `
	err := r.pool.QueryRow(ctx, query).Scan(&total)
	return total, err
}
