package user

import "time"

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInput struct {
	Name  string
	Email string
}

type UpdateInput struct {
	Name  string
	Email string
}

type ListOptions struct {
	Limit  int32
	Offset int32
}
