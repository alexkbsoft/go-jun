package application

import (
	"context"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/domain"
)

// Входящий порт: что ядро умеет. Его вызывают HTTP и RabbitMQ.
type TaskUseCases interface {
	CreateList(ctx context.Context, name string) (domain.List, error)
	GetList(ctx context.Context, listID string) (domain.List, error)
	GetLists(ctx context.Context) ([]domain.List, error)
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
	CreateTask(ctx context.Context, t domain.Task) error
	GetTasks(ctx context.Context, listID string) ([]domain.Task, error)
	UpdateTask(ctx context.Context, id string, p domain.TaskPatch) (domain.Task, error)
	DeleteTask(ctx context.Context, id string) (domain.Task, error)
}

// Исходящий порт: публикация событий. Реализует RabbitMQ.
type EventPublisher interface {
	Publish(ctx context.Context, e domain.Event) error
}