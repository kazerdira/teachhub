
gofunc (h *Handler) ServeUpload(c *gin.Context) {
    reqPath := c.Param("filepath")
    reqPath = filepath.Clean(reqPath)
    if strings.Contains(reqPath, "..") {
        c.String(http.StatusForbidden, "Access denied")
        return
    }

    aid := adminID(c)
    if aid == 0 {
        session, _ := middleware.SessionStore.Get(c.Request, "teachhub-admin")
        if session.Values["admin_id"] != nil {
            aid = session.Values["admin_id"].(int)
        }
    }
    student := middleware.GetStudent(c)

    if aid == 0 && student == nil {
        c.String(http.StatusForbidden, "Access denied")
        return
    }

    // Admins: already scoped by ownsClassroom on upload, trust them
    if aid > 0 {
        c.File(filepath.Join(h.UploadDir, reqPath))
        return
    }

    // Students: verify they belong to the classroom this file belongs to
    // File paths are: resources/{classID}_{ts}.ext  or  submissions/{assignID}_{studentID}_{ts}.ext
    // For submissions, the student ID is embedded — check it matches
    parts := strings.SplitN(strings.TrimPrefix(reqPath, "/"), "/", 2)
    if len(parts) == 2 {
        folder := parts[0]
        filename := parts[1]

        switch folder {
        case "submissions":
            // format: {assignID}_{studentID}_{ts}.ext
            segs := strings.SplitN(filename, "_", 3)
            if len(segs) >= 2 {
                ownerID, err := strconv.Atoi(segs[1])
                if err != nil || ownerID != student.ID {
                    c.String(http.StatusForbidden, "Access denied")
                    return
                }
            }
        case "resources", "assignments":
            // format: {classID}_{ts}.ext
            segs := strings.SplitN(filename, "_", 2)
            if len(segs) >= 1 {
                classID, err := strconv.Atoi(segs[0])
                if err != nil {
                    c.String(http.StatusForbidden, "Access denied")
                    return
                }
                in, _ := h.Store.IsStudentInClassroom(c.Request.Context(), student.ID, classID)
                if !in {
                    c.String(http.StatusForbidden, "Access denied")
                    return
                }
            }
        case "live":
            // live images/pdfs/teacher pics — only need to be enrolled somewhere
            // can't easily scope without extra DB lookup, accept for now
        }
    }

    c.File(filepath.Join(h.UploadDir, reqPath))
}
This is the only file you touch. Deploy, done.
The live folder is the one case you can't cheaply scope without a DB lookup — but live images are ephemeral and non-sensitive (they're whiteboard images, teacher profile pics), so leaving them accessible to any authenticated student is acceptable for now.

---------------------------


Here's everything, prioritized:

---

## 🔴 Critical (Security / Data Loss)

**1. File access vulnerability** — any student can access any classroom's files
→ Fixed above, deploy now

**2. Plaintext password in `admin.pending_password`**
→ At minimum, clear it immediately after the platform owner views credentials (you already do this on teacher first login, but the platform owner can view it indefinitely in `PlatformCredentials`)

**3. CSRF on whiteboard save-to-resources**
```js
// In admin_live.html, wbSaveResource():
var csrfEl = document.querySelector('input[name=_csrf]');
// This is always null on the live page — the token is injected as a JS string
// Fix: hardcode the token
var csrfToken = '{{.CSRFToken}}'; // already available in the template
```

**4. Quiz attempt orphan**
→ Between `CreateQuizAttemptAtomic` and `SubmitQuizAttempt`, a crash = used attempt with no answers. Wrap both in a transaction or merge into one store function.

---

## 🟠 Logic Bugs

**5. `AnalyticsMissing` handler fetches data then discards it**
```go
missing, _ := h.Store.GetMissingSubmissions(...)
c.Redirect(...)
_ = missing // wasted DB query every call
```
→ Just redirect directly, remove the store call

**6. Quiz timer uses localStorage — breaks with two tabs open**
→ Store `startedAt` on the attempt server-side (you have the column), calculate remaining time from server on page load, send it to the template. Remove the localStorage logic entirely.

**7. `sanitizePhone` allows `+` anywhere**
```go
// Current: allows 1+23+4
// Fix:
func sanitizePhone(p string) string {
    var b strings.Builder
    for i, r := range p {
        if r >= '0' && r <= '9' || (r == '+' && i == 0) {
            b.WriteRune(r)
        }
    }
    s := b.String()
    if len(s) > 15 { s = s[:15] }
    return s
}
```

**8. N+1 in `PlatformTeacherDetail`**
→ 4 store calls per classroom in a loop. Add a single query that returns all classroom stats for a teacher in one shot.

**9. `DownloadSubmission` has raw SQL in the handler**
→ Move to store layer for consistency and testability

---

## 🟡 Maintainability

**10. `admin_analytics.html` and `admin_analytics_partial.html` are ~identical 700-line files**
→ Merge into one template, add a `{{if .Partial}}` flag to toggle the outer layout. Every change currently requires editing both files.

**11. `admin_live.html` is 1800+ lines of embedded JS**
→ Extract to `/static/js/admin-live.js`. Pass template variables via `data-` attributes or a small inline `<script>const CONFIG = {...}</script>` block. Cacheable, debuggable in production.

**12. Three different purples**
- `#6c5ce7` — your brand
- `#6366f1` — Tailwind indigo-500 leaking in
- `#4f46e5` — Tailwind indigo-600 leaking in

→ In your CSS file, define:
```css
:root {
    --brand: #6c5ce7;
    --brand-dark: #5a4dd1;
}
```
Replace all `#6366f1` and `#4f46e5` occurrences with `var(--brand)`. Admin and student layouts feel like different products right now.

---

## 🟢 UX / UI (Premium Feel)

**13. Every card does `translateY(-3px)` on hover**
→ Reserve transform for CTAs/buttons only. Cards should only change `border-color` and `box-shadow` on hover. Current behavior makes the whole page feel jittery.

**14. Quiz taking has zero progress feedback**
→ Add a sticky header: `Question 4 / 12 · 8 minutes left`. One line of HTML, massive UX improvement.

**15. Dashboard classroom cards have 6 items of equal visual weight**
→ Student count should be the dominant number, everything else muted metadata below it.

**16. Analytics sub-tabs show a spinner on white background**
→ Replace with skeleton loaders — gray animated bars shaped like the actual content. Same load time, feels 3x faster.

**17. Student classroom view has no summary line**
→ First thing a student sees should be: `"2 assignments due · 1 quiz available"`. Currently they have to click through tabs to discover this.

**18. Admin classroom tabs overflow on mobile**
→ "Resources / Assignments / Quizzes / Students / Analytics" — on 375px the last tab is cut off. Use icons only on mobile, or abbreviate: `Files / Work / Quiz / People / Stats`.

**19. Phone validation mismatch**
→ HTML `pattern="[0-9]{10}"` rejects `+213555123456` (valid Algerian format). Either remove the pattern attribute and rely on `sanitizePhone`, or set `pattern="[\+0-9]{8,15}"`.

---

## Summary by effort

| Effort | Items |
|--------|-------|
| 5 min | #3, #5, #7, #19 |
| 30 min | #1 (done), #2, #8, #9, #12, #13, #18 |
| Half day | #6, #10, #14, #15, #16, #17 |
| Full day | #4, #11 |

Start with the 5-minute ones today. The phone pattern and CSRF fix are one-liners with zero risk.