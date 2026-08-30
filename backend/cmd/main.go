// @title           Notree API
// @version         0.2
// @description     API server for Notree app.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"log/slog"
	"net/http"

	_ "github.com/KubantsevAS/notree/backend/docs"
	"github.com/KubantsevAS/notree/backend/internal/auth"
	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/KubantsevAS/notree/backend/internal/db"
	sqlcAuth "github.com/KubantsevAS/notree/backend/internal/db/auth"
	sqlcHierarchy "github.com/KubantsevAS/notree/backend/internal/db/hierarchy"
	sqlcNode "github.com/KubantsevAS/notree/backend/internal/db/node"
	sqlcUser "github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/hierarchy"
	"github.com/KubantsevAS/notree/backend/internal/mailer"
	"github.com/KubantsevAS/notree/backend/internal/node"
	"github.com/KubantsevAS/notree/backend/internal/router"
	"github.com/KubantsevAS/notree/backend/internal/user"
	"github.com/KubantsevAS/notree/backend/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)
	log.Info("Starting Notree backend v0.2.0", slog.String("env", cfg.Env))

	dbpool := db.CreateDbPool(&cfg.DB, log)
	defer dbpool.Close()

	authDB := sqlcAuth.New(dbpool)
	nodesDB := sqlcNode.New(dbpool)
	usersDB := sqlcUser.New(dbpool)
	hierarchyDB := sqlcHierarchy.New(dbpool)

	mailerService := mailer.NewConsoleMailer()

	authModule := auth.NewModule(cfg, authDB, usersDB, mailerService)
	userModule := user.NewModule(usersDB, mailerService)
	nodeModule := node.NewModule(nodesDB)
	hierarchyModule := hierarchy.NewModule(hierarchyDB, nodesDB)

	router := router.NewRouter(cfg, log, authModule, userModule, nodeModule, hierarchyModule)

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
