package dtos

import (
	"time"

	"github.com/KurmaevAmir/pull-request-service/backend/internal/models"
)

type CreatePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	Title         string `json:"pull_request_name" binding:"required"`
	Author        string `json:"author_id" binding:"required"`
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

type ReassignRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldUserID     string `json:"old_reviewer_id" binding:"required"`
}

type PRResponse struct {
	PR PullRequestDTO `json:"pr"`
}

type ReassignResponse struct {
	PR         PullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

type PullRequestDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	Title             string     `json:"pull_request_name"`
	AuthorUserID      string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         time.Time  `json:"createdAt,omitempty"`
	UpdatedAt         *time.Time `json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	PullRequestID string `json:"pull_request_id"`
	Title         string `json:"pull_request_name"`
	AuthorUserID  string `json:"author_id"`
	Status        string `json:"status"`
}

type PRWithReviewers struct {
	PR        models.PullRequest
	Reviewers []models.User
}
