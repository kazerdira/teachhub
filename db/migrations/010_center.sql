-- Migration 010: Center layer — center table, admin role/center_id, backfill
-- Run once against production, then merged into schema.sql for fresh installs.

-- ─── Center table ───────────────────────────────────────
CREATE TABLE IF NOT EXISTS center (
    id                  SERIAL PRIMARY KEY,
    name                TEXT NOT NULL,
    owner_admin_id      INT,
    address             TEXT NOT NULL DEFAULT '',
    city                TEXT NOT NULL DEFAULT '',
    country             TEXT NOT NULL DEFAULT 'DZ',
    phone               TEXT NOT NULL DEFAULT '',
    email               TEXT NOT NULL DEFAULT '',
    logo_path           TEXT NOT NULL DEFAULT '',
    subscription_status TEXT NOT NULL DEFAULT 'trial'
                        CHECK (subscription_status IN ('trial','active','expired','suspended','cancelled')),
    subscription_start  TIMESTAMPTZ,
    subscription_end    TIMESTAMPTZ,
    seat_count          INT NOT NULL DEFAULT 3,
    price_per_seat      NUMERIC(10,2) NOT NULL DEFAULT 0,
    trial_ends_at       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_center_owner ON center(owner_admin_id);
CREATE INDEX IF NOT EXISTS idx_center_status ON center(subscription_status);

-- ─── Extend admin with role + center scoping ────────────
ALTER TABLE admin ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'teacher'
    CHECK (role IN ('owner','teacher'));
ALTER TABLE admin ADD COLUMN IF NOT EXISTS center_id INT REFERENCES center(id) ON DELETE SET NULL;
ALTER TABLE admin ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT true;

CREATE INDEX IF NOT EXISTS idx_admin_center ON admin(center_id);
CREATE INDEX IF NOT EXISTS idx_admin_role ON admin(role);

-- ─── Extend teacher_application with center fields ──────
ALTER TABLE teacher_application ADD COLUMN IF NOT EXISTS center_name TEXT NOT NULL DEFAULT '';
ALTER TABLE teacher_application ADD COLUMN IF NOT EXISTS expected_teachers INT NOT NULL DEFAULT 1;
ALTER TABLE teacher_application ADD COLUMN IF NOT EXISTS expected_students INT NOT NULL DEFAULT 0;

-- ─── Extend payment with center_id ──────────────────────
ALTER TABLE payment ADD COLUMN IF NOT EXISTS center_id INT REFERENCES center(id) ON DELETE CASCADE;

-- ─── Backfill: every existing admin becomes a solo-center ─
DO $$
DECLARE
    a RECORD;
    new_center_id INT;
BEGIN
    FOR a IN SELECT id, school_name, username, email, country,
                    subscription_status, subscription_start, subscription_end
             FROM admin WHERE center_id IS NULL
    LOOP
        INSERT INTO center (name, owner_admin_id, email, country,
                            subscription_status, subscription_start, subscription_end,
                            seat_count)
        VALUES (
            COALESCE(NULLIF(a.school_name, ''), a.username || ' Center'),
            a.id,
            a.email,
            COALESCE(NULLIF(a.country, ''), 'DZ'),
            CASE WHEN a.subscription_status IN ('active','expired','suspended')
                 THEN a.subscription_status ELSE 'active' END,
            a.subscription_start,
            a.subscription_end,
            1
        )
        RETURNING id INTO new_center_id;

        UPDATE admin SET center_id = new_center_id, role = 'owner' WHERE id = a.id;

        -- Scope existing payments to the center
        UPDATE payment SET center_id = new_center_id
        WHERE teacher_id = a.id AND center_id IS NULL;
    END LOOP;
END $$;

-- Add FK from center.owner_admin_id → admin now that both exist
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'center_owner_admin_fkey') THEN
        ALTER TABLE center ADD CONSTRAINT center_owner_admin_fkey
            FOREIGN KEY (owner_admin_id) REFERENCES admin(id) ON DELETE SET NULL;
    END IF;
END $$;
