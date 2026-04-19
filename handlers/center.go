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
	stats, _ := h.Store.GetCenterStats(c.Request.Context(), center.ID)
	if stats != nil {
		stats.SeatCount = center.SeatCount
	}
	teachers, _ := h.Store.ListCenterTeachers(c.Request.Context(), center.ID)
	dashStats, _ := h.Store.GetCenterDashboardStats(c.Request.Context(), center.ID)
	performance, _ := h.Store.GetCenterTeacherPerformance(c.Request.Context(), center.ID)

	h.render(c, "center_dashboard.html", gin.H{
		"Center":      center,
		"Stats":       stats,
		"Teachers":    teachers,
		"DashStats":   dashStats,
		"Performance": performance,
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
	activeCount, _ := h.Store.CountCenterTeachers(c.Request.Context(), center.ID)

	h.render(c, "center_teachers.html", gin.H{
		"Center":      center,
		"Teachers":    teachers,
		"ActiveCount": activeCount,
		"SeatCount":   center.SeatCount,
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

	c.Redirect(http.StatusFound, "/admin/center/teachers?created="+username+"&pw="+password)
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
		// Soft limit — allow reactivation regardless of seat count
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

	// Parse month param, default to current month
	monthStr := c.Query("month")
	var period time.Time
	if monthStr != "" {
		period, err = time.Parse("2006-01", monthStr)
		if err != nil {
			period = time.Now()
		}
	} else {
		period = time.Now()
	}
	period = time.Date(period.Year(), period.Month(), 1, 0, 0, 0, 0, time.UTC)

	statusFilter := c.DefaultQuery("status", "all")
	invoices, _ := h.Store.ListCenterInvoices(ctx, center.ID, period, statusFilter)
	totalAmount, paidAmount, unpaidCount, _ := h.Store.GetCenterBillingSummary(ctx, center.ID, period)
	parentViews := h.Store.GetParentViewsWeek(ctx, center.ID)

	h.render(c, "center_billing.html", gin.H{
		"Center":       center,
		"Invoices":     invoices,
		"Period":       period,
		"PeriodStr":    period.Format("2006-01"),
		"StatusFilter": statusFilter,
		"TotalAmount":  totalAmount,
		"PaidAmount":   paidAmount,
		"UnpaidCount":  unpaidCount,
		"ParentViews":  parentViews,
	})
}

func (h *Handler) CenterGenerateInvoices(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	ctx := c.Request.Context()

	monthStr := c.PostForm("month")
	period, err := time.Parse("2006-01", monthStr)
	if err != nil {
		period = time.Now()
	}

	count, err := h.Store.GenerateMonthlyInvoices(ctx, *admin.CenterID, period)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/center/billing?month="+period.Format("2006-01")+"&error=generate")
		return
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/center/billing?month=%s&generated=%d", period.Format("2006-01"), count))
}

func (h *Handler) CenterMarkInvoicePaid(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	invoiceID, _ := strconv.Atoi(c.Param("invoiceId"))
	method := c.DefaultPostForm("method", "cash")
	month := c.DefaultPostForm("month", time.Now().Format("2006-01"))

	h.Store.MarkInvoicePaid(c.Request.Context(), invoiceID, *admin.CenterID, method)
	c.Redirect(http.StatusFound, "/admin/center/billing?month="+month)
}

func (h *Handler) CenterCancelInvoice(c *gin.Context) {
	admin := c.MustGet("admin").(*store.Admin)
	invoiceID, _ := strconv.Atoi(c.Param("invoiceId"))
	month := c.DefaultPostForm("month", time.Now().Format("2006-01"))

	h.Store.CancelInvoice(c.Request.Context(), invoiceID, *admin.CenterID)
	c.Redirect(http.StatusFound, "/admin/center/billing?month="+month)
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
