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

	if _, err := service.CreateNode(context.Background(), pgtype.UUID{}, request); err != nil {
		t.Fatalf("first CreateNode() error = %v", err)
	}
	if _, err := service.CreateNode(context.Background(), pgtype.UUID{}, request); err != nil {
		t.Fatalf("second CreateNode() error = %v", err)
	}

	if len(fake.createParams) != 2 {
		t.Fatalf("CreateNode() calls = %d, want 2", len(fake.createParams))
	}
	if fake.createParams[1].SortOrder <= fake.createParams[0].SortOrder {
		t.Fatalf("sort_order did not increase: first=%d, second=%d", fake.createParams[0].SortOrder, fake.createParams[1].SortOrder)
	}
}

func TestNodeServiceCreateNodeValidatesParentID(t *testing.T) {
	t.Run("invalid parent uuid", func(t *testing.T) {
		_, err := NewService(&nodeRepositoryFake{}).CreateNode(context.Background(), pgtype.UUID{}, &CreateNodeRequest{
			ParentID: strPtr("bad-uuid"),
			Type:     "note",
			Title:    "child",
		})
		if !errors.Is(err, ErrInvalidParentID) {
			t.Fatalf("CreateNode() error = %v, want %v", err, ErrInvalidParentID)
		}
	})

	t.Run("parent not found", func(t *testing.T) {
		parentID := newUUIDFromString(t, "11111111-1111-4111-8111-111111111111")
		_, err := NewService(&nodeRepositoryFake{}).CreateNode(context.Background(), pgtype.UUID{}, &CreateNodeRequest{
			ParentID: strPtr(parentID.String()),
			Type:     "note",
			Title:    "child",
		})
		if !errors.Is(err, ErrParentNotFound) {
			t.Fatalf("CreateNode() error = %v, want %v", err, ErrParentNotFound)
		}
	})
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
	if err != nil {
		t.Fatalf("GetChildren() error = %v", err)
	}
	if len(children) != 2 {
		t.Fatalf("GetChildren() items = %d, want 2", len(children))
	}
	if children[0].ID != childOneID.String() || children[1].ID != childTwoID.String() {
		t.Fatalf("GetChildren() order = [%s, %s], want sort_order order [%s, %s]", children[0].ID, children[1].ID, childOneID, childTwoID)
	}
}

func TestNodeServiceGetChildrenReturnsErrParentNotFound(t *testing.T) {
	parentID := newUUIDFromString(t, "11111111-1111-4111-8111-111111111111")
	userID := newUUIDFromString(t, "22222222-2222-4222-8222-222222222222")

	_, err := NewService(&nodeRepositoryFake{}).GetChildren(context.Background(), parentID, userID)
	if !errors.Is(err, ErrParentNotFound) {
		t.Fatalf("GetChildren() error = %v, want %v", err, ErrParentNotFound)
	}
}

func TestNodeServiceDeleteNode(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		nodeID := newUUIDFromString(t, "33333333-3333-4333-8333-333333333333")
		userID := newUUIDFromString(t, "44444444-4444-4444-8444-444444444444")
		fake := &nodeRepositoryFake{softDeleteResult: []pgtype.UUID{nodeID}}

		err := NewService(fake).DeleteNode(context.Background(), nodeID, userID)
		if err != nil {
			t.Fatalf("DeleteNode() error = %v", err)
		}
		if len(fake.softDeleteParams) != 1 {
			t.Fatalf("SoftDeleteNodeCascade() calls = %d, want 1", len(fake.softDeleteParams))
		}
	})

	t.Run("not found or no access", func(t *testing.T) {
		nodeID := newUUIDFromString(t, "55555555-5555-4555-8555-555555555555")
		userID := newUUIDFromString(t, "66666666-6666-4666-8666-666666666666")
		fake := &nodeRepositoryFake{}

		err := NewService(fake).DeleteNode(context.Background(), nodeID, userID)
		if !errors.Is(err, ErrNodeNotFoundOrNoAccess) {
			t.Fatalf("DeleteNode() error = %v, want %v", err, ErrNodeNotFoundOrNoAccess)
		}
	})
}

func TestNodeServiceUpdateNode(t *testing.T) {
	t.Run("empty request", func(t *testing.T) {
		_, err := NewService(&nodeRepositoryFake{}).UpdateNode(context.Background(), pgtype.UUID{}, pgtype.UUID{}, &UpdateNodeRequest{})
		if !errors.Is(err, domain.ErrEmptyUpdate) {
			t.Fatalf("UpdateNode() error = %v, want %v", err, domain.ErrEmptyUpdate)
		}
	})

	t.Run("success", func(t *testing.T) {
		nodeID := newUUIDFromString(t, "77777777-7777-4777-8777-777777777777")
		userID := newUUIDFromString(t, "88888888-8888-4888-8888-888888888888")
		updatedAt := time.Now()
		fake := &nodeRepositoryFake{updateResult: node.UpdateNodeRow{Type: node.NodeTypeTask, Title: "done", UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true}}}
		resp, err := NewService(fake).UpdateNode(context.Background(), nodeID, userID, &UpdateNodeRequest{Type: strPtr("task"), Title: strPtr("done")})
		if err != nil {
			t.Fatalf("UpdateNode() error = %v", err)
		}
		if resp.Type != "task" || resp.Title != "done" {
			t.Fatalf("UpdateNode() response = %+v, want type=task title=done", resp)
		}
	})

	t.Run("node not found or no access", func(t *testing.T) {
		fake := &nodeRepositoryFake{updateErr: pgx.ErrNoRows}
		_, err := NewService(fake).UpdateNode(context.Background(), pgtype.UUID{}, pgtype.UUID{}, &UpdateNodeRequest{Title: strPtr("new title")})
		if !errors.Is(err, ErrNodeNotFoundOrNoAccess) {
			t.Fatalf("UpdateNode() error = %v, want %v", err, ErrNodeNotFoundOrNoAccess)
		}
	})
}

func TestNodeServiceMoveNode(t *testing.T) {
	t.Run("empty request", func(t *testing.T) {
		_, err := NewService(&nodeRepositoryFake{}).MoveNode(context.Background(), pgtype.UUID{}, pgtype.UUID{}, &MoveNodeRequest{})
		if !errors.Is(err, domain.ErrEmptyUpdate) {
			t.Fatalf("MoveNode() error = %v, want %v", err, domain.ErrEmptyUpdate)
		}
	})

	t.Run("invalid parent uuid", func(t *testing.T) {
		_, err := NewService(&nodeRepositoryFake{}).MoveNode(context.Background(), pgtype.UUID{}, pgtype.UUID{}, &MoveNodeRequest{ParentID: NullableString{Value: strPtr("bad-uuid"), IsSet: true}})
		if !errors.Is(err, ErrInvalidParentID) {
			t.Fatalf("MoveNode() error = %v, want %v", err, ErrInvalidParentID)
		}
	})

	t.Run("node cannot be a descendant of itself", func(t *testing.T) {
		nodeID := newUUIDFromString(t, "99999999-9999-4999-8999-999999999999")
		parentID := newUUIDFromString(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
		fake := &nodeRepositoryFake{
			getNodeByIDResult: map[string]node.Node{parentID.String(): {ID: parentID}},
			ancestors:         map[string][]pgtype.UUID{parentID.String(): {nodeID}},
		}
		_, err := NewService(fake).MoveNode(context.Background(), nodeID, pgtype.UUID{}, &MoveNodeRequest{ParentID: NullableString{Value: strPtr(parentID.String()), IsSet: true}})
		if !errors.Is(err, ErrNodeCannotBeADescendantOfItself) {
			t.Fatalf("MoveNode() error = %v, want %v", err, ErrNodeCannotBeADescendantOfItself)
		}
	})

	t.Run("success", func(t *testing.T) {
		nodeID := newUUIDFromString(t, "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
		userID := newUUIDFromString(t, "cccccccc-cccc-4ccc-8ccc-cccccccccccc")
		parentID := newUUIDFromString(t, "dddddddd-dddd-4ddd-8ddd-dddddddddddd")
		updatedAt := time.Now()
		fake := &nodeRepositoryFake{
			getNodeByIDResult: map[string]node.Node{parentID.String(): {ID: parentID, UserID: userID}},
			moveResult: node.MoveNodeRow{
				ParentID:  parentID,
				SortOrder: 42,
				UpdatedAt: pgtype.Timestamptz{Time: updatedAt, Valid: true},
			},
		}
		resp, err := NewService(fake).MoveNode(context.Background(), nodeID, userID, &MoveNodeRequest{ParentID: NullableString{Value: strPtr(parentID.String()), IsSet: true}, SortOrder: int64Ptr(42)})
		if err != nil {
			t.Fatalf("MoveNode() error = %v", err)
		}
		if resp.ParentID == nil || *resp.ParentID != parentID.String() {
			t.Fatalf("MoveNode() parent = %#v, want %s", resp.ParentID, parentID.String())
		}
		if resp.SortOrder != 42 {
			t.Fatalf("MoveNode() sort_order = %d, want 42", resp.SortOrder)
		}
	})
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
	if err != nil {
		t.Fatalf("CreateNode() error = %v", err)
	}
	if resp.ParentID != parentID.String() || resp.Type != "note" || resp.Title != "child" {
		t.Fatalf("CreateNode() response = %+v, want parent=%s type=note title=child", resp, parentID.String())
	}
	if len(fake.createParams) != 1 {
		t.Fatalf("CreateNode() calls = %d, want 1", len(fake.createParams))
	}
}
