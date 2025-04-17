package server

import (
	"TaskManager/pkg/api"
	"log"
	"net/http"
	"os"
)

const (
	DefaultPort = "7540"
	WebDir      = "./web"
	DbFile      = "scheduler.db"
)

func StartServer() {
	api.Init()

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = DefaultPort
	}

	http.Handle("/", http.FileServer(http.Dir(WebDir)))
	http.HandleFunc("/api/signin", api.SignInHandler)
	http.HandleFunc("/api/task", api.Auth(api.TaskHandler))
	http.HandleFunc("/api/task/done", api.Auth(api.DoneTaskHandler))
	http.HandleFunc("/api/tasks", api.Auth(api.TasksHandler))

	log.Printf("Сервер запущен на http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
