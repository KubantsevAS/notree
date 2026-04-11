package service

import (
	"context"

	"github.com/KubantsevAS/notree/backend/internal/db/sqlc"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/jackc/pgx/v5/pgtype"
)

type NodeService struct {
	db *sqlc.Queries
}

func NewNodeService(db *sqlc.Queries) *NodeService {
	return &NodeService{db: db}
}

func (s *NodeService) CreateNode(ctx context.Context, userID pgtype.UUID, req *dto.CreateNodeRequest) (sqlc.Node, error) {
	return s.db.CreateNode(ctx, sqlc.CreateNodeParams{
		UserID:    userID,
		ParentID:  *req.ParentID,
		Type:      sqlc.NodeType(req.Type),
		Title:     req.Title,
		SortOrder: 0,
	})
}
