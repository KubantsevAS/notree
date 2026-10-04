package hierarchy

import (
	"encoding/json"
	"time"
)

type NodeResponse struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	ParentID  *string    `json:"parent_id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	SortOrder int64      `json:"sort_order"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type GetChildrenResponse []NodeResponse
type GetAncestorsResponse []NodeResponse
type GetDescendantsResponse []NodeResponse
type GetSubtreeResponse []NodeResponse

type NullableString struct {
	Value *string
	IsSet bool
}

func (n *NullableString) UnmarshalJSON(data []byte) error {
	n.IsSet = true

	if string(data) == "null" {
		n.Value = nil
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n.Value = &s
	return nil
}

type MoveNodeRequest struct {
	ParentID NullableString `json:"parent_id" swaggertype:"string" binding:"required" extensions:"x-nullable"`
	BeforeID NullableString `json:"before_id" swaggertype:"string" binding:"required" extensions:"x-nullable"`
}

type MoveNodeResponse struct {
	ParentID  *string    `json:"parent_id"`
	SortOrder int64      `json:"sort_order"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type BreadcrumbResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type BreadcrumbsResponse []BreadcrumbResponse
