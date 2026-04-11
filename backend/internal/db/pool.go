package db

import (
	"context"
	"log/slog"

	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateDbPool(config *config.DBConfig, log *slog.Logger) *pgxpool.Pool {
	dbpool, err := pgxpool.New(context.Background(), config.DSN())
	if err != nil {
		log.Error("Database connection failed")
		panic(err)
	}
	log.Info("Database connected")

	return dbpool
}
