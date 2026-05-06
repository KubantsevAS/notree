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
	nodeId := chi.URLParam(r, "id")

	parsedNodeId, err := httputil.PgUUIDFromString(&nodeId)
	if err != nil {
		httputil.WriteErrorJSON(w, "invalid node id format", http.StatusBadRequest)
		return
	}

	userID, err := httputil.GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := h.service.DeleteNode(r.Context(), parsedNodeId, userID); err != nil {
		if errors.Is(err, service.ErrNodeNotFoundOrNoAccess) {
			httputil.WriteErrorJSON(w, "node not found or access denied", http.StatusNotFound)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
