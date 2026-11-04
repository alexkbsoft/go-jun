package application

import (
	"context"
	"fmt"
	"strings"
	"time"
	"uuid"

	"github.com/alexkbsoft/go-jun/internal/domain"
)

type Service struct {
	repo Repository
}

func NewService(r Repository) *Service {
	return &Service{repo: r}
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
	if err := s.repo.CreateTaskWithOutbox(ctx, task); err != nil {
		return domain.Task{}, err
	}
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
	task, err := s.repo.UpdateTaskWithOutbox(ctx, id, p)
	if err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (s *Service) DeleteTask(ctx context.Context, taskID string) (domain.Task, error) {
	task, err := s.repo.DeleteTaskWithOutbox(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	return task, nil
}
