package hierarchy

import (
	"context"

	"github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	"github.com/KubantsevAS/notree/backend/internal/db/node"
)

type NodeStore interface {
	GetNodeByID(context.Context, node.GetNodeByIDParams) (node.Node, error)
}

type Store interface {
	GetParent(context.Context, hierarchy.GetParentParams) (hierarchy.Node, error) //TODO
	//TODO GetChildren(context.Context, node.GetChildrenParams) ([]node.Node, error)

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
