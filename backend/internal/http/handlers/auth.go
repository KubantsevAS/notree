package handlers

import (
	"errors"
	"net/http"

	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/KubantsevAS/notree/backend/internal/httputil"
	"github.com/KubantsevAS/notree/backend/internal/service"
)

type AuthHandler struct {
	Service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{
		Service: s,
	}
}

// Register godoc
// @Summary      User registration
// @Description  Creates a new user and returns access/refresh tokens in HttpOnly cookies
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Data for registration"
// @Success      201  {string}  string "Success"
// @Failure      400  {string}  string "Incorrect request format"
// @Failure      409  {string}  string "User with that email already exists"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.RegisterRequest](&w, r)
	if err != nil {
		return
	}

	tokens, err := h.Service.Register(r.Context(), body)
	if err != nil {
		if errors.Is(err, service.ErrUserExist) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	httputil.SetCookie(w, "access_token", tokens.AccessToken, 15*60, true)
	httputil.SetCookie(w, "refresh_token", tokens.RefreshToken, 7*24*3600, true)
	w.WriteHeader(http.StatusCreated)
}

// Login godoc
// @Summary      User login
// @Description  Authenticates user and returns access/refresh tokens in HttpOnly cookies
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "User credentials"
// @Success      200  {string}  string "Success"
// @Failure      400  {string}  string "Invalid credentials"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.LoginRequest](&w, r)
	if err != nil {
		return
	}

	tokens, err := h.Service.Login(r.Context(), body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	httputil.SetCookie(w, "access_token", tokens.AccessToken, 15*60, true)
	httputil.SetCookie(w, "refresh_token", tokens.RefreshToken, 7*24*3600, true)
	w.WriteHeader(http.StatusOK)
}

// RefreshTokens godoc
// @Summary      Refresh tokens
// @Description  Takes refresh_token from cookie, deletes it, issues a new token pair and sets new cookies
// @Tags         Auth
// @Produce      json
// @Success      200  {string}  string "Tokens refreshed successfully"
// @Failure      401  {string}  string "Missing or invalid refresh token"
// @Failure      500  {string}  string "Internal server error"
// @Router       /auth/refresh-tokens [post]
func (h *AuthHandler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Missing refresh token", http.StatusUnauthorized)
		return
	}

	tokens, err := h.Service.RefreshTokens(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httputil.SetCookie(w, "access_token", tokens.AccessToken, 15*60, true)
	httputil.SetCookie(w, "refresh_token", tokens.RefreshToken, 7*24*3600, true)
	w.WriteHeader(http.StatusOK)
}

// Logout godoc
// @Summary      User logout
// @Description  Deletes refresh token from database and clears cookies
// @Tags         Auth
// @Success      204  {string}  string "No Content"
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		_ = h.Service.Logout(r.Context(), cookie.Value)
	}

	httputil.ClearCookie(w, "access_token")
	httputil.ClearCookie(w, "refresh_token")
	w.WriteHeader(http.StatusNoContent)
}
