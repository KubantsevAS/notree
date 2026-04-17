-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token_hash, user_id, expires_at)
VALUES ($1, $2, $3);

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token_hash = $1 LIMIT 1;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE token_hash = $1;

-- name: DeleteAllUserRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;