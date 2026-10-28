package httpapi

import (
	"net/http"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/application"
)

// Функция принимает роутер и связывает пути с обработчиками
func RegisterRoutes(mux *http.ServeMux, uc application.TaskUseCases) {
	mux.HandleFunc("/debug/info", HandleServiceInfo)
	mux.HandleFunc("POST /v1/lists", HandleCreateList(uc))
	mux.HandleFunc("GET /v1/lists", HandleGetLists(uc))
	mux.HandleFunc("GET /v1/lists/{list_id}", HandleGetList(uc))

	mux.HandleFunc("POST /v1/lists/{list_id}/tasks", HandleCreateTask(uc))
	mux.HandleFunc("GET /v1/lists/{list_id}/tasks", HandleGetTasks(uc))
	mux.HandleFunc("PATCH /v1/tasks/{task_id}", HandleUpdateTask(uc))
	mux.HandleFunc("DELETE /v1/tasks/{task_id}", HandleDeleteTask(uc))
}
