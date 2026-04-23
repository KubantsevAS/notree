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

func (s *UserService) UpdateUserProfile(ctx context.Context, arg user.UpdateUserProfileParams) (user.UpdateUserProfileRow, error) {
	return s.db.UpdateUserProfile(ctx, arg)
}

func (s *UserService) UpdateUserPreferences(ctx context.Context, arg user.UpdateUserPreferencesParams) (user.UpdateUserPreferencesRow, error) {
	return s.db.UpdateUserPreferences(ctx, arg)
}
