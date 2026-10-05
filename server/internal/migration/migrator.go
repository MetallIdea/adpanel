package migration

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

// Migrate runs all pending upward migrations using goose.
func Migrate(db *sql.DB, migrationsDir string) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.RunContext(context.Background(), "up", db, migrationsDir); err != nil {
		return fmt.Errorf("run migrate: %w", err)
	}

	return nil
}
