package models

import "time"

type User struct {
	ID       int64  `db:"id"`
	UserID   string `db:"user_id"` // Внешний идентификатор пользователя
	Name     string `db:"username"`
	TeamID   int64  `db:"team_id"`
	IsActive bool   `db:"is_active"`
	Deleted  *time.Time
}
