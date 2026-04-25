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

	user, err := h.Service.GetUserById(r.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.GetProfileResponse{
		ID:              user.ID.String(),
		Email:           user.Email,
		Username:        &user.Username.String,
		AvatarUrl:       &user.AvatarUrl.String,
		Timezone:        &user.Timezone.String,
		Locale:          &user.Locale.String,
		Preferences:     &user.Preferences,
		IsEmailVerified: &user.IsEmailVerified.Bool,
		LastLoginAt:     &user.LastLoginAt.Time,
		CreatedAt:       &user.CreatedAt.Time,
		UpdatedAt:       &user.UpdatedAt.Time,
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates the name and profile picture of the logged-in user
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body dto.UpdateProfileRequest true "Information to update profile"
// @Success      200 {object} dto.UpdateUserProfileResponse
// @Failure      400 {string} string "Bad Request"
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /profile/me [patch]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.UpdateProfileRequest](r)
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

	userProfile, err := h.Service.UpdateUserProfile(r.Context(), user.UpdateUserProfileParams{
		Username:  httputil.PgTextFromString(body.Username),
		AvatarUrl: httputil.PgTextFromString(body.AvatarUrl),
		ID:        userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.UpdateUserProfileResponse{
		Username:  &userProfile.Username.String,
		AvatarUrl: &userProfile.AvatarUrl.String,
		UpdatedAt: &userProfile.UpdatedAt.Time,
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

	userPreferences, err := h.Service.UpdateUserPreferences(r.Context(), user.UpdateUserPreferencesParams{
		Locale:      httputil.PgTextFromString(body.Locale),
		Timezone:    httputil.PgTextFromString(body.Timezone),
		Preferences: httputil.RawMsgFromPtr(body.Preferences),
		ID:          userID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := dto.UpdateUserPreferencesResponse{
		Locale:      &userPreferences.Locale.String,
		Timezone:    &userPreferences.Timezone.String,
		Preferences: &userPreferences.Preferences,
		UpdatedAt:   &userPreferences.UpdatedAt.Time,
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}
