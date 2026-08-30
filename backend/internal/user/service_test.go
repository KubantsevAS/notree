package user_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	userDb "github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/KubantsevAS/notree/backend/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var (
	userID = testutil.UUIDFromString("11111111-1111-4111-8111-111111111111")
)

type userStoreFake struct {
	getUserByIdResult             *userDb.UsersPublic
	getUserByIdErr                error
	getUserPasswordHashResult     string
	getUserPasswordHashErr        error
	createUserResult              pgtype.UUID
	createUserErr                 error
	setVerificationTokenErr       error
	setVerificationTokenParams    []userDb.SetVerificationTokenParams
	updateUserPasswordErr         error
	updateUserPasswordParams      []userDb.UpdateUserPasswordParams
	updateUserProfileResult       *userDb.UpdateUserProfileRow
	updateUserProfileErr          error
	updateUserProfileParams       []userDb.UpdateUserProfileParams
	updateUserPreferencesResult   *userDb.UpdateUserPreferencesRow
	updateUserPreferencesErr      error
	updateUserPreferencesParams   []userDb.UpdateUserPreferencesParams
	verifyEmailByTokenResult      pgtype.UUID
	verifyEmailByTokenErr         error
	verifyEmailByTokenParams      []userDb.VerifyEmailByTokenParams
	verifyEmailAlreadyVerified    bool
}

func (r *userStoreFake) GetUserById(ctx context.Context, id pgtype.UUID) (userDb.UsersPublic, error) {
	if r.getUserByIdErr != nil {
		return userDb.UsersPublic{}, r.getUserByIdErr
	}
	if r.getUserByIdResult == nil {
		return userDb.UsersPublic{}, sql.ErrNoRows
	}
	return *r.getUserByIdResult, nil
}

func (r *userStoreFake) CreateUser(ctx context.Context, params userDb.CreateUserParams) (pgtype.UUID, error) {
	if r.createUserErr != nil {
		return pgtype.UUID{}, r.createUserErr
	}
	return r.createUserResult, nil
}

func (r *userStoreFake) GetUserPasswordHashById(ctx context.Context, id pgtype.UUID) (string, error) {
	if r.getUserPasswordHashErr != nil {
		return "", r.getUserPasswordHashErr
	}
	return r.getUserPasswordHashResult, nil
}

func (r *userStoreFake) SetVerificationToken(ctx context.Context, params userDb.SetVerificationTokenParams) error {
	r.setVerificationTokenParams = append(r.setVerificationTokenParams, params)
	return r.setVerificationTokenErr
}

func (r *userStoreFake) UpdateUserPassword(ctx context.Context, params userDb.UpdateUserPasswordParams) error {
	r.updateUserPasswordParams = append(r.updateUserPasswordParams, params)
	return r.updateUserPasswordErr
}

func (r *userStoreFake) UpdateUserPreferences(ctx context.Context, params userDb.UpdateUserPreferencesParams) (userDb.UpdateUserPreferencesRow, error) {
	r.updateUserPreferencesParams = append(r.updateUserPreferencesParams, params)
	if r.updateUserPreferencesErr != nil {
		return userDb.UpdateUserPreferencesRow{}, r.updateUserPreferencesErr
	}
	if r.updateUserPreferencesResult == nil {
		return userDb.UpdateUserPreferencesRow{}, sql.ErrNoRows
	}
	return *r.updateUserPreferencesResult, nil
}

func (r *userStoreFake) UpdateUserProfile(ctx context.Context, params userDb.UpdateUserProfileParams) (userDb.UpdateUserProfileRow, error) {
	r.updateUserProfileParams = append(r.updateUserProfileParams, params)
	if r.updateUserProfileErr != nil {
		return userDb.UpdateUserProfileRow{}, r.updateUserProfileErr
	}
	if r.updateUserProfileResult == nil {
		return userDb.UpdateUserProfileRow{}, sql.ErrNoRows
	}
	return *r.updateUserProfileResult, nil
}

func (r *userStoreFake) VerifyEmailByToken(ctx context.Context, params userDb.VerifyEmailByTokenParams) (pgtype.UUID, error) {
	r.verifyEmailByTokenParams = append(r.verifyEmailByTokenParams, params)
	if r.verifyEmailByTokenErr != nil {
		return pgtype.UUID{}, r.verifyEmailByTokenErr
	}
	return r.verifyEmailByTokenResult, nil
}

func TestUserServiceGetUserByIdSuccess(t *testing.T) {
	email := "test@example.com"

	repo := &userStoreFake{
		getUserByIdResult: &userDb.UsersPublic{
			ID:    userID,
			Email: email,
		},
	}

	svc := user.NewService(repo, nil)
	ctx := context.Background()

	profile, err := svc.GetUserById(ctx, userID)

	require.NoError(t, err)
	require.Equal(t, userID.String(), profile.ID)
	require.Equal(t, email, profile.Email)
}

func TestUserServiceGetUserByIdNotFound(t *testing.T) {
	repo := &userStoreFake{
		getUserByIdErr: sql.ErrNoRows,
	}

	svc := user.NewService(repo, nil)
	ctx := context.Background()

	_, err := svc.GetUserById(ctx, userID)

	require.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestUserServiceUpdateUserProfileEmptyUpdate(t *testing.T) {
	repo := &userStoreFake{}
	svc := user.NewService(repo, nil)
	ctx := context.Background()

	req := &user.UpdateUserProfileRequest{
		Username:  nil,
		AvatarUrl: nil,
	}

	_, err := svc.UpdateUserProfile(ctx, userID, req)

	require.Error(t, err)
}

func TestUserServiceUpdateUserProfileSuccess(t *testing.T) {
	newUsername := "newusername"
	newAvatarUrl := "https://example.com/avatar.jpg"
	now := time.Now()

	repo := &userStoreFake{
		updateUserProfileResult: &userDb.UpdateUserProfileRow{
			Username:  testutil.PgText(&newUsername),
			AvatarUrl: testutil.PgText(&newAvatarUrl),
			UpdatedAt: testutil.PgTimestamptz(&now),
		},
	}

	svc := user.NewService(repo, nil)
	ctx := context.Background()

	req := &user.UpdateUserProfileRequest{
		Username:  &newUsername,
		AvatarUrl: &newAvatarUrl,
	}

	resp, err := svc.UpdateUserProfile(ctx, userID, req)

	require.NoError(t, err)
	require.Equal(t, &newUsername, resp.Username)
	require.Equal(t, &newAvatarUrl, resp.AvatarUrl)
}

func TestUserServiceUpdateUserPreferencesEmptyUpdate(t *testing.T) {
	repo := &userStoreFake{}
	svc := user.NewService(repo, nil)
	ctx := context.Background()

	req := &user.UpdateUserPreferencesRequest{
		Locale:      nil,
		Timezone:    nil,
		Preferences: nil,
	}

	_, err := svc.UpdateUserPreferences(ctx, userID, req)

	require.Error(t, err)
}

func TestUserServiceUpdateUserPreferencesSuccess(t *testing.T) {
	locale := "en-US"
	timezone := "America/New_York"
	prefs := json.RawMessage(`{"theme":"dark"}`)
	now := time.Now()

	repo := &userStoreFake{
		updateUserPreferencesResult: &userDb.UpdateUserPreferencesRow{
			Locale:      testutil.PgText(&locale),
			Timezone:    testutil.PgText(&timezone),
			Preferences: prefs,
			UpdatedAt:   testutil.PgTimestamptz(&now),
		},
	}

	svc := user.NewService(repo, nil)
	ctx := context.Background()

	req := &user.UpdateUserPreferencesRequest{
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
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("current"), bcrypt.DefaultCost)

	repo := &userStoreFake{
		getUserPasswordHashResult: string(passwordHash),
	}

	svc := user.NewService(repo, nil)
	ctx := context.Background()

	req := &user.ChangePasswordRequest{
		OldPassword: "wrong",
		NewPassword: "newpass",
	}

	err := svc.UpdateUserPassword(ctx, userID, req)

	require.ErrorIs(t, err, user.ErrWrongCredentials)
}

func TestUserServiceUpdateUserPasswordSuccess(t *testing.T) {
	currentPassword := "current"
	newPassword := "newpass"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)

	repo := &userStoreFake{
		getUserPasswordHashResult: string(passwordHash),
	}

	svc := user.NewService(repo, nil)
	ctx := context.Background()

	req := &user.ChangePasswordRequest{
		OldPassword: currentPassword,
		NewPassword: newPassword,
	}

	err := svc.UpdateUserPassword(ctx, userID, req)

	require.NoError(t, err)
}

func TestUserServiceUpdateUserPasswordUserNotFound(t *testing.T) {
	repo := &userStoreFake{
		getUserPasswordHashErr: sql.ErrNoRows,
	}

	svc := user.NewService(repo, nil)
	ctx := context.Background()

	req := &user.ChangePasswordRequest{
		OldPassword: "current",
		NewPassword: "newpass",
	}

	err := svc.UpdateUserPassword(ctx, userID, req)

	require.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestUserServiceVerifyEmailByTokenSuccess(t *testing.T) {
	token := "valid-token-123"

	repo := &userStoreFake{
		verifyEmailByTokenResult: userID,
	}

	svc := user.NewService(repo, nil)
	ctx := context.Background()

	err := svc.VerifyEmailByToken(ctx, userID, token)

	require.NoError(t, err)
}

func TestUserServiceVerifyEmailByTokenInvalidToken(t *testing.T) {
	token := "invalid-token"

	repo := &userStoreFake{
		verifyEmailByTokenErr: sql.ErrNoRows,
	}

	svc := user.NewService(repo, nil)
	ctx := context.Background()

	err := svc.VerifyEmailByToken(ctx, userID, token)

	require.ErrorIs(t, err, user.ErrInvalidVerificationToken)
}

func TestUserServiceGetUserByIdTableDriven(t *testing.T) {

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
			repo := &userStoreFake{
				getUserByIdResult: &userDb.UsersPublic{
					ID:              userID,
					Email:           tt.email,
					Username:        testutil.PgText(tt.username),
					IsEmailVerified: testutil.PgBool(tt.verified),
				},
			}

			svc := user.NewService(repo, nil)
			ctx := context.Background()

			profile, err := svc.GetUserById(ctx, userID)

			require.NoError(t, err)
			require.Equal(t, tt.email, profile.Email)
		})
	}
}
