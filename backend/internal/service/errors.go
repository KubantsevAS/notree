package service

import "errors"

var (
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
	ErrInvalidRefreshToken      = errors.New("invalid or expired refresh token")
	ErrInvalidResetToken        = errors.New("invalid or expired reset token")
	ErrWrongCredentials         = errors.New("invalid credentials")

	ErrUserExist    = errors.New("user with that email already exist")
	ErrUserNotFound = errors.New("user not found")

	ErrParentNotFound                  = errors.New("parent_id references on nonexistent node")
	ErrNodeCannotBeADescendantOfItself = errors.New("node cannot be a descendant of itself")
	ErrNodeNotFoundOrNoAccess          = errors.New("node not found or access denied")
	ErrInvalidParentID                 = errors.New("invalid parent_id UUID")

	ErrEmptyUpdate = errors.New("no fields provided for update")
)
