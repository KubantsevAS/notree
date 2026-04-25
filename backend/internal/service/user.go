package service

import (
	"context"

	"github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/http/dto"
	"github.com/KubantsevAS/notree/backend/internal/httputil"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserService struct {
	db *user.Queries
}

func NewUserService(db *user.Queries) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetUserById(ctx context.Context, id pgtype.UUID) (dto.GetProfileResponse, error) {
	user, err := s.db.GetUserById(ctx, id)
	if err != nil {
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

func (s *UserService) UpdateUserProfile(ctx context.Context, id pgtype.UUID, arg dto.UpdateUserProfileRequest) (dto.UpdateUserProfileResponse, error) {
	dbParams := user.UpdateUserProfileParams{
		Username:  httputil.PgTextFromString(arg.Username),
		AvatarUrl: httputil.PgTextFromString(arg.AvatarUrl),
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

func (s *UserService) UpdateUserPreferences(ctx context.Context, id pgtype.UUID, arg dto.UpdateUserPreferencesRequest) (dto.UpdateUserPreferencesResponse, error) {
	dbParams := user.UpdateUserPreferencesParams{
		Locale:      httputil.PgTextFromString(arg.Locale),
		Timezone:    httputil.PgTextFromString(arg.Timezone),
		Preferences: httputil.RawMsgFromPtr(arg.Preferences),
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
