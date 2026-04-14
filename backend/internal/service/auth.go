package service

import (
	"context"

	"github.com/KubantsevAS/notree/backend/internal/db/sqlc"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db *sqlc.Queries
}

func NewAuthService(db *sqlc.Queries) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.User, error) {
	_, err := s.db.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrUserExist
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.db.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		Username:     pgtype.Text{String: req.Username, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &dto.User{
		ID:       user.ID.String(),
		Email:    user.Email,
		Username: user.Username.String,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.User, error) {
	user, err := s.db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrWrongCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, ErrWrongCredentials
	}

	return &dto.User{
		ID:       user.ID.String(),
		Email:    user.Email,
		Username: user.Username.String,
	}, nil
}
