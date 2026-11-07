package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	

	"github.com/alexkbsoft/go-jun/internal/domain"
	"github.com/jackc/pgx/v5"
)

// --- Inbox ---

func (p *Postgres) InsertInboxEvent(ctx context.Context, event domain.InboxEvent) error {
	cmd := `
		INSERT INTO inbox (id, event_type, aggregate_id, json_payload, occurred_at, status)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, 'pending')
		ON CONFLICT (event_type, aggregate_id, occurred_at) DO NOTHING
	`
	_, err := p.db.Exec(ctx, cmd, event.EventType, event.AggregateID, event.Payload, event.OccurredAt)
	return err
}

func (p *Postgres) ClaimInboxEvents(ctx context.Context, batchSize int) ([]domain.InboxEvent, error) {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT id, event_type, aggregate_id, json_payload, occurred_at, status, created_at
		FROM inbox
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, batchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.InboxEvent
	for rows.Next() {
		var e domain.InboxEvent
		if err := rows.Scan(&e.ID, &e.EventType, &e.AggregateID, &e.Payload, &e.OccurredAt, &e.Status, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, tx.Commit(ctx)
}

func (p *Postgres) MarkInboxProcessed(ctx context.Context, eventID string) error {
	_, err := p.db.Exec(ctx,
		`UPDATE inbox SET status = 'processed', processed_at = now() WHERE id = $1`,
		eventID,
	)
	return err
}

func (p *Postgres) MarkInboxRejected(ctx context.Context, eventID string) error {
	_, err := p.db.Exec(ctx,
		`UPDATE inbox SET status = 'rejected', processed_at = now() WHERE id = $1`,
		eventID,
	)
	return err
}

// --- Tasks ---

func (p *Postgres) UpsertTask(ctx context.Context, t domain.Task) error {
	cmd := `
		INSERT INTO tasks (id, list_id, title, done, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			done  = EXCLUDED.done
	`
	_, err := p.db.Exec(ctx, cmd, t.ID, t.ListID, t.Title, t.Done, t.CreatedAt)
	return err
}

func (p *Postgres) DeleteTask(ctx context.Context, taskID string) error {
	_, err := p.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, taskID)
	return err
}

// --- Helpers ---

func parseTaskPayload(payload []byte) (string, string, string, bool, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return "", "", "", false, fmt.Errorf("parse json: %w", err)
	}

	taskID, _ := raw["taskId"].(string)
	listID, _ := raw["listId"].(string)
	title, _ := raw["title"].(string)
	done := false
	if d, ok := raw["done"].(bool); ok {
		done = d
	}

	return taskID, listID, title, done, nil
}
