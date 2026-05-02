-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES ($1, $2)
RETURNING id;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserById :one
SELECT * FROM users_public
WHERE id = $1
LIMIT 1;

-- name: UpdateUserProfile :one
UPDATE users
SET 
    username = COALESCE(sqlc.narg('username'), username),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    updated_at = NOW()
WHERE id = @id
RETURNING username, avatar_url, updated_at;

-- name: UpdateUserPreferences :one
UPDATE users
SET 
    locale = COALESCE(sqlc.narg('locale'), locale),
    timezone = COALESCE(sqlc.narg('timezone'), timezone),
    preferences = COALESCE(
        preferences || sqlc.narg('preferences')::jsonb, 
        preferences
    ),
    updated_at = NOW()
WHERE id = @id
RETURNING locale, timezone, preferences, updated_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET
    password_hash = $1, reset_password_token = NULL, reset_password_token_expires_at = NULL
WHERE id = $2;

-- name: GetUserPasswordHashById :one
SELECT password_hash FROM users 
WHERE id = $1 AND deleted_at IS NULL;

-- name: SetResetPasswordToken :exec
UPDATE users
SET 
    reset_password_token = $1, 
    reset_password_token_expires_at = NOW() + INTERVAL '15 minutes'
WHERE id = $2;

-- name: GetUserIdByResetPasswordToken :one
SELECT id FROM users
WHERE reset_password_token = $1
    AND reset_password_token_expires_at > NOW()
    AND deleted_at IS NULL;

-- name: SetVerificationToken :exec
UPDATE users
SET 
    verification_token = $1, 
    verification_token_expires_at = NOW() + INTERVAL '15 minutes'
WHERE id = $2;

-- name: VerifyEmailByCode :one
UPDATE users
SET 
    is_email_verified = true, 
    verification_token = NULL, 
    verification_token_expires_at = NULL
WHERE id = $1 
  AND verification_token = $2 
  AND verification_token_expires_at > NOW()
RETURNING id;