package user

import "time"

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // never return password to client
	Gender    string    `json:"gender"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInput struct {
	Username string
	Name     string
	Email    string
	Password string
	Gender   string
}

type UpdateInput struct {
	Name  string
	Email string
}

type ListOptions struct {
	Limit  int32
	Offset int32
}
