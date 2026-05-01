package handlers

import (
	"database/sql"
	"errors"
	"net/http"

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

// GetProfile   godoc
// @Summary     Get the current user's profile
// @Description Returns the profile data of an authorized user
// @Tags        User
// @Produce     json
// @Success     200 {object} dto.GetProfileResponse
// @Failure     401 {string} string "Unauthorized"
// @Failure     404 {string} string "User not found"
// @Failure     500 {string} string "Internal Server Error"
// @Router      /profile/me [get]
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

	response, err := h.Service.GetUserById(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates the name and profile picture of the logged-in user
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body dto.UpdateUserProfileRequest true "Information to update profile"
// @Success      200 {object} dto.UpdateUserProfileResponse
// @Failure      400 {string} string "Bad Request"
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /profile/me [patch]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.UpdateUserProfileRequest](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	response, err := h.Service.UpdateUserProfile(r.Context(), userID, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}

// UpdatePreferences godoc
// @Summary          Update user preferences
// @Description      Updates the locale, time zone, and user preferences
// @Tags             User
// @Accept           json
// @Produce          json
// @Param            request body dto.UpdateUserPreferencesRequest true "New user preferences"
// @Success          200 {object} dto.UpdateUserPreferencesResponse
// @Failure          400 {string} string "Bad Request"
// @Failure          401 {string} string "Unauthorized"
// @Failure          500 {string} string "Internal Server Error"
// @Router           /profile/me/preference [patch]
func (h *UserHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.UpdateUserPreferencesRequest](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	response, err := h.Service.UpdateUserPreferences(r.Context(), userID, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}

// ChangePassword godoc
// @Summary      Change user password
// @Description  Updates authenticated user's password. Requires old password for verification.
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body dto.ChangePasswordRequest true "Old and new passwords"
// @Success      200 {object} map[string]string "Example: {\"message\": \"password updated\"}"
// @Failure      400 {object} string "Bad Request"
// @Failure      401 {object} string "Unauthorized (Missing or invalid token, or wrong old password)"
// @Failure      500 {object} string "Internal Server Error"
// @Router       /profile/me/change-password [patch]
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.ChangePasswordRequest](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	if err := h.Service.UpdateUserPassword(r.Context(), userID, body); err != nil {
		if errors.Is(err, service.ErrWrongCredentials) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		http.Error(w, service.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, map[string]string{"message": "password updated"}, http.StatusOK)
}
