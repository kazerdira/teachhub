package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ─── Quiz CRUD ──────────────────────────────────────────

func (h *Handler) CreateQuiz(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	title := strings.TrimSpace(c.PostForm("title"))
	desc := strings.TrimSpace(c.PostForm("description"))
	if title == "" {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=quizzes", classID))
		return
	}

	var deadline *time.Time
	if dStr := c.PostForm("deadline"); dStr != "" {
		if t, err := time.Parse("2006-01-02T15:04", dStr); err == nil {
			deadline = &t
		}
	}
	timeLimitMin, _ := strconv.Atoi(c.DefaultPostForm("time_limit_minutes", "0"))
	maxAttempts, _ := strconv.Atoi(c.DefaultPostForm("max_attempts", "1"))

	h.Store.CreateQuiz(c.Request.Context(), classID, title, desc, deadline, timeLimitMin, maxAttempts)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=quizzes", classID))
}

func (h *Handler) DeleteQuiz(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	quizID, _ := strconv.Atoi(c.Param("quizId"))
	h.Store.DeleteQuiz(c.Request.Context(), quizID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=quizzes", classID))
}

func (h *Handler) ToggleQuizPublish(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	quizID, _ := strconv.Atoi(c.Param("quizId"))
	quiz, err := h.Store.GetQuiz(c.Request.Context(), quizID)
	if err == nil {
		h.Store.UpdateQuiz(c.Request.Context(), quizID, quiz.Title, quiz.Description, !quiz.Published, quiz.Deadline, quiz.TimeLimitMinutes, quiz.MaxAttempts)
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d?tab=quizzes", classID))
}

// ─── Quiz Editor ────────────────────────────────────────

func (h *Handler) EditQuiz(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	quizID, _ := strconv.Atoi(c.Param("quizId"))
	classroom, _ := h.Store.GetClassroom(c.Request.Context(), classID)
	quiz, _ := h.Store.GetQuiz(c.Request.Context(), quizID)
	questions, _ := h.Store.ListQuizQuestions(c.Request.Context(), quizID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	const perPage = 50
	offset := (page - 1) * perPage

	attempts, total, _ := h.Store.ListQuizAttemptsPaged(c.Request.Context(), quizID, perPage, offset)
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	h.render(c, "admin_quiz_edit.html", gin.H{
		"Classroom":  classroom,
		"Quiz":       quiz,
		"Questions":  questions,
		"Attempts":   attempts,
		"Page":       page,
		"TotalPages": totalPages,
		"Total":      total,
	})
}

func (h *Handler) AddQuestion(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	quizID, _ := strconv.Atoi(c.Param("quizId"))

	qType := c.PostForm("question_type")
	content := strings.TrimSpace(c.PostForm("content"))
	correctAnswer := strings.TrimSpace(c.PostForm("correct_answer"))
	points, _ := strconv.Atoi(c.DefaultPostForm("points", "1"))
	sortOrder, _ := strconv.Atoi(c.DefaultPostForm("sort_order", "0"))

	var options []string
	if qType == "mcq" {
		for i := 1; i <= 6; i++ {
			opt := strings.TrimSpace(c.PostForm(fmt.Sprintf("option_%d", i)))
			if opt != "" {
				options = append(options, opt)
			}
		}
	} else if qType == "true_false" {
		options = []string{"True", "False"}
	}

	h.Store.CreateQuizQuestion(c.Request.Context(), quizID, sortOrder, qType, content, options, correctAnswer, points)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/quiz/%d/edit", classID, quizID))
}

func (h *Handler) DeleteQuestion(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	quizID, _ := strconv.Atoi(c.Param("quizId"))
	qID, _ := strconv.Atoi(c.Param("qId"))
	h.Store.DeleteQuizQuestion(c.Request.Context(), qID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/quiz/%d/edit", classID, quizID))
}

func (h *Handler) UpdateQuestion(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	quizID, _ := strconv.Atoi(c.Param("quizId"))
	qID, _ := strconv.Atoi(c.Param("qId"))

	qType := c.PostForm("question_type")
	content := strings.TrimSpace(c.PostForm("content"))
	correctAnswer := strings.TrimSpace(c.PostForm("correct_answer"))
	points, _ := strconv.Atoi(c.DefaultPostForm("points", "1"))

	var options []string
	if qType == "mcq" {
		for i := 1; i <= 6; i++ {
			opt := strings.TrimSpace(c.PostForm(fmt.Sprintf("option_%d", i)))
			if opt != "" {
				options = append(options, opt)
			}
		}
	} else if qType == "true_false" {
		options = []string{"True", "False"}
	}

	h.Store.UpdateQuizQuestion(c.Request.Context(), qID, qType, content, options, correctAnswer, points)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/quiz/%d/edit", classID, quizID))
}

func (h *Handler) UpdateQuizSettings(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	quizID, _ := strconv.Atoi(c.Param("quizId"))

	quiz, err := h.Store.GetQuiz(c.Request.Context(), quizID)
	if err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/quiz/%d/edit", classID, quizID))
		return
	}

	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		title = quiz.Title
	}
	desc := strings.TrimSpace(c.PostForm("description"))

	var deadline *time.Time
	if dStr := c.PostForm("deadline"); dStr != "" {
		if t, err := time.Parse("2006-01-02T15:04", dStr); err == nil {
			deadline = &t
		}
	}
	timeLimitMin, _ := strconv.Atoi(c.DefaultPostForm("time_limit_minutes", "0"))
	maxAttempts, _ := strconv.Atoi(c.DefaultPostForm("max_attempts", "1"))

	h.Store.UpdateQuiz(c.Request.Context(), quizID, title, desc, quiz.Published, deadline, timeLimitMin, maxAttempts)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/quiz/%d/edit", classID, quizID))
}

// ─── Quiz Attempt Review ────────────────────────────────

func (h *Handler) ReviewAttempt(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	quizID, _ := strconv.Atoi(c.Param("quizId"))
	attemptID, _ := strconv.Atoi(c.Param("attemptId"))
	score, _ := strconv.Atoi(c.PostForm("score"))
	h.Store.ReviewQuizAttempt(c.Request.Context(), attemptID, score)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/quiz/%d/edit", classID, quizID))
}

// ─── AI Quiz Generation ─────────────────────────────────

type aiQuestion struct {
	QuestionType  string   `json:"question_type"`
	Content       string   `json:"content"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correct_answer"`
}

func (h *Handler) GenerateQuizAI(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}
	quizID, _ := strconv.Atoi(c.Param("quizId"))

	topic := strings.TrimSpace(c.PostForm("topic"))
	numQuestions := c.DefaultPostForm("num_questions", "5")
	difficulty := c.DefaultPostForm("difficulty", "intermediate")
	qTypes := c.PostForm("question_types") // comma-separated: mcq,true_false,fill_blank

	if topic == "" || h.ClaudeKey == "" {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/quiz/%d/edit?error=missing", classID, quizID))
		return
	}

	prompt := fmt.Sprintf(`Generate %s quiz questions about: %s
Difficulty: %s
Question types to use: %s

Return ONLY a JSON array. Each object must have:
- "question_type": one of "mcq", "true_false", "fill_blank", "open_ended"
- "content": the question text
- "options": array of choices (for mcq: 4 options, for true_false: ["True","False"], for fill_blank/open_ended: [])
- "correct_answer": the correct answer text (for open_ended, provide a sample answer)

No markdown, no explanation, just the JSON array.`, numQuestions, topic, difficulty, qTypes)

	body, _ := json.Marshal(map[string]any{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 4096,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	})

	req, _ := http.NewRequestWithContext(c.Request.Context(), "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", h.ClaudeKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/quiz/%d/edit?error=ai_failed", classID, quizID))
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var aiResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	json.Unmarshal(respBody, &aiResp)
	if len(aiResp.Content) == 0 {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/quiz/%d/edit?error=ai_empty", classID, quizID))
		return
	}

	text := aiResp.Content[0].Text
	// Strip markdown fences if present
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var questions []aiQuestion
	if err := json.Unmarshal([]byte(text), &questions); err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/quiz/%d/edit?error=ai_parse", classID, quizID))
		return
	}

	for i, q := range questions {
		h.Store.CreateQuizQuestion(c.Request.Context(), quizID, i+1, q.QuestionType, q.Content, q.Options, q.CorrectAnswer, 1)
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/quiz/%d/edit", classID, quizID))
}
