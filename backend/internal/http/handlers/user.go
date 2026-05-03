package handlers

import (
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
// @Tags        Profile
// @Produce     json
// @Success     200 {object} dto.GetProfileResponse
// @Failure     401 {object} dto.ErrorResponse "unauthorized"
// @Failure     404 {object} dto.ErrorResponse "user not found"
// @Failure     500 {object} dto.ErrorResponse "internal server error"
// @Router      /profile/me [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := httputil.GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response, err := h.Service.GetUserById(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			httputil.WriteErrorJSON(w, err.Error(), http.StatusNotFound)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Updates the name and profile picture of the logged-in user
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Param        request body dto.UpdateUserProfileRequest true "Information to update profile"
// @Success      200 {object} dto.UpdateUserProfileResponse
// @Failure      400 {object} dto.ErrorResponse "bad request"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /profile/me [patch]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.UpdateUserProfileRequest](r)
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := httputil.GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response, err := h.Service.UpdateUserProfile(r.Context(), userID, body)
	if err != nil {
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}

// UpdatePreferences godoc
// @Summary          Update user preferences
// @Description      Updates the locale, time zone, and user preferences
// @Tags             Profile
// @Accept           json
// @Produce          json
// @Param            request body dto.UpdateUserPreferencesRequest true "New user preferences"
// @Success          200 {object} dto.UpdateUserPreferencesResponse
// @Failure          400 {object} dto.ErrorResponse "bad request"
// @Failure          401 {object} dto.ErrorResponse "unauthorized"
// @Failure          500 {object} dto.ErrorResponse "internal server error"
// @Router           /profile/me/preference [patch]
func (h *UserHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.UpdateUserPreferencesRequest](r)
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := httputil.GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusUnauthorized)
		return
	}

	response, err := h.Service.UpdateUserPreferences(r.Context(), userID, body)
	if err != nil {
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, response, http.StatusOK)
}

// ChangePassword godoc
// @Summary      Change user password
// @Description  Updates authenticated user's password. Requires old password for verification.
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Param        request body dto.ChangePasswordRequest true "Old and new passwords"
// @Success      200 {object} dto.MessageResponse "password updated"
// @Failure      400 {object} dto.ErrorResponse "bad request"
// @Failure      401 {object} dto.ErrorResponse "wrong old password"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /profile/me/change-password [patch]
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.ChangePasswordRequest](r)
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := httputil.GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := h.Service.UpdateUserPassword(r.Context(), userID, body); err != nil {
		if errors.Is(err, service.ErrWrongCredentials) {
			httputil.WriteErrorJSON(w, "wrong old password", http.StatusUnauthorized)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, dto.MessageResponse{Message: "password updated"}, http.StatusOK)
}

// SendVerificationToken godoc
// @Summary      Send email verification token
// @Tags         Profile
// @Produce      json
// @Success      200 {object} dto.MessageResponse "email verification link has been sent"
// @Failure      400 {object} dto.ErrorResponse "bad request"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /profile/me/send-verification [post]
func (h *UserHandler) SendVerificationToken(w http.ResponseWriter, r *http.Request) {
	userID, err := httputil.GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// TODO Implement 429 code

	if err := h.Service.SendVerificationEmail(r.Context(), userID); err != nil {
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, dto.MessageResponse{Message: "email verification link has been sent"}, http.StatusOK)
}

// VerifyEmailByToken godoc
// @Summary      Verify email with token
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Param        request body dto.VerifyEmailByTokenRequest true "token"
// @Success      200 {object} dto.MessageResponse "email successfully verified"
// @Failure      400 {object} dto.ErrorResponse "invalid or expired token"
// @Failure      401 {object} dto.ErrorResponse "unauthorized"
// @Failure      500 {object} dto.ErrorResponse "internal server error"
// @Router       /profile/me/verify-email [post]
func (h *UserHandler) VerifyEmailByToken(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.VerifyEmailByTokenRequest](r)
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, err := httputil.GetUserPgUUIDFromCtx(r.Context())
	if err != nil {
		httputil.WriteErrorJSON(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := h.Service.VerifyEmailByToken(r.Context(), userID, body.Token); err != nil {
		if errors.Is(err, service.ErrInvalidVerificationToken) {
			httputil.WriteErrorJSON(w, "invalid or expired token", http.StatusBadRequest)
			return
		}
		httputil.WriteErrorJSON(w, "internal server error", http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, dto.MessageResponse{Message: "email successfully verified"}, http.StatusOK)
}
