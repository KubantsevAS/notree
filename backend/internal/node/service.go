package node

import (
	"context"
	"errors"
	"time"

	"github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/httputil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	CreateNode(context.Context, node.CreateNodeParams) (node.Node, error)
	GetChildren(context.Context, node.GetChildrenParams) ([]node.Node, error)
	GetNodeAncestors(context.Context, pgtype.UUID) ([]pgtype.UUID, error)
	GetNodeByID(context.Context, node.GetNodeByIDParams) (node.Node, error)
	MoveNode(context.Context, node.MoveNodeParams) (node.MoveNodeRow, error)
	SoftDeleteNodeCascade(context.Context, node.SoftDeleteNodeCascadeParams) ([]pgtype.UUID, error)
	UpdateNode(context.Context, node.UpdateNodeParams) (node.UpdateNodeRow, error)
}

type Service struct {
	store Store
}

func NewService(repo Store) *Service {
	return &Service{store: repo}
}

func (s *Service) CreateNode(ctx context.Context, userID pgtype.UUID, req *CreateNodeRequest) (CreateNodeResponse, error) {
	parentID := pgtype.UUID{Valid: false}
	if req.ParentID != nil && *req.ParentID != "" {
		parsedID, err := httputil.PgUUIDFromString(req.ParentID)
		if err != nil {
			return CreateNodeResponse{}, ErrInvalidParentID
		}

		if _, err := s.store.GetNodeByID(ctx, node.GetNodeByIDParams{ID: parsedID, UserID: userID}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return CreateNodeResponse{}, ErrParentNotFound
			}
			return CreateNodeResponse{}, err
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

	nodeRow, err := s.store.CreateNode(ctx, dbParams)
	if err != nil {
		return CreateNodeResponse{}, err
	}

	response := CreateNodeResponse{
		ID:        nodeRow.ID.String(),
		ParentID:  nodeRow.ParentID.String(),
		Type:      string(nodeRow.Type),
		Title:     nodeRow.Title,
		SortOrder: nodeRow.SortOrder,
		CreatedAt: &nodeRow.CreatedAt.Time,
	}

	return response, nil
}

func (s *Service) DeleteNode(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID) error {
	dbParams := &node.SoftDeleteNodeCascadeParams{
		ID:     nodeID,
		UserID: userID,
	}

	deletedIds, err := s.store.SoftDeleteNodeCascade(ctx, *dbParams)
	if err != nil {
		return err
	}

	if len(deletedIds) == 0 {
		return ErrNodeNotFoundOrNoAccess
	}

	return nil
}

func (s *Service) UpdateNode(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID, req *UpdateNodeRequest) (UpdateNodeResponse, error) {
	if req.Type == nil && req.Title == nil {
		return UpdateNodeResponse{}, ErrEmptyUpdate
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

	dbRow, err := s.store.UpdateNode(ctx, *dbParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UpdateNodeResponse{}, ErrNodeNotFoundOrNoAccess
		}
		return UpdateNodeResponse{}, err
	}

	response := UpdateNodeResponse{
		Type:      string(dbRow.Type),
		Title:     dbRow.Title,
		UpdatedAt: &dbRow.UpdatedAt.Time,
	}

	return response, nil
}

func (s *Service) MoveNode(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID, req *MoveNodeRequest) (MoveNodeResponse, error) {
	if !req.ParentID.IsSet && req.SortOrder == nil {
		return MoveNodeResponse{}, ErrEmptyUpdate
	}

	dbParams := &node.MoveNodeParams{
		ID:           nodeID,
		UserID:       userID,
		UpdateParent: pgtype.Bool{Bool: req.ParentID.IsSet, Valid: true},
		ParentID:     pgtype.UUID{Valid: false},
	}

	if req.ParentID.IsSet && req.ParentID.Value != nil {
		if *req.ParentID.Value == "" {
			return MoveNodeResponse{}, ErrInvalidParentID
		}

		parsedID, err := httputil.PgUUIDFromString(req.ParentID.Value)
		if err != nil {
			return MoveNodeResponse{}, ErrInvalidParentID
		}
		dbParams.ParentID = parsedID

		if parsedID == nodeID {
			return MoveNodeResponse{}, ErrNodeCannotBeADescendantOfItself
		}

		if _, err := s.store.GetNodeByID(ctx, node.GetNodeByIDParams{ID: parsedID, UserID: userID}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return MoveNodeResponse{}, ErrParentNotFound
			}
			return MoveNodeResponse{}, err
		}

		isDescendantOfItself, err := s.isNodeDescendantOfItself(ctx, nodeID, parsedID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return MoveNodeResponse{}, ErrParentNotFound
			}
			return MoveNodeResponse{}, ErrInvalidParentID
		}
		if isDescendantOfItself {
			return MoveNodeResponse{}, ErrNodeCannotBeADescendantOfItself
		}
	}

	if req.SortOrder != nil {
		dbParams.SortOrder = pgtype.Int8{Int64: *req.SortOrder, Valid: true}
	}

	dbRow, err := s.store.MoveNode(ctx, *dbParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MoveNodeResponse{}, ErrNodeNotFoundOrNoAccess
		}
		return MoveNodeResponse{}, err
	}

	var responseParentID *string
	if dbRow.ParentID.Valid {
		str := dbRow.ParentID.String()
		responseParentID = &str
	}

	response := MoveNodeResponse{
		ParentID:  responseParentID,
		SortOrder: dbRow.SortOrder,
		UpdatedAt: &dbRow.UpdatedAt.Time,
	}

	return response, nil
}

func (s *Service) isNodeDescendantOfItself(ctx context.Context, nodeID pgtype.UUID, parentID pgtype.UUID) (bool, error) {
	ancestors, err := s.store.GetNodeAncestors(ctx, parentID)
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

func (s *Service) GetChildren(ctx context.Context, parentID pgtype.UUID, userID pgtype.UUID) ([]NodeResponse, error) {
	if _, err := s.store.GetNodeByID(ctx, node.GetNodeByIDParams{ID: parentID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []NodeResponse{}, ErrParentNotFound
		}
		return []NodeResponse{}, err
	}

	dbParams := &node.GetChildrenParams{
		ParentID: parentID,
		UserID:   userID,
	}

	dbRow, err := s.store.GetChildren(ctx, *dbParams)
	if err != nil {
		return []NodeResponse{}, err
	}

	response := make([]NodeResponse, 0, len(dbRow))
	for _, n := range dbRow {
		response = append(response, mapNodeToResponse(n))
	}

	return response, nil
}

func mapNodeToResponse(n node.Node) NodeResponse {
	res := NodeResponse{
		ID:        n.ID.String(),
		UserID:    n.UserID.String(),
		ParentID:  n.ParentID.String(),
		Type:      string(n.Type),
		Title:     n.Title,
		SortOrder: n.SortOrder,
		UpdatedAt: &n.UpdatedAt.Time,
		CreatedAt: &n.CreatedAt.Time,
		DeletedAt: &n.DeletedAt.Time,
	}

	return res
}
