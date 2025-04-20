package main

import (
	"go1f/pkg/db"
	"go1f/pkg/server"
	"log"
)

func main() {
	if err := db.Init(); err != nil {
		log.Fatalf("error initializing database: %v", err)
	}
	defer db.DB.Close()

	db.SetDB(db.DB)

	server.StartServer()
}
