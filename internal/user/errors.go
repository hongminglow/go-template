package user

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailAlreadyUsed    = errors.New("email already exists")
	ErrUsernameAlreadyUsed = errors.New("username already exists")
	ErrInvalidName         = errors.New("name is required")
	ErrInvalidEmail        = errors.New("valid email is required")
	ErrInvalidPassword     = errors.New("password must be at least 6 characters")
	ErrInvalidCredentials  = errors.New("invalid email or password")
)
