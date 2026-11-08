package application

import (
	"context"

	"github.com/alexkbsoft/go-jun/internal/domain"
)

// Входящий порт: что ядро умеет. Его вызывают HTTP и RabbitMQ.
type TaskUseCases interface {
	CreateList(ctx context.Context, name string) (domain.List, error)
	GetList(ctx context.Context, listID string) (domain.List, error)
	GetLists(ctx context.Context) ([]domain.List, error)
	DeleteList(ctx context.Context, listID string) (domain.List, error)
	CreateTask(ctx context.Context, listID, title string, done bool) (domain.Task, error)
	GetTasks(ctx context.Context, listID string) ([]domain.Task, error)
	UpdateTask(ctx context.Context, id string, p domain.TaskPatch) (domain.Task, error)
	DeleteTask(ctx context.Context, id string) (domain.Task, error)
}

// Исходящий порт: хранилище. Реализует Postgres.
type Repository interface {
	CreateList(ctx context.Context, l domain.List) error
	GetList(ctx context.Context, listID string) (domain.List, error)
	GetLists(ctx context.Context) ([]domain.List, error)
	DeleteListWithOutbox(ctx context.Context, listID string) (domain.List, error)
	GetTasks(ctx context.Context, listID string) ([]domain.Task, error)
	CreateTaskWithOutbox(ctx context.Context, task domain.Task) error
	UpdateTaskWithOutbox(ctx context.Context, id string, p domain.TaskPatch) (domain.Task, error)
	DeleteTaskWithOutbox(ctx context.Context, taskID string) (domain.Task, error)
}
