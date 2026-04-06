-- TeachHub Database Schema (Complete — includes all migrations)

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ─── Admin (Teachers) ───────────────────────────────────
CREATE TABLE IF NOT EXISTS admin (
    id                  SERIAL PRIMARY KEY,
    username            TEXT NOT NULL UNIQUE,
    password            TEXT NOT NULL,
    email               TEXT NOT NULL DEFAULT '',
    school_name         TEXT NOT NULL DEFAULT '',
    subscription_status TEXT NOT NULL DEFAULT 'active' CHECK (subscription_status IN ('active', 'expired', 'suspended')),
    subscription_start  TIMESTAMPTZ,
    subscription_end    TIMESTAMPTZ,
    created_by_platform BOOLEAN NOT NULL DEFAULT false,
    application_id      INT,
    pending_password    TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Platform Administrators ────────────────────────────
CREATE TABLE IF NOT EXISTS platform_admin (
    id            SERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Teacher Applications ───────────────────────────────
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

-- FK for application_id (after teacher_application exists)
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'admin_application_id_fkey') THEN
        ALTER TABLE admin ADD CONSTRAINT admin_application_id_fkey
            FOREIGN KEY (application_id) REFERENCES teacher_application(id);
    END IF;
END $$;

-- ─── Payments ───────────────────────────────────────────
CREATE TABLE IF NOT EXISTS payment (
    id          SERIAL PRIMARY KEY,
    teacher_id  INT NOT NULL REFERENCES admin(id) ON DELETE CASCADE,
    amount      NUMERIC(10,2) NOT NULL,
    method      TEXT NOT NULL DEFAULT 'cash' CHECK (method IN ('cash', 'ccp', 'baridi_mob', 'other')),
    reference   TEXT NOT NULL DEFAULT '',
    notes       TEXT NOT NULL DEFAULT '',
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Classrooms ─────────────────────────────────────────
CREATE TABLE IF NOT EXISTS classroom (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    join_code  TEXT NOT NULL UNIQUE DEFAULT encode(gen_random_bytes(4), 'hex'),
    admin_id   INT NOT NULL REFERENCES admin(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- teacher profile picture per classroom
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='classroom' AND column_name='teacher_pic') THEN
        ALTER TABLE classroom ADD COLUMN teacher_pic TEXT NOT NULL DEFAULT '';
    END IF;
END $$;

-- ─── Students ───────────────────────────────────────────
CREATE TABLE IF NOT EXISTS student (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    email      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- phone column (migration-safe)
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='student' AND column_name='phone') THEN
        ALTER TABLE student ADD COLUMN phone TEXT NOT NULL DEFAULT '';
    END IF;
END $$;

-- ─── Student <-> Classroom membership ───────────────────
CREATE TABLE IF NOT EXISTS classroom_student (
    classroom_id INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    student_id   INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'approved' CHECK (status IN ('approved', 'pending', 'rejected')),
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    parent_code  TEXT UNIQUE DEFAULT encode(gen_random_bytes(6), 'hex'),
    PRIMARY KEY (classroom_id, student_id)
);

-- Add parent_code to existing rows (migration-safe)
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='classroom_student' AND column_name='parent_code') THEN
        ALTER TABLE classroom_student ADD COLUMN parent_code TEXT UNIQUE DEFAULT encode(gen_random_bytes(6), 'hex');
    END IF;
END $$;

-- Backfill parent_code for any rows that got NULL
UPDATE classroom_student SET parent_code = encode(gen_random_bytes(6), 'hex') WHERE parent_code IS NULL;

-- ─── Pre-registered allowed students ────────────────────
CREATE TABLE IF NOT EXISTS allowed_student (
    id           SERIAL PRIMARY KEY,
    classroom_id INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    email        TEXT NOT NULL,
    name         TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(classroom_id, email)
);

-- ─── Resource categories ────────────────────────────────
CREATE TABLE IF NOT EXISTS category (
    id           SERIAL PRIMARY KEY,
    classroom_id INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    sort_order   INT NOT NULL DEFAULT 0
);

-- ─── Resources ──────────────────────────────────────────
CREATE TABLE IF NOT EXISTS resource (
    id           SERIAL PRIMARY KEY,
    classroom_id INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    category_id  INT REFERENCES category(id) ON DELETE SET NULL,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    file_path    TEXT,
    file_type    TEXT,
    external_url TEXT,
    file_size    BIGINT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Assignments ────────────────────────────────────────
CREATE TABLE IF NOT EXISTS assignment (
    id             SERIAL PRIMARY KEY,
    classroom_id   INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    title          TEXT NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    deadline       TIMESTAMPTZ,
    response_type  TEXT NOT NULL DEFAULT 'file' CHECK (response_type IN ('file', 'text', 'both')),
    max_chars      INT NOT NULL DEFAULT 0,
    max_file_size  BIGINT NOT NULL DEFAULT 10485760,
    max_grade      DECIMAL(5,2) NOT NULL DEFAULT 20,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Assignment file attachment (teacher uploads a file with the assignment)
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='assignment' AND column_name='file_path') THEN
        ALTER TABLE assignment ADD COLUMN file_path TEXT NOT NULL DEFAULT '';
    END IF;
END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='assignment' AND column_name='file_name') THEN
        ALTER TABLE assignment ADD COLUMN file_name TEXT NOT NULL DEFAULT '';
    END IF;
END $$;

-- ─── Submissions ────────────────────────────────────────
CREATE TABLE IF NOT EXISTS submission (
    id            SERIAL PRIMARY KEY,
    assignment_id INT NOT NULL REFERENCES assignment(id) ON DELETE CASCADE,
    student_id    INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    file_path     TEXT NOT NULL DEFAULT '',
    file_name     TEXT NOT NULL DEFAULT '',
    file_size     BIGINT NOT NULL DEFAULT 0,
    text_content  TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'submitted' CHECK (status IN ('submitted', 'reviewed', 'needs_revision')),
    feedback      TEXT NOT NULL DEFAULT '',
    grade         DECIMAL(5,2),
    max_grade     DECIMAL(5,2),
    graded_at     TIMESTAMPTZ,
    submitted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Quizzes ────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS quiz (
    id                 SERIAL PRIMARY KEY,
    classroom_id       INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    title              TEXT NOT NULL,
    description        TEXT NOT NULL DEFAULT '',
    published          BOOLEAN NOT NULL DEFAULT FALSE,
    deadline           TIMESTAMPTZ,
    time_limit_minutes INT NOT NULL DEFAULT 0,
    max_attempts       INT NOT NULL DEFAULT 1,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Quiz questions ─────────────────────────────────────
CREATE TABLE IF NOT EXISTS quiz_question (
    id             SERIAL PRIMARY KEY,
    quiz_id        INT NOT NULL REFERENCES quiz(id) ON DELETE CASCADE,
    sort_order     INT NOT NULL DEFAULT 0,
    question_type  TEXT NOT NULL CHECK (question_type IN ('mcq', 'true_false', 'fill_blank', 'open_ended', 'file_upload')),
    content        TEXT NOT NULL,
    options        JSONB,
    correct_answer TEXT,
    points         INT NOT NULL DEFAULT 1
);

-- ─── Quiz attempts ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS quiz_attempt (
    id           SERIAL PRIMARY KEY,
    quiz_id      INT NOT NULL REFERENCES quiz(id) ON DELETE CASCADE,
    student_id   INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    answers      JSONB NOT NULL DEFAULT '{}',
    score        INT,
    max_score    INT,
    reviewed     BOOLEAN NOT NULL DEFAULT FALSE,
    file_answers JSONB NOT NULL DEFAULT '{}',
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at  TIMESTAMPTZ
);

-- ─── Live sessions ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS live_session (
    id               SERIAL PRIMARY KEY,
    classroom_id     INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    room_name        TEXT NOT NULL UNIQUE,
    active           BOOLEAN NOT NULL DEFAULT TRUE,
    ended_at         TIMESTAMPTZ,
    duration_minutes INT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Live session attendance ────────────────────────────
CREATE TABLE IF NOT EXISTS live_attendance (
    id              SERIAL PRIMARY KEY,
    live_session_id INT NOT NULL REFERENCES live_session(id) ON DELETE CASCADE,
    student_id      INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at         TIMESTAMPTZ
);

-- ─── Student remarks ────────────────────────────────────
CREATE TABLE IF NOT EXISTS student_remark (
    id           SERIAL PRIMARY KEY,
    classroom_id INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    student_id   INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    content      TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Resource view tracking ─────────────────────────────
CREATE TABLE IF NOT EXISTS resource_view (
    id          SERIAL PRIMARY KEY,
    resource_id INT NOT NULL REFERENCES resource(id) ON DELETE CASCADE,
    student_id  INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    viewed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Login tracking (migration-safe) ────────────────────
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='admin' AND column_name='last_login_at') THEN
        ALTER TABLE admin ADD COLUMN last_login_at TIMESTAMPTZ;
    END IF;
END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='admin' AND column_name='last_login_ip') THEN
        ALTER TABLE admin ADD COLUMN last_login_ip TEXT NOT NULL DEFAULT '';
    END IF;
END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='student' AND column_name='last_login_at') THEN
        ALTER TABLE student ADD COLUMN last_login_at TIMESTAMPTZ;
    END IF;
END $$;
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='student' AND column_name='last_login_ip') THEN
        ALTER TABLE student ADD COLUMN last_login_ip TEXT NOT NULL DEFAULT '';
    END IF;
END $$;

-- ─── Indexes ────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_payment_teacher ON payment(teacher_id);
CREATE INDEX IF NOT EXISTS idx_classroom_admin_id ON classroom(admin_id);
CREATE INDEX IF NOT EXISTS idx_student_remark_student ON student_remark(student_id);
CREATE INDEX IF NOT EXISTS idx_student_remark_classroom ON student_remark(classroom_id);
CREATE INDEX IF NOT EXISTS idx_resource_view_resource ON resource_view(resource_id);
CREATE INDEX IF NOT EXISTS idx_resource_view_student ON resource_view(student_id);
CREATE INDEX IF NOT EXISTS idx_classroom_student_student ON classroom_student(student_id);
CREATE INDEX IF NOT EXISTS idx_classroom_student_status ON classroom_student(status);
CREATE INDEX IF NOT EXISTS idx_allowed_student_classroom ON allowed_student(classroom_id);
CREATE INDEX IF NOT EXISTS idx_allowed_student_email ON allowed_student(email);
CREATE INDEX IF NOT EXISTS idx_resource_classroom ON resource(classroom_id);
CREATE INDEX IF NOT EXISTS idx_resource_category ON resource(category_id);
CREATE INDEX IF NOT EXISTS idx_assignment_classroom ON assignment(classroom_id);
CREATE INDEX IF NOT EXISTS idx_submission_assignment ON submission(assignment_id);
CREATE INDEX IF NOT EXISTS idx_submission_student ON submission(student_id);
CREATE INDEX IF NOT EXISTS idx_submission_assign_student ON submission(assignment_id, student_id);
CREATE INDEX IF NOT EXISTS idx_quiz_classroom ON quiz(classroom_id);
CREATE INDEX IF NOT EXISTS idx_quiz_question_quiz ON quiz_question(quiz_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempt_quiz ON quiz_attempt(quiz_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempt_student ON quiz_attempt(student_id);
CREATE INDEX IF NOT EXISTS idx_live_session_classroom ON live_session(classroom_id);
CREATE INDEX IF NOT EXISTS idx_live_session_active ON live_session(active);
CREATE INDEX IF NOT EXISTS idx_live_attendance_session ON live_attendance(live_session_id);
CREATE INDEX IF NOT EXISTS idx_live_attendance_student ON live_attendance(student_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempt_quiz_student_finished ON quiz_attempt(quiz_id, student_id, finished_at);
CREATE INDEX IF NOT EXISTS idx_cs_parent_code ON classroom_student(parent_code);
