package hierarchy

import (
	"context"
	"errors"

	"github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	"github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type NodeStore interface {
	GetNodeByID(context.Context, node.GetNodeByIDParams) (node.Node, error)
}

type Store interface {
	GetParent(context.Context, hierarchy.GetParentParams) (hierarchy.Node, error) //TODO
	GetChildren(context.Context, hierarchy.GetChildrenParams) ([]hierarchy.Node, error)

	//TODO GetAncestors(context.Context, pgtype.UUID) ([]node.Node, error)
	//TODO GetDescendants(context.Context, pgtype.UUID) ([]node.Node, error)

	//TODO IsDescendant(
	// 	ctx context.Context,
	// 	nodeID pgtype.UUID,
	// 	potentialAncestorID pgtype.UUID,
	// 	userID pgtype.UUID,
	// ) (bool, error)

	//TODO GetSubtree(context.Context, pgtype.UUID) ([]node.Node, error) // GetDescendants + Node itself
	//TODO GetRoot(context.Context, pgtype.UUID) (node.Node, error)

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
	if _, err := s.nodeStore.GetNodeByID(ctx, node.GetNodeByIDParams{ID: nodeID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrParentNotFound
		}
		return nil, err
	}

	params := hierarchy.GetChildrenParams{
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

func mapNodeToResponse(n hierarchy.Node) NodeResponse {
	res := NodeResponse{
		ID:        n.ID.String(),
		UserID:    n.UserID.String(),
		ParentID:  n.ParentID.String(),
		Type:      string(n.Type),
		Title:     n.Title,
		SortOrder: n.SortOrder,
		UpdatedAt: &n.UpdatedAt.Time,
		CreatedAt: &n.CreatedAt.Time,
	}
	if n.DeletedAt.Valid {
		res.DeletedAt = &n.DeletedAt.Time
	}

	return res
}
