package api

import (
	"encoding/json"
	"fmt"
	"go1f/pkg/db"
	"net/http"
	"time"
)

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, map[string]string{"error": fmt.Sprintf("Ошибка чтения JSON: %v", err)})
		return
	}

	if task.Title == "" {
		writeJson(w, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}

	if err := checkDate(&task); err != nil {
		writeJson(w, map[string]string{"error": err.Error()})
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, map[string]string{"error": fmt.Sprintf("Ошибка добавления задачи: %v", err)})
		return
	}

	writeJson(w, map[string]string{"id": fmt.Sprintf("%d", id)})
}

func writeJson(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(data)
}

func checkDate(task *db.Task) error {
	now := time.Now().UTC()
	nowDate, _ := time.Parse(dateFormat, now.Format(dateFormat))

	if task.Date == "" {
		task.Date = now.Format(dateFormat)
	}

	t, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("Неверный формат даты")
	}

	if task.Repeat != "" {
		next, err := NextDate(nowDate, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("Неверное правило повторения")
		}
		if t.Before(nowDate) {
			task.Date = next
		}
	} else {
		if t.Before(nowDate) {
			task.Date = nowDate.Format(dateFormat)
		}
	}

	return nil
}
