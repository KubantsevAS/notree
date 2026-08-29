package node

import "errors"

var (
	ErrParentNotFound                  = errors.New("parent_id references on nonexistent node")
	ErrNodeCannotBeADescendantOfItself = errors.New("node cannot be a descendant of itself")
	ErrNodeNotFoundOrNoAccess          = errors.New("node not found or access denied")
	ErrInvalidParentID                 = errors.New("invalid parent_id UUID")
)
