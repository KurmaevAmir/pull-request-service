package repositories

import "errors"

var (
	ErrTeamExists   = errors.New("team exists")
	ErrUserExists   = errors.New("user exists")
	ErrPRExists     = errors.New("PR exists")
	ErrUserInactive = errors.New("user is inactive")

	ErrNotFound = errors.New("not found")
)
