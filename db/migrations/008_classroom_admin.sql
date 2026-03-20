-- Add admin_id to classroom table to scope classrooms to specific teachers
ALTER TABLE classroom ADD COLUMN IF NOT EXISTS admin_id INT REFERENCES admin(id) ON DELETE CASCADE;

-- Set existing classrooms to the first admin (default teacher)
UPDATE classroom SET admin_id = (SELECT id FROM admin ORDER BY id LIMIT 1) WHERE admin_id IS NULL;

-- Make admin_id NOT NULL after backfill
ALTER TABLE classroom ALTER COLUMN admin_id SET NOT NULL;

-- Index for fast lookups
CREATE INDEX IF NOT EXISTS idx_classroom_admin_id ON classroom(admin_id);
