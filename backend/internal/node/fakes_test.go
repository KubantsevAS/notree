package node_test

import (
	"context"

	nodeDb "github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type nodeStoreFake struct {
	createParams      []nodeDb.CreateNodeParams
	createResult      nodeDb.Node
	createErr         error
	getNodeByIDResult map[string]nodeDb.Node
	getNodeByIDErr    error
	getNodeByIDCalls  []nodeDb.GetNodeByIDParams
	softDeleteParams  []nodeDb.SoftDeleteNodeCascadeParams
	softDeleteResult  []pgtype.UUID
	softDeleteErr     error
	updateParams      []nodeDb.UpdateNodeParams
	updateResult      nodeDb.UpdateNodeRow
	updateErr         error
}

func (f *nodeStoreFake) CreateNode(_ context.Context, params nodeDb.CreateNodeParams) (nodeDb.Node, error) {
	f.createParams = append(f.createParams, params)
	if f.createErr != nil {
		return nodeDb.Node{}, f.createErr
	}
	if f.createResult.ID.Valid {
		return f.createResult, nil
	}

	return nodeDb.Node{
		ID:        pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		ParentID:  params.ParentID,
		Type:      params.Type,
		Title:     params.Title,
		SortOrder: params.SortOrder,
	}, nil
}

func (f *nodeStoreFake) GetNodeByID(_ context.Context, params nodeDb.GetNodeByIDParams) (nodeDb.Node, error) {
	f.getNodeByIDCalls = append(f.getNodeByIDCalls, params)
	if f.getNodeByIDErr != nil {
		return nodeDb.Node{}, f.getNodeByIDErr
	}
	if f.getNodeByIDResult != nil {
		if result, ok := f.getNodeByIDResult[params.ID.String()]; ok {
			return result, nil
		}
	}
	return nodeDb.Node{}, pgx.ErrNoRows
}

func (f *nodeStoreFake) SoftDeleteNodeCascade(_ context.Context, params nodeDb.SoftDeleteNodeCascadeParams) ([]pgtype.UUID, error) {
	f.softDeleteParams = append(f.softDeleteParams, params)
	if f.softDeleteErr != nil {
		return nil, f.softDeleteErr
	}
	return f.softDeleteResult, nil
}

func (f *nodeStoreFake) UpdateNode(_ context.Context, params nodeDb.UpdateNodeParams) (nodeDb.UpdateNodeRow, error) {
	f.updateParams = append(f.updateParams, params)
	if f.updateErr != nil {
		return nodeDb.UpdateNodeRow{}, f.updateErr
	}
	return f.updateResult, nil
}
