package models

import "time"

type Team struct {
	ID      int64
	Name    string
	Deleted *time.Time
}
