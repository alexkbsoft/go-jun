package app

import (
	"context"
	"time"
	"uuid"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/models"
)

type store interface {
	CreateList(ctx context.Context, l models.List) error
	GetList(ctx context.Context, listId string) (models.List, error)
	GetLists(ctx context.Context) ([]models.List, error)
	CreateTask(ctx context.Context, t models.Task) error
	GetTasks(ctx context.Context, listID string) ([]models.Task, error)
	UpdateTask(ctx context.Context, t models.Task) (models.Task, error)
	DeleteTask(ctx context.Context, taskID string) (models.Task, error)
}

type App struct {
	store store
}

func New(s store) *App {
	return &App{
		store: s,
	}
}

func (a *App) GetList(ctx context.Context, listId string) (models.List, error) {
	return a.store.GetList(ctx, listId)
}

func (a *App) CreateList(ctx context.Context, listName string) (models.List, error) {
	list := models.List{
		ID:   uuid.New().String(),
		Name: listName,
	}
	err := a.store.CreateList(ctx, list)
	return list, err

}

func (a *App) GetLists(ctx context.Context) ([]models.List, error) {
	return a.store.GetLists(ctx)
}

func (a *App) CreateTask(ctx context.Context, listId string, title string, done bool) (models.Task, error) {
	task := models.Task{
		ID:        uuid.New().String(),
		Title:     title,
		Done:      done,
		ListID:    listId,
		CreatedAt: time.Now(),
	}
	err := a.store.CreateTask(ctx, task)
	return task, err
}
func (a *App) GetTasks(ctx context.Context, listID string) ([]models.Task, error) {
	return a.store.GetTasks(ctx, listID)
}
func (a *App) UpdateTask(ctx context.Context, t models.Task) (models.Task, error) {
	return a.store.UpdateTask(ctx, t)
}

func (a *App) DeleteTask(ctx context.Context, taskID string) (models.Task, error) {
	return a.store.DeleteTask(ctx, taskID)
}
