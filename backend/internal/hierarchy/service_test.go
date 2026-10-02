package hierarchy_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	sqlcNode "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

const (
	testUUIDBad = "bad-uuid"
	testUUID1   = "11111111-1111-4111-8111-111111111111"
	testUUID2   = "22222222-2222-4222-8222-222222222222"
	testUUID3   = "33333333-3333-4333-8333-333333333333"
	testUUID4   = "44444444-4444-4444-8444-444444444444"
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

func TestGetChildren(t *testing.T) {
	parentID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	userID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	childOneID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	childTwoID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}
	fakeNodeStore := &nodeStoreFake{
		getNodeByIDResult: map[string]sqlcNode.Node{parentID.String(): {ID: parentID, UserID: userID}},
	}
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

func TestGetChildren_NodeNotFound(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	nodeID := testutil.UUIDFromStringT(t, testUUID2)

	service := hierarchy.NewService(&hierarchyStoreFake{}, &nodeStoreFake{})
	_, err := service.GetChildren(context.Background(), nodeID, userID)
	require.ErrorIs(t, err, hierarchy.ErrNodeNotFound)
}

func TestGetParent(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	nodeID := testutil.UUIDFromStringT(t, testUUID2)
	parentID := testutil.UUIDFromStringT(t, testUUID3)
	grandparentID := testutil.UUIDFromStringT(t, testUUID4)

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

	nodeStore := &nodeStoreFake{getNodeByIDResult: map[string]sqlcNode.Node{nodeID.String(): node}}
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

func TestGetParent_NodeNotFound(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	nodeID := testutil.UUIDFromStringT(t, testUUID2)

	service := hierarchy.NewService(&hierarchyStoreFake{}, &nodeStoreFake{})
	_, err := service.GetParent(context.Background(), nodeID, userID)

	require.ErrorIs(t, err, hierarchy.ErrNodeNotFound)
}

func TestGetParent_NodeIsRoot(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	rootID := testutil.UUIDFromStringT(t, testUUID2)

	root := sqlcNode.Node{ID: rootID, UserID: userID}
	nodeStore := &nodeStoreFake{getNodeByIDResult: map[string]sqlcNode.Node{rootID.String(): root}}
	store := &hierarchyStoreFake{parentErr: errors.New("GetParent must not be called for root node")}

	service := hierarchy.NewService(store, nodeStore)
	_, err := service.GetParent(context.Background(), rootID, userID)

	require.ErrorIs(t, err, hierarchy.ErrNodeIsRoot)
	require.Zero(t, store.getParentCalls)
}

func TestGetAncestors(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	nodeID := testutil.UUIDFromStringT(t, testUUID2)
	parentID := testutil.UUIDFromStringT(t, testUUID3)
	rootID := testutil.UUIDFromStringT(t, testUUID4)

	nodeStore := &nodeStoreFake{
		getNodeByIDResult: map[string]sqlcNode.Node{nodeID.String(): {ID: nodeID, UserID: userID, ParentID: parentID}},
	}
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
	userID := testutil.UUIDFromStringT(t, testUUID1)
	rootID := testutil.UUIDFromStringT(t, testUUID2)

	nodeStore := &nodeStoreFake{
		getNodeByIDResult: map[string]sqlcNode.Node{rootID.String(): {ID: rootID, UserID: userID}},
	}

	service := hierarchy.NewService(&hierarchyStoreFake{}, nodeStore)
	res, err := service.GetAncestors(context.Background(), rootID, userID)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Empty(t, res)
}

func TestGetAncestors_NodeNotFound(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	nodeID := testutil.UUIDFromStringT(t, testUUID2)

	service := hierarchy.NewService(&hierarchyStoreFake{}, &nodeStoreFake{})
	_, err := service.GetAncestors(context.Background(), nodeID, userID)

	require.ErrorIs(t, err, hierarchy.ErrNodeNotFound)
}

func TestGetDescendants(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	rootID := testutil.UUIDFromStringT(t, testUUID2)
	childID := testutil.UUIDFromStringT(t, testUUID3)
	grandchildID := testutil.UUIDFromStringT(t, testUUID4)

	nodeStore := &nodeStoreFake{
		getNodeByIDResult: map[string]sqlcNode.Node{rootID.String(): {ID: rootID, UserID: userID}},
	}
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
	userID := testutil.UUIDFromStringT(t, testUUID1)
	nodeID := testutil.UUIDFromStringT(t, testUUID2)

	nodeStore := &nodeStoreFake{
		getNodeByIDResult: map[string]sqlcNode.Node{nodeID.String(): {ID: nodeID, UserID: userID}},
	}

	service := hierarchy.NewService(&hierarchyStoreFake{}, nodeStore)
	res, err := service.GetDescendants(context.Background(), nodeID, userID)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Empty(t, res)
}

func TestGetDescendants_NodeNotFound(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	nodeID := testutil.UUIDFromStringT(t, testUUID2)

	service := hierarchy.NewService(&hierarchyStoreFake{}, &nodeStoreFake{})
	_, err := service.GetDescendants(context.Background(), nodeID, userID)

	require.ErrorIs(t, err, hierarchy.ErrNodeNotFound)
}
