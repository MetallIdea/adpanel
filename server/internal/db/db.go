package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var DB *pgx.Conn

func Init() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	connStr := os.Getenv("POSTGRES_URL")

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return err
	}

	DB = conn
	return nil
}
