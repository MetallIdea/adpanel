package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
)

var DB *pgx.Conn

func Init() error {
	connStr :=
		"postgresql://" +
			os.Getenv("POSTGRES_USER") + ":" +
			os.Getenv("POSTGRES_PASSWORD") + "@" +
			os.Getenv("POSTGRES_HOST") + ":" +
			os.Getenv("POSTGRES_PORT") + "/" +
			os.Getenv("POSTGRES_DB") +
			"?sslmode=disable"

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return err
	}

	DB = conn
	return nil
}
