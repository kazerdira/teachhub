-- Migration 009: Add pending_password to admin table
-- Stores the plaintext password temporarily so the platform owner can re-view it
-- until the teacher logs in for the first time.

ALTER TABLE admin ADD COLUMN IF NOT EXISTS pending_password TEXT;
