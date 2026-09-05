package hierarchy_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	hierarchyDB "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	nodeDB "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy"
	"github.com/KubantsevAS/notree/backend/internal/http/middleware"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/go-chi/chi/v5"
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

func TestHierarchyHandlerGetChildrenSuccess(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	parentID := testutil.UUIDFromStringT(t, testUUID2)
	updatedAt := time.Now()
	fakeNodeStore := &nodeStoreFake{
		getNodeByIDResult: map[string]nodeDB.Node{parentID.String(): {ID: parentID, UserID: userID}},
	}
	fake := &hierarchyStoreFake{
		children: []hierarchyDB.Node{{
			ID:        testutil.UUIDFromStringT(t, testUUID3),
			UserID:    userID,
			ParentID:  parentID,
			Type:      hierarchyDB.NodeTypeNote,
			Title:     "child",
			SortOrder: 1,
			CreatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
		}},
	}
	handler := hierarchy.NewHandler(hierarchy.NewService(fake, fakeNodeStore))

	req := withRouteParam(withNodeUserContext(
		t,
		httptest.NewRequest(http.MethodGet, "/nodes/:id/children", nil),
		userID,
	), "id", parentID.String())
	res := httptest.NewRecorder()

	handler.GetChildren(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var payload hierarchy.GetChildrenResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Len(t, payload, 1)
	require.Equal(t, "child", payload[0].Title)
	require.WithinDuration(t, updatedAt, *payload[0].CreatedAt, time.Second)
	require.Nil(t, payload[0].DeletedAt)
}

func TestHierarchyHandlerGetChildrenUnauthorized(t *testing.T) {
	handler := hierarchy.NewHandler(hierarchy.NewService(&hierarchyStoreFake{}, &nodeStoreFake{}))

	req := withRouteParam(httptest.NewRequest(http.MethodGet, "/nodes/:id/children", nil), "id", testUUID1)
	res := httptest.NewRecorder()

	handler.GetChildren(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
	testutil.AssertErrorJSON(t, res, "User ID not found in context")
}

func TestHierarchyHandlerGetChildrenParentNotFound(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	handler := hierarchy.NewHandler(hierarchy.NewService(&hierarchyStoreFake{}, &nodeStoreFake{}))

	req := withRouteParam(withNodeUserContext(
		t,
		httptest.NewRequest(http.MethodGet, "/nodes/:id/children", nil),
		userID,
	), "id", testUUID2)
	res := httptest.NewRecorder()

	handler.GetChildren(res, req)

	require.Equal(t, http.StatusNotFound, res.Code)
	testutil.AssertErrorJSON(t, res, "parent not found")
}

func TestHandlerGetChildrenInvalidUUID(t *testing.T) {
	handler := hierarchy.NewHandler(hierarchy.NewService(&hierarchyStoreFake{}, &nodeStoreFake{}))

	req := withRouteParam(httptest.NewRequest(http.MethodGet, "/nodes/:id/children", nil), "id", testUUIDBad)
	res := httptest.NewRecorder()

	handler.GetChildren(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "invalid node id format")
}

func TestHierarchyHandlerInternalErrors(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testUUID1)
	tests := []struct {
		name     string
		method   string
		routeKey string
		path     string
		body     any
		setup    func(t *testing.T) (*hierarchyStoreFake, *nodeStoreFake)
		execute  func(h *hierarchy.Handler, w http.ResponseWriter, r *http.Request)
	}{
		{
			name:     "get children db error",
			method:   http.MethodGet,
			routeKey: "id",
			path:     "/nodes/" + testUUID1 + "/children",
			body:     nil,
			setup: func(t *testing.T) (*hierarchyStoreFake, *nodeStoreFake) {
				return &hierarchyStoreFake{}, &nodeStoreFake{getNodeByIDErr: sql.ErrConnDone}
			},
			execute: func(h *hierarchy.Handler, w http.ResponseWriter, r *http.Request) { h.GetChildren(w, r) },
		},
		{
			name:     "get children store error",
			method:   http.MethodGet,
			routeKey: "id",
			path:     "/nodes/" + testUUID1 + "/children",
			body:     nil,
			setup: func(t *testing.T) (*hierarchyStoreFake, *nodeStoreFake) {
				return &hierarchyStoreFake{childrenErr: sql.ErrConnDone},
					&nodeStoreFake{getNodeByIDResult: map[string]nodeDB.Node{
						testUUID1: {ID: testutil.UUIDFromStringT(t, testUUID1)},
					}}
			},
			execute: func(h *hierarchy.Handler, w http.ResponseWriter, r *http.Request) { h.GetChildren(w, r) },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake, fakeNodeStore := tc.setup(t)
			handler := hierarchy.NewHandler(hierarchy.NewService(fake, fakeNodeStore))

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
