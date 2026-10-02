package node

import (
	"context"
	"errors"

	"github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/KubantsevAS/notree/backend/internal/domain"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy/rank"
	"github.com/KubantsevAS/notree/backend/internal/http/httputil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Store interface {
	CreateNode(context.Context, node.CreateNodeParams) (node.Node, error)
	GetNodeByID(context.Context, node.GetNodeByIDParams) (node.Node, error)
	SoftDeleteNodeCascade(context.Context, node.SoftDeleteNodeCascadeParams) ([]pgtype.UUID, error)
	UpdateNode(context.Context, node.UpdateNodeParams) (node.UpdateNodeRow, error)
}

type Service struct {
	store Store
}

func NewService(repo Store) *Service {
	return &Service{store: repo}
}

func (s *Service) CreateNode(ctx context.Context, userID pgtype.UUID, req *CreateNodeRequest) (CreateNodeResponse, error) {
	parentID := pgtype.UUID{Valid: false}
	if req.ParentID != nil && *req.ParentID != "" {
		parsedID, err := httputil.PgUUIDFromString(req.ParentID)
		if err != nil {
			return CreateNodeResponse{}, ErrInvalidParentID
		}

		if _, err := s.store.GetNodeByID(ctx, node.GetNodeByIDParams{ID: parsedID, UserID: userID}); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return CreateNodeResponse{}, ErrParentNotFound
			}
			return CreateNodeResponse{}, err
		}

		parentID = parsedID
	}

	dbParams := node.CreateNodeParams{
		UserID:   userID,
		ParentID: parentID,
		Type:     node.NodeType(req.Type),
		Title:    req.Title,
		Gap:      rank.Gap,
	}

	nodeRow, err := s.store.CreateNode(ctx, dbParams)
	if err != nil {
		return CreateNodeResponse{}, err
	}

	response := CreateNodeResponse{
		ID:        nodeRow.ID.String(),
		ParentID:  nodeRow.ParentID.String(),
		Type:      string(nodeRow.Type),
		Title:     nodeRow.Title,
		SortOrder: nodeRow.SortOrder,
		CreatedAt: &nodeRow.CreatedAt.Time,
	}

	return response, nil
}

func (s *Service) DeleteNode(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID) error {
	dbParams := &node.SoftDeleteNodeCascadeParams{
		ID:     nodeID,
		UserID: userID,
	}

	deletedIds, err := s.store.SoftDeleteNodeCascade(ctx, *dbParams)
	if err != nil {
		return err
	}

	if len(deletedIds) == 0 {
		return ErrNodeNotFoundOrNoAccess
	}

	return nil
}

func (s *Service) UpdateNode(ctx context.Context, nodeID pgtype.UUID, userID pgtype.UUID, req *UpdateNodeRequest) (UpdateNodeResponse, error) {
	if req.Type == nil && req.Title == nil {
		return UpdateNodeResponse{}, domain.ErrEmptyUpdate
	}

	dbParams := &node.UpdateNodeParams{
		ID:     nodeID,
		UserID: userID,
	}
	if req.Type != nil {
		dbParams.Type = node.NullNodeType{NodeType: node.NodeType(*req.Type), Valid: true}
	}
	if req.Title != nil {
		dbParams.Title = httputil.PgTextFromString(req.Title)
	}

	dbRow, err := s.store.UpdateNode(ctx, *dbParams)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UpdateNodeResponse{}, ErrNodeNotFoundOrNoAccess
		}
		return UpdateNodeResponse{}, err
	}

	response := UpdateNodeResponse{
		Type:      string(dbRow.Type),
		Title:     dbRow.Title,
		UpdatedAt: &dbRow.UpdatedAt.Time,
	}

	return response, nil
}
