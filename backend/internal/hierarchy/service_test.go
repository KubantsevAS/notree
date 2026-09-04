package hierarchy_test

import (
	"context"
	"testing"

	hierarchyDb "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	nodeDb "github.com/KubantsevAS/notree/backend/internal/db/node"
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
	children    []hierarchyDb.Node
	childrenErr error
}

func (f *hierarchyStoreFake) GetChildren(context.Context, hierarchyDb.GetChildrenParams) ([]hierarchyDb.Node, error) {
	if f.childrenErr != nil {
		return nil, f.childrenErr
	}
	return f.children, nil
}

func (f *hierarchyStoreFake) GetParent(context.Context, hierarchyDb.GetParentParams) (hierarchyDb.Node, error) {
	return hierarchyDb.Node{}, f.childrenErr
}

type nodeStoreFake struct {
	getNodeByIDResult map[string]nodeDb.Node
	getNodeByIDErr    error
	getNodeByIDCalls  []nodeDb.GetNodeByIDParams
}

func (f *nodeStoreFake) GetNodeByID(_ context.Context, params nodeDb.GetNodeByIDParams) (nodeDb.Node, error) {
	f.getNodeByIDCalls = append(f.getNodeByIDCalls, params)
	if f.getNodeByIDErr != nil {
		return nodeDb.Node{}, f.getNodeByIDErr
	}
	if f.getNodeByIDResult != nil {
		if result, ok := f.getNodeByIDResult[params.ID.String()]; ok {
			return result, nil
		}
	}
	return nodeDb.Node{}, pgx.ErrNoRows
}

func TestNodeServiceGetChildrenReturnsChildrenInSortOrder(t *testing.T) {
	parentID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	userID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	childOneID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	childTwoID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}
	fakeNodeStore := &nodeStoreFake{
		getNodeByIDResult: map[string]nodeDb.Node{parentID.String(): {ID: parentID, UserID: userID}},
	}
	fake := &hierarchyStoreFake{
		children: []hierarchyDb.Node{
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

func TestNodeServiceGetChildrenReturns(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	parentID := testutil.UUIDFromStringT(t, testUUID2)

	_, err := hierarchy.NewService(&hierarchyStoreFake{}, &nodeStoreFake{}).GetChildren(context.Background(), parentID, userID)
	require.ErrorIs(t, err, hierarchy.ErrParentNotFound)
}
