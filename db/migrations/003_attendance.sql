-- Phase 3: Live Session Attendance Tracking
-- Run this on existing databases. New installs use the updated schema.sql.

ALTER TABLE live_session ADD COLUMN IF NOT EXISTS ended_at TIMESTAMPTZ;
ALTER TABLE live_session ADD COLUMN IF NOT EXISTS duration_minutes INT;

CREATE TABLE IF NOT EXISTS live_attendance (
    id              SERIAL PRIMARY KEY,
    live_session_id INT NOT NULL REFERENCES live_session(id) ON DELETE CASCADE,
    student_id      INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at         TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_live_attendance_session ON live_attendance(live_session_id);
CREATE INDEX IF NOT EXISTS idx_live_attendance_student ON live_attendance(student_id);
