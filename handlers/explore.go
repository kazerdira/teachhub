package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"teachhub/geo"

	"github.com/gin-gonic/gin"
)

// ═══════════════════════════════════════════════════════════════
// Explore — Public teacher directory
// ═══════════════════════════════════════════════════════════════

func (h *Handler) ExplorePage(c *gin.Context) {
	// Country from cookie first, then IP geolocation — never from URL
	country, _ := c.Cookie("country")
	if country == "" {
		country = geo.CountryFromIP(c.ClientIP())
	}
	if country == "" {
		country = "DZ" // default
	}
	// Save to cookie for future visits
	c.SetCookie("country", country, 60*60*24*365, "/", "", false, false)

	region := c.Query("region")
	subject := c.Query("subject")
	level := c.Query("level")

	teachers, _ := h.Store.ListPublicTeachers(c.Request.Context(), country, region, subject, level)

	// Build reference data for filters
	lang, _ := c.Cookie("lang")
	if lang != "en" && lang != "fr" {
		lang = "en"
	}

	h.render(c, "explore.html", gin.H{
		"Teachers":   teachers,
		"Country":    country,
		"Region":     region,
		"Subject":    subject,
		"Level":      level,
		"Subjects":   geo.AllSubjects,
		"SubjectMap": geo.SubjectMap(),
		"Levels":     geo.LevelsForCountry(country),
		"Regions":    geo.RegionsForCountry(country),
		"RegionLabel": geo.RegionLabel(country, lang),
	})
}

func (h *Handler) TeacherPublicProfile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	teacher, err := h.Store.GetPublicTeacher(c.Request.Context(), id)
	if err != nil {
		c.Redirect(http.StatusFound, "/explore")
		return
	}
	classrooms, _ := h.Store.ListPublicClassrooms(c.Request.Context(), id)

	lang, _ := c.Cookie("lang")
	if lang != "en" && lang != "fr" {
		lang = "en"
	}

	h.render(c, "explore_teacher.html", gin.H{
		"Teacher":    teacher,
		"Classrooms": classrooms,
		"SubjectMap": geo.SubjectMap(),
		"Levels":     geo.LevelsForCountry(teacher.Country),
	})
}

// ─── Join Request (public form submission) ──────────────

func (h *Handler) JoinRequestPage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	teacher, err := h.Store.GetPublicTeacher(c.Request.Context(), id)
	if err != nil {
		c.Redirect(http.StatusFound, "/explore")
		return
	}
	classrooms, _ := h.Store.ListPublicClassrooms(c.Request.Context(), id)

	lang, _ := c.Cookie("lang")
	if lang != "en" && lang != "fr" {
		lang = "en"
	}

	h.render(c, "explore_join.html", gin.H{
		"Teacher":    teacher,
		"Classrooms": classrooms,
		"SubjectMap": geo.SubjectMap(),
		"Levels":     geo.LevelsForCountry(teacher.Country),
	})
}

func (h *Handler) JoinRequestSubmit(c *gin.Context) {
	teacherID, _ := strconv.Atoi(c.Param("id"))
	// Verify teacher exists and is public
	_, err := h.Store.GetPublicTeacher(c.Request.Context(), teacherID)
	if err != nil {
		c.Redirect(http.StatusFound, "/explore")
		return
	}

	fullName := strings.TrimSpace(c.PostForm("full_name"))
	email := strings.TrimSpace(c.PostForm("email"))
	phone := strings.TrimSpace(c.PostForm("phone"))
	level := strings.TrimSpace(c.PostForm("level"))
	message := strings.TrimSpace(c.PostForm("message"))

	if fullName == "" || email == "" {
		c.Redirect(http.StatusFound, "/explore/teacher/"+strconv.Itoa(teacherID)+"/join")
		return
	}

	var classroomID *int
	if cidStr := c.PostForm("classroom_id"); cidStr != "" {
		if cid, err := strconv.Atoi(cidStr); err == nil {
			classroomID = &cid
		}
	}

	h.Store.CreateJoinRequest(c.Request.Context(), teacherID, classroomID, fullName, email, phone, level, message)
	c.Redirect(http.StatusFound, "/explore/teacher/"+strconv.Itoa(teacherID)+"/join/success")
}

func (h *Handler) JoinRequestSuccess(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	teacher, _ := h.Store.GetPublicTeacher(c.Request.Context(), id)
	h.render(c, "explore_join_success.html", gin.H{
		"Teacher": teacher,
	})
}

// ─── Teacher: Profile Settings ─────────────────────────

func (h *Handler) AdminProfilePage(c *gin.Context) {
	aid := adminID(c)
	admin, err := h.Store.GetAdminByID(c.Request.Context(), aid)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin")
		return
	}
	lang, _ := c.Cookie("lang")
	if lang != "en" && lang != "fr" {
		lang = "en"
	}

	country := admin.Country
	if country == "" {
		country = geo.CountryFromIP(c.ClientIP())
		if country == "" {
			country = "DZ"
		}
	}

	h.render(c, "admin_profile.html", gin.H{
		"Admin":       admin,
		"Subjects":    geo.AllSubjects,
		"Levels":      geo.LevelsForCountry(country),
		"Regions":     geo.RegionsForCountry(country),
		"RegionLabel": geo.RegionLabel(country, lang),
		"Country":     country,
	})
}

func (h *Handler) AdminProfileSave(c *gin.Context) {
	aid := adminID(c)
	bio := strings.TrimSpace(c.PostForm("bio"))
	country := strings.TrimSpace(c.PostForm("country"))
	region := strings.TrimSpace(c.PostForm("region"))
	publicProfile := c.PostForm("public_profile") == "on"

	// Parse subjects (multi-select)
	subjects := c.PostFormArray("subjects")
	if len(subjects) == 0 {
		subjects = []string{}
	}
	// Parse levels (multi-select)
	levels := c.PostFormArray("levels")
	if len(levels) == 0 {
		levels = []string{}
	}

	h.Store.UpdateTeacherProfile(c.Request.Context(), aid, bio, subjects, levels, country, region, publicProfile)
	c.Redirect(http.StatusFound, "/admin/profile?saved=1")
}

// ─── Teacher: Join Request Management ──────────────────

func (h *Handler) AdminJoinRequests(c *gin.Context) {
	aid := adminID(c)
	requests, _ := h.Store.ListJoinRequestsForTeacher(c.Request.Context(), aid)
	classrooms, _ := h.Store.ListClassrooms(c.Request.Context(), aid)
	h.render(c, "admin_requests.html", gin.H{
		"Requests":   requests,
		"Classrooms": classrooms,
	})
}

func (h *Handler) AdminApproveRequest(c *gin.Context) {
	aid := adminID(c)
	reqID, _ := strconv.Atoi(c.Param("reqId"))
	classroomID, _ := strconv.Atoi(c.PostForm("classroom_id"))
	if classroomID == 0 {
		c.Redirect(http.StatusFound, "/admin/requests")
		return
	}
	h.Store.ApproveJoinRequest(c.Request.Context(), reqID, aid, classroomID)
	c.Redirect(http.StatusFound, "/admin/requests")
}

func (h *Handler) AdminRejectRequest(c *gin.Context) {
	aid := adminID(c)
	reqID, _ := strconv.Atoi(c.Param("reqId"))
	h.Store.RejectJoinRequest(c.Request.Context(), reqID, aid)
	c.Redirect(http.StatusFound, "/admin/requests")
}

// ─── API: Levels/Regions for country (AJAX) ────────────

func (h *Handler) APILevelsForCountry(c *gin.Context) {
	country := c.Query("country")
	c.JSON(200, geo.LevelsForCountry(country))
}

func (h *Handler) APIRegionsForCountry(c *gin.Context) {
	country := c.Query("country")
	lang, _ := c.Cookie("lang")
	if lang != "en" && lang != "fr" {
		lang = "en"
	}
	c.JSON(200, gin.H{
		"regions": geo.RegionsForCountry(country),
		"label":   geo.RegionLabel(country, lang),
	})
}
