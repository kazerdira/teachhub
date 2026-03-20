package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"teachhub/handlers"
	"teachhub/i18n"
	"teachhub/middleware"
	"teachhub/store"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Config from env
	dbURL := envOr("DATABASE_URL", "postgres://teachhub:teachhub@localhost:5433/teachhub?sslmode=disable")
	port := envOr("PORT", "8080")
	sessionSecret := envOr("SESSION_SECRET", "change-me-in-production-32chars!")
	adminUser := envOr("ADMIN_USER", "admin")
	adminPass := envOr("ADMIN_PASS", "admin123")
	baseURL := envOr("BASE_URL", "http://localhost:"+port)
	uploadDir := envOr("UPLOAD_DIR", "./uploads")
	claudeKey := os.Getenv("ANTHROPIC_API_KEY")
	lkApiKey := envOr("LIVEKIT_API_KEY", "teachhub")
	lkApiSecret := envOr("LIVEKIT_API_SECRET", "teachhubsecretkey1234567890abc")
	lkUrl := envOr("LIVEKIT_URL", "ws://localhost:7880")

	// In production (GIN_MODE=release), require critical secrets
	if os.Getenv("GIN_MODE") == "release" {
		missing := []string{}
		if os.Getenv("SESSION_SECRET") == "" {
			missing = append(missing, "SESSION_SECRET")
		}
		if os.Getenv("ADMIN_PASS") == "" {
			missing = append(missing, "ADMIN_PASS")
		}
		if os.Getenv("PLATFORM_PASS") == "" {
			missing = append(missing, "PLATFORM_PASS")
		}
		if os.Getenv("BASE_URL") == "" {
			missing = append(missing, "BASE_URL")
		}
		if len(missing) > 0 {
			log.Fatalf("FATAL: Production mode requires these env vars: %s", strings.Join(missing, ", "))
		}
	}

	// DB
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer pool.Close()

	// Run schema
	schema, err := os.ReadFile("db/schema.sql")
	if err != nil {
		log.Fatalf("Cannot read schema: %v", err)
	}
	if _, err := pool.Exec(context.Background(), string(schema)); err != nil {
		// Schema might already exist, log but don't fatal
		log.Printf("Schema exec note: %v", err)
	}

	// Store
	s := store.New(pool)

	// Ensure admin account
	hashed, _ := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	s.CreateAdmin(context.Background(), adminUser, string(hashed))

	// Ensure platform admin account
	platformUser := envOr("PLATFORM_USER", "owner")
	platformPass := envOr("PLATFORM_PASS", "owner123")
	platformHashed, _ := bcrypt.GenerateFromPassword([]byte(platformPass), bcrypt.DefaultCost)
	s.CreatePlatformAdmin(context.Background(), platformUser, string(platformHashed))

	// Sessions
	isProduction := os.Getenv("GIN_MODE") == "release"
	middleware.Init(sessionSecret, isProduction)

	// Templates
	funcMap := template.FuncMap{
		"add":        func(a, b int) int { return a + b },
		"sub":        func(a, b int) int { return a - b },
		"mul":        func(a, b int) int { return a * b },
		"div":        func(a, b int) int { if b == 0 { return 0 }; return a / b },
		"fileSizeMB": func(bytes int64) int64 { return bytes / (1024 * 1024) },
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"deref": func(p *int) int {
			if p == nil {
				return 0
			}
			return *p
		},
		"derefFloat": func(p *float64) float64 {
			if p == nil {
				return 0
			}
			return *p
		},
		"notNil": func(p interface{}) bool {
			return p != nil
		},
		"mapGet": func(m map[string]string, key string) string {
			if m == nil {
				return ""
			}
			return m[key]
		},
		"fileMapGet": func(m map[string]map[string]string, key string) map[string]string {
			if m == nil {
				return nil
			}
			return m[key]
		},
		"formatGrade": func(g *float64) string {
			if g == nil {
				return "—"
			}
			if *g == float64(int(*g)) {
				return fmt.Sprintf("%.0f", *g)
			}
			return fmt.Sprintf("%.1f", *g)
		},
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i + 1
			}
			return s
		},
		"divf": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mulf": func(a, b float64) float64 { return a * b },
		"tof":  func(a int) float64 { return float64(a) },
		"int":  func(a int64) int { return int(a) },
		"index": func(arr [5]int, i int) int {
			if i < 0 || i >= 5 {
				return 0
			}
			return arr[i]
		},
		"formatDuration": func(started time.Time, finished *time.Time) string {
			if finished == nil {
				return "—"
			}
			dur := finished.Sub(started)
			mins := int(dur.Minutes())
			secs := int(dur.Seconds()) % 60
			if mins > 0 {
				return fmt.Sprintf("%dm %ds", mins, secs)
			}
			return fmt.Sprintf("%ds", secs)
		},
		"pctInt": func(score, maxScore *int) int {
			if score == nil || maxScore == nil || *maxScore == 0 {
				return 0
			}
			return *score * 100 / *maxScore
		},
		"t": func(lang, key string) string {
			return i18n.T(lang, key)
		},
		"csrfField": func(token string) template.HTML {
			return template.HTML(`<input type="hidden" name="_csrf" value="` + template.HTMLEscapeString(token) + `">`)
		},
	}

	tmpl := template.New("").Funcs(funcMap)
	// Go's ParseGlob doesn't support **, so load each directory
	for _, pattern := range []string{
		"templates/layouts/*.html",
		"templates/admin/*.html",
		"templates/student/*.html",
		"templates/platform/*.html",
		"templates/landing.html",
	} {
		tmpl, err = tmpl.ParseGlob(pattern)
		if err != nil {
			log.Fatalf("Template parse failed (%s): %v", pattern, err)
		}
	}

	// Upload dirs
	os.MkdirAll(uploadDir+"/resources", 0755)
	os.MkdirAll(uploadDir+"/submissions", 0755)
	os.MkdirAll(uploadDir+"/quiz_files", 0755)

	// Secret platform path (configurable via env, defaults to a random-looking slug)
	platformPath := envOr("PLATFORM_PATH", "/ctrl-p-8x3kf")

	// Handler
	h := handlers.New(s, tmpl, uploadDir, baseURL, claudeKey, lkApiKey, lkApiSecret, lkUrl, platformPath)

	// Rate limiters: 5 attempts then block for 15 minutes
	loginRL := middleware.NewRateLimiter(5, 15*time.Minute)

	// Router
	if isProduction {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.MaxMultipartMemory = 64 << 20 // 64 MB
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CSRFProtection())

	// Static file serving for uploads (resources download handled by handler)
	r.Static("/static", "./static")

	// ─── Platform: Teacher Application (public) ─────────
	r.GET("/apply", h.ApplyPage)
	r.POST("/apply", h.ApplySubmit)
	r.GET("/apply/success", h.ApplySuccess)

	// ─── Language Toggle ─────────────────────────────────
	r.GET("/set-lang", func(c *gin.Context) {
		lang := c.Query("lang")
		if lang != "fr" && lang != "en" {
			lang = "en"
		}
		c.SetCookie("lang", lang, 60*60*24*365, "/", "", isProduction, false)
		ref := c.Request.Referer()
		if ref == "" {
			ref = "/"
		}
		c.Redirect(http.StatusFound, ref)
	})

	// ─── Platform Admin Routes ──────────────────────────
	r.GET(platformPath+"/login", h.PlatformLoginPage)
	r.POST(platformPath+"/login", middleware.RateLimit(loginRL), h.PlatformLogin)
	r.GET(platformPath+"/logout", h.PlatformLogout)

	platformAdmin := r.Group(platformPath, middleware.PlatformAdminRequired(platformPath))
	{
		platformAdmin.GET("", h.PlatformDashboard)
		platformAdmin.GET("/applications", h.PlatformApplications)
		platformAdmin.GET("/applications/:id", h.PlatformAppDetail)
		platformAdmin.POST("/applications/:id/status", h.PlatformUpdateAppStatus)
		platformAdmin.GET("/applications/:id/credentials", h.PlatformCredentials)
		platformAdmin.GET("/teachers", h.PlatformTeachers)
		platformAdmin.GET("/teachers/:id", h.PlatformTeacherDetail)
		platformAdmin.POST("/teachers/:id/toggle", h.PlatformToggleTeacher)
		platformAdmin.POST("/teachers/:id/reset-password", h.PlatformResetPassword)
		platformAdmin.GET("/teachers/:id/credentials", h.PlatformTeacherCredentials)
		platformAdmin.POST("/teachers/:id/subscription", h.PlatformExtendSubscription)
		platformAdmin.POST("/teachers/:id/payment", h.PlatformRecordPayment)
		platformAdmin.POST("/teachers/:id/payment/:paymentId/delete", h.PlatformDeletePayment)
		platformAdmin.GET("/analytics", h.PlatformAnalytics)
		platformAdmin.GET("/export/teachers", h.PlatformExportTeachersCSV)
		platformAdmin.GET("/export/payments", h.PlatformExportPaymentsCSV)
		platformAdmin.GET("/password", h.PlatformChangePasswordPage)
		platformAdmin.POST("/password", h.PlatformChangePassword)
	}

	// ─── Student Routes ─────────────────────────────────
	studentMw := middleware.StudentFromSession(s)

	r.GET("/", studentMw, h.Home)
	r.GET("/join/:code", studentMw, h.JoinPage)
	r.POST("/join/:code", studentMw, h.JoinClassroom)
	r.POST("/join-by-code", studentMw, func(c *gin.Context) {
		code := strings.TrimSpace(c.PostForm("code"))
		c.Redirect(http.StatusFound, "/join/"+code)
	})
	r.GET("/resource/:resId/download", studentMw, h.DownloadResource)

	// Student authenticated routes
	studentAuth := r.Group("/classroom", studentMw, middleware.StudentRequired())
	{
		studentAuth.GET("/:id", h.StudentClassroom)
		studentAuth.GET("/:id/dashboard", h.StudentDashboard)
		studentAuth.GET("/:id/live", h.StudentLivePage)
		studentAuth.POST("/:id/live/leave", h.StudentLiveLeave)
		studentAuth.GET("/:id/assignment/:assignId", h.StudentAssignment)
		studentAuth.POST("/:id/assignment/:assignId/submit", h.StudentSubmit)
		studentAuth.GET("/:id/quiz/:quizId", h.StudentQuiz)
		studentAuth.POST("/:id/quiz/:quizId/submit", h.StudentSubmitQuiz)
	}

	// ─── Admin Routes ───────────────────────────────────
	r.GET("/admin/login", h.AdminLoginPage)
	r.POST("/admin/login", middleware.RateLimit(loginRL), h.AdminLogin)
	r.GET("/admin/logout", h.AdminLogout)

	admin := r.Group("/admin", middleware.AdminRequired(), middleware.AdminSubscriptionCheck(s))
	{
		admin.GET("", h.AdminDashboard)

		// Classrooms
		admin.POST("/classroom", h.CreateClassroom)
		admin.POST("/classroom/:id/delete", h.DeleteClassroom)
		admin.POST("/classroom/:id/regenerate-code", h.RegenerateCode)
		admin.GET("/classroom/:id", h.AdminClassroom)
		admin.GET("/classroom/:id/analytics", h.AdminAnalytics)
		admin.GET("/classroom/:id/report", h.AdminClassroomReport)
		admin.GET("/classroom/:id/export/roster", h.ExportRosterCSV)
		admin.GET("/classroom/:id/export/quizzes", h.ExportQuizzesCSV)
		admin.GET("/classroom/:id/export/assignments", h.ExportAssignmentsCSV)
		admin.GET("/classroom/:id/export/attendance", h.ExportAttendanceCSV)
		admin.GET("/classroom/:id/student/:studentId", h.AdminStudentDetail)
		admin.POST("/classroom/:id/student/:studentId/remark", h.AdminAddRemark)
		admin.POST("/classroom/:id/student/:studentId/remark/:remarkId/delete", h.AdminDeleteRemark)

		// Categories
		admin.POST("/classroom/:id/category", h.CreateCategory)
		admin.POST("/classroom/:id/category/:catId/delete", h.DeleteCategory)

		// Resources
		admin.POST("/classroom/:id/resource", h.UploadResource)
		admin.POST("/classroom/:id/resource/:resId/delete", h.DeleteResource)

		// Assignments
		admin.POST("/classroom/:id/assignment", h.CreateAssignment)
		admin.POST("/classroom/:id/assignment/:assignId/delete", h.DeleteAssignment)
		admin.GET("/classroom/:id/assignment/:assignId/submissions", h.ViewSubmissions)
		admin.POST("/classroom/:id/assignment/:assignId/submission/:subId/review", h.ReviewSubmission)
		admin.GET("/submission/:subId/download", h.DownloadSubmission)

		// Students
		admin.POST("/classroom/:id/student/:studentId/remove", h.RemoveStudent)
		admin.POST("/classroom/:id/student/:studentId/approve", h.ApproveStudent)
		admin.POST("/classroom/:id/student/:studentId/reject", h.RejectStudent)

		// Live Class
		admin.POST("/classroom/:id/live/start", h.StartLiveClass)
		admin.POST("/classroom/:id/live/end", h.EndLiveClass)
		admin.GET("/classroom/:id/live", h.AdminLivePage)

		// Allowed Students (pre-registration)
		admin.POST("/classroom/:id/allowed", h.AddAllowedStudent)
		admin.POST("/classroom/:id/allowed/bulk", h.AddAllowedStudentsBulk)
		admin.POST("/classroom/:id/allowed/:allowedId/delete", h.DeleteAllowedStudent)

		// Quizzes
		admin.POST("/classroom/:id/quiz", h.CreateQuiz)
		admin.POST("/classroom/:id/quiz/:quizId/delete", h.DeleteQuiz)
		admin.POST("/classroom/:id/quiz/:quizId/toggle", h.ToggleQuizPublish)
		admin.GET("/classroom/:id/quiz/:quizId/edit", h.EditQuiz)
		admin.POST("/classroom/:id/quiz/:quizId/question", h.AddQuestion)
		admin.POST("/classroom/:id/quiz/:quizId/question/:qId/delete", h.DeleteQuestion)
		admin.POST("/classroom/:id/quiz/:quizId/question/:qId/update", h.UpdateQuestion)
		admin.POST("/classroom/:id/quiz/:quizId/settings", h.UpdateQuizSettings)
		admin.POST("/classroom/:id/quiz/:quizId/generate", h.GenerateQuizAI)
		admin.POST("/classroom/:id/quiz/:quizId/attempt/:attemptId/review", h.ReviewAttempt)
	}

	log.Printf("📚 TeachHub running on port %s", port)

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
