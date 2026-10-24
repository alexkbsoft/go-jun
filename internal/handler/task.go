package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Task struct {
	ID        string    `json:"id"         db:"id"`
	ListID    string    `json:"list_id"    db:"list_id"`
	Title     string    `json:"title"      db:"title"`
	Done      bool      `json:"done"       db:"done"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func HandleCreateTask(db *pgxpool.Pool) http.HandlerFunc {
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
		cmd := `INSERT INTO tasks (id, list_id, title, done, created_at) VALUES ($1, $2, $3, $4, $5)`
		if _, err := db.Exec(req.Context(), cmd, taskID, listID, reqData.Title, false, time.Now()); err != nil {
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			slog.Error("Failed to create task:", slog.String("error", err.Error()))
			return
		}

		var respData struct {
			Task Task `json:"task"`
		}
		respData.Task.ID = taskID
		respData.Task.ListID = listID
		respData.Task.Title = reqData.Title
		respData.Task.Done = false
		respData.Task.CreatedAt = time.Now()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respData); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}
	}
}

func HandleGetTasks(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		listID := req.PathValue("list_id")
		cmd := `SELECT id, list_id, title, done, created_at FROM tasks WHERE list_id = $1`
		rows, err := db.Query(req.Context(), cmd, listID)
		if err != nil {
			http.Error(w, "Failed to retrieve tasks", http.StatusInternalServerError)
			slog.Error("Failed to retrieve tasks:", slog.String("error", err.Error()))
			return
		}
		defer rows.Close()

		var tasks []Task
		for rows.Next() {
			var task Task
			if err := rows.Scan(&task.ID, &task.ListID, &task.Title, &task.Done, &task.CreatedAt); err != nil {
				http.Error(w, "Failed to scan task", http.StatusInternalServerError)
				slog.Error("Failed to scan task:", slog.String("error", err.Error()))
				return
			}
			tasks = append(tasks, task)
		}

		var respData struct {
			Tasks []Task `json:"tasks"`
		}
		respData.Tasks = tasks

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respData); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}
	}
}

func HandleUpdateTask(db *pgxpool.Pool) http.HandlerFunc {
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
		cmd := `UPDATE tasks SET title = $1, done = $2 WHERE id = $3`
		if _, err := db.Exec(req.Context(), cmd, reqData.Title, reqData.Done, taskID); err != nil {
			http.Error(w, "Failed to update task", http.StatusInternalServerError)
			slog.Error("Failed to update task:", slog.String("error", err.Error()))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func HandleDeleteTask(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		taskID := req.PathValue("task_id")
		cmd := `DELETE FROM tasks WHERE id = $1`
		if _, err := db.Exec(req.Context(), cmd, taskID); err != nil {
			http.Error(w, "Failed to delete task", http.StatusInternalServerError)
			slog.Error("Failed to delete task:", slog.String("error", err.Error()))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
