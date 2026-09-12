package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/http/httputil"
	"github.com/KubantsevAS/notree/backend/internal/http/middleware"
	"github.com/KubantsevAS/notree/backend/internal/http/middleware/auth"
	"github.com/KubantsevAS/notree/backend/internal/testutil"
	"github.com/KubantsevAS/notree/backend/pkg/jwt"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	nextShouldNotBeCalled := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	t.Run("missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		res := httptest.NewRecorder()

		auth.AuthMiddleware(testutil.TestSecret)(nextShouldNotBeCalled).ServeHTTP(res, req)

		require.Equal(t, http.StatusUnauthorized, res.Code)
		testutil.AssertErrorJSON(t, res, "missing token")
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req.AddCookie(&http.Cookie{Name: middleware.CookieAccessToken, Value: "invalid-token"})
		res := httptest.NewRecorder()

		auth.AuthMiddleware(testutil.TestSecret)(nextShouldNotBeCalled).ServeHTTP(res, req)

		require.Equal(t, http.StatusUnauthorized, res.Code)
		testutil.AssertErrorJSON(t, res, "invalid or expired access token")
	})

	t.Run("valid token sets user id in context", func(t *testing.T) {
		userID := testutil.UUIDFromString(testutil.UUID1)

		token, err := jwt.GenerateAccessToken(userID, testutil.TestSecret)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req.AddCookie(&http.Cookie{Name: middleware.CookieAccessToken, Value: token})
		res := httptest.NewRecorder()

		auth.AuthMiddleware(testutil.TestSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxUserID, err := httputil.GetUserIDFromCtx(r.Context())
			require.NoError(t, err)

			require.Equal(t, userID.String(), ctxUserID)

			w.WriteHeader(http.StatusNoContent)
		})).ServeHTTP(res, req)

		require.Equal(t, http.StatusNoContent, res.Code)
	})

	t.Run("empty token is treated as invalid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req.AddCookie(&http.Cookie{Name: middleware.CookieAccessToken, Value: ""})
		res := httptest.NewRecorder()

		auth.AuthMiddleware(testutil.TestSecret)(nextShouldNotBeCalled).ServeHTTP(res, req)

		require.Equal(t, http.StatusUnauthorized, res.Code)
		testutil.AssertErrorJSON(t, res, "invalid or expired access token")
	})
}
