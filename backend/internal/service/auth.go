package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/KubantsevAS/notree/backend/internal/db/sqlc"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/KubantsevAS/notree/backend/pkg/jwt"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db     *sqlc.Queries
	config *config.Config
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewAuthService(c *config.Config, db *sqlc.Queries) *AuthService {
	return &AuthService{
		config: c,
		db:     db,
	}
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

func (s *AuthService) Logout(ctx context.Context, req *dto.LoginRequest) {

}

func (s *AuthService) RefreshTokens(ctx context.Context, incomingRT string) (*TokenPair, error) {
	refreshToken, err := s.db.GetRefreshToken(ctx, incomingRT)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidRefreshToken
		}

		return nil, err
	}

	if time.Now().After(refreshToken.ExpiresAt.Time) {
		s.db.DeleteRefreshToken(ctx, incomingRT)
		return nil, ErrInvalidRefreshToken
	}

	s.db.DeleteRefreshToken(ctx, incomingRT)

	return s.generateTokenPair(ctx, refreshToken.UserID)
}

func (s *AuthService) generateTokenPair(ctx context.Context, userID pgtype.UUID) (*TokenPair, error) {
	accessToken, err := jwt.GenerateAccessToken(userID, s.config.JWT.Secret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	err = s.db.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		TokenHash: refreshToken,
		UserID:    userID,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(7 * 24 * time.Hour),
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
