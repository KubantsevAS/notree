package middleware

type contextKey string

const UserIDKey contextKey = "user_id"
const CookieAccessToken = "access_token"
const CookieRefreshToken = "refresh_token"
