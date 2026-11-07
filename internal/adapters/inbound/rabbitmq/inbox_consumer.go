package rmqin

import (
	"context"
	"log/slog"

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
	// Создаём consumer с prefetch=1 (InitialCredits)
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
			// Receive блокирует до получения сообщения
			delivery, err := consumer.Receive(ctx)
			if err != nil {
				slog.Error("consumer receive", slog.String("error", err.Error()))
				continue
			}

			if err := c.handleMessage(ctx, delivery); err != nil {
				slog.Error("handle message",
					slog.String("error", err.Error()),
				)
				// Плохое сообщение — discarding (не возвращаем в очередь)
				_ = delivery.Discard(ctx, nil)
			} else {
				// Обработано — подтверждаем
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
	// Data — это [][]byte, собираем в один []byte
	body := msg.GetData()
	if body == nil {
		return nil
	}
	return c.svc.ProcessMessage(ctx, body)
}
