package handler

import (
	"net/http"
	"log/slog"
	"github.com/jackc/pgx/v5"
	"github.com/google/uuid"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type List struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func HandleCreateList(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		var reqData struct {
			Name string `json:"name"`
		}

		if err := json.NewDecoder(req.Body).Decode(&reqData); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		cmd := `INSERT INTO lists (id, name) VALUES ($1, $2)`
		listID := uuid.New().String()
		if _, err := db.Exec(req.Context(), cmd, listID, reqData.Name); err != nil {
			http.Error(w, "Failed to create list", http.StatusInternalServerError)
			slog.Error("Failed to create list:", slog.String("error", err.Error()))
			return
		}

		var respData struct {
			List List `json:"list"`
		}
		respData.List.ID = listID
		respData.List.Name = reqData.Name

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respData); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}

		// Здесь вы можете добавить логику для обработки создания списка
		fmt.Fprintln(w, "Создание списка успешно")
	}
}

func HandleGetList(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		listID := req.PathValue("list_id")
		var listName string
		cmd := `SELECT name FROM lists WHERE id = $1`
		err := db.QueryRow(req.Context(), cmd, listID).Scan(&listName)
		if err != nil {
			if err == pgx.ErrNoRows {
				http.Error(w, "List not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to retrieve list", http.StatusInternalServerError)
				slog.Error("Failed to retrieve list:", slog.String("error", err.Error()))
			}
			return
		}

		var respData struct {
			List List `json:"list"`
		}
		respData.List.ID = listID
		respData.List.Name = listName

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respData); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}
	}
}

func HandleGetLists(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cmd := `SELECT id, name FROM lists`
		rows, err := db.Query(req.Context(), cmd)
		if err != nil {
			http.Error(w, "Failed to retrieve lists", http.StatusInternalServerError)
			slog.Error("Failed to retrieve lists:", slog.String("error", err.Error()))
			return
		}
		defer rows.Close()

		var lists []List

		for rows.Next() {
			var list List
			if err := rows.Scan(&list.ID, &list.Name); err != nil {
				http.Error(w, "Failed to scan list", http.StatusInternalServerError)
				slog.Error("Failed to scan list:", slog.String("error", err.Error()))
				return
			}
			lists = append(lists, list)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Error iterating over lists", http.StatusInternalServerError)
			slog.Error("Error iterating over lists:", slog.String("error", err.Error()))
			return
		}

		var respData struct {
			Lists []List `json:"lists"`
		}
		respData.Lists = lists

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respData); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}
	}
}
