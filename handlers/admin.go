package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"teachhub/geo"
	"teachhub/middleware"
	"teachhub/store"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Store        *store.Store
	Tmpl         *template.Template
	UploadDir    string
	BaseURL      string
	ClaudeKey    string
	LKApiKey     string
	LKApiSecret  string
	LKUrl        string
	PlatformPath string
}

func New(s *store.Store, tmpl *template.Template, uploadDir, baseURL, claudeKey, lkKey, lkSecret, lkUrl, platformPath string) *Handler {
	return &Handler{
		Store: s, Tmpl: tmpl, UploadDir: uploadDir, BaseURL: baseURL, ClaudeKey: claudeKey,
		LKApiKey: lkKey, LKApiSecret: lkSecret, LKUrl: lkUrl, PlatformPath: platformPath,
	}
}

func (h *Handler) render(c *gin.Context, tmplName string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["BaseURL"] = h.BaseURL
	// Language
	lang, err := c.Cookie("lang")
	if err != nil || (lang != "en" && lang != "fr") {
		lang = "en"
	}
	data["Lang"] = lang
	// CSRF token
	if t, exists := c.Get("csrf_token"); exists {
		data["CSRFToken"] = t
	}
	// Inject pending join requests count for admin pages
	if aid := adminID(c); aid > 0 {
		if _, exists := data["PendingRequests"]; !exists {
			data["PendingRequests"] = h.Store.CountPendingJoinRequests(c.Request.Context(), aid)
		}
	}
	// Inject admin role / center info for nav
	if admin, exists := c.Get("admin"); exists {
		a := admin.(*store.Admin)
		data["AdminRole"] = a.Role
		if a.CenterID != nil {
			data["AdminCenterID"] = *a.CenterID
		}
	}
	// Query params map for flash messages etc.
	if _, exists := data["Query"]; !exists {
		qm := map[string]string{}
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 {
				qm[k] = v[0]
			}
		}
		data["Query"] = qm
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := h.Tmpl.ExecuteTemplate(c.Writer, tmplName, data); err != nil {
		log.Printf("Template error [%s]: %v", tmplName, err)
		c.String(500, "An internal error occurred. Please try again.")
	}
}

// ─── Admin Login ────────────────────────────────────────

func adminID(c *gin.Context) int {
	if id, ok := c.Get("admin_id"); ok {
		if intID, ok := id.(int); ok {
			return intID
		}
	}
	return 0
}

// ownsClassroom verifies the logged-in admin owns the classroom (teacher) or is the center owner.
// Returns the classroomID if accessible, or 0 and aborts with redirect if not.
func (h *Handler) ownsClassroom(c *gin.Context) int {
	classID, _ := strconv.Atoi(c.Param("id"))
	_, _, err := h.Store.GetClassroomForAdminOrOwner(c.Request.Context(), classID, adminID(c))
	if err != nil {
		c.Redirect(http.StatusFound, "/admin")
		c.Abort()
		return 0
	}
	return classID
}

// ownsClassroomOwnerOnly verifies access AND requires owner role.
// Returns classroomID if the admin is the center owner, 0 otherwise.
func (h *Handler) ownsClassroomOwnerOnly(c *gin.Context) int {
	classID, _ := strconv.Atoi(c.Param("id"))
	admin := c.MustGet("admin").(*store.Admin)
	_, isOwnerAccess, err := h.Store.GetClassroomForAdminOrOwner(c.Request.Context(), classID, adminID(c))
	if err != nil {
		c.Redirect(http.StatusFound, "/admin")
		c.Abort()
		return 0
	}
	// For center classrooms, only owner can perform management actions
	if admin.CenterID != nil && admin.Role != "owner" && !isOwnerAccess {
		// Teacher owns this classroom directly but is a center teacher — still block management
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d", classID))
		c.Abort()
		return 0
	}
	return classID
}

func (h *Handler) AdminLoginPage(c *gin.Context) {
	h.render(c, "admin_login.html", gin.H{"Error": c.Query("error")})
}

func (h *Handler) AdminLogin(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")
	admin, err := h.Store.GetAdmin(c.Request.Context(), username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)) != nil {
		c.Redirect(http.StatusFound, "/admin/login?error=invalid")
		return
	}

	// Block deactivated teachers
	if !admin.Active {
		c.Redirect(http.StatusFound, "/admin/login?error=deactivated")
		return
	}

	// Check subscription: center-based or legacy individual
	if admin.CenterID != nil {
		center, err := h.Store.GetCenter(c.Request.Context(), *admin.CenterID)
		if err != nil {
			c.Redirect(http.StatusFound, "/admin/login?error=invalid")
			return
		}
		if center.SubscriptionEnd != nil && center.SubscriptionEnd.Before(time.Now()) && center.SubscriptionStatus == "active" {
			h.Store.UpdateCenterSubscription(c.Request.Context(), center.ID, "expired", center.SubscriptionStart, center.SubscriptionEnd)
			center.SubscriptionStatus = "expired"
		}
		if center.TrialEndsAt != nil && center.TrialEndsAt.Before(time.Now()) && center.SubscriptionStatus == "trial" {
			h.Store.UpdateCenterSubscription(c.Request.Context(), center.ID, "expired", center.SubscriptionStart, center.SubscriptionEnd)
			center.SubscriptionStatus = "expired"
		}
		if center.SubscriptionStatus != "active" && center.SubscriptionStatus != "trial" {
			errKey := "suspended"
			if center.SubscriptionStatus == "expired" {
				errKey = "expired"
			}
			c.Redirect(http.StatusFound, "/admin/login?error="+errKey)
			return
		}
	} else if admin.CreatedByPlatform {
		if admin.SubscriptionEnd != nil && admin.SubscriptionEnd.Before(time.Now()) && admin.SubscriptionStatus == "active" {
			h.Store.UpdateTeacherSubscription(c.Request.Context(), admin.ID, "expired")
			admin.SubscriptionStatus = "expired"
		}
		if admin.SubscriptionStatus != "active" {
			errKey := "suspended"
			if admin.SubscriptionStatus == "expired" {
				errKey = "expired"
			}
			c.Redirect(http.StatusFound, "/admin/login?error="+errKey)
			return
		}
	}
	middleware.SetAdminSession(c, admin.ID)
	// Record last login IP and time
	h.Store.UpdateAdminLastLogin(c.Request.Context(), admin.ID, c.ClientIP())
	// Clear pending password on first login so platform owner can no longer see it
	if admin.PendingPassword != nil {
		h.Store.ClearPendingPassword(c.Request.Context(), admin.ID)
	}
	// Redirect owners to center dashboard
	if admin.Role == "owner" && admin.CenterID != nil {
		c.Redirect(http.StatusFound, "/admin/center")
		return
	}
	c.Redirect(http.StatusFound, "/admin")
}

func (h *Handler) AdminLogout(c *gin.Context) {
	middleware.ClearAdminSession(c)
	c.Redirect(http.StatusFound, "/admin/login")
}

// ─── Admin Dashboard ────────────────────────────────────

func (h *Handler) AdminDashboard(c *gin.Context) {
	aid := adminID(c)
	admin, _ := h.Store.GetAdminByID(c.Request.Context(), aid)
	// Owners land on the center dashboard, not the teacher dashboard
	if admin != nil && admin.Role == "owner" && admin.CenterID != nil {
		c.Redirect(http.StatusFound, "/admin/center")
		return
	}
	classrooms, _ := h.Store.ListClassrooms(c.Request.Context(), aid)
	country := ""
	if admin != nil && admin.Country != "" {
		country = admin.Country
	}
	if country == "" {
		country = geo.CountryFromIP(c.ClientIP())
		if country == "" {
			country = "DZ"
		}
	}
	h.render(c, "admin_dashboard.html", gin.H{
		"Classrooms": classrooms,
		"Subjects":   geo.AllSubjects,
		"Levels":     geo.LevelsForCountry(country),
	})
}

// ─── Classroom CRUD ─────────────────────────────────────

func (h *Handler) CreateClassroom(c *gin.Context) {
	// Center teachers cannot create classrooms — owner does it via /center/classrooms
	if a, ok := c.Get("admin"); ok {
		if adm, ok := a.(*store.Admin); ok && adm.Role == "teacher" && adm.CenterID != nil {
			c.Redirect(http.StatusFound, "/admin")
			return
		}
	}
	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		c.Redirect(http.StatusFound, "/admin")
		return
	}
	subject := strings.TrimSpace(c.PostForm("subject"))
	level := strings.TrimSpace(c.PostForm("level"))
	h.Store.CreateClassroom(c.Request.Context(), name, subject, level, adminID(c))
	c.Redirect(http.StatusFound, "/admin")
}

func (h *Handler) UpdateClassroomTags(c *gin.Context) {
	// OWNER only for center classrooms
	classID := h.ownsClassroomOwnerOnly(c)
	if classID == 0 {
		return
	}
	subject := strings.TrimSpace(c.PostForm("subject"))
	level := strings.TrimSpace(c.PostForm("level"))
	h.Store.UpdateClassroomTagsAny(c.Request.Context(), classID, subject, level)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d", classID))
}

func (h *Handler) UpdateClassroomBilling(c *gin.Context) {
	// OWNER only for center classrooms
	classID := h.ownsClassroomOwnerOnly(c)
	if classID == 0 {
		return
	}
	rateStr := strings.TrimSpace(c.PostForm("session_rate"))
	rate, _ := strconv.ParseFloat(rateStr, 64)
	if rate < 0 {
		rate = 0
	}
	enabled := c.PostForm("billing_enabled") == "on"
	h.Store.UpdateClassroomBillingAny(c.Request.Context(), classID, rate, enabled)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d", classID))
}

func (h *Handler) DeleteClassroom(c *gin.Context) {
	// OWNER only for center classrooms
	classID := h.ownsClassroomOwnerOnly(c)
	if classID == 0 {
		return
	}
	h.Store.DeleteClassroomAny(c.Request.Context(), classID)
	c.Redirect(http.StatusFound, "/admin")
}

func (h *Handler) RegenerateCode(c *gin.Context) {
	// OWNER only for center classrooms
	classID := h.ownsClassroomOwnerOnly(c)
	if classID == 0 {
		return
	}
	h.Store.RegenerateJoinCodeAny(c.Request.Context(), classID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d", classID))
}

// ─── Classroom Detail (admin) ───────────────────────────

func (h *Handler) AdminClassroom(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	classroom, _, err := h.Store.GetClassroomForAdminOrOwner(c.Request.Context(), id, adminID(c))
	if err != nil {
		c.String(404, "Classroom not found")
		return
	}
	students, _ := h.Store.ListClassroomStudents(c.Request.Context(), id)
	categories, _ := h.Store.ListCategories(c.Request.Context(), id)
	resources, _ := h.Store.ListResources(c.Request.Context(), id)
	assignments, _ := h.Store.ListAssignments(c.Request.Context(), id)
	quizzes, _ := h.Store.ListQuizzes(c.Request.Context(), id)
	allowedStudents, _ := h.Store.ListAllowedStudents(c.Request.Context(), id)

	// Split students by status
	var approved, pending []store.Student
	for _, s := range students {
		switch s.MemberStatus {
		case "pending":
			pending = append(pending, s)
		default:
			approved = append(approved, s)
		}
	}

	tab := c.DefaultQuery("tab", "resources")
	liveSession, _ := h.Store.GetActiveLiveSession(c.Request.Context(), id)
	admin, _ := h.Store.GetAdminByID(c.Request.Context(), adminID(c))
	country := ""
	if admin != nil && admin.Country != "" {
		country = admin.Country
	}
	if country == "" {
		country = geo.CountryFromIP(c.ClientIP())
		if country == "" {
			country = "DZ"
		}
	}
	h.render(c, "admin_classroom.html", gin.H{
		"Classroom":       classroom,
		"Students":        approved,
		"PendingStudents": pending,
		"AllowedStudents": allowedStudents,
		"Categories":      categories,
		"Resources":       resources,
		"Assignments":     assignments,
		"Quizzes":         quizzes,
		"Tab":             tab,
		"JoinURL":         fmt.Sprintf("%s/join/%s", h.BaseURL, classroom.JoinCode),
		"LiveSession":     liveSession,
		"Subjects":        geo.AllSubjects,
		"Levels":          geo.LevelsForCountry(country),
	})
}

// ─── Categories ─────────────────────────────────────────

func (h *Handler) CreateCategory(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	name := strings.TrimSpace(c.PostForm("name"))
	if name != "" {
		h.Store.CreateCategory(c.Request.Context(), classID, name)
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=resources", classID))
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	catID, _ := strconv.Atoi(c.Param("catId"))
	h.Store.DeleteCategory(c.Request.Context(), catID, classID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=resources", classID))
}

// ─── Resources ──────────────────────────────────────────

func (h *Handler) UploadResource(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	title := strings.TrimSpace(c.PostForm("title"))
	desc := strings.TrimSpace(c.PostForm("description"))
	catIDStr := c.PostForm("category_id")
	externalURL := strings.TrimSpace(c.PostForm("external_url"))

	var catID *int
	if catIDStr != "" && catIDStr != "0" {
		v, _ := strconv.Atoi(catIDStr)
		catID = &v
	}

	var filePath, fileType string
	var fileSize int64

	file, header, err := c.Request.FormFile("file")
	if err == nil {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(header.Filename))

		// Block dangerous extensions
		if isBlockedExtension(ext) {
			c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=resources&error=blocked_type", classID))
			return
		}

		// Enforce size limit
		if header.Size > MaxTeacherFileSize {
			c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=resources&error=too_large", classID))
			return
		}

		fileType = detectFileType(ext)
		fname := fmt.Sprintf("%d_%d%s", classID, time.Now().UnixMilli(), ext)
		filePath = filepath.Join("resources", fname)
		fullPath := filepath.Join(h.UploadDir, filePath)
		os.MkdirAll(filepath.Dir(fullPath), 0755)

		dst, err := os.Create(fullPath)
		if err == nil {
			defer dst.Close()
			fileSize, _ = io.Copy(dst, file)
		}
	} else if externalURL != "" {
		fileType = "link"
	}

	h.Store.CreateResource(c.Request.Context(), classID, catID, title, desc, filePath, fileType, externalURL, fileSize)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=resources", classID))
}

func (h *Handler) DeleteResource(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	resID, _ := strconv.Atoi(c.Param("resId"))
	res, err := h.Store.GetResource(c.Request.Context(), resID)
	if err == nil && res.FilePath != "" {
		os.Remove(filepath.Join(h.UploadDir, res.FilePath))
	}
	h.Store.DeleteResource(c.Request.Context(), resID, classID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=resources", classID))
}

func (h *Handler) DownloadResource(c *gin.Context) {
	resID, _ := strconv.Atoi(c.Param("resId"))
	res, err := h.Store.GetResource(c.Request.Context(), resID)
	if err != nil {
		c.String(404, "Not found")
		return
	}

	// Admin who owns this classroom can always download
	aid := adminID(c)
	if aid == 0 {
		session, _ := middleware.SessionStore.Get(c.Request, "teachhub-admin")
		if session.Values["admin_id"] != nil {
			aid, _ = session.Values["admin_id"].(int)
		}
	}
	if aid > 0 {
		// Admin — skip student checks, serve directly
		if res.ExternalURL != "" {
			c.Redirect(http.StatusFound, res.ExternalURL)
			return
		}
		fullPath := filepath.Join(h.UploadDir, res.FilePath)
		c.FileAttachment(fullPath, filepath.Base(res.FilePath))
		return
	}

	// Student: must be enrolled in this classroom
	student := middleware.GetStudent(c)
	if student == nil {
		c.String(403, "Access denied")
		return
	}
	in, _ := h.Store.IsStudentInClassroom(c.Request.Context(), student.ID, res.ClassroomID)
	if !in {
		c.String(403, "Access denied")
		return
	}
	h.Store.TrackResourceView(c.Request.Context(), resID, student.ID)
	if res.ExternalURL != "" {
		c.Redirect(http.StatusFound, res.ExternalURL)
		return
	}
	fullPath := filepath.Join(h.UploadDir, res.FilePath)
	c.FileAttachment(fullPath, filepath.Base(res.FilePath))
}

// ─── Assignments ────────────────────────────────────────

func (h *Handler) CreateAssignment(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	title := strings.TrimSpace(c.PostForm("title"))
	desc := strings.TrimSpace(c.PostForm("description"))
	deadlineStr := c.PostForm("deadline")
	responseType := c.DefaultPostForm("response_type", "file")
	maxChars, _ := strconv.Atoi(c.DefaultPostForm("max_chars", "0"))
	maxFileSizeMB, _ := strconv.Atoi(c.DefaultPostForm("max_file_size", "10"))
	maxFileSize := int64(maxFileSizeMB) * 1024 * 1024
	maxGrade, _ := strconv.ParseFloat(c.DefaultPostForm("max_grade", "20"), 64)

	var deadline *time.Time
	if deadlineStr != "" {
		if t, err := time.Parse("2006-01-02T15:04", deadlineStr); err == nil {
			deadline = &t
		}
	}

	// Handle optional assignment file attachment
	var filePath, fileName string
	file, header, err := c.Request.FormFile("assignment_file")
	if err == nil {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if isBlockedExtension(ext) {
			c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=assignments&error=blocked_type", classID))
			return
		}
		if header.Size > MaxTeacherFileSize {
			c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=assignments&error=too_large", classID))
			return
		}
		fileName = header.Filename
		fname := fmt.Sprintf("assign_%d_%d%s", classID, time.Now().UnixMilli(), ext)
		filePath = filepath.Join("assignments", fname)
		fullPath := filepath.Join(h.UploadDir, filePath)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		dst, err := os.Create(fullPath)
		if err == nil {
			defer dst.Close()
			io.Copy(dst, file)
		}
	}

	h.Store.CreateAssignment(c.Request.Context(), classID, title, desc, deadline, responseType, maxChars, maxFileSize, maxGrade, filePath, fileName)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=assignments", classID))
}

func (h *Handler) DeleteAssignment(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	aID, _ := strconv.Atoi(c.Param("assignId"))
	h.Store.DeleteAssignment(c.Request.Context(), aID, classID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=assignments", classID))
}

func (h *Handler) EditAssignment(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	aID, _ := strconv.Atoi(c.Param("assignId"))

	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=assignments", classID))
		return
	}
	desc := strings.TrimSpace(c.PostForm("description"))
	responseType := c.DefaultPostForm("response_type", "file")
	maxChars, _ := strconv.Atoi(c.DefaultPostForm("max_chars", "0"))
	maxFileSizeMB, _ := strconv.Atoi(c.DefaultPostForm("max_file_size", "10"))
	maxFileSize := int64(maxFileSizeMB) * 1024 * 1024
	maxGrade, _ := strconv.ParseFloat(c.DefaultPostForm("max_grade", "20"), 64)

	var deadline *time.Time
	if dStr := c.PostForm("deadline"); dStr != "" {
		if t, err := time.Parse("2006-01-02T15:04", dStr); err == nil {
			deadline = &t
		}
	}

	// Get existing assignment to preserve file if no new one uploaded
	existing, err := h.Store.GetAssignment(c.Request.Context(), aID)
	if err != nil || existing == nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=assignments", classID))
		return
	}
	filePath := existing.FilePath
	fileName := existing.FileName

	// Check if user wants to remove existing file
	if c.PostForm("remove_file") == "1" {
		if filePath != "" {
			os.Remove(filepath.Join(h.UploadDir, filePath))
		}
		filePath = ""
		fileName = ""
	}

	// Handle new file upload (replaces existing)
	file, header, err := c.Request.FormFile("assignment_file")
	if err == nil {
		defer file.Close()
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if isBlockedExtension(ext) {
			c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=assignments&error=blocked_type", classID))
			return
		}
		if header.Size > MaxTeacherFileSize {
			c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=assignments&error=too_large", classID))
			return
		}
		// Remove old file if replacing
		if filePath != "" {
			os.Remove(filepath.Join(h.UploadDir, filePath))
		}
		fileName = header.Filename
		fname := fmt.Sprintf("assign_%d_%d%s", classID, time.Now().UnixMilli(), ext)
		filePath = filepath.Join("assignments", fname)
		fullPath := filepath.Join(h.UploadDir, filePath)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		dst, err := os.Create(fullPath)
		if err == nil {
			defer dst.Close()
			io.Copy(dst, file)
		}
	}

	h.Store.UpdateAssignment(c.Request.Context(), aID, classID, title, desc, deadline, responseType, maxChars, maxFileSize, maxGrade, filePath, fileName)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=assignments", classID))
}

func (h *Handler) DownloadAssignmentFile(c *gin.Context) {
	assignID, _ := strconv.Atoi(c.Param("assignId"))
	assign, err := h.Store.GetAssignment(c.Request.Context(), assignID)
	if err != nil || assign.FilePath == "" {
		c.String(404, "Not found")
		return
	}
	// Auth: must be an admin who owns this classroom OR an enrolled student
	aid := adminID(c)
	if aid > 0 {
		_, err := h.Store.GetClassroomForAdmin(c.Request.Context(), assign.ClassroomID, aid)
		if err != nil {
			c.String(403, "Access denied")
			return
		}
	} else {
		student := middleware.GetStudent(c)
		if student == nil {
			c.String(403, "Access denied")
			return
		}
		in, _ := h.Store.IsStudentInClassroom(c.Request.Context(), student.ID, assign.ClassroomID)
		if !in {
			c.String(403, "Access denied")
			return
		}
	}
	fullPath := filepath.Join(h.UploadDir, assign.FilePath)
	c.FileAttachment(fullPath, assign.FileName)
}

func (h *Handler) ViewSubmissions(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	assignID, _ := strconv.Atoi(c.Param("assignId"))
	classroom, _ := h.Store.GetClassroom(c.Request.Context(), classID)
	assignment, _ := h.Store.GetAssignment(c.Request.Context(), assignID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	const perPage = 50
	offset := (page - 1) * perPage

	submissions, total, _ := h.Store.ListSubmissionsPaged(c.Request.Context(), assignID, perPage, offset)
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	h.render(c, "admin_submissions.html", gin.H{
		"Classroom":   classroom,
		"Assignment":  assignment,
		"Submissions": submissions,
		"Page":        page,
		"TotalPages":  totalPages,
		"Total":       total,
	})
}

func (h *Handler) ReviewSubmission(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	assignID, _ := strconv.Atoi(c.Param("assignId"))
	subID, _ := strconv.Atoi(c.Param("subId"))
	status := c.PostForm("status")
	feedback := c.PostForm("feedback")

	var grade *float64
	var maxGrade *float64
	if gradeStr := c.PostForm("grade"); gradeStr != "" {
		if g, err := strconv.ParseFloat(gradeStr, 64); err == nil {
			grade = &g
			// Get assignment max_grade
			assignment, aerr := h.Store.GetAssignment(c.Request.Context(), assignID)
			if aerr == nil {
				mg := assignment.MaxGrade
				maxGrade = &mg
			}
		}
	}

	h.Store.UpdateSubmissionStatus(c.Request.Context(), subID, classID, status, feedback, grade, maxGrade)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/assignment/%d/submissions", classID, assignID))
}

func (h *Handler) DownloadSubmission(c *gin.Context) {
	subID, _ := strconv.Atoi(c.Param("subId"))
	filePath, fileName, err := h.Store.GetSubmissionForDownload(c.Request.Context(), subID, adminID(c))
	if err != nil || filePath == "" {
		c.String(404, "Not found")
		return
	}
	fullPath := filepath.Join(h.UploadDir, filePath)
	c.FileAttachment(fullPath, fileName)
}

// ─── Remove Student ─────────────────────────────────────

func (h *Handler) RemoveStudent(c *gin.Context) {
	// OWNER only for center classrooms
	classID := h.ownsClassroomOwnerOnly(c)
	if classID == 0 {
		return
	}
	studentID, _ := strconv.Atoi(c.Param("studentId"))
	h.Store.RemoveStudentFromClassroom(c.Request.Context(), studentID, classID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=students", classID))
}

// ─── Allowed Students (Pre-registration) ───────────────

func (h *Handler) AddAllowedStudent(c *gin.Context) {
	// OWNER only for center classrooms
	classID := h.ownsClassroomOwnerOnly(c)
	if classID == 0 {
		return
	}
	email := strings.TrimSpace(c.PostForm("email"))
	name := strings.TrimSpace(c.PostForm("name"))
	if email != "" {
		h.Store.AddAllowedStudent(c.Request.Context(), classID, email, name)
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=students", classID))
}

func (h *Handler) AddAllowedStudentsBulk(c *gin.Context) {
	// OWNER only for center classrooms
	classID := h.ownsClassroomOwnerOnly(c)
	if classID == 0 {
		return
	}
	bulk := c.PostForm("bulk_emails")
	lines := strings.Split(bulk, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Support "email" or "email,name" format
		parts := strings.SplitN(line, ",", 2)
		email := strings.TrimSpace(parts[0])
		name := ""
		if len(parts) > 1 {
			name = strings.TrimSpace(parts[1])
		}
		if email != "" {
			h.Store.AddAllowedStudent(c.Request.Context(), classID, email, name)
		}
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=students", classID))
}

func (h *Handler) DeleteAllowedStudent(c *gin.Context) {
	// OWNER only for center classrooms
	classID := h.ownsClassroomOwnerOnly(c)
	if classID == 0 {
		return
	}
	allowedID, _ := strconv.Atoi(c.Param("allowedId"))
	h.Store.DeleteAllowedStudent(c.Request.Context(), allowedID, classID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=students", classID))
}

// ─── Student Approval ─────────────────────────────────

func (h *Handler) ApproveStudent(c *gin.Context) {
	// OWNER only for center classrooms
	classID := h.ownsClassroomOwnerOnly(c)
	if classID == 0 {
		return
	}
	studentID, _ := strconv.Atoi(c.Param("studentId"))
	h.Store.ApproveStudent(c.Request.Context(), studentID, classID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=students", classID))
}

func (h *Handler) RejectStudent(c *gin.Context) {
	// OWNER only for center classrooms
	classID := h.ownsClassroomOwnerOnly(c)
	if classID == 0 {
		return
	}
	studentID, _ := strconv.Atoi(c.Param("studentId"))
	h.Store.RejectStudent(c.Request.Context(), studentID, classID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=students", classID))
}

// ─── Helpers ────────────────────────────────────────────

// Maximum upload sizes
const (
	MaxTeacherFileSize = 50 * 1024 * 1024 // 50 MB for teacher resources
	MaxStudentFileSize = 20 * 1024 * 1024 // 20 MB for student submissions
)

// Blocked extensions — dangerous executable/script types
var blockedExtensions = map[string]bool{
	".exe": true, ".bat": true, ".cmd": true, ".com": true, ".msi": true,
	".scr": true, ".pif": true, ".vbs": true, ".vbe": true, ".js": true,
	".jse": true, ".ws": true, ".wsf": true, ".wsc": true, ".wsh": true,
	".ps1": true, ".psd1": true, ".psm1": true,
	".sh": true, ".bash": true, ".csh": true, ".ksh": true,
	".php": true, ".phtml": true, ".php3": true, ".php4": true, ".php5": true,
	".asp": true, ".aspx": true, ".jsp": true,
	".py": true, ".pl": true, ".rb": true, ".cgi": true,
	".dll": true, ".sys": true, ".drv": true,
	".app": true, ".action": true, ".command": true,
	".reg": true, ".inf": true, ".lnk": true, ".url": true,
	".hta": true, ".htm": true, ".html": true, ".svg": false, // svg allowed for resources
}

// Student submissions have a stricter whitelist
var studentAllowedExtensions = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true, ".odt": true,
	".ppt": true, ".pptx": true, ".xls": true, ".xlsx": true,
	".txt": true, ".rtf": true, ".csv": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	".mp4": true, ".webm": true, ".mov": true,
	".mp3": true, ".wav": true, ".m4a": true, ".ogg": true,
	".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true,
}

func isBlockedExtension(ext string) bool {
	return blockedExtensions[strings.ToLower(ext)]
}

func isStudentAllowedExtension(ext string) bool {
	return studentAllowedExtensions[strings.ToLower(ext)]
}

func detectFileType(ext string) string {
	switch ext {
	case ".pdf":
		return "pdf"
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return "image"
	case ".mp4", ".webm", ".mov", ".avi":
		return "video"
	case ".mp3", ".wav", ".ogg", ".m4a":
		return "audio"
	case ".doc", ".docx", ".ppt", ".pptx", ".xls", ".xlsx":
		return "document"
	default:
		return "file"
	}
}
