package domain

import "errors"

var (
	ErrEmptyUpdate = errors.New("no fields provided for update")
)
