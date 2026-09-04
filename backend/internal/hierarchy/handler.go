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
// @Tags         Nodes
// @Produce      json
// @Param        id path string true "Node ID (UUID)"
// @Success      200 {object} GetChildrenResponse
// @Failure      400 {object} dto.ErrorResponse "invalid node id format or parent not found"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{id} [get]
func (h *Handler) GetChildren(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")

	parsedNodeID, err := httputil.PgUUIDFromString(&nodeID)
	if err != nil {
		httputil.WriteErrorJSON(w, "invalid node id format", http.StatusBadRequest)
		return
	}

	userID, err := httputil.GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response, err := h.service.GetChildren(r.Context(), parsedNodeID, userID)
	if err != nil {
		if errors.Is(err, ErrParentNotFound) {
			httputil.WriteErrorJSON(w, "parent not found", http.StatusBadRequest)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}

func (h *Handler) GetParent(w http.ResponseWriter, r *http.Request) {}
