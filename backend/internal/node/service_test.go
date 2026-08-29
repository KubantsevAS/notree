package node

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

type nodeRepositoryFake struct {
	createParams      []node.CreateNodeParams
	createResult      node.Node
	createErr         error
	getNodeByIDResult map[string]node.Node
	getNodeByIDErr    error
	getNodeByIDCalls  []node.GetNodeByIDParams
	children          []node.Node
	childrenErr       error
	ancestors         map[string][]pgtype.UUID
	ancestorsErr      error
	softDeleteParams  []node.SoftDeleteNodeCascadeParams
	softDeleteResult  []pgtype.UUID
	softDeleteErr     error
	updateParams      []node.UpdateNodeParams
	updateResult      node.UpdateNodeRow
	updateErr         error
	moveParams        []node.MoveNodeParams
	moveResult        node.MoveNodeRow
	moveErr           error
}

func (f *nodeRepositoryFake) CreateNode(_ context.Context, params node.CreateNodeParams) (node.Node, error) {
	f.createParams = append(f.createParams, params)
	if f.createErr != nil {
		return node.Node{}, f.createErr
	}
	if f.createResult.ID.Valid {
		return f.createResult, nil
	}

	return node.Node{
		ID:        pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		ParentID:  params.ParentID,
		Type:      params.Type,
		Title:     params.Title,
		SortOrder: params.SortOrder,
	}, nil
}

func (f *nodeRepositoryFake) GetChildren(context.Context, node.GetChildrenParams) ([]node.Node, error) {
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

func (f *nodeRepositoryFake) GetNodeByID(_ context.Context, params node.GetNodeByIDParams) (node.Node, error) {
	f.getNodeByIDCalls = append(f.getNodeByIDCalls, params)
	if f.getNodeByIDErr != nil {
		return node.Node{}, f.getNodeByIDErr
	}
	if f.getNodeByIDResult != nil {
		if result, ok := f.getNodeByIDResult[params.ID.String()]; ok {
			return result, nil
		}
	}
	return node.Node{}, pgx.ErrNoRows
}

func (f *nodeRepositoryFake) MoveNode(_ context.Context, params node.MoveNodeParams) (node.MoveNodeRow, error) {
	f.moveParams = append(f.moveParams, params)
	if f.moveErr != nil {
		return node.MoveNodeRow{}, f.moveErr
	}
	return f.moveResult, nil
}

func (f *nodeRepositoryFake) SoftDeleteNodeCascade(_ context.Context, params node.SoftDeleteNodeCascadeParams) ([]pgtype.UUID, error) {
	f.softDeleteParams = append(f.softDeleteParams, params)
	if f.softDeleteErr != nil {
		return nil, f.softDeleteErr
	}
	return f.softDeleteResult, nil
}

func (f *nodeRepositoryFake) UpdateNode(_ context.Context, params node.UpdateNodeParams) (node.UpdateNodeRow, error) {
	f.updateParams = append(f.updateParams, params)
	if f.updateErr != nil {
		return node.UpdateNodeRow{}, f.updateErr
	}
	return f.updateResult, nil
}

func newUUIDFromString(t *testing.T, value string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		t.Fatalf("pgtype.UUID scan failed: %v", err)
	}
	return id
}

func strPtr(value string) *string {
	return &value
}

func int64Ptr(value int64) *int64 {
	return &value
}

func TestNodeServiceCreateNodeSortOrderIncreases(t *testing.T) {
	fake := &nodeRepositoryFake{}
	service := NewService(fake)
	request := &CreateNodeRequest{Type: "note", Title: "test"}

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
		req     *CreateNodeRequest
		fake    *nodeRepositoryFake
		wantErr error
	}{
		{
			name:    "invalid parent uuid",
			req:     &CreateNodeRequest{ParentID: strPtr("bad-uuid"), Type: "note", Title: "child"},
			fake:    &nodeRepositoryFake{},
			wantErr: ErrInvalidParentID,
		},
		{
			name:    "parent not found",
			req:     &CreateNodeRequest{ParentID: strPtr("11111111-1111-4111-8111-111111111111"), Type: "note", Title: "child"},
			fake:    &nodeRepositoryFake{},
			wantErr: ErrParentNotFound,
		},
		{
			name:    "db error on parent lookup",
			req:     &CreateNodeRequest{ParentID: strPtr("11111111-1111-4111-8111-111111111111"), Type: "note", Title: "child"},
			fake:    &nodeRepositoryFake{getNodeByIDErr: errTimeout},
			wantErr: errTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewService(tt.fake).CreateNode(context.Background(), pgtype.UUID{}, tt.req)
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
		getNodeByIDResult: map[string]node.Node{parentID.String(): {ID: parentID, UserID: userID}},
		children: []node.Node{
			{ID: childOneID, ParentID: parentID, UserID: userID, Title: "first", SortOrder: 10},
			{ID: childTwoID, ParentID: parentID, UserID: userID, Title: "second", SortOrder: 20},
		},
	}

	children, err := NewService(fake).GetChildren(context.Background(), parentID, userID)
	require.NoError(t, err)
	require.Len(t, children, 2)
	require.Equal(t, childOneID.String(), children[0].ID)
	require.Equal(t, childTwoID.String(), children[1].ID)
}

func TestNodeServiceGetChildrenReturnsErrParentNotFound(t *testing.T) {
	parentID := newUUIDFromString(t, "11111111-1111-4111-8111-111111111111")
	userID := newUUIDFromString(t, "22222222-2222-4222-8222-222222222222")

	_, err := NewService(&nodeRepositoryFake{}).GetChildren(context.Background(), parentID, userID)
	require.ErrorIs(t, err, ErrParentNotFound)
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
			nodeID: newUUIDFromString(t, "33333333-3333-4333-8333-333333333333"),
			userID: newUUIDFromString(t, "44444444-4444-4444-8444-444444444444"),
			fake: &nodeRepositoryFake{
				softDeleteResult: []pgtype.UUID{newUUIDFromString(t, "33333333-3333-4333-8333-333333333333")},
			},
			check: func(t *testing.T, fake *nodeRepositoryFake) {
				t.Helper()
				require.Len(t, fake.softDeleteParams, 1)
			},
		},
		{
			name:    "not found or no access",
			nodeID:  newUUIDFromString(t, "55555555-5555-4555-8555-555555555555"),
			userID:  newUUIDFromString(t, "66666666-6666-4666-8666-666666666666"),
			fake:    &nodeRepositoryFake{},
			wantErr: ErrNodeNotFoundOrNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := tt.fake
			if fake == nil {
				fake = &nodeRepositoryFake{}
			}

			err := NewService(fake).DeleteNode(context.Background(), tt.nodeID, tt.userID)
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
		req     *UpdateNodeRequest
		fake    *nodeRepositoryFake
		wantErr error
		check   func(*testing.T, UpdateNodeResponse)
	}{
		{
			name:    "empty request",
			req:     &UpdateNodeRequest{},
			wantErr: domain.ErrEmptyUpdate,
		},
		{
			name:   "success",
			nodeID: newUUIDFromString(t, "77777777-7777-4777-8777-777777777777"),
			userID: newUUIDFromString(t, "88888888-8888-4888-8888-888888888888"),
			fake: &nodeRepositoryFake{
				updateResult: node.UpdateNodeRow{
					Type:      node.NodeTypeTask,
					Title:     "done",
					UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				},
			},
			req: &UpdateNodeRequest{Type: strPtr("task"), Title: strPtr("done")},
			check: func(t *testing.T, resp UpdateNodeResponse) {
				t.Helper()
				require.Equal(t, "task", resp.Type)
				require.Equal(t, "done", resp.Title)
			},
		},
		{
			name:    "node not found or no access",
			fake:    &nodeRepositoryFake{updateErr: pgx.ErrNoRows},
			req:     &UpdateNodeRequest{Title: strPtr("new title")},
			wantErr: ErrNodeNotFoundOrNoAccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := tt.fake
			if fake == nil {
				fake = &nodeRepositoryFake{}
			}

			resp, err := NewService(fake).UpdateNode(context.Background(), tt.nodeID, tt.userID, tt.req)
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
		req     *MoveNodeRequest
		fake    *nodeRepositoryFake
		wantErr error
		check   func(*testing.T, MoveNodeResponse)
	}{
		{
			name:    "empty request",
			req:     &MoveNodeRequest{},
			wantErr: domain.ErrEmptyUpdate,
		},
		{
			name:    "invalid parent uuid",
			req:     &MoveNodeRequest{ParentID: NullableString{Value: strPtr("bad-uuid"), IsSet: true}},
			wantErr: ErrInvalidParentID,
		},
		{
			name:    "node cannot be a descendant of itself when parent is same id",
			nodeID:  newUUIDFromString(t, "99999999-9999-4999-8999-999999999999"),
			req:     &MoveNodeRequest{ParentID: NullableString{Value: strPtr("99999999-9999-4999-8999-999999999999"), IsSet: true}},
			wantErr: ErrNodeCannotBeADescendantOfItself,
		},
		{
			name:   "node cannot be a descendant of itself when ancestor contains node",
			nodeID: newUUIDFromString(t, "99999999-9999-4999-8999-999999999999"),
			fake: &nodeRepositoryFake{
				getNodeByIDResult: map[string]node.Node{
					"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa": {ID: newUUIDFromString(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")},
				},
				ancestors: map[string][]pgtype.UUID{
					"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa": {newUUIDFromString(t, "99999999-9999-4999-8999-999999999999")},
				},
			},
			req:     &MoveNodeRequest{ParentID: NullableString{Value: strPtr("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"), IsSet: true}},
			wantErr: ErrNodeCannotBeADescendantOfItself,
		},
		{
			name:   "success",
			nodeID: newUUIDFromString(t, "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"),
			userID: newUUIDFromString(t, "cccccccc-cccc-4ccc-8ccc-cccccccccccc"),
			fake: &nodeRepositoryFake{
				getNodeByIDResult: map[string]node.Node{
					"dddddddd-dddd-4ddd-8ddd-dddddddddddd": {ID: newUUIDFromString(t, "dddddddd-dddd-4ddd-8ddd-dddddddddddd"), UserID: newUUIDFromString(t, "cccccccc-cccc-4ccc-8ccc-cccccccccccc")},
				},
				moveResult: node.MoveNodeRow{
					ParentID:  newUUIDFromString(t, "dddddddd-dddd-4ddd-8ddd-dddddddddddd"),
					SortOrder: 42,
					UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				},
			},
			req: &MoveNodeRequest{
				ParentID:  NullableString{Value: strPtr("dddddddd-dddd-4ddd-8ddd-dddddddddddd"), IsSet: true},
				SortOrder: int64Ptr(42),
			},
			check: func(t *testing.T, resp MoveNodeResponse) {
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

			resp, err := NewService(fake).MoveNode(context.Background(), tt.nodeID, tt.userID, tt.req)
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
	userID := newUUIDFromString(t, "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee")
	parentID := newUUIDFromString(t, "ffffffff-ffff-4fff-8fff-ffffffffffff")
	fake := &nodeRepositoryFake{
		getNodeByIDResult: map[string]node.Node{parentID.String(): {ID: parentID, UserID: userID}},
		createResult: node.Node{
			ID:        newUUIDFromString(t, "11111111-1111-4111-8111-111111111112"),
			ParentID:  parentID,
			Type:      node.NodeTypeNote,
			Title:     "child",
			SortOrder: 123,
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		},
	}

	resp, err := NewService(fake).CreateNode(context.Background(), userID, &CreateNodeRequest{
		ParentID: strPtr(parentID.String()),
		Type:     "note",
		Title:    "child",
	})
	require.NoError(t, err)
	require.Equal(t, parentID.String(), resp.ParentID)
	require.Equal(t, "note", resp.Type)
	require.Equal(t, "child", resp.Title)
	require.Len(t, fake.createParams, 1)
}
