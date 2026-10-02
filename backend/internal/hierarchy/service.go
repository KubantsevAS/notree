package hierarchy

import (
	"context"
	"errors"

	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	sqlcNode "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy/rank"
	"github.com/KubantsevAS/notree/backend/internal/http/httputil"
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

	IsInSubtree(context.Context, sqlcHierarchy.IsInSubtreeParams) (bool, error)

	//TODO GetBreadcrumbs(context.Context, pgtype.UUID) (BreadcrumbItem, error)

	LockUserHierarchy(context.Context, pgtype.UUID) error
	GetPrevSiblingRank(context.Context, sqlcHierarchy.GetPrevSiblingRankParams) (int64, error)
	RebalanceChildren(context.Context, sqlcHierarchy.RebalanceChildrenParams) error
	MoveNode(context.Context, sqlcHierarchy.MoveNodeParams) (sqlcHierarchy.MoveNodeRow, error)
}

// * Example
// * type BreadcrumbItem struct {
// *   ID    pgtype.UUID `json: "ID"`
// *   Title string      `json: "title"`
// * }

type Service struct {
	store     Store
	nodeStore NodeStore
	tx        Transactor
}

func NewService(store Store, nodeStore NodeStore, tx Transactor) *Service {
	return &Service{store: store, nodeStore: nodeStore, tx: tx}
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

func (s *Service) MoveNode(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID, req *MoveNodeRequest) (MoveNodeResponse, error) {
	if !req.ParentID.IsSet {
		return MoveNodeResponse{}, ErrParentIDRequired
	}
	if !req.BeforeID.IsSet {
		return MoveNodeResponse{}, ErrBeforeIDRequired
	}

	parentID, err := httputil.PgUUIDFromString(req.ParentID.Value)
	if err != nil {
		return MoveNodeResponse{}, ErrInvalidParentID
	}
	beforeID, err := httputil.PgUUIDFromString(req.BeforeID.Value)
	if err != nil {
		return MoveNodeResponse{}, ErrInvalidBeforeID
	}

	var row sqlcHierarchy.MoveNodeRow
	err = s.tx.InTx(ctx, func(r Repos) error {
		if err := r.Store.LockUserHierarchy(ctx, userID); err != nil {
			return err
		}

		node, err := r.NodeStore.GetNodeByID(ctx, sqlcNode.GetNodeByIDParams{ID: nodeID, UserID: userID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNodeNotFound
			}
			return err
		}

		if node.ParentID != parentID && parentID.Valid {
			if err := validateNewParent(ctx, r, nodeID, parentID, userID); err != nil {
				return err
			}
		}

		sortOrder, err := rankFor(ctx, r, nodeID, parentID, beforeID, userID)
		if err != nil {
			return err
		}

		row, err = r.Store.MoveNode(ctx, sqlcHierarchy.MoveNodeParams{
			ID:        nodeID,
			UserID:    userID,
			ParentID:  parentID,
			SortOrder: sortOrder,
		})
		return err
	})
	if err != nil {
		return MoveNodeResponse{}, err
	}

	response := MoveNodeResponse{
		SortOrder: row.SortOrder,
		UpdatedAt: &row.UpdatedAt.Time,
	}
	if row.ParentID.Valid {
		pid := row.ParentID.String()
		response.ParentID = &pid
	}

	return response, nil
}

func validateNewParent(ctx context.Context, r Repos, nodeID, parentID, userID pgtype.UUID) error {
	if parentID == nodeID {
		return ErrNodeCannotBeADescendantOfItself
	}

	if _, err := r.NodeStore.GetNodeByID(ctx, sqlcNode.GetNodeByIDParams{ID: parentID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrParentNotFound
		}
		return err
	}

	inSubtree, err := r.Store.IsInSubtree(ctx, sqlcHierarchy.IsInSubtreeParams{
		RootID: nodeID,
		NodeID: parentID,
		UserID: userID,
	})
	if err != nil {
		return err
	}
	if inSubtree {
		return ErrNodeCannotBeADescendantOfItself
	}

	return nil
}

func rankFor(ctx context.Context, r Repos, nodeID, parentID, beforeID, userID pgtype.UUID) (int64, error) {
	for attempt := 0; ; attempt++ {
		prev, next, err := neighbourRanks(ctx, r, nodeID, parentID, beforeID, userID)
		if err != nil {
			return 0, err
		}

		if sortOrder, ok := rank.Between(prev, next); ok {
			return sortOrder, nil
		}
		if attempt > 0 {
			return 0, ErrNoRankAfterRebalance
		}

		err = r.Store.RebalanceChildren(ctx, sqlcHierarchy.RebalanceChildrenParams{
			Gap:       rank.Gap,
			ParentID:  parentID,
			UserID:    userID,
			ExcludeID: nodeID,
		})
		if err != nil {
			return 0, err
		}
	}
}

func neighbourRanks(ctx context.Context, r Repos, nodeID, parentID, beforeID, userID pgtype.UUID) (prev, next *int64, err error) {
	params := sqlcHierarchy.GetPrevSiblingRankParams{
		ParentID:  parentID,
		UserID:    userID,
		ExcludeID: nodeID,
	}

	if beforeID.Valid {
		if beforeID == nodeID {
			return nil, nil, ErrBeforeNotSibling
		}
		before, err := r.NodeStore.GetNodeByID(ctx, sqlcNode.GetNodeByIDParams{ID: beforeID, UserID: userID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil, ErrBeforeNotSibling
			}
			return nil, nil, err
		}
		if before.ParentID != parentID {
			return nil, nil, ErrBeforeNotSibling
		}

		next = &before.SortOrder
		params.BeforeRank = pgtype.Int8{Int64: before.SortOrder, Valid: true}
		params.BeforeID = beforeID
	}

	prevRank, err := r.Store.GetPrevSiblingRank(ctx, params)
	switch {
	case err == nil:
		prev = &prevRank
	case !errors.Is(err, pgx.ErrNoRows):
		return nil, nil, err
	}

	return prev, next, nil
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
