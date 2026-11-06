package main

import (
	"context"
	"log/slog"
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
		slog.Error("missing env vars",
			slog.Bool("has_db_url", dbURL != ""),
			slog.Bool("has_amqp_url", amqpURL != ""),
		)
		os.Exit(1)
	}

	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("db connect", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	publisher, err := rmqout.NewPublisher(db, amqpURL, "tasks.events")
	if err != nil {
		slog.Error("publisher connect", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer publisher.Close()

	if err := db.Ping(ctx); err != nil {
		slog.Error("db ping", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Outbox poller started")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Outbox poller stopped")
			return
		case <-ticker.C:
			if err := publisher.PublishPending(ctx, 100); err != nil {
				slog.Error("outbox poll error", slog.String("error", err.Error()))
			}
		}
	}
}
