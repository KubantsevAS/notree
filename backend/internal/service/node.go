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
		CreatedAt: &nodeRow.CreatedAt.Time,
	}

	return response, nil
}

func (s *NodeService) DeleteNode(ctx context.Context, nodeId pgtype.UUID, userID pgtype.UUID) error {
	dbParams := &node.SoftDeleteNodeCascadeParams{
		ID:     nodeId,
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

func (s *NodeService) UpdateNode(ctx context.Context, nodeId pgtype.UUID, userID pgtype.UUID, req *dto.UpdateNodeRequest) (dto.UpdateNodeResponse, error) {
	parentID := pgtype.UUID{Valid: false}
	if req.ParentID != nil && *req.ParentID != "" {
		parsedID, err := httputil.PgUUIDFromString(req.ParentID)
		if err != nil {
			return dto.UpdateNodeResponse{}, ErrInvalidParentID
		}

		if _, err := s.db.GetNodeByID(ctx, parsedID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return dto.UpdateNodeResponse{}, ErrParentNotFound
			}
			return dto.UpdateNodeResponse{}, err
		}

		parentID = parsedID

		isDescendantOfItself, err := s.isNodeDescendantOfItself(ctx, nodeId, parentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return dto.UpdateNodeResponse{}, ErrParentNotFound
			}
			return dto.UpdateNodeResponse{}, ErrInvalidParentID
		}
		if isDescendantOfItself {
			return dto.UpdateNodeResponse{}, ErrNodeCannotBeADescendantOfItself
		}
	}

	dbParams := &node.UpdateNodeParams{
		ParentID: parentID,
		ID:       nodeId,
	}
	if req.SortOrder != nil {
		dbParams.SortOrder = pgtype.Int4{Int32: *req.SortOrder, Valid: true}
	}
	if req.Type != nil {
		dbParams.Type = node.NullNodeType{NodeType: node.NodeType(*req.Type), Valid: true}
	}
	if req.Title != nil {
		dbParams.Title = httputil.PgTextFromString(req.Title)
	}

	dbRow, err := s.db.UpdateNode(ctx, *dbParams)
	if err != nil {
		return dto.UpdateNodeResponse{}, err
	}

	response := dto.UpdateNodeResponse{
		ParentID:  &dbRow.ParentID,
		Type:      string(dbRow.Type),
		Title:     dbRow.Title,
		SortOrder: dbRow.SortOrder,
		UpdatedAt: &dbRow.UpdatedAt.Time,
	}

	return response, nil
}

func (s *NodeService) isNodeDescendantOfItself(ctx context.Context, nodeId pgtype.UUID, parentId pgtype.UUID) (bool, error) {
	ancestors, err := s.db.GetNodeAncestors(ctx, parentId)
	if err != nil {
		return false, err
	}

	for _, id := range ancestors {
		if id == nodeId {
			return true, nil
		}
	}

	return false, nil
}
