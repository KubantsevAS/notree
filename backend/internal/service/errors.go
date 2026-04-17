package service

import "errors"

var (
	ErrParentNotFound      = errors.New("parent_id references on nonexistent node")
	ErrInvalidParentID     = errors.New("Invalid parent_id UUID")
	ErrUserExist           = errors.New("User with that email already exist")
	ErrWrongCredentials    = errors.New("Invalid Credentials")
	ErrInvalidRefreshToken = errors.New("Invalid or expired refresh token")
)
