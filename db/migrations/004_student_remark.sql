-- Phase 4: Student remarks table
CREATE TABLE IF NOT EXISTS student_remark (
    id           SERIAL PRIMARY KEY,
    classroom_id INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    student_id   INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    content      TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_student_remark_student ON student_remark(student_id);
CREATE INDEX IF NOT EXISTS idx_student_remark_classroom ON student_remark(classroom_id);
