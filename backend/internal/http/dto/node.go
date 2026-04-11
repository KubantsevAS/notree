package dto

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateNodeRequest struct {
	ParentID *pgtype.UUID `json:"parent_id"`
	Type     string       `json:"type" validate:"required,oneof=folder note task"`
	Title    string       `json:"title" validate:"required,max=255"`
}

type CreateNodeResponse struct {
	ID        pgtype.UUID        `json:"id"`
	ParentID  *pgtype.UUID       `json:"parent_id"`
	Type      string             `json:"type"`
	Title     string             `json:"title"`
	SortOrder int32              `json:"sort_order"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}
