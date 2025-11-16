package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/models"
)

type TeamRepository interface {
	CreateTeamWithMembers(ctx context.Context, teamName string, members []models.User) error
	GetTeamWithMembers(ctx context.Context, teamName string) (models.Team, []models.User, error)
	ExistsTeamWithMembers(ctx context.Context, teamName string, members []models.User) (bool, error)
}

type PgTeamRepository struct {
	pool *pgxpool.Pool
}

func NewPgTeamRepository(pool *pgxpool.Pool) *PgTeamRepository {
	return &PgTeamRepository{pool: pool}
}

func (r *PgTeamRepository) CreateTeamWithMembers(ctx context.Context, teamName string, members []models.User) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var teamID int64
	if err = tx.QueryRow(ctx, `INSERT INTO teams(name) VALUES ($1) RETURNING id`, teamName).Scan(&teamID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrTeamExists
		}
		return err
	}

	batch := &pgx.Batch{}
	for _, m := range members {
		batch.Queue(`
            INSERT INTO users(user_id, name, team_id, is_active)
            VALUES ($1, $2, $3, $4)
        `, m.UserID, m.Name, teamID, m.IsActive)
	}

	br := tx.SendBatch(ctx, batch)
	for i := 0; i < batch.Len(); i++ {
		if _, err = br.Exec(); err != nil {
			_ = br.Close()
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrUserExists
			}
			return err
		}
	}
	if err = br.Close(); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *PgTeamRepository) GetTeamWithMembers(ctx context.Context,
	teamName string) (models.Team, []models.User, error) {
	var t models.Team
	if err := r.pool.QueryRow(ctx, `SELECT id, name FROM teams WHERE name=$1`, teamName).
		Scan(&t.ID, &t.Name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Team{}, nil, ErrNotFound
		}
		return models.Team{}, nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT user_id, name, team_id, is_active
		FROM users
		WHERE team_id = $1
		ORDER BY user_id
	`, t.ID)
	if err != nil {
		return models.Team{}, nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err = rows.Scan(&u.UserID, &u.Name, &u.TeamID, &u.IsActive); err != nil {
			return models.Team{}, nil, err
		}
		users = append(users, u)
	}
	if rows.Err() != nil {
		return models.Team{}, nil, rows.Err()
	}

	return t, users, nil
}

func (r *PgTeamRepository) ExistsTeamWithMembers(ctx context.Context,
	teamName string, members []models.User) (bool, error) {
	var dummy int
	if err := r.pool.QueryRow(ctx,
		`SELECT 1 FROM teams WHERE name = $1`,
		teamName,
	).Scan(&dummy); err != nil && err != pgx.ErrNoRows {
		return false, err
	} else if err == nil {
		return true, nil
	}

	userIDs := make([]string, 0, len(members))
	for _, m := range members {
		userIDs = append(userIDs, m.UserID)
	}

	if err := r.pool.QueryRow(ctx,
		`SELECT 1 FROM users WHERE user_id = ANY($1)`,
		userIDs,
	).Scan(&dummy); err != nil && err != pgx.ErrNoRows {
		return false, err
	} else if err == nil {
		return true, nil
	}

	return false, nil
}
