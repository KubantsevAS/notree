ALTER TABLE users
    ADD COLUMN IF NOT EXISTS verification_token TEXT,
    ADD COLUMN IF NOT EXISTS verification_token_expires_at TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS reset_password_token TEXT,
    ADD COLUMN IF NOT EXISTS reset_password_token_expires_at TIMESTAMP WITH TIME ZONE;