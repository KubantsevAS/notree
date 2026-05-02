ALTER TABLE users
    DROP COLUMN IF EXISTS verification_token,
    DROP COLUMN IF EXISTS verification_token_expires_at,
    DROP COLUMN IF EXISTS reset_password_token,
    DROP COLUMN IF EXISTS reset_password_token_expires_at;