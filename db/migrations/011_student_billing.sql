-- db/migrations/011_student_billing.sql

-- Per-classroom session rate (teacher sets it, owner sees it)
ALTER TABLE classroom ADD COLUMN IF NOT EXISTS session_rate NUMERIC(10,2) NOT NULL DEFAULT 0;
ALTER TABLE classroom ADD COLUMN IF NOT EXISTS billing_enabled BOOLEAN NOT NULL DEFAULT false;

-- Per-student, per-month attendance-based invoice
CREATE TABLE IF NOT EXISTS student_invoice (
    id                SERIAL PRIMARY KEY,
    center_id         INT NOT NULL REFERENCES center(id) ON DELETE CASCADE,
    classroom_id      INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    student_id        INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    period_month      DATE NOT NULL,               -- YYYY-MM-01
    sessions_attended INT NOT NULL DEFAULT 0,
    rate_per_session  NUMERIC(10,2) NOT NULL,
    total_amount      NUMERIC(10,2) NOT NULL,
    status            TEXT NOT NULL DEFAULT 'unpaid'
                      CHECK (status IN ('unpaid','paid','cancelled')),
    paid_at           TIMESTAMPTZ,
    paid_method       TEXT DEFAULT '',             -- cash, card, ccp, etc.
    notes             TEXT NOT NULL DEFAULT '',
    generated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(classroom_id, student_id, period_month)
);

CREATE INDEX IF NOT EXISTS idx_inv_center_period ON student_invoice(center_id, period_month);
CREATE INDEX IF NOT EXISTS idx_inv_status ON student_invoice(status);
CREATE INDEX IF NOT EXISTS idx_inv_student ON student_invoice(student_id);

-- Parent view tracking (so owner can see "23 parents consulted this week")
CREATE TABLE IF NOT EXISTS parent_view_log (
    id          SERIAL PRIMARY KEY,
    parent_code TEXT NOT NULL,
    viewed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip          TEXT DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_pvl_code ON parent_view_log(parent_code);
CREATE INDEX IF NOT EXISTS idx_pvl_date ON parent_view_log(viewed_at);
