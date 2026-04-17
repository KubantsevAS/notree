package auth

import (
	"context"
	"net/http"

	"github.com/KubantsevAS/notree/backend/pkg/jwt"
)

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				http.Error(w, "Missing token", http.StatusUnauthorized)
				return
			}

			tokenString := cookie.Value

			userID, err := jwt.ParseAccessToken(tokenString, secret)
			if err != nil {
				http.Error(w, "Invalid or expired access token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
