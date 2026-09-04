package node_test

import (
	"context"
	"errors"
	"testing"
	"time"

	nodeDb "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/domain"
	"github.com/KubantsevAS/notree/backend/internal/node"
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

type nodeStoreFake struct {
	createParams      []nodeDb.CreateNodeParams
	createResult      nodeDb.Node
	createErr         error
	getNodeByIDResult map[string]nodeDb.Node
	getNodeByIDErr    error
	getNodeByIDCalls  []nodeDb.GetNodeByIDParams
	ancestors         map[string][]pgtype.UUID
	ancestorsErr      error
	softDeleteParams  []nodeDb.SoftDeleteNodeCascadeParams
	softDeleteResult  []pgtype.UUID
	softDeleteErr     error
	updateParams      []nodeDb.UpdateNodeParams
	updateResult      nodeDb.UpdateNodeRow
	updateErr         error
	moveParams        []nodeDb.MoveNodeParams
	moveResult        nodeDb.MoveNodeRow
	moveErr           error
}

func (f *nodeStoreFake) CreateNode(_ context.Context, params nodeDb.CreateNodeParams) (nodeDb.Node, error) {
	f.createParams = append(f.createParams, params)
	if f.createErr != nil {
		return nodeDb.Node{}, f.createErr
	}
	if f.createResult.ID.Valid {
		return f.createResult, nil
	}

	return nodeDb.Node{
		ID:        pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		ParentID:  params.ParentID,
		Type:      params.Type,
		Title:     params.Title,
		SortOrder: params.SortOrder,
	}, nil
}

func (f *nodeStoreFake) GetNodeAncestors(_ context.Context, parentID pgtype.UUID) ([]pgtype.UUID, error) {
	if f.ancestorsErr != nil {
		return nil, f.ancestorsErr
	}
	if f.ancestors != nil {
		if values, ok := f.ancestors[parentID.String()]; ok {
			return values, nil
		}
	}
	return nil, nil
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

func (f *nodeStoreFake) MoveNode(_ context.Context, params nodeDb.MoveNodeParams) (nodeDb.MoveNodeRow, error) {
	f.moveParams = append(f.moveParams, params)
	if f.moveErr != nil {
		return nodeDb.MoveNodeRow{}, f.moveErr
	}
	return f.moveResult, nil
}

func (f *nodeStoreFake) SoftDeleteNodeCascade(_ context.Context, params nodeDb.SoftDeleteNodeCascadeParams) ([]pgtype.UUID, error) {
	f.softDeleteParams = append(f.softDeleteParams, params)
	if f.softDeleteErr != nil {
		return nil, f.softDeleteErr
	}
	return f.softDeleteResult, nil
}

func (f *nodeStoreFake) UpdateNode(_ context.Context, params nodeDb.UpdateNodeParams) (nodeDb.UpdateNodeRow, error) {
	f.updateParams = append(f.updateParams, params)
	if f.updateErr != nil {
		return nodeDb.UpdateNodeRow{}, f.updateErr
	}
	return f.updateResult, nil
}

func TestNodeServiceCreateNodeSortOrderIncreases(t *testing.T) {
	fake := &nodeStoreFake{}
	service := node.NewService(fake)
	request := &node.CreateNodeRequest{Type: "note", Title: "test"}

	_, err := service.CreateNode(context.Background(), pgtype.UUID{}, request)
	require.NoError(t, err)

	_, err = service.CreateNode(context.Background(), pgtype.UUID{}, request)
	require.NoError(t, err)

	require.Len(t, fake.createParams, 2)
	require.Greater(t, fake.createParams[1].SortOrder, fake.createParams[0].SortOrder)
}

func TestNodeServiceCreateNodeValidatesParentID(t *testing.T) {
	var errTimeout = errors.New("timeout")
	tests := []struct {
		name    string
		req     *node.CreateNodeRequest
		fake    *nodeStoreFake
		wantErr error
	}{
		{
			name:    "invalid parent uuid",
			req:     &node.CreateNodeRequest{ParentID: testutil.StringPtr(testUUIDBad), Type: "note", Title: "child"},
			fake:    &nodeStoreFake{},
			wantErr: node.ErrInvalidParentID,
		},
		{
			name:    "parent not found",
			req:     &node.CreateNodeRequest{ParentID: testutil.StringPtr(testUUID1), Type: "note", Title: "child"},
			fake:    &nodeStoreFake{},
			wantErr: node.ErrParentNotFound,
		},
		{
			name:    "db error on parent lookup",
			req:     &node.CreateNodeRequest{ParentID: testutil.StringPtr(testUUID1), Type: "note", Title: "child"},
			fake:    &nodeStoreFake{getNodeByIDErr: errTimeout},
			wantErr: errTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := node.NewService(tt.fake).CreateNode(context.Background(), pgtype.UUID{}, tt.req)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNodeServiceDeleteNode(t *testing.T) {
	tests := []struct {
		name    string
		nodeID  pgtype.UUID
		userID  pgtype.UUID
		fake    *nodeStoreFake
		wantErr error
		check   func(*testing.T, *nodeStoreFake)
	}{
		{
			name:   "success",
			nodeID: testutil.UUIDFromStringT(t, testUUID3),
			userID: testutil.UUIDFromStringT(t, testUUID4),
			fake: &nodeStoreFake{
				softDeleteResult: []pgtype.UUID{testutil.UUIDFromStringT(t, testUUID3)},
			},
			check: func(t *testing.T, fake *nodeStoreFake) {
				t.Helper()
				require.Len(t, fake.softDeleteParams, 1)
			},
		},
		{
			name:    "not found or no access",
			nodeID:  testutil.UUIDFromStringT(t, testUUID2),
			userID:  testutil.UUIDFromStringT(t, testUUID1),
			fake:    &nodeStoreFake{},
			wantErr: node.ErrNodeNotFoundOrNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := tt.fake
			if fake == nil {
				fake = &nodeStoreFake{}
			}

			err := node.NewService(fake).DeleteNode(context.Background(), tt.nodeID, tt.userID)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, fake)
			}
		})
	}
}

func TestNodeServiceUpdateNode(t *testing.T) {
	tests := []struct {
		name    string
		nodeID  pgtype.UUID
		userID  pgtype.UUID
		req     *node.UpdateNodeRequest
		fake    *nodeStoreFake
		wantErr error
		check   func(*testing.T, node.UpdateNodeResponse)
	}{
		{
			name:    "empty request",
			req:     &node.UpdateNodeRequest{},
			wantErr: domain.ErrEmptyUpdate,
		},
		{
			name:   "success",
			nodeID: testutil.UUIDFromStringT(t, testUUID3),
			userID: testutil.UUIDFromStringT(t, testUUID4),
			fake: &nodeStoreFake{
				updateResult: nodeDb.UpdateNodeRow{
					Type:      nodeDb.NodeTypeTask,
					Title:     "done",
					UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				},
			},
			req: &node.UpdateNodeRequest{Type: testutil.StringPtr("task"), Title: testutil.StringPtr("done")},
			check: func(t *testing.T, resp node.UpdateNodeResponse) {
				t.Helper()
				require.Equal(t, "task", resp.Type)
				require.Equal(t, "done", resp.Title)
			},
		},
		{
			name:    "node not found or no access",
			fake:    &nodeStoreFake{updateErr: pgx.ErrNoRows},
			req:     &node.UpdateNodeRequest{Title: testutil.StringPtr("new title")},
			wantErr: node.ErrNodeNotFoundOrNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := tt.fake
			if fake == nil {
				fake = &nodeStoreFake{}
			}

			resp, err := node.NewService(fake).UpdateNode(context.Background(), tt.nodeID, tt.userID, tt.req)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, resp)
			}
		})
	}
}

func TestNodeServiceMoveNode(t *testing.T) {
	tests := []struct {
		name    string
		nodeID  pgtype.UUID
		userID  pgtype.UUID
		req     *node.MoveNodeRequest
		fake    *nodeStoreFake
		wantErr error
		check   func(*testing.T, node.MoveNodeResponse)
	}{
		{
			name:    "empty request",
			req:     &node.MoveNodeRequest{},
			wantErr: domain.ErrEmptyUpdate,
		},
		{
			name:    "invalid parent uuid",
			req:     &node.MoveNodeRequest{ParentID: node.NullableString{Value: testutil.StringPtr("bad-uuid"), IsSet: true}},
			wantErr: node.ErrInvalidParentID,
		},
		{
			name:    "node cannot be a descendant of itself when parent is same id",
			nodeID:  testutil.UUIDFromStringT(t, testUUID1),
			req:     &node.MoveNodeRequest{ParentID: node.NullableString{Value: testutil.StringPtr(testUUID1), IsSet: true}},
			wantErr: node.ErrNodeCannotBeADescendantOfItself,
		},
		{
			name:   "node cannot be a descendant of itself when ancestor contains node",
			nodeID: testutil.UUIDFromStringT(t, testUUID1),
			fake: &nodeStoreFake{
				getNodeByIDResult: map[string]nodeDb.Node{
					testUUID2: {ID: testutil.UUIDFromStringT(t, testUUID2)},
				},
				ancestors: map[string][]pgtype.UUID{
					testUUID2: {testutil.UUIDFromStringT(t, testUUID1)},
				},
			},
			req:     &node.MoveNodeRequest{ParentID: node.NullableString{Value: testutil.StringPtr(testUUID2), IsSet: true}},
			wantErr: node.ErrNodeCannotBeADescendantOfItself,
		},
		{
			name:   "success",
			nodeID: testutil.UUIDFromStringT(t, testUUID3),
			userID: testutil.UUIDFromStringT(t, testUUID4),
			fake: &nodeStoreFake{
				getNodeByIDResult: map[string]nodeDb.Node{
					testUUID2: {ID: testutil.UUIDFromStringT(t, testUUID2), UserID: testutil.UUIDFromStringT(t, testUUID4)},
				},
				moveResult: nodeDb.MoveNodeRow{
					ParentID:  testutil.UUIDFromStringT(t, testUUID2),
					SortOrder: 42,
					UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				},
			},
			req: &node.MoveNodeRequest{
				ParentID:  node.NullableString{Value: testutil.StringPtr(testUUID2), IsSet: true},
				SortOrder: testutil.Int64Ptr(42),
			},
			check: func(t *testing.T, resp node.MoveNodeResponse) {
				t.Helper()
				require.NotNil(t, resp.ParentID)
				require.Equal(t, testUUID2, *resp.ParentID)
				require.EqualValues(t, 42, resp.SortOrder)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := tt.fake
			if fake == nil {
				fake = &nodeStoreFake{}
			}

			resp, err := node.NewService(fake).MoveNode(context.Background(), tt.nodeID, tt.userID, tt.req)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, resp)
			}
		})
	}
}

func TestNodeServiceCreateNodeAddsParentAndSortOrder(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	parentID := testutil.UUIDFromStringT(t, testUUID2)
	fake := &nodeStoreFake{
		getNodeByIDResult: map[string]nodeDb.Node{testUUID2: {ID: parentID, UserID: userID}},
		createResult: nodeDb.Node{
			ID:        testutil.UUIDFromStringT(t, testUUID3),
			ParentID:  parentID,
			Type:      nodeDb.NodeTypeNote,
			Title:     "child",
			SortOrder: 123,
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		},
	}

	resp, err := node.NewService(fake).CreateNode(context.Background(), userID, &node.CreateNodeRequest{
		ParentID: testutil.StringPtr(testUUID2),
		Type:     "note",
		Title:    "child",
	})
	require.NoError(t, err)
	require.Equal(t, testUUID2, resp.ParentID)
	require.Equal(t, "note", resp.Type)
	require.Equal(t, "child", resp.Title)
	require.Len(t, fake.createParams, 1)
}
