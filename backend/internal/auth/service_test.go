package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/KubantsevAS/notree/backend/internal/auth"
	authDb "github.com/KubantsevAS/notree/backend/internal/db/auth"
	userDb "github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister(t *testing.T) {
	testUserID := testutil.UUIDFromString(testutil.UUID1)
	store := &authStoreFake{}
	userStore := &userStoreFake{
		getUserByEmailErr: sql.ErrNoRows,
		createUserResult:  testUserID,
	}

	svc := newService(store, userStore, nil)
	tokens, err := svc.Register(
		context.Background(),
		&auth.RegisterRequest{Email: "new@example.com", Password: "password123"},
	)

	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)
	require.Len(t, store.createRefreshParams, 1)
	require.Equal(t, testUserID, store.createRefreshParams[0].UserID)
	require.Equal(t, "new@example.com", userStore.createUserParams[0].Email)
}

func TestLogin(t *testing.T) {
	testUserID := testutil.UUIDFromString(testutil.UUID1)
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	store := &authStoreFake{}
	userStore := &userStoreFake{
		getUserByEmailResult: userDb.User{
			ID:           testUserID,
			Email:        "user@example.com",
			PasswordHash: string(hash),
		},
	}
	service := newService(store, userStore, nil)

	tokens, err := service.Login(
		context.Background(),
		&auth.LoginRequest{Email: "user@example.com", Password: "password123"},
	)

	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)
	require.Len(t, store.createRefreshParams, 1)
}

func TestForgotPassword(t *testing.T) {
	email := "reset@example.com"
	testUserID := testutil.UUIDFromStringT(t, testutil.UUID1)
	mailer := &fakeMailer{sent: make(chan string, 1)}
	userStore := &userStoreFake{
		getUserByEmailResult: userDb.User{ID: testUserID, Email: email},
	}
	service := newService(&authStoreFake{}, userStore, mailer)

	err := service.ForgotPassword(context.Background(), &auth.ForgotPasswordRequest{Email: email})

	require.NoError(t, err)
	require.Len(t, userStore.setResetPasswordTokenParams, 1)
	require.Equal(t, testUserID, userStore.setResetPasswordTokenParams[0].ID)

	select {
	case token := <-mailer.sent:
		require.NotEmpty(t, token)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected password reset email to be sent")
	}
}

func TestForgotPassword_UserNotFound(t *testing.T) {
	mailer := &fakeMailer{sent: make(chan string, 1)}
	userStore := &userStoreFake{getUserByEmailErr: sql.ErrNoRows}

	service := newService(&authStoreFake{}, userStore, mailer)

	err := service.ForgotPassword(context.Background(), &auth.ForgotPasswordRequest{Email: "missing@example.com"})

	require.NoError(t, err)
	require.Empty(t, userStore.setResetPasswordTokenParams)

	select {
	case <-mailer.sent:
		t.Fatal("email should not be sent for non-existent user")
	default:
	}
}

func TestResetPassword(t *testing.T) {
	testUserID := testutil.UUIDFromStringT(t, testutil.UUID1)
	userStore := &userStoreFake{
		getUserIdByResetPasswordResult: testUserID,
	}
	service := newService(&authStoreFake{}, userStore, nil)

	err := service.ResetPassword(
		context.Background(),
		&auth.ResetPasswordRequest{Token: "token-123", NewPassword: "new-password-123"},
	)

	require.NoError(t, err)
	require.Len(t, userStore.updateUserPasswordParams, 1)
	require.Equal(t, testUserID, userStore.updateUserPasswordParams[0].ID)
	passwordHash := userStore.updateUserPasswordParams[0].PasswordHash
	require.NotEmpty(t, passwordHash)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("new-password-123")))
}

func TestService_Errors(t *testing.T) {
	ctx := context.Background()
	testUserID := testutil.UUIDFromStringT(t, testutil.UUID1)
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	require.NoError(t, err)
	existingUser := userDb.User{ID: testUserID, Email: "user@example.com", PasswordHash: string(hash)}
	dbErr := errors.New("db down")

	register := func(s *auth.Service) error {
		_, err := s.Register(ctx, &auth.RegisterRequest{Email: "new@example.com", Password: "password123"})
		return err
	}
	login := func(password string) func(*auth.Service) error {
		return func(s *auth.Service) error {
			_, err := s.Login(ctx, &auth.LoginRequest{Email: "user@example.com", Password: password})
			return err
		}
	}
	refresh := func(s *auth.Service) error {
		_, err := s.RefreshTokens(ctx, "refresh-token")
		return err
	}
	noRefreshTokenIssued := func(t *testing.T, store *authStoreFake, _ *userStoreFake) {
		t.Helper()
		require.Empty(t, store.createRefreshParams)
	}

	tests := []struct {
		name      string
		store     *authStoreFake
		userStore *userStoreFake
		call      func(*auth.Service) error
		wantErr   error
		check     func(*testing.T, *authStoreFake, *userStoreFake)
	}{
		{
			name:      "register: user already exists",
			userStore: &userStoreFake{getUserByEmailResult: userDb.User{Email: "new@example.com"}},
			call:      register,
			wantErr:   auth.ErrUserExist,
			check:     noRefreshTokenIssued,
		},
		{
			name:      "register: create user db error",
			userStore: &userStoreFake{getUserByEmailErr: sql.ErrNoRows, createUserErr: dbErr},
			call:      register,
			wantErr:   dbErr,
			check:     noRefreshTokenIssued,
		},
		{
			name:      "login: wrong password",
			userStore: &userStoreFake{getUserByEmailResult: existingUser},
			call:      login("wrong-password"),
			wantErr:   auth.ErrWrongCredentials,
			check:     noRefreshTokenIssued,
		},
		{
			name:      "login: user not found",
			userStore: &userStoreFake{getUserByEmailErr: sql.ErrNoRows},
			call:      login("password123"),
			wantErr:   auth.ErrWrongCredentials,
			check:     noRefreshTokenIssued,
		},
		{
			name:      "login: refresh token db error",
			store:     &authStoreFake{createRefreshTokenErr: dbErr},
			userStore: &userStoreFake{getUserByEmailResult: existingUser},
			call:      login("password123"),
			wantErr:   dbErr,
		},
		{
			name:    "refresh: unknown token",
			store:   &authStoreFake{getRefreshTokenErr: sql.ErrNoRows},
			call:    refresh,
			wantErr: auth.ErrInvalidRefreshToken,
		},
		{
			name: "refresh: expired token is revoked",
			store: &authStoreFake{getRefreshTokenResult: authDb.RefreshToken{
				TokenHash: "refresh-token",
				UserID:    testUserID,
				ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
			}},
			call:    refresh,
			wantErr: auth.ErrInvalidRefreshToken,
			check: func(t *testing.T, store *authStoreFake, _ *userStoreFake) {
				t.Helper()
				require.Equal(t, []string{"refresh-token"}, store.deleteRefreshTokenArg)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, userStore := tt.store, tt.userStore
			if store == nil {
				store = &authStoreFake{}
			}
			if userStore == nil {
				userStore = &userStoreFake{}
			}

			err := tt.call(newService(store, userStore, nil))

			require.ErrorIs(t, err, tt.wantErr)
			if tt.check != nil {
				tt.check(t, store, userStore)
			}
		})
	}
}
