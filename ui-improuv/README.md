# TeachHub UI Upgrade

Drop-in replacement for 3 files. No handler changes needed — everything cascades.

## Files to Replace

```
static/css/style.css                → replaces your current style.css
templates/layouts/admin.html        → replaces your current admin layout
templates/layouts/student.html      → replaces your current student layout
```

## What Changed

### Typography
- **Headings**: Plus Jakarta Sans (warm, geometric, modern)
- **Body**: DM Sans (clean, readable, friendlier than Inter)
- Both loaded from Google Fonts CDN

### Layout
- **Admin**: max-width bumped to `80rem` (was `max-w-5xl` = 64rem)
- **Student**: responsive breakpoints — 36rem mobile → 56rem tablet → 64rem desktop (was `max-w-lg` = 32rem, way too narrow)
- Better padding at each breakpoint

### Nav
- Richer gradient (deep indigo → violet)
- Sticky positioning
- Glass-effect shadow
- Logo with backdrop-blur icon
- Admin badge pill
- Better logout button with icon
- Student avatar initial in nav

### Cards
- Warmer border color (`#eceae6` instead of cold gray)
- Default `box-shadow` for depth (not flat)
- Larger border-radius (`16px`)
- Hover: indigo tint + glow + subtle lift
- Cards with forms/tables don't lift on hover (prevents jank)

### Buttons
- Refined gradient with inset highlight
- Better shadow system
- Active press-down effect

### Inputs
- 1.5px border (crisper than 1px)
- Warmer border colors
- Better focus ring with glow
- Hover state for discoverability

### Tables
- Sticky headers
- Better typography (smaller, uppercase, letter-spaced)
- Warmer row borders
- Refined hover background

### Badges
- Consistent sizing
- Inline-flex with gap for icon+text

### Tabs
- Animated underline on hover (grows from center)
- Thicker active indicator (2.5px)

### CSS Global Enhancements (style.css)
- Custom property system for colors, shadows, radii
- Warm-tinted background with subtle radial gradients
- Custom scrollbar styling
- Focus ring system
- Staggered card entrance animations
- Live class banner pulse glow
- Progress bar transitions
- Stat card top-accent on hover
- Tabular number rendering for stats
- Print styles
- Utility classes: `.glass`, `.text-gradient`, `.hover-lift`

### What Didn't Change
- All template names (`admin_head`, `admin_foot`, etc.) — identical
- All data bindings (`.Lang`, `.CSRFToken`, etc.) — identical
- All JavaScript functionality (Quill, MathLive, upload progress, service worker) — identical
- Live session layouts — same structure, just better fonts
- All page templates (dashboard, classroom, quiz, etc.) — untouched, they inherit the improvements
