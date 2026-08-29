package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/KubantsevAS/notree/backend/internal/config"
	"github.com/KubantsevAS/notree/backend/internal/db/auth"
	"github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/http/httputil"
	"github.com/KubantsevAS/notree/backend/internal/mailer"
	"github.com/KubantsevAS/notree/backend/pkg/jwt"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type Store interface {
	CreateRefreshToken(context.Context, auth.CreateRefreshTokenParams) error
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	GetRefreshToken(ctx context.Context, tokenHash string) (auth.RefreshToken, error)
}

type UserStore interface {
	GetUserByEmail(ctx context.Context, email string) (user.User, error)
	CreateUser(context.Context, user.CreateUserParams) (pgtype.UUID, error)
	UpdateUserPassword(context.Context, user.UpdateUserPasswordParams) error
	SetResetPasswordToken(ctx context.Context, params user.SetResetPasswordTokenParams) error
	GetUserIdByResetPasswordToken(ctx context.Context, token pgtype.Text) (pgtype.UUID, error)
}

type Service struct {
	config *config.Config
	store  Store
	userDb UserStore
	mailer mailer.Mailer
}

func NewService(c *config.Config, repo Store, userDb UserStore, mailer mailer.Mailer) *Service {
	return &Service{
		config: c,
		store:  repo,
		userDb: userDb,
		mailer: mailer,
	}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*TokenPair, error) {
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

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*TokenPair, error) {
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

func (s *Service) Logout(ctx context.Context, incomingRT string) error {
	return s.store.DeleteRefreshToken(ctx, incomingRT)
}

func (s *Service) RefreshTokens(ctx context.Context, incomingRT string) (*TokenPair, error) {
	refreshToken, err := s.store.GetRefreshToken(ctx, incomingRT)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidRefreshToken
		}

		return nil, err
	}

	if time.Now().After(refreshToken.ExpiresAt.Time) {
		s.store.DeleteRefreshToken(ctx, incomingRT)
		return nil, ErrInvalidRefreshToken
	}

	s.store.DeleteRefreshToken(ctx, incomingRT)

	return s.generateTokenPair(ctx, refreshToken.UserID)
}

func (s *Service) generateTokenPair(ctx context.Context, userID pgtype.UUID) (*TokenPair, error) {
	accessToken, err := jwt.GenerateAccessToken(userID, s.config.JWT.Secret)
	if err != nil {
		return nil, err
	}

	refreshToken, err := httputil.GenerateSecureToken()
	if err != nil {
		return nil, err
	}

	err = s.store.CreateRefreshToken(ctx, auth.CreateRefreshTokenParams{
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

func (s *Service) ForgotPassword(ctx context.Context, req *ForgotPasswordRequest) error {
	userRow, err := s.userDb.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil
	}

	// TODO 429 status code

	token, err := httputil.GenerateSecureToken()
	if err != nil {
		return err
	}

	dbParams := user.SetResetPasswordTokenParams{
		ResetPasswordToken: httputil.PgTextFromString(&token),
		ID:                 userRow.ID,
	}

	if err := s.userDb.SetResetPasswordToken(ctx, dbParams); err != nil {
		return err
	}

	go func() {
		if err := s.mailer.SendPasswordReset(context.Background(), userRow.Email, token); err != nil {
			log.Printf("Failed to send email to %s: %v", userRow.Email, err)
		}
	}()

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, req *ResetPasswordRequest) error {
	userId, err := s.userDb.GetUserIdByResetPasswordToken(ctx, httputil.PgTextFromString(&req.Token))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidResetToken
		}
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	dbParams := user.UpdateUserPasswordParams{
		PasswordHash: string(passwordHash),
		ID:           userId,
	}

	return s.userDb.UpdateUserPassword(ctx, dbParams)
}
