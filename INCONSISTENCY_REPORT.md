# TeachHub — Comprehensive Inconsistency & UX Audit Report

---

## 1. LOGIN PAGE — DOESN'T DIFFERENTIATE OWNER VS TEACHER

**Severity:** High (User-reported issue)

The login page says "Teacher Portal" / "Espace Enseignant" for **everyone** — both center owners and teachers.

| File | Line | Current Text |
|------|------|-------------|
| `i18n/i18n.go` | 20 | `"login_heading": {"en": "Teacher Portal", "fr": "Espace Enseignant"}` |
| `i18n/i18n.go` | 21 | `"login_subheading": {"en": "Sign in to manage your classrooms", ...}` |
| `templates/admin/admin_login.html` | 36 | `<h1>{{t .Lang "login_heading"}}</h1>` — renders "Teacher Portal" |

**Problem:** Owners are **not** teachers — they manage a center, teachers, and billing. The login page doesn't convey that owners also log in here. Since both roles share one login form, the heading should be neutral (e.g., "Sign In" / "Connexion") or dynamic.

**Related:** After login, owners are redirected to `/admin/center` (handlers/admin.go ~line 186), while teachers go to `/admin`. The login page gives no hint of this.

---

## 2. CENTER PAGES — ALL HARDCODED ENGLISH (NO i18n)

**Severity:** High

Center management pages are entirely in hardcoded English while the rest of the app uses i18n. A French-speaking owner sees French everywhere *except* the center pages.

### 2.1 `templates/admin/center_dashboard.html` (entire file)
All text is hardcoded English:
- Line 12: `"Center Dashboard"`
- Line 24: `"Teachers"` stat label
- Line 32: `"Students"` stat label  
- Line 40: `"Classes"` stat label
- Line 48: `"Sessions"` stat label
- Line 59: `"Revenue (month)"` stat label
- Line 68: `"Parent Views (week)"`
- Line 81: `"Subscription"`, `"Active"`, `"Trial"` 
- Line 88: `"seats included"`
- Line 109–113: Table headers `"Username"`, `"Classes"`, `"Students"`, `"Status"`
- Line 145: `"Teacher Performance"` section heading
- Line 147–152: Table headers `"Teacher"`, `"Classes"`, `"Students"`, `"Avg Quiz %"`, `"Sessions (month)"`, `"Last Active"`
- Line 167: `"Never"` for last active
- Line 175–195: Quick links `"My Classes"`, `"Center Settings"`, `"Center Billing"`

### 2.2 `templates/admin/center_teachers.html` (entire file)
- Line 10: `"Teachers"` heading
- Line 11: `"seats used"` subheading
- Lines 17–22: Error messages all hardcoded
- Line 28: `"Teacher ... created successfully"` 
- Line 32: `"Add Teacher"` heading
- Lines 37–50: Form labels `"Full Name *"`, `"Email *"`, `"Phone"`, `"Create"` button
- Lines 78–85: Table headers `"Name"`, `"Login"`, `"Email"`, `"Role"`, `"Classes"`, `"Students"`, `"Status"`, `"Action"`
- Line 97: `"Owner"`, `"Teacher"` role badges
- Line 103–107: `"Active"`, `"Inactive"` status badges
- Line 115: `"Deactivate"` / `"Activate"` buttons
- Line 135: `"No teachers yet."`

### 2.3 `templates/admin/center_settings.html` (entire file)
All labels: `"Center Settings"`, `"Center Name *"`, `"Address"`, `"City / Wilaya"`, etc.

### 2.4 Exception: `templates/admin/center_billing.html`
This page **does** use i18n properly with `{{t .Lang "billing"}}` etc. — inconsistent with the other 3 center pages.

### 2.5 `templates/layouts/admin.html` — Nav bar
The owner-only nav links are hardcoded English:
```html
<!-- Hardcoded: -->
Center
Teachers
```
Should use `{{t .Lang "nav_center"}}` and `{{t .Lang "nav_teachers"}}`.

---

## 3. PLATFORM PAGES — "Centers" NAV LINK HARDCODED

**Severity:** Medium

Across **all** platform templates, the "Centers" nav link is hardcoded English while the other nav items use i18n:

| File | Hardcoded text |
|------|---------------|
| `templates/platform/platform_dashboard.html` line 29 | `>Centers</a>` |
| `templates/platform/platform_applications.html` line 25 | `>Centers</a>` |
| `templates/platform/platform_credentials.html` line 24 | `>Centers</a>` |
| `templates/platform/platform_teacher_detail.html` line 24 | `>Centers</a>` |
| `templates/platform/platform_analytics.html` line 26 | `>Centers</a>` |
| `templates/platform/platform_centers.html` lines 29, 40, 41, 45 | Multiple: `"Centers"`, `"Total Centers"`, `"Manage all learning centers..."` |
| `templates/platform/platform_center_detail.html` lines 35, 47 | `"Centers"`, `"← Back to Centers"` |

All other platform nav items use `{{t .Lang "plat_nav_dashboard"}}`, `{{t .Lang "plat_nav_applications"}}`, etc.

---

## 4. TEACHER CREATION — SUCCESS MESSAGE PROMISES PASSWORD BUT DOESN'T SHOW IT

**Severity:** Critical (Broken UX flow)

**File:** `templates/admin/center_teachers.html` lines 26–28

```html
<span>Teacher <strong>{{.Query.created}}</strong> created successfully.
They can log in with their temporary password shown below.</span>
```

**Problem:** The message says "temporary password shown below" but **no password is displayed anywhere on the page**. The password is generated in `handlers/center.go` (`CenterCreateTeacher`) and stored in `pending_password`, but the redirect only passes `?created=<username>` — the password is never sent to the UI.

**Backend flow** (`handlers/center.go` ~line 135–170):
1. Generates random 8-char password
2. Stores hashed version + `pending_password` (plain) in DB
3. Redirects to `/admin/center/teachers?created=<username>`

The owner has **no way** to see the generated password for the teacher they just created.

---

## 5. ERROR MESSAGE TEXT MISMATCH — FORM FIELD NAMES

**Severity:** Medium

**File:** `templates/admin/center_teachers.html` line 20

```html
{{else if eq .Query.error "missing_fields"}}Username and email are required.
```

**Problem:** The form fields are `display_name` and `email` (lines 41–42), not `username` and `email`. The username is auto-generated from the email prefix in `handlers/center.go` line ~150. The error message references a field the user doesn't see.

**Should say:** "Name and email are required." or use the i18n system.

---

## 6. CENTER DASHBOARD TEACHER TABLE — SHOWS USERNAME INSTEAD OF DISPLAY NAME

**Severity:** Medium

**File:** `templates/admin/center_dashboard.html` line 109

```html
<th>Username</th>
...
<td>{{.Username}}</td>
```

**vs** `templates/admin/center_teachers.html` line 89:

```html
<span>{{if .DisplayName}}{{.DisplayName}}{{else}}{{.Username}}{{end}}</span>
```

The center_dashboard shows raw `Username` (auto-generated, e.g. "ahmed"), while center_teachers shows `DisplayName` ("Ahmed Benali") with Username as a secondary "Login" column. The dashboard table is less useful and inconsistent.

**Also:** The "Teacher Performance" table at the bottom of center_dashboard.html (line ~155) also uses `{{.Username}}` instead of DisplayName.

---

## 7. EXPLORE FEATURE — REFERENCED BUT FULLY DISABLED

**Severity:** Medium (Dead UI / Confusing)

The Explore feature (public teacher directory) is **commented out** in routes but still referenced in the admin profile page.

**Disabled routes** in `main.go` lines ~290–294:
```go
// Explore routes are commented out
```

**But the profile page still shows:**

`templates/admin/admin_profile.html`:
- Public Profile toggle (checkbox to make profile visible)
- Bio, Subjects, Levels input fields
- Preview link: `<a href="/explore/teacher/{{.Admin.ID}}">`

**Problem:** Teachers can toggle their profile "public" and fill in bio/subjects/levels, but the explore directory doesn't exist. The preview link returns 404. The profile page exposes a feature that does nothing.

---

## 8. HARDCODED PLATFORM ADMIN LINK IN STUDENT HOME

**Severity:** Medium (Potential 404)

**File:** `templates/student/student_home.html` line 88

```html
<a href="/platform/login" ...>
```

**Problem:** The platform admin path is configurable via `PLATFORM_PATH` env var (default: `/ctrl-p-8x3kf`), set in `main.go` line ~80. The student home page hardcodes `/platform/login` instead of using the dynamic path. If the platform path is changed from default, this link breaks (404).

Other platform templates correctly use `{{.PlatformPath}}` (e.g., `platform_dashboard.html` line 23).

---

## 9. PHONE FIELD — VALIDATION MISMATCH

**Severity:** Low

**File:** `templates/student/student_join.html` line 47

```html
<input type="tel" ... pattern="[0-9]{8,15}" maxlength="15" ...>
```

**vs** `i18n/i18n.go` line 498:

```
"phone_help": {"en": "10-digit phone number so the teacher can contact you.", ...}
```

**Problem:** The HTML pattern accepts 8–15 digits, but the help text says "10-digit phone number." In Algeria, phone numbers are 10 digits (0X XX XX XX XX), so the help text is correct for the local context but the validation is too loose. Or conversely, 8 digits could be valid for some countries — but then the help text is wrong.

---

## 10. APPLICATION FORM — INLINE i18n INSTEAD OF TRANSLATION KEYS

**Severity:** Low-Medium

**File:** `templates/platform/apply.html` lines 82–106

Several labels use inline `{{if eq .Lang "fr"}}...{{else}}...{{end}}` instead of proper i18n keys:

```html
<!-- Line 82 -->
<label>{{if eq .Lang "fr"}}Nom du centre{{else}}Center Name{{end}}</label>
<!-- Line 88 -->
<label>{{if eq .Lang "fr"}}Nombre d'enseignants{{else}}Number of Teachers{{end}}</label>
<!-- Line 103 -->
<label>{{if eq .Lang "fr"}}Nombre d'élèves estimé{{else}}Expected Students{{end}}</label>
```

**Problem:** Inconsistent with the rest of the app which uses `{{t .Lang "key"}}`. Makes translation maintenance harder — changes need to be made in templates instead of the centralized i18n file.

---

## 11. PLATFORM CREDENTIALS PAGE — INLINE i18n

**Severity:** Low

**File:** `templates/platform/platform_credentials.html` lines 88–90

```html
<p>{{if eq .Lang "fr"}}Mot de passe non disponible{{else}}Password no longer available{{end}}</p>
<p>{{if eq .Lang "fr"}}L'enseignant s'est déjà connecté...{{else}}The teacher has already logged in...{{end}}</p>
```

Same issue as #10 — inline translations instead of i18n keys.

---

## 12. "SchoolName" FIELD — LEGACY CONCEPT IN CENTER MODEL

**Severity:** Low (Data model confusion)

The `admin` table still has `school_name` (schema.sql line 37), and the application form collects both `school_name` AND `center_name` (apply.html lines 77, 82).

**In `handlers/platform.go` line 230:**
```go
centerName = app.SchoolName  // Falls back to school_name if center_name is empty
```

The `SchoolName` is a pre-center legacy field. Now that centers exist, `school_name` on the admin table is redundant — the center has its own `name`. New applications collect `center_name` separately, but the code still stores/references `school_name` on the admin model.

---

## 13. OWNER POST-LOGIN REDIRECT — POTENTIALLY CONFUSING

**Severity:** Low

**File:** `handlers/admin.go` lines 185–188

```go
if admin.Role == "owner" && admin.CenterID != nil {
    c.Redirect(http.StatusFound, "/admin/center")
    return
}
```

Owners always land on `/admin/center` (center dashboard). To manage their own classrooms, they must click "My Classes" or navigate to `/admin`. This isn't broken, but:
- The nav bar shows both "Center" and "My Classes" links
- There's no visual cue that the owner's personal classrooms are at a different location
- A new owner might not realize they need to navigate away from `/admin/center` to create/manage classrooms

---

## 14. "admin_login.html" — TRAILING PIPE IN FOOTER LINKS

**Severity:** Low (Visual)

**File:** `templates/admin/admin_login.html` line ~89

```html
<a href="/apply" ...>{{t .Lang "nav_become_teacher"}}</a>
<span class="text-gray-200">|</span>

</div>
```

There's a trailing `|` separator with nothing after it (the explore link was presumably removed but the separator remained). This renders a dangling pipe character on the login page.

---

## 15. COOKIE CONSENT — ON LOGIN PAGE ONLY

**Severity:** Low

**File:** `templates/admin/admin_login.html` lines 91–108

The cookie consent banner is embedded directly in `admin_login.html` but not in other entry points (`student_home.html`, `landing.html`). If a student visits the site, they never see the cookie consent banner.

---

## 16. PLATFORM CENTERS PAGE — ALL HARDCODED ENGLISH

**Severity:** Medium

**File:** `templates/platform/platform_centers.html`

- Line 7: `<title>TeachHub Platform — Centers</title>`
- Line 40: `<h1>Centers</h1>`
- Line 41: `<p>Manage all learning centers, seats, and subscriptions</p>`
- Line 45: `<div>Total Centers</div>`
- Line 89: `>Manage</a>` button text

While other platform pages mostly use i18n, the centers page and center detail page are fully hardcoded English.

---

## 17. PLATFORM CENTER DETAIL — HARDCODED ENGLISH

**Severity:** Medium

**File:** `templates/platform/platform_center_detail.html`

- Line 47: `← Back to Centers`
- Various labels for seat management, subscription status, etc.

---

## 18. CENTER DASHBOARD — "Teacher Performance" TABLE SHOWS USERNAME, NOT DISPLAY NAME

**Severity:** Medium (duplicate of point in #6, separate table)

**File:** `templates/admin/center_dashboard.html` line ~155

```html
<td><div style="font-weight:600;color:#1a1a2e;">{{.Username}}</div>
```

The Teacher Performance table shows `Username` (auto-generated login) instead of `DisplayName` (the actual name the owner entered when creating the teacher). This is the same table that also shows email, making the Username column redundant and confusing.

---

## SUMMARY TABLE

| # | Issue | Severity | Files Affected |
|---|-------|----------|---------------|
| 1 | Login page says "Teacher Portal" for owners too | High | i18n/i18n.go:20, admin_login.html |
| 2 | Center pages all hardcoded English (no i18n) | High | center_dashboard.html, center_teachers.html, center_settings.html |
| 3 | Platform "Centers" nav link hardcoded in all templates | Medium | 7+ platform template files |
| 4 | Teacher creation success says "password shown below" but no password displayed | Critical | center_teachers.html:26 |
| 5 | Error message says "Username and email required" but form has "Full Name" not "Username" | Medium | center_teachers.html:20 |
| 6 | Center dashboard teacher table shows Username, not DisplayName | Medium | center_dashboard.html:109, 155 |
| 7 | Explore feature disabled but still shown in admin profile | Medium | main.go (routes), admin_profile.html |
| 8 | Student home hardcodes /platform/login instead of using PlatformPath | Medium | student_home.html:88 |
| 9 | Phone field: pattern accepts 8-15 digits, help text says 10 digits | Low | student_join.html:47, i18n/i18n.go:498 |
| 10 | Application form uses inline i18n instead of translation keys | Low-Med | apply.html:82-106 |
| 11 | Platform credentials page uses inline i18n | Low | platform_credentials.html:88 |
| 12 | Legacy "school_name" field coexists with center.name | Low | schema.sql, handlers/platform.go |
| 13 | Owner always redirected to /admin/center, not to their classrooms | Low | handlers/admin.go:185 |
| 14 | Trailing pipe separator on login page footer | Low | admin_login.html:89 |
| 15 | Cookie consent banner only on admin login page | Low | admin_login.html |
| 16 | Platform centers page fully hardcoded English | Medium | platform_centers.html |
| 17 | Platform center detail hardcoded English | Medium | platform_center_detail.html |
| 18 | Performance table uses Username instead of DisplayName | Medium | center_dashboard.html:155 |
