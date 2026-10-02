package user_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	userDb "github.com/KubantsevAS/notree/backend/internal/db/user"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/KubantsevAS/notree/backend/internal/user"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func newUserHandlerWithFakes(store *userStoreFake, mailer *fakeVerificationMailer) *user.Handler {
	service := user.NewService(store, mailer)
	return user.NewHandler(service)
}

func TestHandlerGetProfile(t *testing.T) {
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

	req := testutil.WithUserID(testutil.NewJSONRequest(t, http.MethodGet, "/profile/me", nil), userID)
	res := httptest.NewRecorder()

	handler.GetProfile(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	var payload user.GetProfileResponse
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, userID.String(), payload.ID)
	require.Equal(t, "alice@example.com", payload.Email)
}

func TestHandlerUpdateProfile(t *testing.T) {
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

	req := testutil.WithUserID(testutil.NewJSONRequest(t, http.MethodPatch, "/profile/me", user.UpdateUserProfileRequest{
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

func TestHandlerUpdatePreferences(t *testing.T) {
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

	req := testutil.WithUserID(testutil.NewJSONRequest(t, http.MethodPatch, "/profile/me/preference", user.UpdateUserPreferencesRequest{
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

func TestHandlerChangePassword(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("current-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	store := &userStoreFake{
		getUserPasswordHashResult: string(passwordHash),
	}
	handler := newUserHandlerWithFakes(store, nil)

	req := testutil.WithUserID(testutil.NewJSONRequest(t, http.MethodPatch, "/profile/me/change-password", user.ChangePasswordRequest{
		OldPassword: "current-password",
		NewPassword: "new-password-123",
	}), userID)
	res := httptest.NewRecorder()

	handler.ChangePassword(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	testutil.AssertMessageJSON(t, res, "password updated")
	require.Len(t, store.updateUserPasswordParams, 1)
}

func TestHandlerSendVerificationToken(t *testing.T) {
	mailer := &fakeVerificationMailer{sent: make(chan string, 1)}
	store := &userStoreFake{
		getUserByIdResult: &userDb.UsersPublic{
			ID:              userID,
			Email:           "alice@example.com",
			IsEmailVerified: testutil.PgBool(testutil.BoolPtr(false)),
		},
	}
	handler := newUserHandlerWithFakes(store, mailer)

	req := testutil.WithUserID(testutil.NewJSONRequest(t, http.MethodPost, "/profile/me/send-verification", nil), userID)
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

func TestHandlerVerifyEmailByToken(t *testing.T) {
	store := &userStoreFake{verifyEmailByTokenResult: userID}
	handler := newUserHandlerWithFakes(store, nil)

	req := testutil.WithUserID(testutil.NewJSONRequest(t, http.MethodPost, "/profile/me/verify-email", user.VerifyEmailByTokenRequest{
		Token: "valid-token",
	}), userID)
	res := httptest.NewRecorder()

	handler.VerifyEmailByToken(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	testutil.AssertMessageJSON(t, res, "email successfully verified")
	require.Len(t, store.verifyEmailByTokenParams, 1)
}

type handlerFunc func(*user.Handler, http.ResponseWriter, *http.Request)

func TestHandler_Errors(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("current-password"), bcrypt.MinCost)
	require.NoError(t, err)
	dbErr := errors.New("db down")
	unauthorized := "User ID not found in context"

	updateProfileBody := user.UpdateUserProfileRequest{Username: testutil.StringPtr("new-name")}
	changePasswordBody := user.ChangePasswordRequest{OldPassword: "current-password", NewPassword: "new-password-123"}

	tests := []struct {
		name       string
		call       handlerFunc
		method     string
		body       any
		anonymous  bool
		store      *userStoreFake
		wantStatus int
		wantMsg    string
	}{
		{name: "get profile unauthorized", call: (*user.Handler).GetProfile, method: http.MethodGet, anonymous: true, wantStatus: http.StatusUnauthorized, wantMsg: unauthorized},
		{name: "update profile unauthorized", call: (*user.Handler).UpdateProfile, method: http.MethodPatch, body: updateProfileBody, anonymous: true, wantStatus: http.StatusUnauthorized, wantMsg: unauthorized},
		{name: "update preferences unauthorized", call: (*user.Handler).UpdatePreferences, method: http.MethodPatch, body: user.UpdateUserPreferencesRequest{Locale: testutil.StringPtr("a")}, anonymous: true, wantStatus: http.StatusUnauthorized, wantMsg: unauthorized},
		{name: "change password unauthorized", call: (*user.Handler).ChangePassword, method: http.MethodPatch, body: changePasswordBody, anonymous: true, wantStatus: http.StatusUnauthorized, wantMsg: unauthorized},
		{name: "send verification unauthorized", call: (*user.Handler).SendVerificationToken, method: http.MethodPost, anonymous: true, wantStatus: http.StatusUnauthorized, wantMsg: unauthorized},
		{name: "verify email unauthorized", call: (*user.Handler).VerifyEmailByToken, method: http.MethodPost, body: user.VerifyEmailByTokenRequest{Token: "a"}, anonymous: true, wantStatus: http.StatusUnauthorized, wantMsg: unauthorized},
		{
			name: "get profile user not found", call: (*user.Handler).GetProfile, method: http.MethodGet,
			store:      &userStoreFake{getUserByIdErr: sql.ErrNoRows},
			wantStatus: http.StatusNotFound, wantMsg: "user not found",
		},
		{
			name: "get profile db error", call: (*user.Handler).GetProfile, method: http.MethodGet,
			store:      &userStoreFake{getUserByIdErr: dbErr},
			wantStatus: http.StatusInternalServerError, wantMsg: "internal server error",
		},
		{
			name: "update profile empty payload", call: (*user.Handler).UpdateProfile, method: http.MethodPatch,
			body:       map[string]any{},
			wantStatus: http.StatusBadRequest, wantMsg: "no fields provided for update",
		},
		{
			name: "update profile db error", call: (*user.Handler).UpdateProfile, method: http.MethodPatch,
			body: updateProfileBody, store: &userStoreFake{updateUserProfileErr: dbErr},
			wantStatus: http.StatusInternalServerError, wantMsg: "internal server error",
		},
		{
			name: "change password wrong old password", call: (*user.Handler).ChangePassword, method: http.MethodPatch,
			body:       user.ChangePasswordRequest{OldPassword: "wrong-password", NewPassword: "new-password-123"},
			store:      &userStoreFake{getUserPasswordHashResult: string(hash)},
			wantStatus: http.StatusUnauthorized, wantMsg: "wrong old password",
		},
		{
			name: "change password db error", call: (*user.Handler).ChangePassword, method: http.MethodPatch,
			body:       changePasswordBody,
			store:      &userStoreFake{getUserPasswordHashResult: string(hash), updateUserPasswordErr: dbErr},
			wantStatus: http.StatusInternalServerError, wantMsg: "internal server error",
		},
		{
			name: "verify email invalid token", call: (*user.Handler).VerifyEmailByToken, method: http.MethodPost,
			body:       user.VerifyEmailByTokenRequest{Token: "invalid-token"},
			store:      &userStoreFake{verifyEmailByTokenErr: sql.ErrNoRows},
			wantStatus: http.StatusBadRequest, wantMsg: "invalid or expired token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := tc.store
			if store == nil {
				store = &userStoreFake{}
			}
			handler := newUserHandlerWithFakes(store, nil)

			req := testutil.NewJSONRequest(t, tc.method, "/profile/me", tc.body)
			if !tc.anonymous {
				req = testutil.WithUserID(req, userID)
			}
			res := httptest.NewRecorder()

			tc.call(handler, res, req)

			require.Equal(t, tc.wantStatus, res.Code)
			testutil.AssertErrorJSON(t, res, tc.wantMsg)
		})
	}
}
