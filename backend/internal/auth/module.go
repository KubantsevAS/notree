package auth

import (
	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/KubantsevAS/notree/backend/internal/db/auth"
	"github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/mailer"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	handler *Handler
}

func NewModule(
	config *config.Config,
	queries *auth.Queries,
	userDb *user.Queries,
	mailer mailer.Mailer,
) *Module {
	service := NewService(
		config,
		queries,
		userDb,
		mailer,
	)
	handler := NewHandler(service)

	return &Module{
		handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", m.handler.Register)
		r.Post("/login", m.handler.Login)
		r.Post("/refresh-tokens", m.handler.RefreshTokens)
		r.Post("/logout", m.handler.Logout)
		r.Post("/forgot-password", m.handler.ForgotPassword)
		r.Post("/reset-password", m.handler.ResetPassword)
	})
}
