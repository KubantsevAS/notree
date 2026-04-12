package service

import (
	"context"
	"errors"

	"github.com/KubantsevAS/notree/backend/internal/db/sqlc"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrUserExist = errors.New("User with that email already exist")
)

type AuthService struct {
	db *sqlc.Queries
}

func NewAuthService(db *sqlc.Queries) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (dto.RegisterResponse, error) {
	_, err := s.db.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return dto.RegisterResponse{}, ErrUserExist
	}

	_, err = s.db.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        req.Email,
		PasswordHash: "", // TODO add bcrypt hash
		Username:     pgtype.Text{String: req.Username, Valid: true},
	})
	if err != nil {
		return dto.RegisterResponse{}, err
	}

	return dto.RegisterResponse{
		Email:    req.Email,
		Username: req.Username,
	}, nil
}
