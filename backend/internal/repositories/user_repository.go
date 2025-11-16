package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/dtos"
	"github.com/KurmaevAmir/pull-request-service/backend/internal/models"
)

type UserRepository interface {
	SetIsActive(ctx context.Context, userID string, active bool) (models.User, string, error)
	GetWithTeam(ctx context.Context, userID string) (models.User, string, error)
	GetReviewPullRequests(ctx context.Context, userID string) ([]dtos.ReviewPullRequest, error)
	GetByUserID(ctx context.Context, userID string) (models.User, error)
	GetTeamMembers(ctx context.Context, teamID int64) ([]models.User, error)
	GetReviewerSlot(ctx context.Context, prInternalID int64, reviewerInternalID int64) (int, error)
	GetByInternalID(ctx context.Context, internalID int64) (models.User, error)
	BulkDeactivate(ctx context.Context, teamID int64, userIDs []string) ([]int64, error)
	GetActiveForReassignment(ctx context.Context, teamID int64, excludeUserIDs []int64) ([]models.User, error)
}

type PgUserRepository struct {
	pool *pgxpool.Pool
}

func NewPgUserRepository(pool *pgxpool.Pool) *PgUserRepository {
	return &PgUserRepository{pool: pool}
}

func (r *PgUserRepository) SetIsActive(ctx context.Context, userID string, active bool) (models.User, string, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE users u
		SET is_active = $2
		FROM teams t
		WHERE u.team_id = t.id AND u.user_id = $1
		RETURNING u.id, u.user_id, u.name, u.team_id, u.is_active, t.name
	`, userID, active)

	var u models.User
	var teamName string
	err := row.Scan(
		&u.ID,
		&u.UserID,
		&u.Name,
		&u.TeamID,
		&u.IsActive,
		&teamName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, "", ErrNotFound
		}
		return models.User{}, "", fmt.Errorf("set is_active: %w", err)
	}

	return u, teamName, nil
}

func (r *PgUserRepository) GetWithTeam(ctx context.Context, userID string) (models.User, string, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT u.id, u.user_id, u.name, u.team_id, u.is_active, t.name
		FROM users u
		JOIN teams t ON t.id = u.team_id
		WHERE u.user_id = $1
	`, userID)

	var u models.User
	var teamName string
	if err := row.Scan(&u.ID, &u.UserID, &u.Name, &u.TeamID, &u.IsActive, &teamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, "", ErrNotFound
		}
		return models.User{}, "", err
	}
	return u, teamName, nil
}

func (r *PgUserRepository) GetReviewPullRequests(ctx context.Context, userID string) ([]dtos.ReviewPullRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT pr.pr_id, pr.title, author.user_id, pr.status::text
		FROM pr_reviews rr
		JOIN pull_requests pr ON pr.id = rr.pr_id
		JOIN users reviewer ON reviewer.id = rr.reviewer_id
		LEFT JOIN users author ON author.id = pr.author_id
		WHERE reviewer.user_id = $1
		ORDER BY pr.id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []dtos.ReviewPullRequest
	for rows.Next() {
		var pr dtos.ReviewPullRequest
		if err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorUserID,
			&pr.Status,
		); err != nil {
			return nil, err
		}
		res = append(res, pr)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return res, nil
}

func (r *PgUserRepository) GetByUserID(ctx context.Context, userID string) (models.User, error) {
	const q = `
     SELECT u.id, u.user_id, u.name, u.is_active, u.team_id
	 FROM users u
     WHERE u.user_id = $1
	`

	var u models.User
	err := r.pool.QueryRow(ctx, q, userID).Scan(&u.ID, &u.UserID, &u.Name, &u.IsActive, &u.TeamID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}
	return u, nil
}

func (r *PgUserRepository) GetTeamMembers(ctx context.Context, teamID int64) ([]models.User, error) {
	const q = `
     SELECT id, user_id, name, is_active, team_id
     FROM users
     WHERE team_id = $1
    `
	rows, err := r.pool.Query(ctx, q, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memebers []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.UserID, &u.Name, &u.IsActive, &u.TeamID); err != nil {
			return nil, err
		}
		memebers = append(memebers, u)
	}
	return memebers, rows.Err()
}

func (r *PgUserRepository) GetReviewerSlot(
	ctx context.Context,
	prInternalID int64,
	reviewerInternalID int64) (int, error) {
	const q = `
	  SELECT slot
	  FROM pr_reviews
	  WHERE pr_id = $1 AND reviewer_id = $2
	 `
	var slot int
	err := r.pool.QueryRow(ctx, q, prInternalID, reviewerInternalID).Scan(&slot)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return slot, nil
}

func (r *PgUserRepository) GetByInternalID(ctx context.Context, internalID int64) (models.User, error) {
	const q = `
        SELECT id, user_id, name, is_active, team_id
        FROM users
        WHERE id = $1 AND deleted_at IS NULL
    `
	var u models.User
	err := r.pool.QueryRow(ctx, q, internalID).Scan(&u.ID, &u.UserID, &u.Name, &u.IsActive, &u.TeamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}
	return u, nil
}

func (r *PgUserRepository) BulkDeactivate(ctx context.Context, teamID int64, userIDs []string) ([]int64, error) {
	const q = `
        UPDATE users
        SET is_active = FALSE
        WHERE team_id = $1
          AND user_id = ANY($2)
          AND deleted_at IS NULL
        RETURNING id
    `
	rows, err := r.pool.Query(ctx, q, teamID, userIDs)
	if err != nil {
		return nil, fmt.Errorf("bulk deactivate: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan deactivated id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PgUserRepository) GetActiveForReassignment(
	ctx context.Context,
	teamID int64,
	excludeUserIDs []int64) ([]models.User, error) {
	const q = `
        SELECT id, user_id, name, team_id, is_active
        FROM users
        WHERE team_id = $1
          AND is_active = TRUE
          AND deleted_at IS NULL
          AND id != ALL($2)
        ORDER BY random()
    `
	rows, err := r.pool.Query(ctx, q, teamID, excludeUserIDs)
	if err != nil {
		return nil, fmt.Errorf("get active for reassignment: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.UserID, &u.Name, &u.TeamID, &u.IsActive); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
