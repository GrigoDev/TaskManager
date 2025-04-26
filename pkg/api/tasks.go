package api

import (
	"encoding/json"
	"net/http"

	"go1f/pkg/db"
)

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

type ErrorResp struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	writeJSON(w, ErrorResp{Error: err.Error()})
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Получаем параметр поиска
	search := r.URL.Query().Get("search")

	// Получаем задачи с разумным ограничением
	tasks, err := db.Tasks(50, search)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, TasksResp{Tasks: tasks})
}
