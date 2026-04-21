-- name: CreateUser :one
INSERT INTO users (email, password_hash, username)
VALUES ($1, $2, $3)
RETURNING *;

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
        sqlc.narg('preferences')::jsonb || preferences, 
        preferences
    ),
    updated_at = NOW()
WHERE id = @id
RETURNING locale, timezone, preferences, updated_at;