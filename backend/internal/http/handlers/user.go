package handlers

import (
	"net/http"

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

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {}

func (h *UserHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {}
