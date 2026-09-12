package auth_test

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KubantsevAS/notree/backend/internal/auth"
	"github.com/KubantsevAS/notree/backend/internal/config"
	authDb "github.com/KubantsevAS/notree/backend/internal/db/auth"
	userDb "github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/http/middleware"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func newAuthHandlerWithFakes(t *testing.T, store *authStoreFake, userStore *userStoreFake, mailer *fakeMailer) *auth.Handler {
	t.Helper()

	service := auth.NewService(
		&config.Config{JWT: config.JWTConfig{Secret: testutil.TestSecret}},
		store,
		userStore,
		mailer,
	)

	return auth.NewHandler(service)
}

func TestHandlerRegister(t *testing.T) {
	testUserID := testutil.UUIDFromString(testutil.UUID1)
	userStore := &userStoreFake{
		getUserByEmailErr: sql.ErrNoRows,
		createUserResult:  testUserID,
	}
	handler := newAuthHandlerWithFakes(t, &authStoreFake{}, userStore, nil)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/register", &auth.RegisterRequest{
		Email:    "new@example.com",
		Password: "password123",
	})
	res := httptest.NewRecorder()

	handler.Register(res, req)

	require.Equal(t, http.StatusCreated, res.Code)

	testutil.AssertAuthCookies(t, res, middleware.CookieAccessToken, middleware.CookieRefreshToken)
	require.Equal(t, "new@example.com", userStore.createUserParams[0].Email)
}

func TestHandlerRegister_Conflict(t *testing.T) {
	userStore := &userStoreFake{
		getUserByEmailResult: userDb.User{Email: "exists@example.com"},
	}
	handler := newAuthHandlerWithFakes(t, &authStoreFake{}, userStore, nil)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/register", &auth.RegisterRequest{
		Email:    "exists@example.com",
		Password: "password123",
	})
	res := httptest.NewRecorder()

	handler.Register(res, req)

	require.Equal(t, http.StatusConflict, res.Code)
	testutil.AssertErrorJSON(t, res, auth.ErrUserExist.Error())
}

func TestHandlerRegister_InternalError(t *testing.T) {
	userStore := &userStoreFake{
		getUserByEmailErr: sql.ErrNoRows,
		createUserErr:     errors.New("db connection lost"),
	}
	handler := newAuthHandlerWithFakes(t, &authStoreFake{}, userStore, nil)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/register", &auth.RegisterRequest{Email: "new@example.com", Password: "password123"})
	res := httptest.NewRecorder()

	handler.Register(res, req)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	testutil.AssertErrorJSON(t, res, "internal server error")
}

func TestHandlerLogin(t *testing.T) {
	testUserID := testutil.UUIDFromString(testutil.UUID1)
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	userStore := &userStoreFake{
		getUserByEmailResult: userDb.User{
			ID:           testUserID,
			Email:        "user@example.com",
			PasswordHash: string(hash),
		},
	}
	handler := newAuthHandlerWithFakes(t, &authStoreFake{}, userStore, nil)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/login", &auth.LoginRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	res := httptest.NewRecorder()

	handler.Login(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	testutil.AssertAuthCookies(t, res, middleware.CookieAccessToken, middleware.CookieRefreshToken)
}

func TestHandlerLogin_Unauthorized(t *testing.T) {
	testUserID := testutil.UUIDFromString(testutil.UUID1)
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	userStore := &userStoreFake{
		getUserByEmailResult: userDb.User{
			ID:           testUserID,
			Email:        "user@example.com",
			PasswordHash: string(hash),
		},
	}
	handler := newAuthHandlerWithFakes(t, &authStoreFake{}, userStore, nil)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/login", &auth.LoginRequest{
		Email:    "user@example.com",
		Password: "wrong-password",
	})
	res := httptest.NewRecorder()

	handler.Login(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
	testutil.AssertErrorJSON(t, res, auth.ErrWrongCredentials.Error())
}

func TestHandler_RefreshTokens(t *testing.T) {
	testUserID := testutil.UUIDFromStringT(t, testutil.UUID1)
	store := &authStoreFake{
		getRefreshTokenResult: authDb.RefreshToken{
			TokenHash: "refresh-token-value",
			UserID:    testUserID,
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
		},
	}
	handler := newAuthHandlerWithFakes(t, store, &userStoreFake{}, nil)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/refresh-tokens", nil)
	req.AddCookie(&http.Cookie{Name: middleware.CookieRefreshToken, Value: "refresh-token-value"})
	res := httptest.NewRecorder()

	handler.RefreshTokens(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	testutil.AssertAuthCookies(t, res, middleware.CookieAccessToken, middleware.CookieRefreshToken)
	require.Len(t, store.deleteRefreshTokenArg, 1)
	require.Equal(t, "refresh-token-value", store.deleteRefreshTokenArg[0])
}

func TestHandlerRefreshTokens_MissingCookie(t *testing.T) {
	handler := newAuthHandlerWithFakes(t, &authStoreFake{}, &userStoreFake{}, nil)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/refresh-tokens", nil)
	res := httptest.NewRecorder()

	handler.RefreshTokens(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
	testutil.AssertErrorJSON(t, res, "missing refresh token")
}

func TestHandlerRefreshTokens_InvalidToken(t *testing.T) {
	store := &authStoreFake{
		getRefreshTokenErr: sql.ErrNoRows,
	}
	handler := newAuthHandlerWithFakes(t, store, &userStoreFake{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh-tokens", nil)
	req.AddCookie(&http.Cookie{Name: middleware.CookieRefreshToken, Value: "bad-token"})
	res := httptest.NewRecorder()

	handler.RefreshTokens(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
	testutil.AssertErrorJSON(t, res, auth.ErrInvalidRefreshToken.Error())
}

func TestHandlerLogout_ClearsCookies(t *testing.T) {
	handler := newAuthHandlerWithFakes(t, &authStoreFake{}, &userStoreFake{}, nil)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: middleware.CookieRefreshToken, Value: "old-refresh-token"})
	res := httptest.NewRecorder()

	handler.Logout(res, req)

	require.Equal(t, http.StatusNoContent, res.Code)
	cookies := testutil.AssertAuthCookies(t, res, middleware.CookieAccessToken, middleware.CookieRefreshToken)
	for _, cookie := range cookies {
		require.Equal(t, -1, cookie.MaxAge)
		require.Equal(t, "", cookie.Value)
	}
}

func TestHandlerLogout_WithoutCookie(t *testing.T) {
	handler := newAuthHandlerWithFakes(t, &authStoreFake{}, &userStoreFake{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	res := httptest.NewRecorder()

	handler.Logout(res, req)

	require.Equal(t, http.StatusNoContent, res.Code)
	cookies := testutil.AssertAuthCookies(t, res, middleware.CookieAccessToken, middleware.CookieRefreshToken)
	for _, cookie := range cookies {
		require.Equal(t, -1, cookie.MaxAge)
	}
}

func TestHandlerForgotPassword(t *testing.T) {
	testUserID := testutil.UUIDFromStringT(t, testutil.UUID1)
	mailer := &fakeMailer{sent: make(chan string, 1)}
	userStore := &userStoreFake{
		getUserByEmailResult: userDb.User{ID: testUserID, Email: "reset@example.com"},
	}
	handler := newAuthHandlerWithFakes(t, &authStoreFake{}, userStore, mailer)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/forgot-password", auth.ForgotPasswordRequest{
		Email: "reset@example.com",
	})
	res := httptest.NewRecorder()

	handler.ForgotPassword(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	testutil.AssertMessageJSON(t, res, "password reset link has been sent")
	require.Len(t, userStore.setResetPasswordTokenParams, 1)
	select {
	case token := <-mailer.sent:
		require.NotEmpty(t, token)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected password reset email to be sent")
	}
}

func TestHandlerResetPassword(t *testing.T) {
	testUserID := testutil.UUIDFromString(testutil.UUID1)
	userStore := &userStoreFake{
		getUserIdByResetPasswordResult: testUserID,
	}
	handler := newAuthHandlerWithFakes(t, &authStoreFake{}, userStore, nil)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/reset-password", auth.ResetPasswordRequest{
		Token:       "valid-token",
		NewPassword: "new-password-123",
	})
	res := httptest.NewRecorder()

	handler.ResetPassword(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	testutil.AssertMessageJSON(t, res, "password has been reset successfully")
	require.Len(t, userStore.updateUserPasswordParams, 1)
	require.Equal(t, testUserID, userStore.updateUserPasswordParams[0].ID)
}

func TestHandlerResetPassword_InvalidToken(t *testing.T) {
	userStore := &userStoreFake{
		getUserIdByResetPasswordErr: sql.ErrNoRows,
	}
	handler := newAuthHandlerWithFakes(t, &authStoreFake{}, userStore, nil)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/reset-password", auth.ResetPasswordRequest{
		Token:       "bad-token",
		NewPassword: "new-password-123",
	})
	res := httptest.NewRecorder()

	handler.ResetPassword(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "invalid or expired token")
}

func TestHandlerRequest_BodyValidation(t *testing.T) {
	handler := newAuthHandlerWithFakes(t, &authStoreFake{}, &userStoreFake{}, nil)

	for _, tc := range []struct {
		name       string
		path       string
		body       string
		statusCode int
	}{
		{name: "register invalid email", path: "/auth/register", body: `{"email":"not-an-email","password":"password123"}`, statusCode: http.StatusBadRequest},
		{name: "login invalid payload", path: "/auth/login", body: `{"email":"user@example.com"}`, statusCode: http.StatusBadRequest},
		{name: "forgot password invalid email", path: "/auth/forgot-password", body: `{"email":"nope"}`, statusCode: http.StatusBadRequest},
		{name: "reset password invalid token", path: "/auth/reset-password", body: `{"token":"","new_password":"new-password-123"}`, statusCode: http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := testutil.NewJSONRequest(t, http.MethodPost, tc.path, strings.NewReader(tc.body))
			res := httptest.NewRecorder()

			switch tc.path {
			case "/auth/register":
				handler.Register(res, req)
			case "/auth/login":
				handler.Login(res, req)
			case "/auth/forgot-password":
				handler.ForgotPassword(res, req)
			case "/auth/reset-password":
				handler.ResetPassword(res, req)
			}

			require.Equal(t, tc.statusCode, res.Code)
		})
	}
}
