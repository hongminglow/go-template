package user

import (
	"context"
	"errors"
	"testing"
)

type stubRepository struct {
	createInput CreateInput
	listOpts    ListOptions

	createUser User
	listUsers  []User

	getUser User

	updateInput UpdateInput
	updateUser  User

	deleteID int64
}

func (s *stubRepository) Create(_ context.Context, input CreateInput) (User, error) {
	s.createInput = input
	return s.createUser, nil
}

func (s *stubRepository) List(_ context.Context, opts ListOptions) ([]User, error) {
	s.listOpts = opts
	return s.listUsers, nil
}

func (s *stubRepository) GetByID(_ context.Context, _ int64) (User, error) {
	return s.getUser, nil
}

func (s *stubRepository) Update(_ context.Context, _ int64, input UpdateInput) (User, error) {
	s.updateInput = input
	return s.updateUser, nil
}

func (s *stubRepository) Delete(_ context.Context, id int64) error {
	s.deleteID = id
	return nil
}

func TestServiceCreate_NormalizesInput(t *testing.T) {
	repo := &stubRepository{
		createUser: User{ID: 1},
	}
	service := NewService(repo)

	_, err := service.Create(context.Background(), CreateInput{
		Name:  "  Alice  ",
		Email: "  ALICE@example.com ",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if repo.createInput.Name != "Alice" {
		t.Fatalf("expected normalized name 'Alice', got '%s'", repo.createInput.Name)
	}

	if repo.createInput.Email != "alice@example.com" {
		t.Fatalf("expected normalized email 'alice@example.com', got '%s'", repo.createInput.Email)
	}
}

func TestServiceCreate_InvalidEmail(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)

	_, err := service.Create(context.Background(), CreateInput{
		Name:  "Alice",
		Email: "not-an-email",
	})
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestServiceList_SanitizesPagination(t *testing.T) {
	repo := &stubRepository{}
	service := NewService(repo)

	_, err := service.List(context.Background(), 999, -5)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if repo.listOpts.Limit != maxListLimit {
		t.Fatalf("expected limit %d, got %d", maxListLimit, repo.listOpts.Limit)
	}

	if repo.listOpts.Offset != 0 {
		t.Fatalf("expected offset 0, got %d", repo.listOpts.Offset)
	}
}
