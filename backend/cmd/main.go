// @title           Notree API
// @version         0.1
// @description     API server for Notree app.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /docs

// @securityDefinitions.basic  BasicAuth

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/KubantsevAS/notree/backend/docs"
	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/KubantsevAS/notree/backend/internal/db"
	sqlcAuth "github.com/KubantsevAS/notree/backend/internal/db/auth"
	sqlcNode "github.com/KubantsevAS/notree/backend/internal/db/node"
	sqlcUser "github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/http/handlers"
	mwAuth "github.com/KubantsevAS/notree/backend/internal/http/middleware/auth"
	mwLogger "github.com/KubantsevAS/notree/backend/internal/http/middleware/logger"
	"github.com/KubantsevAS/notree/backend/internal/service"
	"github.com/KubantsevAS/notree/backend/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println(cfg.JWT.Secret)

	log := logger.SetupLogger(cfg.Env)
	log.Info("Starting Notree backend", slog.String("env", cfg.Env))

	dbpool := db.CreateDbPool(&cfg.DB, log)
	defer dbpool.Close()

	router := chi.NewRouter()

	docs.SwaggerInfo.BasePath = "/"

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwLogger.New(log))
	router.Use(middleware.URLFormat)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	authDB := sqlcAuth.New(dbpool)
	nodesDB := sqlcNode.New(dbpool)
	usersDB := sqlcUser.New(dbpool)

	nodeService := service.NewNodeService(nodesDB)
	authService := service.NewAuthService(cfg, authDB, usersDB)

	authHandler := handlers.NewAuthHandler(authService)

	router.Get("/swagger/*", httpSwagger.WrapHandler)

	router.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh-tokens", authHandler.RefreshTokens)
		r.Post("/logout", authHandler.Logout)
	})
	router.Group(func(r chi.Router) {
		r.Use(mwAuth.AuthMiddleware(cfg.JWT.Secret))
		r.Post("/node", handlers.NewNodeHandler(nodeService).Create)
	})

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
