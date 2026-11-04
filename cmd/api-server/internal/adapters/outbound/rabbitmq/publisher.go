package rmqout

import (
	"context"
	// "encoding/json"
	// "fmt"
	// "time"

	// "github.com/alexkbsoft/go-jun/cmd/api-server/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Publisher struct {
	db       *pgxpool.Pool
	exchange string
}

func NewPublisher(db *pgxpool.Pool, exchange string) *Publisher {
	return &Publisher{db: db, exchange: exchange}
}

// PublishPending читает pending-записи из outbox и публикует их
func (p *Publisher) PublishPending(ctx context.Context, batchSize int) error {
	// 1. SELECT ... FROM outbox WHERE status = 'pending' LIMIT $1 FOR UPDATE SKIP LOCKED
	// 2. Для каждой записи — publish в RabbitMQ
	// 3. UPDATE outbox SET status = 'published', published_at = now() WHERE id = $1
	// Всё в одной транзакции на пакет!
	return nil
}
