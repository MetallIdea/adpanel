package db

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/joho/godotenv"
)

var DB *sql.DB

func Init() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "adpanel.db"
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err = conn.Ping(); err != nil {
		return err
	}

	DB = conn
	return nil
}
