package user

import (
	"github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/mailer"
	"github.com/go-chi/chi/v5"
)

type Module struct {
	handler *Handler
}

func NewModule(
	queries *user.Queries,
	mailer mailer.Mailer,
) *Module {
	service := NewService(queries, mailer)
	handler := NewHandler(service)

	return &Module{
		handler: handler,
	}
}

func (m *Module) RegisterRoutes(r chi.Router) {
	r.Route("/profile", func(r chi.Router) {
		r.Get("/me", m.handler.GetProfile)
		r.Patch("/me", m.handler.UpdateProfile)
		r.Patch("/me/preference", m.handler.UpdatePreferences)
		r.Patch("/me/change-password", m.handler.ChangePassword)
		r.Post("/me/send-verification", m.handler.SendVerificationToken)
		r.Post("/me/verify-email", m.handler.VerifyEmailByToken)
	})
}
