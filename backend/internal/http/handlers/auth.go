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

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
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

	httputil.WriteResponseJSON(w, tokens, http.StatusCreated)
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

	httputil.WriteResponseJSON(w, tokens, http.StatusOK)
}

func (h *AuthHandler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[RefreshRequest](&w, r)
	if err != nil {
		return
	}

	tokens, err := h.Service.RefreshTokens(r.Context(), body.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httputil.WriteResponseJSON(w, tokens, http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[RefreshRequest](&w, r)
	if err != nil {
		return
	}

	if err := h.Service.Logout(r.Context(), body.RefreshToken); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
