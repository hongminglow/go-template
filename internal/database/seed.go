package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SeedConfig struct {
	AdminName  string
	AdminEmail string
}

func SeedDefaultData(ctx context.Context, db *pgxpool.Pool, cfg SeedConfig) error {
	adminName := strings.TrimSpace(cfg.AdminName)
	if adminName == "" {
		adminName = "Admin"
	}

	adminEmail := strings.ToLower(strings.TrimSpace(cfg.AdminEmail))
	if adminEmail == "" {
		adminEmail = "admin@email.com"
	}

	const seedAdminUserQuery = `
		INSERT INTO users (name, email)
		VALUES ($1, $2)
		ON CONFLICT (email)
		DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = NOW()
	`

	if _, err := db.Exec(ctx, seedAdminUserQuery, adminName, adminEmail); err != nil {
		return fmt.Errorf("seed default admin user: %w", err)
	}

	return nil
}
