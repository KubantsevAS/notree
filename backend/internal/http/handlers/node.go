package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/KubantsevAS/notree/backend/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type NodeHandler struct {
	service *service.NodeService
}

func NewNodeHandler(s *service.NodeService) *NodeHandler {
	return &NodeHandler{service: s}
}

func (h *NodeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var reqDTO dto.CreateNodeRequest
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// TODO user_id from JWT token
	userID := pgtype.UUID{
		Bytes: uuid.MustParse("11111111-2222-3333-4444-555555555555"), // TODO Mock data used
		Valid: true,
	}

	dbNode, err := h.service.CreateNode(r.Context(), userID, &reqDTO)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}

	responseDTO := dto.CreateNodeResponse{
		ID:        dbNode.ID,
		ParentID:  &dbNode.ParentID,
		Type:      string(dbNode.Type),
		Title:     dbNode.Title,
		SortOrder: dbNode.SortOrder,
		CreatedAt: dbNode.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responseDTO)
}
