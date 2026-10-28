package application

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"uuid"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/domain"
)

type store interface {
	CreateList(ctx context.Context, l domain.List) error
	GetList(ctx context.Context, listId string) (domain.List, error)
	GetLists(ctx context.Context) ([]domain.List, error)
	CreateTask(ctx context.Context, t domain.Task) error
	GetTasks(ctx context.Context, listID string) ([]domain.Task, error)
	UpdateTask(ctx context.Context, id string, p domain.TaskPatch) (domain.Task, error)
	DeleteTask(ctx context.Context, taskID string) (domain.Task, error)
}

type Service struct {
	repo   Repository
	events EventPublisher
}

func NewService(r Repository, p EventPublisher) *Service {
	return &Service{repo: r, events: p}
}

var _ TaskUseCases = (*Service)(nil)

func (s *Service) GetList(ctx context.Context, listId string) (domain.List, error) {
	return s.repo.GetList(ctx, listId)
}

func (s *Service) CreateList(ctx context.Context, listName string) (domain.List, error) {
	list := domain.List{
		ID:   uuid.New().String(),
		Name: listName,
	}
	err := s.repo.CreateList(ctx, list)
	return list, err

}

func (s *Service) GetLists(ctx context.Context) ([]domain.List, error) {
	return s.repo.GetLists(ctx)
}

func (s *Service) CreateTask(ctx context.Context, listId string, title string, done bool) (domain.Task, error) {
	task := domain.Task{
		ID:        uuid.New().String(),
		Title:     title,
		Done:      done,
		ListID:    listId,
		CreatedAt: time.Now(),
	}
	if err := s.repo.CreateTask(ctx, task); err != nil {
		return domain.Task{}, err
	}
	s.publish(ctx, domain.TaskCreated, task)
	return task, nil
}
func (s *Service) GetTasks(ctx context.Context, listID string) ([]domain.Task, error) {
	return s.repo.GetTasks(ctx, listID)
}
func (s *Service) UpdateTask(ctx context.Context, id string, p domain.TaskPatch) (domain.Task, error) {
	if p.Title == nil && p.Done == nil {
		return domain.Task{}, fmt.Errorf("%w: nothing to update", domain.ErrInvalid)
	}
	if p.Title != nil && strings.TrimSpace(*p.Title) == "" {
		return domain.Task{}, fmt.Errorf("%w: title must not be empty", domain.ErrInvalid)
	}
	task, err := s.repo.UpdateTask(ctx, id, p)
	if err != nil {
		return domain.Task{}, err
	}
	s.publish(ctx, domain.TaskUpdated, task)
	return task, nil
}

func (s *Service) DeleteTask(ctx context.Context, taskID string) (domain.Task, error) {
	task, err := s.repo.DeleteTask(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	s.publish(ctx, domain.TaskDeleted, task)
	return task, nil
}

func (s *Service) publish(ctx context.Context, t domain.EventType, task domain.Task) {
	err := s.events.Publish(ctx, domain.Event{
		Type: t, TaskID: task.ID, ListID: task.ListID, OccurredAt: time.Now(),
	})
	if err != nil {
		// Задача уже сохранена; падение брокера не должно ломать запрос.
		slog.Error("publish event failed", slog.String("error", err.Error()))
	}
}
