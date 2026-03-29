package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"teachhub/middleware"

	"github.com/gin-gonic/gin"
)

// ─── Student Dashboard (My Progress) ────────────────────

func (h *Handler) StudentDashboard(c *gin.Context) {
	student := middleware.GetStudent(c)
	classID, _ := strconv.Atoi(c.Param("id"))

	in, _ := h.Store.IsStudentInClassroom(c.Request.Context(), student.ID, classID)
	if !in {
		c.Redirect(http.StatusFound, "/")
		return
	}

	classroom, _ := h.Store.GetClassroom(c.Request.Context(), classID)
	liveSession, _ := h.Store.GetActiveLiveSession(c.Request.Context(), classID)

	// Reuse existing analytics functions from Phase 2
	quizDetails, _ := h.Store.GetStudentQuizDetails(c.Request.Context(), student.ID, classID)
	assignDetails, _ := h.Store.GetStudentAssignmentDetails(c.Request.Context(), student.ID, classID)

	// Attendance record
	attendanceRecords, attended, totalSessions, _ := h.Store.GetStudentAttendanceRecord(c.Request.Context(), student.ID, classID)

	// Class averages for above/below comparison
	classStats, _ := h.Store.GetStudentDashboardStats(c.Request.Context(), classID)

	// Teacher remarks
	remarks, _ := h.Store.GetStudentRemarks(c.Request.Context(), student.ID, classID)

	// Compute student's own averages
	var avgQuizPct, avgAssignPct float64
	if len(quizDetails) > 0 {
		total := 0.0
		for _, q := range quizDetails {
			total += q.Pct
		}
		avgQuizPct = total / float64(len(quizDetails))
	}
	if len(assignDetails) > 0 {
		gradedCount := 0
		total := 0.0
		for _, a := range assignDetails {
			if a.Grade != nil {
				total += a.Pct
				gradedCount++
			}
		}
		if gradedCount > 0 {
			avgAssignPct = total / float64(gradedCount)
		}
	}

	// Attendance percentage
	var attendancePct float64
	if totalSessions > 0 {
		attendancePct = float64(attended) * 100.0 / float64(totalSessions)
	}

	sub := c.DefaultQuery("sub", "overview")

	h.render(c, "student_dashboard.html", gin.H{
		"Classroom":         classroom,
		"Student":           student,
		"Tab":               "dashboard",
		"Sub":               sub,
		"LiveSession":       liveSession,
		"QuizDetails":       quizDetails,
		"AssignDetails":     assignDetails,
		"AttendanceRecords": attendanceRecords,
		"AttendedCount":     attended,
		"TotalSessions":     totalSessions,
		"AttendancePct":     attendancePct,
		"AvgQuizPct":        avgQuizPct,
		"AvgAssignPct":      avgAssignPct,
		"ClassAvgQuizPct":   classStats.ClassAvgQuizPct,
		"ClassAvgAssignPct": classStats.ClassAvgAssignPct,
		"Remarks":           remarks,
	})
}

// ─── Admin: Add Remark to Student ───────────────────────

func (h *Handler) AdminAddRemark(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	studentID, _ := strconv.Atoi(c.Param("studentId"))
	content := strings.TrimSpace(c.PostForm("content"))

	if content != "" {
		h.Store.AddStudentRemark(c.Request.Context(), classID, studentID, content)
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/student/%d", classID, studentID))
}

// ─── Admin: Delete Remark ───────────────────────────────

func (h *Handler) AdminDeleteRemark(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	studentID, _ := strconv.Atoi(c.Param("studentId"))
	remarkID, _ := strconv.Atoi(c.Param("remarkId"))

	h.Store.DeleteStudentRemark(c.Request.Context(), remarkID, classID)

	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/student/%d", classID, studentID))
}
