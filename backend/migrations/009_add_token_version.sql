-- Add token_version to users table for JWT invalidation.
-- Incrementing token_version instantly invalidates all existing JWTs for that user.
ALTER TABLE users ADD COLUMN token_version INTEGER NOT NULL DEFAULT 1;
