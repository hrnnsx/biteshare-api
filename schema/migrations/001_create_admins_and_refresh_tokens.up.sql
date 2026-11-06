-- Create the admins table used by the login flow.
CREATE TABLE IF NOT EXISTS admins (
    admin_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL
);

-- Store only hashed refresh tokens, never the original token value.
CREATE TABLE IF NOT EXISTS admin_refresh_tokens (
    token_id SERIAL PRIMARY KEY,
    admin_id INTEGER NOT NULL REFERENCES admins(admin_id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Speed up refresh-token lookups by admin and expiration time.
CREATE INDEX IF NOT EXISTS idx_admin_refresh_tokens_admin_id
    ON admin_refresh_tokens(admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_refresh_tokens_expires_at
    ON admin_refresh_tokens(expires_at);