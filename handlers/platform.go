package handlers

import (
	"crypto/rand"
	"encoding/csv"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"teachhub/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// helper: get lang from cookie
func platformLang(c *gin.Context) string {
	lang, _ := c.Cookie("lang")
	if lang != "fr" {
		lang = "en"
	}
	return lang
}

// platformRender renders a platform template with common data (Lang, CSRFToken).
func (h *Handler) platformRender(c *gin.Context, tmplName string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["Lang"] = platformLang(c)
	data["PlatformPath"] = h.PlatformPath
	if t, exists := c.Get("csrf_token"); exists {
		data["CSRFToken"] = t
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	h.Tmpl.ExecuteTemplate(c.Writer, tmplName, data)
}

// pp builds a platform-relative URL (e.g. "/ctrl-p-8x3kf/teachers/5")
func (h *Handler) pp(path string) string {
	return h.PlatformPath + path
}

// ─── Public: Teacher Application Form ───────────────────

func (h *Handler) ApplyPage(c *gin.Context) {
	h.platformRender(c, "apply.html", gin.H{
		"Error":   c.Query("error"),
		"Success": c.Query("success"),
	})
}

func (h *Handler) ApplySubmit(c *gin.Context) {
	fullName := strings.TrimSpace(c.PostForm("full_name"))
	email := strings.TrimSpace(c.PostForm("email"))
	phone := strings.TrimSpace(c.PostForm("phone"))
	school := strings.TrimSpace(c.PostForm("school_name"))
	country := inferCountry(c)
	centerName := strings.TrimSpace(c.PostForm("center_name"))
	message := strings.TrimSpace(c.PostForm("message"))
	expectedTeachers, _ := strconv.Atoi(c.PostForm("expected_teachers"))
	expectedStudents, _ := strconv.Atoi(c.PostForm("expected_students"))
	if expectedTeachers < 1 {
		expectedTeachers = 1
	}

	if fullName == "" || email == "" {
		c.Redirect(http.StatusFound, "/apply?error=name_email_required")
		return
	}

	err := h.Store.CreateTeacherApplication(c.Request.Context(), fullName, email, phone, school, country, message, centerName, expectedTeachers, expectedStudents)
	if err != nil {
		c.Redirect(http.StatusFound, "/apply?error=submit_failed")
		return
	}

	c.Redirect(http.StatusFound, "/apply/success")
}

func (h *Handler) ApplySuccess(c *gin.Context) {
	h.platformRender(c, "apply_success.html", nil)
}

// ─── Platform Admin: Login ──────────────────────────────

func (h *Handler) PlatformLoginPage(c *gin.Context) {
	h.platformRender(c, "platform_login.html", gin.H{
		"Error": c.Query("error"),
	})
}

func (h *Handler) PlatformLogin(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")

	admin, err := h.Store.GetPlatformAdminByUsername(c.Request.Context(), username)
	if err != nil {
		c.Redirect(http.StatusFound, h.pp("/login?error=1"))
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) != nil {
		c.Redirect(http.StatusFound, h.pp("/login?error=1"))
		return
	}

	middleware.SetPlatformSession(c, admin.ID)
	c.Redirect(http.StatusFound, h.pp(""))
}

func (h *Handler) PlatformLogout(c *gin.Context) {
	middleware.ClearPlatformSession(c)
	c.Redirect(http.StatusFound, h.pp("/login"))
}

// ─── Platform Admin: Dashboard ──────────────────────────

func (h *Handler) PlatformDashboard(c *gin.Context) {
	ctx := c.Request.Context()
	pending, approved, rejected, contacted, _ := h.Store.CountApplicationsByStatus(ctx)
	active, suspended, expired, _ := h.Store.CountActiveTeachers(ctx)
	expiringSoon, _ := h.Store.CountExpiringSoon(ctx, 7)
	totalRevenue, _ := h.Store.GetTotalRevenue(ctx)
	monthlyRevenue, _ := h.Store.GetMonthlyRevenue(ctx)

	// Auto-expire subscriptions
	h.Store.CheckAndExpireSubscriptions(ctx)

	h.platformRender(c, "platform_dashboard.html", gin.H{
		"Pending":           pending,
		"Approved":          approved,
		"Rejected":          rejected,
		"Contacted":         contacted,
		"Total":             pending + approved + rejected + contacted,
		"ActiveTeachers":    active,
		"SuspendedTeachers": suspended,
		"ExpiredTeachers":   expired,
		"TotalTeachers":     active + suspended + expired,
		"ExpiringSoon":      expiringSoon,
		"TotalRevenue":      totalRevenue,
		"MonthlyRevenue":    monthlyRevenue,
	})
}

// ─── Platform Admin: Applications List ──────────────────

func (h *Handler) PlatformApplications(c *gin.Context) {
	ctx := c.Request.Context()
	filter := c.DefaultQuery("status", "all")

	apps, err := h.Store.ListTeacherApplications(ctx, filter)
	if err != nil {
		c.String(500, "Error: %v", err)
		return
	}

	pending, approved, rejected, contacted, _ := h.Store.CountApplicationsByStatus(ctx)

	h.platformRender(c, "platform_applications.html", gin.H{
		"Applications": apps,
		"Filter":       filter,
		"Pending":      pending,
		"Approved":     approved,
		"Rejected":     rejected,
		"Contacted":    contacted,
	})
}

// ─── Platform Admin: Application Detail ─────────────────

func (h *Handler) PlatformAppDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	app, err := h.Store.GetTeacherApplication(c.Request.Context(), id)
	if err != nil {
		c.Redirect(http.StatusFound, h.pp("/applications"))
		return
	}

	h.platformRender(c, "platform_app_detail.html", gin.H{
		"App":   app,
		"Saved": c.Query("saved"),
	})
}

func (h *Handler) PlatformUpdateAppStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	status := c.PostForm("status")
	notes := c.PostForm("admin_notes")

	validStatuses := map[string]bool{"pending": true, "contacted": true, "approved": true, "rejected": true}
	if !validStatuses[status] {
		status = "pending"
	}

	ctx := c.Request.Context()

	// If approving, check if teacher account already exists for this application
	if status == "approved" {
		app, err := h.Store.GetTeacherApplication(ctx, id)
		if err != nil {
			c.Redirect(http.StatusFound, h.pp("/applications"))
			return
		}

		// Check if already approved (account already created)
		if app.Status == "approved" {
			// Just update notes, don't create another account
			h.Store.UpdateApplicationStatus(ctx, id, status, notes)
			c.Redirect(http.StatusFound, h.pp("/applications/"+strconv.Itoa(id)+"?saved=1"))
			return
		}

		// Generate username from email (before @)
		username := strings.Split(app.Email, "@")[0]
		username = strings.ToLower(strings.ReplaceAll(username, " ", ""))

		// Generate random password
		password := generatePassword(10)
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			c.Redirect(http.StatusFound, h.pp("/applications/"+strconv.Itoa(id)+"?error=password"))
			return
		}

		// 1) Create center
		centerName := app.CenterName
		if centerName == "" {
			centerName = app.SchoolName
		}
		if centerName == "" {
			centerName = app.FullName + " Center"
		}
		seats := app.ExpectedTeachers
		if seats < 1 {
			seats = 1
		}
		currency := "DZD"
		country := strings.ToUpper(strings.TrimSpace(app.Wilaya))
		if country == "" {
			country = "DZ"
		}
		// FR (and other EU) → EUR; everything else stays DZD until we add more locales
		if country == "FR" {
			currency = "EUR"
		}
		centerID, err := h.Store.CreateCenterWithCountry(ctx, centerName, app.Email, 0, currency, 0, country)
		if err != nil {
			log.Printf("[approve] CreateCenter failed (app=%d): %v", id, err)
			c.Redirect(http.StatusFound, h.pp("/applications/"+strconv.Itoa(id)+"?error=center"))
			return
		}
		// Set expected teacher count for reference
		_ = seats

		// 2) Create owner admin with center_id
		ownerID, err := h.Store.CreateOwnerAdmin(ctx, centerID, username, string(hashed), password, app.Email, app.Phone, app.SchoolName, id, app.FullName)
		if err != nil {
			// Username conflict — append number
			username = fmt.Sprintf("%s%d", username, id)
			ownerID, err = h.Store.CreateOwnerAdmin(ctx, centerID, username, string(hashed), password, app.Email, app.Phone, app.SchoolName, id, app.FullName)
			if err != nil {
				log.Printf("[approve] CreateOwnerAdmin failed (app=%d, center=%d): %v", id, centerID, err)
				c.Redirect(http.StatusFound, h.pp("/applications/"+strconv.Itoa(id)+"?error=create"))
				return
			}
		}

		// 3) Link center to owner
		h.Store.DB.Exec(ctx, `UPDATE center SET owner_admin_id=$1 WHERE id=$2`, ownerID, centerID)

		// Update application status
		h.Store.UpdateApplicationStatus(ctx, id, status, notes)

		// Redirect to credentials page (no longer passes password in URL)
		c.Redirect(http.StatusFound, h.pp("/applications/"+strconv.Itoa(id)+"/credentials"))
		return
	}

	h.Store.UpdateApplicationStatus(ctx, id, status, notes)
	c.Redirect(http.StatusFound, h.pp("/applications/"+strconv.Itoa(id)+"?saved=1"))
}

// generatePassword creates a random alphanumeric password of the given length
func generatePassword(length int) string {
	const chars = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[n.Int64()]
	}
	return string(result)
}

// ─── Platform Admin: Credentials Page ───────────────────

func (h *Handler) PlatformCredentials(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	ctx := c.Request.Context()
	app, err := h.Store.GetTeacherApplication(ctx, id)
	if err != nil {
		c.Redirect(http.StatusFound, h.pp("/applications"))
		return
	}

	// Find the teacher account created from this application
	var teacherUsername string
	var pendingPassword *string
	h.Store.DB.QueryRow(ctx,
		`SELECT username, pending_password FROM admin WHERE application_id=$1`, id).
		Scan(&teacherUsername, &pendingPassword)

	password := ""
	if pendingPassword != nil {
		password = *pendingPassword
	}

	h.platformRender(c, "platform_credentials.html", gin.H{
		"App":        app,
		"Username":   teacherUsername,
		"Password":   password,
		"HasPending": pendingPassword != nil,
	})
}

// ─── Platform Admin: Teachers List ──────────────────────

func (h *Handler) PlatformTeachers(c *gin.Context) {
	ctx := c.Request.Context()
	teachers, err := h.Store.ListTeachers(ctx)
	if err != nil {
		c.String(500, "Error: %v", err)
		return
	}

	h.platformRender(c, "platform_teachers.html", gin.H{
		"Teachers": teachers,
	})
}

// ─── Platform Admin: Teacher Detail ─────────────────────

func (h *Handler) PlatformTeacherDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	ctx := c.Request.Context()

	teacher, err := h.Store.GetAdminByID(ctx, id)
	if err != nil {
		c.Redirect(http.StatusFound, h.pp("/teachers"))
		return
	}

	classroomCount, students, quizzes, resources, _ := h.Store.GetTeacherStats(ctx, id)
	payments, _ := h.Store.ListPaymentsByTeacher(ctx, id)
	totalPaid, paymentCount, _ := h.Store.GetTeacherTotalPayments(ctx, id)

	// Fetch full classroom details for the owner panel
	classroomList, _ := h.Store.ListClassrooms(ctx, id)
	type ClassroomDetail struct {
		Classroom   interface{}
		Students    interface{}
		Resources   interface{}
		Assignments interface{}
		Quizzes     interface{}
	}
	var classroomDetails []ClassroomDetail
	for _, cl := range classroomList {
		stu, _ := h.Store.ListClassroomStudents(ctx, cl.ID)
		res, _ := h.Store.ListResources(ctx, cl.ID)
		asgn, _ := h.Store.ListAssignments(ctx, cl.ID)
		qz, _ := h.Store.ListQuizzes(ctx, cl.ID)
		classroomDetails = append(classroomDetails, ClassroomDetail{
			Classroom:   cl,
			Students:    stu,
			Resources:   res,
			Assignments: asgn,
			Quizzes:     qz,
		})
	}

	h.platformRender(c, "platform_teacher_detail.html", gin.H{
		"Teacher":          teacher,
		"Classrooms":       classroomCount,
		"Students":         students,
		"Quizzes":          quizzes,
		"Resources":        resources,
		"ClassroomDetails": classroomDetails,
		"Payments":         payments,
		"TotalPaid":        totalPaid,
		"PaymentCount":     paymentCount,
		"Saved":            c.Query("saved"),
		"HasPendingPW":     teacher.PendingPassword != nil,
	})
}

func (h *Handler) PlatformToggleTeacher(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	action := c.PostForm("action") // suspend or activate

	status := "active"
	if action == "suspend" {
		status = "suspended"
	}

	h.Store.UpdateTeacherSubscription(c.Request.Context(), id, status)
	c.Redirect(http.StatusFound, h.pp("/teachers/"+strconv.Itoa(id)+"?saved=1"))
}

func (h *Handler) PlatformResetPassword(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	ctx := c.Request.Context()

	teacher, err := h.Store.GetAdminByID(ctx, id)
	if err != nil {
		c.Redirect(http.StatusFound, h.pp("/teachers"))
		return
	}

	password := generatePassword(10)
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	h.Store.DB.Exec(ctx, `UPDATE admin SET password=$1, pending_password=$2 WHERE id=$3`, string(hashed), password, id)

	h.platformRender(c, "platform_credentials.html", gin.H{
		"App":        nil,
		"Teacher":    teacher,
		"Username":   teacher.Username,
		"Password":   password,
		"IsReset":    true,
		"HasPending": true,
	})
}

// ─── Platform Admin: Subscription Management ────────────

// PlatformTeacherCredentials shows credentials for a teacher by teacher ID (not application ID)
func (h *Handler) PlatformTeacherCredentials(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	ctx := c.Request.Context()

	teacher, err := h.Store.GetAdminByID(ctx, id)
	if err != nil {
		c.Redirect(http.StatusFound, h.pp("/teachers"))
		return
	}

	password := ""
	if teacher.PendingPassword != nil {
		password = *teacher.PendingPassword
	}

	h.platformRender(c, "platform_credentials.html", gin.H{
		"App":        nil,
		"Teacher":    teacher,
		"Username":   teacher.Username,
		"Password":   password,
		"HasPending": teacher.PendingPassword != nil,
	})
}

func (h *Handler) PlatformExtendSubscription(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	monthsStr := c.PostForm("months")
	endDateStr := c.PostForm("end_date")
	ctx := c.Request.Context()

	if monthsStr != "" {
		months, _ := strconv.Atoi(monthsStr)
		if months > 0 && months <= 24 {
			h.Store.ExtendSubscription(ctx, id, months)
		}
	} else if endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			h.Store.SetSubscriptionEnd(ctx, id, endDate)
			if endDate.After(time.Now()) {
				h.Store.UpdateTeacherSubscription(ctx, id, "active")
			}
		}
	}

	c.Redirect(http.StatusFound, h.pp("/teachers/"+strconv.Itoa(id)+"?saved=1"))
}

// ─── Platform Admin: Payment Recording ──────────────────

func (h *Handler) PlatformRecordPayment(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	amountStr := c.PostForm("amount")
	method := c.PostForm("method")
	reference := strings.TrimSpace(c.PostForm("reference"))
	notes := strings.TrimSpace(c.PostForm("notes"))

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		c.Redirect(http.StatusFound, h.pp("/teachers/"+strconv.Itoa(id)+"?error=amount"))
		return
	}

	validMethods := map[string]bool{"cash": true, "ccp": true, "baridi_mob": true, "other": true}
	if !validMethods[method] {
		method = "cash"
	}

	h.Store.CreatePayment(c.Request.Context(), id, amount, method, reference, notes)
	c.Redirect(http.StatusFound, h.pp("/teachers/"+strconv.Itoa(id)+"?saved=1"))
}

func (h *Handler) PlatformDeletePayment(c *gin.Context) {
	teacherID, _ := strconv.Atoi(c.Param("id"))
	paymentID, _ := strconv.Atoi(c.Param("paymentId"))

	h.Store.DeletePayment(c.Request.Context(), paymentID)
	c.Redirect(http.StatusFound, h.pp("/teachers/"+strconv.Itoa(teacherID)+"?saved=1"))
}

// ─── Platform Admin: Analytics ──────────────────────────

func (h *Handler) PlatformAnalytics(c *gin.Context) {
	ctx := c.Request.Context()

	active, suspended, expired, _ := h.Store.CountActiveTeachers(ctx)
	totalStudents, _ := h.Store.TotalStudentsOnPlatform(ctx)
	totalClassrooms, _ := h.Store.TotalClassroomsOnPlatform(ctx)
	totalQuizzes, _ := h.Store.TotalQuizzesOnPlatform(ctx)
	totalRevenue, _ := h.Store.GetTotalRevenue(ctx)
	monthlyRevenue, _ := h.Store.GetMonthlyRevenue(ctx)
	topTeachers, _ := h.Store.TopTeachersByStudents(ctx, 10)
	revenueBreakdown, _ := h.Store.MonthlyRevenueBreakdown(ctx, 6)
	appTrend, _ := h.Store.ApplicationsTrend(ctx, 6)

	h.platformRender(c, "platform_analytics.html", gin.H{
		"ActiveTeachers":    active,
		"SuspendedTeachers": suspended,
		"ExpiredTeachers":   expired,
		"TotalTeachers":     active + suspended + expired,
		"TotalStudents":     totalStudents,
		"TotalClassrooms":   totalClassrooms,
		"TotalQuizzes":      totalQuizzes,
		"TotalRevenue":      totalRevenue,
		"MonthlyRevenue":    monthlyRevenue,
		"TopTeachers":       topTeachers,
		"RevenueBreakdown":  revenueBreakdown,
		"AppTrend":          appTrend,
	})
}

// ─── Platform Admin: CSV Exports ────────────────────────

func (h *Handler) PlatformExportTeachersCSV(c *gin.Context) {
	ctx := c.Request.Context()
	teachers, err := h.Store.ListTeachers(ctx)
	if err != nil {
		c.String(500, "Error: %v", err)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=teachers.csv")
	c.Writer.Write([]byte("\xEF\xBB\xBF")) // UTF-8 BOM

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"ID", "Username", "Email", "School", "Status", "Start Date", "End Date", "Classrooms", "Students"})

	for _, t := range teachers {
		start := ""
		if t.SubscriptionStart != nil {
			start = t.SubscriptionStart.Format("2006-01-02")
		}
		end := ""
		if t.SubscriptionEnd != nil {
			end = t.SubscriptionEnd.Format("2006-01-02")
		}
		w.Write([]string{
			strconv.Itoa(t.ID),
			t.Username,
			t.Email,
			t.SchoolName,
			t.SubscriptionStatus,
			start,
			end,
			strconv.Itoa(t.ClassroomCount),
			strconv.Itoa(t.StudentCount),
		})
	}
	w.Flush()
}

func (h *Handler) PlatformExportPaymentsCSV(c *gin.Context) {
	ctx := c.Request.Context()
	payments, err := h.Store.ListAllPayments(ctx)
	if err != nil {
		c.String(500, "Error: %v", err)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=payments.csv")
	c.Writer.Write([]byte("\xEF\xBB\xBF")) // UTF-8 BOM

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"ID", "Teacher ID", "Amount (DA)", "Method", "Reference", "Notes", "Date"})

	for _, p := range payments {
		w.Write([]string{
			strconv.Itoa(p.ID),
			strconv.Itoa(p.TeacherID),
			fmt.Sprintf("%.2f", p.Amount),
			p.Method,
			p.Reference,
			p.Notes,
			p.RecordedAt.Format("2006-01-02 15:04"),
		})
	}
	w.Flush()
}

// ─── Platform Admin: Password Change ────────────────────

func (h *Handler) PlatformChangePasswordPage(c *gin.Context) {
	h.platformRender(c, "platform_password.html", gin.H{
		"Success": c.Query("success"),
		"Error":   c.Query("error"),
	})
}

func (h *Handler) PlatformChangePassword(c *gin.Context) {
	currentPw := c.PostForm("current_password")
	newPw := c.PostForm("new_password")
	confirmPw := c.PostForm("confirm_password")

	if newPw == "" || len(newPw) < 6 {
		c.Redirect(http.StatusFound, h.pp("/password?error=short"))
		return
	}
	if newPw != confirmPw {
		c.Redirect(http.StatusFound, h.pp("/password?error=mismatch"))
		return
	}

	adminID := middleware.GetPlatformAdminID(c)
	admin, err := h.Store.GetPlatformAdminByID(c.Request.Context(), adminID)
	if err != nil {
		c.Redirect(http.StatusFound, h.pp("/password?error=unknown"))
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(currentPw)) != nil {
		c.Redirect(http.StatusFound, h.pp("/password?error=wrong"))
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(newPw), bcrypt.DefaultCost)
	h.Store.UpdatePlatformAdminPassword(c.Request.Context(), adminID, string(hashed))
	c.Redirect(http.StatusFound, h.pp("/password?success=1"))
}

// ─── Platform Admin: Center Management ──────────────────

func (h *Handler) PlatformCenters(c *gin.Context) {
	ctx := c.Request.Context()
	centers, _ := h.Store.ListCenters(ctx)
	h.platformRender(c, "platform_centers.html", gin.H{
		"Centers": centers,
	})
}

func (h *Handler) PlatformCenterDetail(c *gin.Context) {
	ctx := c.Request.Context()
	centerID, _ := strconv.Atoi(c.Param("id"))
	center, err := h.Store.GetCenter(ctx, centerID)
	if err != nil {
		c.Redirect(http.StatusFound, h.pp("/centers"))
		return
	}
	teachers, _ := h.Store.ListCenterTeachers(ctx, centerID)
	stats, _ := h.Store.GetCenterStats(ctx, centerID)
	teacherCount, _ := h.Store.CountCenterTeachers(ctx, centerID)
	invoices, _ := h.Store.ListCenterInvoices(ctx, centerID)

	h.platformRender(c, "platform_center_detail.html", gin.H{
		"Center":       center,
		"Teachers":     teachers,
		"Stats":        stats,
		"TeacherCount": teacherCount,
		"Invoices":     invoices,
		"Saved":        c.Query("saved"),
		"Error":        c.Query("error"),
	})
}

func (h *Handler) PlatformCenterUpdatePricing(c *gin.Context) {
	ctx := c.Request.Context()
	centerID, _ := strconv.Atoi(c.Param("id"))
	price, _ := strconv.ParseFloat(c.PostForm("price_per_teacher"), 64)
	currency := strings.TrimSpace(c.PostForm("currency"))
	if price < 0 {
		price = 0
	}
	if currency == "" {
		currency = "DZD"
	}
	h.Store.UpdateCenterPricing(ctx, centerID, price, currency)
	c.Redirect(http.StatusFound, h.pp("/centers/"+strconv.Itoa(centerID)+"?saved=pricing"))
}

func (h *Handler) PlatformGenerateCenterInvoice(c *gin.Context) {
	ctx := c.Request.Context()
	centerID, _ := strconv.Atoi(c.Param("id"))
	monthStr := c.PostForm("month")
	period, err := time.Parse("2006-01", monthStr)
	if err != nil {
		period = time.Now()
	}
	_, err = h.Store.GenerateCenterMonthlyInvoice(ctx, centerID, period)
	if err != nil {
		c.Redirect(http.StatusFound, h.pp("/centers/"+strconv.Itoa(centerID)+"?error=invoice_failed"))
		return
	}
	c.Redirect(http.StatusFound, h.pp("/centers/"+strconv.Itoa(centerID)+"?saved=invoice"))
}

func (h *Handler) PlatformMarkCenterInvoicePaid(c *gin.Context) {
	ctx := c.Request.Context()
	centerID, _ := strconv.Atoi(c.Param("id"))
	invoiceID, _ := strconv.Atoi(c.Param("invoiceId"))
	method := strings.TrimSpace(c.PostForm("method"))
	reference := strings.TrimSpace(c.PostForm("reference"))
	h.Store.MarkCenterInvoicePaid(ctx, invoiceID, centerID, method, reference)
	c.Redirect(http.StatusFound, h.pp("/centers/"+strconv.Itoa(centerID)+"?saved=invoice"))
}

func (h *Handler) PlatformCancelCenterInvoice(c *gin.Context) {
	ctx := c.Request.Context()
	centerID, _ := strconv.Atoi(c.Param("id"))
	invoiceID, _ := strconv.Atoi(c.Param("invoiceId"))
	h.Store.CancelCenterInvoice(ctx, invoiceID, centerID)
	c.Redirect(http.StatusFound, h.pp("/centers/"+strconv.Itoa(centerID)+"?saved=invoice"))
}

func (h *Handler) PlatformCenterUpdateSubscription(c *gin.Context) {
	ctx := c.Request.Context()
	centerID, _ := strconv.Atoi(c.Param("id"))
	status := c.PostForm("status")
	validStatuses := map[string]bool{"trial": true, "active": true, "expired": true, "suspended": true, "cancelled": true}
	if !validStatuses[status] {
		c.Redirect(http.StatusFound, h.pp("/centers/"+strconv.Itoa(centerID)+"?error=invalid_status"))
		return
	}
	var start, end *time.Time
	if s := c.PostForm("subscription_start"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			start = &t
		}
	}
	if e := c.PostForm("subscription_end"); e != "" {
		if t, err := time.Parse("2006-01-02", e); err == nil {
			end = &t
		}
	}
	h.Store.UpdateCenterSubscription(ctx, centerID, status, start, end)
	c.Redirect(http.StatusFound, h.pp("/centers/"+strconv.Itoa(centerID)+"?saved=subscription"))
}
