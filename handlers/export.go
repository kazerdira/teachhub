package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ─── Export Handlers (Print-friendly HTML) ───────────────

func (h *Handler) ExportRosterCSV(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	classroom, _ := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))

	data, err := h.Store.GetRosterExport(c.Request.Context(), classID)
	if err != nil {
		c.String(500, "Export failed")
		return
	}

	h.render(c, "admin_export_roster.html", gin.H{
		"Classroom": classroom,
		"Data":      data,
		"Now":       time.Now(),
	})
}

type quizGroup struct {
	Title string
	Rows  []quizGroupRow
}
type quizGroupRow struct {
	StudentName  string
	StudentEmail string
	Score        int
	MaxScore     int
	Pct          float64
	StartedAt    time.Time
}

func (h *Handler) ExportQuizzesCSV(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	classroom, _ := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))

	data, err := h.Store.GetQuizExport(c.Request.Context(), classID)
	if err != nil {
		c.String(500, "Export failed")
		return
	}

	// Group by quiz title
	var groups []quizGroup
	groupMap := map[string]int{}
	for _, r := range data {
		idx, ok := groupMap[r.QuizTitle]
		if !ok {
			idx = len(groups)
			groupMap[r.QuizTitle] = idx
			groups = append(groups, quizGroup{Title: r.QuizTitle})
		}
		groups[idx].Rows = append(groups[idx].Rows, quizGroupRow{
			StudentName:  r.StudentName,
			StudentEmail: r.StudentEmail,
			Score:        r.Score,
			MaxScore:     r.MaxScore,
			Pct:          r.Pct,
			StartedAt:    r.StartedAt,
		})
	}

	h.render(c, "admin_export_quizzes.html", gin.H{
		"Classroom": classroom,
		"Groups":    groups,
		"Now":       time.Now(),
	})
}

type assignGroup struct {
	Title    string
	MaxGrade float64
	Rows     []assignGroupRow
}
type assignGroupRow struct {
	StudentName  string
	StudentEmail string
	Grade        *float64
	MaxGrade     float64
	Pct          float64
	Status       string
	SubmittedAt  time.Time
}

func (h *Handler) ExportAssignmentsCSV(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	classroom, _ := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))

	data, err := h.Store.GetAssignmentExport(c.Request.Context(), classID)
	if err != nil {
		c.String(500, "Export failed")
		return
	}

	// Group by assignment title
	var groups []assignGroup
	groupMap := map[string]int{}
	for _, r := range data {
		idx, ok := groupMap[r.AssignmentTitle]
		if !ok {
			idx = len(groups)
			groupMap[r.AssignmentTitle] = idx
			groups = append(groups, assignGroup{Title: r.AssignmentTitle, MaxGrade: r.MaxGrade})
		}
		groups[idx].Rows = append(groups[idx].Rows, assignGroupRow{
			StudentName:  r.StudentName,
			StudentEmail: r.StudentEmail,
			Grade:        r.Grade,
			MaxGrade:     r.MaxGrade,
			Pct:          r.Pct,
			Status:       r.Status,
			SubmittedAt:  r.SubmittedAt,
		})
	}

	h.render(c, "admin_export_assignments.html", gin.H{
		"Classroom": classroom,
		"Groups":    groups,
		"Now":       time.Now(),
	})
}

type attendSession struct {
	Date     time.Time
	Duration *int
	Students []attendStudent
}
type attendStudent struct {
	Name         string
	Attended     bool
	TimeSpentMin int
}

func (h *Handler) ExportAttendanceCSV(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	classroom, _ := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))

	data, err := h.Store.GetAttendanceExport(c.Request.Context(), classID)
	if err != nil {
		c.String(500, "Export failed")
		return
	}

	// Get per-student summary
	summary, _ := h.Store.GetStudentAttendanceRates(c.Request.Context(), classID)

	// Group by session date
	var sessions []attendSession
	sessionMap := map[string]int{}
	for _, r := range data {
		key := r.SessionDate.Format("2006-01-02T15:04")
		idx, ok := sessionMap[key]
		if !ok {
			idx = len(sessions)
			sessionMap[key] = idx
			sessions = append(sessions, attendSession{Date: r.SessionDate, Duration: r.SessionDuration})
		}
		sessions[idx].Students = append(sessions[idx].Students, attendStudent{
			Name:         r.StudentName,
			Attended:     r.Attended,
			TimeSpentMin: r.TimeSpentMin,
		})
	}

	h.render(c, "admin_export_attendance.html", gin.H{
		"Classroom": classroom,
		"Summary":   summary,
		"Sessions":  sessions,
		"Now":       time.Now(),
	})
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
