package models

import "time"

const (
	PROpen   = "OPEN"
	PRMerged = "MERGED"
)

type PullRequest struct {
	ID            int64      `db:"id"`
	PullRequestID string     `db:"pr_id"` // Внешний идентификатор pull request
	Title         string     `db:"title"`
	AuthorUserID  int64      `db:"author_id"`
	Status        string     `db:"status"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
	DeletedAt     time.Time  `db:"deleted_at"`
	Reviewers     []string
}
