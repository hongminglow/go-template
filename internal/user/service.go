package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultListLimit int32 = 20
	maxListLimit     int32 = 100
	maxNameLength          = 120
)

type Service struct {
	repo  Repository
	cache *redis.Client
}

func NewService(repo Repository, cache *redis.Client) *Service {
	return &Service{
		repo:  repo,
		cache: cache,
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (User, error) {
	normalizedInput, err := normalizeCreateInput(input)
	if err != nil {
		return User{}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(normalizedInput.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hash password failed: %w", err)
	}
	normalizedInput.Password = string(hashedPassword)

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

func (s *Service) Auth(ctx context.Context, email, password string) (User, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	u, err := s.repo.GetByEmail(ctx, normalizedEmail)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	return u, nil
}

func (s *Service) GetProfile(ctx context.Context, id int64) (User, error) {
	cacheKey := fmt.Sprintf("user:profile:%d", id)

	// Try reading from Redis first
	cachedData, err := s.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var u User
		if err := json.Unmarshal([]byte(cachedData), &u); err == nil {
			return u, nil
		}
	}

	// Fallback to database
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return User{}, err
	}

	// Save back to Redis with 5 minutes TTL
	if bytes, err := json.Marshal(u); err == nil {
		_ = s.cache.Set(ctx, cacheKey, bytes, 5*time.Minute).Err()
	}

	return u, nil
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

	username := strings.TrimSpace(input.Username)
	if username == "" {
		return CreateInput{}, ErrInvalidName
	}

	gender := strings.TrimSpace(input.Gender)
	if gender == "" {
		gender = "unspecified"
	}

	if len(input.Password) < 6 {
		return CreateInput{}, ErrInvalidPassword
	}

	return CreateInput{
		Username: username,
		Name:     name,
		Email:    email,
		Password: input.Password,
		Gender:   gender,
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
