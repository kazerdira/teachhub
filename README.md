# TeachHub

A lightweight teaching platform with live video sessions, quizzes, assignments, and analytics. Built for Algerian teachers and students — server-rendered HTML loads fast on any connection.

## Features

- **Classrooms** — Create separate spaces for different groups, each with a unique join code
- **Resource Library** — Upload PDFs, images, audio, video, or external links, organized by categories
- **Assignments** — Create assignments with deadlines and response types (file/text/both), review and grade submissions
- **Quizzes** — MCQ, True/False, Fill-in-the-blank (auto-graded) + Open-ended + File upload (manual review)
- **AI Quiz Generation** — Paste a topic → Claude generates quiz questions for you to review
- **Live Sessions** — Video/audio/screen-share via LiveKit SFU, with interactive whiteboard, chat, polls, hand-raise
- **Analytics** — Quiz stats, assignment trends, student roster, at-risk detection, attendance rates, resource views
- **Platform Admin** — Multi-tenant: manage teacher accounts, subscriptions, applications, payments
- **Bilingual** — Full English + French support
- **Mobile-first** — Students tap a link, enter their name, and they're in

## Quick Start

### Local Development

```bash
# 1. Start PostgreSQL
docker compose up -d db

# 2. Configure
cp .env.example .env
# Edit .env — defaults work for local dev

# 3. Run
go mod tidy
go run .
```

### Production (Docker Compose)

The full stack runs as 4 containers: app, postgres, livekit, caddy.

```bash
# 1. Clone and configure
git clone <repo> /opt/teachhub
cd /opt/teachhub
cp .env.example .env
# Edit .env with production values (DATABASE_URL, BASE_URL, LIVEKIT keys, etc.)

# 2. Launch everything
docker compose up -d --build

# 3. Update after changes
git pull origin main
docker compose up -d --build app
```

### Access

- **Admin panel**: `https://yourdomain.com/admin` (default: admin / admin123)
- **Student view**: `https://yourdomain.com`
- **Platform admin**: `https://yourdomain.com/ctrl-p-8x3kf/login`

## How It Works

### For you (admin)

1. Go to `/admin`, create a classroom
2. Add categories and upload resources
3. Create assignments and quizzes
4. Share the join link with students (e.g. `yoursite.com/join/a1b2c3d4`)
5. Review submissions and quiz results

### For students

1. Open the join link you shared
2. Enter their name → they're in the classroom
3. Browse resources, download files
4. Upload submissions for assignments
5. Take quizzes and see scores

## AI Quiz Generation

Set `ANTHROPIC_API_KEY` in your `.env` file. In the quiz editor, expand "Generate questions with AI", describe the topic, pick difficulty and question types, and hit Generate. Questions are added to the quiz for you to review/edit before publishing.

## Deployment

Production runs 4 Docker containers behind Caddy (auto-HTTPS):

| Container | Purpose |
| --------- | ------- |
| `app` | Go binary serving HTML + API |
| `db` | PostgreSQL 16 with persistent volume |
| `livekit` | LiveKit SFU for live video sessions |
| `caddy` | HTTPS reverse proxy with auto-TLS |

```bash
# Deploy updates
ssh root@your-vps "cd /opt/teachhub && git pull origin main && docker compose up -d --build app"
```

Zero CDN dependencies — all assets (Tailwind CSS, htmx, LiveKit SDK) are compiled and self-hosted.

## Project Structure

```text
teachhub/
├── main.go              # Entry point, routes, config, template funcs
├── handlers/
│   ├── admin.go         # Admin: login, classrooms, resources, assignments
│   ├── analytics.go     # Analytics dashboard (7 sub-tabs)
│   ├── dashboard.go     # Student dashboard
│   ├── export.go        # PDF/report export
│   ├── live.go          # Live session: join, leave, image/file upload
│   ├── platform.go      # Platform admin: teachers, applications, payments
│   ├── quiz.go          # Quiz CRUD, AI generation, attempt review
│   └── student.go       # Student: join, browse, submit, take quizzes
├── store/
│   └── store.go         # All database operations (~2500 lines)
├── middleware/
│   └── middleware.go     # Auth, CSRF, rate limiting, security headers
├── i18n/
│   └── i18n.go          # English + French translations
├── db/
│   └── schema.sql       # PostgreSQL schema (runs on startup, idempotent)
├── static/
│   ├── css/style.css    # Compiled Tailwind CSS (58KB)
│   └── js/
│       ├── htmx.min.js          # Self-hosted htmx (47KB)
│       └── livekit-client.umd.js # Self-hosted LiveKit SDK (469KB)
├── templates/
│   ├── layouts/          # Base HTML layouts (admin, student, platform, live)
│   ├── admin/            # Admin templates (9 pages)
│   ├── student/          # Student templates (7 pages)
│   └── platform/         # Platform admin templates (11 pages)
├── uploads/              # File storage (Docker volume, gitignored)
├── docker-compose.yml    # Production: app + postgres + livekit + caddy
├── Dockerfile            # Multi-stage Go build
├── Caddyfile             # HTTPS reverse proxy config
├── livekit.yaml          # LiveKit server config
├── SCALING.md            # Scaling guide for 50+ teachers
└── .env.example          # Environment config template
```

## Tech Stack

- **Backend**: Go 1.21 + Gin + PostgreSQL 16 (pgxpool)
- **Frontend**: Server-rendered HTML + Tailwind CSS (compiled, self-hosted) + HTMX (self-hosted)
- **Live Video**: LiveKit v1.10.0 SFU (self-hosted, livekit-client self-hosted)
- **Sessions**: Cookie-based (gorilla/sessions)
- **AI**: Anthropic Claude API (optional)
- **HTTPS**: Caddy with auto-TLS
- **Deploy**: Docker Compose (app + postgres + livekit + caddy)

## Scaling

See [SCALING.md](SCALING.md) for a detailed guide on what to optimize when you reach 50+ teachers — Redis caching, query timeouts, object storage, and more. Includes exact code snippets, estimated time, and priority order.
