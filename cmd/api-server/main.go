package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/adapters/inbound/httpapi"
	rmqin "github.com/alexkbsoft/go-jun/cmd/api-server/internal/adapters/inbound/rabbitmq"
	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/adapters/outbound/postgres"
	rmqout "github.com/alexkbsoft/go-jun/cmd/api-server/internal/adapters/outbound/rabbitmq"
	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/application"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

const PORT = ":8090"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbURL := os.Getenv("API_SERVER_DB_URL")
	if dbURL == "" {
		log.Fatal("API_SERVER_DB_URL is not set")
		os.Exit(1)
	}

	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
		os.Exit(1)
	}
	defer db.Close()

	// conn, err := amqp.Dial(os.Getenv("API_SERVER_AMQP_URL")) // amqp://guest:guest@localhost:5672/
	// defer conn.Close()
	// pubCh, _ := conn.Channel()  // отдельные каналы: для публикации
	// conCh, _ := conn.Channel()  // и для потребления

	publisher, _ := rmqout.NewPublisher("tasks.events")
	repo := postgres.New(db)
	svc := application.NewService(repo, publisher)

	mux := http.NewServeMux()

	httpapi.RegisterRoutes(mux, svc)

	srv := &http.Server{Addr: PORT, Handler: mux}
	consumer := rmqin.NewConsumer("tasks.commands", svc)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
	g.Go(func() error {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(sctx)
	})
	g.Go(func() error { return consumer.Run(ctx) })

	fmt.Println("Server started on", PORT)
	return g.Wait()
}
