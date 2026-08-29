package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type userRepositoryFake struct {
	getUserByIdResult           *user.UsersPublic
	getUserByIdErr              error
	getUserPasswordHashResult   string
	getUserPasswordHashErr      error
	createUserResult            pgtype.UUID
	createUserErr               error
	setVerificationTokenErr     error
	updateUserPasswordErr       error
	updateUserProfileResult     *user.UpdateUserProfileRow
	updateUserProfileErr        error
	updateUserPreferencesResult *user.UpdateUserPreferencesRow
	updateUserPreferencesErr    error
	verifyEmailByTokenResult    pgtype.UUID
	verifyEmailByTokenErr       error
	verifyEmailAlreadyVerified  bool
}

func (r *userRepositoryFake) GetUserById(ctx context.Context, id pgtype.UUID) (user.UsersPublic, error) {
	if r.getUserByIdErr != nil {
		return user.UsersPublic{}, r.getUserByIdErr
	}
	if r.getUserByIdResult == nil {
		return user.UsersPublic{}, sql.ErrNoRows
	}
	return *r.getUserByIdResult, nil
}

func (r *userRepositoryFake) CreateUser(ctx context.Context, params user.CreateUserParams) (pgtype.UUID, error) {
	if r.createUserErr != nil {
		return pgtype.UUID{}, r.createUserErr
	}
	return r.createUserResult, nil
}

func (r *userRepositoryFake) GetUserPasswordHashById(ctx context.Context, id pgtype.UUID) (string, error) {
	if r.getUserPasswordHashErr != nil {
		return "", r.getUserPasswordHashErr
	}
	return r.getUserPasswordHashResult, nil
}

func (r *userRepositoryFake) SetVerificationToken(ctx context.Context, params user.SetVerificationTokenParams) error {
	return r.setVerificationTokenErr
}

func (r *userRepositoryFake) UpdateUserPassword(ctx context.Context, params user.UpdateUserPasswordParams) error {
	return r.updateUserPasswordErr
}

func (r *userRepositoryFake) UpdateUserPreferences(ctx context.Context, params user.UpdateUserPreferencesParams) (user.UpdateUserPreferencesRow, error) {
	if r.updateUserPreferencesErr != nil {
		return user.UpdateUserPreferencesRow{}, r.updateUserPreferencesErr
	}
	if r.updateUserPreferencesResult == nil {
		return user.UpdateUserPreferencesRow{}, sql.ErrNoRows
	}
	return *r.updateUserPreferencesResult, nil
}

func (r *userRepositoryFake) UpdateUserProfile(ctx context.Context, params user.UpdateUserProfileParams) (user.UpdateUserProfileRow, error) {
	if r.updateUserProfileErr != nil {
		return user.UpdateUserProfileRow{}, r.updateUserProfileErr
	}
	if r.updateUserProfileResult == nil {
		return user.UpdateUserProfileRow{}, sql.ErrNoRows
	}
	return *r.updateUserProfileResult, nil
}

func (r *userRepositoryFake) VerifyEmailByToken(ctx context.Context, params user.VerifyEmailByTokenParams) (pgtype.UUID, error) {
	if r.verifyEmailByTokenErr != nil {
		return pgtype.UUID{}, r.verifyEmailByTokenErr
	}
	return r.verifyEmailByTokenResult, nil
}

func TestUserServiceGetUserByIdSuccess(t *testing.T) {
	userID := testutil.UUIDFromString("ebde9d75-dd29-4702-afde-1f93772f905d")
	email := "test@example.com"

	repo := &userRepositoryFake{
		getUserByIdResult: &user.UsersPublic{
			ID:    userID,
			Email: email,
		},
	}

	svc := NewService(repo, nil)
	ctx := context.Background()

	profile, err := svc.GetUserById(ctx, userID)

	require.NoError(t, err)
	require.Equal(t, userID.String(), profile.ID)
	require.Equal(t, email, profile.Email)
}

func TestUserServiceGetUserByIdNotFound(t *testing.T) {
	userID := testutil.UUIDFromString("ebde9d75-dd29-4702-afde-1f93772f905d")

	repo := &userRepositoryFake{
		getUserByIdErr: sql.ErrNoRows,
	}

	svc := NewService(repo, nil)
	ctx := context.Background()

	_, err := svc.GetUserById(ctx, userID)

	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserServiceUpdateUserProfileEmptyUpdate(t *testing.T) {
	userID := testutil.UUIDFromString("ebde9d75-dd29-4702-afde-1f93772f905d")
	repo := &userRepositoryFake{}
	svc := NewService(repo, nil)
	ctx := context.Background()

	req := &UpdateUserProfileRequest{
		Username:  nil,
		AvatarUrl: nil,
	}

	_, err := svc.UpdateUserProfile(ctx, userID, req)

	require.Error(t, err)
}

func TestUserServiceUpdateUserProfileSuccess(t *testing.T) {
	userID := testutil.UUIDFromString("ebde9d75-dd29-4702-afde-1f93772f905d")
	newUsername := "newusername"
	newAvatarUrl := "https://example.com/avatar.jpg"
	now := time.Now()

	repo := &userRepositoryFake{
		updateUserProfileResult: &user.UpdateUserProfileRow{
			Username:  testutil.PgText(&newUsername),
			AvatarUrl: testutil.PgText(&newAvatarUrl),
			UpdatedAt: testutil.PgTimestamptz(&now),
		},
	}

	svc := NewService(repo, nil)
	ctx := context.Background()

	req := &UpdateUserProfileRequest{
		Username:  &newUsername,
		AvatarUrl: &newAvatarUrl,
	}

	resp, err := svc.UpdateUserProfile(ctx, userID, req)

	require.NoError(t, err)
	require.Equal(t, &newUsername, resp.Username)
	require.Equal(t, &newAvatarUrl, resp.AvatarUrl)
}

func TestUserServiceUpdateUserPreferencesEmptyUpdate(t *testing.T) {
	userID := testutil.UUIDFromString("ebde9d75-dd29-4702-afde-1f93772f905d")
	repo := &userRepositoryFake{}
	svc := NewService(repo, nil)
	ctx := context.Background()

	req := &UpdateUserPreferencesRequest{
		Locale:      nil,
		Timezone:    nil,
		Preferences: nil,
	}

	_, err := svc.UpdateUserPreferences(ctx, userID, req)

	require.Error(t, err)
}

func TestUserServiceUpdateUserPreferencesSuccess(t *testing.T) {
	userID := testutil.UUIDFromString("ebde9d75-dd29-4702-afde-1f93772f905d")
	locale := "en-US"
	timezone := "America/New_York"
	prefs := json.RawMessage(`{"theme":"dark"}`)
	now := time.Now()

	repo := &userRepositoryFake{
		updateUserPreferencesResult: &user.UpdateUserPreferencesRow{
			Locale:      testutil.PgText(&locale),
			Timezone:    testutil.PgText(&timezone),
			Preferences: prefs,
			UpdatedAt:   testutil.PgTimestamptz(&now),
		},
	}

	svc := NewService(repo, nil)
	ctx := context.Background()

	req := &UpdateUserPreferencesRequest{
		Locale:      &locale,
		Timezone:    &timezone,
		Preferences: &prefs,
	}

	resp, err := svc.UpdateUserPreferences(ctx, userID, req)

	require.NoError(t, err)
	require.Equal(t, &locale, resp.Locale)
	require.Equal(t, &timezone, resp.Timezone)
}

func TestUserServiceUpdateUserPasswordWrongCurrentPassword(t *testing.T) {
	userID := testutil.UUIDFromString("ebde9d75-dd29-4702-afde-1f93772f905d")
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("current"), bcrypt.DefaultCost)

	repo := &userRepositoryFake{
		getUserPasswordHashResult: string(passwordHash),
	}

	svc := NewService(repo, nil)
	ctx := context.Background()

	req := &ChangePasswordRequest{
		OldPassword: "wrong",
		NewPassword: "newpass",
	}

	err := svc.UpdateUserPassword(ctx, userID, req)

	require.ErrorIs(t, err, ErrWrongCredentials)
}

func TestUserServiceUpdateUserPasswordSuccess(t *testing.T) {
	userID := testutil.UUIDFromString("ebde9d75-dd29-4702-afde-1f93772f905d")
	currentPassword := "current"
	newPassword := "newpass"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)

	repo := &userRepositoryFake{
		getUserPasswordHashResult: string(passwordHash),
	}

	svc := NewService(repo, nil)
	ctx := context.Background()

	req := &ChangePasswordRequest{
		OldPassword: currentPassword,
		NewPassword: newPassword,
	}

	err := svc.UpdateUserPassword(ctx, userID, req)

	require.NoError(t, err)
}

func TestUserServiceUpdateUserPasswordUserNotFound(t *testing.T) {
	userID := testutil.UUIDFromString("ebde9d75-dd29-4702-afde-1f93772f905d")

	repo := &userRepositoryFake{
		getUserPasswordHashErr: sql.ErrNoRows,
	}

	svc := NewService(repo, nil)
	ctx := context.Background()

	req := &ChangePasswordRequest{
		OldPassword: "current",
		NewPassword: "newpass",
	}

	err := svc.UpdateUserPassword(ctx, userID, req)

	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserServiceVerifyEmailByTokenSuccess(t *testing.T) {
	userID := testutil.UUIDFromString("ebde9d75-dd29-4702-afde-1f93772f905d")
	token := "valid-token-123"

	repo := &userRepositoryFake{
		verifyEmailByTokenResult: userID,
	}

	svc := NewService(repo, nil)
	ctx := context.Background()

	err := svc.VerifyEmailByToken(ctx, userID, token)

	require.NoError(t, err)
}

func TestUserServiceVerifyEmailByTokenInvalidToken(t *testing.T) {
	userID := testutil.UUIDFromString("ebde9d75-dd29-4702-afde-1f93772f905d")
	token := "invalid-token"

	repo := &userRepositoryFake{
		verifyEmailByTokenErr: sql.ErrNoRows,
	}

	svc := NewService(repo, nil)
	ctx := context.Background()

	err := svc.VerifyEmailByToken(ctx, userID, token)

	require.ErrorIs(t, err, ErrInvalidVerificationToken)
}

func TestUserServiceGetUserByIdTableDriven(t *testing.T) {
	userID := testutil.UUIDFromString("ebde9d75-dd29-4702-afde-1f93772f905d")

	tests := []struct {
		name     string
		email    string
		username *string
		verified *bool
	}{
		{
			name:     "User with all optional fields",
			email:    "test@example.com",
			username: func() *string { s := "john_doe"; return &s }(),
			verified: func() *bool { b := true; return &b }(),
		},
		{
			name:     "User with only email",
			email:    "minimal@example.com",
			username: nil,
			verified: nil,
		},
		{
			name:     "User not verified",
			email:    "unverified@example.com",
			username: func() *string { s := "jane"; return &s }(),
			verified: func() *bool { b := false; return &b }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &userRepositoryFake{
				getUserByIdResult: &user.UsersPublic{
					ID:              userID,
					Email:           tt.email,
					Username:        testutil.PgText(tt.username),
					IsEmailVerified: testutil.PgBool(tt.verified),
				},
			}

			svc := NewService(repo, nil)
			ctx := context.Background()

			profile, err := svc.GetUserById(ctx, userID)

			require.NoError(t, err)
			require.Equal(t, tt.email, profile.Email)
		})
	}
}
