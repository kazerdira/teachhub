# 🔍 Teacher Explore & Directory Feature

## Overview

Transform TeachHub from a private classroom tool into a discoverable marketplace where students can find teachers by subject, level, and region — then request to join.

---

## Architecture Decisions

| Decision | Choice | Reason |
|----------|--------|--------|
| Country detection | MaxMind GeoLite2 (local DB) | Silent, fast, no external API, no user prompt |
| Search engine | PostgreSQL first, Meilisearch later | PG handles <1000 teachers fine. Meilisearch when we scale (instant typo-tolerant search, faceted filters) |
| Chat | None | Legal risk with minors, moderation burden, scope creep. Email + phone is enough. |
| Student account | Created on approval | No signup friction — student fills a form, teacher creates the account |
| Subject/Level data | Stored on both classroom AND teacher profile | Classroom = specific, Profile = summary for public display |

---

## DB Changes

### Alter `classroom` table
```sql
ALTER TABLE classroom ADD COLUMN subject TEXT NOT NULL DEFAULT '';
ALTER TABLE classroom ADD COLUMN level TEXT NOT NULL DEFAULT '';
```

### Alter `admin` (teacher) table
```sql
ALTER TABLE admin ADD COLUMN bio TEXT NOT NULL DEFAULT '';
ALTER TABLE admin ADD COLUMN subjects TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE admin ADD COLUMN levels TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE admin ADD COLUMN country TEXT NOT NULL DEFAULT '';    -- 'DZ', 'FR', etc.
ALTER TABLE admin ADD COLUMN public_profile BOOLEAN NOT NULL DEFAULT false;
```

### New `join_request` table
```sql
CREATE TABLE IF NOT EXISTS join_request (
    id           SERIAL PRIMARY KEY,
    teacher_id   INT NOT NULL REFERENCES admin(id) ON DELETE CASCADE,
    classroom_id INT REFERENCES classroom(id) ON DELETE SET NULL,  -- set on approval
    full_name    TEXT NOT NULL,
    email        TEXT NOT NULL,
    phone        TEXT NOT NULL DEFAULT '',
    level        TEXT NOT NULL DEFAULT '',
    message      TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at  TIMESTAMPTZ
);
```

---

## Subject & Level Lists

### Subjects (shared, both countries)
```
Math, Physics, Chemistry, Biology, French, English, Arabic, 
History-Geography, Philosophy, Islamic Sciences, 
Computer Science, Economics, Spanish, German, Italian,
Civil Engineering, Electrical Engineering, Mechanical Engineering
```

### Levels — Algeria 🇩🇿
```
Primary: 1AP, 2AP, 3AP, 4AP, 5AP
Middle: 1AM, 2AM, 3AM, 4AM (BEM)
Secondary: 1AS, 2AS, 3AS (BAC)
University: Licence, Master, Doctorat
```

### Levels — France 🇫🇷
```
Primary: CP, CE1, CE2, CM1, CM2
Collège: 6ème, 5ème, 4ème, 3ème (Brevet)
Lycée: Seconde, Première, Terminale (BAC)
Supérieur: Licence, Master, Prépa, BTS, DUT
```

---

## Phase 1 — Foundation (DB + Geolocation)

### 1.1 Database migrations
- [ ] Add `subject`, `level` to `classroom`
- [ ] Add `bio`, `subjects`, `levels`, `country`, `public_profile` to `admin`
- [ ] Create `join_request` table
- [ ] Update Go structs: `Classroom`, `Admin`
- [ ] Update all SQL queries that SELECT/INSERT on these tables

### 1.2 MaxMind GeoLite2 integration
- [ ] Download GeoLite2-Country.mmdb (free, ~5MB)
- [ ] Add `github.com/oschwald/maxminddb-golang` to go.mod
- [ ] Create `geo/geo.go` — helper: `CountryFromIP(ip string) string` → "DZ", "FR", etc.
- [ ] Middleware or helper to detect country on each request, store in cookie `country`
- [ ] Bundle .mmdb file in Docker image

### 1.3 Subject/Level reference data
- [ ] Create `geo/subjects.go` — lists of subjects
- [ ] Create `geo/levels.go` — level lists per country (DZ, FR)
- [ ] i18n keys for all subjects and levels (fr + en)

---

## Phase 2 — Classroom Subject & Level

### 2.1 Create Classroom flow update
- [ ] When teacher clicks "Create Classroom" → popup/modal asks: Name + Subject + Level
- [ ] Subject dropdown populated from reference data
- [ ] Level dropdown adapts to teacher's detected country
- [ ] Store `subject` and `level` on classroom record

### 2.2 Edit Classroom
- [ ] Teacher can edit subject/level on existing classrooms
- [ ] Prompt existing teachers to fill subject/level on first visit (soft migration)

### 2.3 Display
- [ ] Show subject + level badge on classroom cards in teacher dashboard
- [ ] Show in student classroom view

---

## Phase 3 — Teacher Public Profile

### 3.1 Profile settings page
- [ ] New page or section in teacher settings: `/admin/profile`
- [ ] Fields: Bio (textarea, 300 char), Subjects (multi-select), Levels (multi-select)
- [ ] Country auto-detected (displayed, editable)
- [ ] Region: Wilaya dropdown (DZ) or Département dropdown (FR)
- [ ] "Show my profile publicly" toggle (off by default)
- [ ] Save → updates admin record

### 3.2 Auto-populate from classrooms
- [ ] If teacher hasn't set subjects/levels, auto-suggest from their classroom data
- [ ] "You teach Math 3AS and Physics 2AS — add these to your profile?"

---

## Phase 4 — Public Explore Page

### 4.1 Route & handler
- [ ] `GET /explore` — public, no auth required
- [ ] Handler: detect country from IP/cookie, load teachers with `public_profile=true`
- [ ] Filter params: `?subject=math&level=3AS&region=alger`

### 4.2 Template: `templates/explore.html`
- [ ] Standalone page (own layout, no admin/student chrome)
- [ ] Hero section: "Find a Teacher" + search/filter bar
- [ ] Filters:
  - Region (Wilaya or Département, auto-detected from country)
  - Subject (dropdown)
  - Level (dropdown, adapts to country)
- [ ] Teacher cards grid:
  - Avatar (first letter)
  - Name, school
  - Subjects + levels (badges)
  - Region
  - Student count
  - Bio snippet
  - "Request to Join" button
- [ ] Empty state: "No teachers found matching your filters"
- [ ] Responsive (mobile-first)

### 4.3 Search (future: Meilisearch)
- [ ] **V1**: PostgreSQL `ILIKE` + `ANY(subjects)` + `ANY(levels)` — good enough for <1000 teachers
- [ ] **V2 (later)**: Add Meilisearch container to docker-compose
  - Index: teacher profiles (id, name, school, subjects, levels, region, bio)
  - Instant search with typo tolerance
  - Faceted filtering
  - Auto-sync on profile update

---

## Phase 5 — Join Request Flow

### 5.1 Request form
- [ ] Accessible from explore page teacher card
- [ ] Modal or dedicated page: `/explore/teacher/{id}/request`
- [ ] Fields: Full name, Email, Phone, Level (dropdown), Message (optional)
- [ ] CSRF protection
- [ ] Rate limiting (max 3 requests per email per day)
- [ ] Success message: "Your request has been sent! The teacher will contact you."

### 5.2 Teacher request management
- [ ] New tab/badge in teacher dashboard: "Join Requests (3)"
- [ ] List view: name, email, phone, level, message, date
- [ ] Actions per request:
  - **Approve**: dropdown to pick which classroom → auto-creates student account → adds to classroom
  - **Reject**: optional reason
- [ ] On approve:
  - Create student record (name, email, phone)
  - Add to selected classroom with status='approved'
  - Generate login credentials
  - Send email with credentials (if SMTP configured) OR show credentials on screen

### 5.3 Store functions
- [ ] `CreateJoinRequest(teacherID, name, email, phone, level, message)`
- [ ] `ListJoinRequests(teacherID, status)` — pending/all
- [ ] `ApproveJoinRequest(requestID, classroomID)` — creates student + joins classroom
- [ ] `RejectJoinRequest(requestID)`
- [ ] `CountPendingJoinRequests(teacherID)` — for badge count

---

## Phase 6 — Owner Panel Visibility

### 6.1 Join requests overview
- [ ] New section or tab in owner panel: "Join Requests"
- [ ] Table: student name, teacher name, subject, status, date
- [ ] Stats: total requests, approved, rejected, pending, conversion rate
- [ ] Filter by teacher, by date range

### 6.2 Public profile stats
- [ ] How many teachers have public profiles
- [ ] Most viewed/requested teachers

---

## Phase 7 — Meilisearch (when needed)

### 7.1 Setup
- [ ] Add Meilisearch container to `docker-compose.yml`
- [ ] Configure API key
- [ ] Create index: `teachers`

### 7.2 Indexing
- [ ] On teacher profile save → upsert document in Meilisearch
- [ ] On teacher delete/unpublish → remove from index
- [ ] Fields: id, name, school, subjects, levels, region, country, bio, student_count

### 7.3 Search endpoint
- [ ] `GET /api/search?q=math+alger` → queries Meilisearch → returns JSON
- [ ] Frontend: instant search with debounce (fetch as you type)
- [ ] Faceted filters: subject, level, region

---

## File Changes Summary

### New files
```
geo/geo.go                           — IP geolocation helper
geo/subjects.go                      — Subject reference lists
geo/levels.go                        — Level lists per country
templates/explore.html               — Public explore page
templates/admin/admin_profile.html   — Teacher profile settings
templates/admin/admin_requests.html  — Teacher join request management
static/GeoLite2-Country.mmdb        — MaxMind DB file
```

### Modified files
```
db/schema.sql                        — New columns + join_request table
store/store.go                       — New structs, new queries, updated queries
handlers/admin.go                    — Profile settings, request management
handlers/explore.go (new)            — Public explore page + request submission
handlers/platform.go                 — Owner visibility for requests
main.go                              — New routes, geolocation init
docker-compose.yml                   — (Phase 7) Meilisearch container
i18n/en.json + fr.json              — New translation keys
templates/admin/admin_dashboard.html — Join request badge, subject/level on classroom cards
templates/layouts/admin.html         — Nav link for requests
```

---

## Implementation Order

```
Phase 1  →  Foundation (DB + geo + reference data)
Phase 2  →  Classroom subject/level
Phase 3  →  Teacher public profile
Phase 4  →  Public explore page
Phase 5  →  Join request flow
Phase 6  →  Owner visibility
Phase 7  →  Meilisearch (later, when scale demands it)
```

Start with Phase 1 when ready.
