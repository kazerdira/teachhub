-- Phase 5: Resource view tracking
CREATE TABLE IF NOT EXISTS resource_view (
    id          SERIAL PRIMARY KEY,
    resource_id INT NOT NULL REFERENCES resource(id) ON DELETE CASCADE,
    student_id  INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    viewed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_resource_view_resource ON resource_view(resource_id);
CREATE INDEX IF NOT EXISTS idx_resource_view_student ON resource_view(student_id);
