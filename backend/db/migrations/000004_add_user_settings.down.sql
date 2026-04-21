ALTER TABLE users
    DROP COLUMN IF EXISTS avatar_url,
    DROP COLUMN IF EXISTS timezone,
    DROP COLUMN IF EXISTS locale,
    DROP COLUMN IF EXISTS preferences,
    DROP COLUMN IF EXISTS is_email_verified,
    DROP COLUMN IF EXISTS last_login_at,
    DROP COLUMN IF EXISTS deleted_at;