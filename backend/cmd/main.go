package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/KubantsevAS/notree/backend/internal/db"
	sqlc "github.com/KubantsevAS/notree/backend/internal/db/sqlc"
	"github.com/KubantsevAS/notree/backend/internal/http/handlers"
	mwAuth "github.com/KubantsevAS/notree/backend/internal/http/middleware/auth"
	mwLogger "github.com/KubantsevAS/notree/backend/internal/http/middleware/logger"
	"github.com/KubantsevAS/notree/backend/internal/service"
	"github.com/KubantsevAS/notree/backend/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg.JWT.Secret)

	log := logger.SetupLogger(cfg.Env)
	log.Info("Starting Notree backend", slog.String("env", cfg.Env))

	dbpool := db.CreateDbPool(&cfg.DB, log)
	defer dbpool.Close()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwLogger.New(log))
	router.Use(mwAuth.AuthMiddleware(cfg.JWT.Secret))
	router.Use(middleware.URLFormat)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	queries := sqlc.New(dbpool)
	nodeService := service.NewNodeService(queries)
	authService := service.NewAuthService(cfg, queries)

	authHandler := handlers.NewAuthHandler(authService)

	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh-tokens", authHandler.RefreshTokens)
		r.Post("/logout", authHandler.Logout)

	})
	router.Post("/node", handlers.NewNodeHandler(nodeService).Create)

	server := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("Failed to start server")
	}

	log.Error("Server stopped")
}
