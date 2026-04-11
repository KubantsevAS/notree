package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/KubantsevAS/notree/backend/internal/db"
	"github.com/KubantsevAS/notree/backend/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)
	log.Info("Starting Notree backend", slog.String("env", cfg.Env))

	dbpool := db.CreateDbPool(&cfg.DB, log)
	defer dbpool.Close()

	port := os.Getenv("SERVER_PORT")
	address := fmt.Sprintf(":%s", port)
	server := http.Server{
		Addr: address,
	}

	log.Info("Server is listening", slog.String("port", address))
	server.ListenAndServe()
}
