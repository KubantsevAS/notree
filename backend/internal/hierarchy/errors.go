package hierarchy

import "errors"

var (
	ErrParentNotFound = errors.New("parent_id references on nonexistent node")
)
