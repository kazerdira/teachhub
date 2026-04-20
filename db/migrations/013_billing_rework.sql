-- db/migrations/013_billing_rework.sql
-- Billing rework: remove student/parent billing, add platform-to-center invoicing.
-- B2B only: TeachHub invoices centers (not students/parents).
-- Idempotent: safe to re-run. Uses IF EXISTS / IF NOT EXISTS / DO $$ throughout.

-- ═══════════════════════════════════════════════════════
-- PART A — Drop student/parent billing artifacts
-- ═══════════════════════════════════════════════════════

-- 1. student_invoice (center-to-student invoicing — removed, B2B only)
DROP TABLE IF EXISTS student_invoice CASCADE;

-- 2. parent_view_log (parent report view tracking — removed, no parent analytics)
DROP TABLE IF EXISTS parent_view_log CASCADE;

-- 3. classroom billing columns
ALTER TABLE classroom DROP COLUMN IF EXISTS session_rate;
ALTER TABLE classroom DROP COLUMN IF EXISTS billing_enabled;

-- 4. center seat cap (no seat limits in per-teacher model)
ALTER TABLE center DROP COLUMN IF EXISTS seat_count;

-- ═══════════════════════════════════════════════════════
-- PART B — Platform-to-center billing
-- ═══════════════════════════════════════════════════════

-- 5. Rename center.price_per_seat → center.price_per_teacher
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'center' AND column_name = 'price_per_seat'
    ) THEN
        ALTER TABLE center RENAME COLUMN price_per_seat TO price_per_teacher;
    END IF;
END $$;

-- 6. Add center.currency (explicit per-center currency, never inferred at runtime)
ALTER TABLE center
    ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'DZD';

-- Backfill: existing centers in France get EUR
UPDATE center SET currency = 'EUR' WHERE country = 'FR' AND currency = 'DZD';

-- 7. Add center.billing_mode (enum-ready, only valid value is 'per_teacher')
ALTER TABLE center
    ADD COLUMN IF NOT EXISTS billing_mode TEXT NOT NULL DEFAULT 'per_teacher'
    CHECK (billing_mode IN ('per_teacher'));

-- 8. Add admin.billable_from
--    Set on first teacher login: NOW() + 30 days. Never reset.
--    Guard is enforced in SQL: UPDATE ... WHERE billable_from IS NULL.
ALTER TABLE admin
    ADD COLUMN IF NOT EXISTS billable_from TIMESTAMPTZ;

-- 9. Add admin.deactivated_at
--    Set to NOW() when owner deactivates a teacher.
--    Cleared to NULL when owner reactivates. Single value — not an audit log.
ALTER TABLE admin
    ADD COLUMN IF NOT EXISTS deactivated_at TIMESTAMPTZ;

-- 10. Platform-to-center invoice table
CREATE TABLE IF NOT EXISTS center_invoice (
    id                SERIAL PRIMARY KEY,
    center_id         INT NOT NULL REFERENCES center(id) ON DELETE CASCADE,
    period_month      DATE NOT NULL,                      -- always YYYY-MM-01 UTC
    teacher_count     INT NOT NULL DEFAULT 0,
    price_per_teacher NUMERIC(10,2) NOT NULL,             -- snapshot at generation time
    currency          TEXT NOT NULL,                      -- snapshot (no default — must be explicit)
    total_amount      NUMERIC(10,2) NOT NULL,             -- teacher_count × price_per_teacher
    status            TEXT NOT NULL DEFAULT 'unpaid'
                      CHECK (status IN ('unpaid', 'paid', 'cancelled')),
    paid_at           TIMESTAMPTZ,
    paid_method       TEXT NOT NULL DEFAULT '',           -- cash, ccp, virement, other
    paid_reference    TEXT NOT NULL DEFAULT '',
    generated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (center_id, period_month)
);

CREATE INDEX IF NOT EXISTS idx_center_invoice_center  ON center_invoice(center_id);
CREATE INDEX IF NOT EXISTS idx_center_invoice_status  ON center_invoice(status);
CREATE INDEX IF NOT EXISTS idx_center_invoice_period  ON center_invoice(period_month);
