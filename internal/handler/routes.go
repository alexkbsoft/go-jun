package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Функция принимает роутер и связывает пути с обработчиками
func RegisterRoutes(mux *http.ServeMux, db *pgxpool.Pool) {
	mux.HandleFunc("/debug/info", HandleServiceInfo)
	mux.HandleFunc("POST /v1/lists", HandleCreateList(db))
	mux.HandleFunc("GET /v1/lists", HandleGetLists(db))
	mux.HandleFunc("GET /v1/lists/{list_id}", HandleGetList(db))

	mux.HandleFunc("POST /v1/lists/{list_id}/tasks", HandleCreateTask(db))
	mux.HandleFunc("GET /v1/lists/{list_id}/tasks", HandleGetTasks(db))
	mux.HandleFunc("PATCH /v1/tasks/{task_id}", HandleUpdateTask(db))
	mux.HandleFunc("DELETE /v1/tasks/{task_id}", HandleDeleteTask(db))
}
