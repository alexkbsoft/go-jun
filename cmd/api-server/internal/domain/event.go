package domain

import "time"

type EventType string

const (
	TaskCreated EventType = "task.created"
	TaskUpdated EventType = "task.updated"
	TaskDeleted EventType = "task.deleted"
)

type Event struct {
	Type       EventType
	TaskID     string
	ListID     string
	OccurredAt time.Time
}