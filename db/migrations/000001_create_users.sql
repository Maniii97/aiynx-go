-- Migration: 000001_create_users
-- Creates the users table with name and date of birth.
-- NOTE: age is NEVER stored — it is calculated dynamically in the service layer.

CREATE TABLE IF NOT EXISTS users (
    id   BIGSERIAL    PRIMARY KEY,
    name TEXT         NOT NULL,
    dob  DATE         NOT NULL
);
