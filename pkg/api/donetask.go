package api

import (
	"go1f/pkg/db"
	"go1f/pkg/nextdate"
	"net/http"
)

func DoneTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeJSON(w, map[string]string{"error": "Не указан идентификатор"})
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	if task.Repeat == "" {
		// Для разовых задач - удаляем их
		if err := db.DeleteTask(id); err != nil {
			writeJSON(w, map[string]string{"error": err.Error()})
			return
		}
	} else {
		// Для повторяющихся задач - вычисляем следующую дату и обновляем
		next := nextdate.NextDate(task.Date, task.Repeat)
		if err := db.UpdateDate(next, id); err != nil {
			writeJSON(w, map[string]string{"error": err.Error()})
			return
		}
	}

	writeJSON(w, map[string]interface{}{})
}
