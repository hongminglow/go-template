package postgres

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
		INSERT INTO users (username, name, email, password, gender)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email)
		DO UPDATE SET
			name = EXCLUDED.name,
			username = EXCLUDED.username,
			gender = EXCLUDED.gender,
			updated_at = NOW()
	`

	// "password" hashed with bcrypt
	defaultPassword := "$2a$10$T1K7.lD22W.k4/98S.Z.I.C76PjQYxK/b1YpToZ/L5l1e/bZ2l63G"
	if _, err := db.Exec(ctx, seedAdminUserQuery, "admin", adminName, adminEmail, defaultPassword, "unspecified"); err != nil {
		return fmt.Errorf("seed default admin user: %w", err)
	}

	return nil
}
