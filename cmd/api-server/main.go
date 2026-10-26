package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/handler"
	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/store"
	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/app"

)

const PORT = ":8090"

func main() {
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

	store := store.New(db)
	app := app.New(store)

	mux := http.NewServeMux()
	// Вызываем вынесенную регистрацию маршрутов
	handler.RegisterRoutes(mux, app)

	fmt.Println("Server started on", PORT)
	http.ListenAndServe(PORT, mux)
}
