package dtos

type ReviewerStats struct {
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	AssignedCount int    `json:"assigned_count"`
	ActivePRCount int    `json:"active_pr_count"`
	MergedPRCount int    `json:"merged_pr_count"`
}

type PRStats struct {
	PullRequestID  string `json:"pull_request_id"`
	Title          string `json:"title"`
	ReviewersCount int    `json:"reviewers_count"`
	Status         string `json:"status"`
}

type AssignmentStats struct {
	ReviewerStats    []ReviewerStats `json:"reviewer_stats"`
	PRStats          []PRStats       `json:"pr_stats"`
	TotalAssignments int             `json:"total_assignments"`
}
