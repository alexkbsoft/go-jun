package handler

import (
	"net/http"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/app"
)

// Функция принимает роутер и связывает пути с обработчиками
func RegisterRoutes(mux *http.ServeMux, app *app.App) {
	mux.HandleFunc("/debug/info", HandleServiceInfo)
	mux.HandleFunc("POST /v1/lists", HandleCreateList(app))
	mux.HandleFunc("GET /v1/lists", HandleGetLists(app))
	mux.HandleFunc("GET /v1/lists/{list_id}", HandleGetList(app))

	mux.HandleFunc("POST /v1/lists/{list_id}/tasks", HandleCreateTask(app))
	mux.HandleFunc("GET /v1/lists/{list_id}/tasks", HandleGetTasks(app))
	mux.HandleFunc("PATCH /v1/tasks/{task_id}", HandleUpdateTask(app))
	mux.HandleFunc("DELETE /v1/tasks/{task_id}", HandleDeleteTask(app))
}
