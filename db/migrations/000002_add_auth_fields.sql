-- Migration: 000002_add_auth_fields
-- Adds authentication columns to the users table.
-- NOTE: DO NOT edit 000001_create_users.sql — always add new migrations.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email         TEXT        NOT NULL DEFAULT '' UNIQUE,
    ADD COLUMN IF NOT EXISTS password_hash TEXT        NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS role          TEXT        NOT NULL DEFAULT 'user',
    ADD COLUMN IF NOT EXISTS created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Index for fast email lookups on login
CREATE UNIQUE INDEX IF NOT EXISTS users_email_idx ON users (email);
