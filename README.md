# TeachHub

A lightweight teaching platform for sharing resources, collecting student submissions, and running quizzes. Built for teachers using HelloTalk or similar platforms who need a simple link-based system.

## Features

- **Classrooms** — Create separate spaces for different groups, each with a unique join code
- **Resource Library** — Upload PDFs, images, audio, video, or external links, organized by categories
- **Student Submissions** — Create assignments with deadlines, students upload files, you review and give feedback
- **Quizzes** — MCQ, True/False, Fill-in-the-blank (auto-graded) + Open-ended (manual review)
- **AI Quiz Generation** — Paste a topic or text → Claude generates quiz questions for you to review
- **Mobile-first** — Students tap a link in chat, enter their name, and they're in

## Quick Start

### 1. Prerequisites
- Go 1.22+
- PostgreSQL (or Docker)

### 2. Database
```bash
# Option A: Docker (easiest)
docker compose up -d

# Option B: Manual PostgreSQL
createdb teachhub
```

### 3. Configure
```bash
cp .env.example .env
# Edit .env with your settings (defaults work for local dev)
```

### 4. Run
```bash
go mod tidy
go run .
```

Or with Make:
```bash
make setup   # starts DB + app
```

### 5. Access
- **Admin panel**: http://localhost:8080/admin (default: admin / admin123)
- **Student view**: http://localhost:8080

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

Single Go binary + PostgreSQL. Deploy on any VPS:

```bash
# Build
go build -o teachhub .

# Run with env vars
DATABASE_URL=postgres://... BASE_URL=https://yourdomain.com ./teachhub
```

For file storage at scale, swap `UPLOAD_DIR` to an R2/S3 mounted path.

## Project Structure

```
teachhub/
├── main.go              # Entry point, routes, config
├── handlers/
│   ├── admin.go         # Admin: login, classrooms, resources, assignments
│   ├── quiz.go          # Admin: quiz CRUD, AI generation
│   └── student.go       # Student: join, browse, submit, take quizzes
├── store/
│   └── store.go         # All database operations
├── middleware/
│   └── middleware.go     # Session management, auth
├── db/
│   └── schema.sql       # PostgreSQL schema
├── templates/
│   ├── layouts/          # Base HTML layouts
│   ├── admin/            # Admin templates
│   └── student/          # Student templates
├── uploads/              # File storage (gitignored)
├── docker-compose.yml    # PostgreSQL container
├── .env.example          # Environment config
└── Makefile
```

## Tech Stack

- **Backend**: Go + Gin + PostgreSQL
- **Frontend**: Server-rendered HTML + Tailwind CSS (CDN) + HTMX
- **Sessions**: Cookie-based (gorilla/sessions)
- **AI**: Anthropic Claude API (optional)
