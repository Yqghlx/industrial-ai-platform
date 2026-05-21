-- Remove token_version column from users table

ALTER TABLE users DROP COLUMN IF EXISTS token_version;