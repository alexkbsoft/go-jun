package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	rmqin "github.com/alexkbsoft/go-jun/internal/adapters/inbound/rabbitmq"
	"github.com/alexkbsoft/go-jun/internal/adapters/outbound/maxbot"
	"github.com/alexkbsoft/go-jun/internal/adapters/outbound/postgres"
	"github.com/alexkbsoft/go-jun/internal/application"
	"github.com/jackc/pgx/v5/pgxpool"
	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbURL := os.Getenv("INBOX_POLLER_DB_URL")
	amqpURL := os.Getenv("INBOX_POLLER_AMQP_URL")
	maxBotToken := os.Getenv("MAX_BOT_TOKEN")
	maxChatID := os.Getenv("MAX_CHAT_ID")

	if dbURL == "" || amqpURL == "" || maxBotToken == "" {
		slog.Error("missing required env vars",
			slog.Bool("has_db_url", dbURL != ""),
			slog.Bool("has_amqp_url", amqpURL != ""),
			slog.Bool("has_max_token", maxBotToken != ""),
			slog.Bool("has_max_chat_id", maxChatID != ""),
		)
		os.Exit(1)
	}

	// Подключение к БД inbox
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("db connect", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// Подключение к RabbitMQ
	env := rmq.NewEnvironment(amqpURL, nil)
	conn, err := env.NewConnection(ctx)
	if err != nil {
		slog.Error("rabbitmq connect", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer conn.Close(ctx)
	defer env.CloseConnections(ctx)

	// Объявляем exchange
	_, err = conn.Management().DeclareExchange(ctx, &rmq.TopicExchangeSpecification{
		Name: "tasks.events",
	})
	if err != nil {
		slog.Error("declare exchange", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Объявляем queue для inbox-консюмера
	_, err = conn.Management().DeclareQueue(ctx, &rmq.QuorumQueueSpecification{
		Name: "tasks.inbox",
	})
	if err != nil {
		slog.Error("declare queue", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Binding: exchange → queue по routing key "#"
	_, err = conn.Management().Bind(ctx, &rmq.ExchangeToQueueBindingSpecification{
		SourceExchange:   "tasks.events",
		DestinationQueue: "tasks.inbox",
		BindingKey:       "#",
	})
	if err != nil {
		slog.Error("bind queue", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// MAX Bot клиент
	maxClient := maxbot.New(maxBotToken, maxChatID)

	// Репозиторий
	repo := postgres.New(db)

	// Сервис
	inboxSvc := application.NewInboxService(repo, maxClient)

	// Проверка соединения с БД
	if err := db.Ping(ctx); err != nil {
		slog.Error("db ping", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Inbox poller started")

	// Создаём consumer для RabbitMQ с prefetch=1
	consumer, err := rmqin.NewInboxConsumer(conn, "tasks.inbox", inboxSvc)
	if err != nil {
		slog.Error("consumer connect", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Запускаем consumer в горутине
	go func() {
		if err := consumer.Run(ctx); err != nil {
			slog.Error("consumer run", slog.String("error", err.Error()))
		}
	}()

	// Тикер для обработки pending-событий из inbox → tasks + MAX
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Inbox poller stopped")
			return
		case <-ticker.C:
			if err := inboxSvc.ProcessPendingBatch(ctx); err != nil {
				slog.Error("process pending batch", slog.String("error", err.Error()))
			}
		}
	}
}
