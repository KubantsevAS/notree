package router

import (
	"log/slog"

	"github.com/KubantsevAS/notree/backend/internal/auth"
	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy"
	mwAuth "github.com/KubantsevAS/notree/backend/internal/http/middleware/auth"
	mwLogger "github.com/KubantsevAS/notree/backend/internal/http/middleware/logger"
	"github.com/KubantsevAS/notree/backend/internal/node"
	"github.com/KubantsevAS/notree/backend/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(
	cfg *config.Config,
	log *slog.Logger,
	authModule *auth.Module,
	userModule *user.Module,
	nodeModule *node.Module,
	hierarchyModule *hierarchy.Module,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(mwLogger.New(log))
	r.Use(middleware.URLFormat)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1", func(r chi.Router) {
		authModule.RegisterRoutes(r)

		r.Group(func(r chi.Router) {
			r.Use(mwAuth.AuthMiddleware(cfg.JWT.Secret))

			userModule.RegisterRoutes(r)
			r.Route("/nodes", func(r chi.Router) {
				nodeModule.RegisterRoutes(r)
				hierarchyModule.RegisterRoutes(r)
			})
		})
	})

	return r
}
