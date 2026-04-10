package main_test

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"teachhub/handlers"
	"teachhub/middleware"
	"teachhub/store"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// ─── Test Infrastructure ────────────────────────────────

type testEnv struct {
	Server *httptest.Server
	Store  *store.Store
	DB     *pgxpool.Pool
	URL    string
}

func setup(t *testing.T) *testEnv {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://teachhub:teachhub@localhost:5433/teachhub?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Skipping integration tests: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("Skipping integration tests: Postgres not reachable: %v", err)
	}

	// Create isolated test schema
	schemaName := fmt.Sprintf("test_%d", time.Now().UnixNano())
	if _, err := pool.Exec(ctx, fmt.Sprintf("CREATE SCHEMA %s", schemaName)); err != nil {
		pool.Close()
		t.Fatalf("Cannot create test schema: %v", err)
	}

	// All connections in this pool use the test schema
	pool.Close()
	pool, err = pgxpool.New(ctx, dbURL+"&search_path="+schemaName+",public")
	if err != nil {
		t.Fatalf("Cannot reconnect with test schema: %v", err)
	}

	// Apply schema.sql
	schemaSQL, err := os.ReadFile("db/schema.sql")
	if err != nil {
		pool.Close()
		t.Fatalf("Cannot read schema.sql: %v", err)
	}
	if _, err := pool.Exec(ctx, string(schemaSQL)); err != nil {
		pool.Close()
		t.Fatalf("Schema apply failed: %v", err)
	}

	s := store.New(pool)

	// Init session middleware
	middleware.Init("integ-test-secret-32-characters!", false)

	// Seed an admin
	hashed, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	s.CreateAdmin(ctx, "teacher1", string(hashed))

	// Minimal template stubs — we test status codes and redirects, not HTML
	tmpl := buildStubTemplates()

	uploadDir := t.TempDir()
	os.MkdirAll(uploadDir+"/resources", 0755)
	os.MkdirAll(uploadDir+"/submissions", 0755)
	os.MkdirAll(uploadDir+"/quiz_files", 0755)

	h := handlers.New(s, tmpl, uploadDir, "http://localhost", "", "", "", "", "/ctrl-p-test")

	// Build full router (mirrors main.go)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CSRFProtection())

	// Test-only endpoint: returns the CSRF token for integration tests
	r.GET("/test/csrf", func(c *gin.Context) {
		c.String(http.StatusOK, middleware.GetCSRFToken(c))
	})

	loginRL := middleware.NewRateLimiter(100, 1*time.Minute)
	studentMw := middleware.StudentFromSession(s)

	r.GET("/uploads/*filepath", studentMw, h.ServeUpload)
	r.GET("/", studentMw, h.Home)
	r.GET("/join/:code", studentMw, h.JoinPage)
	r.POST("/join/:code", studentMw, h.JoinClassroom)
	r.GET("/cgu", h.CGUPage)

	studentAuth := r.Group("/classroom", studentMw, middleware.StudentRequired())
	{
		studentAuth.GET("/:id", h.StudentClassroom)
		studentAuth.GET("/:id/assignment/:assignId", h.StudentAssignment)
		studentAuth.POST("/:id/assignment/:assignId/submit", h.StudentSubmit)
		studentAuth.GET("/:id/quiz/:quizId", h.StudentQuiz)
		studentAuth.POST("/:id/quiz/:quizId/submit", h.StudentSubmitQuiz)
	}

	r.GET("/admin/login", h.AdminLoginPage)
	r.POST("/admin/login", middleware.RateLimit(loginRL), h.AdminLogin)
	r.GET("/admin/logout", h.AdminLogout)

	admin := r.Group("/admin", middleware.AdminRequired(), middleware.AdminSubscriptionCheck(s))
	{
		admin.GET("", h.AdminDashboard)
		admin.POST("/classroom", h.CreateClassroom)
		admin.GET("/classroom/:id", h.AdminClassroom)
		admin.POST("/classroom/:id/assignment", h.CreateAssignment)
		admin.GET("/classroom/:id/assignment/:assignId/submissions", h.ViewSubmissions)
		admin.POST("/classroom/:id/quiz", h.CreateQuiz)
		admin.POST("/classroom/:id/quiz/:quizId/toggle", h.ToggleQuizPublish)
		admin.GET("/classroom/:id/quiz/:quizId/edit", h.EditQuiz)
		admin.POST("/classroom/:id/quiz/:quizId/question", h.AddQuestion)
		admin.POST("/classroom/:id/quiz/:quizId/settings", h.UpdateQuizSettings)
		admin.POST("/classroom/:id/student/:studentId/remove", h.RemoveStudent)
		admin.POST("/classroom/:id/student/:studentId/approve", h.ApproveStudent)
		admin.POST("/classroom/:id/delete", h.DeleteClassroom)
	}

	server := httptest.NewServer(r)

	t.Cleanup(func() {
		server.Close()
		pool.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA %s CASCADE", schemaName))
		pool.Close()
	})

	return &testEnv{Server: server, Store: s, DB: pool, URL: server.URL}
}

// ─── HTTP Helpers ───────────────────────────────────────

func newClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// csrf fetches the CSRF token using the test-only endpoint.
func (env *testEnv) csrf(t *testing.T, client *http.Client) string {
	t.Helper()
	resp, err := client.Get(env.URL + "/test/csrf")
	if err != nil {
		t.Fatalf("GET /test/csrf: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	token := strings.TrimSpace(string(b))
	if token == "" {
		t.Fatal("empty CSRF token")
	}
	return token
}

// post sends a POST with form data including the CSRF token.
func (env *testEnv) post(t *testing.T, client *http.Client, path string, data url.Values) *http.Response {
	t.Helper()
	token := env.csrf(t, client)
	data.Set("_csrf", token)
	resp, err := client.PostForm(env.URL+path, data)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

// get sends a GET request.
func (env *testEnv) get(t *testing.T, client *http.Client, path string) *http.Response {
	t.Helper()
	resp, err := client.Get(env.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}

func location(resp *http.Response) string {
	return resp.Header.Get("Location")
}

// ─── Admin Login Helper ─────────────────────────────────

func (env *testEnv) adminLogin(t *testing.T) *http.Client {
	t.Helper()
	client := newClient()
	resp := env.post(t, client, "/admin/login", url.Values{
		"username": {"teacher1"},
		"password": {"testpass"},
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("admin login expected 302, got %d", resp.StatusCode)
	}
	loc := location(resp)
	if !strings.HasPrefix(loc, "/admin") {
		t.Fatalf("admin login should redirect to /admin, got %s", loc)
	}
	return client
}

// ─── Student Join Helper ────────────────────────────────

func (env *testEnv) studentJoin(t *testing.T, joinCode, name, email string) *http.Client {
	t.Helper()
	client := newClient()
	resp := env.post(t, client, "/join/"+joinCode, url.Values{
		"name":  {name},
		"email": {email},
	})
	resp.Body.Close()
	// 302 = auto-approved (allow-listed), 200 = pending (rendered join page with status)
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		t.Fatalf("student join expected 302 or 200, got %d", resp.StatusCode)
	}
	return client
}

// studentJoinApproved joins a student and has the admin approve them.
func (env *testEnv) studentJoinApproved(t *testing.T, adminClient *http.Client, joinCode, name, email string, classID int) *http.Client {
	t.Helper()
	student := env.studentJoin(t, joinCode, name, email)

	// Find the student ID and approve
	students, _ := env.Store.ListClassroomStudents(context.Background(), classID)
	for _, s := range students {
		if s.Email == email && s.MemberStatus == "pending" {
			env.post(t, adminClient, fmt.Sprintf("/admin/classroom/%d/student/%d/approve", classID, s.ID), url.Values{}).Body.Close()
			break
		}
	}
	return student
}

// ═══════════════════════════════════════════════════════════
// SCENARIO 1: Full Admin Workflow
// ═══════════════════════════════════════════════════════════

func TestAdminWorkflow_LoginCreateClassAssignmentQuiz(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// Step 1: Admin login
	admin := env.adminLogin(t)

	// Step 2: Access dashboard
	resp := env.get(t, admin, "/admin")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("dashboard expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Step 3: Create classroom
	resp = env.post(t, admin, "/admin/classroom", url.Values{
		"name":    {"Math 101"},
		"subject": {"math"},
		"level":   {"1as"},
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("create classroom expected 302, got %d", resp.StatusCode)
	}

	classrooms, err := env.Store.ListClassrooms(ctx, 1)
	if err != nil || len(classrooms) == 0 {
		t.Fatalf("classroom not created: %v", err)
	}
	classID := classrooms[0].ID
	t.Logf("Created classroom %d (code=%s)", classID, classrooms[0].JoinCode)

	// Step 4: View classroom
	resp = env.get(t, admin, fmt.Sprintf("/admin/classroom/%d", classID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("classroom page expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Step 5: Create assignment
	resp = env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/assignment", classID), url.Values{
		"title":         {"Homework 1"},
		"description":   {"Solve exercises 1-5"},
		"response_type": {"text"},
		"max_chars":     {"500"},
		"max_grade":     {"20"},
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("create assignment expected 302, got %d", resp.StatusCode)
	}
	assignments, _ := env.Store.ListAssignments(ctx, classID)
	if len(assignments) == 0 {
		t.Fatal("assignment not created")
	}
	t.Logf("Created assignment %d", assignments[0].ID)

	// Step 6: Create quiz
	resp = env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/quiz", classID), url.Values{
		"title":       {"Chapter 1 Quiz"},
		"description": {"Test your knowledge"},
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("create quiz expected 302, got %d", resp.StatusCode)
	}
	quizzes, _ := env.Store.ListQuizzes(ctx, classID)
	if len(quizzes) == 0 {
		t.Fatal("quiz not created")
	}
	quizID := quizzes[0].ID

	// Step 7: Add question
	resp = env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/quiz/%d/question", classID, quizID), url.Values{
		"question_type":  {"mcq"},
		"content":        {"What is 2+2?"},
		"option_0":       {"3"},
		"option_1":       {"4"},
		"option_2":       {"5"},
		"correct_answer": {"1"},
		"points":         {"2"},
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("add question expected 302, got %d", resp.StatusCode)
	}
	questions, _ := env.Store.ListQuizQuestions(ctx, quizID)
	if len(questions) == 0 {
		t.Fatal("question not created")
	}

	// Step 8: Publish quiz
	resp = env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/quiz/%d/toggle", classID, quizID), url.Values{})
	resp.Body.Close()
	quiz, _ := env.Store.GetQuiz(ctx, quizID)
	if !quiz.Published {
		t.Fatal("quiz should be published")
	}

	t.Log("✅ Admin workflow: login → dashboard → classroom → assignment → quiz → publish")
}

// ═══════════════════════════════════════════════════════════
// SCENARIO 2: Student Workflow
// ═══════════════════════════════════════════════════════════

func TestStudentWorkflow_JoinViewSubmit(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// Setup: admin creates classroom + assignment + published quiz
	admin := env.adminLogin(t)
	env.post(t, admin, "/admin/classroom", url.Values{"name": {"Physics"}}).Body.Close()
	classrooms, _ := env.Store.ListClassrooms(ctx, 1)
	classID := classrooms[0].ID
	joinCode := classrooms[0].JoinCode

	env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/assignment", classID), url.Values{
		"title": {"Lab Report"}, "description": {"Write report"},
		"response_type": {"text"}, "max_chars": {"1000"}, "max_grade": {"20"},
	}).Body.Close()
	assignments, _ := env.Store.ListAssignments(ctx, classID)
	assignID := assignments[0].ID

	env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/quiz", classID), url.Values{
		"title": {"Physics Quiz"},
	}).Body.Close()
	quizzes, _ := env.Store.ListQuizzes(ctx, classID)
	quizID := quizzes[0].ID
	env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/quiz/%d/question", classID, quizID), url.Values{
		"question_type": {"true_false"}, "content": {"Gravity is 9.8"},
		"correct_answer": {"true"}, "points": {"1"},
	}).Body.Close()
	env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/quiz/%d/toggle", classID, quizID), url.Values{}).Body.Close()

	// ── Student joins and gets approved ──
	student := env.studentJoinApproved(t, admin, joinCode, "Alice", "alice@test.com", classID)
	students, _ := env.Store.ListClassroomStudents(ctx, classID)
	if len(students) == 0 {
		t.Fatal("student not in classroom")
	}
	studentID := students[0].ID

	// View classroom
	resp := env.get(t, student, fmt.Sprintf("/classroom/%d", classID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("classroom page expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// View + submit assignment
	resp = env.get(t, student, fmt.Sprintf("/classroom/%d/assignment/%d", classID, assignID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("assignment page expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = env.post(t, student, fmt.Sprintf("/classroom/%d/assignment/%d/submit", classID, assignID), url.Values{
		"text_content": {"My lab report content here."},
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("submit expected 302, got %d", resp.StatusCode)
	}
	subs, _ := env.Store.GetStudentSubmissions(ctx, assignID, studentID)
	if len(subs) == 0 {
		t.Fatal("submission not created")
	}

	// View + submit quiz
	resp = env.get(t, student, fmt.Sprintf("/classroom/%d/quiz/%d", classID, quizID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("quiz page expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = env.post(t, student, fmt.Sprintf("/classroom/%d/quiz/%d/submit", classID, quizID), url.Values{
		"answer_1": {"true"},
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("quiz submit expected 302, got %d", resp.StatusCode)
	}
	attempts, _ := env.Store.ListQuizAttempts(ctx, quizID)
	if len(attempts) == 0 {
		t.Fatal("quiz attempt not created")
	}

	t.Log("✅ Student workflow: join → classroom → assignment → submit → quiz → submit")
}

// ═══════════════════════════════════════════════════════════
// SCENARIO 3: Cross-Classroom Authorization
// ═══════════════════════════════════════════════════════════

func TestStudentAuth_CannotAccessOtherClassroom(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	admin := env.adminLogin(t)

	// Create two classrooms
	env.post(t, admin, "/admin/classroom", url.Values{"name": {"Class A"}}).Body.Close()
	env.post(t, admin, "/admin/classroom", url.Values{"name": {"Class B"}}).Body.Close()
	classrooms, _ := env.Store.ListClassrooms(ctx, 1)
	classA := classrooms[0]
	classB := classrooms[1]

	// Assignment + quiz in Class A
	env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/assignment", classA.ID), url.Values{
		"title": {"Private HW"}, "response_type": {"text"}, "max_chars": {"100"}, "max_grade": {"20"},
	}).Body.Close()
	assignsA, _ := env.Store.ListAssignments(ctx, classA.ID)

	env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/quiz", classA.ID), url.Values{"title": {"Quiz A"}}).Body.Close()
	quizzesA, _ := env.Store.ListQuizzes(ctx, classA.ID)
	env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/quiz/%d/question", classA.ID, quizzesA[0].ID), url.Values{
		"question_type": {"true_false"}, "content": {"Test?"}, "correct_answer": {"true"}, "points": {"1"},
	}).Body.Close()
	env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/quiz/%d/toggle", classA.ID, quizzesA[0].ID), url.Values{}).Body.Close()

	// Bob joins Class B only (and gets approved)
	bob := env.studentJoinApproved(t, admin, classB.JoinCode, "Bob", "bob@test.com", classB.ID)

	// Bob tries Class A's assignment — should NOT get 200
	resp := env.get(t, bob, fmt.Sprintf("/classroom/%d/assignment/%d", classA.ID, assignsA[0].ID))
	if resp.StatusCode == http.StatusOK {
		t.Error("SECURITY: Bob accessed Class A assignment!")
	}
	resp.Body.Close()

	// Bob tries Class A's quiz — should NOT get 200
	resp = env.get(t, bob, fmt.Sprintf("/classroom/%d/quiz/%d", classA.ID, quizzesA[0].ID))
	if resp.StatusCode == http.StatusOK {
		t.Error("SECURITY: Bob accessed Class A quiz!")
	}
	resp.Body.Close()

	// Bob CAN access Class B
	resp = env.get(t, bob, fmt.Sprintf("/classroom/%d", classB.ID))
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Bob should access his own classroom, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	t.Log("✅ Cross-classroom auth enforced")
}

// ═══════════════════════════════════════════════════════════
// SCENARIO 4: Unauthenticated Access Blocked
// ═══════════════════════════════════════════════════════════

func TestUnauthenticated_ProtectedRoutesBlocked(t *testing.T) {
	env := setup(t)
	anon := newClient()

	adminRoutes := []string{"/admin", "/admin/classroom/1"}
	for _, path := range adminRoutes {
		resp := env.get(t, anon, path)
		resp.Body.Close()
		if resp.StatusCode != http.StatusFound {
			t.Errorf("%s: expected redirect, got %d", path, resp.StatusCode)
			continue
		}
		if !strings.Contains(location(resp), "login") {
			t.Errorf("%s: expected redirect to login, got %s", path, location(resp))
		}
	}

	studentRoutes := []string{"/classroom/1", "/classroom/1/assignment/1", "/classroom/1/quiz/1"}
	for _, path := range studentRoutes {
		resp := env.get(t, anon, path)
		resp.Body.Close()
		if resp.StatusCode != http.StatusFound {
			t.Errorf("%s: expected redirect for anon student, got %d", path, resp.StatusCode)
		}
	}

	t.Log("✅ Unauthenticated access blocked")
}

// ═══════════════════════════════════════════════════════════
// SCENARIO 5: CSRF Enforcement (Integration)
// ═══════════════════════════════════════════════════════════

func TestCSRF_PostWithoutToken_Blocked(t *testing.T) {
	env := setup(t)
	client := newClient()

	// Establish session cookies
	env.get(t, client, "/cgu").Body.Close()

	// POST without CSRF token
	resp, err := client.PostForm(env.URL+"/admin/login", url.Values{
		"username": {"teacher1"},
		"password": {"testpass"},
	})
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("POST without CSRF should be 403, got %d", resp.StatusCode)
	}

	t.Log("✅ CSRF enforcement at integration level")
}

// ═══════════════════════════════════════════════════════════
// SCENARIO 6: Upload Auth
// ═══════════════════════════════════════════════════════════

func TestUploads_RequireAuth(t *testing.T) {
	env := setup(t)

	anon := newClient()
	resp := env.get(t, anon, "/uploads/resources/secret.pdf")
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("anon upload access should be 403, got %d", resp.StatusCode)
	}

	admin := env.adminLogin(t)
	resp = env.get(t, admin, "/uploads/resources/nonexistent.pdf")
	resp.Body.Close()
	// Should be 404 (not found) or similar, NOT 403
	if resp.StatusCode == http.StatusForbidden {
		t.Error("authenticated admin should not get 403 on uploads")
	}
	if resp.StatusCode == http.StatusOK {
		t.Error("nonexistent file should not return 200")
	}

	t.Log("✅ Upload auth enforced")
}

// ═══════════════════════════════════════════════════════════
// SCENARIO 7: Admin Isolation
// ═══════════════════════════════════════════════════════════

func TestAdminAuth_CannotAccessOtherAdminClassroom(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	// teacher2
	hashed2, _ := bcrypt.GenerateFromPassword([]byte("pass2"), bcrypt.DefaultCost)
	env.Store.CreateAdmin(ctx, "teacher2", string(hashed2))

	// teacher1 creates classroom
	t1 := env.adminLogin(t)
	env.post(t, t1, "/admin/classroom", url.Values{"name": {"T1 Class"}}).Body.Close()
	classrooms1, _ := env.Store.ListClassrooms(ctx, 1)
	classID1 := classrooms1[0].ID

	// teacher2 logs in
	t2 := newClient()
	env.post(t, t2, "/admin/login", url.Values{
		"username": {"teacher2"},
		"password": {"pass2"},
	}).Body.Close()

	// teacher2 tries teacher1's classroom
	resp := env.get(t, t2, fmt.Sprintf("/admin/classroom/%d", classID1))
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Error("SECURITY: teacher2 accessed teacher1's classroom!")
	}

	t.Log("✅ Admin isolation enforced")
}

// ═══════════════════════════════════════════════════════════
// SCENARIO 8: Full E2E Lifecycle
// ═══════════════════════════════════════════════════════════

func TestE2E_FullClassroomLifecycle(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	admin := env.adminLogin(t)
	env.post(t, admin, "/admin/classroom", url.Values{"name": {"Biology"}}).Body.Close()
	classrooms, _ := env.Store.ListClassrooms(ctx, 1)
	classID := classrooms[0].ID
	joinCode := classrooms[0].JoinCode

	// Create assignment
	env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/assignment", classID), url.Values{
		"title": {"Cell Diagram"}, "description": {"Draw a cell"},
		"response_type": {"text"}, "max_chars": {"2000"}, "max_grade": {"20"},
	}).Body.Close()
	assignments, _ := env.Store.ListAssignments(ctx, classID)
	assignID := assignments[0].ID

	// Student joins + gets approved + submits
	alice := env.studentJoinApproved(t, admin, joinCode, "Alice", "alice@bio.test", classID)
	students, _ := env.Store.ListClassroomStudents(ctx, classID)
	studentID := students[0].ID

	env.post(t, alice, fmt.Sprintf("/classroom/%d/assignment/%d/submit", classID, assignID), url.Values{
		"text_content": {"The cell has a nucleus and mitochondria."},
	}).Body.Close()

	// Admin views submissions
	resp := env.get(t, admin, fmt.Sprintf("/admin/classroom/%d/assignment/%d/submissions", classID, assignID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("submissions page expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	subs, _ := env.Store.GetStudentSubmissions(ctx, assignID, studentID)
	if len(subs) == 0 || !strings.Contains(subs[0].TextContent, "nucleus") {
		t.Fatal("submission missing or wrong content")
	}

	// Admin deletes classroom (cascading delete)
	env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/delete", classID), url.Values{}).Body.Close()
	remaining, _ := env.Store.ListClassrooms(ctx, 1)
	if len(remaining) != 0 {
		t.Error("classroom should be deleted")
	}

	t.Log("✅ E2E: create → join → submit → review → delete")
}

// ═══════════════════════════════════════════════════════════
// SCENARIO 9: Multiple Students
// ═══════════════════════════════════════════════════════════

func TestMultipleStudents_IndependentSubmissions(t *testing.T) {
	env := setup(t)
	ctx := context.Background()

	admin := env.adminLogin(t)
	env.post(t, admin, "/admin/classroom", url.Values{"name": {"History"}}).Body.Close()
	classrooms, _ := env.Store.ListClassrooms(ctx, 1)
	classID := classrooms[0].ID
	joinCode := classrooms[0].JoinCode

	env.post(t, admin, fmt.Sprintf("/admin/classroom/%d/assignment", classID), url.Values{
		"title": {"Essay"}, "response_type": {"text"}, "max_chars": {"5000"}, "max_grade": {"20"},
	}).Body.Close()
	assignments, _ := env.Store.ListAssignments(ctx, classID)
	assignID := assignments[0].ID

	alice := env.studentJoinApproved(t, admin, joinCode, "Alice", "alice@h.test", classID)
	bob := env.studentJoinApproved(t, admin, joinCode, "Bob", "bob@h.test", classID)
	charlie := env.studentJoinApproved(t, admin, joinCode, "Charlie", "charlie@h.test", classID)

	students, _ := env.Store.ListClassroomStudents(ctx, classID)
	if len(students) != 3 {
		t.Fatalf("expected 3 students, got %d", len(students))
	}

	env.post(t, alice, fmt.Sprintf("/classroom/%d/assignment/%d/submit", classID, assignID), url.Values{
		"text_content": {"Alice essay about Rome"},
	}).Body.Close()
	env.post(t, bob, fmt.Sprintf("/classroom/%d/assignment/%d/submit", classID, assignID), url.Values{
		"text_content": {"Bob essay about Greece"},
	}).Body.Close()
	env.post(t, charlie, fmt.Sprintf("/classroom/%d/assignment/%d/submit", classID, assignID), url.Values{
		"text_content": {"Charlie essay about Egypt"},
	}).Body.Close()

	allSubs, _ := env.Store.ListSubmissions(ctx, assignID)
	if len(allSubs) != 3 {
		t.Fatalf("expected 3 submissions, got %d", len(allSubs))
	}

	t.Log("✅ Multiple students: 3 independent submissions")
}

// ═══════════════════════════════════════════════════════════
// SCENARIO 10: Wrong Password
// ═══════════════════════════════════════════════════════════

func TestAdminLogin_WrongPassword(t *testing.T) {
	env := setup(t)
	client := newClient()

	resp := env.post(t, client, "/admin/login", url.Values{
		"username": {"teacher1"},
		"password": {"wrong-password"},
	})
	resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected redirect, got %d", resp.StatusCode)
	}
	loc := location(resp)
	if !strings.Contains(loc, "login") {
		t.Errorf("should redirect to login, got %s", loc)
	}

	// Dashboard still blocked
	resp = env.get(t, client, "/admin")
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		t.Error("SECURITY: dashboard accessible after failed login")
	}

	t.Log("✅ Wrong password rejected")
}

// ─── Template Stubs ─────────────────────────────────────

func buildStubTemplates() *template.Template {
	tmpl := template.New("").Funcs(template.FuncMap{
		"add":            func(a, b int) int { return a + b },
		"sub":            func(a, b int) int { return a - b },
		"mul":            func(a, b int) int { return a * b },
		"div":            func(a, b int) int { return a / b },
		"fileSizeMB":     func(b int64) int64 { return b / (1024 * 1024) },
		"lower":          strings.ToLower,
		"upper":          strings.ToUpper,
		"deref":          func(p *int) int { return 0 },
		"derefFloat":     func(p *float64) float64 { return 0 },
		"notNil":         func(p interface{}) bool { return p != nil },
		"mapGet":         func(m map[string]string, k string) string { return "" },
		"fileMapGet":     func(m map[string]map[string]string, k string) map[string]string { return nil },
		"formatGrade":    func(g *float64) string { return "—" },
		"seq":            func(n int) []int { return nil },
		"divf":           func(a, b float64) float64 { return 0 },
		"mulf":           func(a, b float64) float64 { return 0 },
		"tof":            func(a int) float64 { return 0 },
		"int":            func(a int64) int { return 0 },
		"index":          func(arr [5]int, i int) int { return 0 },
		"formatDuration": func(s time.Time, f *time.Time) string { return "" },
		"pctInt":         func(s, m *int) int { return 0 },
		"t":              func(lang, key string) string { return key },
		"csrfField":      func(token string) template.HTML { return template.HTML("") },
		"urlquery":       func(s string) string { return url.QueryEscape(s) },
		"safeContent":    func(s string) template.HTML { return template.HTML(s) },
		"letter":         func(i int) string { return "A" },
		"subjectInfo":    func(k string) interface{} { return nil },
		"contains":       func(slice []string, val string) bool { return false },
	})

	for _, name := range []string{
		"admin_login.html", "admin_dashboard.html", "admin_classroom.html",
		"admin_quiz_edit.html", "admin_submissions.html", "admin_student_detail.html",
		"admin_analytics.html", "admin_report.html", "admin_profile.html",
		"admin_requests.html", "admin_live.html",
		"student_home.html", "student_classroom.html", "student_assignment.html",
		"student_quiz.html", "student_join.html", "student_dashboard.html",
		"student_live.html",
		"landing.html", "cgu.html",
		"explore.html", "explore_teacher.html", "explore_join.html", "explore_join_success.html",
		"apply.html", "apply_success.html",
		"parent_report.html",
		"platform_login.html", "platform_dashboard.html", "platform_applications.html",
		"platform_app_detail.html", "platform_credentials.html",
		"platform_teachers.html", "platform_teacher_detail.html",
		"platform_teacher_credentials.html", "platform_analytics.html",
		"platform_password.html",
		"error_rate_limit.html",
	} {
		template.Must(tmpl.New(name).Parse("OK"))
	}
	return tmpl
}

var _ = strconv.Atoi // keep import
