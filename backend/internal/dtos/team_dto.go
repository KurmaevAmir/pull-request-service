package dtos

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type AddTeamRequest struct {
	TeamName string          `json:"team_name" binding:"required"`
	Members  []TeamMemberDTO `json:"members" binding:"required"`
}

type TeamResponse struct {
	Team TeamDTO `json:"team"`
}

type BulkDeactivateRequest struct {
	TeamName string   `json:"team_name"`
	UserIDs  []string `json:"user_ids"`
}

type BulkDeactivateResponse struct {
	DeactivatedUsers []string              `json:"deactivated_users"`
	ReassignedPRs    []ReassignedPRSummary `json:"reassigned_prs"`
}

type ReassignedPRSummary struct {
	PullRequestID string            `json:"pull_request_id"`
	Replacements  map[string]string `json:"replacements"` // old_user_id -> new_user_id
}
