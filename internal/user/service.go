package user

import (
	"context"
	"net/mail"
	"strings"
)

const (
	defaultListLimit int32 = 20
	maxListLimit     int32 = 100
	maxNameLength          = 120
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (User, error) {
	normalizedInput, err := normalizeCreateInput(input)
	if err != nil {
		return User{}, err
	}

	return s.repo.Create(ctx, normalizedInput)
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]User, error) {
	opts := ListOptions{
		Limit:  sanitizeLimit(limit),
		Offset: sanitizeOffset(offset),
	}

	return s.repo.List(ctx, opts)
}

func (s *Service) GetByID(ctx context.Context, id int64) (User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (User, error) {
	normalizedInput, err := normalizeUpdateInput(input)
	if err != nil {
		return User{}, err
	}

	return s.repo.Update(ctx, id, normalizedInput)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func normalizeCreateInput(input CreateInput) (CreateInput, error) {
	name, err := normalizeName(input.Name)
	if err != nil {
		return CreateInput{}, err
	}

	email, err := normalizeEmail(input.Email)
	if err != nil {
		return CreateInput{}, err
	}

	return CreateInput{
		Name:  name,
		Email: email,
	}, nil
}

func normalizeUpdateInput(input UpdateInput) (UpdateInput, error) {
	name, err := normalizeName(input.Name)
	if err != nil {
		return UpdateInput{}, err
	}

	email, err := normalizeEmail(input.Email)
	if err != nil {
		return UpdateInput{}, err
	}

	return UpdateInput{
		Name:  name,
		Email: email,
	}, nil
}

func sanitizeLimit(limit int32) int32 {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func sanitizeOffset(offset int32) int32 {
	if offset < 0 {
		return 0
	}
	return offset
}

func normalizeName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" || len(name) > maxNameLength {
		return "", ErrInvalidName
	}
	return name, nil
}

func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", ErrInvalidEmail
	}

	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", ErrInvalidEmail
	}

	return email, nil
}
