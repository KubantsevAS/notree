package handlers

import (
	"net/http"

	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/KubantsevAS/notree/backend/internal/httputil"
	"github.com/KubantsevAS/notree/backend/internal/service"
)

type AuthHandler struct {
	Config  *config.Config
	Service *service.AuthService
}

type authHTTPResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

func NewAuthHandler(c *config.Config, s *service.AuthService) *AuthHandler {
	return &AuthHandler{
		Config:  c,
		Service: s,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.RegisterRequest](&w, r)
	if err != nil {
		return
	}

	user, err := h.Service.Register(r.Context(), body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resDTO := authHTTPResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Token:    "",
	}

	httputil.WriteResponseJSON(w, resDTO, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	body, err := httputil.HandleBody[dto.LoginRequest](&w, r)
	if err != nil {
		return
	}

	user, err := h.Service.Login(r.Context(), body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resDTO := authHTTPResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Token:    "",
	}

	httputil.WriteResponseJSON(w, resDTO, http.StatusOK)
}
