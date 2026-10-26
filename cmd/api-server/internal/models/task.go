package models

import "time"

type Task struct {
	ID        string
	ListID    string
	Title     string
	Done      bool
	CreatedAt time.Time
}
