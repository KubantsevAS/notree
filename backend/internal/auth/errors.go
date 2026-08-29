package auth

import "errors"

var (
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
	ErrInvalidRefreshToken      = errors.New("invalid or expired refresh token")
	ErrInvalidResetToken        = errors.New("invalid or expired reset token")
	ErrWrongCredentials         = errors.New("invalid credentials")

	ErrUserExist    = errors.New("user with that email already exist")
	ErrUserNotFound = errors.New("user not found")
)
