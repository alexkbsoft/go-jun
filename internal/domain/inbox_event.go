package domain

import "time"

type InboxEvent struct {
	ID          string
	EventType   string
	AggregateID string
	Payload     []byte
	OccurredAt  time.Time
	Status      string // pending | processed | rejected
	CreatedAt   time.Time
	ProcessedAt *time.Time
}
