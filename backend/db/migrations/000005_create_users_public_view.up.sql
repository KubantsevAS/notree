CREATE VIEW users_public AS
SELECT id, email, username, avatar_url, timezone, locale, preferences, is_email_verified, last_login_at, created_at, updated_at
FROM users
WHERE deleted_at IS NULL;