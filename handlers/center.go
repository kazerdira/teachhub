package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strconv"
	"strings"

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
