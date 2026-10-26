package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/app"
	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/models"
	"github.com/jackc/pgx/v5"
)

type Task struct {
	ID        string    `json:"id"`
	ListID    string    `json:"listId"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"createdAt"`
}

func HandleCreateTask(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var reqData struct {
			Title string `json:"title"`
		}

		if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		listID := req.PathValue("list_id")

		task, err := a.CreateTask(req.Context(), listID, reqData.Title, false)
		if err != nil {
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			slog.Error("Failed to create task:", slog.String("error", err.Error()))
			return
		}

		var respTask Task
		respTask.ID = task.ID
		respTask.Title = task.Title
		respTask.ListID = task.ListID
		respTask.Done = task.Done
		respTask.CreatedAt = task.CreatedAt

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respTask); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}
	}
}

func HandleGetTasks(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		listID := req.PathValue("list_id")

		tasks, err := a.GetTasks(req.Context(), listID)
		if err != nil {
			http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
			slog.Error("Failed to retrieve tasks:", slog.String("error", err.Error()))
			return
		}

		var tasksJson []Task

		for _, l := range tasks {
			tasksJson = append(tasksJson, Task{
				ID:        l.ID,
				Title:     l.Title,
				ListID:    l.ListID,
				Done:      l.Done,
				CreatedAt: l.CreatedAt,
			})
		}

		var respData struct {
			Tasks []Task `json:"tasks"`
		}
		respData.Tasks = tasksJson

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respData); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}
	}
}

func HandleUpdateTask(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var reqData struct {
			Title *string `json:"title"`
			Done  *bool   `json:"done"`
		}
		if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		task, err := a.UpdateTask(req.Context(), req.PathValue("task_id"),
			models.TaskPatch{Title: reqData.Title, Done: reqData.Done})
		switch {
		case errors.Is(err, models.ErrInvalid):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case errors.Is(err, models.ErrNotFound):
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		case err != nil:
			slog.Error("Failed to update task", slog.String("error", err.Error()))
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			return
		}

		var respTask Task
		respTask.ID = task.ID
		respTask.Title = task.Title
		respTask.ListID = task.ListID
		respTask.Done = task.Done
		respTask.CreatedAt = task.CreatedAt

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respTask); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}
	}
}

func HandleDeleteTask(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		taskID := req.PathValue("task_id")

		deletedTask, err := a.DeleteTask(req.Context(), taskID)
		if err != nil {
			if err == pgx.ErrNoRows {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to delete task", http.StatusInternalServerError)
			slog.Error("Failed to delete task:", slog.String("error", err.Error()))
			return
		}

		var respTask Task
		respTask.ID = deletedTask.ID
		respTask.Title = deletedTask.Title
		respTask.ListID = deletedTask.ListID
		respTask.Done = deletedTask.Done
		respTask.CreatedAt = deletedTask.CreatedAt

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respTask); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}
	}
}
