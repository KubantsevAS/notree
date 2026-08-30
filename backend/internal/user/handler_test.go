package user_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	userDb "github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/http/middleware"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/KubantsevAS/notree/backend/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type fakeVerificationMailer struct {
	sent chan string
}

func (f *fakeVerificationMailer) SendPasswordReset(_ context.Context, _ string, _ string) error {
	return nil
}

func (f *fakeVerificationMailer) SendVerificationEmail(_ context.Context, _ string, token string) error {
	if f.sent != nil {
		f.sent <- token
	}
	return nil
}

func newUserHandlerWithFakes(store *userStoreFake, mailer *fakeVerificationMailer) *user.Handler {
	service := user.NewService(store, mailer)
	return user.NewHandler(service)
}

func withUserContext(t *testing.T, req *http.Request, userID pgtype.UUID) *http.Request {
	t.Helper()
	return req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, userID.String()))
}

func TestUserHandlerGetProfileSuccess(t *testing.T) {
	now := time.Now()
	store := &userStoreFake{
		getUserByIdResult: &userDb.UsersPublic{
			ID:              userID,
			Email:           "alice@example.com",
			Username:        testutil.PgText(testutil.StringPtr("alice")),
			AvatarUrl:       testutil.PgText(testutil.StringPtr("https://example.com/avatar.png")),
			Timezone:        testutil.PgText(testutil.StringPtr("UTC")),
			Locale:          testutil.PgText(testutil.StringPtr("en-US")),
			Preferences:     json.RawMessage(`{"theme":"dark"}`),
			IsEmailVerified: testutil.PgBool(testutil.BoolPtr(true)),
			LastLoginAt:     testutil.PgTimestamptz(&now),
			CreatedAt:       testutil.PgTimestamptz(&now),
			UpdatedAt:       testutil.PgTimestamptz(&now),
		},
	}
	handler := newUserHandlerWithFakes(store, nil)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodGet, "/profile/me", nil), userID)
	res := httptest.NewRecorder()

	handler.GetProfile(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var payload user.GetProfileResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, userID.String(), payload.ID)
	require.Equal(t, "alice@example.com", payload.Email)
}

func TestUserHandlerGetProfileUnauthorized(t *testing.T) {
	handler := newUserHandlerWithFakes(&userStoreFake{}, nil)

	req := testutil.NewJSONRequest(t, http.MethodGet, "/profile/me", nil)
	res := httptest.NewRecorder()

	handler.GetProfile(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
	testutil.AssertErrorJSON(t, res, "User ID not found in context")
}

func TestUserHandlerGetProfileNotFound(t *testing.T) {
	handler := newUserHandlerWithFakes(&userStoreFake{getUserByIdErr: sql.ErrNoRows}, nil)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodGet, "/profile/me", nil), userID)
	res := httptest.NewRecorder()

	handler.GetProfile(res, req)

	require.Equal(t, http.StatusNotFound, res.Code)
	testutil.AssertErrorJSON(t, res, "user not found")
}

func TestUserHandlerGetProfileInternalError(t *testing.T) {
	store := &userStoreFake{getUserByIdErr: errors.New("db timeout")}
	handler := newUserHandlerWithFakes(store, nil)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodGet, "/profile/me", nil), userID)
	res := httptest.NewRecorder()

	handler.GetProfile(res, req)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	testutil.AssertErrorJSON(t, res, "internal server error")
}

func TestUserHandlerUpdateProfileSuccess(t *testing.T) {
	username := "new-name"
	avatarURL := "https://example.com/avatar.jpg"
	updatedAt := time.Now()
	store := &userStoreFake{
		updateUserProfileResult: &userDb.UpdateUserProfileRow{
			Username:  testutil.PgText(testutil.StringPtr(username)),
			AvatarUrl: testutil.PgText(testutil.StringPtr(avatarURL)),
			UpdatedAt: testutil.PgTimestamptz(&updatedAt),
		},
	}
	handler := newUserHandlerWithFakes(store, nil)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodPatch, "/profile/me", user.UpdateUserProfileRequest{
		Username:  testutil.StringPtr(username),
		AvatarUrl: testutil.StringPtr(avatarURL),
	}), userID)
	res := httptest.NewRecorder()

	handler.UpdateProfile(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var payload user.UpdateUserProfileResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, username, *payload.Username)
	require.Equal(t, avatarURL, *payload.AvatarUrl)
	require.Len(t, store.updateUserProfileParams, 1)
}

func TestUserHandlerUpdateProfileEmptyPayload(t *testing.T) {
	handler := newUserHandlerWithFakes(&userStoreFake{}, nil)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodPatch, "/profile/me", map[string]any{}), userID)
	res := httptest.NewRecorder()

	handler.UpdateProfile(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "no fields provided for update")
}

func TestUserHandlerUpdatePreferencesSuccess(t *testing.T) {
	locale := "ru-RU"
	timezone := "Europe/Moscow"
	preferences := json.RawMessage(`{"theme":"light"}`)
	updatedAt := time.Now()
	store := &userStoreFake{
		updateUserPreferencesResult: &userDb.UpdateUserPreferencesRow{
			Locale:      testutil.PgText(testutil.StringPtr(locale)),
			Timezone:    testutil.PgText(testutil.StringPtr(timezone)),
			Preferences: preferences,
			UpdatedAt:   testutil.PgTimestamptz(&updatedAt),
		},
	}
	handler := newUserHandlerWithFakes(store, nil)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodPatch, "/profile/me/preference", user.UpdateUserPreferencesRequest{
		Locale:      testutil.StringPtr(locale),
		Timezone:    testutil.StringPtr(timezone),
		Preferences: &preferences,
	}), userID)
	res := httptest.NewRecorder()

	handler.UpdatePreferences(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var payload user.UpdateUserPreferencesResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, locale, *payload.Locale)
	require.Equal(t, timezone, *payload.Timezone)
	require.Len(t, store.updateUserPreferencesParams, 1)
}

func TestUserHandlerChangePasswordSuccess(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("current-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	store := &userStoreFake{
		getUserPasswordHashResult: string(passwordHash),
	}
	handler := newUserHandlerWithFakes(store, nil)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodPatch, "/profile/me/change-password", user.ChangePasswordRequest{
		OldPassword: "current-password",
		NewPassword: "new-password-123",
	}), userID)
	res := httptest.NewRecorder()

	handler.ChangePassword(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	testutil.AssertMessageJSON(t, res, "password updated")
	require.Len(t, store.updateUserPasswordParams, 1)
}

func TestUserHandlerChangePasswordWrongOldPassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("current-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	store := &userStoreFake{getUserPasswordHashResult: string(passwordHash)}
	handler := newUserHandlerWithFakes(store, nil)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodPatch, "/profile/me/change-password", user.ChangePasswordRequest{
		OldPassword: "wrong-password",
		NewPassword: "new-password-123",
	}), userID)
	res := httptest.NewRecorder()

	handler.ChangePassword(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
	testutil.AssertErrorJSON(t, res, "wrong old password")
}

func TestUserHandlerChangePasswordInternalError(t *testing.T) {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("current-password"), bcrypt.DefaultCost)

	store := &userStoreFake{
		getUserPasswordHashResult: string(passwordHash),
		updateUserPasswordErr:     errors.New("connection lost"),
	}
	handler := newUserHandlerWithFakes(store, nil)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodPatch, "/profile/me/change-password", user.ChangePasswordRequest{
		OldPassword: "current-password",
		NewPassword: "new-password-123",
	}), userID)
	res := httptest.NewRecorder()

	handler.ChangePassword(res, req)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	testutil.AssertErrorJSON(t, res, "internal server error")
}

func TestUserHandlerSendVerificationTokenSuccess(t *testing.T) {
	mailer := &fakeVerificationMailer{sent: make(chan string, 1)}
	store := &userStoreFake{
		getUserByIdResult: &userDb.UsersPublic{
			ID:              userID,
			Email:           "alice@example.com",
			IsEmailVerified: testutil.PgBool(testutil.BoolPtr(false)),
		},
	}
	handler := newUserHandlerWithFakes(store, mailer)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodPost, "/profile/me/send-verification", nil), userID)
	res := httptest.NewRecorder()

	handler.SendVerificationToken(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	testutil.AssertMessageJSON(t, res, "email verification link has been sent")
	require.Len(t, store.setVerificationTokenParams, 1)

	select {
	case token := <-mailer.sent:
		require.NotEmpty(t, token)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected verification email to be sent")
	}
}

func TestUserHandlerVerifyEmailByTokenSuccess(t *testing.T) {
	store := &userStoreFake{verifyEmailByTokenResult: userID}
	handler := newUserHandlerWithFakes(store, nil)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodPost, "/profile/me/verify-email", user.VerifyEmailByTokenRequest{
		Token: "valid-token",
	}), userID)
	res := httptest.NewRecorder()

	handler.VerifyEmailByToken(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	testutil.AssertMessageJSON(t, res, "email successfully verified")
	require.Len(t, store.verifyEmailByTokenParams, 1)
}

func TestUserHandlerVerifyEmailByTokenInvalidToken(t *testing.T) {
	store := &userStoreFake{verifyEmailByTokenErr: sql.ErrNoRows}
	handler := newUserHandlerWithFakes(store, nil)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodPost, "/profile/me/verify-email", user.VerifyEmailByTokenRequest{
		Token: "invalid-token",
	}), userID)
	res := httptest.NewRecorder()

	handler.VerifyEmailByToken(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
	testutil.AssertErrorJSON(t, res, "invalid or expired token")
}

func TestUserHandlerRequiresAuthentication(t *testing.T) {
	handler := newUserHandlerWithFakes(&userStoreFake{}, nil)
	tests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{"GetProfile", http.MethodGet, "/profile/me", nil},
		{"UpdateProfile", http.MethodPatch, "/profile/me", user.UpdateUserProfileRequest{Username: testutil.StringPtr("a")}},
		{"UpdatePreferences", http.MethodPatch, "/profile/me/preference", user.UpdateUserPreferencesRequest{Locale: testutil.StringPtr("a")}},
		{"ChangePassword", http.MethodPatch, "/profile/me/change-password", user.ChangePasswordRequest{OldPassword: "a", NewPassword: "b"}},
		{"SendVerificationToken", http.MethodPost, "/profile/me/send-verification", nil},
		{"VerifyEmailByToken", http.MethodPost, "/profile/me/verify-email", user.VerifyEmailByTokenRequest{Token: "a"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := testutil.NewJSONRequest(t, tc.method, tc.path, tc.body)
			res := httptest.NewRecorder()

			switch tc.path {
			case "/profile/me":
				if tc.method == http.MethodGet {
					handler.GetProfile(res, req)
				} else {
					handler.UpdateProfile(res, req)
				}
			case "/profile/me/preference":
				handler.UpdatePreferences(res, req)
			case "/profile/me/change-password":
				handler.ChangePassword(res, req)
			case "/profile/me/send-verification":
				handler.SendVerificationToken(res, req)
			case "/profile/me/verify-email":
				handler.VerifyEmailByToken(res, req)
			}

			require.Equal(t, http.StatusUnauthorized, res.Code)
			testutil.AssertErrorJSON(t, res, "User ID not found in context")
		})
	}
}

func TestUserHandlerUpdateProfileInternalError(t *testing.T) {
	store := &userStoreFake{updateUserProfileErr: errors.New("db error")}
	handler := newUserHandlerWithFakes(store, nil)

	req := withUserContext(t, testutil.NewJSONRequest(t, http.MethodPatch, "/profile/me", user.UpdateUserProfileRequest{Username: testutil.StringPtr("new-name")}), userID)
	res := httptest.NewRecorder()

	handler.UpdateProfile(res, req)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	testutil.AssertErrorJSON(t, res, "internal server error")
}
