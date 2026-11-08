package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alexkbsoft/go-jun/internal/domain"
)

type Postgres struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) CreateList(ctx context.Context, l domain.List) error {
	cmd := `INSERT INTO lists (id, name) VALUES ($1, $2)`
	_, err := p.db.Exec(ctx, cmd, l.ID, l.Name)
	return err
}

func (p *Postgres) GetList(ctx context.Context, listId string) (domain.List, error) {
	var listName string
	cmd := `SELECT name FROM lists WHERE id = $1`
	err := p.db.QueryRow(ctx, cmd, listId).Scan(&listName)
	if err != nil {
		return domain.List{}, mapErr(err)
	}
	return domain.List{
		ID:   listId,
		Name: listName,
	}, nil
}

func (p *Postgres) GetLists(ctx context.Context) ([]domain.List, error) {
	cmd := `SELECT id, name FROM lists`
	rows, err := p.db.Query(ctx, cmd)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []domain.List

	for rows.Next() {
		var list domain.List
		if err := rows.Scan(&list.ID, &list.Name); err != nil {
			return nil, err
		}
		result = append(result, list)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

type taskEvent struct {
	Type       string    `json:"type"`
	TaskID     string    `json:"taskId"`
	ListID     string    `json:"listId"`
	Title      string    `json:"title"`
	Done       bool      `json:"done"`
	OccurredAt time.Time `json:"occurredAt"`
}

type listDeletedEvent struct {
	Type       string    `json:"type"`
	ListID     string    `json:"listId"`
	OccurredAt time.Time `json:"occurredAt"`
}

func (p *Postgres) CreateTaskWithOutbox(ctx context.Context, t domain.Task) error {
	payload, err := json.Marshal(taskEvent{
		Type:       string(domain.TaskCreated),
		TaskID:     t.ID,
		ListID:     t.ListID,
		Title:      t.Title,
		Done:       t.Done,
		OccurredAt: time.Now(),
	})
	if err != nil {
		return err
	}

	tx, err := p.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := p.insertTask(ctx, tx, t); err != nil {
		return err
	}
	if err := p.insertOutbox(ctx, tx, t.ID, string(domain.TaskCreated), payload); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (p *Postgres) UpdateTaskWithOutbox(ctx context.Context, id string, patch domain.TaskPatch) (domain.Task, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return domain.Task{}, err
	}
	defer tx.Rollback(ctx)

	task, err := p.updateTask(ctx, tx, id, patch)
	if err != nil {
		return domain.Task{}, err
	}

	payload, err := json.Marshal(taskEvent{
		Type:       string(domain.TaskUpdated),
		TaskID:     task.ID,
		ListID:     task.ListID,
		Title:      task.Title,
		Done:       task.Done,
		OccurredAt: time.Now(),
	})
	if err != nil {
		return domain.Task{}, err
	}

	if err := p.insertOutbox(ctx, tx, task.ID, string(domain.TaskUpdated), payload); err != nil {
		return domain.Task{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (p *Postgres) DeleteTaskWithOutbox(ctx context.Context, taskID string) (domain.Task, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return domain.Task{}, err
	}
	defer tx.Rollback(ctx)

	task, err := p.deleteTask(ctx, tx, taskID)
	if err != nil {
		return domain.Task{}, err
	}

	payload, err := json.Marshal(taskEvent{
		Type:       string(domain.TaskDeleted),
		TaskID:     task.ID,
		ListID:     task.ListID,
		Title:      task.Title,
		Done:       task.Done,
		OccurredAt: time.Now(),
	})
	if err != nil {
		return domain.Task{}, err
	}

	if err := p.insertOutbox(ctx, tx, task.ID, string(domain.TaskDeleted), payload); err != nil {
		return domain.Task{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (p *Postgres) DeleteListWithOutbox(ctx context.Context, listID string) (domain.List, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return domain.List{}, err
	}
	defer tx.Rollback(ctx)

	list, err := p.deleteList(ctx, tx, listID)
	if err != nil {
		return domain.List{}, err
	}

	payload, err := json.Marshal(listDeletedEvent{
		Type:       string(domain.ListDeleted),
		ListID:     list.ID,
		OccurredAt: time.Now(),
	})
	if err != nil {
		return domain.List{}, err
	}

	if err := p.insertOutbox(ctx, tx, list.ID, string(domain.ListDeleted), payload); err != nil {
		return domain.List{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.List{}, err
	}
	return list, nil
}

func (p *Postgres) GetTasks(ctx context.Context, listID string) ([]domain.Task, error) {
	cmd := `SELECT id, list_id, title, done, created_at 
				FROM tasks 
				WHERE list_id = $1
				ORDER BY created_at, id`
	rows, err := p.db.Query(ctx, cmd, listID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tasks []domain.Task

	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(&task.ID, &task.ListID, &task.Title, &task.Done, &task.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (p *Postgres) insertTask(ctx context.Context, tx pgx.Tx, t domain.Task) error {
	cmd := `INSERT INTO tasks (id, list_id, title, done, created_at)
			VALUES ($1, $2, $3, $4, $5)`
	_, err := tx.Exec(ctx, cmd, t.ID, t.ListID, t.Title, t.Done, t.CreatedAt)
	return err
}

func (p *Postgres) updateTask(ctx context.Context, tx pgx.Tx, id string, patch domain.TaskPatch) (domain.Task, error) {
	const cmd = `
		UPDATE tasks
		   SET title = COALESCE($2, title),
		       done  = COALESCE($3, done)
		 WHERE id = $1
		RETURNING id, list_id, title, done, created_at`

	var t domain.Task
	err := tx.QueryRow(ctx, cmd, id, patch.Title, patch.Done).
		Scan(&t.ID, &t.ListID, &t.Title, &t.Done, &t.CreatedAt)
	if err != nil {
		return domain.Task{}, mapErr(err)
	}
	return t, err
}

func (p *Postgres) deleteTask(ctx context.Context, tx pgx.Tx, taskID string) (domain.Task, error) {
	cmd := `DELETE FROM tasks WHERE id = $1 RETURNING id, list_id, title, done, created_at`
	var t domain.Task
	err := tx.QueryRow(ctx, cmd, taskID).Scan(
		&t.ID, &t.ListID, &t.Title, &t.Done, &t.CreatedAt,
	)
	if err != nil {
		return domain.Task{}, mapErr(err)
	}
	return t, nil
}

func (p *Postgres) deleteList(ctx context.Context, tx pgx.Tx, listID string) (domain.List, error) {
	cmd := `DELETE FROM lists WHERE id = $1 RETURNING id, name`
	var l domain.List
	err := tx.QueryRow(ctx, cmd, listID).Scan(&l.ID, &l.Name)
	if err != nil {
		return domain.List{}, mapErr(err)
	}
	return l, nil
}

func (p *Postgres) insertOutbox(ctx context.Context, tx pgx.Tx, aggregateID, eventType string, payload []byte) error {
	cmd := `INSERT INTO outbox (id, event_type, aggregate_id, json_payload, status, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, 'pending', now())`
	_, err := tx.Exec(ctx, cmd, eventType, aggregateID, payload)
	return err
}

func mapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}
