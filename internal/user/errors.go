package user

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailAlreadyUsed = errors.New("email already exists")
	ErrInvalidName      = errors.New("name is required")
	ErrInvalidEmail     = errors.New("valid email is required")
)
