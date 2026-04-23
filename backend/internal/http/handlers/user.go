package handlers

import (
	"net/http"

	"github.com/KubantsevAS/notree/backend/internal/httputil"
	"github.com/KubantsevAS/notree/backend/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserHandler struct {
	Service *service.UserService
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{
		Service: s,
	}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDContext, err := httputil.GetUserIDFromCtx(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID := pgtype.UUID{
		Bytes: uuid.MustParse(userIDContext),
		Valid: true,
	}

	user, err := h.Service.GetUserById(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	httputil.WriteResponseJSON(w, user, http.StatusOK)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {}

func (h *UserHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {}
