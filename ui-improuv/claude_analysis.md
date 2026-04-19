You're right — the current code assumes individual teachers signing up through `/apply`, getting approved by *you* (platform admin), then running their own classroom. There's no notion of a center owning teachers. Here's the full restructure.

---

## 1. The conceptual shift

**Today:**
```
platform_admin (you)
  └─ admin (teacher) — standalone, approved by you
       └─ classroom
            └─ students
```

**Target:**
```
platform_admin (you)
  └─ center (owned by a center_owner admin)
       ├─ admin (teacher, role='owner' or 'teacher')
       │    └─ classroom
       │         └─ students + invoices
       └─ billing, aggregated analytics
```

Key changes:
- `admin` gets a `role` (`owner` / `teacher`) and a `center_id`
- New `center` table
- New `student_invoice` table
- `/apply` flow now creates *centers*, not individual teachers
- Platform admin (you) approves centers, not teachers
- Center owner onboards their own teachers — no platform approval needed
- Billing is per-center (subscription + per-seat)

---

## 2. Database migrations

### Migration 010 — center + role + teacher scoping

```sql
-- db/migrations/010_center.sql

CREATE TABLE IF NOT EXISTS center (
    id                  SERIAL PRIMARY KEY,
    name                TEXT NOT NULL,
    owner_admin_id      INT REFERENCES admin(id) ON DELETE SET NULL,
    address             TEXT NOT NULL DEFAULT '',
    city                TEXT NOT NULL DEFAULT '',
    country             TEXT NOT NULL DEFAULT 'FR',
    phone               TEXT NOT NULL DEFAULT '',
    email               TEXT NOT NULL DEFAULT '',
    logo_path           TEXT NOT NULL DEFAULT '',
    -- subscription: billed to the center, not the individual teacher
    subscription_status TEXT NOT NULL DEFAULT 'trial' 
                        CHECK (subscription_status IN ('trial','active','expired','suspended','cancelled')),
    subscription_start  TIMESTAMPTZ,
    subscription_end    TIMESTAMPTZ,
    seat_count          INT NOT NULL DEFAULT 3,       -- seats purchased
    price_per_seat      NUMERIC(10,2) NOT NULL DEFAULT 10.00,
    trial_ends_at       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_center_owner ON center(owner_admin_id);
CREATE INDEX IF NOT EXISTS idx_center_status ON center(subscription_status);

-- Extend admin table with role + center scoping
ALTER TABLE admin ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'teacher'
    CHECK (role IN ('owner','teacher'));
ALTER TABLE admin ADD COLUMN IF NOT EXISTS center_id INT REFERENCES center(id) ON DELETE CASCADE;
ALTER TABLE admin ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT true;

CREATE INDEX IF NOT EXISTS idx_admin_center ON admin(center_id);
CREATE INDEX IF NOT EXISTS idx_admin_role ON admin(role);

-- Rework teacher_application into center_application
ALTER TABLE teacher_application RENAME TO center_application;
ALTER TABLE center_application ADD COLUMN IF NOT EXISTS center_name TEXT NOT NULL DEFAULT '';
ALTER TABLE center_application ADD COLUMN IF NOT EXISTS expected_teachers INT NOT NULL DEFAULT 1;
ALTER TABLE center_application ADD COLUMN IF NOT EXISTS expected_students INT NOT NULL DEFAULT 0;
-- admin.application_id still points here — rename column for clarity
ALTER TABLE admin RENAME COLUMN application_id TO center_application_id;

-- Payments are now against the center subscription, not the individual teacher
ALTER TABLE payment ADD COLUMN IF NOT EXISTS center_id INT REFERENCES center(id) ON DELETE CASCADE;
ALTER TABLE payment ADD COLUMN IF NOT EXISTS kind TEXT NOT NULL DEFAULT 'subscription'
    CHECK (kind IN ('subscription','other'));

-- Backfill: every existing admin becomes a solo-center of 1 seat
DO $$
DECLARE
    a RECORD;
    new_center_id INT;
BEGIN
    FOR a IN SELECT id, school_name, username, email FROM admin WHERE center_id IS NULL LOOP
        INSERT INTO center (name, owner_admin_id, email, subscription_status, subscription_start, seat_count)
        VALUES (
            COALESCE(NULLIF(a.school_name, ''), a.username || ' Center'),
            a.id,
            a.email,
            'active',
            NOW(),
            1
        )
        RETURNING id INTO new_center_id;
        
        UPDATE admin SET center_id = new_center_id, role = 'owner' WHERE id = a.id;
        
        -- Move any payments to be center-scoped
        UPDATE payment SET center_id = new_center_id WHERE teacher_id = a.id AND center_id IS NULL;
    END LOOP;
END $$;

-- Now enforce NOT NULL
ALTER TABLE admin ALTER COLUMN center_id SET NOT NULL;
```

### Migration 011 — student invoicing

```sql
-- db/migrations/011_student_billing.sql

-- Per-classroom session rate (teacher sets it, owner sees it)
ALTER TABLE classroom ADD COLUMN IF NOT EXISTS session_rate NUMERIC(10,2) NOT NULL DEFAULT 0;
ALTER TABLE classroom ADD COLUMN IF NOT EXISTS billing_enabled BOOLEAN NOT NULL DEFAULT false;

-- Per-student, per-month attendance-based invoice
CREATE TABLE IF NOT EXISTS student_invoice (
    id                SERIAL PRIMARY KEY,
    center_id         INT NOT NULL REFERENCES center(id) ON DELETE CASCADE,
    classroom_id      INT NOT NULL REFERENCES classroom(id) ON DELETE CASCADE,
    student_id        INT NOT NULL REFERENCES student(id) ON DELETE CASCADE,
    period_month      DATE NOT NULL,               -- YYYY-MM-01
    sessions_attended INT NOT NULL DEFAULT 0,
    rate_per_session  NUMERIC(10,2) NOT NULL,
    total_amount      NUMERIC(10,2) NOT NULL,
    status            TEXT NOT NULL DEFAULT 'unpaid'
                      CHECK (status IN ('unpaid','paid','cancelled')),
    paid_at           TIMESTAMPTZ,
    paid_method       TEXT DEFAULT '',             -- cash, card, ccp, etc.
    notes             TEXT NOT NULL DEFAULT '',
    generated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(classroom_id, student_id, period_month)
);

CREATE INDEX IF NOT EXISTS idx_inv_center_period ON student_invoice(center_id, period_month);
CREATE INDEX IF NOT EXISTS idx_inv_status ON student_invoice(status);
CREATE INDEX IF NOT EXISTS idx_inv_student ON student_invoice(student_id);

-- Parent view tracking (so owner can see "23 parents consulted this week")
CREATE TABLE IF NOT EXISTS parent_view_log (
    id          SERIAL PRIMARY KEY,
    parent_code TEXT NOT NULL,
    viewed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip          TEXT DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_pvl_code ON parent_view_log(parent_code);
CREATE INDEX IF NOT EXISTS idx_pvl_date ON parent_view_log(viewed_at);
```

Update `db/schema.sql` too — add the same `center` table, new columns, new indexes. Keep it in sync.

---

## 3. Store layer changes

### `store/store.go` — add models

```go
type Center struct {
    ID                 int
    Name               string
    OwnerAdminID       *int
    Address            string
    City               string
    Country            string
    Phone              string
    Email              string
    LogoPath           string
    SubscriptionStatus string
    SubscriptionStart  *time.Time
    SubscriptionEnd    *time.Time
    SeatCount          int
    PricePerSeat       float64
    TrialEndsAt        *time.Time
    CreatedAt          time.Time
    // computed
    TeacherCount  int
    StudentCount  int
    SeatsUsed     int
}

type StudentInvoice struct {
    ID               int
    CenterID         int
    ClassroomID      int
    StudentID        int
    StudentName      string  // joined
    ClassroomName    string  // joined
    PeriodMonth      time.Time
    SessionsAttended int
    RatePerSession   float64
    TotalAmount      float64
    Status           string
    PaidAt           *time.Time
    PaidMethod       string
    Notes            string
    GeneratedAt      time.Time
}
```

Extend the existing `Admin` struct:

```go
type Admin struct {
    // ... existing fields ...
    Role     string  // "owner" or "teacher"
    CenterID int
    Active   bool
}
```

### Center queries (new file: `store/center.go`)

```go
package store

import (
    "context"
    "fmt"
    "time"
)

func (s *Store) CreateCenter(ctx context.Context, name, email string, ownerID int) (int, error) {
    var id int
    err := s.DB.QueryRow(ctx, `
        INSERT INTO center (name, email, owner_admin_id, subscription_status, trial_ends_at)
        VALUES ($1, $2, $3, 'trial', NOW() + INTERVAL '30 days')
        RETURNING id`, name, email, ownerID).Scan(&id)
    return id, err
}

func (s *Store) GetCenter(ctx context.Context, id int) (*Center, error) {
    c := &Center{}
    err := s.DB.QueryRow(ctx, `
        SELECT id, name, owner_admin_id, address, city, country, phone, email, logo_path,
               subscription_status, subscription_start, subscription_end, seat_count, price_per_seat,
               trial_ends_at, created_at,
               (SELECT COUNT(*) FROM admin WHERE center_id=$1 AND active=true) AS teacher_count,
               (SELECT COUNT(DISTINCT cs.student_id) FROM classroom_student cs 
                JOIN classroom cl ON cl.id=cs.classroom_id 
                WHERE cl.admin_id IN (SELECT id FROM admin WHERE center_id=$1) 
                AND cs.status='approved') AS student_count
        FROM center WHERE id=$1`, id).
        Scan(&c.ID, &c.Name, &c.OwnerAdminID, &c.Address, &c.City, &c.Country, &c.Phone, &c.Email, &c.LogoPath,
            &c.SubscriptionStatus, &c.SubscriptionStart, &c.SubscriptionEnd, &c.SeatCount, &c.PricePerSeat,
            &c.TrialEndsAt, &c.CreatedAt, &c.TeacherCount, &c.StudentCount)
    if err != nil {
        return nil, err
    }
    c.SeatsUsed = c.TeacherCount
    return c, nil
}

// ListCenterTeachers lists all teachers in a center (owner + teachers)
func (s *Store) ListCenterTeachers(ctx context.Context, centerID int) ([]Admin, error) {
    rows, err := s.DB.Query(ctx, `
        SELECT id, username, email, role, active, last_login_at, created_at,
               (SELECT COUNT(*) FROM classroom WHERE admin_id=admin.id) AS classroom_count
        FROM admin WHERE center_id=$1 ORDER BY role DESC, username`, centerID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var list []Admin
    for rows.Next() {
        var a Admin
        var cc int
        if err := rows.Scan(&a.ID, &a.Username, &a.Email, &a.Role, &a.Active,
            &a.LastLoginAt, &a.CreatedAt, &cc); err != nil {
            return nil, err
        }
        list = append(list, a)
    }
    return list, nil
}

// CreateTeacherInCenter — owner creates a teacher seat. Returns the temp password.
func (s *Store) CreateTeacherInCenter(ctx context.Context, centerID int, username, email, hashedPassword, plaintextPassword string) (int, error) {
    // Check seat availability
    var seatCount, used int
    s.DB.QueryRow(ctx, `SELECT seat_count FROM center WHERE id=$1`, centerID).Scan(&seatCount)
    s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM admin WHERE center_id=$1 AND active=true`, centerID).Scan(&used)
    if used >= seatCount {
        return 0, fmt.Errorf("seat_limit")
    }
    var id int
    err := s.DB.QueryRow(ctx, `
        INSERT INTO admin (center_id, username, password, pending_password, email, role, created_by_platform)
        VALUES ($1, $2, $3, $4, $5, 'teacher', true)
        RETURNING id`, centerID, username, hashedPassword, plaintextPassword, email).Scan(&id)
    return id, err
}

func (s *Store) DeactivateTeacher(ctx context.Context, adminID, centerID int) error {
    _, err := s.DB.Exec(ctx, `UPDATE admin SET active=false WHERE id=$1 AND center_id=$2 AND role='teacher'`, adminID, centerID)
    return err
}

func (s *Store) UpdateCenterSubscription(ctx context.Context, centerID int, status string, endDate *time.Time, seats int, price float64) error {
    _, err := s.DB.Exec(ctx, `
        UPDATE center SET subscription_status=$2, subscription_end=$3, seat_count=$4, price_per_seat=$5
        WHERE id=$1`, centerID, status, endDate, seats, price)
    return err
}
```

### Center dashboard analytics (`store/center_analytics.go`)

```go
package store

import (
    "context"
    "time"
)

type CenterDashboardStats struct {
    TeacherCount      int
    ActiveStudents    int
    SessionsThisMonth int
    RevenueThisMonth  float64
    ParentViewsWeek   int
    AtRiskCount       int
}

func (s *Store) GetCenterDashboardStats(ctx context.Context, centerID int) (*CenterDashboardStats, error) {
    st := &CenterDashboardStats{}
    s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM admin WHERE center_id=$1 AND active=true`, centerID).Scan(&st.TeacherCount)
    s.DB.QueryRow(ctx, `
        SELECT COUNT(DISTINCT cs.student_id) FROM classroom_student cs
        JOIN classroom cl ON cl.id=cs.classroom_id
        JOIN admin a ON a.id=cl.admin_id
        WHERE a.center_id=$1 AND cs.status='approved'`, centerID).Scan(&st.ActiveStudents)
    s.DB.QueryRow(ctx, `
        SELECT COUNT(*) FROM live_session ls
        JOIN classroom cl ON cl.id=ls.classroom_id
        JOIN admin a ON a.id=cl.admin_id
        WHERE a.center_id=$1 AND ls.created_at >= date_trunc('month', NOW())`, centerID).Scan(&st.SessionsThisMonth)
    s.DB.QueryRow(ctx, `
        SELECT COALESCE(SUM(total_amount),0) FROM student_invoice
        WHERE center_id=$1 AND status='paid' AND paid_at >= date_trunc('month', NOW())`, centerID).Scan(&st.RevenueThisMonth)
    s.DB.QueryRow(ctx, `
        SELECT COUNT(*) FROM parent_view_log pvl
        JOIN classroom_student cs ON cs.parent_code = pvl.parent_code
        JOIN classroom cl ON cl.id = cs.classroom_id
        JOIN admin a ON a.id = cl.admin_id
        WHERE a.center_id=$1 AND pvl.viewed_at >= NOW() - INTERVAL '7 days'`, centerID).Scan(&st.ParentViewsWeek)
    return st, nil
}

type TeacherPerformanceRow struct {
    TeacherID        int
    Username         string
    Email            string
    ClassroomCount   int
    StudentCount     int
    AvgQuizPct       float64
    AvgAttendancePct float64
    SessionsThisMonth int
    LastActive       *time.Time
}

func (s *Store) GetCenterTeacherPerformance(ctx context.Context, centerID int) ([]TeacherPerformanceRow, error) {
    rows, err := s.DB.Query(ctx, `
        SELECT a.id, a.username, a.email,
            (SELECT COUNT(*) FROM classroom WHERE admin_id=a.id),
            (SELECT COUNT(DISTINCT cs.student_id) FROM classroom_student cs 
             JOIN classroom cl ON cl.id=cs.classroom_id 
             WHERE cl.admin_id=a.id AND cs.status='approved'),
            COALESCE((SELECT AVG(CASE WHEN qa.max_score>0 THEN qa.score*100.0/qa.max_score ELSE 0 END) 
             FROM quiz_attempt qa 
             JOIN quiz q ON q.id=qa.quiz_id 
             JOIN classroom cl ON cl.id=q.classroom_id 
             WHERE cl.admin_id=a.id AND qa.finished_at IS NOT NULL), 0),
            COALESCE((SELECT COUNT(*) FROM live_session ls 
             JOIN classroom cl ON cl.id=ls.classroom_id 
             WHERE cl.admin_id=a.id AND ls.created_at >= date_trunc('month', NOW())), 0),
            a.last_login_at
        FROM admin a
        WHERE a.center_id=$1 AND a.role='teacher' AND a.active=true
        ORDER BY a.username`, centerID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var list []TeacherPerformanceRow
    for rows.Next() {
        var r TeacherPerformanceRow
        if err := rows.Scan(&r.TeacherID, &r.Username, &r.Email, &r.ClassroomCount, &r.StudentCount,
            &r.AvgQuizPct, &r.SessionsThisMonth, &r.LastActive); err != nil {
            return nil, err
        }
        list = append(list, r)
    }
    return list, nil
}
```

### Billing (`store/invoicing.go`)

```go
package store

import (
    "context"
    "time"
)

// GenerateMonthlyInvoices — run for a given center for the previous month.
// Idempotent (UNIQUE constraint on classroom_id, student_id, period_month).
func (s *Store) GenerateMonthlyInvoices(ctx context.Context, centerID int, period time.Time) (int, error) {
    periodStart := time.Date(period.Year(), period.Month(), 1, 0, 0, 0, 0, time.UTC)
    periodEnd := periodStart.AddDate(0, 1, 0)

    // For each classroom in the center with billing_enabled and session_rate > 0,
    // count distinct sessions attended per student in that month, insert invoice rows.
    tag, err := s.DB.Exec(ctx, `
        INSERT INTO student_invoice (center_id, classroom_id, student_id, period_month,
                                     sessions_attended, rate_per_session, total_amount)
        SELECT $1,
               cl.id,
               la.student_id,
               $2::date,
               COUNT(DISTINCT la.live_session_id),
               cl.session_rate,
               COUNT(DISTINCT la.live_session_id) * cl.session_rate
        FROM live_attendance la
        JOIN live_session ls ON ls.id = la.live_session_id
        JOIN classroom cl ON cl.id = ls.classroom_id
        JOIN admin a ON a.id = cl.admin_id
        WHERE a.center_id = $1
          AND cl.billing_enabled = true
          AND cl.session_rate > 0
          AND ls.created_at >= $2
          AND ls.created_at < $3
          AND ls.duration_minutes >= 5  -- skip aborted sessions
        GROUP BY cl.id, la.student_id, cl.session_rate
        HAVING COUNT(DISTINCT la.live_session_id) > 0
        ON CONFLICT (classroom_id, student_id, period_month) DO NOTHING`,
        centerID, periodStart, periodEnd)
    if err != nil {
        return 0, err
    }
    return int(tag.RowsAffected()), nil
}

func (s *Store) ListCenterInvoices(ctx context.Context, centerID int, period time.Time, status string) ([]StudentInvoice, error) {
    periodStart := time.Date(period.Year(), period.Month(), 1, 0, 0, 0, 0, time.UTC)
    q := `SELECT si.id, si.center_id, si.classroom_id, si.student_id, s.name, cl.name,
                 si.period_month, si.sessions_attended, si.rate_per_session, si.total_amount,
                 si.status, si.paid_at, si.paid_method, si.notes, si.generated_at
          FROM student_invoice si
          JOIN student s ON s.id = si.student_id
          JOIN classroom cl ON cl.id = si.classroom_id
          WHERE si.center_id = $1 AND si.period_month = $2`
    args := []interface{}{centerID, periodStart}
    if status != "" && status != "all" {
        q += ` AND si.status = $3`
        args = append(args, status)
    }
    q += ` ORDER BY si.status, s.name`
    rows, err := s.DB.Query(ctx, q, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var list []StudentInvoice
    for rows.Next() {
        var inv StudentInvoice
        if err := rows.Scan(&inv.ID, &inv.CenterID, &inv.ClassroomID, &inv.StudentID, &inv.StudentName, &inv.ClassroomName,
            &inv.PeriodMonth, &inv.SessionsAttended, &inv.RatePerSession, &inv.TotalAmount,
            &inv.Status, &inv.PaidAt, &inv.PaidMethod, &inv.Notes, &inv.GeneratedAt); err != nil {
            return nil, err
        }
        list = append(list, inv)
    }
    return list, nil
}

func (s *Store) MarkInvoicePaid(ctx context.Context, invoiceID, centerID int, method string) error {
    _, err := s.DB.Exec(ctx, `
        UPDATE student_invoice SET status='paid', paid_at=NOW(), paid_method=$3
        WHERE id=$1 AND center_id=$2`, invoiceID, centerID, method)
    return err
}

func (s *Store) CancelInvoice(ctx context.Context, invoiceID, centerID int) error {
    _, err := s.DB.Exec(ctx, `
        UPDATE student_invoice SET status='cancelled'
        WHERE id=$1 AND center_id=$2`, invoiceID, centerID)
    return err
}

// GetStudentInvoices — used by parent report to show "ce mois: 12 séances, 240€"
func (s *Store) GetStudentInvoices(ctx context.Context, studentID, classroomID int) ([]StudentInvoice, error) {
    rows, err := s.DB.Query(ctx, `
        SELECT id, center_id, classroom_id, student_id, '', '',
               period_month, sessions_attended, rate_per_session, total_amount,
               status, paid_at, paid_method, notes, generated_at
        FROM student_invoice
        WHERE student_id=$1 AND classroom_id=$2
        ORDER BY period_month DESC
        LIMIT 6`, studentID, classroomID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var list []StudentInvoice
    for rows.Next() {
        var inv StudentInvoice
        if err := rows.Scan(&inv.ID, &inv.CenterID, &inv.ClassroomID, &inv.StudentID, &inv.StudentName, &inv.ClassroomName,
            &inv.PeriodMonth, &inv.SessionsAttended, &inv.RatePerSession, &inv.TotalAmount,
            &inv.Status, &inv.PaidAt, &inv.PaidMethod, &inv.Notes, &inv.GeneratedAt); err != nil {
            return nil, err
        }
        list = append(list, inv)
    }
    return list, nil
}
```

---

## 4. Middleware — owner vs teacher

Add to `middleware/middleware.go`:

```go
// OwnerRequired ensures the admin has role='owner'
func OwnerRequired(db *store.Store) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, _ := c.Get("admin_id")
        aid, _ := id.(int)
        if aid == 0 {
            c.Redirect(http.StatusFound, "/admin/login")
            c.Abort()
            return
        }
        admin, err := db.GetAdminByID(c.Request.Context(), aid)
        if err != nil || admin.Role != "owner" {
            c.String(http.StatusForbidden, "Accès réservé au propriétaire du centre")
            c.Abort()
            return
        }
        c.Set("center_id", admin.CenterID)
        c.Next()
    }
}

// CenterContext injects center_id for any authenticated admin
func CenterContext(db *store.Store) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, _ := c.Get("admin_id")
        aid, _ := id.(int)
        if aid > 0 {
            admin, err := db.GetAdminByID(c.Request.Context(), aid)
            if err == nil {
                c.Set("center_id", admin.CenterID)
                c.Set("admin_role", admin.Role)
            }
        }
        c.Next()
    }
}
```

Update `AdminSubscriptionCheck` to check the **center's** subscription, not the teacher's:

```go
func AdminSubscriptionCheck(db *store.Store) gin.HandlerFunc {
    return func(c *gin.Context) {
        id, _ := c.Get("admin_id")
        aid, _ := id.(int)
        if aid == 0 { c.Next(); return }
        admin, err := db.GetAdminByID(c.Request.Context(), aid)
        if err != nil {
            ClearAdminSession(c); c.Redirect(http.StatusFound, "/admin/login"); c.Abort(); return
        }
        if !admin.Active {
            ClearAdminSession(c); c.Redirect(http.StatusFound, "/admin/login?error=deactivated"); c.Abort(); return
        }
        center, err := db.GetCenter(c.Request.Context(), admin.CenterID)
        if err != nil {
            ClearAdminSession(c); c.Redirect(http.StatusFound, "/admin/login"); c.Abort(); return
        }
        if center.SubscriptionStatus == "expired" || center.SubscriptionStatus == "suspended" || center.SubscriptionStatus == "cancelled" {
            ClearAdminSession(c); c.Redirect(http.StatusFound, "/admin/login?error="+center.SubscriptionStatus); c.Abort(); return
        }
        // Auto-expire trials
        if center.SubscriptionStatus == "trial" && center.TrialEndsAt != nil && center.TrialEndsAt.Before(time.Now()) {
            db.UpdateCenterSubscription(c.Request.Context(), center.ID, "expired", nil, center.SeatCount, center.PricePerSeat)
            ClearAdminSession(c); c.Redirect(http.StatusFound, "/admin/login?error=trial_expired"); c.Abort(); return
        }
        c.Next()
    }
}
```

---

## 5. Handlers — new flows

### Center application (replaces individual `/apply`)

`handlers/platform.go::ApplyPage` stays but the form now asks:

```html
<!-- apply.html — rename fields -->
<input name="center_name" placeholder="Nom du centre" required>
<input name="full_name" placeholder="Votre nom (propriétaire)" required>
<input name="email" placeholder="Email" required>
<input name="phone" placeholder="Téléphone" required>
<input name="city" placeholder="Ville" required>
<input name="expected_teachers" type="number" placeholder="Nombre de profs" required>
<input name="expected_students" type="number" placeholder="Nombre d'élèves estimé">
```

`ApplySubmit` stores into `center_application` with the new fields.

When you (platform admin) approve an application, `PlatformUpdateAppStatus` now:
1. Creates a `center` row (status=`trial`, trial_ends_at=NOW()+30d, seat_count=`expected_teachers`)
2. Creates an `admin` with `role='owner'`, `center_id=` new center id
3. Sets owner's `pending_password` so you can show it in the credentials page

Change in `handlers/platform.go::PlatformUpdateAppStatus`:

```go
if status == "approved" {
    app, _ := h.Store.GetCenterApplication(ctx, id)
    // Generate owner credentials
    username := strings.Split(app.Email, "@")[0]
    password := generatePassword(10)
    hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    
    // Create center first, then owner, then link
    centerID, err := h.Store.CreateCenter(ctx, app.CenterName, app.Email, 0)
    if err != nil {
        c.Redirect(http.StatusFound, h.pp("/applications/"+strconv.Itoa(id)+"?error=center"))
        return
    }
    // Owner admin
    ownerID, err := h.Store.CreateOwnerAdmin(ctx, centerID, username, string(hashed), password, app.Email, app.Phone)
    if err != nil {
        // handle username conflict
        username = fmt.Sprintf("%s%d", username, id)
        ownerID, _ = h.Store.CreateOwnerAdmin(ctx, centerID, username, string(hashed), password, app.Email, app.Phone)
    }
    // Link owner to center
    h.Store.DB.Exec(ctx, `UPDATE center SET owner_admin_id=$1 WHERE id=$2`, ownerID, centerID)
    h.Store.DB.Exec(ctx, `UPDATE center SET seat_count=$1 WHERE id=$2`, app.ExpectedTeachers, centerID)
    
    h.Store.UpdateApplicationStatus(ctx, id, status, notes)
    c.Redirect(http.StatusFound, h.pp("/applications/"+strconv.Itoa(id)+"/credentials"))
    return
}
```

Add `Store.CreateOwnerAdmin`:

```go
func (s *Store) CreateOwnerAdmin(ctx context.Context, centerID int, username, hashedPassword, plaintextPassword, email, phone string) (int, error) {
    var id int
    err := s.DB.QueryRow(ctx, `
        INSERT INTO admin (center_id, username, password, pending_password, email, phone, role, active, created_by_platform)
        VALUES ($1, $2, $3, $4, $5, $6, 'owner', true, true)
        RETURNING id`, centerID, username, hashedPassword, plaintextPassword, email, phone).Scan(&id)
    return id, err
}
```

### Center owner dashboard

New file: `handlers/center.go`

```go
package handlers

import (
    "net/http"
    "strconv"
    "strings"
    "teachhub/middleware"
    "time"

    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "crypto/rand"
    "math/big"
)

func centerID(c *gin.Context) int {
    if id, ok := c.Get("center_id"); ok {
        if v, ok := id.(int); ok { return v }
    }
    return 0
}

func (h *Handler) CenterDashboard(c *gin.Context) {
    cid := centerID(c)
    center, _ := h.Store.GetCenter(c.Request.Context(), cid)
    stats, _ := h.Store.GetCenterDashboardStats(c.Request.Context(), cid)
    teachers, _ := h.Store.GetCenterTeacherPerformance(c.Request.Context(), cid)
    atRisk, _ := h.Store.GetCenterAtRiskStudents(c.Request.Context(), cid)  // aggregate across classrooms

    h.render(c, "center_dashboard.html", gin.H{
        "Center":     center,
        "Stats":      stats,
        "Teachers":   teachers,
        "AtRisk":     atRisk,
        "IsOwner":    true,
    })
}

func (h *Handler) CenterTeachers(c *gin.Context) {
    cid := centerID(c)
    center, _ := h.Store.GetCenter(c.Request.Context(), cid)
    teachers, _ := h.Store.ListCenterTeachers(c.Request.Context(), cid)
    h.render(c, "center_teachers.html", gin.H{
        "Center":   center,
        "Teachers": teachers,
        "Saved":    c.Query("saved"),
        "Error":    c.Query("error"),
    })
}

func (h *Handler) CenterCreateTeacher(c *gin.Context) {
    cid := centerID(c)
    username := strings.ToLower(strings.TrimSpace(c.PostForm("username")))
    email := strings.TrimSpace(c.PostForm("email"))
    if username == "" || email == "" {
        c.Redirect(http.StatusFound, "/admin/center/teachers?error=missing")
        return
    }
    password := generatePassword(10)
    hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    _, err := h.Store.CreateTeacherInCenter(c.Request.Context(), cid, username, email, string(hashed), password)
    if err != nil {
        if err.Error() == "seat_limit" {
            c.Redirect(http.StatusFound, "/admin/center/teachers?error=seat_limit")
        } else {
            c.Redirect(http.StatusFound, "/admin/center/teachers?error=create")
        }
        return
    }
    c.Redirect(http.StatusFound, "/admin/center/teachers?saved=1")
}

func (h *Handler) CenterDeactivateTeacher(c *gin.Context) {
    cid := centerID(c)
    tid, _ := strconv.Atoi(c.Param("teacherId"))
    h.Store.DeactivateTeacher(c.Request.Context(), tid, cid)
    c.Redirect(http.StatusFound, "/admin/center/teachers?saved=1")
}

func (h *Handler) CenterBilling(c *gin.Context) {
    cid := centerID(c)
    center, _ := h.Store.GetCenter(c.Request.Context(), cid)
    periodStr := c.DefaultQuery("period", time.Now().Format("2006-01"))
    period, err := time.Parse("2006-01", periodStr)
    if err != nil { period = time.Now() }
    status := c.DefaultQuery("status", "all")
    invoices, _ := h.Store.ListCenterInvoices(c.Request.Context(), cid, period, status)
    h.render(c, "center_billing.html", gin.H{
        "Center":   center,
        "Invoices": invoices,
        "Period":   period,
        "Status":   status,
    })
}

func (h *Handler) CenterGenerateInvoices(c *gin.Context) {
    cid := centerID(c)
    periodStr := c.PostForm("period")
    period, err := time.Parse("2006-01", periodStr)
    if err != nil { period = time.Now().AddDate(0, -1, 0) }  // default: previous month
    count, _ := h.Store.GenerateMonthlyInvoices(c.Request.Context(), cid, period)
    c.Redirect(http.StatusFound, "/admin/center/billing?period="+periodStr+"&generated="+strconv.Itoa(count))
}

func (h *Handler) CenterMarkInvoicePaid(c *gin.Context) {
    cid := centerID(c)
    invID, _ := strconv.Atoi(c.Param("invoiceId"))
    method := c.DefaultPostForm("method", "cash")
    h.Store.MarkInvoicePaid(c.Request.Context(), invID, cid, method)
    c.Redirect(http.StatusFound, c.GetHeader("Referer"))
}

func (h *Handler) CenterSettings(c *gin.Context) {
    cid := centerID(c)
    center, _ := h.Store.GetCenter(c.Request.Context(), cid)
    h.render(c, "center_settings.html", gin.H{"Center": center, "Saved": c.Query("saved")})
}

func (h *Handler) CenterSettingsSave(c *gin.Context) {
    cid := centerID(c)
    name := strings.TrimSpace(c.PostForm("name"))
    address := strings.TrimSpace(c.PostForm("address"))
    city := strings.TrimSpace(c.PostForm("city"))
    phone := strings.TrimSpace(c.PostForm("phone"))
    email := strings.TrimSpace(c.PostForm("email"))
    h.Store.DB.Exec(c.Request.Context(), `
        UPDATE center SET name=$2, address=$3, city=$4, phone=$5, email=$6 WHERE id=$1`,
        cid, name, address, city, phone, email)
    c.Redirect(http.StatusFound, "/admin/center/settings?saved=1")
}
```

### Classroom billing toggle (teacher-side)

Add a "Tarif par séance" input in the classroom edit/create form. Add handler:

```go
func (h *Handler) UpdateClassroomBilling(c *gin.Context) {
    classID := h.ownsClassroom(c)
    if classID == 0 { return }
    rate, _ := strconv.ParseFloat(c.PostForm("session_rate"), 64)
    enabled := c.PostForm("billing_enabled") == "on"
    h.Store.DB.Exec(c.Request.Context(), `
        UPDATE classroom SET session_rate=$2, billing_enabled=$3 WHERE id=$1`,
        classID, rate, enabled)
    c.Redirect(http.StatusFound, "/admin/classroom/"+strconv.Itoa(classID))
}
```

### Parent report — log views + show invoice

In `handlers/parent.go::ParentReport`, add:

```go
// Log the view (fire and forget)
go func() {
    h.Store.DB.Exec(context.Background(),
        `INSERT INTO parent_view_log (parent_code, ip) VALUES ($1, $2)`,
        code, c.ClientIP())
}()

// Fetch invoices for the student
invoices, _ := h.Store.GetStudentInvoices(ctx, data.StudentID, data.ClassroomID)
// ... pass to template
```

Add a small section in `parent_report.html`:

```html
{{if .Invoices}}
<h3>Facturation</h3>
<table>
  {{range .Invoices}}
  <tr>
    <td>{{.PeriodMonth.Format "January 2006"}}</td>
    <td>{{.SessionsAttended}} séances × {{printf "%.0f" .RatePerSession}}€</td>
    <td>{{printf "%.2f" .TotalAmount}} €</td>
    <td>{{if eq .Status "paid"}}✓ Payé{{else}}À régler{{end}}</td>
  </tr>
  {{end}}
</table>
{{end}}
```

---

## 6. Routing changes

In your router setup (wherever routes are wired — probably `main.go`):

```go
// Center-owner-only routes
centerGroup := r.Group("/admin/center")
centerGroup.Use(middleware.AdminRequired())
centerGroup.Use(middleware.AdminSubscriptionCheck(store))
centerGroup.Use(middleware.OwnerRequired(store))
{
    centerGroup.GET("", h.CenterDashboard)
    centerGroup.GET("/teachers", h.CenterTeachers)
    centerGroup.POST("/teachers", h.CenterCreateTeacher)
    centerGroup.POST("/teachers/:teacherId/deactivate", h.CenterDeactivateTeacher)
    centerGroup.GET("/billing", h.CenterBilling)
    centerGroup.POST("/billing/generate", h.CenterGenerateInvoices)
    centerGroup.POST("/billing/:invoiceId/paid", h.CenterMarkInvoicePaid)
    centerGroup.GET("/settings", h.CenterSettings)
    centerGroup.POST("/settings", h.CenterSettingsSave)
}

// Regular teacher routes — stay the same, but AdminSubscriptionCheck now validates the center
```

After login, redirect owners to `/admin/center`, teachers to `/admin`:

```go
// In AdminLogin handler, after SetAdminSession
if admin.Role == "owner" {
    c.Redirect(http.StatusFound, "/admin/center")
} else {
    c.Redirect(http.StatusFound, "/admin")
}
```

---

## 7. Templates to build

You need these new templates (I'll leave detailed HTML to you, but here are the structures):

- `center_dashboard.html` — 4 top cards (Teachers, Students, Sessions this month, Revenue this month), at-risk table, teacher performance table, "view parent reports" counter
- `center_teachers.html` — list + create form + deactivate button
- `center_billing.html` — month picker, generate button, invoice table with mark-paid
- `center_settings.html` — center name, address, logo upload

Update existing:
- `admin_head.html` — add nav links for owners: Dashboard Centre / Profs / Facturation / Paramètres
- `admin_dashboard.html` — show a "Gérant:" breadcrumb if logged in via a center
- `apply.html` — rebrand to "Inscrire mon centre" with the new fields
- `admin_login.html` — error messages for `trial_expired`, `deactivated`

---

## 8. What you delete / deprecate

- The `Explore` / public teacher directory — hide until you have 20+ active centers. Comment out those routes or wrap in a feature flag.
- `/admin/profile` with `public_profile` toggle — disable the public part (keep the bio/subjects for internal use).
- `join_request` from `/explore/teacher/:id/join` — deprecated, students join via classroom code only.
- Teacher-level subscription fields (`admin.subscription_status`, `admin.subscription_start`, `admin.subscription_end`) — still in the table for now (don't drop), but stop reading them. Only read `center.subscription_*`.

---

## 9. Sequencing — exact order to ship

Do this in order, don't jump ahead:

1. **Day 1-2**: Migration 010 + store/center.go + middleware changes. Deploy. Verify existing teachers auto-migrated into solo-centers and can still log in.
2. **Day 3-4**: `CenterDashboard` + `center_dashboard.html` + routing. Test with your own account (promote yourself to `owner`).
3. **Day 5**: `CenterTeachers` + create/deactivate.
4. **Day 6**: Migration 011 + `classroom.session_rate` + store/invoicing.go.
5. **Day 7-8**: `CenterBilling` page + generate/mark-paid flow.
6. **Day 9**: Parent report updates (invoice section + view logging).
7. **Day 10**: Rework `/apply` to be center-focused + update `PlatformUpdateAppStatus`.
8. **Day 11**: Demo seeding for new centers (seed 5 students, 1 classroom, fake attendance so the dashboard is populated).
9. **Day 12**: Polish nav, error messages, empty states. Print 20 one-pagers.
10. **Day 13**: Walk into Cours des Possibles.

---

## 10. What I'd skip for now

To keep the scope tight for a sellable v1:

- Stripe integration for parent payment — not needed. Centers collect cash/CCP, mark as paid manually.
- Logo upload for centers — leave the field but defer the UI.
- Multi-classroom per student across teachers — existing model works.
- White-label parent reports — phase 2.
- AI session summaries — phase 2.

---

This gives you a sellable center product in ~2 weeks. The code is in good shape — you're mostly adding a layer on top, not rewriting. Want me to draft the `center_dashboard.html` template next, or the migration 010 full file with the exact backfill SQL?