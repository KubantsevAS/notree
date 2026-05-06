package service

import "errors"

var (
	ErrParentNotFound           = errors.New("parent_id references on nonexistent node")
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
	ErrInvalidRefreshToken      = errors.New("invalid or expired refresh token")
	ErrInvalidResetToken        = errors.New("invalid or expired reset token")
	ErrUserExist                = errors.New("user with that email already exist")
	ErrInvalidParentID          = errors.New("invalid parent_id UUID")
	ErrWrongCredentials         = errors.New("invalid credentials")
	ErrUserNotFound             = errors.New("user not found")
	ErrNodeNotFoundOrNoAccess   = errors.New("node not found or access denied")
)
