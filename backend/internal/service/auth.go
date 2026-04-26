package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/KubantsevAS/notree/backend/internal/db/auth"
	"github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/KubantsevAS/notree/backend/pkg/jwt"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	config *config.Config
	db     *auth.Queries
	userDb *user.Queries
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewAuthService(c *config.Config, authDb *auth.Queries, userDb *user.Queries) *AuthService {
	return &AuthService{
		config: c,
		db:     authDb,
		userDb: userDb,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*TokenPair, error) {
	_, err := s.userDb.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrUserExist
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID, err := s.userDb.CreateUser(ctx, user.CreateUserParams{
		Email:        req.Email,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		return nil, err
	}

	return s.generateTokenPair(ctx, userID)
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*TokenPair, error) {
	user, err := s.userDb.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrWrongCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, ErrWrongCredentials
	}

	return s.generateTokenPair(ctx, user.ID)
}

func (s *AuthService) Logout(ctx context.Context, incomingRT string) error {
	return s.db.DeleteRefreshToken(ctx, incomingRT)
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

	err = s.db.CreateRefreshToken(ctx, auth.CreateRefreshTokenParams{
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
