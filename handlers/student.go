package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"teachhub/middleware"
	"teachhub/store"

	"github.com/gin-gonic/gin"
)

// ─── CGU / Privacy ──────────────────────────────────────

func (h *Handler) CGUPage(c *gin.Context) {
	h.render(c, "cgu.html", gin.H{})
}

// ─── Home / Join ────────────────────────────────────────

func (h *Handler) Home(c *gin.Context) {
	student := middleware.GetStudent(c)
	if student != nil {
		classrooms, _ := h.Store.GetStudentClassrooms(c.Request.Context(), student.ID)
		// Only update last-login once per day to avoid a DB write on every page load
		if student.LastLoginAt == nil || time.Since(*student.LastLoginAt) > 24*time.Hour {
			h.Store.UpdateStudentLastLogin(c.Request.Context(), student.ID, c.ClientIP())
		}
		h.render(c, "student_home.html", gin.H{"Student": student, "Classrooms": classrooms})
		return
	}
	// Non-authenticated visitors see the landing page
	h.render(c, "landing.html", gin.H{})
}

func (h *Handler) JoinPage(c *gin.Context) {
	code := c.Param("code")
	classroom, err := h.Store.GetClassroomByCode(c.Request.Context(), code)
	if err != nil {
		h.render(c, "student_join.html", gin.H{"Error": "Invalid join code"})
		return
	}

	student := middleware.GetStudent(c)
	if student != nil {
		// Already logged in — check membership
		isMember, status, _ := h.Store.IsStudentMemberOfClassroom(c.Request.Context(), student.ID, classroom.ID)
		if isMember {
			if status == "approved" {
				c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d", classroom.ID))
				return
			}
			// Pending or rejected
			h.render(c, "student_join.html", gin.H{"Classroom": classroom, "Code": code, "Status": status})
			return
		}
		// Not a member yet — check allow-list by email
		if student.Email != "" {
			allowed, _ := h.Store.IsEmailAllowed(c.Request.Context(), student.Email, classroom.ID)
			if allowed {
				h.Store.CreateStudentAndJoinExisting(c.Request.Context(), student.ID, classroom.ID)
				c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d", classroom.ID))
				return
			}
		}
		// Not on allow-list, join as pending
		h.Store.CreateStudentAndJoinExistingWithStatus(c.Request.Context(), student.ID, classroom.ID, "pending")
		h.render(c, "student_join.html", gin.H{"Classroom": classroom, "Code": code, "Status": "pending"})
		return
	}

	h.render(c, "student_join.html", gin.H{"Classroom": classroom, "Code": code})
}

func sanitizePhone(p string) string {
	var b strings.Builder
	for _, r := range p {
		if r >= '0' && r <= '9' || r == '+' {
			b.WriteRune(r)
		}
	}
	s := b.String()
	if len(s) > 15 {
		s = s[:15]
	}
	return s
}

func (h *Handler) JoinClassroom(c *gin.Context) {
	code := c.Param("code")
	name := strings.TrimSpace(c.PostForm("name"))
	email := strings.TrimSpace(c.PostForm("email"))
	phone := sanitizePhone(c.PostForm("phone"))

	classroom, err := h.Store.GetClassroomByCode(c.Request.Context(), code)
	if err != nil {
		h.render(c, "student_join.html", gin.H{"Error": "Invalid code"})
		return
	}

	// Check if student already in session
	student := middleware.GetStudent(c)
	if student != nil {
		isMember, status, _ := h.Store.IsStudentMemberOfClassroom(c.Request.Context(), student.ID, classroom.ID)
		if isMember {
			if status == "approved" {
				c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d", classroom.ID))
			} else {
				h.render(c, "student_join.html", gin.H{"Classroom": classroom, "Code": code, "Status": status})
			}
			return
		}
		// Check allow-list
		if student.Email != "" {
			allowed, _ := h.Store.IsEmailAllowed(c.Request.Context(), student.Email, classroom.ID)
			if allowed {
				h.Store.CreateStudentAndJoinExisting(c.Request.Context(), student.ID, classroom.ID)
				c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d", classroom.ID))
				return
			}
		}
		h.Store.CreateStudentAndJoinExistingWithStatus(c.Request.Context(), student.ID, classroom.ID, "pending")
		h.render(c, "student_join.html", gin.H{"Classroom": classroom, "Code": code, "Status": "pending"})
		return
	}

	if name == "" {
		h.render(c, "student_join.html", gin.H{"Classroom": classroom, "Code": code, "Error": "Name is required"})
		return
	}
	if email == "" {
		h.render(c, "student_join.html", gin.H{"Classroom": classroom, "Code": code, "Error": "Email is required"})
		return
	}

	// Check if a student with this email already exists in this classroom
	// (e.g. created via an approved join request — teacher sent them the link)
	existingID, existingStatus, err := h.Store.FindStudentByEmailInClassroom(c.Request.Context(), email, classroom.ID)
	if err == nil && existingID > 0 {
		middleware.SetStudentSession(c, existingID)
		h.Store.UpdateStudentLastLogin(c.Request.Context(), existingID, c.ClientIP())
		if existingStatus == "approved" {
			c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d", classroom.ID))
		} else {
			h.render(c, "student_join.html", gin.H{"Classroom": classroom, "Code": code, "Status": existingStatus})
		}
		return
	}

	// Check allow-list
	allowed, _ := h.Store.IsEmailAllowed(c.Request.Context(), email, classroom.ID)
	status := "pending"
	if allowed {
		status = "approved"
	}

	studentID, err := h.Store.CreateStudentAndJoinWithStatus(c.Request.Context(), name, email, phone, classroom.ID, status)
	if err != nil {
		h.render(c, "student_join.html", gin.H{"Classroom": classroom, "Code": code, "Error": "Failed to join"})
		return
	}
	middleware.SetStudentSession(c, studentID)
	h.Store.UpdateStudentLastLogin(c.Request.Context(), studentID, c.ClientIP())

	if status == "approved" {
		c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d", classroom.ID))
	} else {
		h.render(c, "student_join.html", gin.H{"Classroom": classroom, "Code": code, "Status": "pending"})
	}
}

// ─── Student Classroom View ─────────────────────────────

func (h *Handler) StudentClassroom(c *gin.Context) {
	student := middleware.GetStudent(c)
	classID, _ := strconv.Atoi(c.Param("id"))

	in, _ := h.Store.IsStudentInClassroom(c.Request.Context(), student.ID, classID)
	if !in {
		c.Redirect(http.StatusFound, "/")
		return
	}

	classroom, _ := h.Store.GetClassroom(c.Request.Context(), classID)
	tab := c.DefaultQuery("tab", "resources")
	liveSession, _ := h.Store.GetActiveLiveSession(c.Request.Context(), classID)

	data := gin.H{
		"Classroom":   classroom,
		"Student":     student,
		"Tab":         tab,
		"LiveSession": liveSession,
	}

	switch tab {
	case "resources":
		resources, _ := h.Store.ListResources(c.Request.Context(), classID)
		data["Resources"] = resources
	case "assignments":
		assignments, _ := h.Store.ListAssignments(c.Request.Context(), classID)
		data["Assignments"] = assignments
	case "quizzes":
		quizzes, _ := h.Store.ListQuizzes(c.Request.Context(), classID)
		var published []store.Quiz
		for _, q := range quizzes {
			if q.Published {
				published = append(published, q)
			}
		}
		data["Quizzes"] = published
	}

	h.render(c, "student_classroom.html", data)
}

// ─── Student Resource Download ──────────────────────────

func (h *Handler) StudentDownload(c *gin.Context) {
	h.DownloadResource(c) // reuse admin download handler
}

// ─── Student Submissions ────────────────────────────────

func (h *Handler) StudentAssignment(c *gin.Context) {
	student := middleware.GetStudent(c)
	classID, _ := strconv.Atoi(c.Param("id"))
	assignID, _ := strconv.Atoi(c.Param("assignId"))

	classroom, _ := h.Store.GetClassroom(c.Request.Context(), classID)
	assignment, _ := h.Store.GetAssignment(c.Request.Context(), assignID)
	submissions, _ := h.Store.GetStudentSubmissions(c.Request.Context(), assignID, student.ID)

	h.render(c, "student_assignment.html", gin.H{
		"Classroom":   classroom,
		"Assignment":  assignment,
		"Submissions": submissions,
		"Student":     student,
	})
}

func (h *Handler) StudentSubmit(c *gin.Context) {
	student := middleware.GetStudent(c)
	classID, _ := strconv.Atoi(c.Param("id"))
	assignID, _ := strconv.Atoi(c.Param("assignId"))

	assignment, err := h.Store.GetAssignment(c.Request.Context(), assignID)
	if err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d/assignment/%d", classID, assignID))
		return
	}

	// Deadline enforcement
	if assignment.Deadline != nil && time.Now().After(*assignment.Deadline) {
		classroom, _ := h.Store.GetClassroom(c.Request.Context(), classID)
		submissions, _ := h.Store.GetStudentSubmissions(c.Request.Context(), assignID, student.ID)
		h.render(c, "student_assignment.html", gin.H{
			"Classroom": classroom, "Assignment": assignment,
			"Submissions": submissions, "Student": student,
			"Error": "The deadline for this assignment has passed. Submissions are no longer accepted.",
		})
		return
	}

	textContent := strings.TrimSpace(c.PostForm("text_content"))
	var filePath, fileName string
	var fileSize int64

	// Handle file upload if response type allows it
	if assignment.ResponseType == "file" || assignment.ResponseType == "both" {
		file, header, ferr := c.Request.FormFile("file")
		if ferr == nil {
			defer file.Close()

			// Validate extension (student whitelist)
			ext := strings.ToLower(filepath.Ext(header.Filename))
			if !isStudentAllowedExtension(ext) {
				classroom, _ := h.Store.GetClassroom(c.Request.Context(), classID)
				submissions, _ := h.Store.GetStudentSubmissions(c.Request.Context(), assignID, student.ID)
				h.render(c, "student_assignment.html", gin.H{
					"Classroom": classroom, "Assignment": assignment,
					"Submissions": submissions, "Student": student,
					"Error": "This file type is not allowed. Accepted: PDF, Word, Excel, PowerPoint, images, audio, video, archives.",
				})
				return
			}

			// Validate file size
			if assignment.MaxFileSize > 0 && header.Size > assignment.MaxFileSize {
				classroom, _ := h.Store.GetClassroom(c.Request.Context(), classID)
				submissions, _ := h.Store.GetStudentSubmissions(c.Request.Context(), assignID, student.ID)
				h.render(c, "student_assignment.html", gin.H{
					"Classroom": classroom, "Assignment": assignment,
					"Submissions": submissions, "Student": student,
					"Error": fmt.Sprintf("File too large. Maximum size is %d MB.", assignment.MaxFileSize/(1024*1024)),
				})
				return
			}

			fname := fmt.Sprintf("%d_%d_%d%s", assignID, student.ID, time.Now().UnixMilli(), ext)
			filePath = filepath.Join("submissions", fname)
			fullPath := filepath.Join(h.UploadDir, filePath)
			os.MkdirAll(filepath.Dir(fullPath), 0755)

			dst, derr := os.Create(fullPath)
			if derr == nil {
				defer dst.Close()
				fileSize, _ = io.Copy(dst, file)
			}
			fileName = header.Filename
		} else if assignment.ResponseType == "file" {
			// File required but not provided
			c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d/assignment/%d", classID, assignID))
			return
		}
	}

	// Validate text content
	if (assignment.ResponseType == "text" || assignment.ResponseType == "both") && textContent == "" && filePath == "" {
		c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d/assignment/%d", classID, assignID))
		return
	}
	if assignment.MaxChars > 0 && len(textContent) > assignment.MaxChars {
		textContent = textContent[:assignment.MaxChars]
	}

	h.Store.CreateSubmission(c.Request.Context(), assignID, student.ID, filePath, fileName, fileSize, textContent)
	c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d/assignment/%d", classID, assignID))
}

// ─── Student Quiz Taking ────────────────────────────────

func (h *Handler) StudentQuiz(c *gin.Context) {
	student := middleware.GetStudent(c)
	classID, _ := strconv.Atoi(c.Param("id"))
	quizID, _ := strconv.Atoi(c.Param("quizId"))

	classroom, _ := h.Store.GetClassroom(c.Request.Context(), classID)
	quiz, _ := h.Store.GetQuiz(c.Request.Context(), quizID)
	if !quiz.Published {
		c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d?tab=quizzes", classID))
		return
	}

	questions, _ := h.Store.ListQuizQuestions(c.Request.Context(), quizID)
	attempts, _ := h.Store.GetAllStudentAttempts(c.Request.Context(), quizID, student.ID)
	attemptCount := len(attempts)

	// Check if can still take quiz
	canTake := true
	var message string
	if quiz.MaxAttempts > 0 && attemptCount >= quiz.MaxAttempts {
		canTake = false
		message = fmt.Sprintf("You have used all %d attempts.", quiz.MaxAttempts)
	}
	if quiz.Deadline != nil && time.Now().After(*quiz.Deadline) {
		canTake = false
		message = "The deadline for this quiz has passed."
	}

	h.render(c, "student_quiz.html", gin.H{
		"Classroom":    classroom,
		"Quiz":         quiz,
		"Questions":    questions,
		"Attempts":     attempts,
		"AttemptCount": attemptCount,
		"CanTake":      canTake,
		"Message":      message,
		"Student":      student,
	})
}

func (h *Handler) StudentSubmitQuiz(c *gin.Context) {
	student := middleware.GetStudent(c)
	classID, _ := strconv.Atoi(c.Param("id"))
	quizID, _ := strconv.Atoi(c.Param("quizId"))

	quiz, _ := h.Store.GetQuiz(c.Request.Context(), quizID)
	if quiz == nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d", classID))
		return
	}

	// Deadline enforcement
	if quiz.Deadline != nil && time.Now().After(*quiz.Deadline) {
		c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d/quiz/%d", classID, quizID))
		return
	}

	// Server-side time limit enforcement: check if student has an open attempt that started too long ago
	if quiz.TimeLimitMinutes > 0 {
		openAttempt, _ := h.Store.GetOpenAttempt(c.Request.Context(), quizID, student.ID)
		if openAttempt != nil {
			elapsed := time.Since(openAttempt.StartedAt)
			// Allow 60s grace period for network delays
			if elapsed > time.Duration(quiz.TimeLimitMinutes)*time.Minute+60*time.Second {
				c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d/quiz/%d", classID, quizID))
				return
			}
		}
	}

	// Attempt limit enforcement
	attemptCount, _ := h.Store.CountStudentAttempts(c.Request.Context(), quizID, student.ID)
	if quiz.MaxAttempts > 0 && attemptCount >= quiz.MaxAttempts {
		c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d/quiz/%d", classID, quizID))
		return
	}

	questions, _ := h.Store.ListQuizQuestions(c.Request.Context(), quizID)

	// Collect answers
	answers := make(map[string]string)
	fileAnswers := make(map[string]map[string]string)
	score := 0
	maxScore := 0
	hasOpenEnded := false

	for _, q := range questions {
		key := fmt.Sprintf("q_%d", q.ID)
		qIDStr := strconv.Itoa(q.ID)
		maxScore += q.Points

		if q.QuestionType == "file_upload" {
			// Handle file upload answer
			file, header, ferr := c.Request.FormFile(key)
			if ferr == nil {
				// Validate extension and size
				ext := strings.ToLower(filepath.Ext(header.Filename))
				if !isStudentAllowedExtension(ext) || header.Size > MaxStudentFileSize {
					file.Close()
					// Skip bad files silently (quiz continues)
					hasOpenEnded = true
					continue
				}

				fname := fmt.Sprintf("quiz_%d_q%d_s%d_%d%s", quizID, q.ID, student.ID, time.Now().UnixMilli(), ext)
				fPath := filepath.Join("submissions", fname)
				fullPath := filepath.Join(h.UploadDir, fPath)
				os.MkdirAll(filepath.Dir(fullPath), 0755)

				dst, derr := os.Create(fullPath)
				if derr == nil {
					io.Copy(dst, file)
					dst.Close()
				}
				file.Close()
				fileAnswers[qIDStr] = map[string]string{
					"file_path": fPath,
					"file_name": header.Filename,
				}
			}
			hasOpenEnded = true // file_upload always needs manual review
			continue
		}

		answer := strings.TrimSpace(c.PostForm(key))
		answers[qIDStr] = answer

		switch q.QuestionType {
		case "mcq", "true_false":
			if strings.EqualFold(answer, q.CorrectAnswer) {
				score += q.Points
			}
		case "fill_blank":
			if strings.EqualFold(strings.TrimSpace(answer), strings.TrimSpace(q.CorrectAnswer)) {
				score += q.Points
			}
		case "open_ended":
			hasOpenEnded = true
		}
	}

	// Create attempt (atomic: checks max_attempts in the same INSERT to prevent race conditions)
	attemptID, err := h.Store.CreateQuizAttemptAtomic(c.Request.Context(), quizID, student.ID, quiz.MaxAttempts)
	if err != nil || attemptID == 0 {
		// Race condition caught: another submission sneaked in
		c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d/quiz/%d", classID, quizID))
		return
	}
	_ = hasOpenEnded // mark as needing review handled by reviewed=false default
	h.Store.SubmitQuizAttempt(c.Request.Context(), attemptID, answers, fileAnswers, score, maxScore)

	c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d/quiz/%d", classID, quizID))
}

// ─── Student Quiz Result (JSON for HTMX) ────────────────

func (h *Handler) StudentQuizResult(c *gin.Context) {
	student := middleware.GetStudent(c)
	classID, _ := strconv.Atoi(c.Param("id"))
	quizID, _ := strconv.Atoi(c.Param("quizId"))
	// Verify student belongs to this classroom
	in, _ := h.Store.IsStudentInClassroom(c.Request.Context(), student.ID, classID)
	if !in {
		c.JSON(403, gin.H{"error": "access denied"})
		return
	}
	attempt, _ := h.Store.GetStudentAttempt(c.Request.Context(), quizID, student.ID)
	if attempt == nil {
		c.JSON(200, gin.H{"status": "not_taken"})
		return
	}
	questions, _ := h.Store.ListQuizQuestions(c.Request.Context(), quizID)

	type resultQ struct {
		Content       string `json:"content"`
		YourAnswer    string `json:"your_answer"`
		CorrectAnswer string `json:"correct_answer"`
		Correct       bool   `json:"correct"`
		Type          string `json:"type"`
	}
	var results []resultQ
	for _, q := range questions {
		ans := attempt.Answers[strconv.Itoa(q.ID)]
		correct := false
		if q.QuestionType != "open_ended" {
			correct = strings.EqualFold(strings.TrimSpace(ans), strings.TrimSpace(q.CorrectAnswer))
		}
		results = append(results, resultQ{
			Content:       q.Content,
			YourAnswer:    ans,
			CorrectAnswer: q.CorrectAnswer,
			Correct:       correct,
			Type:          q.QuestionType,
		})
	}
	c.JSON(200, gin.H{"attempt": attempt, "results": results})
}
