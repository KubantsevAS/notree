package service

import (
	"context"
	"errors"

	"github.com/KubantsevAS/notree/backend/internal/db/sqlc"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type NodeService struct {
	db *sqlc.Queries
}

func NewNodeService(db *sqlc.Queries) *NodeService {
	return &NodeService{db: db}
}

func (s *NodeService) CreateNode(ctx context.Context, userID pgtype.UUID, req *dto.CreateNodeRequest) (sqlc.Node, error) {
	parentID := pgtype.UUID{Valid: false}
	if req.ParentID != nil && *req.ParentID != "" {
		parsedID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return sqlc.Node{}, ErrInvalidParentID
		}
		parentID = pgtype.UUID{Bytes: parsedID, Valid: true}

		if _, err := s.db.GetNodeByID(ctx, parentID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return sqlc.Node{}, ErrParentNotFound
			}
			return sqlc.Node{}, err
		}
	}

	return s.db.CreateNode(ctx, sqlc.CreateNodeParams{
		UserID:    userID,
		ParentID:  parentID,
		Type:      sqlc.NodeType(req.Type),
		Title:     req.Title,
		SortOrder: 0,
	})
}
