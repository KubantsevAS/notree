package service

import (
	"context"
	"errors"
	"time"

	"github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/KubantsevAS/notree/backend/internal/httputil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type NodeService struct {
	db nodeRepository
}

type nodeRepository interface {
	CreateNode(context.Context, node.CreateNodeParams) (node.Node, error)
	GetChildren(context.Context, node.GetChildrenParams) ([]node.Node, error)
	GetNodeAncestors(context.Context, pgtype.UUID) ([]pgtype.UUID, error)
	GetNodeByID(context.Context, node.GetNodeByIDParams) (node.Node, error)
	MoveNode(context.Context, node.MoveNodeParams) (node.MoveNodeRow, error)
	SoftDeleteNodeCascade(context.Context, node.SoftDeleteNodeCascadeParams) ([]pgtype.UUID, error)
	UpdateNode(context.Context, node.UpdateNodeParams) (node.UpdateNodeRow, error)
}

func NewNodeService(db nodeRepository) *NodeService {
	return &NodeService{db: db}
}

func (s *NodeService) CreateNode(ctx context.Context, userID pgtype.UUID, req *dto.CreateNodeRequest) (dto.CreateNodeResponse, error) {
	parentID := pgtype.UUID{Valid: false}
	if req.ParentID != nil && *req.ParentID != "" {
		parsedID, err := httputil.PgUUIDFromString(req.ParentID)
		if err != nil {
			return dto.CreateNodeResponse{}, ErrInvalidParentID
		}

		if _, err := s.db.GetNodeByID(ctx, node.GetNodeByIDParams{ID: parsedID, UserID: userID}); err != nil {
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
		SortOrder: time.Now().UnixNano(),
	}

	nodeRow, err := s.db.CreateNode(ctx, dbParams)
	if err != nil {
		return dto.CreateNodeResponse{}, err
	}

	response := dto.CreateNodeResponse{
		ID:        nodeRow.ID.String(),
		ParentID:  nodeRow.ParentID.String(),
		Type:      string(nodeRow.Type),
		Title:     nodeRow.Title,
		SortOrder: nodeRow.SortOrder,
		CreatedAt: &nodeRow.CreatedAt.Time,
	}

	return response, nil
}

func (s *NodeService) DeleteNode(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID) error {
	dbParams := &node.SoftDeleteNodeCascadeParams{
		ID:     nodeID,
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

func (s *NodeService) UpdateNode(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID, req *dto.UpdateNodeRequest) (dto.UpdateNodeResponse, error) {
	if req.Type == nil && req.Title == nil {
		return dto.UpdateNodeResponse{}, ErrEmptyUpdate
	}

	dbParams := &node.UpdateNodeParams{
		ID:     nodeID,
		UserID: userID,
	}
	if req.Type != nil {
		dbParams.Type = node.NullNodeType{NodeType: node.NodeType(*req.Type), Valid: true}
	}
	if req.Title != nil {
		dbParams.Title = httputil.PgTextFromString(req.Title)
	}

	dbRow, err := s.db.UpdateNode(ctx, *dbParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.UpdateNodeResponse{}, ErrNodeNotFoundOrNoAccess
		}
		return dto.UpdateNodeResponse{}, err
	}

	response := dto.UpdateNodeResponse{
		Type:      string(dbRow.Type),
		Title:     dbRow.Title,
		UpdatedAt: &dbRow.UpdatedAt.Time,
	}

	return response, nil
}

func (s *NodeService) MoveNode(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID, req *dto.MoveNodeRequest) (dto.MoveNodeResponse, error) {
	if !req.ParentID.IsSet && req.SortOrder == nil {
		return dto.MoveNodeResponse{}, ErrEmptyUpdate
	}

	dbParams := &node.MoveNodeParams{
		ID:           nodeID,
		UserID:       userID,
		UpdateParent: pgtype.Bool{Bool: req.ParentID.IsSet, Valid: true},
		ParentID:     pgtype.UUID{Valid: false},
	}

	if req.ParentID.IsSet && req.ParentID.Value != nil {
		if *req.ParentID.Value == "" {
			return dto.MoveNodeResponse{}, ErrInvalidParentID
		}

		parsedID, err := httputil.PgUUIDFromString(req.ParentID.Value)
		if err != nil {
			return dto.MoveNodeResponse{}, ErrInvalidParentID
		}
		dbParams.ParentID = parsedID

		if parsedID == nodeID {
			return dto.MoveNodeResponse{}, ErrNodeCannotBeADescendantOfItself
		}

		if _, err := s.db.GetNodeByID(ctx, node.GetNodeByIDParams{ID: parsedID, UserID: userID}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return dto.MoveNodeResponse{}, ErrParentNotFound
			}
			return dto.MoveNodeResponse{}, err
		}

		isDescendantOfItself, err := s.isNodeDescendantOfItself(ctx, nodeID, parsedID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return dto.MoveNodeResponse{}, ErrParentNotFound
			}
			return dto.MoveNodeResponse{}, ErrInvalidParentID
		}
		if isDescendantOfItself {
			return dto.MoveNodeResponse{}, ErrNodeCannotBeADescendantOfItself
		}
	}

	if req.SortOrder != nil {
		dbParams.SortOrder = pgtype.Int8{Int64: *req.SortOrder, Valid: true}
	}

	dbRow, err := s.db.MoveNode(ctx, *dbParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.MoveNodeResponse{}, ErrNodeNotFoundOrNoAccess
		}
		return dto.MoveNodeResponse{}, err
	}

	var responseParentID *string
	if dbRow.ParentID.Valid {
		str := dbRow.ParentID.String()
		responseParentID = &str
	}

	response := dto.MoveNodeResponse{
		ParentID:  responseParentID,
		SortOrder: dbRow.SortOrder,
		UpdatedAt: &dbRow.UpdatedAt.Time,
	}

	return response, nil
}

func (s *NodeService) isNodeDescendantOfItself(ctx context.Context, nodeID pgtype.UUID, parentID pgtype.UUID) (bool, error) {
	ancestors, err := s.db.GetNodeAncestors(ctx, parentID)
	if err != nil {
		return false, err
	}

	for _, id := range ancestors {
		if id == nodeID {
			return true, nil
		}
	}

	return false, nil
}

func (s *NodeService) GetChildren(ctx context.Context, parentID pgtype.UUID, userID pgtype.UUID) ([]dto.NodeResponse, error) {
	if _, err := s.db.GetNodeByID(ctx, node.GetNodeByIDParams{ID: parentID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []dto.NodeResponse{}, ErrParentNotFound
		}
		return []dto.NodeResponse{}, err
	}

	dbParams := &node.GetChildrenParams{
		ParentID: parentID,
		UserID:   userID,
	}

	dbRow, err := s.db.GetChildren(ctx, *dbParams)
	if err != nil {
		return []dto.NodeResponse{}, err
	}

	response := make([]dto.NodeResponse, 0, len(dbRow))
	for _, n := range dbRow {
		response = append(response, dto.MapNodeToResponse(n))
	}

	return response, nil
}
