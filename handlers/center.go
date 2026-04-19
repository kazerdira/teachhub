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

	h.render(c, "center_dashboard.html", gin.H{
		"Center":   center,
		"Stats":    stats,
		"Teachers": teachers,
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

	activeCount, _ := h.Store.CountCenterTeachers(ctx, *admin.CenterID)
	center, _ := h.Store.GetCenter(ctx, *admin.CenterID)
	if center == nil {
		c.Redirect(http.StatusFound, "/admin")
		return
	}
	if activeCount >= center.SeatCount {
		c.Redirect(http.StatusFound, "/admin/center/teachers?error=seat_limit")
		return
	}

	username := strings.TrimSpace(c.PostForm("username"))
	email := strings.TrimSpace(c.PostForm("email"))
	phone := strings.TrimSpace(c.PostForm("phone"))

	if username == "" || email == "" {
		c.Redirect(http.StatusFound, "/admin/center/teachers?error=missing_fields")
		return
	}

	password := generateCenterPassword(10)
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/center/teachers?error=password")
		return
	}

	_, err = h.Store.CreateTeacherInCenter(ctx, center.ID, username, string(hashed), password, email, phone)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/center/teachers?error=username_taken")
		return
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
		// Check seat limit before reactivating
		activeCount, _ := h.Store.CountCenterTeachers(ctx, *admin.CenterID)
		center, _ := h.Store.GetCenter(ctx, *admin.CenterID)
		if center != nil && activeCount >= center.SeatCount {
			c.Redirect(http.StatusFound, "/admin/center/teachers?error=seat_limit")
			return
		}
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
