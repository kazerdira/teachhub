-- Migration 012: Add display_name to admin for proper human names
ALTER TABLE admin ADD COLUMN IF NOT EXISTS display_name TEXT NOT NULL DEFAULT '';

-- Backfill: use username as display_name where empty
UPDATE admin SET display_name = username WHERE display_name = '';
