package hierarchy

import (
	"context"
	"errors"

	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	sqlcNode "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type NodeStore interface {
	GetNodeByID(context.Context, sqlcNode.GetNodeByIDParams) (sqlcNode.Node, error)
}

type Store interface {
	GetParent(context.Context, sqlcHierarchy.GetParentParams) (sqlcHierarchy.Node, error)
	GetChildren(context.Context, sqlcHierarchy.GetChildrenParams) ([]sqlcHierarchy.Node, error)

	GetAncestors(context.Context, sqlcHierarchy.GetAncestorsParams) ([]sqlcHierarchy.Node, error)
	GetDescendants(context.Context, sqlcHierarchy.GetDescendantsParams) ([]sqlcHierarchy.Node, error)

	GetSubtree(context.Context, sqlcHierarchy.GetSubtreeParams) ([]sqlcHierarchy.Node, error)
	GetRoot(context.Context, sqlcHierarchy.GetRootParams) (sqlcHierarchy.Node, error)

	//TODO IsDescendant(
	// 	ctx context.Context,
	// 	nodeID pgtype.UUID,
	// 	potentialAncestorID pgtype.UUID,
	// 	userID pgtype.UUID,
	// ) (bool, error)

	//TODO GetBreadcrumbs(context.Context, pgtype.UUID) (BreadcrumbItem, error)

	//TODO MoveNode(context.Context, node.MoveNodeParams) (node.MoveNodeRow, error)

	//TODO ReorderNode(context.Context, pgtype.UUID) error
}

// * Example
// * type BreadcrumbItem struct {
// *   ID    pgtype.UUID `json: "ID"`
// *   Title string      `json: "title"`
// * }

type Service struct {
	store     Store
	nodeStore NodeStore
}

func NewService(store Store, nodeStore NodeStore) *Service {
	return &Service{store: store, nodeStore: nodeStore}
}

func (s *Service) GetChildren(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID) ([]NodeResponse, error) {
	if err := s.ensureNodeExists(ctx, nodeID, userID); err != nil {
		return nil, err
	}

	params := sqlcHierarchy.GetChildrenParams{
		ParentID: nodeID,
		UserID:   userID,
	}

	nodes, err := s.store.GetChildren(ctx, params)
	if err != nil {
		return nil, err
	}

	response := make([]NodeResponse, 0, len(nodes))
	for _, n := range nodes {
		response = append(response, mapNodeToResponse(n))
	}

	return response, nil
}

func mapNodeToResponse(n sqlcHierarchy.Node) NodeResponse {
	res := NodeResponse{
		ID:        n.ID.String(),
		UserID:    n.UserID.String(),
		Type:      string(n.Type),
		Title:     n.Title,
		SortOrder: n.SortOrder,
		UpdatedAt: &n.UpdatedAt.Time,
		CreatedAt: &n.CreatedAt.Time,
	}
	if n.ParentID.Valid {
		pid := n.ParentID.String()
		res.ParentID = &pid
	}
	if n.DeletedAt.Valid {
		res.DeletedAt = &n.DeletedAt.Time
	}

	return res
}

func (s *Service) GetParent(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID) (NodeResponse, error) {
	node, err := s.nodeStore.GetNodeByID(ctx, sqlcNode.GetNodeByIDParams{ID: nodeID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return NodeResponse{}, ErrNodeNotFound
		}
		return NodeResponse{}, err
	}

	if !node.ParentID.Valid {
		return NodeResponse{}, ErrNodeIsRoot
	}

	parent, err := s.store.GetParent(ctx, sqlcHierarchy.GetParentParams{
		ID:     nodeID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return NodeResponse{}, ErrParentNotFound
		}
		return NodeResponse{}, err
	}

	return mapNodeToResponse(parent), nil
}

func (s *Service) GetAncestors(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID) (GetAncestorsResponse, error) {
	if err := s.ensureNodeExists(ctx, nodeID, userID); err != nil {
		return nil, err
	}

	ancestors, err := s.store.GetAncestors(ctx, sqlcHierarchy.GetAncestorsParams{
		ID:     nodeID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	response := make([]NodeResponse, 0, len(ancestors))
	for _, n := range ancestors {
		response = append(response, mapNodeToResponse(n))
	}

	return response, nil
}

func (s *Service) GetDescendants(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID) (GetDescendantsResponse, error) {
	if err := s.ensureNodeExists(ctx, nodeID, userID); err != nil {
		return nil, err
	}

	descendants, err := s.store.GetDescendants(ctx, sqlcHierarchy.GetDescendantsParams{
		ParentID: nodeID,
		UserID:   userID,
	})
	if err != nil {
		return nil, err
	}

	response := make([]NodeResponse, 0, len(descendants))
	for _, n := range descendants {
		response = append(response, mapNodeToResponse(n))
	}

	return response, nil
}

func (s *Service) GetSubtree(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID) (GetSubtreeResponse, error) {
	if err := s.ensureNodeExists(ctx, nodeID, userID); err != nil {
		return nil, err
	}

	subtree, err := s.store.GetSubtree(ctx, sqlcHierarchy.GetSubtreeParams{
		ID:     nodeID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	response := make([]NodeResponse, 0, len(subtree))
	for _, n := range subtree {
		response = append(response, mapNodeToResponse(n))
	}

	return response, nil
}

func (s *Service) GetRoot(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID) (NodeResponse, error) {
	if err := s.ensureNodeExists(ctx, nodeID, userID); err != nil {
		return NodeResponse{}, err
	}

	root, err := s.store.GetRoot(ctx, sqlcHierarchy.GetRootParams{
		ID:     nodeID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return NodeResponse{}, ErrRootNotFound
		}
		return NodeResponse{}, err
	}

	return mapNodeToResponse(root), nil
}

func (s *Service) ensureNodeExists(ctx context.Context, nodeID, userID pgtype.UUID) error {
	_, err := s.nodeStore.GetNodeByID(ctx, sqlcNode.GetNodeByIDParams{
		ID:     nodeID,
		UserID: userID,
	})
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows):
		return ErrNodeNotFound
	default:
		return err
	}
}
