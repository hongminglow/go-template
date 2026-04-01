package postgres

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func EnsureSchema(dsn string) error {
	slog.Info("running database migrations...")

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get working directory: %w", err)
	}
	
	migrationPath := filepath.Join(cwd, "migrations")
	m, err := migrate.New("file://"+migrationPath, dsn)
	if err != nil {
		return fmt.Errorf("init migration failed: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration up failed: %w", err)
	}

	slog.Info("database migrations completed successfully")
	return nil
}
