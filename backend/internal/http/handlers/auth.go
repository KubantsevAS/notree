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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.RegisterRequest](&w, r)
	if err != nil {
		return
	}

	tokens, err := h.Service.Register(r.Context(), body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	httputil.SetCookie(w, "access_token", tokens.AccessToken, 15*60, true)
	httputil.SetCookie(w, "refresh_token", tokens.RefreshToken, 7*24*3600, true)
	httputil.WriteResponseJSON(w, map[string]string{"message": "success"}, http.StatusCreated)
}

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
	httputil.WriteResponseJSON(w, map[string]string{"message": "success"}, http.StatusOK)
}

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

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		_ = h.Service.Logout(r.Context(), cookie.Value)
	}

	httputil.ClearCookie(w, "access_token")
	httputil.ClearCookie(w, "refresh_token")
	w.WriteHeader(http.StatusNoContent)
}
