package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/app"
	"github.com/jackc/pgx/v5"
)

type List struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func HandleCreateList(a *app.App) http.HandlerFunc {
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

		list, err := a.CreateList(req.Context(), reqData.Name)
		if err != nil {
			http.Error(w, "Failed to create list", http.StatusInternalServerError)
			slog.Error("Failed to create list:", slog.String("error", err.Error()))
			return
		}

		var respData List
		respData.ID = list.ID
		respData.Name = list.Name

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respData); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}


	}
}

func HandleGetList(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		listID := req.PathValue("list_id")

		list, err := a.GetList(req.Context(), listID)
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
		respData.List.ID = list.ID
		respData.List.Name = list.Name

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respData); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}
	}
}

func HandleGetLists(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		lists, err := a.GetLists(req.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve lists", http.StatusInternalServerError)
			slog.Error("Failed to retrieve lists:", slog.String("error", err.Error()))
			return
		}

		var listsJson []List

		for _, l := range lists {
			listsJson = append(listsJson, List{
				ID:   l.ID,
				Name: l.Name,
			})
		}

		var respData struct {
			Lists []List `json:"lists"`
		}
		respData.Lists = listsJson

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(respData); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			slog.Error("Failed to encode response:", slog.String("error", err.Error()))
			return
		}
	}
}
