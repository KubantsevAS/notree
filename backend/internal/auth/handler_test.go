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
	authDb "github.com/KubantsevAS/notree/backend/internal/db/auth"
	userDb "github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/http/middleware"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func newAuthHandlerWithFakes(store *authStoreFake, userStore *userStoreFake, mailer *fakeMailer) *auth.Handler {
	return auth.NewHandler(newService(store, userStore, mailer))
}

func TestHandlerRegister(t *testing.T) {
	testUserID := testutil.UUIDFromString(testutil.UUID1)
	userStore := &userStoreFake{
		getUserByEmailErr: sql.ErrNoRows,
		createUserResult:  testUserID,
	}
	handler := newAuthHandlerWithFakes(&authStoreFake{}, userStore, nil)

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
	handler := newAuthHandlerWithFakes(&authStoreFake{}, userStore, nil)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/login", &auth.LoginRequest{
		Email:    "user@example.com",
		Password: "password123",
	})
	res := httptest.NewRecorder()

	handler.Login(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	testutil.AssertAuthCookies(t, res, middleware.CookieAccessToken, middleware.CookieRefreshToken)
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
	handler := newAuthHandlerWithFakes(store, &userStoreFake{}, nil)

	req := testutil.NewJSONRequest(t, http.MethodPost, "/auth/refresh-tokens", nil)
	req.AddCookie(&http.Cookie{Name: middleware.CookieRefreshToken, Value: "refresh-token-value"})
	res := httptest.NewRecorder()

	handler.RefreshTokens(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	testutil.AssertAuthCookies(t, res, middleware.CookieAccessToken, middleware.CookieRefreshToken)
	require.Len(t, store.deleteRefreshTokenArg, 1)
	require.Equal(t, "refresh-token-value", store.deleteRefreshTokenArg[0])
}

func TestHandlerLogout_ClearsCookies(t *testing.T) {
	handler := newAuthHandlerWithFakes(&authStoreFake{}, &userStoreFake{}, nil)

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
	handler := newAuthHandlerWithFakes(&authStoreFake{}, &userStoreFake{}, nil)

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
	handler := newAuthHandlerWithFakes(&authStoreFake{}, userStore, mailer)

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
	handler := newAuthHandlerWithFakes(&authStoreFake{}, userStore, nil)

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

type handlerFunc func(*auth.Handler, http.ResponseWriter, *http.Request)

func TestHandler_Errors(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	require.NoError(t, err)
	existingUser := userDb.User{ID: testutil.UUIDFromString(testutil.UUID1), Email: "user@example.com", PasswordHash: string(hash)}

	tests := []struct {
		name          string
		call          handlerFunc
		body          string
		refreshCookie string
		store         *authStoreFake
		userStore     *userStoreFake
		wantStatus    int
		wantMsg       string
	}{
		// Request validation
		{
			name: "register invalid email", call: (*auth.Handler).Register,
			body:       `{"email":"not-an-email","password":"password123"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "login missing password", call: (*auth.Handler).Login,
			body:       `{"email":"user@example.com"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "forgot password invalid email", call: (*auth.Handler).ForgotPassword,
			body:       `{"email":"nope"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "reset password empty token", call: (*auth.Handler).ResetPassword,
			body:       `{"token":"","new_password":"new-password-123"}`,
			wantStatus: http.StatusBadRequest,
		},

		// Register
		{
			name: "register user exists", call: (*auth.Handler).Register,
			body:       `{"email":"user@example.com","password":"password123"}`,
			userStore:  &userStoreFake{getUserByEmailResult: existingUser},
			wantStatus: http.StatusConflict, wantMsg: auth.ErrUserExist.Error(),
		},
		{
			name: "register db error", call: (*auth.Handler).Register,
			body:       `{"email":"new@example.com","password":"password123"}`,
			userStore:  &userStoreFake{getUserByEmailErr: sql.ErrNoRows, createUserErr: errors.New("db down")},
			wantStatus: http.StatusInternalServerError, wantMsg: "internal server error",
		},

		// Login
		{
			name: "login wrong password", call: (*auth.Handler).Login,
			body:       `{"email":"user@example.com","password":"wrong-password"}`,
			userStore:  &userStoreFake{getUserByEmailResult: existingUser},
			wantStatus: http.StatusUnauthorized, wantMsg: auth.ErrWrongCredentials.Error(),
		},

		// RefreshTokens
		{
			name: "refresh missing cookie", call: (*auth.Handler).RefreshTokens,
			wantStatus: http.StatusUnauthorized, wantMsg: "missing refresh token",
		},
		{
			name: "refresh unknown token", call: (*auth.Handler).RefreshTokens,
			refreshCookie: "bad-token",
			store:         &authStoreFake{getRefreshTokenErr: sql.ErrNoRows},
			wantStatus:    http.StatusUnauthorized, wantMsg: auth.ErrInvalidRefreshToken.Error(),
		},

		// ResetPassword
		{
			name: "reset password unknown token", call: (*auth.Handler).ResetPassword,
			body:       `{"token":"bad-token","new_password":"new-password-123"}`,
			userStore:  &userStoreFake{getUserIdByResetPasswordErr: sql.ErrNoRows},
			wantStatus: http.StatusBadRequest, wantMsg: "invalid or expired token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store, userStore := tc.store, tc.userStore
			if store == nil {
				store = &authStoreFake{}
			}
			if userStore == nil {
				userStore = &userStoreFake{}
			}
			handler := newAuthHandlerWithFakes(store, userStore, nil)

			req := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			if tc.refreshCookie != "" {
				req.AddCookie(&http.Cookie{Name: middleware.CookieRefreshToken, Value: tc.refreshCookie})
			}
			res := httptest.NewRecorder()

			tc.call(handler, res, req)

			require.Equal(t, tc.wantStatus, res.Code)
			if tc.wantMsg != "" {
				testutil.AssertErrorJSON(t, res, tc.wantMsg)
			}
		})
	}
}
