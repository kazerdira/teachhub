# Billing Rework Spec — Migration 013
**Status:** Awaiting approval. No code written yet.  
**Covers:** Part A (delete student/parent billing) + Part B (add platform-to-center billing)

---

## Artifact 1 — `db/migrations/013_billing_rework.sql`

```sql
-- db/migrations/013_billing_rework.sql
-- Billing rework: remove student/parent billing, add platform-to-center invoicing.
-- Idempotent: safe to re-run. Uses IF EXISTS / IF NOT EXISTS / DO $$ throughout.

-- ═══════════════════════════════════════════════════════
-- PART A — Drop student/parent billing artifacts
-- ═══════════════════════════════════════════════════════

-- 1. student_invoice (center-to-student invoicing, removed: B2B only)
DROP TABLE IF EXISTS student_invoice CASCADE;

-- 2. parent_view_log (parent report view tracking, removed: no parent analytics)
DROP TABLE IF EXISTS parent_view_log CASCADE;

-- 3. classroom billing columns
ALTER TABLE classroom DROP COLUMN IF EXISTS session_rate;
ALTER TABLE classroom DROP COLUMN IF EXISTS billing_enabled;

-- ═══════════════════════════════════════════════════════
-- PART B — Platform-to-center billing
-- ═══════════════════════════════════════════════════════

-- 4. Rename center.price_per_seat → center.price_per_teacher
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'center' AND column_name = 'price_per_seat'
    ) THEN
        ALTER TABLE center RENAME COLUMN price_per_seat TO price_per_teacher;
    END IF;
END $$;

-- 5. Add center.currency (explicit per-center currency, never inferred at runtime)
ALTER TABLE center
    ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT 'DZD';

-- Backfill: centers in France get EUR
UPDATE center SET currency = 'EUR' WHERE country = 'FR' AND currency = 'DZD';

-- 6. Add center.billing_mode (enum-ready, only value today is 'per_teacher')
ALTER TABLE center
    ADD COLUMN IF NOT EXISTS billing_mode TEXT NOT NULL DEFAULT 'per_teacher'
    CHECK (billing_mode IN ('per_teacher'));

-- 7. Add admin.billable_from (set on first login: NOW() + 30 days; never reset)
ALTER TABLE admin
    ADD COLUMN IF NOT EXISTS billable_from TIMESTAMPTZ;

-- 8. Add admin.deactivated_at (set on deactivate, cleared on reactivate; audit trail)
ALTER TABLE admin
    ADD COLUMN IF NOT EXISTS deactivated_at TIMESTAMPTZ;

-- 9. Platform-to-center invoice table
CREATE TABLE IF NOT EXISTS center_invoice (
    id                SERIAL PRIMARY KEY,
    center_id         INT NOT NULL REFERENCES center(id) ON DELETE CASCADE,
    period_month      DATE NOT NULL,                      -- always YYYY-MM-01 UTC
    teacher_count     INT NOT NULL DEFAULT 0,
    price_per_teacher NUMERIC(10,2) NOT NULL,             -- snapshot at generation time
    currency          TEXT NOT NULL DEFAULT 'DZD',        -- snapshot at generation time
    total_amount      NUMERIC(10,2) NOT NULL,             -- teacher_count × price_per_teacher
    status            TEXT NOT NULL DEFAULT 'unpaid'
                      CHECK (status IN ('unpaid', 'paid', 'cancelled')),
    paid_at           TIMESTAMPTZ,
    paid_method       TEXT NOT NULL DEFAULT '',           -- cash, ccp, virement, other
    paid_reference    TEXT NOT NULL DEFAULT '',
    generated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (center_id, period_month)
);

CREATE INDEX IF NOT EXISTS idx_center_invoice_center  ON center_invoice(center_id);
CREATE INDEX IF NOT EXISTS idx_center_invoice_status  ON center_invoice(status);
CREATE INDEX IF NOT EXISTS idx_center_invoice_period  ON center_invoice(period_month);
```

**Notes on idempotency:**
- `DROP TABLE IF EXISTS` — safe if tables already dropped.
- `DO $$ ... IF EXISTS` block — safe if column already renamed.
- `ADD COLUMN IF NOT EXISTS` — safe on repeat runs.
- `UPDATE center SET currency ...` — harmless if already backfilled (WHERE currency='DZD' prevents double-apply to FR centers that were manually set to EUR).
- `CREATE TABLE IF NOT EXISTS` + `CREATE INDEX IF NOT EXISTS` — safe.

---

## Artifact 2 — Handler table

Every handler affected. Status: **NEW** = create, **EDIT** = modify, **DELETE** = remove function + route.

### Deleted handlers

| Route | Handler function | File | One-line purpose |
|-------|-----------------|------|-----------------|
| `GET /admin/center/billing` | `CenterBilling` | handlers/center.go | Show student invoice list for a center month |
| `POST /admin/center/billing/generate` | `CenterGenerateInvoices` | handlers/center.go | Generate student invoices for a month |
| `POST /admin/center/billing/:invoiceId/paid` | `CenterMarkInvoicePaid` | handlers/center.go | Mark a student invoice as paid |
| `POST /admin/center/billing/:invoiceId/cancel` | `CenterCancelInvoice` | handlers/center.go | Cancel a student invoice |
| `POST /admin/classroom/:id/billing` | `UpdateClassroomBilling` | handlers/admin.go | Update classroom session_rate and billing_enabled |

### New handlers

| Route | Handler function | File | One-line purpose |
|-------|-----------------|------|-----------------|
| `POST /admin/center/teachers/:id/reset-password` | `CenterResetTeacherPassword` | handlers/center.go | Owner generates new password for a teacher; verifies teacher.center_id == owner.center_id; shows password once |
| `POST /platform/centers/:id/generate-invoices` | `PlatformGenerateCenterInvoice` | handlers/platform.go | Platform admin manually triggers invoice generation for `?month=YYYY-MM`; inserts one `center_invoice` row |

### Edited handlers

| Route | Handler function | File | What changes |
|-------|-----------------|------|-------------|
| `POST /admin/login` | `AdminLogin` | handlers/admin.go | (1) Remove center subscription-status gate block. (2) Remove legacy individual subscription-status gate block. (3) After successful login, if `admin.Role == "teacher"` and `admin.BillableFrom == nil`: call `store.SetBillableFrom(adminID, NOW()+30d)` |
| `POST /admin/center/teachers/:id/toggle` | `CenterToggleTeacher` | handlers/center.go | On deactivate: call `store.SetDeactivatedAt(teacherID, &now)` in addition to `DeactivateTeacher`. On reactivate: call `store.SetDeactivatedAt(teacherID, nil)` in addition to `ActivateTeacher` |
| `GET /admin/center` | `CenterDashboard` | handlers/center.go | Swap `GetCenterDashboardStats` (which queries removed tables) for `GetCenterTeachers` + `HasUnpaidCenterInvoice`. Pass `ThisMonthTotal = activeCount × center.PricePerTeacher`, `Currency = center.Currency`, `NextInvoiceDate = 1st of next month`. Remove `Stats.RevenueThisMonth` and `Stats.ParentViewsWeek` data |
| `GET /admin/center/settings` | `CenterSettings` | handlers/center.go | Replace `geo.CurrencyForCountry(center.Country)` with `center.Currency` (read from DB field, not inferred) |
| `GET /p/:code` | `ParentReport` | handlers/parent.go | Remove `h.Store.LogParentView(...)` call. Remove `h.Store.GetStudentInvoices(...)` call and `Invoices` template key |
| `GET /platform/centers/:id` | `PlatformCenterDetail` | handlers/platform.go | Display `price_per_teacher` + `currency` instead of `price_per_seat`. Pass `center.Currency` to template |
| `POST /platform/centers/:id/seats` | `PlatformCenterUpdateSeats` | handlers/platform.go | Rename route to `/pricing`; update `price_per_teacher` and `currency` (not `price_per_seat` + `seat_count`) |

---

## Artifact 3 — Store method table

Status: **NEW** = add, **EDIT** = modify body, **DELETE** = remove entirely.

### Deleted methods

| Method | File | Reason |
|--------|------|--------|
| `GenerateMonthlyInvoices(ctx, centerID int, period time.Time) (int, error)` | store/invoicing.go | student_invoice removed |
| `ListCenterInvoices(ctx, centerID int, period time.Time, status string) ([]StudentInvoice, error)` | store/invoicing.go | student_invoice removed |
| `MarkInvoicePaid(ctx, invoiceID, centerID int, method string) error` | store/invoicing.go | student_invoice removed |
| `CancelInvoice(ctx, invoiceID, centerID int) error` | store/invoicing.go | student_invoice removed |
| `GetStudentInvoices(ctx, studentID, classroomID int) ([]StudentInvoice, error)` | store/invoicing.go | student_invoice removed |
| `GetCenterBillingSummary(ctx, centerID int, period time.Time) (float64, float64, int, error)` | store/invoicing.go | student_invoice removed |
| `LogParentView(ctx, parentCode, ip string)` | store/invoicing.go | parent_view_log removed |
| `GetParentViewsWeek(ctx, centerID int) int` | store/invoicing.go | parent_view_log removed |
| `UpdateClassroomBilling(ctx, classroomID, adminID int, sessionRate float64, billingEnabled bool) error` | store/store.go | classroom.session_rate + billing_enabled removed |
| `UpdateClassroomBillingAny(ctx, classroomID int, sessionRate float64, billingEnabled bool) error` | store/store.go | classroom.session_rate + billing_enabled removed |
| `UpdateCenterSeats(ctx, centerID, seats int, price float64) error` | store/center.go | Replaced by UpdateCenterPricing |

> After removing all 8 methods from `store/invoicing.go`, the file itself is deleted.

### New methods

| Method | File | Signature | Purpose |
|--------|------|-----------|---------|
| `SetBillableFrom` | store/center.go | `(ctx, adminID int, t time.Time) error` | Writes `billable_from` once on first teacher login |
| `SetDeactivatedAt` | store/center.go | `(ctx, adminID int, t *time.Time) error` | Sets or clears `deactivated_at`; nil = clear (reactivation) |
| `UpdateCenterPricing` | store/center.go | `(ctx, centerID int, pricePerTeacher float64, currency string) error` | Replaces UpdateCenterSeats; updates `price_per_teacher` + `currency` |
| `GenerateCenterMonthlyInvoice` | store/invoicing.go (new file) | `(ctx, centerID int, periodMonth time.Time) (*CenterInvoice, error)` | Counts billable teachers, inserts `center_invoice` row, returns the row (existing if conflict) |
| `HasUnpaidCenterInvoice` | store/invoicing.go (new file) | `(ctx, centerID int) bool` | Returns true if any `center_invoice.status='unpaid'` for this center — used for dunning banner |
| `ListCenterInvoices` | store/invoicing.go (new file) | `(ctx, centerID int) ([]CenterInvoice, error)` | Lists all invoices for a center for the platform admin detail view |
| `MarkCenterInvoicePaid` | store/invoicing.go (new file) | `(ctx, invoiceID, centerID int, method, reference string) error` | Platform admin marks a center_invoice as paid |
| `CancelCenterInvoice` | store/invoicing.go (new file) | `(ctx, invoiceID, centerID int) error` | Platform admin cancels a center_invoice |

> `store/invoicing.go` is fully deleted and recreated with the 5 new methods above.

### Edited methods

| Method | File | What changes |
|--------|------|-------------|
| `GetCenter` | store/center.go | Add scan for `currency`, `billing_mode`, `price_per_teacher` (renamed column) |
| `GetCenterStats` | store/center.go | Remove `SeatCount` from query/struct (or keep as-is if seat_count column is retained; see open Q1) |
| `ListCenterTeachers` | store/center.go | Add `billable_from`, `deactivated_at` to scan so dashboard can show billing status |
| `DeactivateTeacher` | store/center.go | After setting `active=false`, also set `deactivated_at = NOW()` in the same UPDATE |
| `ActivateTeacher` | store/center.go | After setting `active=true`, also set `deactivated_at = NULL` in the same UPDATE |
| `CreateCenter` | store/center.go | Accept `currency string, pricePerTeacher float64` params; set them in INSERT |
| `ListCenters` | store/center.go | Replace `seat_count` column reference with `price_per_teacher`; add `currency` |
| `GetCenterDashboardStats` | store/center_analytics.go | Remove the `student_invoice` revenue query and the `parent_view_log` query. Remove `RevenueThisMonth` and `ParentViewsWeek` fields from `CenterDashboardStats` struct |
| `GetClassroomByCode` | store/store.go | Remove `session_rate`, `billing_enabled` from SELECT and Scan |

### Edited models (store/store.go)

| Model | Change |
|-------|--------|
| `Admin` struct | Add `BillableFrom *time.Time`, `DeactivatedAt *time.Time` |
| `Classroom` struct | Remove `SessionRate float64`, `BillingEnabled bool` |
| `StudentInvoice` struct | **Delete** entire struct |
| `Center` struct | In `store/center.go`: rename `PricePerSeat float64` → `PricePerTeacher float64`; add `Currency string`, `BillingMode string` |
| `CenterTeacher` struct | Add `BillableFrom *time.Time`, `DeactivatedAt *time.Time` |
| `CenterDashboardStats` struct | Remove `RevenueThisMonth float64`, `ParentViewsWeek int` |
| `CenterInvoice` struct | **New** — add to `store/store.go`: fields mirror `center_invoice` table |

`CenterInvoice` struct shape:
```
type CenterInvoice struct {
    ID               int
    CenterID         int
    PeriodMonth      time.Time
    TeacherCount     int
    PricePerTeacher  float64
    Currency         string
    TotalAmount      float64
    Status           string
    PaidAt           *time.Time
    PaidMethod       string
    PaidReference    string
    GeneratedAt      time.Time
}
```

---

## Artifact 4 — Middleware table

| Function | File | Status | What changes |
|----------|------|--------|-------------|
| `AdminSubscriptionCheck` | middleware/middleware.go | EDIT | Remove the center subscription-status block (expired/suspended redirect). Keep the `!admin.Active` deactivated check. Remove the legacy individual `admin.SubscriptionStatus` check. The function becomes: verify admin exists + is active → inject `admin` into context → proceed |
| `CenterOwnerRequired` | middleware/middleware.go | EDIT | Remove the `GetCenter` + subscription-status check block (the `if center.SubscriptionStatus != "active" && ...` block). Keep: session cookie check, admin lookup, `active` check, `role == "owner"` check |

---

## Artifact 5 — Template table

| Path | Status | What changes |
|------|--------|-------------|
| `templates/admin/center_billing.html` | **DELETE** | Entire file removed (student invoicing UI) |
| `templates/admin/admin_classroom.html` | **EDIT** | Remove the "Billing Settings" form block (lines ~40–52: `<form action="/admin/classroom/:id/billing">`, `billing_enabled` checkbox, `session_rate` input, CSRF, submit). Keep everything else |
| `templates/admin/center_dashboard.html` | **FULL REWRITE** | See design spec below |
| `templates/admin/center_teachers.html` | **REWRITE** | See design spec below |
| `templates/admin/center_settings.html` | **EDIT** | Replace `{{.Currency}}` + `{{.PricePerSeat}}` references with `{{.Center.Currency}}` + `{{.Center.PricePerTeacher}}` |
| Center nav layout (within center templates) | **EDIT** | Remove "Facturation" nav link that points to `/admin/center/billing` |
| `templates/admin/admin_login.html` | **EDIT** | Remove error message cases for `expired`, `trial_expired`, `suspended` (or keep them as dead code — see open Q2) |

### center_dashboard.html redesign spec

**Remove:**
- 4-stat grid (Teachers/Students/Sessions/Revenue cards)
- Seat/subscription status card (`{{.Stats.ActiveSeats}} / {{.Stats.SeatCount}} seats`, `price/seat/month`)
- `RevenueThisMonth` and `ParentViewsWeek` stats

**Add:**
1. **Dunning banner** (conditional, renders when `{{.HasUnpaidInvoice}}` is true):  
   Amber/red strip at top: "Facture impayée — Veuillez contacter l'équipe TeachHub pour régulariser votre situation."  
   No lockout, no redirect, just visual.

2. **"This month" billing card** (always visible):  
   - `N enseignants actifs × price currency/mois = total currency`  
   - "Prochaine facture : 1er [next month name]"  
   - Read from template data: `{{.ActiveTeacherCount}}`, `{{.Center.PricePerTeacher}}`, `{{.Center.Currency}}`, `{{.NextInvoiceDate}}`

3. **Teacher cards grid** (replaces old table/list):  
   One card per teacher from `{{.Teachers}}` (type `[]CenterTeacher`). Each card shows:
   - **Avatar** from initials (first letter of DisplayName or Username, CSS circle)
   - DisplayName / username
   - Email, Phone
   - Subjects + levels (from admin.Subjects, admin.Levels — already in Admin struct)
   - `# salles · # élèves` — from `CenterTeacher.ClassroomCount`, `CenterTeacher.StudentCount`
   - Last login (`LastLoginAt` formatted, or "Jamais connecté" if nil)
   - Trial badge if `BillableFrom != nil && BillableFrom > now` ("Période d'essai — facturable le DD/MM/YYYY")
   - **Quick actions** inline: "Désactiver" / "Réactiver" button (POST toggle), "Réinitialiser MDP" button (POST reset-password)

4. **Stats row** (lightweight, replaces old 4-card grid):  
   3 counters: `N enseignants · N élèves · N salles`  
   Read from existing `CenterStats` fields (TeacherCount, StudentCount, ClassCount).

### center_teachers.html redesign spec

**Remove:**
- Table-based layout
- "Billing total" line (MonthlyTotal — removed in 39c14a4, already gone per context)

**Add:**
- Card-based layout matching center_dashboard.html teacher cards (same fields)
- **Reset password button** on each card: small button "Réinitialiser MDP" → POST `/admin/center/teachers/{{.ID}}/reset-password` → redirects back with `?pw=xxx&user=yyy` flash, shown once inline
- Add form stays (create teacher), but remove any seat-limit warning text

---

## Artifact 6 — Files touched

### CREATE (new files)
| File | Contents |
|------|----------|
| `db/migrations/013_billing_rework.sql` | The migration SQL above |
| `docs/billing.md` | Billing policy document |
| `store/invoicing.go` | Rebuilt: `CenterInvoice` model + 5 new methods (`GenerateCenterMonthlyInvoice`, `HasUnpaidCenterInvoice`, `ListCenterInvoices`, `MarkCenterInvoicePaid`, `CancelCenterInvoice`) |

### DELETE (files fully removed)
| File | Reason |
|------|--------|
| `store/invoicing.go` (old) | All 8 methods reference dropped tables; file deleted and recreated (see CREATE above) |
| `templates/admin/center_billing.html` | Student invoicing UI, deleted |

### EDIT (existing files modified)
| File | Summary of changes |
|------|--------------------|
| `store/store.go` | Add `BillableFrom`, `DeactivatedAt` to `Admin`. Add `CenterInvoice` struct. Remove `StudentInvoice` struct. Remove `SessionRate`, `BillingEnabled` from `Classroom`. Remove `UpdateClassroomBilling`, `UpdateClassroomBillingAny` functions |
| `store/center.go` | Rename `PricePerSeat→PricePerTeacher` in `Center` struct + SQL. Add `Currency`, `BillingMode`. Add `BillableFrom`, `DeactivatedAt` to `CenterTeacher`. Update `GetCenter`, `DeactivateTeacher`, `ActivateTeacher`, `ListCenterTeachers`, `ListCenters`, `CreateCenter`. Add `SetBillableFrom`, `SetDeactivatedAt`, `UpdateCenterPricing`. Delete `UpdateCenterSeats` |
| `store/center_analytics.go` | Remove `student_invoice` and `parent_view_log` queries from `GetCenterDashboardStats`. Remove `RevenueThisMonth`, `ParentViewsWeek` from `CenterDashboardStats` struct |
| `middleware/middleware.go` | Edit `AdminSubscriptionCheck`: remove subscription gate. Edit `CenterOwnerRequired`: remove subscription gate |
| `handlers/admin.go` | Edit `AdminLogin`: remove both subscription check blocks; add `SetBillableFrom` call on first teacher login. Delete `UpdateClassroomBilling` function |
| `handlers/center.go` | Delete `CenterBilling`, `CenterGenerateInvoices`, `CenterMarkInvoicePaid`, `CenterCancelInvoice`. Edit `CenterDashboard`, `CenterToggleTeacher`, `CenterSettings`. Add `CenterResetTeacherPassword` |
| `handlers/parent.go` | Remove `LogParentView` call. Remove `GetStudentInvoices` call and `Invoices` key |
| `handlers/platform.go` | Edit `PlatformCenterDetail`: use `price_per_teacher`/`currency`. Edit/rename `PlatformCenterUpdateSeats` → `PlatformCenterUpdatePricing`. Add `PlatformGenerateCenterInvoice` |
| `main.go` | Remove routes: `billing`, `billing/generate`, `billing/:invoiceId/paid`, `billing/:invoiceId/cancel`, `/classroom/:id/billing`. Add routes: `/teachers/:id/reset-password`, `/platform/centers/:id/generate-invoices`. Rename `/platform/centers/:id/seats` → `/pricing` |
| `templates/admin/admin_classroom.html` | Remove billing form block (~lines 40–52) |
| `templates/admin/center_dashboard.html` | Full rewrite per design spec |
| `templates/admin/center_teachers.html` | Rewrite to card layout + add reset-password button |
| `templates/admin/center_settings.html` | Replace `PricePerSeat` / `geo.CurrencyForCountry` refs with `Center.PricePerTeacher` / `Center.Currency` |
| `templates/admin/admin_login.html` | Remove or neutralise error messages for `expired`, `suspended`, `trial_expired` |
| Platform center templates (`platform_centers.html`, `platform_center_detail.html`) | Replace `seat_count`/`price_per_seat` display with `price_per_teacher`/`currency`; add invoice list section to detail page |

**Not touched (zero changes):**
- `geo/geo.go`, `geo/`
- `handlers/quiz.go`, `handlers/student.go`, `handlers/analytics.go`, `handlers/live.go`, `handlers/export.go`, `handlers/explore.go`
- `store/center_analytics.go::GetCenterTeacherPerformance` (uses only live_session, no billing)
- All quiz/assignment/submission/resource/live/remark tables and their store methods
- `payment` table
- `teacher_application` table, `/apply` routes, `PlatformUpdateAppStatus` (except the one-line center creation default)
- `classroom_student.parent_code` column — kept as-is
- `/p/:code` parent report — kept, academic data only

---

## Artifact 7 — Open questions

**Q1 — Drop `seat_count` from `center` table?**  
The spec renames `price_per_seat → price_per_teacher` but does not mention `seat_count`. The column is currently used in `GetCenter`, `UpdateCenterSeats`, `ListCenters`, and `center_dashboard.html` ("X / Y seats"). Since we're removing seat caps entirely, do you want to:
- (a) Keep `seat_count` as dead legacy metadata (no UI, no enforcement)
- (b) DROP it in migration 013

Recommendation: (a) keep — DROP requires touching more code and the data has no harm. If you choose (b), add `ALTER TABLE center DROP COLUMN IF EXISTS seat_count;` to migration 013 and remove it from `GetCenter`, `ListCenters`, `CenterStats`.

**Q2 — `admin_login.html` error messages for `expired`/`suspended`/`trial_expired`**  
With the subscription gate removed from login, these query-param errors (`?error=expired` etc.) will never be set. Do you want to:
- (a) Leave the HTML as dead code (harmless)
- (b) Remove the `{{if eq .Error "expired"}}` blocks to clean up the template

Recommendation: (b) remove — cleaner, but low priority. Mark for Phase 3.

**Q3 — Platform center detail page: show invoice history?**  
`PlatformGenerateCenterInvoice` creates invoices but the platform admin needs to see them somewhere. Should the existing `platform_center_detail.html` be extended with an invoice table (list from `ListCenterInvoices`)? Or is that a Phase 3 task?

Recommendation: include it now since `PlatformGenerateCenterInvoice` is being built in the same phase — otherwise there's no way to verify what was generated without a DB query. Adds ~1 table to `platform_center_detail.html` only.

**Q4 — `CenterResetTeacherPassword`: show password to owner once — how?**  
Two approaches:
- (a) Redirect to `/admin/center/teachers?reset=username&pw=PLAINTEXT` — plain text in URL (visible in browser history, server logs)
- (b) Store `pending_password` on the admin row (already exists), redirect to `/admin/center/teachers?reset=username`, then the teacher list page reads and clears `pending_password` from the DB when rendering

Recommendation: (b) — consistent with how initial passwords are shown (`pending_password` pattern already implemented in `CenterCreateTeacher`). Safer: no password in URL.

**Q5 — `PlatformCenterUpdateSeats` route rename: breaking change?**  
Renaming `/centers/:id/seats` → `/centers/:id/pricing` changes the form `action` in `platform_center_detail.html`. This is fine since it's an internal admin form, not a public API. Confirm before code.

**Q6 — `db/schema.sql` sync**  
Existing convention: `db/schema.sql` is kept in sync with migrations for fresh deployments. Should migration 013's changes be reflected in `schema.sql` in the same PR, or is schema.sql updated separately?  
Recommendation: update `schema.sql` in the same commit as 013 to avoid drift.
