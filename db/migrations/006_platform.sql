-- Migration 006: Platform Admin layer
-- Adds platform owner system, teacher applications, and subscription management

-- Platform administrators (the business owners)
CREATE TABLE IF NOT EXISTS platform_admin (
    id            SERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Teacher applications (public registration)
CREATE TABLE IF NOT EXISTS teacher_application (
    id          SERIAL PRIMARY KEY,
    full_name   TEXT NOT NULL,
    email       TEXT NOT NULL,
    phone       TEXT NOT NULL DEFAULT '',
    school_name TEXT NOT NULL DEFAULT '',
    wilaya      TEXT NOT NULL DEFAULT '',
    message     TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'contacted')),
    admin_notes TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ
);

-- Extend admin table with subscription & contact info
ALTER TABLE admin ADD COLUMN IF NOT EXISTS email TEXT NOT NULL DEFAULT '';
ALTER TABLE admin ADD COLUMN IF NOT EXISTS school_name TEXT NOT NULL DEFAULT '';
ALTER TABLE admin ADD COLUMN IF NOT EXISTS subscription_status TEXT NOT NULL DEFAULT 'active' CHECK (subscription_status IN ('active', 'expired', 'suspended'));
ALTER TABLE admin ADD COLUMN IF NOT EXISTS subscription_start TIMESTAMPTZ;
ALTER TABLE admin ADD COLUMN IF NOT EXISTS subscription_end TIMESTAMPTZ;
ALTER TABLE admin ADD COLUMN IF NOT EXISTS created_by_platform BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE admin ADD COLUMN IF NOT EXISTS application_id INT REFERENCES teacher_application(id);
