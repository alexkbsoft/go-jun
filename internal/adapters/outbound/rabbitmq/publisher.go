package rmqout

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

type Publisher struct {
	db       *pgxpool.Pool
	env      *rmq.Environment
	conn     *rmq.AmqpConnection
	exchange string
}

func NewPublisher(db *pgxpool.Pool, amqpURL, exchange string) (*Publisher, error) {
	// 1. Создаём окружение (shared settings)
	env := rmq.NewEnvironment(amqpURL, nil)

	// 2. Создаём соединение с брокером
	conn, err := env.NewConnection(context.Background())
	if err != nil {
		return nil, fmt.Errorf("rabbitmq connect: %w", err)
	}

	// 3. Объявляем exchange (idempotent — не создаст, если уже есть)
	_, err = conn.Management().DeclareExchange(context.Background(), &rmq.TopicExchangeSpecification{
		Name: exchange,
	})
	if err != nil {
		conn.Close(context.Background())
		env.CloseConnections(context.Background())
		return nil, fmt.Errorf("declare exchange %s: %w", exchange, err)
	}

	// 4. Объявляем очередь
	queueName := "tasks.events"
	_, err = conn.Management().DeclareQueue(context.Background(), &rmq.QuorumQueueSpecification{
		Name: queueName,
	})
	if err != nil {
		conn.Close(context.Background())
		env.CloseConnections(context.Background())
		return nil, fmt.Errorf("declare queue %s: %w", queueName, err)
	}

	// 5. Привязываем очередь к exchange (routing key = "#" — ловим все события)
	_, err = conn.Management().Bind(context.Background(), &rmq.ExchangeToQueueBindingSpecification{
		SourceExchange:   exchange,
		DestinationQueue: queueName,
		BindingKey:       "#",
	})
	if err != nil {
		conn.Close(context.Background())
		env.CloseConnections(context.Background())
		return nil, fmt.Errorf("bind queue: %w", err)
	}

	return &Publisher{
		db:       db,
		env:      env,
		conn:     conn,
		exchange: exchange,
	}, nil
}

// PublishPending читает pending-записи из outbox, публикует в RabbitMQ
// и помечает как published — всё в одной транзакции БД
func (p *Publisher) PublishPending(ctx context.Context, batchSize int) error {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Читаем pending-записи (SKIP LOCKED — другие poller не прочитают)
	rows, err := tx.Query(ctx,
		`SELECT id, event_type, aggregate_id, json_payload
		 FROM outbox
		 WHERE status = 'pending'
		 ORDER BY created_at ASC
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`,
		batchSize,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	type outboxRow struct {
		id          string
		eventType   string
		aggregateID string
		payload     []byte
	}

	var records []outboxRow
	for rows.Next() {
		var r outboxRow
		if err := rows.Scan(&r.id, &r.eventType, &r.aggregateID, &r.payload); err != nil {
			return err
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(records) == 0 {
		return tx.Commit(ctx)
	}

	// 2. Публикуем каждую запись
	for _, r := range records {
		// Создаём publisher для exchange с routing key = event type
		publisher, err := p.conn.NewPublisher(ctx, &rmq.ExchangeAddress{
			Exchange: p.exchange,
			Key:      r.eventType,
		}, nil)
		if err != nil {
			return fmt.Errorf("create publisher: %w", err)
		}

		msg := rmq.NewMessage(r.payload)

		res, err := publisher.Publish(ctx, msg)
		if err != nil {
			publisher.Close(ctx)
			slog.Error("publish failed",
				slog.String("outbox_id", r.id),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("publish outbox %s: %w", r.id, err)
		}

		switch res.Outcome.(type) {
		case *rmq.StateAccepted:
			slog.Info("published event",
				slog.String("type", r.eventType),
				slog.String("aggregate", r.aggregateID),
			)
		case *rmq.StateReleased:
			slog.Error("message not routed",
				slog.String("outbox_id", r.id),
				slog.String("routing_key", r.eventType),
			)
			publisher.Close(ctx)
			return fmt.Errorf("message not routed for outbox %s", r.id)
		case *rmq.StateRejected:
			slog.Error("message rejected",
				slog.String("outbox_id", r.id),
			)
			publisher.Close(ctx)
			return fmt.Errorf("message rejected for outbox %s", r.id)
		}

		_ = publisher.Close(ctx)
	}

	// 3. Помечаем все как published
	for _, r := range records {
		_, err := tx.Exec(ctx,
			`UPDATE outbox SET status = 'published', published_at = now() WHERE id = $1`,
			r.id,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// Close закрывает соединение с RabbitMQ
func (p *Publisher) Close() error {
	_ = p.conn.Close(context.Background())
	return p.env.CloseConnections(context.Background())
}