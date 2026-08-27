package service

import (
	"context"
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/jackc/pgx/v5/pgtype"
)

type nodeRepositoryFake struct {
	createParams []node.CreateNodeParams
	parent       node.Node
	children     []node.Node
}

func (f *nodeRepositoryFake) CreateNode(_ context.Context, params node.CreateNodeParams) (node.Node, error) {
	f.createParams = append(f.createParams, params)
	return node.Node{
		ID:        pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		ParentID:  params.ParentID,
		Type:      params.Type,
		Title:     params.Title,
		SortOrder: params.SortOrder,
	}, nil
}

func (f *nodeRepositoryFake) GetChildren(context.Context, node.GetChildrenParams) ([]node.Node, error) {
	return f.children, nil
}

func (f *nodeRepositoryFake) GetNodeAncestors(context.Context, pgtype.UUID) ([]pgtype.UUID, error) {
	return nil, nil
}

func (f *nodeRepositoryFake) GetNodeByID(context.Context, node.GetNodeByIDParams) (node.Node, error) {
	return f.parent, nil
}

func (f *nodeRepositoryFake) MoveNode(context.Context, node.MoveNodeParams) (node.MoveNodeRow, error) {
	return node.MoveNodeRow{}, nil
}

func (f *nodeRepositoryFake) SoftDeleteNodeCascade(context.Context, node.SoftDeleteNodeCascadeParams) ([]pgtype.UUID, error) {
	return nil, nil
}

func (f *nodeRepositoryFake) UpdateNode(context.Context, node.UpdateNodeParams) (node.UpdateNodeRow, error) {
	return node.UpdateNodeRow{}, nil
}

func TestNodeServiceCreateNodeSortOrderIncreases(t *testing.T) {
	fake := &nodeRepositoryFake{}
	service := NewNodeService(fake)
	request := &dto.CreateNodeRequest{Type: "note", Title: "test"}

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

func TestNodeServiceGetChildrenReturnsChildrenInSortOrder(t *testing.T) {
	parentID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	userID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	childOneID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}
	childTwoID := pgtype.UUID{Bytes: [16]byte{4}, Valid: true}
	fake := &nodeRepositoryFake{
		parent: node.Node{ID: parentID},
		children: []node.Node{
			{ID: childOneID, ParentID: parentID, UserID: userID, Title: "first", SortOrder: 10},
			{ID: childTwoID, ParentID: parentID, UserID: userID, Title: "second", SortOrder: 20},
		},
	}

	children, err := NewNodeService(fake).GetChildren(context.Background(), parentID, userID)
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
