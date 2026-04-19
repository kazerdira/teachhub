package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ParentReport renders the public parent progress page.
// No authentication — the parent_code in the URL IS the access token.
func (h *Handler) ParentReport(c *gin.Context) {
	code := c.Param("code")
	if len(code) < 8 {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	ctx := c.Request.Context()

	// Look up student + classroom from the secret code
	data, err := h.Store.GetStudentByParentCode(ctx, code)
	if err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	// Log this view for center analytics
	h.Store.LogParentView(ctx, code, c.ClientIP())

	// Fetch all data using existing store functions
	quizDetails, _ := h.Store.GetStudentQuizDetails(ctx, data.StudentID, data.ClassroomID)
	assignDetails, _ := h.Store.GetStudentAssignmentDetails(ctx, data.StudentID, data.ClassroomID)
	remarks, _ := h.Store.GetStudentRemarks(ctx, data.StudentID, data.ClassroomID)
	attendance, attended, totalSessions, _ := h.Store.GetStudentAttendanceRecord(ctx, data.StudentID, data.ClassroomID)
	classStats, _ := h.Store.GetStudentDashboardStats(ctx, data.ClassroomID)

	// Compute student averages
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

	// Overall student average (quiz + assignments combined)
	overallParts := 0
	overallTotal := 0.0
	if len(quizDetails) > 0 {
		overallTotal += avgQuizPct
		overallParts++
	}
	if avgAssignPct > 0 {
		overallTotal += avgAssignPct
		overallParts++
	}
	overallAvg := 0.0
	if overallParts > 0 {
		overallAvg = overallTotal / float64(overallParts)
	}

	// Class average for comparison
	classOverallParts := 0
	classOverallTotal := 0.0
	if classStats != nil {
		if classStats.ClassAvgQuizPct > 0 {
			classOverallTotal += classStats.ClassAvgQuizPct
			classOverallParts++
		}
		if classStats.ClassAvgAssignPct > 0 {
			classOverallTotal += classStats.ClassAvgAssignPct
			classOverallParts++
		}
	}
	classOverallAvg := 0.0
	if classOverallParts > 0 {
		classOverallAvg = classOverallTotal / float64(classOverallParts)
	}

	// Banner color: green = above class avg, amber = within 10%, red = below
	banner := "green"
	if overallParts == 0 {
		banner = "gray" // no data yet
	} else if overallAvg < classOverallAvg-10 {
		banner = "red"
	} else if overallAvg < classOverallAvg {
		banner = "amber"
	}

	// Limit remarks to latest 5
	if len(remarks) > 5 {
		remarks = remarks[:5]
	}

	// Limit attendance to last 10 sessions for display
	displayAttendance := attendance
	if len(displayAttendance) > 10 {
		displayAttendance = displayAttendance[:10]
	}

	// Compute quiz trend: compare first half vs second half average (chronological)
	// quizDetails are ordered DESC (newest first), so reverse for trend calc
	trend := "stable"
	if len(quizDetails) >= 3 {
		mid := len(quizDetails) / 2
		// older half = quizDetails[mid:] (these were taken earlier)
		// newer half = quizDetails[:mid]
		olderSum, newerSum := 0.0, 0.0
		for i := 0; i < mid; i++ {
			newerSum += quizDetails[i].Pct
		}
		for i := mid; i < len(quizDetails); i++ {
			olderSum += quizDetails[i].Pct
		}
		olderAvg := olderSum / float64(len(quizDetails)-mid)
		newerAvg := newerSum / float64(mid)
		if newerAvg > olderAvg+5 {
			trend = "up"
		} else if newerAvg < olderAvg-5 {
			trend = "down"
		}
	}

	// Language: default to French (Algeria)
	lang, err := c.Cookie("lang")
	if err != nil || (lang != "en" && lang != "fr") {
		lang = "fr"
	}

	// Fetch student invoices for this classroom
	invoices, _ := h.Store.GetStudentInvoices(ctx, data.StudentID, data.ClassroomID)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	if err := h.Tmpl.ExecuteTemplate(c.Writer, "parent_report.html", gin.H{
		"StudentName":   data.StudentName,
		"ClassroomName": data.ClassroomName,
		"TeacherName":   data.TeacherName,
		"QuizDetails":   quizDetails,
		"AssignDetails": assignDetails,
		"Remarks":       remarks,
		"Attendance":    displayAttendance,
		"Attended":      attended,
		"TotalSessions": totalSessions,
		"AttendancePct": attendancePct,
		"AvgQuizPct":    avgQuizPct,
		"AvgAssignPct":  avgAssignPct,
		"OverallAvg":    overallAvg,
		"ClassAvg":      classOverallAvg,
		"Banner":        banner,
		"Trend":         trend,
		"Invoices":      invoices,
		"GeneratedAt":   time.Now(),
		"Lang":          lang,
		"BaseURL":       h.BaseURL,
		"FooterLink":    fmt.Sprintf("%s/apply", h.BaseURL),
	}); err != nil {
		log.Printf("Parent report template error: %v", err)
		c.String(http.StatusInternalServerError, "Something went wrong")
	}
}

// RegenerateParentCode generates a new parent code for a student (invalidates old link).
func (h *Handler) RegenerateParentCode(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	studentID := 0
	fmt.Sscanf(c.Param("studentId"), "%d", &studentID)
	if studentID == 0 {
		c.String(http.StatusBadRequest, "Invalid student")
		return
	}

	_, err := h.Store.RegenerateParentCode(c.Request.Context(), classID, studentID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to regenerate code")
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/student/%d", classID, studentID))
}
