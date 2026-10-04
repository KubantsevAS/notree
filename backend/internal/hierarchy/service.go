package hierarchy

import (
	"context"
	"errors"

	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy/rank"
	"github.com/KubantsevAS/notree/backend/internal/http/httputil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	GetNode(context.Context, sqlcHierarchy.GetNodeParams) (sqlcHierarchy.Node, error)
	GetParent(context.Context, sqlcHierarchy.GetParentParams) (sqlcHierarchy.Node, error)
	GetChildren(context.Context, sqlcHierarchy.GetChildrenParams) ([]sqlcHierarchy.Node, error)

	GetAncestors(context.Context, sqlcHierarchy.GetAncestorsParams) ([]sqlcHierarchy.Node, error)
	GetDescendants(context.Context, sqlcHierarchy.GetDescendantsParams) ([]sqlcHierarchy.Node, error)

	GetSubtree(context.Context, sqlcHierarchy.GetSubtreeParams) ([]sqlcHierarchy.Node, error)
	GetRoot(context.Context, sqlcHierarchy.GetRootParams) (sqlcHierarchy.Node, error)

	IsInSubtree(context.Context, sqlcHierarchy.IsInSubtreeParams) (bool, error)

	GetBreadcrumbs(context.Context, sqlcHierarchy.GetBreadcrumbsParams) ([]sqlcHierarchy.GetBreadcrumbsRow, error)

	LockUserHierarchy(context.Context, pgtype.UUID) error
	GetPrevSiblingRank(context.Context, sqlcHierarchy.GetPrevSiblingRankParams) (int64, error)
	GetLastSiblingRank(context.Context, sqlcHierarchy.GetLastSiblingRankParams) (int64, error)
	RebalanceChildren(context.Context, sqlcHierarchy.RebalanceChildrenParams) error
	MoveNode(context.Context, sqlcHierarchy.MoveNodeParams) (sqlcHierarchy.MoveNodeRow, error)
}

type Service struct {
	store Store
	tx    Transactor
}

func NewService(store Store, tx Transactor) *Service {
	return &Service{store: store, tx: tx}
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

	response := mapSlice(nodes, mapNodeToResponse)

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
	node, err := s.store.GetNode(ctx, sqlcHierarchy.GetNodeParams{ID: nodeID, UserID: userID})
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

	response := mapSlice(ancestors, mapNodeToResponse)

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

	response := mapSlice(descendants, mapNodeToResponse)

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

	response := mapSlice(subtree, mapNodeToResponse)

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
	err = s.tx.InTx(ctx, func(store Store) error {
		if err := store.LockUserHierarchy(ctx, userID); err != nil {
			return err
		}

		node, err := store.GetNode(ctx, sqlcHierarchy.GetNodeParams{ID: nodeID, UserID: userID})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNodeNotFound
			}
			return err
		}

		if node.ParentID != parentID && parentID.Valid {
			if err := validateNewParent(ctx, store, nodeID, parentID, userID); err != nil {
				return err
			}
		}

		sortOrder, err := rankFor(ctx, store, nodeID, parentID, beforeID, userID)
		if err != nil {
			return err
		}

		row, err = store.MoveNode(ctx, sqlcHierarchy.MoveNodeParams{
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

func validateNewParent(ctx context.Context, store Store, nodeID, parentID, userID pgtype.UUID) error {
	if parentID == nodeID {
		return ErrNodeCannotBeADescendantOfItself
	}

	if _, err := store.GetNode(ctx, sqlcHierarchy.GetNodeParams{ID: parentID, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrParentNotFound
		}
		return err
	}

	inSubtree, err := store.IsInSubtree(ctx, sqlcHierarchy.IsInSubtreeParams{
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

func rankFor(ctx context.Context, store Store, nodeID, parentID, beforeID, userID pgtype.UUID) (int64, error) {
	for attempt := 0; ; attempt++ {
		prev, next, err := neighbourRanks(ctx, store, nodeID, parentID, beforeID, userID)
		if err != nil {
			return 0, err
		}

		if sortOrder, ok := rank.Between(prev, next); ok {
			return sortOrder, nil
		}
		if attempt > 0 {
			return 0, ErrNoRankAfterRebalance
		}

		err = store.RebalanceChildren(ctx, sqlcHierarchy.RebalanceChildrenParams{
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

func neighbourRanks(ctx context.Context, store Store, nodeID, parentID, beforeID, userID pgtype.UUID) (prev, next *int64, err error) {
	var prevRank int64
	if beforeID.Valid {
		if beforeID == nodeID {
			return nil, nil, ErrBeforeNotSibling
		}
		before, getErr := store.GetNode(ctx, sqlcHierarchy.GetNodeParams{ID: beforeID, UserID: userID})
		if getErr != nil {
			if errors.Is(getErr, pgx.ErrNoRows) {
				return nil, nil, ErrBeforeNotSibling
			}
			return nil, nil, getErr
		}
		if before.ParentID != parentID {
			return nil, nil, ErrBeforeNotSibling
		}

		next = &before.SortOrder
		prevRank, err = store.GetPrevSiblingRank(ctx, sqlcHierarchy.GetPrevSiblingRankParams{
			UserID:     userID,
			ParentID:   parentID,
			ExcludeID:  nodeID,
			BeforeRank: before.SortOrder,
			BeforeID:   beforeID,
		})
	} else {
		prevRank, err = store.GetLastSiblingRank(ctx, sqlcHierarchy.GetLastSiblingRankParams{
			UserID:    userID,
			ParentID:  parentID,
			ExcludeID: nodeID,
		})
	}

	switch {
	case err == nil:
		prev = &prevRank
	case !errors.Is(err, pgx.ErrNoRows):
		return nil, nil, err
	}

	return prev, next, nil
}

func (s *Service) ensureNodeExists(ctx context.Context, nodeID, userID pgtype.UUID) error {
	_, err := s.store.GetNode(ctx, sqlcHierarchy.GetNodeParams{
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

func (s *Service) GetBreadcrumbs(ctx context.Context, nodeID, userID pgtype.UUID) (BreadcrumbsResponse, error) {
	if err := s.ensureNodeExists(ctx, nodeID, userID); err != nil {
		return nil, err
	}

	breadcrumbs, err := s.store.GetBreadcrumbs(ctx, sqlcHierarchy.GetBreadcrumbsParams{
		ID:     nodeID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	response := mapSlice(breadcrumbs, mapBreadcrumbToResponse)

	return response, nil
}

func mapBreadcrumbToResponse(r sqlcHierarchy.GetBreadcrumbsRow) BreadcrumbResponse {
	return BreadcrumbResponse{
		ID:    r.ID.String(),
		Title: r.Title,
	}
}

func mapSlice[T, R any](items []T, fn func(T) R) []R {
	out := make([]R, 0, len(items))
	for _, item := range items {
		out = append(out, fn(item))
	}
	return out
}
