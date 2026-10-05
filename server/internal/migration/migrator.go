package migration

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const createMigrationsTableSQL = `
CREATE TABLE IF NOT EXISTS migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

type Migrator struct {
	db         *sql.DB
	migrations map[string]string
}

func NewMigrator(db *sql.DB, migrationsDir string) (*Migrator, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	m := &Migrator{
		db:         db,
		migrations: make(map[string]string),
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		path := filepath.Join(migrationsDir, entry.Name())
		m.migrations[entry.Name()] = path
	}

	return m, nil
}

func (m *Migrator) Init() error {
	_, err := m.db.Exec(createMigrationsTableSQL)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}
	return nil
}

func (m *Migrator) Run() error {
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Сортируем миграции по имени
	names := make([]string, 0, len(m.migrations))
	for name := range m.migrations {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if applied[name] {
			fmt.Printf("Migration %s already applied, skipping\n", name)
			continue
		}

		path := m.migrations[name]
		sql, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := m.db.Begin()
		if err != nil {
			return fmt.Errorf("begin transaction for %s: %w", name, err)
		}

		if _, err = tx.Exec(string(sql)); err != nil {
			tx.Rollback()
			return fmt.Errorf("exec migration %s: %w", name, err)
		}

		_, err = tx.Exec(
			"INSERT INTO migrations(name, applied_at) VALUES(?, ?)",
			name,
			time.Now().Format("2006-01-02 15:04:05"),
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}

		fmt.Printf("Migration %s applied successfully\n", name)
	}

	return nil
}

func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	rows, err := m.db.Query("SELECT name FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}

	return applied, nil
}
