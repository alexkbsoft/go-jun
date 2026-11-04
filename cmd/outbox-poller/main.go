package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	rmqout "github.com/alexkbsoft/go-jun/internal/adapters/outbound/rabbitmq"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbURL := os.Getenv("API_SERVER_DB_URL")
	amqpURL := os.Getenv("API_SERVER_AMQP_URL")
	if dbURL == "" || amqpURL == "" {
		log.Fatal("API_SERVER_DB_URL and API_SERVER_AMQP_URL must be set")
	}

	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("db connect:", err)
	}
	defer db.Close()

	publisher, err := rmqout.NewPublisher(db, amqpURL, "tasks.events")
	if err != nil {
		log.Fatal("publisher connect:", err)
	}
	defer publisher.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatal("db ping:", err)
	}

	log.Println("Outbox poller started")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Outbox poller stopped")
			return
		case <-ticker.C:
			if err := publisher.PublishPending(ctx, 100); err != nil {
				log.Println("outbox poll error:", err)
			}
		}
	}
}