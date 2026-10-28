package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/domain"
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) CreateList(ctx context.Context, l domain.List) error {
	cmd := `INSERT INTO lists (id, name) VALUES ($1, $2)`
	_, err := s.db.Exec(ctx, cmd, l.ID, l.Name)
	return err
}

func (s *Store) GetList(ctx context.Context, listId string) (domain.List, error) {
	var listName string
	cmd := `SELECT name FROM lists WHERE id = $1`
	err := s.db.QueryRow(ctx, cmd, listId).Scan(&listName)
	if err != nil {
		return domain.List{}, mapErr(err)
	}
	return domain.List{
		ID:   listId,
		Name: listName,
	}, nil
}

func (s *Store) GetLists(ctx context.Context) ([]domain.List, error) {
	cmd := `SELECT id, name FROM lists`
	rows, err := s.db.Query(ctx, cmd)
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

func (s *Store) CreateTask(ctx context.Context, t domain.Task) error {
	cmd := `INSERT INTO tasks (id, list_id, title, done, created_at) 
				VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(ctx, cmd, t.ID, t.ListID, t.Title, t.Done, t.CreatedAt)
	return err
}

func (s *Store) GetTasks(ctx context.Context, listID string) ([]domain.Task, error) {
	cmd := `SELECT id, list_id, title, done, created_at 
				FROM tasks 
				WHERE list_id = $1
				ORDER BY created_at, id`
	rows, err := s.db.Query(ctx, cmd, listID)
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

func (s *Store) UpdateTask(ctx context.Context, id string, p domain.TaskPatch) (domain.Task, error) {
	const cmd = `
		UPDATE tasks
		   SET title = COALESCE($2, title),
		       done  = COALESCE($3, done)
		 WHERE id = $1
		RETURNING id, list_id, title, done, created_at`

	var t domain.Task
	err := s.db.QueryRow(ctx, cmd, id, p.Title, p.Done).
		Scan(&t.ID, &t.ListID, &t.Title, &t.Done, &t.CreatedAt)
	if err != nil {
		return domain.Task{}, mapErr(err)
	}
	return t, err
}

func (s *Store) DeleteTask(ctx context.Context, taskID string) (domain.Task, error) {
	cmd := `DELETE FROM tasks WHERE id = $1 RETURNING id, list_id, title, done, created_at`
	var task domain.Task
	err := s.db.QueryRow(ctx, cmd, taskID).Scan(
		&task.ID, &task.ListID, &task.Title, &task.Done, &task.CreatedAt,
	)
	if err != nil {
		return domain.Task{}, mapErr(err)
	}
	return task, nil
}

func mapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}