package user

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/httputil"
	"github.com/KubantsevAS/notree/backend/internal/mailer"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type Store interface {
	GetUserById(context.Context, pgtype.UUID) (user.UsersPublic, error)
	CreateUser(context.Context, user.CreateUserParams) (pgtype.UUID, error)
	GetUserPasswordHashById(context.Context, pgtype.UUID) (string, error)
	SetVerificationToken(context.Context, user.SetVerificationTokenParams) error
	UpdateUserPassword(context.Context, user.UpdateUserPasswordParams) error
	UpdateUserPreferences(context.Context, user.UpdateUserPreferencesParams) (user.UpdateUserPreferencesRow, error)
	UpdateUserProfile(context.Context, user.UpdateUserProfileParams) (user.UpdateUserProfileRow, error)
	VerifyEmailByToken(context.Context, user.VerifyEmailByTokenParams) (pgtype.UUID, error)
}

type Service struct {
	store  Store
	mailer mailer.Mailer
}

func NewService(repo Store, mailer mailer.Mailer) *Service {
	return &Service{store: repo, mailer: mailer}
}

func (s *Service) GetUserById(ctx context.Context, id pgtype.UUID) (GetProfileResponse, error) {
	user, err := s.store.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GetProfileResponse{}, ErrUserNotFound
		}
		return GetProfileResponse{}, err
	}

	response := GetProfileResponse{
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

func (s *Service) UpdateUserProfile(ctx context.Context, id pgtype.UUID, req *UpdateUserProfileRequest) (UpdateUserProfileResponse, error) {
	if req.Username == nil && req.AvatarUrl == nil {
		return UpdateUserProfileResponse{}, ErrEmptyUpdate
	}

	dbParams := user.UpdateUserProfileParams{
		Username:  httputil.PgTextFromString(req.Username),
		AvatarUrl: httputil.PgTextFromString(req.AvatarUrl),
		ID:        id,
	}

	dbRow, err := s.store.UpdateUserProfile(ctx, dbParams)
	if err != nil {
		return UpdateUserProfileResponse{}, err
	}

	response := UpdateUserProfileResponse{
		Username:  &dbRow.Username.String,
		AvatarUrl: &dbRow.AvatarUrl.String,
		UpdatedAt: &dbRow.UpdatedAt.Time,
	}

	return response, nil
}

func (s *Service) UpdateUserPreferences(ctx context.Context, id pgtype.UUID, req *UpdateUserPreferencesRequest) (UpdateUserPreferencesResponse, error) {
	if req.Locale == nil && req.Timezone == nil && req.Preferences == nil {
		return UpdateUserPreferencesResponse{}, ErrEmptyUpdate
	}

	dbParams := user.UpdateUserPreferencesParams{
		Locale:      httputil.PgTextFromString(req.Locale),
		Timezone:    httputil.PgTextFromString(req.Timezone),
		Preferences: httputil.RawMsgFromPtr(req.Preferences),
		ID:          id,
	}

	dbRow, err := s.store.UpdateUserPreferences(ctx, dbParams)
	if err != nil {
		return UpdateUserPreferencesResponse{}, err
	}

	response := UpdateUserPreferencesResponse{
		Locale:      &dbRow.Locale.String,
		Timezone:    &dbRow.Timezone.String,
		Preferences: &dbRow.Preferences,
		UpdatedAt:   &dbRow.UpdatedAt.Time,
	}

	return response, nil
}

func (s *Service) UpdateUserPassword(ctx context.Context, id pgtype.UUID, req *ChangePasswordRequest) error {
	hashRow, err := s.store.GetUserPasswordHashById(ctx, id)
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

	return s.store.UpdateUserPassword(ctx, dbParams)
}

func (s *Service) SendVerificationEmail(ctx context.Context, id pgtype.UUID) error {
	userRow, err := s.store.GetUserById(ctx, id)
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

	if err := s.store.SetVerificationToken(ctx, dbParams); err != nil {
		return err
	}

	go func() {
		if err := s.mailer.SendVerificationEmail(context.Background(), userRow.Email, token); err != nil {
			log.Printf("Failed to send email to %s: %v", userRow.Email, err)
		}
	}()

	return nil
}

func (s *Service) VerifyEmailByToken(ctx context.Context, userID pgtype.UUID, token string) error {
	dbParams := user.VerifyEmailByTokenParams{
		ID:                userID,
		VerificationToken: httputil.PgTextFromString(&token),
	}

	if _, err := s.store.VerifyEmailByToken(ctx, dbParams); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidVerificationToken
		}
		return err
	}

	return nil
}
