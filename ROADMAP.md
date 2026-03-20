# TeachHub — Feature Roadmap

> **Current state:** TeachHub is a full-featured teaching platform with classrooms, resources, assignments (file + text submissions with numeric grading + deadlines), quizzes (MCQ, True/False, Fill-in-the-blank, Open-ended, File Upload with time limits + attempt limits), student access control, LiveKit live video classes with attendance tracking, and a teacher analytics dashboard.

This document outlines the plan to add **analytics, enhanced quizzes, grading, and tracking** in logical phases. Each phase builds on the previous one so nothing breaks mid-way.

---

## What We Have Today (v1.0 → v3.0)

| Area | Status | Notes |
|------|--------|-------|
| Classrooms | ✅ Complete | Create, delete, join codes, regenerate codes |
| Resources | ✅ Complete | Upload files, links, categories |
| Assignments | ✅ Complete | File/text/both responses, deadlines enforced, numeric grading, max grade |
| Quizzes | ✅ Complete | MCQ, T/F, Fill-blank, Open-ended, File Upload, time limits, max attempts, deadlines |
| Students | ✅ Complete | Join via link, allow-list, pending approval, remove |
| Live Classes | ✅ Complete | LiveKit video/audio, teacher controls, text chat, attendance tracking |
| Analytics | ✅ Complete | Quiz analytics, assignment analytics, student roster, student detail page, live session history |
| Deadlines | ✅ Complete | Enforced on both assignments and quizzes |

### Known Gaps in Current Code
- ~~`assignment.deadline` is stored but **never enforced** on submit~~ ✅ Fixed
- ~~`UpdateQuizQuestion` store function exists but **no handler**~~ ✅ Fixed
- ~~`StudentQuizResult` handler exists but **no route** registered~~ ✅ Fixed
- ~~Students can take the same quiz **multiple times** (no limit)~~ ✅ Fixed (max_attempts)
- ~~No attendance tracking for live sessions~~ ✅ Fixed
- ~~No numeric grades/marks on assignments (only text feedback + status)~~ ✅ Fixed

---

## Phase 1 — Grading & Deadline Enforcement ✅ COMPLETE

**Goal:** Give teachers the ability to assign numeric grades and enforce deadlines so the data foundation exists for analytics.

### 1.1 Assignment Grading

**Database changes:**
```sql
-- Add grade columns to submission
ALTER TABLE submission ADD COLUMN grade DECIMAL(5,2);        -- e.g. 17.5 out of 20
ALTER TABLE submission ADD COLUMN max_grade DECIMAL(5,2);    -- e.g. 20
ALTER TABLE submission ADD COLUMN graded_at TIMESTAMPTZ;
```

**What to build:**
- [x] Admin submission review page gets a **grade input** (numeric) + max grade field
- [x] `ReviewSubmission` handler saves grade alongside existing feedback + status
- [x] Assignment list shows **graded / total** count instead of just submission count
- [x] Student sees their grade + feedback on the assignment page after grading
- [x] Assignment creation form gets a **max grade** field (default 20) so all submissions inherit it

### 1.2 Deadline Enforcement

**What to build:**
- [x] `StudentSubmit` handler checks `assignment.deadline` — if past deadline, reject with error message
- [x] Student assignment page shows a **countdown timer** or "Deadline passed" badge
- [ ] Admin assignment list shows ⏰ icon for upcoming deadlines, ❌ for expired
- [ ] Optional: Admin toggle "Allow late submissions" (boolean on assignment) — late ones get flagged

### 1.3 Quiz Improvements

**Database changes:**
```sql
-- Add deadline and attempt limit to quiz
ALTER TABLE quiz ADD COLUMN deadline TIMESTAMPTZ;
ALTER TABLE quiz ADD COLUMN time_limit_minutes INT DEFAULT 0;     -- 0 = unlimited
ALTER TABLE quiz ADD COLUMN max_attempts INT DEFAULT 1;            -- 0 = unlimited
```

**What to build:**
- [x] Quiz creation form: deadline picker, time limit (minutes), max attempts
- [x] `StudentSubmitQuiz` handler: reject if past deadline or max attempts reached
- [x] If `time_limit_minutes > 0`, student quiz page shows a **live countdown timer** (JS) — auto-submit on expiry
- [x] Student quiz page: show "Attempt X of Y" and disable if exhausted
- [ ] Register the existing `StudentQuizResult` handler on a route so students can see their detailed results
- [x] Wire up `UpdateQuizQuestion` so admin can **edit** existing questions (not just add/delete)

### 1.4 Enhanced Quiz Question Types

**Database changes:**
```sql
-- Expand allowed question types
ALTER TABLE quiz_question DROP CONSTRAINT quiz_question_question_type_check;
ALTER TABLE quiz_question ADD CONSTRAINT quiz_question_question_type_check
    CHECK (question_type IN ('mcq', 'true_false', 'fill_blank', 'open_ended', 'file_upload'));

-- Store file-based answers
ALTER TABLE quiz_attempt ADD COLUMN file_answers JSONB DEFAULT '{}';
-- format: {"question_id": {"file_path": "...", "file_name": "..."}}
```

**What to build:**
- [x] New question type `file_upload` — student uploads a file (PDF, image, etc.) as their answer
- [x] Admin quiz editor: option to add `file_upload` type question with instructions
- [x] Student quiz page: file input for `file_upload` questions
- [x] `StudentSubmitQuiz` handler: save uploaded files, store paths in `file_answers` JSONB
- [x] Admin quiz review: display/download uploaded file answers
- [x] `file_upload` questions are **never auto-graded** — always require admin review

**Estimated effort:** ~3-4 days

---

## Phase 2 — Teacher Analytics Dashboard ✅ COMPLETE

**Goal:** Give the teacher a data-driven overview of how the class is performing.

### 2.1 Quiz Analytics

**What to build:**
- [x] New page: `/admin/classroom/:id/analytics` (new tab alongside Resources, Assignments, Quizzes, Students)
- [x] **Quiz performance table:**
  - Per quiz: title, # attempts, average score %, highest, lowest, median
  - Click to expand → per-question breakdown
- [x] **Question difficulty analysis:**
  - For each question: % of students who answered correctly
  - Color-coded: 🟢 >80% correct (easy), 🟡 40-80% (medium), 🔴 <40% (hard)
  - Shows the most common wrong answer for MCQ/True-False
- [x] **Per-student quiz breakdown:**
  - Table: student name, score, time taken (finished_at - started_at), attempt number
  - Sortable by score, time, name
- [x] **Quiz comparison chart:**
  - Simple bar chart (CSS-only, no JS library needed) showing average scores across quizzes

### 2.2 Assignment Analytics

**What to build:**
- [x] **Assignment overview table:**
  - Per assignment: title, submissions received vs enrolled students, average grade, grade distribution
  - Deadline status: on-time vs late submissions count
- [x] **Grade distribution:**
  - Simple histogram bins: A (90-100%), B (80-89%), C (70-79%), D (60-69%), F (<60%)
  - Or configurable by teacher
- [x] **Missing submissions alert:**
  - List of students who haven't submitted for each assignment
  - One-click to send reminder (future: email/notification, for now just a list)

### 2.3 Student Roster Analytics

**What to build:**
- [x] **Class roster with aggregated data:**
  - Per student row: name, email, avg quiz score, assignments submitted/total, avg assignment grade
  - Overall "engagement score" (simple: % of quizzes taken + % of assignments submitted)
- [x] **Student detail page** (`/admin/classroom/:id/student/:studentId`):
  - All quiz attempts with scores
  - All assignment submissions with grades + feedback
  - Timeline of activity

**Estimated effort:** ~4-5 days

---

## Phase 3 — Live Session Tracking & Attendance ✅ COMPLETE (3.1 + 3.2)

**Goal:** Track who attends live classes, for how long, and give the teacher a history.

### 3.1 Attendance Tracking

**Database changes:**
```sql
-- Track live session attendance
CREATE TABLE live_attendance (
    id            SERIAL PRIMARY KEY,
    live_session_id INT NOT NULL REFERENCES live_session(id) ON DELETE CASCADE,
    student_id    INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    joined_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at       TIMESTAMPTZ,
    duration_seconds INT GENERATED ALWAYS AS (
        EXTRACT(EPOCH FROM (COALESCE(left_at, NOW()) - joined_at))
    ) STORED
);

-- Track when teacher starts/ends sessions with more detail
ALTER TABLE live_session ADD COLUMN ended_at TIMESTAMPTZ;
ALTER TABLE live_session ADD COLUMN duration_minutes INT;
```

**What to build:**
- [x] When a student joins a live class → insert `live_attendance` row with `joined_at`
- [x] When a student leaves (disconnect event or page unload) → update `left_at`
- [x] When teacher ends class → update `live_session.ended_at`, compute duration
- [x] Use `beforeunload` JS event + a `/api/live/leave` endpoint to catch page closes

### 3.2 Session History

**What to build:**
- [x] New section in teacher analytics: **Live Class History**
  - Table of past sessions: date, start time, end time, duration, # students who joined
  - Click to expand → per-student attendance: name, joined at, left at, time spent
- [x] **Attendance patterns:**
  - Per student: total sessions attended / total sessions held = attendance rate %
  - Flag students with <50% attendance
- [ ] Show attendance data on the student roster (Phase 2.3 integration)

### 3.3 Teacher Activity Log

**Database changes:**
```sql
CREATE TABLE activity_log (
    id          SERIAL PRIMARY KEY,
    classroom_id INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    action      TEXT NOT NULL,   -- 'live_start', 'live_end', 'assignment_created', 'quiz_published', 'grade_given', etc.
    details     JSONB,           -- flexible payload
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**What to build:**
- [ ] Log key teacher actions: start/end live class, create assignment, grade submission, publish quiz, etc.
- [ ] Timeline view in analytics showing recent classroom activity
- [ ] This becomes the foundation for any future notification system

**Estimated effort:** ~3-4 days

---

## Phase 4 — Student Dashboard & Self-Analytics ✅ COMPLETE

**Goal:** Let students see their own performance, attendance, and teacher remarks.

### 4.1 Student Performance Dashboard

**What to build:**
- [x] New page: `/classroom/:id/dashboard` ("My Progress" tab in the student classroom view)
- [x] **My Grades:**
  - List of all assignments with: title, grade received (or "Not graded yet" / "Not submitted"), teacher feedback snippet
  - List of all quizzes with: title, score, attempt count, best score
  - Overall average grade across assignments
  - Overall average quiz score
- [x] **Visual progress:**
  - Score cards with quiz performance bar chart
  - "Above class average" / "Below average" indicator (without revealing other students' grades)

### 4.2 Attendance Record

**What to build:**
- [x] **My Attendance:**
  - List of live sessions held, which ones the student attended
  - Total attendance rate as percentage
  - "Sessions attended: 8 / 12 (67%)"
  - Progress bar and color-coded attendance rate
- [x] Session table with Present/Absent status and time spent

### 4.3 Teacher Remarks & Feedback

**Database changes:**
```sql
-- General teacher remarks (not tied to a specific assignment/quiz)
CREATE TABLE student_remark (
    id           SERIAL PRIMARY KEY,
    classroom_id INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    student_id   INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    content      TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**What to build:**
- [x] Admin can add free-text **remarks** to any student from the student detail page
- [x] Student sees their remarks in their dashboard — like teacher notes/comments
- [x] Useful for things like "Great improvement this week" or "Please review Chapter 3"
- [x] Remarks appear as a timeline feed on the student dashboard
- [x] Admin can delete remarks from student detail page

**Estimated effort:** ~3-4 days

---

## Phase 5 — Advanced Analytics & Reports ✅ COMPLETE

**Goal:** Deeper insights and exportable reports.

### 5.1 Class-Wide Analytics

**What to build:**
- [x] **Performance trends over time:**
  - Chart showing class average score per quiz over time (are they improving?)
  - Assignment grade trends
- [x] **Engagement metrics:**
  - Resource download/view counts (add tracking to download handler)
  - Time-on-quiz (already have started_at/finished_at)
  - Submission timing patterns (how many submit on deadline day vs. early)
- [x] **Risk detection:**
  - Flag students who: missed 2+ assignments, scored <50% on last 2 quizzes, attendance <50%
  - Simple "At Risk" badge on student roster

### 5.2 Export & Reports

**What to build:**
- [x] **CSV export** for:
  - Student roster with grades
  - Quiz results per quiz
  - Assignment grades
  - Attendance records
- [x] **Classroom summary report** (HTML page, print-friendly CSS):
  - Class stats, per-student summary table, quiz/assignment averages
  - Printable as PDF via browser print

### 5.3 Resource Tracking

**Database changes:**
```sql
CREATE TABLE resource_view (
    id          SERIAL PRIMARY KEY,
    resource_id INT NOT NULL REFERENCES resource(id) ON DELETE CASCADE,
    student_id  INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    viewed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**What to build:**
- [x] Track when students download/view resources
- [x] Admin sees view count per resource
- [x] Analytics: which resources are most/least accessed
- [x] Per-student: which resources they've accessed

**Estimated effort:** ~4-5 days

---

## Summary Timeline

| Phase | Focus | Key Deliverables | Status |
|-------|-------|-----------------|--------|
| **Phase 1** | Grading + Deadlines + Quiz Enhancements | Numeric grades, deadline enforcement, quiz time limits, file-upload answers | ✅ Done |
| **Phase 2** | Teacher Analytics Dashboard | Quiz difficulty analysis, grade distributions, student roster analytics, student detail page | ✅ Done |
| **Phase 3** | Live Session Tracking | Attendance records, session history, attendance rates | ✅ Done (3.1 + 3.2) |
| **Phase 4** | Student Dashboard | Self-service grades, attendance record, teacher remarks | ✅ Done |
| **Phase 5** | Advanced Analytics + Export | Trend charts, risk detection, CSV exports, resource tracking | ✅ Done |

**Total estimated: ~17-22 days of implementation work**

---

## Database Migration Strategy

Each phase has its own set of `ALTER TABLE` / `CREATE TABLE` statements. We'll add them as migration files:

```
db/
├── schema.sql              # Original schema (don't modify)
├── migrations/
│   ├── 001_grading.sql     # Phase 1: grade columns, quiz deadline/limits, file_upload type
│   ├── 002_analytics.sql   # Phase 2: (no schema changes, just queries)
│   ├── 003_attendance.sql  # Phase 3: live_attendance, activity_log tables
│   ├── 004_student.sql     # Phase 4: student_remark table
│   └── 005_tracking.sql    # Phase 5: resource_view table
```

For now (dev mode), we'll just add columns directly. For production, proper migration tooling (like `golang-migrate`) can be added later.

---

## Architecture Notes

- **No JS charting libraries needed.** We'll use CSS-only bar charts and progress bars via Tailwind utility classes. Simple, fast, no dependencies.
- **All analytics are server-rendered.** Computed by SQL aggregate queries, rendered in Go templates. No client-side data fetching or SPA patterns.
- **Queries will use PostgreSQL aggregate functions:** `AVG()`, `COUNT()`, `PERCENTILE_CONT()`, `GROUP BY`, window functions for trends.
- **Same tech stack throughout:** Go + Gin + pgx + Tailwind + HTMX. No new dependencies unless absolutely necessary.

---

## How to Use This Document

1. **Pick a phase** — start with Phase 1 (it's the foundation for everything else)
2. **Tell me to implement it** — I'll add the schema changes, store functions, handlers, and templates
3. **Test it** — run the app, try the features
4. **Move to the next phase** — each one builds on the previous

You can also pick individual items within a phase if you want to go more granularly.
