package dtos

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type SetIsActiveResponse struct {
	User User `json:"user"`
}

type GetReviewResponse struct {
	UserID       string              `json:"user_id"`
	PullRequests []ReviewPullRequest `json:"pull_requests"`
}

type ReviewPullRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorUserID    string `json:"author_id"`
	Status          string `json:"status"`
}
