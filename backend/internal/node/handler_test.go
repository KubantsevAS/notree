package node_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	nodeDb "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/node"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestHandlerCreate(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	fake := &nodeStoreFake{}
	handler := node.NewHandler(node.NewService(fake))

	req := testutil.WithUserID(testutil.NewJSONRequest(t, http.MethodPost, "/nodes", node.CreateNodeRequest{
		Type:  "note",
		Title: "hello",
	}), userID)
	res := httptest.NewRecorder()

	handler.Create(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	require.Len(t, fake.createParams, 1)
	require.Equal(t, userID, fake.createParams[0].UserID)
	var payload node.CreateNodeResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, "note", payload.Type)
	require.Equal(t, "hello", payload.Title)
}

func TestHandlerDelete(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	fake := &nodeStoreFake{softDeleteResult: []pgtype.UUID{testutil.UUIDFromStringT(t, testutil.UUID2)}}
	handler := node.NewHandler(node.NewService(fake))

	req := testutil.WithRouteParam(
		testutil.WithUserID(httptest.NewRequest(http.MethodDelete, "/nodes/"+testutil.UUID1, nil), userID),
		"id",
		testutil.UUID1,
	)
	res := httptest.NewRecorder()

	handler.Delete(res, req)

	require.Equal(t, http.StatusNoContent, res.Code)
	require.Len(t, fake.softDeleteParams, 1)
}

func TestHandlerUpdate(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	updatedAt := time.Now()
	fake := &nodeStoreFake{
		updateResult: nodeDb.UpdateNodeRow{
			Type:      nodeDb.NodeTypeTask,
			Title:     "done",
			UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
		},
	}
	handler := node.NewHandler(node.NewService(fake))

	req := testutil.WithRouteParam(
		testutil.WithUserID(testutil.NewJSONRequest(
			t,
			http.MethodPatch,
			"/nodes/:id",
			node.UpdateNodeRequest{
				Type:  testutil.StringPtr("task"),
				Title: testutil.StringPtr("done"),
			}),
			userID,
		),
		"id",
		nodeID.String(),
	)
	res := httptest.NewRecorder()

	handler.Update(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var payload node.UpdateNodeResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, "task", payload.Type)
	require.Equal(t, "done", payload.Title)
	require.Len(t, fake.updateParams, 1)
}

type handlerFunc func(*node.Handler, http.ResponseWriter, *http.Request)

func TestHandler_Errors(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	dbErr := errors.New("db down")
	createBody := node.CreateNodeRequest{Type: "note", Title: "child"}
	updateBody := node.UpdateNodeRequest{Title: testutil.StringPtr("title")}

	tests := []struct {
		name       string
		call       handlerFunc
		method     string
		nodeID     string
		body       any
		rawBody    string
		anonymous  bool
		store      *nodeStoreFake
		wantStatus int
		wantMsg    string
	}{
		{
			name: "create unauthorized", call: (*node.Handler).Create, method: http.MethodPost,
			body: createBody, anonymous: true,
			wantStatus: http.StatusUnauthorized, wantMsg: "User ID not found in context",
		},
		{
			name: "create malformed body", call: (*node.Handler).Create, method: http.MethodPost,
			rawBody:    "{bad json}",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "create invalid parent id", call: (*node.Handler).Create, method: http.MethodPost,
			body:       node.CreateNodeRequest{ParentID: testutil.StringPtr(testutil.BadUUID), Type: "note", Title: "child"},
			wantStatus: http.StatusBadRequest, wantMsg: "invalid parent id",
		},
		{
			name: "create parent not found", call: (*node.Handler).Create, method: http.MethodPost,
			body:       node.CreateNodeRequest{ParentID: testutil.StringPtr(testutil.UUID2), Type: "note", Title: "child"},
			wantStatus: http.StatusBadRequest, wantMsg: "parent not found",
		},
		{
			name: "create db error", call: (*node.Handler).Create, method: http.MethodPost,
			body: createBody, store: &nodeStoreFake{createErr: dbErr},
			wantStatus: http.StatusInternalServerError, wantMsg: "internal server error",
		},
		{
			name: "delete unauthorized", call: (*node.Handler).Delete, method: http.MethodDelete,
			nodeID: testutil.UUID2, anonymous: true,
			wantStatus: http.StatusUnauthorized, wantMsg: "User ID not found in context",
		},
		{
			name: "delete invalid node id", call: (*node.Handler).Delete, method: http.MethodDelete,
			nodeID:     testutil.BadUUID,
			wantStatus: http.StatusBadRequest, wantMsg: "invalid node id format",
		},
		{
			name: "delete not found", call: (*node.Handler).Delete, method: http.MethodDelete,
			nodeID:     testutil.UUID2,
			wantStatus: http.StatusNotFound, wantMsg: "node not found or access denied",
		},
		{
			name: "delete db error", call: (*node.Handler).Delete, method: http.MethodDelete,
			nodeID: testutil.UUID2, store: &nodeStoreFake{softDeleteErr: dbErr},
			wantStatus: http.StatusInternalServerError, wantMsg: "internal server error",
		},
		{
			name: "update unauthorized", call: (*node.Handler).Update, method: http.MethodPatch,
			nodeID: testutil.UUID2, body: updateBody, anonymous: true,
			wantStatus: http.StatusUnauthorized, wantMsg: "User ID not found in context",
		},
		{
			name: "update invalid node id", call: (*node.Handler).Update, method: http.MethodPatch,
			nodeID: testutil.BadUUID, body: updateBody,
			wantStatus: http.StatusBadRequest, wantMsg: "invalid node id format",
		},
		{
			name: "update empty payload", call: (*node.Handler).Update, method: http.MethodPatch,
			nodeID: testutil.UUID2, body: map[string]any{},
			wantStatus: http.StatusBadRequest, wantMsg: "no fields provided for update",
		},
		{
			name: "update not found", call: (*node.Handler).Update, method: http.MethodPatch,
			nodeID: testutil.UUID2, body: updateBody, store: &nodeStoreFake{updateErr: pgx.ErrNoRows},
			wantStatus: http.StatusNotFound, wantMsg: "node not found or access denied",
		},
		{
			name: "update db error", call: (*node.Handler).Update, method: http.MethodPatch,
			nodeID: testutil.UUID2, body: updateBody, store: &nodeStoreFake{updateErr: dbErr},
			wantStatus: http.StatusInternalServerError, wantMsg: "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := tc.store
			if store == nil {
				store = &nodeStoreFake{}
			}
			handler := node.NewHandler(node.NewService(store))

			path := "/nodes/" + tc.nodeID
			var req *http.Request
			if tc.rawBody != "" {
				req = httptest.NewRequest(tc.method, path, strings.NewReader(tc.rawBody))
			} else {
				req = testutil.NewJSONRequest(t, tc.method, path, tc.body)
			}
			if tc.nodeID != "" {
				req = testutil.WithRouteParam(req, "id", tc.nodeID)
			}
			if !tc.anonymous {
				req = testutil.WithUserID(req, userID)
			}
			res := httptest.NewRecorder()

			tc.call(handler, res, req)

			require.Equal(t, tc.wantStatus, res.Code)
			if tc.wantMsg != "" {
				testutil.AssertErrorJSON(t, res, tc.wantMsg)
			}
		})
	}
}
