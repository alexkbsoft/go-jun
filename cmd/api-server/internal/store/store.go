package store

import (
	"context"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) CreateList(ctx context.Context, l models.List) error {
	cmd := `INSERT INTO lists (id, name) VALUES ($1, $2)`
	_, err := s.db.Exec(ctx, cmd, l.ID, l.Name)
	return err
}

func (s *Store) GetList(ctx context.Context, listId string) (models.List, error) {
	var listName string
	cmd := `SELECT name FROM lists WHERE id = $1`
	err := s.db.QueryRow(ctx, cmd, listId).Scan(&listName)
	if err != nil {
		return models.List{}, err
	}
	return models.List{
		ID:   listId,
		Name: listName,
	}, nil
}

func (s *Store) GetLists(ctx context.Context) ([]models.List, error) {
	cmd := `SELECT id, name FROM lists`
	rows, err := s.db.Query(ctx, cmd)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []models.List

	for rows.Next() {
		var list models.List
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

func (s *Store) CreateTask(ctx context.Context, t models.Task) error {
	cmd := `INSERT INTO tasks (id, list_id, title, done, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := s.db.Exec(ctx, cmd, t.ID, t.ListID, t.Title, t.Done, t.CreatedAt)
	return err
}

func (s *Store) GetTasks(ctx context.Context, listID string) ([]models.Task, error) {
	cmd := `SELECT id, list_id, title, done, created_at FROM tasks WHERE list_id = $1`
	rows, err := s.db.Query(ctx, cmd, listID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {
		var task models.Task
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

func (s *Store) UpdateTask(ctx context.Context, t models.Task) (models.Task, error) {
	cmd := `UPDATE tasks SET title = $1, done = $2 WHERE id = $3 RETURNING id, list_id, title, done, created_at`
	var task models.Task
	err := s.db.QueryRow(ctx, cmd, t.Title, t.Done, t.ID).Scan(
		&task.ID, &task.ListID, &task.Title, &task.Done, &task.CreatedAt,
	)
	if err != nil {
		return models.Task{}, err
	}
	return task, nil
}

func (s *Store) DeleteTask(ctx context.Context, taskID string) (models.Task, error) {
	cmd := `DELETE FROM tasks WHERE id = $1 RETURNING id, list_id, title, done, created_at`
	var task models.Task
	err := s.db.QueryRow(ctx, cmd, taskID).Scan(
		&task.ID, &task.ListID, &task.Title, &task.Done, &task.CreatedAt,
	)
	if err != nil {
		return models.Task{}, err
	}
	return task, nil
}