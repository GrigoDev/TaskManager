package api

import "net/http"

func Init() {
	http.HandleFunc("/api/nextdate", nextDayHandler)
}

// Экспортируем функции, связанные с аутентификацией
var (
	SignInHandler = signInHandler
	Auth          = auth
	TasksHandler  = tasksHandler
)
