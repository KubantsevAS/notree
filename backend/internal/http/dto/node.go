package dto

import (
	"time"
)

type CreateNodeRequest struct {
	ParentID *string `json:"parent_id"`
	Type     string  `json:"type" validate:"required,oneof=folder note task"`
	Title    string  `json:"title" validate:"required,max=255"`
}

type CreateNodeResponse struct {
	ID        string     `json:"id"`
	ParentID  string     `json:"parent_id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	SortOrder int32      `json:"sort_order"`
	CreatedAt *time.Time `json:"created_at"`
}

type UpdateNodeRequest struct {
	ParentID  *string `json:"parent_id" validate:"omitempty"`
	Type      *string `json:"type" validate:"omitempty,oneof=folder note task"`
	Title     *string `json:"title" validate:"omitempty,max=255"`
	SortOrder *int32  `json:"sort_order" validate:"omitempty"`
}

type UpdateNodeResponse struct {
	ParentID  string     `json:"parent_id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	SortOrder int32      `json:"sort_order"`
	UpdatedAt *time.Time `json:"updated_at"`
}
