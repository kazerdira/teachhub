# 🏢 TeachHub Platform Admin — Roadmap

> Transform TeachHub from a single-teacher tool into a **multi-teacher SaaS platform**
> with owner dashboard, teacher onboarding, and subscription management.

---

## Phase 1 — Foundation & Teacher Registration
> Public-facing teacher application + DB schema

**Database:**
- `platform_admin` table (id, username, password_hash, created_at)
- `teacher_application` table (id, full_name, email, phone, school_name, wilaya, message, status [pending/approved/rejected/contacted], admin_notes, created_at, reviewed_at)
- Modify `admin` table → add columns: email, school_name, subscription_status [active/expired/suspended], subscription_start, subscription_end, created_by_platform

**Pages:**
- `/apply` — Public teacher registration form (name, email, phone, school, wilaya, why they want to use TeachHub)
- `/apply/success` — Confirmation page ("We'll contact you soon")

**Result:** Teachers can submit applications from the public site.

---

## Phase 2 — Platform Admin Login & Dashboard
> The business owner gets their own protected area

**Auth:**
- `/platform/login` — Platform admin login page (separate from teacher login)
- Platform admin session/JWT (separate cookie from teacher session)
- Middleware: `PlatformAuthRequired`

**Dashboard Pages:**
- `/platform/` — Overview: pending applications count, active teachers count, expiring subscriptions count
- `/platform/applications` — List all teacher applications with filters (pending / approved / rejected / all)
- `/platform/applications/:id` — Application detail: view info, add internal notes, change status (approve / reject / contact)

**Layout:**
- `templates/platform/` folder with its own layout (`layouts/platform.html`)
- Clean dark theme to distinguish from teacher/student UI

**Result:** Platform owner can log in, review applications, approve/reject teachers.

---

## Phase 3 — Teacher Account Creation & Onboarding
> When an application is approved, create the teacher's account

**Workflow:**
- Platform admin clicks "Approve" on an application
- System generates a secure random password
- Creates entry in `admin` table with email, school, subscription_status=active
- Shows the generated credentials to platform admin (to share with teacher via email/phone)
- Optionally: send email with credentials (if SMTP configured)

**Pages:**
- Approval confirmation page showing generated credentials (copy button)
- `/platform/teachers` — List all active teachers (name, email, school, subscription status, classrooms count, students count)
- `/platform/teachers/:id` — Teacher detail: subscription info, usage stats, suspend/reactivate actions

**Middleware Update:**
- Teacher login (`/admin/login`) → check `subscription_status == active` before allowing login
- Show "Subscription expired — contact platform admin" if expired/suspended

**Result:** Full onboarding flow from application → approval → teacher account.

---

## Phase 4 — Subscription Management
> Track and manage teacher subscriptions

**Features:**
- Subscription plans: Monthly / Yearly (configurable by platform admin)
- `/platform/teachers/:id/subscription` — Extend, suspend, or reactivate subscription
- Subscription expiry tracking with visual indicators (active ✅, expiring soon ⚠️, expired ❌)
- Dashboard alerts: "3 teachers expiring this week"

**Auto-actions:**
- Daily check (or on login): if `subscription_end < now`, set status to `expired`
- Teacher sees "Your subscription expired on [date]" on login attempt

**Payment tracking:**
- `payment` table (id, teacher_id, amount, method [cash/ccp/baridi_mob/other], reference, date, notes)
- Platform admin logs payments manually (since Algeria has limited online payment)
- Payment history per teacher

**Result:** Full subscription lifecycle with manual payment tracking suited for Algeria.

---

## Phase 5 — Platform Analytics & Polish
> Business intelligence for the platform owner

**Analytics Dashboard:**
- Total teachers, total students across platform
- Applications trend (per week/month)
- Revenue tracking (total payments, monthly breakdown)
- Most active teachers (by student count, quiz count)
- Platform growth chart

**Quality of Life:**
- Platform admin can impersonate a teacher (view their dashboard as them — read-only)
- Export teachers list + payment history as CSV
- Application status email notifications (if SMTP)
- Platform admin password change
- Activity log: who was approved when, payments logged, etc.

**Result:** Platform owner has full visibility into the business.

---

## Summary

| Phase | What | Key Deliverable |
|-------|------|----------------|
| **1** | Foundation & Registration | Public `/apply` form + DB schema |
| **2** | Platform Login & Dashboard | Owner sees & manages applications |
| **3** | Teacher Onboarding | Approve → auto-create teacher account |
| **4** | Subscriptions & Payments | Expiry tracking + manual payment log |
| **5** | Analytics & Polish | Business dashboard + CSV exports |

Each phase builds on the previous one. The teacher + student experience stays unchanged.
