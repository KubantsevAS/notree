package handlers

import (
	"database/sql"
	"net/http"

	"github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/KubantsevAS/notree/backend/internal/httputil"
	"github.com/KubantsevAS/notree/backend/internal/service"
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
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := httputil.PgUUIDFromString(&userIDContext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	user, err := h.Service.GetUserById(r.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, user, http.StatusOK)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.UpdateProfileRequest](&w, r)
	if err != nil {
		return
	}

	userIDContext, err := httputil.GetUserIDFromCtx(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := httputil.PgUUIDFromString(&userIDContext)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	profileResponse, err := h.Service.UpdateUserProfile(r.Context(), user.UpdateUserProfileParams{
		Username:  httputil.PgTextFromString(body.Username),
		AvatarUrl: httputil.PgTextFromString(body.AvatarUrl),
		ID:        userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, profileResponse, http.StatusOK)
}

func (h *UserHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {}
