package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/alexkbsoft/go-jun/internal/domain"
)

// TelegramClient — порт для MAX Bot
type TelegramClient interface {
	SendText(ctx context.Context, text string) error
}

// InboxRepository — порт для работы с inbox и tasks
type InboxRepository interface {
	InsertInboxEvent(ctx context.Context, event domain.InboxEvent) error
	MarkInboxProcessed(ctx context.Context, eventID string) error
	MarkInboxRejected(ctx context.Context, eventID string) error
	UpsertTask(ctx context.Context, t domain.Task) error
	DeleteTask(ctx context.Context, taskID string) error
}

type InboxService struct {
	repo   InboxRepository
	maxBot TelegramClient
}

func NewInboxService(repo InboxRepository, maxBot TelegramClient) *InboxService {
	return &InboxService{
		repo:   repo,
		maxBot: maxBot,
	}
}

func (s *InboxService) ProcessMessage(ctx context.Context, body []byte) error {
	// Парсим входящее событие
	var raw struct {
		Type       string    `json:"type"`
		TaskID     string    `json:"taskId"`
		ListID     string    `json:"listId"`
		Title      string    `json:"title"`
		Done       bool      `json:"done"`
		OccurredAt time.Time `json:"occurredAt"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return fmt.Errorf("parse event: %w", err)
	}

	event := domain.InboxEvent{
		EventType:   raw.Type,
		AggregateID: raw.TaskID,
		Payload:     body,
		OccurredAt:  raw.OccurredAt,
	}

	// 1. Проверяем идемпотентность — уникальный индекс (event_type, aggregate_id, occurred_at)
	err := s.repo.InsertInboxEvent(ctx, event)
	if err != nil {
		// Дубликат — игнорируем
		slog.Info("duplicate event, skipping",
			slog.String("type", raw.Type),
			slog.String("task_id", raw.TaskID),
		)
		return nil
	}

	// 2. Обновляем tasks и отправляем уведомление
	return s.processTaskEvent(ctx, raw.Type, raw.TaskID, raw.ListID, raw.Title, raw.Done)
}

func (s *InboxService) processTaskEvent(ctx context.Context, eventType, taskID, listID, title string, done bool) error {
	switch eventType {
	case "task.created":
		task := domain.Task{
			ID:        taskID,
			ListID:    listID,
			Title:     title,
			Done:      done,
			CreatedAt: time.Now(),
		}
		if err := s.repo.UpsertTask(ctx, task); err != nil {
			return err
		}
		// Уведомляем
		if err := s.notifyMax(ctx, fmt.Sprintf("задача с именем %s создана", title)); err != nil {
			slog.Error("notify max", slog.String("error", err.Error()))
		}

	case "task.updated":
		// Обновляем задачу (ON CONFLICT)
		task := domain.Task{
			ID:     taskID,
			ListID: listID,
			Title:  title,
			Done:   done,
		}
		if err := s.repo.UpsertTask(ctx, task); err != nil {
			return err
		}
		// Уведомляем только если стало done
		if done {
			if err := s.notifyMax(ctx, fmt.Sprintf("задача с именем %s завершена", title)); err != nil {
				slog.Error("notify max", slog.String("error", err.Error()))
			}
		}

	case "task.deleted":
		if err := s.repo.DeleteTask(ctx, taskID); err != nil {
			return err
		}

	default:
		slog.Warn("unknown event type", slog.String("type", eventType))
	}

	return nil
}

func (s *InboxService) notifyMax(ctx context.Context, text string) error {
	return s.maxBot.SendText(ctx, text)
}
