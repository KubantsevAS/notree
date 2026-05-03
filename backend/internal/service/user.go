package service

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/KubantsevAS/notree/backend/internal/httputil"
	"github.com/KubantsevAS/notree/backend/internal/mailer"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db     *user.Queries
	mailer mailer.Mailer
}

func NewUserService(db *user.Queries, mailer mailer.Mailer) *UserService {
	return &UserService{
		db:     db,
		mailer: mailer,
	}
}

func (s *UserService) GetUserById(ctx context.Context, id pgtype.UUID) (dto.GetProfileResponse, error) {
	user, err := s.db.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.GetProfileResponse{}, ErrUserNotFound
		}
		return dto.GetProfileResponse{}, err
	}

	response := dto.GetProfileResponse{
		ID:              user.ID.String(),
		Email:           user.Email,
		Username:        &user.Username.String,
		AvatarUrl:       &user.AvatarUrl.String,
		Timezone:        &user.Timezone.String,
		Locale:          &user.Locale.String,
		Preferences:     &user.Preferences,
		IsEmailVerified: &user.IsEmailVerified.Bool,
		LastLoginAt:     &user.LastLoginAt.Time,
		CreatedAt:       &user.CreatedAt.Time,
		UpdatedAt:       &user.UpdatedAt.Time,
	}

	return response, nil
}

func (s *UserService) UpdateUserProfile(ctx context.Context, id pgtype.UUID, req *dto.UpdateUserProfileRequest) (dto.UpdateUserProfileResponse, error) {
	dbParams := user.UpdateUserProfileParams{
		Username:  httputil.PgTextFromString(req.Username),
		AvatarUrl: httputil.PgTextFromString(req.AvatarUrl),
		ID:        id,
	}

	dbRow, err := s.db.UpdateUserProfile(ctx, dbParams)
	if err != nil {
		return dto.UpdateUserProfileResponse{}, err
	}

	response := dto.UpdateUserProfileResponse{
		Username:  &dbRow.Username.String,
		AvatarUrl: &dbRow.AvatarUrl.String,
		UpdatedAt: &dbRow.UpdatedAt.Time,
	}

	return response, nil
}

func (s *UserService) UpdateUserPreferences(ctx context.Context, id pgtype.UUID, req *dto.UpdateUserPreferencesRequest) (dto.UpdateUserPreferencesResponse, error) {
	dbParams := user.UpdateUserPreferencesParams{
		Locale:      httputil.PgTextFromString(req.Locale),
		Timezone:    httputil.PgTextFromString(req.Timezone),
		Preferences: httputil.RawMsgFromPtr(req.Preferences),
		ID:          id,
	}

	dbRow, err := s.db.UpdateUserPreferences(ctx, dbParams)
	if err != nil {
		return dto.UpdateUserPreferencesResponse{}, err
	}

	response := dto.UpdateUserPreferencesResponse{
		Locale:      &dbRow.Locale.String,
		Timezone:    &dbRow.Timezone.String,
		Preferences: &dbRow.Preferences,
		UpdatedAt:   &dbRow.UpdatedAt.Time,
	}

	return response, nil
}

func (s *UserService) UpdateUserPassword(ctx context.Context, id pgtype.UUID, req *dto.ChangePasswordRequest) error {
	hashRow, err := s.db.GetUserPasswordHashById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashRow), []byte(req.OldPassword)); err != nil {
		return ErrWrongCredentials
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	dbParams := user.UpdateUserPasswordParams{
		PasswordHash: string(passwordHash),
		ID:           id,
	}

	return s.db.UpdateUserPassword(ctx, dbParams)
}

func (s *UserService) SendVerificationEmail(ctx context.Context, id pgtype.UUID) error {
	userRow, err := s.db.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if userRow.IsEmailVerified.Bool {
		return nil
	}

	// TODO 429 status code

	token, err := httputil.GenerateSecureToken()
	if err != nil {
		return err
	}

	dbParams := user.SetVerificationTokenParams{
		VerificationToken: httputil.PgTextFromString(&token),
		ID:                id,
	}

	if err := s.db.SetVerificationToken(ctx, dbParams); err != nil {
		return err
	}

	go func() {
		if err := s.mailer.SendVerificationEmail(context.Background(), userRow.Email, token); err != nil {
			log.Printf("Failed to send email to %s: %v", userRow.Email, err)
		}
	}()

	return nil
}

func (s *UserService) VerifyEmailByToken(ctx context.Context, userID pgtype.UUID, token string) error {
	dbParams := user.VerifyEmailByTokenParams{
		ID:                userID,
		VerificationToken: httputil.PgTextFromString(&token),
	}

	if _, err := s.db.VerifyEmailByToken(ctx, dbParams); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidVerificationToken
		}
		return err
	}

	return nil
}
