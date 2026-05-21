-- Add token_version column to users table for token revocation mechanism
-- This allows invalidating all tokens for a user by incrementing their version

ALTER TABLE users ADD COLUMN IF NOT EXISTS token_version INTEGER NOT NULL DEFAULT 0;

-- Add comment for documentation
COMMENT ON COLUMN users.token_version IS 'Token version for revocation. Increment to invalidate all existing tokens';