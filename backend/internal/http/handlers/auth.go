package handlers

import (
	"net/http"

	"github.com/KubantsevAS/notree/backend/internal/service"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {}
