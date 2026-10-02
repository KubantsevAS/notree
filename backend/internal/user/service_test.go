package user_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	userDb "github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/domain"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/KubantsevAS/notree/backend/internal/user"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var (
	userID = testutil.UUIDFromString(testutil.UUID1)
)

func TestGetUserById(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		username *string
		verified *bool
	}{
		{name: "all optional fields", email: "test@example.com", username: testutil.StringPtr("john_doe"), verified: testutil.BoolPtr(true)},
		{name: "only email", email: "minimal@example.com"},
		{name: "not verified", email: "unverified@example.com", username: testutil.StringPtr("jane"), verified: testutil.BoolPtr(false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &userStoreFake{
				getUserByIdResult: &userDb.UsersPublic{
					ID:              userID,
					Email:           tt.email,
					Username:        testutil.PgText(tt.username),
					IsEmailVerified: testutil.PgBool(tt.verified),
				},
			}

			profile, err := user.NewService(store, nil).GetUserById(context.Background(), userID)

			require.NoError(t, err)
			require.Equal(t, userID.String(), profile.ID)
			require.Equal(t, tt.email, profile.Email)
			if tt.username != nil {
				require.Equal(t, *tt.username, *profile.Username)
			}
			if tt.verified != nil {
				require.Equal(t, *tt.verified, *profile.IsEmailVerified)
			}
		})
	}
}

func TestService_Errors(t *testing.T) {
	ctx := context.Background()
	hash, err := bcrypt.GenerateFromPassword([]byte("current"), bcrypt.MinCost)
	require.NoError(t, err)

	tests := []struct {
		name    string
		store   *userStoreFake
		call    func(*user.Service) error
		wantErr error
	}{
		{
			name:  "get user: not found",
			store: &userStoreFake{getUserByIdErr: sql.ErrNoRows},
			call: func(s *user.Service) error {
				_, err := s.GetUserById(ctx, userID)
				return err
			},
			wantErr: user.ErrUserNotFound,
		},
		{
			name:  "update profile: empty update",
			store: &userStoreFake{},
			call: func(s *user.Service) error {
				_, err := s.UpdateUserProfile(ctx, userID, &user.UpdateUserProfileRequest{})
				return err
			},
			wantErr: domain.ErrEmptyUpdate,
		},
		{
			name:  "update preferences: empty update",
			store: &userStoreFake{},
			call: func(s *user.Service) error {
				_, err := s.UpdateUserPreferences(ctx, userID, &user.UpdateUserPreferencesRequest{})
				return err
			},
			wantErr: domain.ErrEmptyUpdate,
		},
		{
			name:  "change password: wrong old password",
			store: &userStoreFake{getUserPasswordHashResult: string(hash)},
			call: func(s *user.Service) error {
				return s.UpdateUserPassword(ctx, userID, &user.ChangePasswordRequest{OldPassword: "wrong", NewPassword: "newpass"})
			},
			wantErr: user.ErrWrongCredentials,
		},
		{
			name:  "change password: user not found",
			store: &userStoreFake{getUserPasswordHashErr: sql.ErrNoRows},
			call: func(s *user.Service) error {
				return s.UpdateUserPassword(ctx, userID, &user.ChangePasswordRequest{OldPassword: "current", NewPassword: "newpass"})
			},
			wantErr: user.ErrUserNotFound,
		},
		{
			name:  "verify email: invalid token",
			store: &userStoreFake{verifyEmailByTokenErr: sql.ErrNoRows},
			call: func(s *user.Service) error {
				return s.VerifyEmailByToken(ctx, userID, "invalid-token")
			},
			wantErr: user.ErrInvalidVerificationToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call(user.NewService(tt.store, nil))
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestUpdateUserProfile(t *testing.T) {
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

func TestUpdateUserPreferences(t *testing.T) {
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

func TestUpdateUserPassword(t *testing.T) {
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

func TestVerifyEmailByToken(t *testing.T) {
	token := "valid-token-123"

	repo := &userStoreFake{
		verifyEmailByTokenResult: userID,
	}

	svc := user.NewService(repo, nil)
	ctx := context.Background()

	err := svc.VerifyEmailByToken(ctx, userID, token)

	require.NoError(t, err)
}
