package node

import (
	"github.com/KubantsevAS/notree/backend/internal/db/node"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	handler *Handler
}

func NewModule(
	queries *node.Queries,
) *Module {

	service := NewService(queries)
	handler := NewHandler(service)

	return &Module{
		handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) {
	r.Route("/nodes", func(r chi.Router) {
		r.Get("/{parent_id}", m.handler.GetChildren)
		r.Post("/", m.handler.Create)
		r.Post("/{id}/move", m.handler.Move)
		r.Patch("/{id}", m.handler.Update)
		r.Delete("/{id}", m.handler.Delete)
	})
}
