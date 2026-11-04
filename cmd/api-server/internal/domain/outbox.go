package domain

import "time"

type Outbox struct {
	ID          string
	EventType   string
	AggregateID string
	Payload     []byte
	Status      string // "pending" | "published"
	CreatedAt   time.Time
	PublishedAt *time.Time
}
