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

func (h *NodeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	nodeId := chi.URLParam(r, "id")
	if nodeId == "" {
		httputil.WriteErrorJSON(w, "node id required", http.StatusBadRequest)
		return
	}

	userID, err := httputil.GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := h.service.DeleteNode(r.Context(), nodeId, userID); err != nil {
		if errors.Is(err, service.ErrNodeNotFoundOrNoAccess) {
			httputil.WriteErrorJSON(w, "node not found or no access to delete it", http.StatusBadRequest)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
