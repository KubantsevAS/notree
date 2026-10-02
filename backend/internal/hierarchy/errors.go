package hierarchy

import "errors"

var (
	ErrParentNotFound = errors.New("parent_id references on nonexistent node")
	ErrNodeNotFound   = errors.New("node not found")
	ErrNodeIsRoot     = errors.New("node is root and has no parent")
	ErrRootNotFound   = errors.New("root node not found")

	ErrParentIDRequired                = errors.New("parent_id is required")
	ErrInvalidParentID                 = errors.New("invalid parent_id UUID")
	ErrNodeCannotBeADescendantOfItself = errors.New("node cannot be a descendant of itself")

	ErrBeforeIDRequired = errors.New("before_id is required")
	ErrInvalidBeforeID  = errors.New("invalid before_id UUID")
	ErrBeforeNotSibling = errors.New("before_id must reference another child of parent_id")

	ErrNoRankAfterRebalance = errors.New("no free rank between siblings after rebalance")
)
