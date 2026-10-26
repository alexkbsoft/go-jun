package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/app"
	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Task struct {
	ID        string    `json:"name"`
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
		taskID := uuid.New().String()

		var taskJson struct {
			task Task
		}
		task, err := a.CreateTask(req.Context(), taskID, listID, reqData.Title, false)
		if err != nil {
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			slog.Error("Failed to create task:", slog.String("error", err.Error()))
			return
		}

		taskJson.task.ID = task.ID
		taskJson.task.Title = task.Title
		taskJson.task.ListID = task.ListID
		taskJson.task.Done = task.Done
		taskJson.task.CreatedAt = task.CreatedAt

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(taskJson); err != nil {
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

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string][]models.Task{"tasks": tasks}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}
	}
}

func HandleUpdateTask(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var reqData struct {
			Title string `json:"title"`
			Done  bool   `json:"done"`
		}

		if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		taskID := req.PathValue("task_id")

		task := models.Task{
			ID:    taskID,
			Title: reqData.Title,
			Done:  reqData.Done,
		}

		updatedTask, err := a.UpdateTask(req.Context(), task)
		if err != nil {
			if err == pgx.ErrNoRows {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			slog.Error("Failed to update task:", slog.String("error", err.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(updatedTask); err != nil {
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

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(deletedTask); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}
	}
}
