package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/KubantsevAS/notree/backend/internal/auth"
	"github.com/KubantsevAS/notree/backend/internal/config"
	authDb "github.com/KubantsevAS/notree/backend/internal/db/auth"
	userDb "github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var (
	userID = "11111111-1111-4111-8111-111111111111"
)

type authStoreFake struct {
	getRefreshTokenResult authDb.RefreshToken
	getRefreshTokenErr    error
	createRefreshTokenErr error
	createRefreshParams   []authDb.CreateRefreshTokenParams
	deleteRefreshTokenErr error
	deleteRefreshTokenArg []string
}

func (f *authStoreFake) CreateRefreshToken(ctx context.Context, params authDb.CreateRefreshTokenParams) error {
	f.createRefreshParams = append(f.createRefreshParams, params)
	return f.createRefreshTokenErr
}

func (f *authStoreFake) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	f.deleteRefreshTokenArg = append(f.deleteRefreshTokenArg, tokenHash)
	return f.deleteRefreshTokenErr
}

func (f *authStoreFake) GetRefreshToken(ctx context.Context, tokenHash string) (authDb.RefreshToken, error) {
	return f.getRefreshTokenResult, f.getRefreshTokenErr
}

type userStoreFake struct {
	getUserByEmailResult           userDb.User
	getUserByEmailErr              error
	createUserResult               pgtype.UUID
	createUserErr                  error
	createUserParams               []userDb.CreateUserParams
	setResetPasswordTokenParams    []userDb.SetResetPasswordTokenParams
	setResetPasswordTokenErr       error
	getUserIdByResetPasswordResult pgtype.UUID
	getUserIdByResetPasswordErr    error
	updateUserPasswordParams       []userDb.UpdateUserPasswordParams
	updateUserPasswordErr          error
}

func (f *userStoreFake) GetUserByEmail(_ context.Context, _ string) (userDb.User, error) {
	return f.getUserByEmailResult, f.getUserByEmailErr
}

func (f *userStoreFake) CreateUser(_ context.Context, params userDb.CreateUserParams) (pgtype.UUID, error) {
	f.createUserParams = append(f.createUserParams, params)
	return f.createUserResult, f.createUserErr
}

func (f *userStoreFake) SetResetPasswordToken(_ context.Context, params userDb.SetResetPasswordTokenParams) error {
	f.setResetPasswordTokenParams = append(f.setResetPasswordTokenParams, params)
	return f.setResetPasswordTokenErr
}

func (f *userStoreFake) GetUserIdByResetPasswordToken(_ context.Context, _ pgtype.Text) (pgtype.UUID, error) {
	return f.getUserIdByResetPasswordResult, f.getUserIdByResetPasswordErr
}

func (f *userStoreFake) UpdateUserPassword(_ context.Context, params userDb.UpdateUserPasswordParams) error {
	f.updateUserPasswordParams = append(f.updateUserPasswordParams, params)
	return f.updateUserPasswordErr
}

func TestAuthServiceRegisterSuccess(t *testing.T) {
	store := &authStoreFake{}
	userStore := &userStoreFake{
		getUserByEmailErr: sql.ErrNoRows,
		createUserResult:  testutil.UUIDFromString(userID),
	}

	svc := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: "secret"}},
		store,
		userStore,
		nil,
	)
	tokens, err := svc.Register(
		context.Background(),
		&auth.RegisterRequest{Email: "new@example.com", Password: "password123"},
	)

	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)
	require.Len(t, store.createRefreshParams, 1)
	require.Equal(t, testutil.UUIDFromString(userID), store.createRefreshParams[0].UserID)
	require.Equal(t, "new@example.com", userStore.createUserParams[0].Email)
}

func TestAuthServiceRegisterUserAlreadyExists(t *testing.T) {
	userStore := &userStoreFake{
		getUserByEmailResult: userDb.User{Email: "exists@example.com"},
	}
	store := &authStoreFake{}
	service := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: "secret"}},
		store,
		userStore,
		nil,
	)

	_, err := service.Register(
		context.Background(),
		&auth.RegisterRequest{Email: "exists@example.com", Password: "password123"},
	)

	require.ErrorIs(t, err, auth.ErrUserExist)
}

func TestAuthServiceRegisterDBErrorOnCreateUser(t *testing.T) {
	dbErr := errors.New("connection lost")

	store := &authStoreFake{}
	userStore := &userStoreFake{
		getUserByEmailErr: sql.ErrNoRows,
		createUserErr:     dbErr,
	}

	svc := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: "secret"}},
		store,
		userStore,
		nil,
	)

	_, err := svc.Register(
		context.Background(),
		&auth.RegisterRequest{Email: "new@example.com", Password: "password123"},
	)

	require.ErrorIs(t, err, dbErr)
	require.Empty(t, store.createRefreshParams)
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	store := &authStoreFake{}
	userStore := &userStoreFake{
		getUserByEmailResult: userDb.User{
			ID:           testutil.UUIDFromString(userID),
			Email:        "user@example.com",
			PasswordHash: string(hash),
		},
	}
	service := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: "secret"}},
		store,
		userStore,
		nil,
	)

	tokens, err := service.Login(
		context.Background(),
		&auth.LoginRequest{Email: "user@example.com", Password: "password123"},
	)

	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)
	require.Len(t, store.createRefreshParams, 1)
}

func TestAuthServiceLoginWrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	store := &authStoreFake{}
	userStore := &userStoreFake{
		getUserByEmailResult: userDb.User{
			ID:           testutil.UUIDFromString(userID),
			Email:        "user@example.com",
			PasswordHash: string(hash),
		},
	}
	service := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: "secret"}},
		store,
		userStore,
		nil,
	)

	_, err = service.Login(
		context.Background(),
		&auth.LoginRequest{Email: "user@example.com", Password: "wrong-password"},
	)

	require.ErrorIs(t, err, auth.ErrWrongCredentials)
	require.Empty(t, store.createRefreshParams)
}

func TestAuthServiceLoginWrongCredentials(t *testing.T) {
	service := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: "secret"}},
		&authStoreFake{},
		&userStoreFake{getUserByEmailErr: sql.ErrNoRows},
		nil,
	)

	_, err := service.Login(
		context.Background(),
		&auth.LoginRequest{Email: "missing@example.com", Password: "password123"},
	)

	require.ErrorIs(t, err, auth.ErrWrongCredentials)
}

func TestAuthServiceLoginDBErrorOnCreateRefreshToken(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	dbErr := errors.New("db timeout")
	store := &authStoreFake{
		createRefreshTokenErr: dbErr,
	}
	userStore := &userStoreFake{
		getUserByEmailResult: userDb.User{
			ID:           testutil.UUIDFromString(userID),
			Email:        "user@example.com",
			PasswordHash: string(hash),
		},
	}
	service := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: "secret"}},
		store,
		userStore,
		nil,
	)

	_, err = service.Login(
		context.Background(),
		&auth.LoginRequest{Email: "user@example.com", Password: "password123"},
	)

	require.ErrorIs(t, err, dbErr)
}

func TestAuthServiceRefreshTokensInvalidToken(t *testing.T) {
	service := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: "secret"}},
		&authStoreFake{getRefreshTokenErr: sql.ErrNoRows},
		&userStoreFake{},
		nil,
	)

	_, err := service.RefreshTokens(context.Background(), "bad-token")

	require.ErrorIs(t, err, auth.ErrInvalidRefreshToken)
}

func TestAuthServiceRefreshTokensExpiredToken(t *testing.T) {
	store := &authStoreFake{
		getRefreshTokenResult: authDb.RefreshToken{
			TokenHash: "expired-token",
			UserID:    testutil.UUIDFromStringT(t, userID),
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Hour), Valid: true},
		},
	}
	service := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: "secret"}},
		store,
		&userStoreFake{},
		nil,
	)

	_, err := service.RefreshTokens(context.Background(), "expired-token")

	require.ErrorIs(t, err, auth.ErrInvalidRefreshToken)
	require.Equal(t, []string{"expired-token"}, store.deleteRefreshTokenArg)
}

func TestAuthServiceForgotPasswordSuccess(t *testing.T) {
	email := "reset@example.com"
	userId := testutil.UUIDFromStringT(t, userID)
	mailer := &fakeMailer{sent: make(chan string, 1)}
	userStore := &userStoreFake{
		getUserByEmailResult: userDb.User{ID: userId, Email: email},
	}
	service := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: "secret"}},
		&authStoreFake{},
		userStore,
		mailer,
	)

	err := service.ForgotPassword(context.Background(), &auth.ForgotPasswordRequest{Email: email})

	require.NoError(t, err)
	require.Len(t, userStore.setResetPasswordTokenParams, 1)
	require.Equal(t, userId, userStore.setResetPasswordTokenParams[0].ID)

	select {
	case token := <-mailer.sent:
		require.NotEmpty(t, token)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected password reset email to be sent")
	}
}

func TestAuthServiceForgotPasswordUserNotFound(t *testing.T) {
	mailer := &fakeMailer{sent: make(chan string, 1)}
	userStore := &userStoreFake{getUserByEmailErr: sql.ErrNoRows}

	service := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: "secret"}},
		&authStoreFake{},
		userStore,
		mailer,
	)

	err := service.ForgotPassword(context.Background(), &auth.ForgotPasswordRequest{Email: "missing@example.com"})

	require.NoError(t, err)
	require.Empty(t, userStore.setResetPasswordTokenParams)

	select {
	case <-mailer.sent:
		t.Fatal("email should not be sent for non-existent user")
	default:
	}
}

func TestAuthServiceResetPasswordSuccess(t *testing.T) {
	userId := testutil.UUIDFromStringT(t, userID)
	userStore := &userStoreFake{
		getUserIdByResetPasswordResult: userId,
	}
	service := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: "secret"}},
		&authStoreFake{},
		userStore,
		nil,
	)

	err := service.ResetPassword(
		context.Background(),
		&auth.ResetPasswordRequest{Token: "token-123", NewPassword: "new-password-123"},
	)

	require.NoError(t, err)
	require.Len(t, userStore.updateUserPasswordParams, 1)
	require.Equal(t, userId, userStore.updateUserPasswordParams[0].ID)
	passwordHash := userStore.updateUserPasswordParams[0].PasswordHash
	require.NotEmpty(t, passwordHash)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("new-password-123")))
}

type fakeMailer struct {
	sent chan string
}

func (m *fakeMailer) SendPasswordReset(ctx context.Context, email string, token string) error {
	if m.sent != nil {
		m.sent <- token
	}
	return nil
}

func (m *fakeMailer) SendVerificationEmail(ctx context.Context, email string, token string) error {
	return nil
}
