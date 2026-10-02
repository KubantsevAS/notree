package hierarchy_test

import (
	"context"
	"time"

	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type hierarchyStoreFake struct {
	nodes               *nodeSet
	children            []sqlcHierarchy.Node
	childrenErr         error
	parent              sqlcHierarchy.Node
	parentErr           error
	getParentCalls      int
	lastGetParentParams sqlcHierarchy.GetParentParams
	ancestors           []sqlcHierarchy.Node
	ancestorsErr        error
	lastAncestorsArgs   sqlcHierarchy.GetAncestorsParams
	descendants         []sqlcHierarchy.Node
	descendantsErr      error
	lastDescendantsArgs sqlcHierarchy.GetDescendantsParams
	subtree             []sqlcHierarchy.Node
	subtreeErr          error
	lastSubtreeArgs     sqlcHierarchy.GetSubtreeParams
	root                sqlcHierarchy.Node
	rootErr             error
	lastRootArgs        sqlcHierarchy.GetRootParams
	inSubtree           bool
	inSubtreeErr        error
	inSubtreeCalls      []sqlcHierarchy.IsInSubtreeParams
	lockErr             error
	lockCalls           []pgtype.UUID
	siblingRanks        []*int64
	prevRankCalls       []sqlcHierarchy.GetPrevSiblingRankParams
	lastRankCalls       []sqlcHierarchy.GetLastSiblingRankParams
	rebalanceErr        error
	rebalanceCalls      []sqlcHierarchy.RebalanceChildrenParams
	moveErr             error
	moveCalls           []sqlcHierarchy.MoveNodeParams
}

func (f *hierarchyStoreFake) GetChildren(context.Context, sqlcHierarchy.GetChildrenParams) ([]sqlcHierarchy.Node, error) {
	if f.childrenErr != nil {
		return []sqlcHierarchy.Node{}, f.childrenErr
	}
	return f.children, nil
}

func (f *hierarchyStoreFake) GetParent(_ context.Context, params sqlcHierarchy.GetParentParams) (sqlcHierarchy.Node, error) {
	f.getParentCalls++
	f.lastGetParentParams = params
	if f.parentErr != nil {
		return sqlcHierarchy.Node{}, f.parentErr
	}
	return f.parent, nil
}

func (f *hierarchyStoreFake) GetAncestors(_ context.Context, params sqlcHierarchy.GetAncestorsParams) ([]sqlcHierarchy.Node, error) {
	f.lastAncestorsArgs = params
	if f.ancestorsErr != nil {
		return nil, f.ancestorsErr
	}
	return f.ancestors, nil
}

func (f *hierarchyStoreFake) GetDescendants(_ context.Context, params sqlcHierarchy.GetDescendantsParams) ([]sqlcHierarchy.Node, error) {
	f.lastDescendantsArgs = params
	if f.descendantsErr != nil {
		return nil, f.descendantsErr
	}
	return f.descendants, nil
}

func (f *hierarchyStoreFake) GetSubtree(_ context.Context, params sqlcHierarchy.GetSubtreeParams) ([]sqlcHierarchy.Node, error) {
	f.lastSubtreeArgs = params
	if f.subtreeErr != nil {
		return nil, f.subtreeErr
	}
	return f.subtree, nil
}

func (f *hierarchyStoreFake) GetRoot(_ context.Context, params sqlcHierarchy.GetRootParams) (sqlcHierarchy.Node, error) {
	f.lastRootArgs = params
	if f.rootErr != nil {
		return sqlcHierarchy.Node{}, f.rootErr
	}
	return f.root, nil
}

func (f *hierarchyStoreFake) IsInSubtree(_ context.Context, params sqlcHierarchy.IsInSubtreeParams) (bool, error) {
	f.inSubtreeCalls = append(f.inSubtreeCalls, params)
	if f.inSubtreeErr != nil {
		return false, f.inSubtreeErr
	}
	return f.inSubtree, nil
}

func (f *hierarchyStoreFake) LockUserHierarchy(_ context.Context, userID pgtype.UUID) error {
	f.lockCalls = append(f.lockCalls, userID)
	return f.lockErr
}

func (f *hierarchyStoreFake) GetPrevSiblingRank(_ context.Context, params sqlcHierarchy.GetPrevSiblingRankParams) (int64, error) {
	f.prevRankCalls = append(f.prevRankCalls, params)
	return f.nextSiblingRank()
}

func (f *hierarchyStoreFake) GetLastSiblingRank(_ context.Context, params sqlcHierarchy.GetLastSiblingRankParams) (int64, error) {
	f.lastRankCalls = append(f.lastRankCalls, params)
	return f.nextSiblingRank()
}

func (f *hierarchyStoreFake) nextSiblingRank() (int64, error) {
	call := len(f.prevRankCalls) + len(f.lastRankCalls) - 1
	if call >= len(f.siblingRanks) || f.siblingRanks[call] == nil {
		return 0, pgx.ErrNoRows
	}
	return *f.siblingRanks[call], nil
}

func (f *hierarchyStoreFake) RebalanceChildren(_ context.Context, params sqlcHierarchy.RebalanceChildrenParams) error {
	f.rebalanceCalls = append(f.rebalanceCalls, params)
	return f.rebalanceErr
}

func (f *hierarchyStoreFake) MoveNode(_ context.Context, params sqlcHierarchy.MoveNodeParams) (sqlcHierarchy.MoveNodeRow, error) {
	f.moveCalls = append(f.moveCalls, params)
	if f.moveErr != nil {
		return sqlcHierarchy.MoveNodeRow{}, f.moveErr
	}
	return sqlcHierarchy.MoveNodeRow{
		ParentID:  params.ParentID,
		SortOrder: params.SortOrder,
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, nil
}

type nodeSet struct {
	nodes map[string]sqlcHierarchy.Node
	err   error
}

func nodeSetOf(nodes ...sqlcHierarchy.Node) *nodeSet {
	result := make(map[string]sqlcHierarchy.Node, len(nodes))
	for _, n := range nodes {
		result[n.ID.String()] = n
	}
	return &nodeSet{nodes: result}
}

func (f *hierarchyStoreFake) GetNode(_ context.Context, params sqlcHierarchy.GetNodeParams) (sqlcHierarchy.Node, error) {
	if f.nodes == nil {
		return sqlcHierarchy.Node{}, pgx.ErrNoRows
	}
	if f.nodes.err != nil {
		return sqlcHierarchy.Node{}, f.nodes.err
	}
	if node, ok := f.nodes.nodes[params.ID.String()]; ok && node.UserID == params.UserID {
		return node, nil
	}
	return sqlcHierarchy.Node{}, pgx.ErrNoRows
}

type txFake struct {
	store *hierarchyStoreFake
	calls int
}

func (f *txFake) InTx(_ context.Context, fn func(hierarchy.Store) error) error {
	f.calls++
	return fn(f.store)
}

func newTestService(store *hierarchyStoreFake, nodes *nodeSet) *hierarchy.Service {
	store.nodes = nodes
	return hierarchy.NewService(store, &txFake{store: store})
}

func int64Ptr(v int64) *int64 { return &v }
