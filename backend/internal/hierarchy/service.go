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
	store  Store
	nodeDb NodeStore
}

func NewService(store Store, nodeDb NodeStore) *Service {
	return &Service{store: store, nodeDb: nodeDb}
}

func (s *Service) GetChildren(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID) ([]NodeResponse, error) {
	if _, err := s.nodeDb.GetNodeByID(ctx, node.GetNodeByIDParams{ID: nodeID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrParentNotFound
		}
		return nil, err
	}

	dbParams := hierarchy.GetChildrenParams{
		ParentID: nodeID,
		UserID:   userID,
	}

	dbRow, err := s.store.GetChildren(ctx, dbParams)
	if err != nil {
		return nil, err
	}

	response := make([]NodeResponse, 0, len(dbRow))
	for _, n := range dbRow {
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
