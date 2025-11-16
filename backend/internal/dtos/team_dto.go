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
