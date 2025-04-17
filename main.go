package main

import (
	"go1f/pkg/db"
	"go1f/pkg/server"
	"log"
)

func main() {
	if err := db.Init(); err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer db.DB.Close()

	db.SetDB(db.DB)

	server.StartServer()
}
