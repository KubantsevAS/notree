package node

import (
	"errors"
	"net/http"

	"github.com/KubantsevAS/notree/backend/internal/domain"
	"github.com/KubantsevAS/notree/backend/internal/http/httputil"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// Create godoc
// @Summary      Create a new node
// @Tags         Nodes
// @Accept       json
// @Produce      json
// @Param        request body CreateNodeRequest true "Information to create node"
// @Success      201 {object} CreateNodeResponse
// @Failure      400 {object} dto.ErrorResponse "bad request"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[CreateNodeRequest](r)
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := httputil.GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusUnauthorized)
		return
	}

	node, err := h.service.CreateNode(r.Context(), userID, body)
	if err != nil {
		if errors.Is(err, ErrInvalidParentID) {
			httputil.WriteErrorJSON(w, "invalid parent id", http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrParentNotFound) {
			httputil.WriteErrorJSON(w, "parent not found", http.StatusBadRequest)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, node, http.StatusCreated)
}

// Delete godoc
// @Summary      Delete a node
// @Description  Soft deletes a specific node by ID and all its nested children recursively. The nodes are marked as deleted and can be restored later.
// @Tags         Nodes
// @Produce      json
// @Param        id path string true "Node ID (UUID)"
// @Success      204 {object} nil "No Content"
// @Failure      400 {object} dto.ErrorResponse "invalid node id format"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      404 {object} dto.ErrorResponse "node not found or access denied"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.DeleteNode(r.Context(), parsedNodeID, userID); err != nil {
		if errors.Is(err, ErrNodeNotFoundOrNoAccess) {
			httputil.WriteErrorJSON(w, "node not found or access denied", http.StatusNotFound)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Update godoc
// @Summary      Update node metadata
// @Description  Partially updates a specific node (e.g., title or type) by ID. Does not affect hierarchy.
// @Tags         Nodes
// @Accept       json
// @Produce      json
// @Param        id path string true "Node ID (UUID)"
// @Param        request body UpdateNodeRequest true "Fields to update"
// @Success      200 {object} UpdateNodeResponse
// @Failure      400 {object} dto.ErrorResponse "invalid request body or node ID format"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      404 {object} dto.ErrorResponse "node not found or access denied"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{id} [patch]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[UpdateNodeRequest](r)
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	response, err := h.service.UpdateNode(r.Context(), parsedNodeID, userID, body)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyUpdate) {
			httputil.WriteErrorJSON(w, "no fields provided for update", http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrNodeNotFoundOrNoAccess) {
			httputil.WriteErrorJSON(w, "node not found or access denied", http.StatusNotFound)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}

// Move godoc
// @Summary      Move node in tree
// @Description  Changes the parent or sort order of a node. Pass `null` as parent_id to move the node to the root.
// @Tags         Nodes
// @Accept       json
// @Produce      json
// @Param        id path string true "Node ID (UUID)"
// @Param        request body MoveNodeRequest true "Move parameters"
// @Success      200 {object} MoveNodeResponse
// @Failure      400 {object} dto.ErrorResponse "invalid parent ID format or parent not found"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      404 {object} dto.ErrorResponse "node not found or access denied"
// @Failure      409 {object} dto.ErrorResponse "node cannot be a descendant of itself (circular reference)"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{id}/move [post]
func (h *Handler) Move(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[MoveNodeRequest](r)
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

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

	response, err := h.service.MoveNode(r.Context(), parsedNodeID, userID, body)
	if err != nil {
		if errors.Is(err, domain.ErrEmptyUpdate) {
			httputil.WriteErrorJSON(w, "no fields provided for update", http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrInvalidParentID) {
			httputil.WriteErrorJSON(w, "invalid parent id", http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrNodeCannotBeADescendantOfItself) {
			httputil.WriteErrorJSON(w, "node cannot be a descendant of itself (circular reference)", http.StatusConflict)
			return
		}
		if errors.Is(err, ErrParentNotFound) {
			httputil.WriteErrorJSON(w, "parent not found", http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrNodeNotFoundOrNoAccess) {
			httputil.WriteErrorJSON(w, "node not found or access denied", http.StatusNotFound)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}
