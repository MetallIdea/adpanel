package migration

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

// Migrate runs all pending upward migrations using goose with embedded SQL files.
func Migrate(db *sql.DB) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	goose.SetBaseFS(migrationFS)

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("run migrate: %w", err)
	}

	return nil
}
