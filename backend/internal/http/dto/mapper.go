package dto

import "github.com/KubantsevAS/notree/backend/internal/db/node"

func MapNodeToResponse(n node.Node) NodeResponse {
	res := NodeResponse{
		ID:        n.ID.String(),
		UserID:    n.UserID.String(),
		ParentID:  n.ParentID.String(),
		Type:      string(n.Type),
		Title:     n.Title,
		SortOrder: n.SortOrder,
		UpdatedAt: &n.UpdatedAt.Time,
		CreatedAt: &n.CreatedAt.Time,
		DeletedAt: &n.DeletedAt.Time,
	}

	return res
}
