package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ─── Analytics Dashboard ────────────────────────────────

func (h *Handler) AdminAnalytics(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	classroom, err := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))
	if err != nil {
		c.String(404, "Classroom not found")
		return
	}

	sub := c.DefaultQuery("sub", "quizzes")

	data := gin.H{
		"Classroom": classroom,
		"Tab":       "analytics",
		"Sub":       sub,
	}

	switch sub {
	case "quizzes":
		quizStats, _ := h.Store.GetQuizAnalytics(c.Request.Context(), classID)

		// If a specific quiz is selected, get question breakdown + student breakdown
		var questionStats []interface{}
		var studentBreakdown []interface{}
		var selectedQuizID int
		var selectedQuizTitle string

		if qidStr := c.Query("quiz"); qidStr != "" {
			selectedQuizID, _ = strconv.Atoi(qidStr)
			qAnalytics, _ := h.Store.GetQuestionAnalytics(c.Request.Context(), selectedQuizID)
			attempts, _ := h.Store.GetQuizStudentBreakdown(c.Request.Context(), selectedQuizID)

			for _, qa := range qAnalytics {
				questionStats = append(questionStats, qa)
			}
			for _, a := range attempts {
				studentBreakdown = append(studentBreakdown, a)
			}

			// Find quiz title
			for _, qs := range quizStats {
				if qs.QuizID == selectedQuizID {
					selectedQuizTitle = qs.Title
					break
				}
			}
		}

		data["QuizStats"] = quizStats
		data["QuestionStats"] = questionStats
		data["StudentBreakdown"] = studentBreakdown
		data["SelectedQuizID"] = selectedQuizID
		data["SelectedQuizTitle"] = selectedQuizTitle

	case "assignments":
		assignStats, _ := h.Store.GetAssignmentAnalytics(c.Request.Context(), classID)
		data["AssignmentStats"] = assignStats

		// If a specific assignment is selected, show missing submissions
		if aidStr := c.Query("assign"); aidStr != "" {
			aid, _ := strconv.Atoi(aidStr)
			missing, _ := h.Store.GetMissingSubmissions(c.Request.Context(), aid, classID)
			data["MissingStudents"] = missing
			data["SelectedAssignID"] = aid
			for _, as := range assignStats {
				if as.AssignmentID == aid {
					data["SelectedAssignTitle"] = as.Title
					break
				}
			}
		}

	case "students":
		rosterStats, _ := h.Store.GetStudentRosterAnalytics(c.Request.Context(), classID)
		data["RosterStats"] = rosterStats

	case "live":
		sessionHistory, _ := h.Store.GetSessionHistory(c.Request.Context(), classID)
		attendanceRates, _ := h.Store.GetStudentAttendanceRates(c.Request.Context(), classID)
		data["SessionHistory"] = sessionHistory
		data["AttendanceRates"] = attendanceRates

		// If a specific session is selected, show its attendance
		if sidStr := c.Query("session"); sidStr != "" {
			sid, _ := strconv.Atoi(sidStr)
			sessionAttendance, _ := h.Store.GetSessionAttendance(c.Request.Context(), sid)
			data["SessionAttendance"] = sessionAttendance
			data["SelectedSessionID"] = sid
		}

	case "trends":
		quizTrends, _ := h.Store.GetQuizTrends(c.Request.Context(), classID)
		assignTrends, _ := h.Store.GetAssignmentTrends(c.Request.Context(), classID)
		timingStats, _ := h.Store.GetSubmissionTimingStats(c.Request.Context(), classID)
		data["QuizTrends"] = quizTrends
		data["AssignmentTrends"] = assignTrends
		data["TimingStats"] = timingStats

	case "risk":
		atRiskStudents, _ := h.Store.GetAtRiskStudents(c.Request.Context(), classID)
		data["AtRiskStudents"] = atRiskStudents

	case "resources":
		resourceViews, _ := h.Store.GetResourceViewCounts(c.Request.Context(), classID)
		data["ResourceViews"] = resourceViews
	}

	h.render(c, "admin_analytics.html", data)
}

// ─── Student Detail Page ────────────────────────────────

func (h *Handler) AdminStudentDetail(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	studentID, _ := strconv.Atoi(c.Param("studentId"))

	classroom, err := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))
	if err != nil {
		c.String(404, "Classroom not found")
		return
	}

	student, err := h.Store.GetStudent(c.Request.Context(), studentID)
	if err != nil {
		c.String(404, "Student not found")
		return
	}

	quizDetails, _ := h.Store.GetStudentQuizDetails(c.Request.Context(), studentID, classID)
	assignDetails, _ := h.Store.GetStudentAssignmentDetails(c.Request.Context(), studentID, classID)
	remarks, _ := h.Store.GetStudentRemarks(c.Request.Context(), studentID, classID)

	// Compute overall averages
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

	// Get parent code for this student
	parentCode, _ := h.Store.GetParentCode(c.Request.Context(), classID, studentID)
	parentURL := ""
	if parentCode != "" {
		parentURL = fmt.Sprintf("%s/p/%s", h.BaseURL, parentCode)
	}

	h.render(c, "admin_student_detail.html", gin.H{
		"Classroom":     classroom,
		"Student":       student,
		"QuizDetails":   quizDetails,
		"AssignDetails": assignDetails,
		"AvgQuizPct":    avgQuizPct,
		"AvgAssignPct":  avgAssignPct,
		"Remarks":       remarks,
		"ParentCode":    parentCode,
		"ParentURL":     parentURL,
	})
}

// ─── Analytics API (for HTMX partials) ──────────────────

func (h *Handler) AnalyticsMissing(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	assignID, _ := strconv.Atoi(c.Param("assignId"))
	missing, _ := h.Store.GetMissingSubmissions(c.Request.Context(), assignID, classID)
	c.Redirect(http.StatusFound,
		fmt.Sprintf("/admin/classroom/%d/analytics?sub=assignments&assign=%d", classID, assignID))
	_ = missing // redirect handles display
}
