package domain

import "time"

type EventType string

const (
	TaskCreated EventType = "task.created"
	TaskUpdated EventType = "task.updated"
	TaskDeleted EventType = "task.deleted"
	ListDeleted EventType = "list.deleted"
)

type Event struct {
	Type       EventType
	TaskID     string
	ListID     string
	OccurredAt time.Time
	Task       Task
}
