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
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{"error": fmt.Sprintf("error reading JSON: %v", err)})
		return
	}

	if task.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{"error": "task title is required"})
		return
	}

	if err := checkDate(&task); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, map[string]string{"error": err.Error()})
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, map[string]string{"error": fmt.Sprintf("error adding task: %v", err)})
		return
	}

	writeJSON(w, map[string]string{"id": fmt.Sprintf("%d", id)})
}

func checkDate(task *db.Task) error {
	now := time.Now().UTC()
	nowDate, _ := time.Parse(dateFormat, now.Format(dateFormat))

	if task.Date == "" {
		task.Date = now.Format(dateFormat)
	}

	t, err := time.Parse(dateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("invalid date format")
	}

	if task.Repeat != "" {
		next, err := NextDate(nowDate, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("invalid repeat rule")
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
