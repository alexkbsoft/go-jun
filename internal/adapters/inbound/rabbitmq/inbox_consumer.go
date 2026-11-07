package rmqin

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/alexkbsoft/go-jun/internal/application"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

type InboxConsumer struct {
	conn  *rmq.AmqpConnection
	queue string
	svc   *application.InboxService
}

func NewInboxConsumer(conn *rmq.AmqpConnection, queue string, svc *application.InboxService) (*InboxConsumer, error) {
	return &InboxConsumer{
		conn:  conn,
		queue: queue,
		svc:   svc,
	}, nil
}

func (c *InboxConsumer) Run(ctx context.Context) error {
	consumer, err := c.conn.NewConsumer(ctx, c.queue, &rmq.ConsumerOptions{
		InitialCredits: 1,
		SettleStrategy: rmq.ExplicitSettle,
	})
	if err != nil {
		return err
	}
	defer consumer.Close(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			delivery, err := consumer.Receive(ctx)
			if err != nil {
				slog.Error("consumer receive", slog.String("error", err.Error()))
				continue
			}

			if err := c.handleMessage(ctx, delivery); err != nil {
				slog.Error("handle message", slog.String("error", err.Error()))
				// Ошибка записи в inbox — Reject (не возвращаем в очередь, битое сообщение)
				_ = delivery.Discard(ctx, nil)
			} else {
				// Записано в inbox (или дубликат) — Accept
				_ = delivery.Accept(ctx)
			}
		}
	}
}

func (c *InboxConsumer) handleMessage(ctx context.Context, delivery rmq.IDeliveryContext) error {
	msg := delivery.Message()
	if msg == nil || len(msg.Data) == 0 {
		return nil
	}
	body := msg.GetData()
	if body == nil {
		return nil
	}

	// Парсим payload для извлечения occurred_at
	var raw struct {
		Type       string    `json:"type"`
		TaskID     string    `json:"taskId"`
		ListID     string    `json:"listId"`
		Title      string    `json:"title"`
		Done       bool      `json:"done"`
		OccurredAt time.Time `json:"occurredAt"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return err
	}

	// Записываем в inbox в транзакции
	event := application.InboxPendingEvent{
		EventType:   raw.Type,
		AggregateID: raw.TaskID,
		Payload:     body,
		OccurredAt:  raw.OccurredAt,
	}

	// Если ошибка — это не дубликат (ON CONFLICT DO NOTHING не генерирует ошибку)
	// Значит Reject. Если успех — Accept (и новая запись, и дубликат).
	return c.svc.InsertInboxEvent(ctx, event)
}
