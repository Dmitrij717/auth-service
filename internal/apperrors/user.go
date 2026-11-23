package apperrors

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("user with that email already exists")
)

var (
	ErrEmailIsEmpty = errors.New("empty email")

	ErrPasswordIsEmpty = errors.New("empty password")
	ErrPasswordTooLong = errors.New("password too long. The password must not exceed 72 bytes in UTF-8.")
)
