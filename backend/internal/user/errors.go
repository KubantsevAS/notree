package user

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")

	ErrWrongCredentials         = errors.New("invalid credentials")
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
)
