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

// GetParent godoc
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

// GetAncestors godoc
// @Summary      Get ancestor nodes
// @Description  Retrieves all ancestor nodes for a specific node ordered from root to direct parent.
// @Tags         Hierarchy
// @Produce      json
// @Param        id path string true "Node ID (UUID)"
// @Success      200 {array} NodeResponse
// @Failure      400 {object} dto.ErrorResponse "invalid node id format"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      404 {object} dto.ErrorResponse "node not found"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{id}/ancestors [get]
func (h *Handler) GetAncestors(w http.ResponseWriter, r *http.Request) {
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

	response, err := h.service.GetAncestors(r.Context(), parsedNodeID, userID)
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

// GetDescendants godoc
// @Summary      Get descendant nodes
// @Description  Retrieves all nested descendant nodes for a specific node in depth-first order.
// @Tags         Hierarchy
// @Produce      json
// @Param        id path string true "Node ID (UUID)"
// @Success      200 {array} NodeResponse
// @Failure      400 {object} dto.ErrorResponse "invalid node id format"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      404 {object} dto.ErrorResponse "node not found"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{id}/descendants [get]
func (h *Handler) GetDescendants(w http.ResponseWriter, r *http.Request) {
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

	response, err := h.service.GetDescendants(r.Context(), parsedNodeID, userID)
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

// GetSubtree godoc
// @Summary      Get subtree
// @Description  Retrieves a specific node together with all its nested descendants in depth-first order.
// @Tags         Hierarchy
// @Produce      json
// @Param        id path string true "Node ID (UUID)"
// @Success      200 {array} NodeResponse
// @Failure      400 {object} dto.ErrorResponse "invalid node id format"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      404 {object} dto.ErrorResponse "node not found"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{id}/subtree [get]
func (h *Handler) GetSubtree(w http.ResponseWriter, r *http.Request) {
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

	response, err := h.service.GetSubtree(r.Context(), parsedNodeID, userID)
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

// GetRoot godoc
// @Summary      Get root node
// @Description  Retrieves the top-level root node of the tree containing a specific node. Returns the node itself if it is a root.
// @Tags         Hierarchy
// @Produce      json
// @Param        id path string true "Node ID (UUID)"
// @Success      200 {object} NodeResponse
// @Failure      400 {object} dto.ErrorResponse "invalid node id format"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      404 {object} dto.ErrorResponse "node or root not found"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /nodes/{id}/root [get]
func (h *Handler) GetRoot(w http.ResponseWriter, r *http.Request) {
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

	response, err := h.service.GetRoot(r.Context(), parsedNodeID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNodeNotFound):
			httputil.WriteErrorJSON(w, "node not found", http.StatusNotFound)
		case errors.Is(err, ErrRootNotFound):
			httputil.WriteErrorJSON(w, "root not found", http.StatusNotFound)
		default:
			httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}

// Move godoc
// @Summary      Move node in tree
// @Description  Places a node under parent_id right before before_id. Both fields are required: `null` parent_id means the root, `null` before_id means the end of the list. The same parent_id reorders the node among its siblings.
// @Tags         Hierarchy
// @Accept       json
// @Produce      json
// @Param        id path string true "Node ID (UUID)"
// @Param        request body MoveNodeRequest true "Move parameters"
// @Success      200 {object} MoveNodeResponse
// @Failure      400 {object} dto.ErrorResponse "parent_id or before_id missing, invalid or not found"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      404 {object} dto.ErrorResponse "node not found"
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

	response, err := h.service.MoveNode(r.Context(), parsedNodeID, userID, body)
	if err != nil {
		switch {
		case errors.Is(err, ErrParentIDRequired):
			httputil.WriteErrorJSON(w, "parent_id is required", http.StatusBadRequest)
		case errors.Is(err, ErrBeforeIDRequired):
			httputil.WriteErrorJSON(w, "before_id is required", http.StatusBadRequest)
		case errors.Is(err, ErrInvalidBeforeID):
			httputil.WriteErrorJSON(w, "invalid before id", http.StatusBadRequest)
		case errors.Is(err, ErrBeforeNotSibling):
			httputil.WriteErrorJSON(w, "before_id must reference another child of parent_id", http.StatusBadRequest)
		case errors.Is(err, ErrInvalidParentID):
			httputil.WriteErrorJSON(w, "invalid parent id", http.StatusBadRequest)
		case errors.Is(err, ErrParentNotFound):
			httputil.WriteErrorJSON(w, "parent not found", http.StatusBadRequest)
		case errors.Is(err, ErrNodeCannotBeADescendantOfItself):
			httputil.WriteErrorJSON(w, "node cannot be a descendant of itself (circular reference)", http.StatusConflict)
		case errors.Is(err, ErrNodeNotFound):
			httputil.WriteErrorJSON(w, "node not found", http.StatusNotFound)
		default:
			httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}
