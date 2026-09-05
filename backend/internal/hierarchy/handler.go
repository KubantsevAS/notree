package hierarchy

import (
	"errors"
	"net/http"

	"github.com/KubantsevAS/notree/backend/internal/http/httputil"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// GetChildren godoc
// @Summary      Get child nodes
// @Description  Retrieves a list of direct child nodes for a specific parent node.
// @Tags         Hierarchy
// @Produce      json
// @Param        id path string true "Node ID (UUID)"
// @Success      200 {array} NodeResponse
// @Failure      400 {object} dto.ErrorResponse "invalid node id format"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      404 {object} dto.ErrorResponse "node not found"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{id}/children [get]
func (h *Handler) GetChildren(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")

	userID, err := httputil.GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusUnauthorized)
		return
	}

	parsedNodeID, err := httputil.PgUUIDFromString(&nodeID)
	if err != nil {
		httputil.WriteErrorJSON(w, "invalid node id format", http.StatusBadRequest)
		return
	}

	response, err := h.service.GetChildren(r.Context(), parsedNodeID, userID)
	if err != nil {
		if errors.Is(err, ErrNodeNotFound) {
			httputil.WriteErrorJSON(w, "node not found", http.StatusNotFound)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}

// GetChildren godoc
// @Summary      Get parent node
// @Description  Retrieves parent for a specific node.
// @Tags         Hierarchy
// @Produce      json
// @Param        id path string true "Node ID (UUID)"
// @Success      200 {object} NodeResponse
// @Success      204 "node is root and has no parent"
// @Failure      400 {object} dto.ErrorResponse "invalid node id format"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      404 {object} dto.ErrorResponse "node or parent not found"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{id}/parent [get]
func (h *Handler) GetParent(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")

	userID, err := httputil.GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusUnauthorized)
		return
	}

	parsedNodeID, err := httputil.PgUUIDFromString(&nodeID)
	if err != nil {
		httputil.WriteErrorJSON(w, "invalid node id format", http.StatusBadRequest)
		return
	}

	response, err := h.service.GetParent(r.Context(), parsedNodeID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNodeNotFound):
			httputil.WriteErrorJSON(w, "node not found", http.StatusNotFound)
		case errors.Is(err, ErrNodeIsRoot):
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, ErrParentNotFound):
			httputil.WriteErrorJSON(w, "parent not found", http.StatusNotFound)
		default:
			httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}
