package hierarchy_test

import (
	"context"
	"time"

	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	sqlcNode "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type hierarchyStoreFake struct {
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
	prevRanks           []*int64
	prevRankCalls       []sqlcHierarchy.GetPrevSiblingRankParams
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
	call := len(f.prevRankCalls)
	f.prevRankCalls = append(f.prevRankCalls, params)
	if call >= len(f.prevRanks) || f.prevRanks[call] == nil {
		return 0, pgx.ErrNoRows
	}
	return *f.prevRanks[call], nil
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

type nodeStoreFake struct {
	getNodeByIDResult map[string]sqlcNode.Node
	getNodeByIDErr    error
}

func (f *nodeStoreFake) GetNodeByID(_ context.Context, params sqlcNode.GetNodeByIDParams) (sqlcNode.Node, error) {
	if f.getNodeByIDErr != nil {
		return sqlcNode.Node{}, f.getNodeByIDErr
	}
	if node, ok := f.getNodeByIDResult[params.ID.String()]; ok && node.UserID == params.UserID {
		return node, nil
	}
	return sqlcNode.Node{}, pgx.ErrNoRows
}

func nodeStoreWith(nodes ...sqlcNode.Node) *nodeStoreFake {
	result := make(map[string]sqlcNode.Node, len(nodes))
	for _, n := range nodes {
		result[n.ID.String()] = n
	}
	return &nodeStoreFake{getNodeByIDResult: result}
}

type txFake struct {
	repos hierarchy.Repos
	calls int
}

func (f *txFake) InTx(_ context.Context, fn func(hierarchy.Repos) error) error {
	f.calls++
	return fn(f.repos)
}

func newTestService(store *hierarchyStoreFake, nodeStore *nodeStoreFake) *hierarchy.Service {
	return hierarchy.NewService(store, nodeStore, &txFake{repos: hierarchy.Repos{Store: store, NodeStore: nodeStore}})
}

func int64Ptr(v int64) *int64 { return &v }
