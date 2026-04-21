package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"teachhub/geo"
	"teachhub/store"
)

// ─── Center Dashboard ───────────────────────────────────

func (h *Handler) CenterDashboard(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	center, err := h.Store.GetCenter(c.Request.Context(), *admin.CenterID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin")
		return
	}
	teachers, _ := h.Store.ListCenterTeachers(c.Request.Context(), center.ID)
	stats, _ := h.Store.GetCenterStats(c.Request.Context(), center.ID)
	hasUnpaid := h.Store.HasUnpaidCenterInvoice(c.Request.Context(), center.ID)

	activeCount := 0
	for _, t := range teachers {
		if t.Active {
			activeCount++
		}
	}
	now := time.Now()
	nextInvoice := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	h.render(c, "center_dashboard.html", gin.H{
		"Center":             center,
		"Stats":              stats,
		"Teachers":           teachers,
		"ActiveTeacherCount": activeCount,
		"MonthlyTotal":       float64(activeCount) * center.PricePerTeacher,
		"HasUnpaidInvoice":   hasUnpaid,
		"NextInvoiceDate":    nextInvoice,
	})
}

// ─── Center Teachers ────────────────────────────────────

func (h *Handler) CenterTeachers(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	center, err := h.Store.GetCenter(c.Request.Context(), *admin.CenterID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin")
		return
	}
	teachers, _ := h.Store.ListCenterTeachers(c.Request.Context(), center.ID)

	h.render(c, "center_teachers.html", gin.H{
		"Center":   center,
		"Teachers": teachers,
		"Error":    c.Query("error"),
		"Created":  c.Query("created"),
		"Reset":    c.Query("reset"),
	})
}

func (h *Handler) CenterCreateTeacher(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	ctx := c.Request.Context()

	center, _ := h.Store.GetCenter(ctx, *admin.CenterID)
	if center == nil {
		c.Redirect(http.StatusFound, "/admin")
		return
	}

	displayName := strings.TrimSpace(c.PostForm("display_name"))
	email := strings.TrimSpace(c.PostForm("email"))
	phone := strings.TrimSpace(c.PostForm("phone"))

	if displayName == "" || email == "" {
		c.Redirect(http.StatusFound, "/admin/center/teachers?error=missing_fields")
		return
	}

	// Auto-generate username from email prefix
	username := strings.Split(email, "@")[0]
	username = strings.ToLower(strings.ReplaceAll(username, " ", ""))

	password := generateCenterPassword(10)
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/center/teachers?error=password")
		return
	}

	_, err = h.Store.CreateTeacherInCenter(ctx, center.ID, username, string(hashed), password, email, phone, displayName)
	if err != nil {
		// Username conflict — append number
		username = fmt.Sprintf("%s%d", username, time.Now().Unix()%1000)
		_, err = h.Store.CreateTeacherInCenter(ctx, center.ID, username, string(hashed), password, email, phone, displayName)
		if err != nil {
			c.Redirect(http.StatusFound, "/admin/center/teachers?error=username_taken")
			return
		}
	}

	c.Redirect(http.StatusFound, "/admin/center/teachers?created="+username)
}

func (h *Handler) CenterToggleTeacher(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	teacherID, _ := strconv.Atoi(c.Param("id"))
	ctx := c.Request.Context()

	// Verify teacher belongs to same center
	teacher, err := h.Store.GetAdminByID(ctx, teacherID)
	if err != nil || teacher.CenterID == nil || *teacher.CenterID != *admin.CenterID {
		c.Redirect(http.StatusFound, "/admin/center/teachers?error=not_found")
		return
	}
	// Don't let owner deactivate themselves
	if teacherID == admin.ID {
		c.Redirect(http.StatusFound, "/admin/center/teachers?error=self")
		return
	}

	if teacher.Active {
		h.Store.DeactivateTeacher(ctx, teacherID)
	} else {
		h.Store.ActivateTeacher(ctx, teacherID)
	}
	c.Redirect(http.StatusFound, "/admin/center/teachers")
}

// ─── Center Settings ────────────────────────────────────

func (h *Handler) CenterSettings(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	center, err := h.Store.GetCenter(c.Request.Context(), *admin.CenterID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin")
		return
	}
	h.render(c, "center_settings.html", gin.H{
		"Center": center,
	})
}

func (h *Handler) CenterResetTeacherPassword(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	ctx := c.Request.Context()
	teacherID, _ := strconv.Atoi(c.Param("id"))

	// Verify teacher belongs to same center and is a teacher
	teacher, err := h.Store.GetAdminByID(ctx, teacherID)
	if err != nil || teacher.CenterID == nil || *teacher.CenterID != *admin.CenterID || teacher.Role != "teacher" {
		c.Redirect(http.StatusFound, "/admin/center/teachers?error=not_found")
		return
	}

	password := generateCenterPassword(10)
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/center/teachers?error=password")
		return
	}

	_, err = h.Store.DB.Exec(ctx,
		`UPDATE admin SET password=$1, pending_password=$2 WHERE id=$3`,
		string(hashed), password, teacherID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/center/teachers?error=reset_failed")
		return
	}
	c.Redirect(http.StatusFound, "/admin/center/teachers?reset="+teacher.Username)
}

func (h *Handler) CenterSettingsSave(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	name := strings.TrimSpace(c.PostForm("name"))
	address := strings.TrimSpace(c.PostForm("address"))
	city := strings.TrimSpace(c.PostForm("city"))
	phone := strings.TrimSpace(c.PostForm("phone"))
	email := strings.TrimSpace(c.PostForm("email"))

	if name == "" {
		c.Redirect(http.StatusFound, "/admin/center/settings?error=name_required")
		return
	}

	err := h.Store.UpdateCenter(c.Request.Context(), *admin.CenterID, name, address, city, phone, email)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/center/settings?error=save_failed")
		return
	}
	c.Redirect(http.StatusFound, "/admin/center/settings?saved=1")
}

func generateCenterPassword(length int) string {
	const chars = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[n.Int64()]
	}
	return string(result)
}

// ─── Center Billing ─────────────────────────────────────

func (h *Handler) CenterBilling(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	ctx := c.Request.Context()
	center, err := h.Store.GetCenter(ctx, *admin.CenterID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin")
		return
	}
	invoices, _ := h.Store.ListCenterInvoices(ctx, center.ID)
	teachers, _ := h.Store.ListCenterTeachers(ctx, center.ID)

	activeCount := 0
	for _, t := range teachers {
		if t.Active {
			activeCount++
		}
	}
	now := time.Now()
	currentPeriod := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextInvoice := currentPeriod.AddDate(0, 1, 0)
	estimate := float64(activeCount) * center.PricePerTeacher

	unpaidTotal := 0.0
	unpaidCount := 0
	for _, inv := range invoices {
		if inv.Status == "unpaid" {
			unpaidTotal += inv.TotalAmount
			unpaidCount++
		}
	}

	h.render(c, "center_billing.html", gin.H{
		"Center":             center,
		"Invoices":           invoices,
		"ActiveTeacherCount": activeCount,
		"CurrentPeriod":      currentPeriod,
		"NextInvoiceDate":    nextInvoice,
		"Estimate":           estimate,
		"UnpaidTotal":        unpaidTotal,
		"UnpaidCount":        unpaidCount,
	})
}

// ─── Center Pending Join Requests (read-only aggregate) ─

func (h *Handler) CenterPendingRequests(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	ctx := c.Request.Context()
	center, err := h.Store.GetCenter(ctx, *admin.CenterID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin")
		return
	}
	requests, _ := h.Store.ListCenterPendingJoinRequests(ctx, center.ID)
	h.render(c, "center_requests.html", gin.H{
		"Center":   center,
		"Requests": requests,
	})
}

// ─── Center Students ────────────────────────────────────

func (h *Handler) CenterStudents(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	ctx := c.Request.Context()
	center, err := h.Store.GetCenter(ctx, *admin.CenterID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin")
		return
	}
	students, _ := h.Store.ListCenterStudents(ctx, center.ID)
	classrooms, _ := h.Store.ListCenterClassrooms(ctx, center.ID)
	studentCount := len(students)

	h.render(c, "center_students.html", gin.H{
		"Center":       center,
		"Students":     students,
		"Classrooms":   classrooms,
		"StudentCount": studentCount,
	})
}

func (h *Handler) CenterCreateStudent(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	ctx := c.Request.Context()

	name := strings.TrimSpace(c.PostForm("name"))
	email := strings.TrimSpace(c.PostForm("email"))
	phone := strings.TrimSpace(c.PostForm("phone"))

	if name == "" {
		c.Redirect(http.StatusFound, "/admin/center/students?error=missing_name")
		return
	}

	_, err := h.Store.CreateCenterStudent(ctx, *admin.CenterID, name, email, phone)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/center/students?error=create_failed")
		return
	}
	c.Redirect(http.StatusFound, "/admin/center/students?created=1")
}

func (h *Handler) CenterAssignStudent(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	ctx := c.Request.Context()

	studentID, _ := strconv.Atoi(c.PostForm("student_id"))
	classroomID, _ := strconv.Atoi(c.PostForm("classroom_id"))

	if studentID == 0 || classroomID == 0 {
		c.Redirect(http.StatusFound, "/admin/center/students?error=invalid")
		return
	}

	// Verify student belongs to this center
	student, err := h.Store.GetStudent(ctx, studentID)
	if err != nil || student.CenterID == nil || *student.CenterID != *admin.CenterID {
		c.Redirect(http.StatusFound, "/admin/center/students?error=not_found")
		return
	}

	// Verify classroom belongs to this center (teacher's center_id matches)
	classrooms, _ := h.Store.ListCenterClassrooms(ctx, *admin.CenterID)
	found := false
	for _, cl := range classrooms {
		if cl.ID == classroomID {
			found = true
			break
		}
	}
	if !found {
		c.Redirect(http.StatusFound, "/admin/center/students?error=classroom_not_found")
		return
	}

	err = h.Store.AssignStudentToClassroom(ctx, studentID, classroomID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/center/students?error=assign_failed")
		return
	}
	c.Redirect(http.StatusFound, "/admin/center/students?assigned=1")
}

func (h *Handler) CenterRemoveStudentFromClassroom(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	ctx := c.Request.Context()

	studentID, _ := strconv.Atoi(c.PostForm("student_id"))
	classroomID, _ := strconv.Atoi(c.PostForm("classroom_id"))

	if studentID == 0 || classroomID == 0 {
		c.Redirect(http.StatusFound, "/admin/center/students?error=invalid")
		return
	}

	// Verify student belongs to this center
	student, err := h.Store.GetStudent(ctx, studentID)
	if err != nil || student.CenterID == nil || *student.CenterID != *admin.CenterID {
		c.Redirect(http.StatusFound, "/admin/center/students?error=not_found")
		return
	}

	h.Store.RemoveStudentFromClassroom(ctx, studentID, classroomID)
	c.Redirect(http.StatusFound, "/admin/center/students?removed=1")
}

// ─── Center Classrooms ──────────────────────────────────

func (h *Handler) CenterClassrooms(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	ctx := c.Request.Context()
	center, err := h.Store.GetCenter(ctx, *admin.CenterID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin")
		return
	}
	classrooms, _ := h.Store.ListCenterClassrooms(ctx, center.ID)
	teachers, _ := h.Store.ListCenterTeachers(ctx, center.ID)

	country := center.Country
	if country == "" {
		country = "FR"
	}

	h.render(c, "center_classrooms.html", gin.H{
		"Center":     center,
		"Classrooms": classrooms,
		"Teachers":   teachers,
		"Subjects":   geo.AllSubjects,
		"Levels":     geo.LevelsForCountry(country),
	})
}

func (h *Handler) CenterCreateClassroom(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	ctx := c.Request.Context()

	center, _ := h.Store.GetCenter(ctx, *admin.CenterID)
	if center == nil {
		c.Redirect(http.StatusFound, "/admin")
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	subject := strings.TrimSpace(c.PostForm("subject"))
	level := strings.TrimSpace(c.PostForm("level"))
	teacherIDStr := c.PostForm("teacher_id")
	teacherID, _ := strconv.Atoi(teacherIDStr)

	if name == "" || teacherID == 0 {
		c.Redirect(http.StatusFound, "/admin/center/classrooms?error=missing_fields")
		return
	}

	// Verify teacher belongs to this center
	teacher, err := h.Store.GetAdminByID(ctx, teacherID)
	if err != nil || teacher.CenterID == nil || *teacher.CenterID != *admin.CenterID {
		c.Redirect(http.StatusFound, "/admin/center/classrooms?error=invalid_teacher")
		return
	}

	_, err = h.Store.CreateClassroom(ctx, name, subject, level, teacherID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/center/classrooms?error=create_failed")
		return
	}
	c.Redirect(http.StatusFound, "/admin/center/classrooms?created=1")
}

func (h *Handler) CenterDeleteClassroom(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	ctx := c.Request.Context()
	classroomID, _ := strconv.Atoi(c.Param("id"))

	// Verify classroom belongs to this center
	classrooms, _ := h.Store.ListCenterClassrooms(ctx, *admin.CenterID)
	var teacherID int
	for _, cl := range classrooms {
		if cl.ID == classroomID {
			teacherID = cl.AdminID
			break
		}
	}
	if teacherID == 0 {
		c.Redirect(http.StatusFound, "/admin/center/classrooms?error=not_found")
		return
	}

	h.Store.DeleteClassroom(ctx, classroomID, teacherID)
	c.Redirect(http.StatusFound, "/admin/center/classrooms?deleted=1")
}
