package hierarchy_test

import (
	"context"
	"testing"

	hierarchyDB "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	nodeDB "github.com/KubantsevAS/notree/backend/internal/db/node"
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
)

type hierarchyStoreFake struct {
	children    []hierarchyDB.Node
	childrenErr error
	parent      hierarchyDB.Node
	parentErr   error
}

func (f *hierarchyStoreFake) GetChildren(context.Context, hierarchyDB.GetChildrenParams) ([]hierarchyDB.Node, error) {
	if f.childrenErr != nil {
		return []hierarchyDB.Node{}, f.childrenErr
	}
	return f.children, nil
}

func (f *hierarchyStoreFake) GetParent(context.Context, hierarchyDB.GetParentParams) (hierarchyDB.Node, error) {
	return hierarchyDB.Node{}, f.parentErr
}

type nodeStoreFake struct {
	getNodeByIDResult map[string]nodeDB.Node
	getNodeByIDErr    error
}

func (f *nodeStoreFake) GetNodeByID(_ context.Context, params nodeDB.GetNodeByIDParams) (nodeDB.Node, error) {
	if f.getNodeByIDErr != nil {
		return nodeDB.Node{}, f.getNodeByIDErr
	}
	if f.getNodeByIDResult != nil {
		if result, ok := f.getNodeByIDResult[params.ID.String()]; ok {
			return result, nil
		}
	}
	return nodeDB.Node{}, pgx.ErrNoRows
}

func TestHierarchyServiceGetChildrenReturnsChildrenInSortOrder(t *testing.T) {
	parentID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	userID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	childOneID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	childTwoID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}
	fakeNodeStore := &nodeStoreFake{
		getNodeByIDResult: map[string]nodeDB.Node{parentID.String(): {ID: parentID, UserID: userID}},
	}
	fake := &hierarchyStoreFake{
		children: []hierarchyDB.Node{
			{ID: childOneID, ParentID: parentID, UserID: userID, Title: "first", SortOrder: 10},
			{ID: childTwoID, ParentID: parentID, UserID: userID, Title: "second", SortOrder: 20},
		},
	}

	children, err := hierarchy.NewService(fake, fakeNodeStore).GetChildren(context.Background(), parentID, userID)
	require.NoError(t, err)
	require.Len(t, children, 2)
	require.Equal(t, childOneID.String(), children[0].ID)
	require.Equal(t, childTwoID.String(), children[1].ID)
}

func TestHierarchyServiceGetChildrenReturnsErrParentNotFoundWhenParentMissing(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	parentID := testutil.UUIDFromStringT(t, testUUID2)

	_, err := hierarchy.NewService(&hierarchyStoreFake{}, &nodeStoreFake{}).GetChildren(context.Background(), parentID, userID)
	require.ErrorIs(t, err, hierarchy.ErrParentNotFound)
}
