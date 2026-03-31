package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, input CreateInput) (User, error)
	List(ctx context.Context, opts ListOptions) ([]User, error)
	GetByID(ctx context.Context, id int64) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	Update(ctx context.Context, id int64, input UpdateInput) (User, error)
	Delete(ctx context.Context, id int64) error
}

type PGRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *PGRepository {
	return &PGRepository{db: db}
}

func (r *PGRepository) Create(ctx context.Context, input CreateInput) (User, error) {
	const query = `
		INSERT INTO users (username, name, email, password, gender)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, name, email, password, gender, created_at, updated_at
	`

	var u User
	err := r.db.QueryRow(ctx, query, input.Username, input.Name, input.Email, input.Password, input.Gender).Scan(
		&u.ID,
		&u.Username,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.Gender,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if known := mapRepositoryError(err); known != nil {
			return User{}, known
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return u, nil
}

func (r *PGRepository) List(ctx context.Context, opts ListOptions) ([]User, error) {
	const query = `
		SELECT id, username, name, email, password, gender, created_at, updated_at
		FROM users
		ORDER BY id ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, opts.Limit, opts.Offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Name,
			&u.Email,
			&u.Password,
			&u.Gender,
			&u.CreatedAt,
			&u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user rows: %w", err)
	}

	return users, nil
}

func (r *PGRepository) GetByID(ctx context.Context, id int64) (User, error) {
	const query = `
		SELECT id, username, name, email, password, gender, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var u User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Username,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.Gender,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if known := mapRepositoryError(err); known != nil {
			return User{}, known
		}
		return User{}, fmt.Errorf("get user by id: %w", err)
	}

	return u, nil
}

func (r *PGRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	const query = `
		SELECT id, username, name, email, password, gender, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var u User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Username,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.Gender,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if known := mapRepositoryError(err); known != nil {
			return User{}, known
		}
		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	return u, nil
}

func (r *PGRepository) Update(ctx context.Context, id int64, input UpdateInput) (User, error) {
	const query = `
		UPDATE users
		SET name = $2, email = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, username, name, email, password, gender, created_at, updated_at
	`

	var u User
	err := r.db.QueryRow(ctx, query, id, input.Name, input.Email).Scan(
		&u.ID,
		&u.Username,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.Gender,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if known := mapRepositoryError(err); known != nil {
			return User{}, known
		}
		return User{}, fmt.Errorf("update user: %w", err)
	}

	return u, nil
}

func (r *PGRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM users WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

func mapRepositoryError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrUserNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return ErrUsernameAlreadyUsed
			}
			return ErrEmailAlreadyUsed
		}
	}

	return nil
}
