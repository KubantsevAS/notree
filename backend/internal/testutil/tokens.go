package testutil

import (
	"net/http"
	"testing"

	"github.com/KubantsevAS/notree/backend/internal/http/middleware"
	pkgJwt "github.com/KubantsevAS/notree/backend/pkg/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func AccessTokenCookie(token string) http.Cookie {
	return http.Cookie{Name: middleware.CookieAccessToken, Value: token}
}

func MakeAccessToken(t *testing.T, secret string, userID string) string {
	t.Helper()
	token, err := pkgJwt.GenerateAccessToken(MustUUID(t, userID), secret)
	require.NoError(t, err)
	return token
}

func MakeTokenWithClaims(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}
