package node_test

import (
	"context"
	"errors"
	"testing"
	"time"

	nodeDb "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/domain"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy/rank"
	"github.com/KubantsevAS/notree/backend/internal/node"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestCreateNode_AppendsWithRankGap(t *testing.T) {
	fake := &nodeStoreFake{}
	request := &node.CreateNodeRequest{Type: "note", Title: "test"}

	_, err := node.NewService(fake).CreateNode(context.Background(), pgtype.UUID{}, request)

	require.NoError(t, err)
	require.Len(t, fake.createParams, 1)
	require.Equal(t, rank.Gap, fake.createParams[0].Gap)
}

func TestCreateNode_ValidatesParentID(t *testing.T) {
	var errTimeout = errors.New("timeout")
	tests := []struct {
		name    string
		req     *node.CreateNodeRequest
		fake    *nodeStoreFake
		wantErr error
	}{
		{
			name:    "invalid parent uuid",
			req:     &node.CreateNodeRequest{ParentID: testutil.StringPtr(testutil.BadUUID), Type: "note", Title: "child"},
			fake:    &nodeStoreFake{},
			wantErr: node.ErrInvalidParentID,
		},
		{
			name:    "parent not found",
			req:     &node.CreateNodeRequest{ParentID: testutil.StringPtr(testutil.UUID1), Type: "note", Title: "child"},
			fake:    &nodeStoreFake{},
			wantErr: node.ErrParentNotFound,
		},
		{
			name:    "db error on parent lookup",
			req:     &node.CreateNodeRequest{ParentID: testutil.StringPtr(testutil.UUID1), Type: "note", Title: "child"},
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

func TestDeleteNode(t *testing.T) {
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
			nodeID: testutil.UUIDFromStringT(t, testutil.UUID3),
			userID: testutil.UUIDFromStringT(t, testutil.UUID4),
			fake: &nodeStoreFake{
				softDeleteResult: []pgtype.UUID{testutil.UUIDFromStringT(t, testutil.UUID3)},
			},
			check: func(t *testing.T, fake *nodeStoreFake) {
				t.Helper()
				require.Len(t, fake.softDeleteParams, 1)
			},
		},
		{
			name:    "not found or no access",
			nodeID:  testutil.UUIDFromStringT(t, testutil.UUID2),
			userID:  testutil.UUIDFromStringT(t, testutil.UUID1),
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

func TestUpdateNode(t *testing.T) {
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
			nodeID: testutil.UUIDFromStringT(t, testutil.UUID3),
			userID: testutil.UUIDFromStringT(t, testutil.UUID4),
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

func TestCreateNode_AddsParentAndSortOrder(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID2)
	fake := &nodeStoreFake{
		getNodeByIDResult: map[string]nodeDb.Node{testutil.UUID2: {ID: parentID, UserID: userID}},
		createResult: nodeDb.Node{
			ID:        testutil.UUIDFromStringT(t, testutil.UUID3),
			ParentID:  parentID,
			Type:      nodeDb.NodeTypeNote,
			Title:     "child",
			SortOrder: 123,
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		},
	}

	resp, err := node.NewService(fake).CreateNode(context.Background(), userID, &node.CreateNodeRequest{
		ParentID: testutil.StringPtr(testutil.UUID2),
		Type:     "note",
		Title:    "child",
	})
	require.NoError(t, err)
	require.Equal(t, testutil.UUID2, resp.ParentID)
	require.Equal(t, "note", resp.Type)
	require.Equal(t, "child", resp.Title)
	require.Len(t, fake.createParams, 1)
}
