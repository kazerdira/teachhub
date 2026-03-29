# TEACHHUB — Development Specification

**Phase: Parent Reports & Smart Alerts**
**Version:** 1.0 · **Date:** March 2026 · **Status:** Ready for Development

---

## 1. Why We're Building This

The parent pays for private tutoring. The parent chooses the teacher. The parent decides whether to continue or switch. But right now, the parent gets zero visibility into their child's progress. They hand over cash and ask their kid "how was the lesson?" — and get a shrug.

This spec covers two features that turn parents into our growth engine:

- **Parent Progress Page** — a shareable, mobile-optimized view of each student's performance
- **Smart Alerts** — automatic notifications for missed deadlines, absent students, and falling grades

---

## 2. Feature: Parent Progress Page

### 2.1 Overview

Every student enrolled in a classroom gets a unique secret URL. When the parent opens this URL on their phone, they see a clean read-only page showing their child's quiz scores, assignment grades, attendance record, and teacher remarks. No login. No app download. No signup.

### 2.2 Database Changes

Add a single column to the `classroom_student` join table:

**Migration: `010_parent_code.sql`**

```sql
ALTER TABLE classroom_student
  ADD COLUMN IF NOT EXISTS parent_code TEXT
  UNIQUE DEFAULT encode(gen_random_bytes(6), 'hex');

CREATE INDEX IF NOT EXISTS idx_cs_parent_code
  ON classroom_student(parent_code);
```

The `parent_code` is a 12-character hex string, generated automatically. One code per student per classroom. Not guessable, not sequential.

### 2.3 New Route

**`GET /p/{parent_code}`** — public, no authentication required.

This route looks up the `parent_code` in `classroom_student`, finds the `student_id` and `classroom_id`, then renders the parent report page using existing store functions.

### 2.4 Data to Display

The page reuses existing store queries. No new analytics queries needed. **All 5 functions below already exist in `store/store.go` and are battle-tested by the admin student detail page.**

| Section | Store Function | Display | Status |
|---------|---------------|---------|--------|
| Quiz Scores | `GetStudentQuizDetails()` | Score, %, trend arrow | ✅ Exists |
| Assignments | `GetStudentAssignmentDetails()` | Grade, status badge | ✅ Exists |
| Attendance | `GetStudentAttendanceRecord()` | Green/red dots, % | ✅ Exists |
| Class Position | `GetStudentDashboardStats()` | Above/below class avg | ✅ Exists |
| Teacher Remarks | `GetStudentRemarks()` | Latest 5 remarks | ✅ Exists |

### 2.5 Page Design Requirements

- Mobile-first. 90%+ of parents will open this on a phone via WhatsApp.
- No navigation, no header, no footer. Just a clean white card with the student's data.
- Top of page: student name, classroom name, teacher name, last updated timestamp.
- Language: follow the classroom's teacher language preference (FR or EN). Consider Arabic later.
- No interactive elements. Pure read-only. No JavaScript required — server-rendered HTML.
- Must load fast on 3G. No Tailwind CDN. Inline critical CSS only.
- Show a colored banner at top: green if student is above class average, amber if near, red if below.

### 2.6 WhatsApp Share Integration

On the admin student detail page and the classroom student list, add a "Share with Parent" button per student. This button opens WhatsApp with a pre-filled message:

> **French:**
> Bonjour, voici le suivi de {StudentName} dans ma classe: {URL}

> **Arabic:**
> متابعة {StudentName}: {URL}

Implementation: standard WhatsApp click-to-chat URL scheme:

```
https://wa.me/?text=Bonjour%2C+voici+le+suivi+de+...
```

### 2.7 Teacher Controls

- Teacher can regenerate a student's `parent_code` (invalidates old link).
- Teacher can see all parent codes in the student list (small "copy link" icon per student).
- No bulk sharing. Teacher sends one link per student. This is intentional — it forces a personal touch.

### 2.8 Security

- The `parent_code` is 12 hex chars = 48 bits of entropy. Not guessable.
- No authentication. No session. No cookies. The URL IS the access token.
- Rate-limit the `/p/` route: max 60 requests per IP per minute.
- The page shows student name, scores, and teacher remarks. No email, no phone, no other students' data.

---

## 3. Feature: Smart Alerts

### 3.1 Overview

Smart alerts are pre-composed WhatsApp messages that the teacher can send with one tap. TeachHub detects situations that need attention and surfaces them as actionable alerts on the dashboard. The teacher decides whether to send each one.

**Important:** TeachHub does NOT send messages automatically. It prepares the message and opens WhatsApp. The teacher taps send. This keeps the teacher in control and avoids any SMS/API cost.

### 3.2 Alert Types

#### Alert 1: Assignment Deadline Approaching

| | |
|---|---|
| **Trigger** | Assignment deadline is within 24 hours AND student has not submitted |
| **Check Frequency** | Every time teacher opens classroom page |
| **Target** | Student (via parent WhatsApp if parent_code exists) |
| **Message (FR)** | Rappel: {StudentName} n'a pas encore soumis "{AssignmentTitle}". Date limite: {Deadline}. |

#### Alert 2: Quiz Not Attempted

| | |
|---|---|
| **Trigger** | Published quiz has deadline within 24 hours AND student has zero attempts |
| **Check Frequency** | Every time teacher opens classroom page |
| **Target** | Student (via parent WhatsApp if parent_code exists) |
| **Message (FR)** | Rappel: {StudentName} n'a pas encore passé le quiz "{QuizTitle}". Date limite: {Deadline}. |

#### Alert 3: Student Absent from Live Session

| | |
|---|---|
| **Trigger** | Live session has been active for 10+ minutes AND student has not joined |
| **Check Frequency** | Checked once per session, 10 minutes after session start |
| **Target** | Parent (via WhatsApp) |
| **Message (FR)** | {StudentName} n'est pas encore connecté(e) au cours en direct de {TeacherName}. Le cours a commencé à {StartTime}. |

#### Alert 4: Grade Drop Warning

| | |
|---|---|
| **Trigger** | Student scored below 50% on their last 2 quizzes in the same classroom |
| **Check Frequency** | After quiz grading / review |
| **Target** | Parent (via WhatsApp) |
| **Message (FR)** | {StudentName} a obtenu moins de 50% sur les 2 derniers quiz. Scores: {Score1}%, {Score2}%. Un suivi supplémentaire serait bénéfique. |

#### Alert 5: Attendance Drop

| | |
|---|---|
| **Trigger** | Student has missed 3+ of the last 5 live sessions |
| **Check Frequency** | After a live session ends |
| **Target** | Parent (via WhatsApp) |
| **Message (FR)** | {StudentName} a manqué {MissedCount} des 5 dernières séances. L'assiduité est importante pour la réussite. |

### 3.3 UI: Alert Center

Alerts appear in two places:

**1. Classroom dashboard banner:** A yellow/red banner at the top of the classroom page showing the count of pending alerts. Example: "⚠ 3 students need attention — View alerts".

**2. Alert panel:** A slide-out panel (or dedicated page) listing all current alerts for that classroom. Each alert card shows: the student name, the alert type (icon + label), the pre-composed message, and two buttons: "📱 Send via WhatsApp" and "✕ Dismiss".

Dismissed alerts are hidden for 7 days (store dismissal timestamp per alert type per student in a new table). After 7 days, if the condition still holds, the alert reappears.

### 3.4 Database for Dismissals

```sql
CREATE TABLE IF NOT EXISTS alert_dismissal (
    id            SERIAL PRIMARY KEY,
    classroom_id  INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    student_id    INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    alert_type    TEXT NOT NULL,
    dismissed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## 4. Implementation Order

Work in this exact sequence. Each step is independently shippable.

### Step 1: Parent Code Infrastructure
1. Run migration `010_parent_code.sql`
2. Add `parent_code` to the Student struct and all relevant queries in `store.go`
3. Add store function: `GetStudentByParentCode(ctx, code)` returning `student_id`, `classroom_id`
4. Add store function: `RegenerateParentCode(ctx, classroomID, studentID)`

### Step 2: Parent Report Page
1. Create handler: `ParentReport(c *gin.Context)` — `GET /p/:code`
2. Create template: `parent_report.html` — mobile-first, self-contained CSS, no external dependencies
3. Wire the route in `main.go` (public group, no auth middleware)
4. Add rate limiting on `/p/` routes (60 req/min per IP)

### Step 3: WhatsApp Share Button
1. Add "Share with Parent" button to `admin_student_detail.html`
2. Add small share icon per student row in the classroom student list
3. Add "Copy parent link" button to the same locations

### Step 4: Alert Detection Queries
1. Add store function: `GetPendingAlerts(ctx, classroomID)` returning `[]Alert`
2. This function runs all 5 alert checks in a single query or batched queries
3. Each Alert struct contains: `StudentID`, `StudentName`, `AlertType`, `MessageFR`, `MessageAR`, `WhatsAppURL`
4. Run migration for `alert_dismissal` table

### Step 5: Alert UI
1. Add alert banner to `admin_classroom.html` template
2. Create alert panel page/partial: `admin_alerts.html`
3. Add dismiss handler: `POST /admin/classroom/:id/alert/dismiss`

---

## 5. What NOT to Build

Explicitly out of scope for this phase:

- No push notifications. No SMS. No email. WhatsApp click-to-chat only.
- No parent accounts or parent login. The URL is the access.
- No real-time updates on the parent page. It shows data as of page load. Parent refreshes to see new data.
- No automatic message sending. Teacher always taps the send button.
- No Arabic RTL layout yet. French and English only. Arabic is Phase 2.
- No n8n, no webhooks, no API integrations.

---

## 6. Priority Matrix

P0 = must ship. P1 = should ship. P2 = nice to have.

| Priority | Item | Effort | Notes |
|----------|------|--------|-------|
| **P0** | Parent code migration + store functions | 1 hour | Only 1 new column + 2 small store functions. All 5 data queries already exist. |
| **P0** | Parent report page (handler + template) | 3–4 hours | Handler calls existing functions. Template is standalone HTML, hand-written CSS. |
| **P0** | WhatsApp share button on student detail + student list | 1 hour | wa.me link with URL-encoded message |
| **P0** | `/p/` route rate limiting | 30 min | Reuse existing RateLimiter middleware |
| **P0** | Copy link button on student list | 30 min | Moved from P2 — trivial and teachers need it day one |
| **P0** | Regenerate parent code button | 30 min | Moved from P2 — security essential |
| **P1** | Alert detection queries (all 5 types) | 6–10 hours | Realistic: edge cases (0 quizzes, 1 session, no attendance), timezone handling, WhatsApp URL escaping |
| **P1** | Alert banner on classroom dashboard | 2 hours | |
| **P1** | Alert panel with dismiss + WhatsApp buttons | 3 hours | |
| **P1** | Alert dismissal table + 7-day cooldown logic | 1 hour | |
| **P2** | Arabic message variants | 1 hour | Phase 2 — RTL layout needed |

> **Implementation note:** Consider shipping a "lite alerts" version first — just show "⚠ X students haven't submitted" per approaching-deadline assignment on the classroom page. One SQL query, zero new tables, covers 60% of alert value. Full 5-type alerts can wait until 10+ active teachers.

---

## 7. Success Metrics

How we know this worked:

- Parent page views per week (target: 3+ views per active parent per week)
- WhatsApp share button taps per teacher per week (target: 5+ per active teacher)
- Teacher retention: teachers who use parent reports should have higher 30-day retention than those who don't
- Viral coefficient: track how many new teacher signups come from parents asking "what tool is this?" (add UTM parameter to parent page: "Powered by TeachHub — Apply here")

**The parent report page footer must include:** "Powered by TeachHub — Become a teacher" with a link to `/apply`. This is the viral loop.

---

## 8. Technical Notes for Developers

- The parent report handler must NOT use the admin or student session. It reads `parent_code` from URL, queries the DB, renders the template. Zero auth.
- The parent report template must be a standalone HTML file with inlined CSS. No Tailwind CDN. No JS frameworks. Target: under 15KB total page size.
- Alert detection should be lazy — computed when the teacher opens the classroom, not via background jobs. Keep the architecture simple.
- WhatsApp URLs use the `wa.me` scheme: `https://wa.me/?text={urlencoded message}`. This works on both mobile and desktop WhatsApp.
- All alert messages must be available in both French and English (use the existing `t()` translation system). Arabic translations can be added later.
- The alert dismissal check is simple: `WHERE dismissed_at > NOW() - INTERVAL '7 days' AND alert_type = $1 AND student_id = $2 AND classroom_id = $3`.
- For the live session absence alert (Alert 3), compute absent students by comparing `classroom_student` roster against `live_attendance` for the current active session. Only show this alert if the session has been active for 10+ minutes (compare `live_session.created_at` to `NOW()`).