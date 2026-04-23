package service

import (
	"context"

	"github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserService struct {
	db *user.Queries
}

func NewUserService(db *user.Queries) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetUserById(ctx context.Context, id pgtype.UUID) (user.UsersPublic, error) {
	return s.db.GetUserById(ctx, id)
}
