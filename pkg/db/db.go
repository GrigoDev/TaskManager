package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

const schema = `
CREATE TABLE scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL CHECK(length(date) = 8),
    title VARCHAR(255) NOT NULL CHECK(title != ''),
    comment TEXT NOT NULL DEFAULT '',
    repeat VARCHAR(128) NOT NULL DEFAULT ''
);

CREATE INDEX idx_date ON scheduler(date);
`

func Init() error {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}

	if install {
		if _, err := db.Exec(schema); err != nil {
			db.Close()
			return fmt.Errorf("error creating table: %w", err)
		}
	}

	DB = db
	return nil
}
