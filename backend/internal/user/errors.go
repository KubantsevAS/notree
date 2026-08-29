package user

import "errors"

var (
	ErrUserExist    = errors.New("user with that email already exist")
	ErrUserNotFound = errors.New("user not found")

	ErrWrongCredentials         = errors.New("invalid credentials")
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
)
