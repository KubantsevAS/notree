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
	queries *hierarchy.Queries,
	nodeDb *node.Queries,
) *Module {
	service := NewService(queries, nodeDb)
	handler := NewHandler(service)

	return &Module{
		handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) {}
