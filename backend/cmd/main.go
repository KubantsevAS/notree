package main

import (
	"log/slog"

	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/KubantsevAS/notree/backend/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	log.Info("Starting Notree backend", slog.String("env", cfg.Env))
}
