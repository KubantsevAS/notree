package hierarchy

import "errors"

var (
	ErrParentNotFound = errors.New("parent_id references on nonexistent node")
	ErrNodeNotFound   = errors.New("node not found")
	ErrNodeIsRoot     = errors.New("node is root and has no parent")
)
