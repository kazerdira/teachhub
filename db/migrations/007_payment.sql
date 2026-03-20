-- Migration 007: Payment tracking for subscriptions
-- Manual payment logging suited for Algeria (cash, CCP, BaridiMob)

CREATE TABLE IF NOT EXISTS payment (
    id          SERIAL PRIMARY KEY,
    teacher_id  INT NOT NULL REFERENCES admin(id) ON DELETE CASCADE,
    amount      NUMERIC(10,2) NOT NULL,
    method      TEXT NOT NULL DEFAULT 'cash' CHECK (method IN ('cash', 'ccp', 'baridi_mob', 'other')),
    reference   TEXT NOT NULL DEFAULT '',
    notes       TEXT NOT NULL DEFAULT '',
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_teacher ON payment(teacher_id);
