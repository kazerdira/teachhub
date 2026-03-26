# TeachHub — Scaling Guide

> **When to read this:** You have 50+ active teachers and are starting to notice
> slow analytics pages, or you're planning to run multiple server instances.
> Before that point, none of this is necessary — the current architecture handles
> 50 teachers × 100 students each without breaking a sweat.

---

## Table of Contents

1. [Current Architecture (as of March 2026)](#current-architecture)
2. [What's Already Optimized](#whats-already-optimized)
3. [Scaling Step 1: Query Timeouts & Connection Pool Tuning](#step-1-query-timeouts)
4. [Scaling Step 2: Rewrite GetQuestionAnalytics to SQL](#step-2-rewrite-getquestionanalytics)
5. [Scaling Step 3: Add Composite Indexes](#step-3-composite-indexes)
6. [Scaling Step 4: Redis Caching for Analytics](#step-4-redis-caching)
7. [Scaling Step 5: Object Storage for File Uploads](#step-5-object-storage)
8. [Scaling Step 6: Paginate Remaining List Endpoints](#step-6-paginate-remaining-lists)
9. [Scaling Step 7: Redis Rate Limiter](#step-7-redis-rate-limiter)
10. [What Does NOT Need Fixing](#what-does-not-need-fixing)

---

## Current Architecture

```
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│    Caddy      │──────│   Go App     │──────│ PostgreSQL   │
│  (HTTPS/WSS) │      │  (Gin, 1     │      │  16-alpine   │
│  port 80/443 │      │   instance)  │      │  port 5432   │
└──────┬───────┘      └──────────────┘      └──────────────┘
       │
       │ /rtc/*
       ▼
┌──────────────┐
│   LiveKit     │
│   v1.10.0    │
│  SFU server  │
│ 7880/7881/UDP│
└──────────────┘
```

**Server:** Contabo VPS, 144.91.77.206, Ubuntu 24.04  
**Domain:** teachhub.chickenkiller.com (HTTPS via Caddy auto-TLS)  
**Deploy:** `git push` → `ssh` → `git pull` → `docker compose up -d --build app`  
**Static assets:** Self-hosted in `/static/` — compiled Tailwind CSS (58KB), htmx (47KB), livekit-client (469KB). Zero CDN dependencies.

### Database Tables

| Table | Purpose | Typical row count at 50 teachers × 100 students |
|-------|---------|------------------------------------------------|
| `admin` | Teacher accounts | 50 |
| `student` | Student accounts | ~3,000 (many shared across classrooms) |
| `classroom` | Teacher classrooms | ~150 (3 per teacher avg) |
| `classroom_student` | Enrollment (many-to-many) | ~5,000 |
| `resource` | Uploaded files/links | ~2,000 |
| `assignment` | Homework assignments | ~500 |
| `submission` | Student file/text submissions | ~10,000 |
| `quiz` | Quizzes | ~300 |
| `quiz_question` | Questions per quiz | ~3,000 |
| `quiz_attempt` | Student quiz attempts with JSONB answers | ~15,000 |
| `live_session` | Live video session records | ~500 |
| `live_attendance` | Who joined which session | ~5,000 |
| `allowed_student` | Pre-registered students | ~3,000 |
| `category` | Resource categories | ~500 |
| `student_remark` | Teacher notes on students | ~1,000 |
| `resource_view` | View tracking | ~20,000 |
| `payment` | Payment records (platform) | ~200 |
| `teacher_application` | Teacher signups | ~200 |
| `platform_admin` | Platform owner accounts | 1-2 |

### Existing Indexes

All tables have proper single-column indexes on foreign keys. Additionally:
- `idx_submission_assign_student ON submission(assignment_id, student_id)` — composite for filtered queries

### What Each Handler Queries

| Page | Handler | Queries per load |
|------|---------|-----------------|
| Teacher dashboard | `AdminDashboard` | 1 (ListClassrooms with subquery counts) |
| Classroom detail | `AdminClassroom` | 7 (classroom + students + categories + resources + assignments + quizzes + allowed_students + live_session) |
| Analytics — quizzes | `AdminAnalytics` sub=quizzes | 1-3 (quiz stats + optional question analytics + student breakdown) |
| Analytics — assignments | `AdminAnalytics` sub=assignments | 1-2 (assignment stats with inline grade bins + optional missing students) |
| Analytics — students | `AdminAnalytics` sub=students | 1 (single CTE query) |
| Analytics — trends | `AdminAnalytics` sub=trends | 3 (quiz trends + assignment trends + timing stats) |
| Analytics — risk | `AdminAnalytics` sub=risk | 1 (single CTE query) |
| Analytics — live | `AdminAnalytics` sub=live | 2-3 (session history + attendance rates + optional session detail) |
| Submissions list | `ViewSubmissions` | 2 (count + paginated list) |
| Quiz attempts | `EditQuiz` | 3-4 (quiz + questions + paginated attempts count + attempts) |
| Student classroom | `StudentClassroom` | 4 (classroom + resources + assignments + quizzes) |
| Student quiz | `StudentQuiz` | 3-4 (quiz + questions + attempt count + optional existing attempt) |

---

## What's Already Optimized

These were fixed in March 2026. Don't redo them:

### N+1 Queries — Fixed

| Function | Before | After |
|----------|--------|-------|
| `GetStudentRosterAnalytics` | 2-3 queries per student in a loop | 1 CTE query with LEFT JOINs |
| `GetAtRiskStudents` | 3 queries per student in a loop | 1 CTE query with ROW_NUMBER window function |
| `GetAssignmentAnalytics` | 1 extra query per assignment for grade bins | Inline `COUNT(*) FILTER (WHERE ...)` |
| `GetSubmissionTimingStats` | 1-3 queries per assignment for timing | Inline `COUNT(*) FILTER (WHERE ...)` |

### Transactions — Fixed

| Function | What it wraps |
|----------|--------------|
| `CreateLiveSession` | UPDATE (deactivate old) + DELETE (cleanup inactive) + INSERT (new session) |
| `EndLiveSession` | UPDATE (attendance left_at) + UPDATE (session active=false) |

### Pagination — Added

| Endpoint | Function | Default page size |
|----------|----------|-------------------|
| Submissions list | `ListSubmissionsPaged` | 50 |
| Quiz attempts | `ListQuizAttemptsPaged` | 50 |

### Static Assets — Self-hosted

All CSS/JS served from `/static/`. No CDN dependencies. Tailwind compiled to static CSS.

---

## Step 1: Query Timeouts

**When:** If you ever see a hung connection or a query running for >10 seconds.

**What to do:** In `main.go`, replace the bare `pgxpool.New` call with a configured pool:

```go
// In main.go, replace:
//   pool, err := pgxpool.New(context.Background(), dbURL)
// With:

config, err := pgxpool.ParseConfig(dbURL)
if err != nil {
    log.Fatalf("DB config failed: %v", err)
}
config.MaxConns = 20                           // max connections (default 4)
config.MaxConnLifetime = 1 * time.Hour         // recycle connections every hour
config.MaxConnIdleTime = 30 * time.Minute      // close idle connections after 30min
config.HealthCheckPeriod = 1 * time.Minute     // check connection health every minute

pool, err := pgxpool.NewWithConfig(context.Background(), config)
```

Then for heavy analytics queries, wrap them with a timeout context:

```go
// In store functions that run CTEs:
ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
defer cancel()
```

**Estimated time:** 15 minutes.

**Why it's not done yet:** The default pgxpool config is fine for single-digit teachers. `c.Request.Context()` already cancels queries when HTTP clients disconnect. Explicit timeouts only matter when you have enough concurrent queries to exhaust the connection pool.

---

## Step 2: Rewrite GetQuestionAnalytics

**When:** A teacher with 200+ students clicks on a specific quiz in analytics and it takes >1 second.

**What's wrong:** `GetQuestionAnalytics` in `store/store.go` (~line 1270) calls `ListQuizAttempts` which loads ALL quiz attempts into Go memory (including full JSONB `answers` and `file_answers`), then loops through them to compute per-question stats. With 200 students × 3 attempts × 50 questions, that's 600 JSON objects parsed into RAM and 30,000 loop iterations.

**How to fix:** Rewrite it as a SQL query that does the aggregation in PostgreSQL:

```sql
-- For MCQ/true_false/fill_blank questions:
SELECT
    qq.id AS question_id,
    qq.content,
    qq.question_type,
    qq.points,
    qq.correct_answer,
    COUNT(DISTINCT qa.id) FILTER (WHERE qa.finished_at IS NOT NULL) AS total_count,
    COUNT(DISTINCT qa.id) FILTER (
        WHERE qa.finished_at IS NOT NULL
        AND LOWER(TRIM(qa.answers->>qq.id::text)) = LOWER(TRIM(qq.correct_answer))
    ) AS correct_count
FROM quiz_question qq
LEFT JOIN quiz_attempt qa ON qa.quiz_id = qq.quiz_id AND qa.finished_at IS NOT NULL
WHERE qq.quiz_id = $1
GROUP BY qq.id, qq.content, qq.question_type, qq.points, qq.correct_answer
ORDER BY qq.sort_order
```

The tricky part is the "most common wrong answer" — that requires a subquery or a mode aggregate. You could do it with:

```sql
-- Separate query for common wrong answers per question:
SELECT DISTINCT ON (qq.id)
    qq.id AS question_id,
    qa.answers->>qq.id::text AS wrong_answer,
    COUNT(*) AS cnt
FROM quiz_question qq
JOIN quiz_attempt qa ON qa.quiz_id = qq.quiz_id AND qa.finished_at IS NOT NULL
WHERE qq.quiz_id = $1
  AND qa.answers->>qq.id::text IS NOT NULL
  AND qa.answers->>qq.id::text != ''
  AND LOWER(TRIM(qa.answers->>qq.id::text)) != LOWER(TRIM(qq.correct_answer))
GROUP BY qq.id, qa.answers->>qq.id::text
ORDER BY qq.id, cnt DESC
```

**Also fix `GetQuizStudentBreakdown`:** Currently it's just `return s.ListQuizAttempts(ctx, quizID)` — loads all attempts with full JSONB. For the student breakdown view, you only need student name, score, max_score, reviewed status, timestamps. Write a lighter query:

```sql
SELECT a.id, a.student_id, s.name, a.score, a.max_score, a.reviewed, a.started_at, a.finished_at
FROM quiz_attempt a JOIN student s ON a.student_id = s.id
WHERE a.quiz_id = $1 ORDER BY a.started_at DESC
```

This skips loading the heavy `answers` and `file_answers` JSONB columns.

**Estimated time:** 1-2 hours.

---

## Step 3: Composite Indexes

**When:** You have 10,000+ quiz attempts or 5,000+ live attendance records.

**What to add to `db/schema.sql`:**

```sql
-- Quiz attempt lookups by quiz + student (CountStudentAttempts, GetAllStudentAttempts)
CREATE INDEX IF NOT EXISTS idx_quiz_attempt_quiz_student
    ON quiz_attempt(quiz_id, student_id);

-- Live attendance "already joined" check (JoinLiveSession duplicate check)
CREATE INDEX IF NOT EXISTS idx_live_attendance_session_student
    ON live_attendance(live_session_id, student_id);

-- Quiz attempt finished lookups (analytics queries filter on finished_at IS NOT NULL)
CREATE INDEX IF NOT EXISTS idx_quiz_attempt_quiz_finished
    ON quiz_attempt(quiz_id, finished_at)
    WHERE finished_at IS NOT NULL;
```

**How to apply:** Just add these lines to `db/schema.sql` before the closing section. They use `CREATE INDEX IF NOT EXISTS` so they're safe to run repeatedly. The app runs `schema.sql` on startup.

**Estimated time:** 5 minutes.

**Why it's not done yet:** With <1,000 rows in these tables, PostgreSQL does a sequential scan in <1ms. Indexes start mattering at 10,000+ rows.

---

## Step 4: Redis Caching

**When:** Analytics pages take >500ms, or multiple teachers are hitting analytics simultaneously.

### 4a. Add Redis to Docker Compose

Add this to `docker-compose.yml`:

```yaml
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
```

Add `redis_data:` to the `volumes:` section at the bottom.

Add to app environment:
```yaml
      - REDIS_URL=redis://redis:6379
```

### 4b. Add Redis Client to Go

```bash
go get github.com/redis/go-redis/v9
```

In `main.go`:
```go
import "github.com/redis/go-redis/v9"

// After pool setup:
redisURL := envOr("REDIS_URL", "redis://localhost:6379")
opt, _ := redis.ParseURL(redisURL)
rdb := redis.NewClient(opt)
defer rdb.Close()
```

### 4c. Create a Cache Helper

Create `cache/cache.go`:

```go
package cache

import (
    "context"
    "encoding/json"
    "time"
    "github.com/redis/go-redis/v9"
)

type Cache struct {
    rdb *redis.Client
}

func New(rdb *redis.Client) *Cache {
    return &Cache{rdb: rdb}
}

// GetOrSet checks cache, returns cached value if present,
// otherwise calls fn(), caches the result, and returns it.
func GetOrSet[T any](c *Cache, ctx context.Context, key string, ttl time.Duration, fn func() (T, error)) (T, error) {
    var result T

    // Try cache
    data, err := c.rdb.Get(ctx, key).Bytes()
    if err == nil {
        if json.Unmarshal(data, &result) == nil {
            return result, nil
        }
    }

    // Cache miss — call the real function
    result, err = fn()
    if err != nil {
        return result, err
    }

    // Store in cache (ignore errors — cache is best-effort)
    if encoded, err := json.Marshal(result); err == nil {
        c.rdb.Set(ctx, key, encoded, ttl)
    }

    return result, nil
}

// Invalidate removes a cached key (call when data changes)
func (c *Cache) Invalidate(ctx context.Context, pattern string) {
    keys, _ := c.rdb.Keys(ctx, pattern).Result()
    if len(keys) > 0 {
        c.rdb.Del(ctx, keys...)
    }
}
```

### 4d. Which Queries to Cache and How Long

| Function | Cache key pattern | TTL | Invalidate when |
|----------|------------------|-----|-----------------|
| `GetStudentRosterAnalytics` | `analytics:roster:{classroomID}` | 60s | Student submits, quiz taken, grade saved |
| `GetAtRiskStudents` | `analytics:risk:{classroomID}` | 60s | Same as above |
| `GetAssignmentAnalytics` | `analytics:assign:{classroomID}` | 60s | Submission created, graded |
| `GetQuizAnalytics` | `analytics:quiz:{classroomID}` | 60s | Quiz attempt finished, reviewed |
| `GetQuestionAnalytics` | `analytics:question:{quizID}` | 60s | Quiz attempt finished |
| `GetSubmissionTimingStats` | `analytics:timing:{classroomID}` | 120s | Submission created |
| `GetSessionHistory` | `analytics:sessions:{classroomID}` | 300s | Live session ended |

### 4e. Invalidation Strategy

The simplest approach: invalidate by classroom. When anything changes in a classroom (submission, quiz attempt, grade), do:

```go
cache.Invalidate(ctx, "analytics:*:"+classroomID)
```

This clears all analytics caches for that classroom. A 60-second TTL means even without invalidation, data is at most 1 minute stale.

**Estimated time:** 4-6 hours (includes testing).

**Why it's not done yet:** At current scale, every analytics CTE query returns in <100ms. Caching adds complexity (invalidation bugs, stale data) for no visible improvement.

---

## Step 5: Object Storage

**When:** You need multiple server instances, or you want off-server backup of student files.

### Option A: Cloudflare R2 (Recommended for Algeria)

R2 has **zero egress fees** — students downloading files costs nothing. S3-compatible API.

1. Create R2 bucket at dash.cloudflare.com
2. Get API credentials (access key ID + secret)
3. Add to `.env`:
   ```
   STORAGE_TYPE=r2
   R2_ACCOUNT_ID=your-account-id
   R2_ACCESS_KEY=your-access-key
   R2_SECRET_KEY=your-secret-key
   R2_BUCKET=teachhub-uploads
   ```

4. Add Go S3 client:
   ```bash
   go get github.com/aws/aws-sdk-go-v2/service/s3
   go get github.com/aws/aws-sdk-go-v2/credentials
   ```

5. Create `storage/storage.go` with an interface:
   ```go
   type Storage interface {
       Upload(ctx context.Context, key string, reader io.Reader, contentType string) error
       GetURL(key string) string
       Delete(ctx context.Context, key string) error
   }
   ```

6. Implement `LocalStorage` (current behavior) and `R2Storage` (S3 client).

7. Change all handlers that use `h.UploadDir` + `filepath.Join` to use the `Storage` interface instead.

**Files that need changing:**
- `handlers/admin.go` — `UploadResource`, `DeleteResource`, `DownloadResource`, `UploadTeacherPic`
- `handlers/student.go` — `SubmitAssignment`, `SubmitQuiz` (file upload questions)
- `handlers/live.go` — `LiveUploadImage`, `LiveUploadFile`
- `main.go` — `r.Static("/uploads", uploadDir)` needs to become a proxy or signed-URL redirect

**Migration of existing files:** Write a one-time script that uploads everything from `uploads/` to R2.

**Estimated time:** 6-8 hours.

### Option B: S3-Compatible Mount (Simpler)

Use `s3fs` or `goofys` to mount R2/S3 bucket as a local filesystem. The app code doesn't change at all — it still writes to `./uploads/`. The OS handles syncing to object storage.

```bash
# On the VPS:
apt install s3fs
echo "ACCESS_KEY:SECRET_KEY" > ~/.passwd-s3fs
chmod 600 ~/.passwd-s3fs
s3fs teachhub-uploads /app/uploads -o url=https://ACCOUNT_ID.r2.cloudflarestorage.com -o passwd_file=~/.passwd-s3fs
```

**Pros:** Zero code changes.  
**Cons:** Higher latency on file writes, FUSE overhead, needs careful Docker volume config.

**Estimated time:** 1-2 hours (but harder to debug issues).

---

## Step 6: Paginate Remaining Lists

**When:** A teacher has 200+ resources or 300+ students in one classroom.

These endpoints currently return all rows:

| Function | Used by | Add pagination? |
|----------|---------|----------------|
| `ListClassroomStudents` | `AdminClassroom` handler | Yes if >200 students per classroom |
| `ListResources` | `AdminClassroom` + `StudentClassroom` | Yes if >100 resources |
| `ListAssignments` | `AdminClassroom` + `StudentClassroom` | Unlikely to need it (<50 per year) |
| `ListQuizzes` | `AdminClassroom` + `StudentClassroom` | Unlikely to need it (<30 per year) |
| `ListAllowedStudents` | `AdminClassroom` | Yes if >200 allowed students |

**Pattern to follow:** Same as `ListSubmissionsPaged` and `ListQuizAttemptsPaged` — add a `Paged` variant that takes `limit, offset int` and returns `(items, totalCount, error)`. Update the handler to parse `?page=N`, compute offset, pass pagination data to the template.

**Estimated time:** 1-2 hours per endpoint.

---

## Step 7: Redis Rate Limiter

**When:** You run 2+ app instances behind a load balancer.

The current rate limiter in `middleware/middleware.go` uses an in-memory `sync.Mutex + map`. This works perfectly for a single instance. If you ever run multiple instances, replace it with Redis:

```go
func RateLimitRedis(rdb *redis.Client, max int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method != "POST" {
            c.Next()
            return
        }
        ip := c.ClientIP()
        key := "ratelimit:" + ip

        count, _ := rdb.Incr(c.Request.Context(), key).Result()
        if count == 1 {
            rdb.Expire(c.Request.Context(), key, window)
        }

        if count > int64(max) {
            ttl, _ := rdb.TTL(c.Request.Context(), key).Result()
            mins := int(ttl.Minutes()) + 1
            c.HTML(http.StatusTooManyRequests, "error_rate_limit.html", gin.H{"Minutes": mins})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

**Estimated time:** 30 minutes (if Redis is already in the stack from Step 4).

---

## What Does NOT Need Fixing

These were raised in a code review but are either wrong or irrelevant:

### "LiveKit data channel abuse / whiteboard is O(n) per participant"

**Not true.** The whiteboard uses LiveKit's SFU `publishData` which sends ONE message from the teacher — LiveKit server fans it out to N students. The teacher's upload bandwidth is O(1), not O(n). Move events use `reliable: false` (UDP) for real-time drawing, which is correct. Completed strokes send as a single `wb-stroke` message. The `sendWbStateTo` function for late joiners sends to ONE specific student via `destinationIdentities`, not to everyone.

### "muteAllStudents sends individual commands"

**Wrong.** `muteAllStudents()` calls `broadcastCommand('mute-all')` which calls `room.localParticipant.publishData(...)` once. LiveKit SFU relays it to all participants.

### "CookieStore session data in cookies is a problem"

**It's not.** `gorilla/sessions` CookieStore stores session data signed and encrypted in the cookie. The session contains only `admin_id` (an integer) or `student_id` (an integer). Total cookie size is ~100 bytes. This is fine and used by many production Go applications.

### "Need pagination on resources/assignments/quizzes"

**Not at current scale.** A teacher with 30 resources and 15 quizzes does not need pagination. The templates render these lists in <5ms. Only worth adding if a teacher reports slow loading with 200+ items.

---

## Production Checklist (Current State)

- [x] HTTPS via Caddy auto-TLS
- [x] Security headers (X-Content-Type-Options, X-Frame-Options, HSTS, CSP-adjacent)
- [x] CSRF protection on all POST forms
- [x] Rate limiting on login endpoints
- [x] Session cookies: HttpOnly, Secure, SameSite=Lax
- [x] Non-root Docker user
- [x] Health checks on all containers
- [x] Database schema runs on startup (idempotent with IF NOT EXISTS)
- [x] N+1 queries fixed on all analytics
- [x] Transactions on multi-step DB operations
- [x] Pagination on heavy list endpoints (submissions, quiz attempts)
- [x] Self-hosted static assets (Tailwind CSS compiled, htmx, livekit-client)
- [x] Bilingual (English + French)
- [x] Production secrets required in release mode (SESSION_SECRET, ADMIN_PASS, PLATFORM_PASS, BASE_URL)

---

## Deploy Commands Reference

```bash
# Local: compile Tailwind after template changes
npx tailwindcss -i static/css/input.css -o static/css/style.css --minify

# Push and deploy
cd F:\teachhub
git add -A && git commit -m "description" && git push origin main
ssh root@144.91.77.206 "cd /opt/teachhub && git pull origin main && docker compose up -d --build app"

# Check logs
ssh root@144.91.77.206 "docker compose -f /opt/teachhub/docker-compose.yml logs -f app --tail 50"

# Check if app is healthy
curl -sI https://teachhub.chickenkiller.com/
```

## Access URLs

| What | URL | Credentials |
|------|-----|-------------|
| Student portal | https://teachhub.chickenkiller.com | (students join via link) |
| Teacher login | https://teachhub.chickenkiller.com/admin/login | (per-teacher credentials) |
| Platform admin | https://teachhub.chickenkiller.com/ctrl-p-8x3kf/login | owner / pLKr37NupmT7c9OJBvzJ |
