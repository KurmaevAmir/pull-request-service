package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/models"
)

type PRRepository interface {
	Create(ctx context.Context, pr models.PullRequest) (models.PullRequest, error)
	GetByPullRequestID(ctx context.Context, prID string) (models.PullRequest, error)
	UpdateStatus(ctx context.Context, prID string, status string, mergedAt *time.Time) error
	AssignReviewers(ctx context.Context, prInternalID int64, reviewerInternalIDs []int64) error
	GetReviewers(ctx context.Context, prID int64) ([]string, error)
	RemoveReviewer(ctx context.Context, prID int64, userID int64) error
	AddReviewer(ctx context.Context, prID int64, userID int64, slot int) error
	ExistsByPullRequestID(ctx context.Context, prID string) (bool, error)
	GetOpenPRsWithReviewers(ctx context.Context, deactivatedInternalIDs []int64) ([]dtos.PRWithReviewers, error)
	ReplaceReviewer(ctx context.Context, prID int64, oldReviewerID, newReviewerID int64, slot int) error
	GetReviewerSlot(ctx context.Context, prID, reviewerID int64) (int, error)
}

type pgPRRepository struct {
	pool *pgxpool.Pool
}

func NewPRRepository(pool *pgxpool.Pool) PRRepository {
	return &pgPRRepository{pool: pool}
}

func (r *pgPRRepository) Create(ctx context.Context, pr models.PullRequest) (models.PullRequest, error) {
	const q = `
		INSERT INTO pull_requests (pr_id, title, author_id, status, created_at)
		VALUES ($1, $2, $3, $4::pr_status, $5)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, q, pr.PullRequestID, pr.Title, pr.AuthorUserID, pr.Status, time.Now()).
		Scan(&pr.ID, &pr.CreatedAt, &pr.UpdatedAt)
	if err != nil {
		return models.PullRequest{}, fmt.Errorf("create pr: %w", err)
	}
	return pr, nil
}

func (r *pgPRRepository) GetByPullRequestID(ctx context.Context, prID string) (models.PullRequest, error) {
	const q = `
		SELECT id, pr_id, title, author_id, status::text, created_at, updated_at
		FROM pull_requests
		WHERE pr_id = $1 AND deleted_at IS NULL
	`

	var pr models.PullRequest
	err := r.pool.QueryRow(ctx, q, prID).Scan(
		&pr.ID, &pr.PullRequestID, &pr.Title, &pr.AuthorUserID, &pr.Status, &pr.CreatedAt, &pr.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.PullRequest{}, ErrNotFound
		}
		return models.PullRequest{}, fmt.Errorf("get pr: %w", err)
	}

	reviewers, err := r.GetReviewers(ctx, pr.ID)
	if err != nil {
		return models.PullRequest{}, err
	}
	pr.Reviewers = reviewers

	return pr, nil
}

func (r *pgPRRepository) UpdateStatus(ctx context.Context, prID string, status string, updatedAt *time.Time) error {
	const q = `
		UPDATE pull_requests
		SET status = $2::pr_status, updated_at = $3
		WHERE pr_id = $1 AND deleted_at IS NULL
	`
	res, err := r.pool.Exec(ctx, q, prID, status, updatedAt)
	if err != nil {
		return fmt.Errorf("update pr status: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgPRRepository) AssignReviewers(ctx context.Context, prInternalID int64, reviewerInternalIDs []int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	const qInsert = `
	  INSERT INTO pr_reviews (pr_id, reviewer_id, slot)
	  VALUES ($1, $2, $3)
	 `
	for i, rid := range reviewerInternalIDs {
		slot := i + 1
		_, err := tx.Exec(ctx, qInsert, prInternalID, rid, slot)
		if err != nil {
			return fmt.Errorf("assign reviewer: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *pgPRRepository) GetReviewers(ctx context.Context, prID int64) ([]string, error) {
	const q = `
		SELECT u.user_id
		FROM pr_reviews pr
		JOIN users u ON u.id = pr.reviewer_id
		WHERE pr.pr_id = $1
		ORDER BY pr.slot
	`
	rows, err := r.pool.Query(ctx, q, prID)
	if err != nil {
		return nil, fmt.Errorf("get reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scan reviewer: %w", err)
		}
		reviewers = append(reviewers, uid)
	}
	return reviewers, rows.Err()
}

func (r *pgPRRepository) RemoveReviewer(ctx context.Context, prID int64, userID int64) error {
	const q = `
		DELETE FROM pr_reviews
		WHERE pr_id = $1 AND reviewer_id = $2
	`
	res, err := r.pool.Exec(ctx, q, prID, userID)
	if err != nil {
		return fmt.Errorf("remove reviewer: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgPRRepository) AddReviewer(ctx context.Context, prID int64, userID int64, slot int) error {
	const q = `
		INSERT INTO pr_reviews (pr_id, reviewer_id, slot)
		VALUES ($1, $2, $3)
	`
	_, err := r.pool.Exec(ctx, q, prID, userID, slot)
	if err != nil {
		return fmt.Errorf("add reviewer: %w", err)
	}
	return nil
}

func (r *pgPRRepository) ExistsByPullRequestID(ctx context.Context, prID string) (bool, error) {
	const q = `
        SELECT EXISTS(
            SELECT 1 
            FROM pull_requests 
            WHERE pr_id = $1 AND deleted_at IS NULL
        )
    `
	var exists bool
	err := r.pool.QueryRow(ctx, q, prID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check pr exists: %w", err)
	}
	return exists, nil
}

func (r *pgPRRepository) GetOpenPRsWithReviewers(ctx context.Context,
	deactivatedInternalIDs []int64) ([]dtos.PRWithReviewers, error) {
	const q = `
        SELECT DISTINCT
            pr.id, pr.pr_id, pr.title, pr.author_id, pr.status::text,
            prr.reviewer_id, prr.slot,
            u.user_id, u.name, u.team_id, u.is_active
        FROM pull_requests pr
        JOIN pr_reviews prr ON pr.id = prr.pr_id
        JOIN users u ON prr.reviewer_id = u.id
        WHERE pr.status = 'OPEN'
          AND pr.deleted_at IS NULL
          AND prr.reviewer_id = ANY($1)
        ORDER BY pr.id, prr.slot
    `
	rows, err := r.pool.Query(ctx, q, deactivatedInternalIDs)
	if err != nil {
		return nil, fmt.Errorf("get open PRs with reviewers: %w", err)
	}
	defer rows.Close()

	prMap := make(map[int64]*dtos.PRWithReviewers)
	for rows.Next() {
		var prID, authorID, reviewerID, teamID int64
		var prExtID, title, status, userID, name string
		var slot int
		var isActive bool

		if err := rows.Scan(&prID, &prExtID, &title, &authorID,
			&status, &reviewerID, &slot, &userID, &name, &teamID,
			&isActive); err != nil {
			return nil, fmt.Errorf("scan pr with reviewer: %w", err)
		}

		if _, exists := prMap[prID]; !exists {
			prMap[prID] = &dtos.PRWithReviewers{
				PR: models.PullRequest{
					ID:            prID,
					PullRequestID: prExtID,
					Title:         title,
					AuthorUserID:  authorID,
					Status:        status,
				},
				Reviewers: []models.User{},
			}
		}

		prMap[prID].Reviewers = append(prMap[prID].Reviewers, models.User{
			ID:       reviewerID,
			UserID:   userID,
			Name:     name,
			TeamID:   teamID,
			IsActive: isActive,
		})
	}

	result := make([]dtos.PRWithReviewers, 0, len(prMap))
	for _, prwr := range prMap {
		result = append(result, *prwr)
	}
	return result, rows.Err()
}

func (r *pgPRRepository) ReplaceReviewer(
	ctx context.Context,
	prID int64,
	oldReviewerID,
	newReviewerID int64,
	slot int) error {
	const q = `
        UPDATE pr_reviews
        SET reviewer_id = $1, assigned_at = NOW()
        WHERE pr_id = $2 AND reviewer_id = $3 AND slot = $4
    `
	res, err := r.pool.Exec(ctx, q, newReviewerID, prID, oldReviewerID, slot)
	if err != nil {
		return fmt.Errorf("replace reviewer: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *pgPRRepository) GetReviewerSlot(ctx context.Context, prID, reviewerID int64) (int, error) {
	const q = `SELECT slot FROM pr_reviews WHERE pr_id = $1 AND reviewer_id = $2`
	var slot int
	err := r.pool.QueryRow(ctx, q, prID, reviewerID).Scan(&slot)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, fmt.Errorf("get reviewer slot: %w", err)
	}
	return slot, nil
}
