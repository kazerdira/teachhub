package handlers

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ─── CSV Export Handlers ────────────────────────────────

func (h *Handler) ExportRosterCSV(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	classroom, _ := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))

	data, err := h.Store.GetRosterExport(c.Request.Context(), classID)
	if err != nil {
		c.String(500, "Export failed")
		return
	}

	filename := fmt.Sprintf("%s_roster.csv", classroom.Name)
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"Name", "Email", "Quiz Avg %", "Quizzes Taken", "Assignment Avg %", "Assignments Submitted", "Attendance %", "Engagement %"})
	for _, r := range data {
		w.Write([]string{
			r.Name, r.Email,
			fmt.Sprintf("%.1f", r.AvgQuizPct),
			strconv.Itoa(r.QuizzesTaken),
			fmt.Sprintf("%.1f", r.AvgAssignmentPct),
			strconv.Itoa(r.AssignmentsSubmitted),
			fmt.Sprintf("%.1f", r.AttendancePct),
			fmt.Sprintf("%.1f", r.EngagementPct),
		})
	}
	w.Flush()
}

func (h *Handler) ExportQuizzesCSV(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	classroom, _ := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))

	data, err := h.Store.GetQuizExport(c.Request.Context(), classID)
	if err != nil {
		c.String(500, "Export failed")
		return
	}

	filename := fmt.Sprintf("%s_quiz_results.csv", classroom.Name)
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"Quiz", "Student", "Email", "Score", "Max Score", "Percentage", "Started At", "Finished At"})
	for _, r := range data {
		finished := ""
		if r.FinishedAt != nil {
			finished = r.FinishedAt.Format("2006-01-02 15:04:05")
		}
		w.Write([]string{
			r.QuizTitle, r.StudentName, r.StudentEmail,
			strconv.Itoa(r.Score), strconv.Itoa(r.MaxScore),
			fmt.Sprintf("%.1f", r.Pct),
			r.StartedAt.Format("2006-01-02 15:04:05"),
			finished,
		})
	}
	w.Flush()
}

func (h *Handler) ExportAssignmentsCSV(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	classroom, _ := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))

	data, err := h.Store.GetAssignmentExport(c.Request.Context(), classID)
	if err != nil {
		c.String(500, "Export failed")
		return
	}

	filename := fmt.Sprintf("%s_assignment_grades.csv", classroom.Name)
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"Assignment", "Student", "Email", "Grade", "Max Grade", "Percentage", "Status", "Submitted At"})
	for _, r := range data {
		grade := ""
		if r.Grade != nil {
			grade = fmt.Sprintf("%.1f", *r.Grade)
		}
		w.Write([]string{
			r.AssignmentTitle, r.StudentName, r.StudentEmail,
			grade, fmt.Sprintf("%.1f", r.MaxGrade),
			fmt.Sprintf("%.1f", r.Pct),
			r.Status,
			r.SubmittedAt.Format("2006-01-02 15:04:05"),
		})
	}
	w.Flush()
}

func (h *Handler) ExportAttendanceCSV(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	classroom, _ := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))

	data, err := h.Store.GetAttendanceExport(c.Request.Context(), classID)
	if err != nil {
		c.String(500, "Export failed")
		return
	}

	filename := fmt.Sprintf("%s_attendance.csv", classroom.Name)
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"Student", "Email", "Session Date", "Session Duration (min)", "Attended", "Joined At", "Left At", "Time Spent (min)"})
	for _, r := range data {
		dur := ""
		if r.SessionDuration != nil {
			dur = strconv.Itoa(*r.SessionDuration)
		}
		attended := "No"
		if r.Attended {
			attended = "Yes"
		}
		joined := ""
		if r.JoinedAt != nil {
			joined = r.JoinedAt.Format("2006-01-02 15:04:05")
		}
		left := ""
		if r.LeftAt != nil {
			left = r.LeftAt.Format("2006-01-02 15:04:05")
		}
		timeSpent := ""
		if r.Attended {
			timeSpent = strconv.Itoa(r.TimeSpentMin)
		}
		w.Write([]string{
			r.StudentName, r.StudentEmail,
			r.SessionDate.Format("2006-01-02 15:04:05"),
			dur, attended, joined, left, timeSpent,
		})
	}
	w.Flush()
}

// ─── Classroom Report (Print-friendly) ──────────────────

type reportStudent struct {
	Name                 string
	Email                string
	AvgQuizPct           float64
	QuizzesTaken         int
	QuizzesTotal         int
	AssignmentsSubmitted int
	AssignmentsTotal     int
	AvgAssignmentPct     float64
	AttendancePct        float64
	OverallPct           float64
	Status               string // "excellent","good","average","struggling"
}

func studentStatus(pct float64) string {
	switch {
	case pct >= 80:
		return "excellent"
	case pct >= 60:
		return "good"
	case pct >= 40:
		return "average"
	default:
		return "struggling"
	}
}

func (h *Handler) AdminClassroomReport(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	classroom, err := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))
	if err != nil {
		c.String(404, "Classroom not found")
		return
	}

	quizStats, _ := h.Store.GetQuizAnalytics(c.Request.Context(), classID)
	assignStats, _ := h.Store.GetAssignmentAnalytics(c.Request.Context(), classID)
	rosterStats, _ := h.Store.GetStudentRosterAnalytics(c.Request.Context(), classID)
	attendanceRates, _ := h.Store.GetStudentAttendanceRates(c.Request.Context(), classID)
	atRiskStudents, _ := h.Store.GetAtRiskStudents(c.Request.Context(), classID)

	// Build attendance lookup by student ID
	attendMap := make(map[int]float64, len(attendanceRates))
	for _, a := range attendanceRates {
		attendMap[a.StudentID] = a.AttendancePct
	}

	// Build per-student report cards merging roster + attendance
	students := make([]reportStudent, 0, len(rosterStats))
	for _, r := range rosterStats {
		attPct := attendMap[r.StudentID]
		// Overall = weighted average: 40% quiz, 40% assignments, 20% attendance
		parts := 0.0
		weights := 0.0
		if r.QuizzesTaken > 0 {
			parts += r.AvgQuizPct * 0.4
			weights += 0.4
		}
		if r.AssignmentsSubmitted > 0 {
			parts += r.AvgAssignmentPct * 0.4
			weights += 0.4
		}
		if attPct > 0 {
			parts += attPct * 0.2
			weights += 0.2
		}
		overall := 0.0
		if weights > 0 {
			overall = parts / weights * 1.0 // normalise back
		}
		students = append(students, reportStudent{
			Name:                 r.Name,
			Email:                r.Email,
			AvgQuizPct:           r.AvgQuizPct,
			QuizzesTaken:         r.QuizzesTaken,
			QuizzesTotal:         r.QuizzesTotal,
			AssignmentsSubmitted: r.AssignmentsSubmitted,
			AssignmentsTotal:     r.AssignmentsTotal,
			AvgAssignmentPct:     r.AvgAssignmentPct,
			AttendancePct:        attPct,
			OverallPct:           overall,
			Status:               studentStatus(overall),
		})
	}

	// Compute class-wide stats
	var avgQuizPct, avgAssignPct, avgAttendPct float64
	if len(quizStats) > 0 {
		total := 0.0
		for _, q := range quizStats {
			total += q.AvgPct
		}
		avgQuizPct = total / float64(len(quizStats))
	}
	if len(assignStats) > 0 {
		total := 0.0
		count := 0
		for _, a := range assignStats {
			if a.GradedCount > 0 {
				total += a.AvgPct
				count++
			}
		}
		if count > 0 {
			avgAssignPct = total / float64(count)
		}
	}
	if len(attendanceRates) > 0 {
		total := 0.0
		for _, a := range attendanceRates {
			total += a.AttendancePct
		}
		avgAttendPct = total / float64(len(attendanceRates))
	}

	now := time.Now()

	h.render(c, "admin_report.html", gin.H{
		"Classroom":       classroom,
		"QuizStats":       quizStats,
		"AssignmentStats": assignStats,
		"Students":        students,
		"AtRiskStudents":  atRiskStudents,
		"AvgQuizPct":      avgQuizPct,
		"AvgAssignPct":    avgAssignPct,
		"AvgAttendPct":    avgAttendPct,
		"StudentCount":    len(rosterStats),
		"Now":             now,
	})
}
