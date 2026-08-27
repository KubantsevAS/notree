package handlers

import (
	"errors"
	"net/http"

	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/KubantsevAS/notree/backend/internal/httputil"
	"github.com/KubantsevAS/notree/backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type NodeHandler struct {
	service *service.NodeService
}

func NewNodeHandler(s *service.NodeService) *NodeHandler {
	return &NodeHandler{service: s}
}

// Create godoc
// @Summary      Create a new node
// @Tags         Nodes
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateNodeRequest true "Information to create node"
// @Success      201 {object} dto.CreateNodeResponse
// @Failure      400 {object} dto.ErrorResponse "bad request"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes [post]
func (h *NodeHandler) Create(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.CreateNodeRequest](r)
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
		if errors.Is(err, service.ErrInvalidParentID) {
			httputil.WriteErrorJSON(w, "invalid parent id", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrParentNotFound) {
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
func (h *NodeHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, service.ErrNodeNotFoundOrNoAccess) {
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
// @Param        request body dto.UpdateNodeRequest true "Fields to update"
// @Success      200 {object} dto.UpdateNodeResponse
// @Failure      400 {object} dto.ErrorResponse "invalid request body or node ID format"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      404 {object} dto.ErrorResponse "node not found or access denied"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{id} [patch]
func (h *NodeHandler) Update(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.UpdateNodeRequest](r)
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
		if errors.Is(err, service.ErrEmptyUpdate) {
			httputil.WriteErrorJSON(w, "no fields provided for update", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrNodeNotFoundOrNoAccess) {
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
// @Param        request body dto.MoveNodeRequest true "Move parameters"
// @Success      200 {object} dto.MoveNodeResponse
// @Failure      400 {object} dto.ErrorResponse "invalid parent ID format or parent not found"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      404 {object} dto.ErrorResponse "node not found or access denied"
// @Failure      409 {object} dto.ErrorResponse "node cannot be a descendant of itself (circular reference)"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{id}/move [post]
func (h *NodeHandler) Move(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.MoveNodeRequest](r)
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
		if errors.Is(err, service.ErrEmptyUpdate) {
			httputil.WriteErrorJSON(w, "no fields provided for update", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrInvalidParentID) {
			httputil.WriteErrorJSON(w, "invalid parent id", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrNodeCannotBeADescendantOfItself) {
			httputil.WriteErrorJSON(w, "node cannot be a descendant of itself (circular reference)", http.StatusConflict)
			return
		}
		if errors.Is(err, service.ErrParentNotFound) {
			httputil.WriteErrorJSON(w, "parent not found", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrNodeNotFoundOrNoAccess) {
			httputil.WriteErrorJSON(w, "node not found or access denied", http.StatusNotFound)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}

// GetChildren godoc
// @Summary      Get child nodes
// @Description  Retrieves a list of direct child nodes for a specific parent node.
// @Tags         Nodes
// @Produce      json
// @Param        parent_id path string true "Parent Node ID (UUID)"
// @Success      200 {object} dto.GetChildrenResponse
// @Failure      400 {object} dto.ErrorResponse "invalid node id format or parent not found"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{parent_id} [get]
func (h *NodeHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	parentID := chi.URLParam(r, "parent_id")

	parsedNodeID, err := httputil.PgUUIDFromString(&parentID)
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
		if errors.Is(err, service.ErrParentNotFound) {
			httputil.WriteErrorJSON(w, "parent not found", http.StatusBadRequest)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}
