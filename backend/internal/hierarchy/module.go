package hierarchy

import (
	"github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	"github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	handler *Handler
}

func NewModule(
	store *hierarchy.Queries,
	nodeStore *node.Queries,
) *Module {
	service := NewService(store, nodeStore)
	handler := NewHandler(service)

	return &Module{
		handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) {
	r.Get("/{id}/children", m.handler.GetChildren)
	r.Get("/{id}/parent", m.handler.GetParent)
	r.Get("/{id}/ancestors", m.handler.GetAncestors)
	r.Get("/{id}/descendants", m.handler.GetDescendants)
	r.Get("/{id}/subtree", m.handler.GetSubtree)
	r.Get("/{id}/root", m.handler.GetRoot)
	r.Post("/{id}/move", m.handler.Move)
}
