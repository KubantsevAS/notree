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

type nodeRepositoryFake struct {
	createParams      []nodeDb.CreateNodeParams
	createResult      nodeDb.Node
	createErr         error
	getNodeByIDResult map[string]nodeDb.Node
	getNodeByIDErr    error
	getNodeByIDCalls  []nodeDb.GetNodeByIDParams
	children          []nodeDb.Node
	childrenErr       error
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

func (f *nodeRepositoryFake) CreateNode(_ context.Context, params nodeDb.CreateNodeParams) (nodeDb.Node, error) {
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

func (f *nodeRepositoryFake) GetChildren(context.Context, nodeDb.GetChildrenParams) ([]nodeDb.Node, error) {
	if f.childrenErr != nil {
		return nil, f.childrenErr
	}
	return f.children, nil
}

func (f *nodeRepositoryFake) GetNodeAncestors(_ context.Context, parentID pgtype.UUID) ([]pgtype.UUID, error) {
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

func (f *nodeRepositoryFake) GetNodeByID(_ context.Context, params nodeDb.GetNodeByIDParams) (nodeDb.Node, error) {
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

func (f *nodeRepositoryFake) MoveNode(_ context.Context, params nodeDb.MoveNodeParams) (nodeDb.MoveNodeRow, error) {
	f.moveParams = append(f.moveParams, params)
	if f.moveErr != nil {
		return nodeDb.MoveNodeRow{}, f.moveErr
	}
	return f.moveResult, nil
}

func (f *nodeRepositoryFake) SoftDeleteNodeCascade(_ context.Context, params nodeDb.SoftDeleteNodeCascadeParams) ([]pgtype.UUID, error) {
	f.softDeleteParams = append(f.softDeleteParams, params)
	if f.softDeleteErr != nil {
		return nil, f.softDeleteErr
	}
	return f.softDeleteResult, nil
}

func (f *nodeRepositoryFake) UpdateNode(_ context.Context, params nodeDb.UpdateNodeParams) (nodeDb.UpdateNodeRow, error) {
	f.updateParams = append(f.updateParams, params)
	if f.updateErr != nil {
		return nodeDb.UpdateNodeRow{}, f.updateErr
	}
	return f.updateResult, nil
}

func TestNodeServiceCreateNodeSortOrderIncreases(t *testing.T) {
	fake := &nodeRepositoryFake{}
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
	var validUUID = "11111111-1111-4111-8111-111111111111"
	var badUUID = "bad-uuid"
	var errTimeout = errors.New("timeout")
	tests := []struct {
		name    string
		req     *node.CreateNodeRequest
		fake    *nodeRepositoryFake
		wantErr error
	}{
		{
			name:    "invalid parent uuid",
			req:     &node.CreateNodeRequest{ParentID: &badUUID, Type: "note", Title: "child"},
			fake:    &nodeRepositoryFake{},
			wantErr: node.ErrInvalidParentID,
		},
		{
			name:    "parent not found",
			req:     &node.CreateNodeRequest{ParentID: &validUUID, Type: "note", Title: "child"},
			fake:    &nodeRepositoryFake{},
			wantErr: node.ErrParentNotFound,
		},
		{
			name:    "db error on parent lookup",
			req:     &node.CreateNodeRequest{ParentID: &validUUID, Type: "note", Title: "child"},
			fake:    &nodeRepositoryFake{getNodeByIDErr: errTimeout},
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

func TestNodeServiceGetChildrenReturnsChildrenInSortOrder(t *testing.T) {
	parentID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	userID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	childOneID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	childTwoID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}
	fake := &nodeRepositoryFake{
		getNodeByIDResult: map[string]nodeDb.Node{parentID.String(): {ID: parentID, UserID: userID}},
		children: []nodeDb.Node{
			{ID: childOneID, ParentID: parentID, UserID: userID, Title: "first", SortOrder: 10},
			{ID: childTwoID, ParentID: parentID, UserID: userID, Title: "second", SortOrder: 20},
		},
	}

	children, err := node.NewService(fake).GetChildren(context.Background(), parentID, userID)
	require.NoError(t, err)
	require.Len(t, children, 2)
	require.Equal(t, childOneID.String(), children[0].ID)
	require.Equal(t, childTwoID.String(), children[1].ID)
}

func TestNodeServiceGetChildrenReturns(t *testing.T) {
	userID := testutil.UUIDFromString("11111111-1111-4111-8111-111111111111")
	parentID := testutil.UUIDFromString("22222222-2222-4222-8222-222222222222")

	_, err := node.NewService(&nodeRepositoryFake{}).GetChildren(context.Background(), parentID, userID)
	require.ErrorIs(t, err, node.ErrParentNotFound)
}

func TestNodeServiceDeleteNode(t *testing.T) {
	tests := []struct {
		name    string
		nodeID  pgtype.UUID
		userID  pgtype.UUID
		fake    *nodeRepositoryFake
		wantErr error
		check   func(*testing.T, *nodeRepositoryFake)
	}{
		{
			name:   "success",
			nodeID: testutil.UUIDFromStringT(t, "33333333-3333-4333-8333-333333333333"),
			userID: testutil.UUIDFromStringT(t, "44444444-4444-4444-8444-444444444444"),
			fake: &nodeRepositoryFake{
				softDeleteResult: []pgtype.UUID{testutil.UUIDFromStringT(t, "33333333-3333-4333-8333-333333333333")},
			},
			check: func(t *testing.T, fake *nodeRepositoryFake) {
				t.Helper()
				require.Len(t, fake.softDeleteParams, 1)
			},
		},
		{
			name:    "not found or no access",
			nodeID:  testutil.UUIDFromStringT(t, "55555555-5555-4555-8555-555555555555"),
			userID:  testutil.UUIDFromStringT(t, "66666666-6666-4666-8666-666666666666"),
			fake:    &nodeRepositoryFake{},
			wantErr: node.ErrNodeNotFoundOrNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := tt.fake
			if fake == nil {
				fake = &nodeRepositoryFake{}
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
		fake    *nodeRepositoryFake
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
			nodeID: testutil.UUIDFromStringT(t, "77777777-7777-4777-8777-777777777777"),
			userID: testutil.UUIDFromStringT(t, "88888888-8888-4888-8888-888888888888"),
			fake: &nodeRepositoryFake{
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
			fake:    &nodeRepositoryFake{updateErr: pgx.ErrNoRows},
			req:     &node.UpdateNodeRequest{Title: testutil.StringPtr("new title")},
			wantErr: node.ErrNodeNotFoundOrNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := tt.fake
			if fake == nil {
				fake = &nodeRepositoryFake{}
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
		fake    *nodeRepositoryFake
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
			nodeID:  testutil.UUIDFromStringT(t, "99999999-9999-4999-8999-999999999999"),
			req:     &node.MoveNodeRequest{ParentID: node.NullableString{Value: testutil.StringPtr("99999999-9999-4999-8999-999999999999"), IsSet: true}},
			wantErr: node.ErrNodeCannotBeADescendantOfItself,
		},
		{
			name:   "node cannot be a descendant of itself when ancestor contains node",
			nodeID: testutil.UUIDFromStringT(t, "99999999-9999-4999-8999-999999999999"),
			fake: &nodeRepositoryFake{
				getNodeByIDResult: map[string]nodeDb.Node{
					"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa": {ID: testutil.UUIDFromStringT(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")},
				},
				ancestors: map[string][]pgtype.UUID{
					"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa": {testutil.UUIDFromStringT(t, "99999999-9999-4999-8999-999999999999")},
				},
			},
			req:     &node.MoveNodeRequest{ParentID: node.NullableString{Value: testutil.StringPtr("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"), IsSet: true}},
			wantErr: node.ErrNodeCannotBeADescendantOfItself,
		},
		{
			name:   "success",
			nodeID: testutil.UUIDFromStringT(t, "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"),
			userID: testutil.UUIDFromStringT(t, "cccccccc-cccc-4ccc-8ccc-cccccccccccc"),
			fake: &nodeRepositoryFake{
				getNodeByIDResult: map[string]nodeDb.Node{
					"dddddddd-dddd-4ddd-8ddd-dddddddddddd": {ID: testutil.UUIDFromStringT(t, "dddddddd-dddd-4ddd-8ddd-dddddddddddd"), UserID: testutil.UUIDFromStringT(t, "cccccccc-cccc-4ccc-8ccc-cccccccccccc")},
				},
				moveResult: nodeDb.MoveNodeRow{
					ParentID:  testutil.UUIDFromStringT(t, "dddddddd-dddd-4ddd-8ddd-dddddddddddd"),
					SortOrder: 42,
					UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				},
			},
			req: &node.MoveNodeRequest{
				ParentID:  node.NullableString{Value: testutil.StringPtr("dddddddd-dddd-4ddd-8ddd-dddddddddddd"), IsSet: true},
				SortOrder: testutil.Int64Ptr(42),
			},
			check: func(t *testing.T, resp node.MoveNodeResponse) {
				t.Helper()
				require.NotNil(t, resp.ParentID)
				require.Equal(t, "dddddddd-dddd-4ddd-8ddd-dddddddddddd", *resp.ParentID)
				require.EqualValues(t, 42, resp.SortOrder)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := tt.fake
			if fake == nil {
				fake = &nodeRepositoryFake{}
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
	parentId := "ffffffff-ffff-4fff-8fff-ffffffffffff"
	userID := testutil.UUIDFromStringT(t, "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee")
	parentUUID := testutil.UUIDFromStringT(t, parentId)
	fake := &nodeRepositoryFake{
		getNodeByIDResult: map[string]nodeDb.Node{parentId: {ID: parentUUID, UserID: userID}},
		createResult: nodeDb.Node{
			ID:        testutil.UUIDFromStringT(t, "11111111-1111-4111-8111-111111111112"),
			ParentID:  parentUUID,
			Type:      nodeDb.NodeTypeNote,
			Title:     "child",
			SortOrder: 123,
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		},
	}

	resp, err := node.NewService(fake).CreateNode(context.Background(), userID, &node.CreateNodeRequest{
		ParentID: &parentId,
		Type:     "note",
		Title:    "child",
	})
	require.NoError(t, err)
	require.Equal(t, parentId, resp.ParentID)
	require.Equal(t, "note", resp.Type)
	require.Equal(t, "child", resp.Title)
	require.Len(t, fake.createParams, 1)
}
