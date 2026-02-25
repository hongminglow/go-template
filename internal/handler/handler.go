package handler

import (
	"github.com/hongminglow/go-template/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB     *pgxpool.Pool
	Config config.Config
}

func New(db *pgxpool.Pool, cfg config.Config) *Handler {
	return &Handler{
		DB:     db,
		Config: cfg,
	}
}
