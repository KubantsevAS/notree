package hierarchy

import "time"

type NodeResponse struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	ParentID  string     `json:"parent_id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	SortOrder int64      `json:"sort_order"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type GetChildrenResponse []NodeResponse
