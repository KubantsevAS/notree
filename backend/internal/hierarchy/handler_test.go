package hierarchy_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy/rank"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestHandlerGetChildren(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID2)
	updatedAt := time.Now()
	nodes := nodeSetOf(sqlcHierarchy.Node{ID: parentID, UserID: userID})
	fake := &hierarchyStoreFake{
		children: []sqlcHierarchy.Node{{
			ID:        testutil.UUIDFromStringT(t, testutil.UUID3),
			UserID:    userID,
			ParentID:  parentID,
			Type:      sqlcHierarchy.NodeTypeNote,
			Title:     "child",
			SortOrder: 1,
			CreatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
		}},
	}
	handler := hierarchy.NewHandler(newTestService(fake, nodes))

	req := testutil.WithRouteParam(testutil.WithUserID(httptest.NewRequest(http.MethodGet, "/nodes/:id/children", nil), userID), "id", parentID.String())
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

func TestHandlerGetParent(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID3)
	grandparentID := testutil.UUIDFromStringT(t, testutil.UUID4)
	updatedAt := time.Now()

	nodes := nodeSetOf(sqlcHierarchy.Node{ID: nodeID, UserID: userID, ParentID: parentID})
	fake := &hierarchyStoreFake{
		parent: sqlcHierarchy.Node{
			ID:        parentID,
			UserID:    userID,
			ParentID:  grandparentID,
			Type:      sqlcHierarchy.NodeTypeNote,
			Title:     "parent node",
			SortOrder: 5,
			CreatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
		},
	}
	handler := hierarchy.NewHandler(newTestService(fake, nodes))

	req := testutil.WithRouteParam(testutil.WithUserID(httptest.NewRequest(http.MethodGet, "/nodes/:id/parent", nil), userID), "id", nodeID.String())
	res := httptest.NewRecorder()

	handler.GetParent(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	var payload hierarchy.NodeResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, parentID.String(), payload.ID)
	require.Equal(t, "parent node", payload.Title)
	require.WithinDuration(t, updatedAt, *payload.CreatedAt, time.Second)
	require.Nil(t, payload.DeletedAt)
}

func TestHandlerGetAncestors(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID3)
	rootID := testutil.UUIDFromStringT(t, testutil.UUID4)
	updatedAt := time.Now()

	nodes := nodeSetOf(sqlcHierarchy.Node{ID: nodeID, UserID: userID, ParentID: parentID})
	fake := &hierarchyStoreFake{
		ancestors: []sqlcHierarchy.Node{
			{
				ID:        rootID,
				UserID:    userID,
				Type:      sqlcHierarchy.NodeTypeFolder,
				Title:     "root",
				CreatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
			},
			{
				ID:        parentID,
				UserID:    userID,
				ParentID:  rootID,
				Type:      sqlcHierarchy.NodeTypeFolder,
				Title:     "parent",
				CreatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
			},
		},
	}
	handler := hierarchy.NewHandler(newTestService(fake, nodes))

	req := testutil.WithRouteParam(testutil.WithUserID(httptest.NewRequest(http.MethodGet, "/nodes/:id/ancestors", nil), userID), "id", nodeID.String())
	res := httptest.NewRecorder()

	handler.GetAncestors(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var payload hierarchy.GetAncestorsResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Len(t, payload, 2)
	require.Equal(t, "root", payload[0].Title)
	require.Nil(t, payload[0].ParentID)
	require.Equal(t, "parent", payload[1].Title)
	require.Equal(t, rootID.String(), *payload[1].ParentID)
}

func TestHandler_NodeIsRoot(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	rootID := testutil.UUIDFromStringT(t, testutil.UUID2)

	tests := []struct {
		name string
		call handlerFunc
	}{
		{"ancestors", (*hierarchy.Handler).GetAncestors},
		{"breadcrumbs", (*hierarchy.Handler).GetBreadcrumbs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := nodeSetOf(sqlcHierarchy.Node{ID: rootID, UserID: userID})
			handler := hierarchy.NewHandler(newTestService(&hierarchyStoreFake{}, nodes))
			req := testutil.WithUserID(newReadRequest(tt.name, rootID.String()), userID)
			res := httptest.NewRecorder()

			tt.call(handler, res, req)

			require.Equal(t, http.StatusOK, res.Code)
			require.JSONEq(t, "[]", res.Body.String())
		})
	}
}

func TestHandlerGetBreadcrumbs(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID3)
	rootID := testutil.UUIDFromStringT(t, testutil.UUID4)
	nodes := nodeSetOf(sqlcHierarchy.Node{ID: nodeID, UserID: userID, ParentID: parentID})
	store := &hierarchyStoreFake{
		breadcrumbs: []sqlcHierarchy.GetBreadcrumbsRow{
			{ID: rootID, Title: "root"},
			{ID: parentID, Title: "parent"},
		},
	}
	handler := hierarchy.NewHandler(newTestService(store, nodes))
	req := testutil.WithRouteParam(
		testutil.WithUserID(httptest.NewRequest(http.MethodGet, "/nodes/:id/breadcrumbs", nil), userID),
		"id",
		nodeID.String(),
	)
	res := httptest.NewRecorder()

	handler.GetBreadcrumbs(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.JSONEq(t, `[{"id":"`+rootID.String()+`","title":"root"},{"id":"`+parentID.String()+`","title":"parent"}]`, res.Body.String())
	require.Equal(t, sqlcHierarchy.GetBreadcrumbsParams{ID: nodeID, UserID: userID}, store.breadcrumbsParam)
}

func TestHandlerGetDescendants(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	rootID := testutil.UUIDFromStringT(t, testutil.UUID2)
	childID := testutil.UUIDFromStringT(t, testutil.UUID3)
	grandchildID := testutil.UUIDFromStringT(t, testutil.UUID4)
	updatedAt := time.Now()

	nodes := nodeSetOf(sqlcHierarchy.Node{ID: rootID, UserID: userID})
	fake := &hierarchyStoreFake{
		descendants: []sqlcHierarchy.Node{
			{
				ID:        childID,
				UserID:    userID,
				ParentID:  rootID,
				Type:      sqlcHierarchy.NodeTypeFolder,
				Title:     "child",
				CreatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
			},
			{
				ID:        grandchildID,
				UserID:    userID,
				ParentID:  childID,
				Type:      sqlcHierarchy.NodeTypeNote,
				Title:     "grandchild",
				CreatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
			},
		},
	}
	handler := hierarchy.NewHandler(newTestService(fake, nodes))

	req := testutil.WithRouteParam(testutil.WithUserID(httptest.NewRequest(http.MethodGet, "/nodes/:id/descendants", nil), userID), "id", rootID.String())
	res := httptest.NewRecorder()

	handler.GetDescendants(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var payload hierarchy.GetDescendantsResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Len(t, payload, 2)
	require.Equal(t, "child", payload[0].Title)
	require.Equal(t, "grandchild", payload[1].Title)
	require.Equal(t, childID.String(), *payload[1].ParentID)
	require.Nil(t, payload[0].DeletedAt)
}

func TestHandlerGetSubtree(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	childID := testutil.UUIDFromStringT(t, testutil.UUID3)
	updatedAt := time.Now()

	nodes := nodeSetOf(sqlcHierarchy.Node{ID: nodeID, UserID: userID})
	fake := &hierarchyStoreFake{
		subtree: []sqlcHierarchy.Node{
			{
				ID:        nodeID,
				UserID:    userID,
				Type:      sqlcHierarchy.NodeTypeFolder,
				Title:     "node",
				CreatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
			},
			{
				ID:        childID,
				UserID:    userID,
				ParentID:  nodeID,
				Type:      sqlcHierarchy.NodeTypeNote,
				Title:     "child",
				CreatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
			},
		},
	}
	handler := hierarchy.NewHandler(newTestService(fake, nodes))

	req := testutil.WithRouteParam(testutil.WithUserID(httptest.NewRequest(http.MethodGet, "/nodes/:id/subtree", nil), userID), "id", nodeID.String())
	res := httptest.NewRecorder()

	handler.GetSubtree(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var payload hierarchy.GetSubtreeResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Len(t, payload, 2)
	require.Equal(t, "node", payload[0].Title)
	require.Equal(t, "child", payload[1].Title)
	require.Equal(t, nodeID.String(), *payload[1].ParentID)
}

func TestHandlerGetRoot(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	rootID := testutil.UUIDFromStringT(t, testutil.UUID3)
	updatedAt := time.Now()

	nodes := nodeSetOf(sqlcHierarchy.Node{ID: nodeID, UserID: userID, ParentID: rootID})
	fake := &hierarchyStoreFake{
		root: sqlcHierarchy.Node{
			ID:        rootID,
			UserID:    userID,
			Type:      sqlcHierarchy.NodeTypeFolder,
			Title:     "root",
			CreatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
		},
	}
	handler := hierarchy.NewHandler(newTestService(fake, nodes))

	req := testutil.WithRouteParam(testutil.WithUserID(httptest.NewRequest(http.MethodGet, "/nodes/:id/root", nil), userID), "id", nodeID.String())
	res := httptest.NewRecorder()

	handler.GetRoot(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var payload hierarchy.NodeResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, rootID.String(), payload.ID)
	require.Equal(t, "root", payload.Title)
	require.Nil(t, payload.ParentID)
}

type handlerFunc func(*hierarchy.Handler, http.ResponseWriter, *http.Request)

var readEndpoints = []struct {
	name string
	call handlerFunc
}{
	{"children", (*hierarchy.Handler).GetChildren},
	{"parent", (*hierarchy.Handler).GetParent},
	{"ancestors", (*hierarchy.Handler).GetAncestors},
	{"descendants", (*hierarchy.Handler).GetDescendants},
	{"subtree", (*hierarchy.Handler).GetSubtree},
	{"root", (*hierarchy.Handler).GetRoot},
	{"breadcrumbs", (*hierarchy.Handler).GetBreadcrumbs},
}

func newReadRequest(endpoint, nodeID string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/nodes/"+nodeID+"/"+endpoint, nil)
	return testutil.WithRouteParam(req, "id", nodeID)
}

func TestHandler_Unauthorized(t *testing.T) {
	for _, ep := range readEndpoints {
		t.Run(ep.name, func(t *testing.T) {
			handler := hierarchy.NewHandler(newTestService(&hierarchyStoreFake{}, &nodeSet{}))
			res := httptest.NewRecorder()

			ep.call(handler, res, newReadRequest(ep.name, testutil.UUID1))

			require.Equal(t, http.StatusUnauthorized, res.Code)
			wantMessage := "unauthorized"
			testutil.AssertErrorJSON(t, res, wantMessage)
		})
	}
}

func TestHandler_InvalidUUID(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)

	for _, ep := range readEndpoints {
		t.Run(ep.name, func(t *testing.T) {
			handler := hierarchy.NewHandler(newTestService(&hierarchyStoreFake{}, &nodeSet{}))
			res := httptest.NewRecorder()

			ep.call(handler, res, testutil.WithUserID(newReadRequest(ep.name, testutil.BadUUID), userID))

			require.Equal(t, http.StatusBadRequest, res.Code)
			testutil.AssertErrorJSON(t, res, "invalid node id format")
		})
	}
}

func TestHandler_NodeNotFound(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)

	for _, ep := range readEndpoints {
		t.Run(ep.name, func(t *testing.T) {
			handler := hierarchy.NewHandler(newTestService(&hierarchyStoreFake{}, &nodeSet{}))
			res := httptest.NewRecorder()

			ep.call(handler, res, testutil.WithUserID(newReadRequest(ep.name, testutil.UUID2), userID))

			require.Equal(t, http.StatusNotFound, res.Code)
			testutil.AssertErrorJSON(t, res, "node not found")
		})
	}
}

func TestHandler_NodeLookupError(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)

	for _, ep := range readEndpoints {
		t.Run(ep.name, func(t *testing.T) {
			nodes := &nodeSet{err: sql.ErrConnDone}
			handler := hierarchy.NewHandler(newTestService(&hierarchyStoreFake{}, nodes))
			res := httptest.NewRecorder()

			ep.call(handler, res, testutil.WithUserID(newReadRequest(ep.name, testutil.UUID2), userID))

			require.Equal(t, http.StatusInternalServerError, res.Code)
			testutil.AssertErrorJSON(t, res, "internal server error")
		})
	}
}

func TestHandler_StoreError(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	nodes := nodeSetOf(sqlcHierarchy.Node{ID: nodeID, UserID: userID, ParentID: testutil.UUIDFromStringT(t, testutil.UUID3)})

	tests := []struct {
		endpoint string
		call     handlerFunc
		store    *hierarchyStoreFake
	}{
		{"children", (*hierarchy.Handler).GetChildren, &hierarchyStoreFake{childrenErr: sql.ErrConnDone}},
		{"parent", (*hierarchy.Handler).GetParent, &hierarchyStoreFake{parentErr: sql.ErrConnDone}},
		{"ancestors", (*hierarchy.Handler).GetAncestors, &hierarchyStoreFake{ancestorsErr: sql.ErrConnDone}},
		{"descendants", (*hierarchy.Handler).GetDescendants, &hierarchyStoreFake{descendantsErr: sql.ErrConnDone}},
		{"subtree", (*hierarchy.Handler).GetSubtree, &hierarchyStoreFake{subtreeErr: sql.ErrConnDone}},
		{"root", (*hierarchy.Handler).GetRoot, &hierarchyStoreFake{rootErr: sql.ErrConnDone}},
		{"breadcrumbs", (*hierarchy.Handler).GetBreadcrumbs, &hierarchyStoreFake{breadcrumbsErr: sql.ErrConnDone}},
	}

	for _, tc := range tests {
		t.Run(tc.endpoint, func(t *testing.T) {
			handler := hierarchy.NewHandler(newTestService(tc.store, nodes))
			res := httptest.NewRecorder()

			tc.call(handler, res, testutil.WithUserID(newReadRequest(tc.endpoint, nodeID.String()), userID))

			require.Equal(t, http.StatusInternalServerError, res.Code)
			testutil.AssertErrorJSON(t, res, "internal server error")
		})
	}
}

func TestHandler_DomainErrors(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	rootNode := sqlcHierarchy.Node{ID: nodeID, UserID: userID}
	childNode := sqlcHierarchy.Node{ID: nodeID, UserID: userID, ParentID: testutil.UUIDFromStringT(t, testutil.UUID3)}

	tests := []struct {
		name       string
		endpoint   string
		call       handlerFunc
		node       sqlcHierarchy.Node
		store      *hierarchyStoreFake
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "parent of root node",
			endpoint:   "parent",
			call:       (*hierarchy.Handler).GetParent,
			node:       rootNode,
			store:      &hierarchyStoreFake{parentErr: errors.New("GetParent must not be called for root node")},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "parent row missing",
			endpoint:   "parent",
			call:       (*hierarchy.Handler).GetParent,
			node:       childNode,
			store:      &hierarchyStoreFake{parentErr: pgx.ErrNoRows},
			wantStatus: http.StatusNotFound,
			wantMsg:    "parent not found",
		},
		{
			name:       "root not found",
			endpoint:   "root",
			call:       (*hierarchy.Handler).GetRoot,
			node:       childNode,
			store:      &hierarchyStoreFake{rootErr: pgx.ErrNoRows},
			wantStatus: http.StatusNotFound,
			wantMsg:    "root not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := hierarchy.NewHandler(newTestService(tc.store, nodeSetOf(tc.node)))
			res := httptest.NewRecorder()

			tc.call(handler, res, testutil.WithUserID(newReadRequest(tc.endpoint, nodeID.String()), userID))

			require.Equal(t, tc.wantStatus, res.Code)
			if tc.wantMsg == "" {
				require.Empty(t, res.Body.String())
				return
			}
			testutil.AssertErrorJSON(t, res, tc.wantMsg)
		})
	}
}

func newMoveRequest(t *testing.T, userID pgtype.UUID, nodeID string, body any) *http.Request {
	t.Helper()
	return testutil.WithRouteParam(testutil.WithUserID(testutil.NewJSONRequest(t, http.MethodPost, "/nodes/"+nodeID+"/move", body), userID), "id", nodeID)
}

func TestHandlerMove(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	parentID := testutil.UUIDFromStringT(t, testutil.UUID3)
	nodes := nodeSetOf(
		sqlcHierarchy.Node{ID: nodeID, UserID: userID},
		sqlcHierarchy.Node{ID: parentID, UserID: userID},
	)
	store := &hierarchyStoreFake{}
	handler := hierarchy.NewHandler(newTestService(store, nodes))
	res := httptest.NewRecorder()

	handler.Move(res, newMoveRequest(t, userID, testutil.UUID2, map[string]any{"parent_id": testutil.UUID3, "before_id": nil}))

	require.Equal(t, http.StatusOK, res.Code)
	var payload hierarchy.MoveNodeResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, testutil.UUID3, *payload.ParentID)
	require.Equal(t, rank.Gap, payload.SortOrder)
	require.Len(t, store.moveCalls, 1)
}

func TestHandlerMove_Errors(t *testing.T) {
	userID := testutil.UUIDFromStringT(t, testutil.UUID1)
	nodeID := testutil.UUIDFromStringT(t, testutil.UUID2)
	moving := sqlcHierarchy.Node{ID: nodeID, UserID: userID}
	parent := sqlcHierarchy.Node{ID: testutil.UUIDFromStringT(t, testutil.UUID3), UserID: userID}
	toEnd := func(parentID any) map[string]any { return map[string]any{"parent_id": parentID, "before_id": nil} }
	cycle := "node cannot be a descendant of itself (circular reference)"

	tests := []struct {
		name       string
		nodeID     string
		body       any
		store      *hierarchyStoreFake
		nodes      *nodeSet
		wantStatus int
		wantMsg    string
	}{
		{name: "parent_id missing", body: map[string]any{"before_id": nil}, wantStatus: http.StatusBadRequest, wantMsg: "parent_id is required"},
		{name: "before_id missing", body: map[string]any{"parent_id": nil}, wantStatus: http.StatusBadRequest, wantMsg: "before_id is required"},
		{name: "invalid node id", nodeID: testutil.BadUUID, body: toEnd(nil), wantStatus: http.StatusBadRequest, wantMsg: "invalid node id format"},
		{name: "invalid parent id", body: toEnd(testutil.BadUUID), wantStatus: http.StatusBadRequest, wantMsg: "invalid parent id"},
		{name: "invalid before id", body: map[string]any{"parent_id": nil, "before_id": testutil.BadUUID}, wantStatus: http.StatusBadRequest, wantMsg: "invalid before id"},
		{name: "node not found", body: toEnd(nil), wantStatus: http.StatusNotFound, wantMsg: "node not found"},
		{name: "parent not found", body: toEnd(testutil.UUID3), nodes: nodeSetOf(moving), wantStatus: http.StatusBadRequest, wantMsg: "parent not found"},
		{name: "self parent", body: toEnd(testutil.UUID2), nodes: nodeSetOf(moving), wantStatus: http.StatusConflict, wantMsg: cycle},
		{
			name: "parent inside subtree", body: toEnd(testutil.UUID3),
			store: &hierarchyStoreFake{inSubtree: true}, nodes: nodeSetOf(moving, parent),
			wantStatus: http.StatusConflict, wantMsg: cycle,
		},
		{
			name: "before not a sibling", body: map[string]any{"parent_id": testutil.UUID3, "before_id": testutil.UUID4},
			nodes:      nodeSetOf(moving, parent),
			wantStatus: http.StatusBadRequest, wantMsg: "before_id must reference another child of parent_id",
		},
		{
			name: "db error", body: toEnd(nil),
			store: &hierarchyStoreFake{moveErr: sql.ErrConnDone}, nodes: nodeSetOf(moving),
			wantStatus: http.StatusInternalServerError, wantMsg: "internal server error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := tc.store
			if store == nil {
				store = &hierarchyStoreFake{}
			}
			nodes := tc.nodes
			if nodes == nil {
				nodes = &nodeSet{}
			}
			nodeIDParam := tc.nodeID
			if nodeIDParam == "" {
				nodeIDParam = testutil.UUID2
			}
			handler := hierarchy.NewHandler(newTestService(store, nodes))
			res := httptest.NewRecorder()

			handler.Move(res, newMoveRequest(t, userID, nodeIDParam, tc.body))

			require.Equal(t, tc.wantStatus, res.Code)
			testutil.AssertErrorJSON(t, res, tc.wantMsg)
		})
	}
}

func TestHandlerMove_Unauthorized(t *testing.T) {
	handler := hierarchy.NewHandler(newTestService(&hierarchyStoreFake{}, &nodeSet{}))
	req := testutil.WithRouteParam(
		testutil.NewJSONRequest(t, http.MethodPost, "/nodes/"+testutil.UUID1+"/move", map[string]any{"parent_id": nil, "before_id": nil}),
		"id", testutil.UUID1,
	)
	res := httptest.NewRecorder()

	handler.Move(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
	testutil.AssertErrorJSON(t, res, "unauthorized")
}
