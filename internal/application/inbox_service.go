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
	SendText(text string) error
}

// InboxPendingEvent — сырое событие из RabbitMQ, ещё не обработанное
type InboxPendingEvent struct {
	EventType   string
	AggregateID string
	Payload     []byte
	OccurredAt  time.Time
}

// InboxRepository — порт для работы с inbox и tasks
type InboxRepository interface {
	InsertInboxEvent(ctx context.Context, event InboxPendingEvent) error
	ClaimPendingEvents(ctx context.Context, batchSize int) ([]domain.InboxEvent, error)
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

// InsertInboxEvent записывает событие из RabbitMQ в inbox в транзакции
func (s *InboxService) InsertInboxEvent(ctx context.Context, event InboxPendingEvent) error {
	return s.repo.InsertInboxEvent(ctx, event)
}

// ProcessPendingBatch читает pending-события из inbox и обрабатывает пачкой
func (s *InboxService) ProcessPendingBatch(ctx context.Context) error {
	events, err := s.repo.ClaimPendingEvents(ctx, 100)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		return nil
	}

	slog.Info("processing inbox batch", slog.Int("count", len(events)))

	for _, event := range events {
		if err := s.processSingleEvent(ctx, event); err != nil {
			slog.Error("process inbox event",
				slog.String("event_id", event.ID),
				slog.String("error", err.Error()),
			)
			// Помечаем как rejected — чтобы не крутить бесконечно
			_ = s.repo.MarkInboxRejected(ctx, event.ID)
		} else {
			_ = s.repo.MarkInboxProcessed(ctx, event.ID)
		}
	}

	return nil
}

func (s *InboxService) processSingleEvent(ctx context.Context, event domain.InboxEvent) error {
	// Парсим payload
	var raw struct {
		Type       string    `json:"type"`
		TaskID     string    `json:"taskId"`
		ListID     string    `json:"listId"`
		Title      string    `json:"title"`
		Done       bool      `json:"done"`
	}
	if err := json.Unmarshal(event.Payload, &raw); err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	switch raw.Type {
	case "task.created":
		task := domain.Task{
			ID:        raw.TaskID,
			ListID:    raw.ListID,
			Title:     raw.Title,
			Done:      raw.Done,
			CreatedAt: time.Now(),
		}
		if err := s.repo.UpsertTask(ctx, task); err != nil {
			return err
		}
		// Уведомляем
		if err := s.notifyMax(ctx, fmt.Sprintf("задача с именем %s создана", raw.Title)); err != nil {
			slog.Error("notify max", slog.String("error", err.Error()))
		}

	case "task.updated":
		task := domain.Task{
			ID:     raw.TaskID,
			ListID: raw.ListID,
			Title:  raw.Title,
			Done:   raw.Done,
		}
		if err := s.repo.UpsertTask(ctx, task); err != nil {
			return err
		}
		// Уведомляем только если стало done
		if raw.Done {
			if err := s.notifyMax(ctx, fmt.Sprintf("задача с именем %s завершена", raw.Title)); err != nil {
				slog.Error("notify max", slog.String("error", err.Error()))
			}
		}

	case "task.deleted":
		if err := s.repo.DeleteTask(ctx, raw.TaskID); err != nil {
			return err
		}

	case "list.deleted":
		// Список удалён — уведомим пользователя
		if err := s.notifyMax(ctx, "список удалён"); err != nil {
			slog.Error("notify max", slog.String("error", err.Error()))
		}

	default:
		slog.Warn("unknown event type", slog.String("type", raw.Type))
	}

	return nil
}

func (s *InboxService) notifyMax(_ context.Context, text string) error {
	return s.maxBot.SendText(text)
}
