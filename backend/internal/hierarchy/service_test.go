package hierarchy_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	sqlcNode "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/domain"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy"
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
			err := tt.call(hierarchy.NewService(&hierarchyStoreFake{}, &nodeStoreFake{}))
			require.ErrorIs(t, err, hierarchy.ErrNodeNotFound)
		})
	}
}

func TestGetChildren(t *testing.T) {
	parentID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	userID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	childOneID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	childTwoID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}
	fakeNodeStore := nodeStoreWith(sqlcNode.Node{ID: parentID, UserID: userID})
	fake := &hierarchyStoreFake{
		children: []sqlcHierarchy.Node{
			{ID: childOneID, ParentID: parentID, UserID: userID, Title: "first", SortOrder: 10},
			{ID: childTwoID, ParentID: parentID, UserID: userID, Title: "second", SortOrder: 20},
		},
	}

	service := hierarchy.NewService(fake, fakeNodeStore)
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

	node := sqlcNode.Node{ID: nodeID, UserID: userID, ParentID: parentID}

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

	nodeStore := nodeStoreWith(node)
	store := &hierarchyStoreFake{parent: parent}

	service := hierarchy.NewService(store, nodeStore)
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

	root := sqlcNode.Node{ID: rootID, UserID: userID}
	nodeStore := nodeStoreWith(root)
	store := &hierarchyStoreFake{parentErr: errors.New("GetParent must not be called for root node")}

	service := hierarchy.NewService(store, nodeStore)
	_, err := service.GetParent(context.Background(), rootID, userID)

	require.ErrorIs(t, err, hierarchy.ErrNodeIsRoot)
	require.Zero(t, store.getParentCalls)
}

func TestGetAncestors(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID3)
	rootID := testutil.UUIDFromStringT(t, testutil.UUID4)

	nodeStore := nodeStoreWith(sqlcNode.Node{ID: nodeID, UserID: userID, ParentID: parentID})
	store := &hierarchyStoreFake{
		ancestors: []sqlcHierarchy.Node{
			{ID: rootID, UserID: userID, Title: "root"},
			{ID: parentID, ParentID: rootID, UserID: userID, Title: "parent"},
		},
	}

	service := hierarchy.NewService(store, nodeStore)
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

	nodeStore := nodeStoreWith(sqlcNode.Node{ID: rootID, UserID: userID})

	service := hierarchy.NewService(&hierarchyStoreFake{}, nodeStore)
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

	nodeStore := nodeStoreWith(sqlcNode.Node{ID: rootID, UserID: userID})
	store := &hierarchyStoreFake{
		descendants: []sqlcHierarchy.Node{
			{ID: childID, ParentID: rootID, UserID: userID, Title: "child"},
			{ID: grandchildID, ParentID: childID, UserID: userID, Title: "grandchild"},
		},
	}

	service := hierarchy.NewService(store, nodeStore)
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

	nodeStore := nodeStoreWith(sqlcNode.Node{ID: nodeID, UserID: userID})

	service := hierarchy.NewService(&hierarchyStoreFake{}, nodeStore)
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

	nodeStore := nodeStoreWith(sqlcNode.Node{ID: nodeID, UserID: userID})
	store := &hierarchyStoreFake{
		subtree: []sqlcHierarchy.Node{
			{ID: nodeID, UserID: userID, Title: "node"},
			{ID: childID, ParentID: nodeID, UserID: userID, Title: "child"},
			{ID: grandchildID, ParentID: childID, UserID: userID, Title: "grandchild"},
		},
	}

	service := hierarchy.NewService(store, nodeStore)
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

	nodeStore := nodeStoreWith(sqlcNode.Node{ID: nodeID, UserID: userID, ParentID: parentID})
	store := &hierarchyStoreFake{
		root: sqlcHierarchy.Node{ID: rootID, UserID: userID, Title: "root"},
	}

	service := hierarchy.NewService(store, nodeStore)
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

	nodeStore := nodeStoreWith(sqlcNode.Node{ID: nodeID, UserID: userID})
	store := &hierarchyStoreFake{rootErr: pgx.ErrNoRows}

	service := hierarchy.NewService(store, nodeStore)
	_, err := service.GetRoot(context.Background(), nodeID, userID)

	require.ErrorIs(t, err, hierarchy.ErrRootNotFound)
}

func TestMoveNode(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID3)
	parentExists := func() *nodeStoreFake {
		return nodeStoreWith(sqlcNode.Node{ID: parentID, UserID: userID})
	}
	moveTo := func(id string) *hierarchy.MoveNodeRequest {
		return &hierarchy.MoveNodeRequest{ParentID: hierarchy.NullableString{Value: testutil.StringPtr(id), IsSet: true}}
	}

	tests := []struct {
		name      string
		req       *hierarchy.MoveNodeRequest
		store     *hierarchyStoreFake
		nodeStore *nodeStoreFake
		wantErr   error
		check     func(*testing.T, hierarchy.MoveNodeResponse, *hierarchyStoreFake)
	}{
		{
			name:    "empty request",
			req:     &hierarchy.MoveNodeRequest{},
			wantErr: domain.ErrEmptyUpdate,
		},
		{
			name:    "invalid parent uuid",
			req:     moveTo(testutil.BadUUID),
			wantErr: hierarchy.ErrInvalidParentID,
		},
		{
			name:    "empty parent uuid",
			req:     moveTo(""),
			wantErr: hierarchy.ErrInvalidParentID,
		},
		{
			name:    "self parent",
			req:     moveTo(testutil.UUID2),
			wantErr: hierarchy.ErrNodeCannotBeADescendantOfItself,
		},
		{
			name:    "parent not found",
			req:     moveTo(testutil.UUID3),
			wantErr: hierarchy.ErrParentNotFound,
		},
		{
			name:      "parent is inside node subtree",
			req:       moveTo(testutil.UUID3),
			store:     &hierarchyStoreFake{inSubtree: true},
			nodeStore: parentExists(),
			wantErr:   hierarchy.ErrNodeCannotBeADescendantOfItself,
		},
		{
			name:      "subtree check error",
			req:       moveTo(testutil.UUID3),
			store:     &hierarchyStoreFake{inSubtreeErr: sql.ErrConnDone},
			nodeStore: parentExists(),
			wantErr:   sql.ErrConnDone,
		},
		{
			name:      "node not found",
			req:       moveTo(testutil.UUID3),
			store:     &hierarchyStoreFake{moveErr: pgx.ErrNoRows},
			nodeStore: parentExists(),
			wantErr:   hierarchy.ErrNodeNotFound,
		},
		{
			name: "success",
			req: &hierarchy.MoveNodeRequest{
				ParentID:  hierarchy.NullableString{Value: testutil.StringPtr(testutil.UUID3), IsSet: true},
				SortOrder: testutil.Int64Ptr(42),
			},
			store: &hierarchyStoreFake{moveResult: sqlcHierarchy.MoveNodeRow{
				ParentID:  parentID,
				SortOrder: 42,
				UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}},
			nodeStore: parentExists(),
			check: func(t *testing.T, resp hierarchy.MoveNodeResponse, store *hierarchyStoreFake) {
				t.Helper()
				require.NotNil(t, resp.ParentID)
				require.Equal(t, testutil.UUID3, *resp.ParentID)
				require.EqualValues(t, 42, resp.SortOrder)

				require.Len(t, store.inSubtreeCalls, 1)
				require.Equal(t, sqlcHierarchy.IsInSubtreeParams{RootID: nodeID, NodeID: parentID, UserID: userID}, store.inSubtreeCalls[0])

				require.Len(t, store.moveCalls, 1)
				require.True(t, store.moveCalls[0].UpdateParent.Bool)
				require.Equal(t, parentID, store.moveCalls[0].ParentID)
			},
		},
		{
			name: "move to root skips parent checks",
			req:  &hierarchy.MoveNodeRequest{ParentID: hierarchy.NullableString{IsSet: true}},
			store: &hierarchyStoreFake{moveResult: sqlcHierarchy.MoveNodeRow{
				UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}},
			check: func(t *testing.T, resp hierarchy.MoveNodeResponse, store *hierarchyStoreFake) {
				t.Helper()
				require.Nil(t, resp.ParentID)
				require.Empty(t, store.inSubtreeCalls)
				require.True(t, store.moveCalls[0].UpdateParent.Bool)
				require.False(t, store.moveCalls[0].ParentID.Valid)
			},
		},
		{
			name: "reorder only keeps parent",
			req:  &hierarchy.MoveNodeRequest{SortOrder: testutil.Int64Ptr(7)},
			store: &hierarchyStoreFake{moveResult: sqlcHierarchy.MoveNodeRow{
				SortOrder: 7,
				UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}},
			check: func(t *testing.T, resp hierarchy.MoveNodeResponse, store *hierarchyStoreFake) {
				t.Helper()
				require.EqualValues(t, 7, resp.SortOrder)
				require.Empty(t, store.inSubtreeCalls)
				require.False(t, store.moveCalls[0].UpdateParent.Bool)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.store
			if store == nil {
				store = &hierarchyStoreFake{}
			}
			nodeStore := tt.nodeStore
			if nodeStore == nil {
				nodeStore = &nodeStoreFake{}
			}

			resp, err := hierarchy.NewService(store, nodeStore).MoveNode(context.Background(), nodeID, userID, tt.req)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				if store.moveErr == nil {
					require.Empty(t, store.moveCalls, "validation errors must not reach MoveNode")
				}
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, resp, store)
			}
		})
	}
}
