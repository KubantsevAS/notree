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

func TestHandlerCreate_Unauthorized(t *testing.T) {
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := testutil.NewJSONRequest(t, http.MethodPost, "/nodes", map[string]string{"type": "note", "title": "hello"})
	res := httptest.NewRecorder()

	handler.Create(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
	testutil.AssertErrorJSON(t, res, "User ID not found in context")
}

func TestHandlerCreate_InvalidParentID(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := testutil.WithUserID(testutil.NewJSONRequest(t, http.MethodPost, "/nodes", node.CreateNodeRequest{
		ParentID: testutil.StringPtr(testutil.BadUUID),
		Type:     "note",
		Title:    "child",
	}), userID)
	res := httptest.NewRecorder()

	handler.Create(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "invalid parent id")
}

func TestHandlerCreate_ParentNotFound(t *testing.T) {
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))
	req := testutil.WithUserID(testutil.NewJSONRequest(t, http.MethodPost, "/nodes", node.CreateNodeRequest{
		ParentID: testutil.StringPtr(testutil.UUID2),
		Type:     "note",
		Title:    "child",
	}), testutil.UUIDFromStringT(t, testutil.UUID1))
	res := httptest.NewRecorder()

	handler.Create(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "parent not found")
}

func TestHandlerCreate_InvalidBody(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := testutil.WithUserID(httptest.NewRequest(http.MethodPost, "/nodes", strings.NewReader("{bad json}")), userID)
	res := httptest.NewRecorder()

	handler.Create(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
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

func TestHandlerDelete_InvalidNodeID(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := testutil.WithRouteParam(
		testutil.WithUserID(httptest.NewRequest(http.MethodDelete,
			"/nodes/bad-id",
			nil,
		),
			userID,
		),
		"id",
		testutil.BadUUID,
	)
	res := httptest.NewRecorder()

	handler.Delete(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "invalid node id format")
}

func TestHandlerDelete_NotFound(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := testutil.WithRouteParam(
		testutil.WithUserID(httptest.NewRequest(http.MethodDelete, "/nodes/"+testutil.UUID1, nil), userID),
		"id",
		testutil.UUID1,
	)
	res := httptest.NewRecorder()

	handler.Delete(res, req)

	require.Equal(t, http.StatusNotFound, res.Code)
	testutil.AssertErrorJSON(t, res, "node not found or access denied")
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

func TestHandlerUpdate_EmptyPayload(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := testutil.WithRouteParam(
		testutil.WithUserID(testutil.NewJSONRequest(t, http.MethodPatch, "/nodes/:id", map[string]any{}), userID),
		"id",
		testutil.UUID1,
	)
	res := httptest.NewRecorder()

	handler.Update(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "no fields provided for update")
}

func TestHandlerUpdate_InvalidNodeID(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := testutil.WithRouteParam(
		testutil.WithUserID(testutil.NewJSONRequest(
			t,
			http.MethodPatch,
			"/nodes/:id",
			node.UpdateNodeRequest{Title: testutil.StringPtr("t")},
		),
			userID,
		),
		"id",
		testutil.BadUUID,
	)
	res := httptest.NewRecorder()

	handler.Update(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "invalid node id format")
}

func TestHandlerUpdate_NotFound(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{updateErr: pgx.ErrNoRows}))

	req := testutil.WithRouteParam(testutil.WithUserID(testutil.NewJSONRequest(
		t,
		http.MethodPatch,
		"/nodes/:id",
		node.UpdateNodeRequest{Title: testutil.StringPtr("title")},
	),
		userID,
	),
		"id",
		testutil.UUID1,
	)
	res := httptest.NewRecorder()

	handler.Update(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
	testutil.AssertErrorJSON(t, res, "node not found or access denied")
}

func TestHandler_Unauthorized(t *testing.T) {
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	tests := []struct {
		name    string
		method  string
		body    any
		execute func(h *node.Handler, w http.ResponseWriter, r *http.Request)
	}{
		{
			"Delete",
			http.MethodDelete,
			nil,
			func(h *node.Handler, w http.ResponseWriter, r *http.Request) { h.Delete(w, r) },
		},
		{
			"Update",
			http.MethodPatch,
			node.UpdateNodeRequest{Title: testutil.StringPtr("t")},
			func(h *node.Handler, w http.ResponseWriter, r *http.Request) { h.Update(w, r) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := testutil.WithRouteParam(testutil.NewJSONRequest(t, tc.method, "/nodes", tc.body), "id", testutil.UUID1)
			res := httptest.NewRecorder()

			tc.execute(handler, res, req)

			require.Equal(t, http.StatusUnauthorized, res.Code)
			testutil.AssertErrorJSON(t, res, "User ID not found in context")
		})
	}
}

func TestHandler_InternalErrors(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	tests := []struct {
		name     string
		method   string
		routeKey string
		path     string
		body     any
		setup    func(t *testing.T) *nodeStoreFake
		execute  func(h *node.Handler, w http.ResponseWriter, r *http.Request)
	}{
		{
			name:     "delete db error",
			method:   http.MethodDelete,
			routeKey: "id",
			path:     "/nodes/" + testutil.UUID1,
			body:     nil,
			setup: func(t *testing.T) *nodeStoreFake {
				return &nodeStoreFake{softDeleteErr: errors.New("db down")}
			},
			execute: func(h *node.Handler, w http.ResponseWriter, r *http.Request) { h.Delete(w, r) },
		},
		{
			name:     "update db error",
			method:   http.MethodPatch,
			routeKey: "id",
			path:     "/nodes/" + testutil.UUID1,
			body:     node.UpdateNodeRequest{Title: testutil.StringPtr("new title")},
			setup: func(t *testing.T) *nodeStoreFake {
				return &nodeStoreFake{updateErr: errors.New("db down")}
			},
			execute: func(h *node.Handler, w http.ResponseWriter, r *http.Request) { h.Update(w, r) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := tc.setup(t)
			handler := node.NewHandler(node.NewService(fake))

			req := testutil.NewJSONRequest(t, tc.method, tc.path, tc.body)
			req = testutil.WithUserID(req, userID)

			req = testutil.WithRouteParam(req, tc.routeKey, testutil.UUID1)

			res := httptest.NewRecorder()

			tc.execute(handler, res, req)

			require.Equal(t, http.StatusInternalServerError, res.Code)
			testutil.AssertErrorJSON(t, res, "internal server error")
		})
	}
}
