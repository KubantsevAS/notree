package node_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	nodeDb "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/http/middleware"
	"github.com/KubantsevAS/notree/backend/internal/node"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func withNodeUserContext(t *testing.T, req *http.Request, userID pgtype.UUID) *http.Request {
	t.Helper()
	return req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID.String()))
}

func withRouteParam(req *http.Request, key, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func TestNodeHandlerCreateSuccess(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	fake := &nodeStoreFake{}
	handler := node.NewHandler(node.NewService(fake))

	req := withNodeUserContext(t, testutil.NewJSONRequest(t, http.MethodPost, "/nodes", node.CreateNodeRequest{
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

func TestNodeHandlerCreateUnauthorized(t *testing.T) {
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := testutil.NewJSONRequest(t, http.MethodPost, "/nodes", map[string]string{"type": "note", "title": "hello"})
	res := httptest.NewRecorder()

	handler.Create(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
	testutil.AssertErrorJSON(t, res, "User ID not found in context")
}

func TestNodeHandlerCreateInvalidParentID(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := withNodeUserContext(t, testutil.NewJSONRequest(t, http.MethodPost, "/nodes", node.CreateNodeRequest{
		ParentID: testutil.StringPtr(testUUIDBad),
		Type:     "note",
		Title:    "child",
	}), userID)
	res := httptest.NewRecorder()

	handler.Create(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "invalid parent id")
}

func TestNodeHandlerCreateParentNotFound(t *testing.T) {
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))
	req := withNodeUserContext(t, testutil.NewJSONRequest(t, http.MethodPost, "/nodes", node.CreateNodeRequest{
		ParentID: testutil.StringPtr(testUUID2),
		Type:     "note",
		Title:    "child",
	}), testutil.UUIDFromStringT(t, testUUID1))
	res := httptest.NewRecorder()

	handler.Create(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "parent not found")
}

func TestNodeHandlerCreateInvalidBody(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := withNodeUserContext(
		t,
		httptest.NewRequest(http.MethodPost, "/nodes", strings.NewReader("{bad json}")),
		userID,
	)
	res := httptest.NewRecorder()

	handler.Create(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestNodeHandlerDeleteSuccess(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	fake := &nodeStoreFake{softDeleteResult: []pgtype.UUID{testutil.UUIDFromStringT(t, testUUID2)}}
	handler := node.NewHandler(node.NewService(fake))

	req := withRouteParam(
		withNodeUserContext(
			t,
			httptest.NewRequest(http.MethodDelete, "/nodes/"+testUUID1, nil),
			userID,
		),
		"id",
		testUUID1,
	)
	res := httptest.NewRecorder()

	handler.Delete(res, req)

	require.Equal(t, http.StatusNoContent, res.Code)
	require.Len(t, fake.softDeleteParams, 1)
}

func TestNodeHandlerDeleteInvalidNodeID(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := withRouteParam(
		withNodeUserContext(
			t,
			httptest.NewRequest(http.MethodDelete,
				"/nodes/bad-id",
				nil,
			),
			userID,
		),
		"id",
		"bad-id",
	)
	res := httptest.NewRecorder()

	handler.Delete(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "invalid node id format")
}

func TestNodeHandlerDeleteNotFound(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := withRouteParam(
		withNodeUserContext(
			t,
			httptest.NewRequest(http.MethodDelete, "/nodes/"+testUUID1, nil),
			userID,
		),
		"id",
		testUUID1,
	)
	res := httptest.NewRecorder()

	handler.Delete(res, req)

	require.Equal(t, http.StatusNotFound, res.Code)
	testutil.AssertErrorJSON(t, res, "node not found or access denied")
}

func TestNodeHandlerUpdateSuccess(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	nodeID := testutil.UUIDFromStringT(t, testUUID2)
	updatedAt := time.Now()
	fake := &nodeStoreFake{
		updateResult: nodeDb.UpdateNodeRow{
			Type:      nodeDb.NodeTypeTask,
			Title:     "done",
			UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
		},
	}
	handler := node.NewHandler(node.NewService(fake))

	req := withRouteParam(
		withNodeUserContext(
			t,
			testutil.NewJSONRequest(
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

func TestNodeHandlerUpdateEmptyPayload(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := withRouteParam(
		withNodeUserContext(t,
			testutil.NewJSONRequest(t, http.MethodPatch, "/nodes/:id", map[string]any{}),
			userID,
		),
		"id",
		testUUID1,
	)
	res := httptest.NewRecorder()

	handler.Update(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "no fields provided for update")
}

func TestNodeHandlerUpdateInvalidNodeID(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := withRouteParam(
		withNodeUserContext(t,
			testutil.NewJSONRequest(
				t,
				http.MethodPatch,
				"/nodes/:id",
				node.UpdateNodeRequest{Title: testutil.StringPtr("t")},
			),
			userID,
		),
		"id",
		"bad-id",
	)
	res := httptest.NewRecorder()

	handler.Update(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "invalid node id format")
}

func TestNodeHandlerUpdateAccessDenied(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{updateErr: pgx.ErrNoRows}))

	req := withRouteParam(withNodeUserContext(
		t,
		testutil.NewJSONRequest(
			t,
			http.MethodPatch,
			"/nodes/:id",
			node.UpdateNodeRequest{Title: testutil.StringPtr("title")},
		),
		userID,
	),
		"id",
		testUUID1,
	)
	res := httptest.NewRecorder()

	handler.Update(res, req)
	require.Equal(t, http.StatusNotFound, res.Code)
	testutil.AssertErrorJSON(t, res, "node not found or access denied")
}

func TestNodeHandlerMoveSuccess(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	nodeID := testutil.UUIDFromStringT(t, testUUID2)
	parentID := testutil.UUIDFromStringT(t, testUUID3)
	updatedAt := time.Now()
	fake := &nodeStoreFake{
		getNodeByIDResult: map[string]nodeDb.Node{parentID.String(): {ID: parentID, UserID: userID}},
		moveResult: nodeDb.MoveNodeRow{
			ParentID:  parentID,
			SortOrder: 42,
			UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
		},
	}
	handler := node.NewHandler(node.NewService(fake))

	req := withRouteParam(
		withNodeUserContext(t, testutil.NewJSONRequest(
			t,
			http.MethodPost,
			"/nodes/:id/move",
			map[string]any{
				"parent_id":  parentID.String(),
				"sort_order": 42,
			},
		), userID), "id", nodeID.String(),
	)
	res := httptest.NewRecorder()

	handler.Move(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var payload node.MoveNodeResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.NotNil(t, payload.ParentID)
	require.Equal(t, parentID.String(), *payload.ParentID)
	require.EqualValues(t, 42, payload.SortOrder)
}

func TestNodeHandlerMoveCircularReference(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	nodeID := testutil.UUIDFromStringT(t, testUUID2)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := withRouteParam(
		withNodeUserContext(t, testutil.NewJSONRequest(
			t,
			http.MethodPost,
			"/nodes/:id/move",
			map[string]any{
				"parent_id":  nodeID.String(),
				"sort_order": 42,
			},
		), userID), "id", nodeID.String(),
	)
	res := httptest.NewRecorder()

	handler.Move(res, req)

	require.Equal(t, http.StatusConflict, res.Code)
	testutil.AssertErrorJSON(t, res, "node cannot be a descendant of itself (circular reference)")
}

func TestNodeHandlerMoveInvalidParentUUID(t *testing.T) {
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))
	req := withRouteParam(
		withNodeUserContext(
			t,
			testutil.NewJSONRequest(t, http.MethodPost, "/nodes/:id/move", map[string]string{"parent_id": "bad-uuid"}),
			testutil.UUIDFromStringT(t, testUUID1),
		),
		"id",
		testUUID1,
	)
	res := httptest.NewRecorder()

	handler.Move(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "invalid parent id")
}

func TestNodeHandlerMoveParentNotFound(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	handler := node.NewHandler(node.NewService(&nodeStoreFake{}))

	req := withRouteParam(
		withNodeUserContext(
			t,
			testutil.NewJSONRequest(t, http.MethodPost, "/nodes/:id/move", map[string]any{"parent_id": testUUID2}),
			userID,
		),
		"id",
		testUUID1,
	)
	res := httptest.NewRecorder()

	handler.Move(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "parent not found")
}

func TestNodeHandlerRequiresAuthentication(t *testing.T) {
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
		{
			"Move",
			http.MethodPost,
			map[string]any{"parent_id": testUUID1},
			func(h *node.Handler, w http.ResponseWriter, r *http.Request) { h.Move(w, r) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := withRouteParam(testutil.NewJSONRequest(t, tc.method, "/nodes", tc.body), "id", testUUID1)
			res := httptest.NewRecorder()

			tc.execute(handler, res, req)

			require.Equal(t, http.StatusUnauthorized, res.Code)
			testutil.AssertErrorJSON(t, res, "User ID not found in context")
		})
	}
}

func TestNodeHandlerInternalErrors(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
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
			name:     "move db error",
			method:   http.MethodPost,
			routeKey: "id",
			path:     "/nodes/" + testUUID1 + "/move",
			body: map[string]any{
				"parent_id":  testUUID2,
				"sort_order": 10,
			},
			setup: func(t *testing.T) *nodeStoreFake {
				return &nodeStoreFake{
					getNodeByIDResult: map[string]nodeDb.Node{
						testUUID2: {ID: testutil.UUIDFromStringT(t, testUUID2), UserID: userID},
					},
					moveErr: errors.New("db down"),
				}
			},
			execute: func(h *node.Handler, w http.ResponseWriter, r *http.Request) { h.Move(w, r) },
		},
		{
			name:     "delete db error",
			method:   http.MethodDelete,
			routeKey: "id",
			path:     "/nodes/" + testUUID1,
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
			path:     "/nodes/" + testUUID1,
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
			req = withNodeUserContext(t, req, userID)

			req = withRouteParam(req, tc.routeKey, testUUID1)

			res := httptest.NewRecorder()

			tc.execute(handler, res, req)

			require.Equal(t, http.StatusInternalServerError, res.Code)
			testutil.AssertErrorJSON(t, res, "internal server error")
		})
	}
}
