package service

import (
	"context"
	"errors"

	"github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/KubantsevAS/notree/backend/internal/httputil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type NodeService struct {
	db *node.Queries
}

func NewNodeService(db *node.Queries) *NodeService {
	return &NodeService{db: db}
}

func (s *NodeService) CreateNode(ctx context.Context, userID pgtype.UUID, req *dto.CreateNodeRequest) (dto.CreateNodeResponse, error) {
	parentID := pgtype.UUID{Valid: false}
	if req.ParentID != nil && *req.ParentID != "" {
		parsedID, err := httputil.PgUUIDFromString(req.ParentID)
		if err != nil {
			return dto.CreateNodeResponse{}, ErrInvalidParentID
		}

		if _, err := s.db.GetNodeByID(ctx, parsedID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return dto.CreateNodeResponse{}, ErrParentNotFound
			}
			return dto.CreateNodeResponse{}, err
		}

		parentID = parsedID
	}

	dbParams := node.CreateNodeParams{
		UserID:    userID,
		ParentID:  parentID,
		Type:      node.NodeType(req.Type),
		Title:     req.Title,
		SortOrder: 0,
	}

	nodeRow, err := s.db.CreateNode(ctx, dbParams)
	if err != nil {
		return dto.CreateNodeResponse{}, err
	}

	response := dto.CreateNodeResponse{
		ID:        nodeRow.ID,
		ParentID:  &nodeRow.ParentID,
		Type:      string(nodeRow.Type),
		Title:     nodeRow.Title,
		SortOrder: nodeRow.SortOrder,
		CreatedAt: nodeRow.CreatedAt,
	}

	return response, nil
}

func (s *NodeService) DeleteNode(ctx context.Context, nodeId string, userID pgtype.UUID) error {
	parsedNodeId, err := httputil.PgUUIDFromString(&nodeId)
	if err != nil {
		return err
	}

	dbParams := &node.SoftDeleteNodeCascadeParams{
		ID:     parsedNodeId,
		UserID: userID,
	}

	deletedIds, err := s.db.SoftDeleteNodeCascade(ctx, *dbParams)
	if err != nil {
		return err
	}

	if len(deletedIds) == 0 {
		return ErrNodeNotFoundOrNoAccess
	}

	return nil
}
