package hierarchy_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy/rank"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestService_NodeNotFound(t *testing.T) {
	ctx := context.Background()
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)

	tests := []struct {
		name string
		call func(*hierarchy.Service) error
	}{
		{"GetChildren", func(s *hierarchy.Service) error { _, err := s.GetChildren(ctx, nodeID, userID); return err }},
		{"GetParent", func(s *hierarchy.Service) error { _, err := s.GetParent(ctx, nodeID, userID); return err }},
		{"GetAncestors", func(s *hierarchy.Service) error { _, err := s.GetAncestors(ctx, nodeID, userID); return err }},
		{"GetDescendants", func(s *hierarchy.Service) error { _, err := s.GetDescendants(ctx, nodeID, userID); return err }},
		{"GetSubtree", func(s *hierarchy.Service) error { _, err := s.GetSubtree(ctx, nodeID, userID); return err }},
		{"GetRoot", func(s *hierarchy.Service) error { _, err := s.GetRoot(ctx, nodeID, userID); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call(newTestService(&hierarchyStoreFake{}, &nodeSet{}))
			require.ErrorIs(t, err, hierarchy.ErrNodeNotFound)
		})
	}
}

func TestGetChildren(t *testing.T) {
	parentID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	userID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	childOneID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	childTwoID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}
	nodes := nodeSetOf(sqlcHierarchy.Node{ID: parentID, UserID: userID})
	fake := &hierarchyStoreFake{
		children: []sqlcHierarchy.Node{
			{ID: childOneID, ParentID: parentID, UserID: userID, Title: "first", SortOrder: 10},
			{ID: childTwoID, ParentID: parentID, UserID: userID, Title: "second", SortOrder: 20},
		},
	}

	service := newTestService(fake, nodes)
	children, err := service.GetChildren(context.Background(), parentID, userID)
	require.NoError(t, err)
	require.Len(t, children, 2)
	require.Equal(t, childOneID.String(), children[0].ID)
	require.Equal(t, childTwoID.String(), children[1].ID)
}

func TestGetParent(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID3)
	grandparentID := testutil.UUIDFromStringT(t, testutil.UUID4)

	node := sqlcHierarchy.Node{ID: nodeID, UserID: userID, ParentID: parentID}

	createdAt := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 2, 3, 4, 5, 6, 0, time.UTC)
	parent := sqlcHierarchy.Node{
		ID:        parentID,
		UserID:    userID,
		ParentID:  grandparentID,
		Type:      "folder",
		Title:     "parent node",
		SortOrder: 7,
		CreatedAt: pgtype.Timestamptz{Time: createdAt, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
	}

	nodes := nodeSetOf(node)
	store := &hierarchyStoreFake{parent: parent}

	service := newTestService(store, nodes)
	res, err := service.GetParent(context.Background(), nodeID, userID)

	require.NoError(t, err)
	require.Equal(t, parentID.String(), res.ID)
	require.Equal(t, userID.String(), res.UserID)
	require.Equal(t, grandparentID.String(), *res.ParentID)
	require.Equal(t, "parent node", res.Title)
	require.EqualValues(t, "folder", res.Type)
	require.EqualValues(t, 7, res.SortOrder)
	require.NotNil(t, res.CreatedAt)
	require.True(t, res.CreatedAt.Equal(createdAt))
	require.NotNil(t, res.UpdatedAt)
	require.True(t, res.UpdatedAt.Equal(updatedAt))
	require.Nil(t, res.DeletedAt)

	require.Equal(t, 1, store.getParentCalls)
	require.Equal(t, nodeID, store.lastGetParentParams.ID)
	require.Equal(t, userID, store.lastGetParentParams.UserID)
}

func TestGetParent_NodeIsRoot(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	rootID := testutil.UUIDFromStringT(t, testutil.UUID2)

	root := sqlcHierarchy.Node{ID: rootID, UserID: userID}
	nodes := nodeSetOf(root)
	store := &hierarchyStoreFake{parentErr: errors.New("GetParent must not be called for root node")}

	service := newTestService(store, nodes)
	_, err := service.GetParent(context.Background(), rootID, userID)

	require.ErrorIs(t, err, hierarchy.ErrNodeIsRoot)
	require.Zero(t, store.getParentCalls)
}

func TestGetAncestors(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID3)
	rootID := testutil.UUIDFromStringT(t, testutil.UUID4)

	nodes := nodeSetOf(sqlcHierarchy.Node{ID: nodeID, UserID: userID, ParentID: parentID})
	store := &hierarchyStoreFake{
		ancestors: []sqlcHierarchy.Node{
			{ID: rootID, UserID: userID, Title: "root"},
			{ID: parentID, ParentID: rootID, UserID: userID, Title: "parent"},
		},
	}

	service := newTestService(store, nodes)
	res, err := service.GetAncestors(context.Background(), nodeID, userID)

	require.NoError(t, err)
	require.Len(t, res, 2)
	require.Equal(t, rootID.String(), res[0].ID)
	require.Nil(t, res[0].ParentID)
	require.Equal(t, parentID.String(), res[1].ID)
	require.Equal(t, rootID.String(), *res[1].ParentID)
	require.Equal(t, nodeID, store.lastAncestorsArgs.ID)
	require.Equal(t, userID, store.lastAncestorsArgs.UserID)
}

func TestGetAncestors_NodeIsRoot(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	rootID := testutil.UUIDFromStringT(t, testutil.UUID2)

	nodes := nodeSetOf(sqlcHierarchy.Node{ID: rootID, UserID: userID})

	service := newTestService(&hierarchyStoreFake{}, nodes)
	res, err := service.GetAncestors(context.Background(), rootID, userID)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Empty(t, res)
}

func TestGetDescendants(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	rootID := testutil.UUIDFromStringT(t, testutil.UUID2)
	childID := testutil.UUIDFromStringT(t, testutil.UUID3)
	grandchildID := testutil.UUIDFromStringT(t, testutil.UUID4)

	nodes := nodeSetOf(sqlcHierarchy.Node{ID: rootID, UserID: userID})
	store := &hierarchyStoreFake{
		descendants: []sqlcHierarchy.Node{
			{ID: childID, ParentID: rootID, UserID: userID, Title: "child"},
			{ID: grandchildID, ParentID: childID, UserID: userID, Title: "grandchild"},
		},
	}

	service := newTestService(store, nodes)
	res, err := service.GetDescendants(context.Background(), rootID, userID)

	require.NoError(t, err)
	require.Len(t, res, 2)
	require.Equal(t, childID.String(), res[0].ID)
	require.Equal(t, grandchildID.String(), res[1].ID)
	require.Equal(t, childID.String(), *res[1].ParentID)
	require.Equal(t, rootID, store.lastDescendantsArgs.ParentID)
	require.Equal(t, userID, store.lastDescendantsArgs.UserID)
}

func TestGetDescendants_Empty(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)

	nodes := nodeSetOf(sqlcHierarchy.Node{ID: nodeID, UserID: userID})

	service := newTestService(&hierarchyStoreFake{}, nodes)
	res, err := service.GetDescendants(context.Background(), nodeID, userID)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Empty(t, res)
}

func TestGetSubtree(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	childID := testutil.UUIDFromStringT(t, testutil.UUID3)
	grandchildID := testutil.UUIDFromStringT(t, testutil.UUID4)

	nodes := nodeSetOf(sqlcHierarchy.Node{ID: nodeID, UserID: userID})
	store := &hierarchyStoreFake{
		subtree: []sqlcHierarchy.Node{
			{ID: nodeID, UserID: userID, Title: "node"},
			{ID: childID, ParentID: nodeID, UserID: userID, Title: "child"},
			{ID: grandchildID, ParentID: childID, UserID: userID, Title: "grandchild"},
		},
	}

	service := newTestService(store, nodes)
	res, err := service.GetSubtree(context.Background(), nodeID, userID)

	require.NoError(t, err)
	require.Len(t, res, 3)
	require.Equal(t, nodeID.String(), res[0].ID)
	require.Equal(t, childID.String(), res[1].ID)
	require.Equal(t, grandchildID.String(), res[2].ID)
	require.Equal(t, childID.String(), *res[2].ParentID)
	require.Equal(t, nodeID, store.lastSubtreeArgs.ID)
	require.Equal(t, userID, store.lastSubtreeArgs.UserID)
}

func TestGetRoot(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID3)
	rootID := testutil.UUIDFromStringT(t, testutil.UUID4)

	nodes := nodeSetOf(sqlcHierarchy.Node{ID: nodeID, UserID: userID, ParentID: parentID})
	store := &hierarchyStoreFake{
		root: sqlcHierarchy.Node{ID: rootID, UserID: userID, Title: "root"},
	}

	service := newTestService(store, nodes)
	res, err := service.GetRoot(context.Background(), nodeID, userID)

	require.NoError(t, err)
	require.Equal(t, rootID.String(), res.ID)
	require.Equal(t, "root", res.Title)
	require.Nil(t, res.ParentID)
	require.Equal(t, nodeID, store.lastRootArgs.ID)
	require.Equal(t, userID, store.lastRootArgs.UserID)
}

func TestGetRoot_RootNotFound(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)

	nodes := nodeSetOf(sqlcHierarchy.Node{ID: nodeID, UserID: userID})
	store := &hierarchyStoreFake{rootErr: pgx.ErrNoRows}

	service := newTestService(store, nodes)
	_, err := service.GetRoot(context.Background(), nodeID, userID)

	require.ErrorIs(t, err, hierarchy.ErrRootNotFound)
}

func TestMoveNode(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID3)
	siblingID := testutil.UUIDFromStringT(t, testutil.UUID4)
	oldParentID := testutil.UUIDFromStringT(t, "55555555-5555-4555-8555-555555555555")

	moving := sqlcHierarchy.Node{ID: nodeID, UserID: userID, ParentID: oldParentID}
	parent := sqlcHierarchy.Node{ID: parentID, UserID: userID}
	sibling := sqlcHierarchy.Node{ID: siblingID, UserID: userID, ParentID: parentID, SortOrder: 2000}

	moveReq := func(parent, before *string) *hierarchy.MoveNodeRequest {
		return &hierarchy.MoveNodeRequest{
			ParentID: hierarchy.NullableString{Value: parent, IsSet: true},
			BeforeID: hierarchy.NullableString{Value: before, IsSet: true},
		}
	}
	str := testutil.StringPtr

	tests := []struct {
		name    string
		req     *hierarchy.MoveNodeRequest
		store   *hierarchyStoreFake
		nodes   *nodeSet
		wantErr error
		check   func(*testing.T, hierarchy.MoveNodeResponse, *hierarchyStoreFake)
	}{
		{
			name:    "parent_id missing",
			req:     &hierarchy.MoveNodeRequest{BeforeID: hierarchy.NullableString{IsSet: true}},
			wantErr: hierarchy.ErrParentIDRequired,
		},
		{
			name:    "before_id missing",
			req:     &hierarchy.MoveNodeRequest{ParentID: hierarchy.NullableString{IsSet: true}},
			wantErr: hierarchy.ErrBeforeIDRequired,
		},
		{name: "invalid parent uuid", req: moveReq(str(testutil.BadUUID), nil), wantErr: hierarchy.ErrInvalidParentID},
		{name: "empty parent uuid", req: moveReq(str(""), nil), wantErr: hierarchy.ErrInvalidParentID},
		{name: "invalid before uuid", req: moveReq(nil, str(testutil.BadUUID)), wantErr: hierarchy.ErrInvalidBeforeID},

		{name: "node not found", req: moveReq(nil, nil), wantErr: hierarchy.ErrNodeNotFound},
		{
			name:    "lock error",
			req:     moveReq(nil, nil),
			store:   &hierarchyStoreFake{lockErr: sql.ErrConnDone},
			nodes:   nodeSetOf(moving),
			wantErr: sql.ErrConnDone,
		},
		{
			name:    "self parent",
			req:     moveReq(str(testutil.UUID2), nil),
			nodes:   nodeSetOf(moving),
			wantErr: hierarchy.ErrNodeCannotBeADescendantOfItself,
		},
		{
			name:    "parent not found",
			req:     moveReq(str(testutil.UUID3), nil),
			nodes:   nodeSetOf(moving),
			wantErr: hierarchy.ErrParentNotFound,
		},
		{
			name:    "parent inside node subtree",
			req:     moveReq(str(testutil.UUID3), nil),
			store:   &hierarchyStoreFake{inSubtree: true},
			nodes:   nodeSetOf(moving, parent),
			wantErr: hierarchy.ErrNodeCannotBeADescendantOfItself,
		},
		{
			name:    "before is the node itself",
			req:     moveReq(nil, str(testutil.UUID2)),
			nodes:   nodeSetOf(moving),
			wantErr: hierarchy.ErrBeforeNotSibling,
		},
		{
			name:    "before not found",
			req:     moveReq(str(testutil.UUID3), str(testutil.UUID4)),
			nodes:   nodeSetOf(moving, parent),
			wantErr: hierarchy.ErrBeforeNotSibling,
		},
		{
			name:    "before under another parent",
			req:     moveReq(nil, str(testutil.UUID4)),
			nodes:   nodeSetOf(moving, sibling),
			wantErr: hierarchy.ErrBeforeNotSibling,
		},
		{
			name:  "before a sibling",
			req:   moveReq(str(testutil.UUID3), str(testutil.UUID4)),
			store: &hierarchyStoreFake{prevRanks: []*int64{int64Ptr(1000)}},
			nodes: nodeSetOf(moving, parent, sibling),
			check: func(t *testing.T, resp hierarchy.MoveNodeResponse, store *hierarchyStoreFake) {
				t.Helper()
				require.Equal(t, testutil.UUID3, *resp.ParentID)
				require.EqualValues(t, 1500, resp.SortOrder)
				require.Equal(t, []pgtype.UUID{userID}, store.lockCalls)
				require.Equal(t, []sqlcHierarchy.GetPrevSiblingRankParams{{
					ParentID:   parentID,
					UserID:     userID,
					ExcludeID:  nodeID,
					BeforeRank: pgtype.Int8{Int64: 2000, Valid: true},
					BeforeID:   siblingID,
				}}, store.prevRankCalls)
				require.Equal(t, []sqlcHierarchy.MoveNodeParams{{ID: nodeID, UserID: userID, ParentID: parentID, SortOrder: 1500}}, store.moveCalls)
				require.Len(t, store.inSubtreeCalls, 1, "a new parent must be checked for cycles")
			},
		},
		{
			name:  "first in list",
			req:   moveReq(str(testutil.UUID3), str(testutil.UUID4)),
			nodes: nodeSetOf(moving, parent, sibling),
			check: func(t *testing.T, resp hierarchy.MoveNodeResponse, _ *hierarchyStoreFake) {
				t.Helper()
				require.Equal(t, 2000-rank.Gap, resp.SortOrder)
			},
		},
		{
			name:  "end of list",
			req:   moveReq(str(testutil.UUID3), nil),
			store: &hierarchyStoreFake{prevRanks: []*int64{int64Ptr(5000)}},
			nodes: nodeSetOf(moving, parent),
			check: func(t *testing.T, resp hierarchy.MoveNodeResponse, store *hierarchyStoreFake) {
				t.Helper()
				require.Equal(t, 5000+rank.Gap, resp.SortOrder)
				require.False(t, store.prevRankCalls[0].BeforeRank.Valid, "without before_id the last sibling is looked up")
			},
		},
		{
			name:  "into empty parent",
			req:   moveReq(str(testutil.UUID3), nil),
			nodes: nodeSetOf(moving, parent),
			check: func(t *testing.T, resp hierarchy.MoveNodeResponse, _ *hierarchyStoreFake) {
				t.Helper()
				require.Equal(t, rank.Gap, resp.SortOrder)
			},
		},
		{
			name:  "to root skips parent checks",
			req:   moveReq(nil, nil),
			nodes: nodeSetOf(moving),
			check: func(t *testing.T, resp hierarchy.MoveNodeResponse, store *hierarchyStoreFake) {
				t.Helper()
				require.Nil(t, resp.ParentID)
				require.Empty(t, store.inSubtreeCalls)
				require.False(t, store.moveCalls[0].ParentID.Valid)
			},
		},
		{
			name:  "reorder within same parent skips parent checks",
			req:   moveReq(str(testutil.UUID3), str(testutil.UUID4)),
			store: &hierarchyStoreFake{prevRanks: []*int64{int64Ptr(1000)}},
			nodes: nodeSetOf(sqlcHierarchy.Node{ID: nodeID, UserID: userID, ParentID: parentID}, sibling),
			check: func(t *testing.T, resp hierarchy.MoveNodeResponse, store *hierarchyStoreFake) {
				t.Helper()
				require.EqualValues(t, 1500, resp.SortOrder)
				require.Empty(t, store.inSubtreeCalls)
			},
		},
		{
			name:  "no gap rebalances siblings",
			req:   moveReq(str(testutil.UUID3), str(testutil.UUID4)),
			store: &hierarchyStoreFake{prevRanks: []*int64{int64Ptr(1999), int64Ptr(0)}},
			nodes: nodeSetOf(moving, parent, sibling),
			check: func(t *testing.T, resp hierarchy.MoveNodeResponse, store *hierarchyStoreFake) {
				t.Helper()
				require.Equal(t, []sqlcHierarchy.RebalanceChildrenParams{{
					Gap:       rank.Gap,
					ParentID:  parentID,
					UserID:    userID,
					ExcludeID: nodeID,
				}}, store.rebalanceCalls)
				require.Len(t, store.prevRankCalls, 2, "neighbours are read again after rebalance")
				require.EqualValues(t, 1000, resp.SortOrder)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.store
			if store == nil {
				store = &hierarchyStoreFake{}
			}
			nodes := tt.nodes
			if nodes == nil {
				nodes = &nodeSet{}
			}

			resp, err := newTestService(store, nodes).MoveNode(context.Background(), nodeID, userID, tt.req)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Empty(t, store.moveCalls, "a failed move must not write")
				return
			}
			require.NoError(t, err)
			tt.check(t, resp, store)
		})
	}
}

func TestMoveNode_NoGapAfterRebalance(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID3)
	store := &hierarchyStoreFake{prevRanks: []*int64{int64Ptr(1999), int64Ptr(1999)}}
	nodes := nodeSetOf(
		sqlcHierarchy.Node{ID: nodeID, UserID: userID, ParentID: parentID},
		sqlcHierarchy.Node{ID: testutil.UUIDFromStringT(t, testutil.UUID4), UserID: userID, ParentID: parentID, SortOrder: 2000},
	)
	req := &hierarchy.MoveNodeRequest{
		ParentID: hierarchy.NullableString{Value: testutil.StringPtr(testutil.UUID3), IsSet: true},
		BeforeID: hierarchy.NullableString{Value: testutil.StringPtr(testutil.UUID4), IsSet: true},
	}

	_, err := newTestService(store, nodes).MoveNode(context.Background(), nodeID, userID, req)

	require.ErrorIs(t, err, hierarchy.ErrNoRankAfterRebalance)
	require.Len(t, store.rebalanceCalls, 1, "rebalance is attempted only once")
	require.Empty(t, store.moveCalls)
}
