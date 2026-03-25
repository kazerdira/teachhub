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

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ─── LiveKit Token Generation ───────────────────────────

type liveKitGrant struct {
	RoomJoin     bool   `json:"roomJoin,omitempty"`
	Room         string `json:"room,omitempty"`
	CanPublish   bool   `json:"canPublish,omitempty"`
	CanSubscribe bool   `json:"canSubscribe,omitempty"`
}

type liveKitClaims struct {
	jwt.RegisteredClaims
	Video *liveKitGrant `json:"video,omitempty"`
	Name  string        `json:"name,omitempty"`
}

func (h *Handler) generateLiveKitToken(roomName, identity, displayName string, canPublish bool) (string, error) {
	claims := liveKitClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    h.LKApiKey,
			Subject:   identity,
			ID:        fmt.Sprintf("%s-%s-%d", roomName, identity, time.Now().UnixNano()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Video: &liveKitGrant{
			RoomJoin:     true,
			Room:         roomName,
			CanPublish:   canPublish,
			CanSubscribe: true,
		},
		Name: displayName,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.LKApiSecret))
}

// ─── Admin Live Class Handlers ──────────────────────────

func (h *Handler) StartLiveClass(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))

	// Verify ownership
	_, err := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))
	if err != nil {
		c.String(403, "Not your classroom")
		return
	}

	roomName := fmt.Sprintf("classroom-%d", classID)

	_, err = h.Store.CreateLiveSession(c.Request.Context(), classID, roomName)
	if err != nil {
		fmt.Printf("LiveKit: CreateLiveSession error: %v\n", err)
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d", classID))
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d/live", classID))
}

func (h *Handler) EndLiveClass(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))

	// Verify ownership
	_, err := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))
	if err != nil {
		c.String(403, "Not your classroom")
		return
	}

	h.Store.EndLiveSession(c.Request.Context(), classID)
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d", classID))
}

func (h *Handler) AdminLivePage(c *gin.Context) {
	classID, _ := strconv.Atoi(c.Param("id"))
	classroom, _ := h.Store.GetClassroomForAdmin(c.Request.Context(), classID, adminID(c))
	liveSession, _ := h.Store.GetActiveLiveSession(c.Request.Context(), classID)

	if liveSession == nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/classroom/%d", classID))
		return
	}

	token, err := h.generateLiveKitToken(liveSession.RoomName, "admin", "Teacher", true)
	if err != nil {
		c.String(500, "Token error: %v", err)
		return
	}

	h.render(c, "admin_live.html", gin.H{
		"Classroom":   classroom,
		"LiveSession": liveSession,
		"Token":       token,
		"LKUrl":       h.LKUrl,
	})
}

// ─── Student Live Class Handler ─────────────────────────

func (h *Handler) StudentLivePage(c *gin.Context) {
	student := middleware.GetStudent(c)
	classID, _ := strconv.Atoi(c.Param("id"))
	classroom, _ := h.Store.GetClassroom(c.Request.Context(), classID)
	liveSession, _ := h.Store.GetActiveLiveSession(c.Request.Context(), classID)

	if liveSession == nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/classroom/%d", classID))
		return
	}

	// Record attendance join
	h.Store.JoinLiveAttendance(c.Request.Context(), liveSession.ID, student.ID)

	identity := fmt.Sprintf("student-%d", student.ID)
	token, err := h.generateLiveKitToken(liveSession.RoomName, identity, student.Name, true)
	if err != nil {
		c.String(500, "Token error: %v", err)
		return
	}

	h.render(c, "student_live.html", gin.H{
		"Classroom":   classroom,
		"LiveSession": liveSession,
		"Token":       token,
		"LKUrl":       h.LKUrl,
		"Student":     student,
	})
}

// ─── Live Attendance API ────────────────────────────────

func (h *Handler) StudentLiveLeave(c *gin.Context) {
	student := middleware.GetStudent(c)
	if student == nil {
		c.JSON(401, gin.H{"error": "not authenticated"})
		return
	}
	classID, _ := strconv.Atoi(c.Param("id"))
	liveSession, _ := h.Store.GetActiveLiveSession(c.Request.Context(), classID)
	if liveSession != nil {
		h.Store.LeaveLiveAttendance(c.Request.Context(), liveSession.ID, student.ID)
	}
	c.JSON(200, gin.H{"ok": true})
}

// ─── Live Image Presenter Upload ────────────────────────

func (h *Handler) LiveImageUpload(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(400, gin.H{"error": "no file"})
		return
	}
	defer file.Close()

	// Validate file type (images only)
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".bmp": true}
	if !allowed[ext] {
		c.JSON(400, gin.H{"error": "only images allowed"})
		return
	}

	// Max 10MB
	if header.Size > 10<<20 {
		c.JSON(400, gin.H{"error": "file too large (max 10MB)"})
		return
	}

	fname := fmt.Sprintf("live_%d_%d%s", classID, time.Now().UnixMilli(), ext)
	filePath := filepath.Join("live", fname)
	fullPath := filepath.Join(h.UploadDir, filePath)
	os.MkdirAll(filepath.Dir(fullPath), 0755)

	dst, err := os.Create(fullPath)
	if err != nil {
		c.JSON(500, gin.H{"error": "save failed"})
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	c.JSON(200, gin.H{"url": "/uploads/" + filePath})
}

// ─── Teacher Profile Picture Upload ─────────────────────

func (h *Handler) TeacherPicUpload(c *gin.Context) {
	classID := h.ownsClassroom(c)
	if classID == 0 {
		return
	}

	file, header, err := c.Request.FormFile("pic")
	if err != nil {
		c.JSON(400, gin.H{"error": "no file"})
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowed[ext] {
		c.JSON(400, gin.H{"error": "only jpg/png/webp allowed"})
		return
	}
	if header.Size > 5<<20 {
		c.JSON(400, gin.H{"error": "file too large (max 5MB)"})
		return
	}

	fname := fmt.Sprintf("teacher_%d_%d%s", classID, time.Now().UnixMilli(), ext)
	filePath := filepath.Join("live", fname)
	fullPath := filepath.Join(h.UploadDir, filePath)
	os.MkdirAll(filepath.Dir(fullPath), 0755)

	dst, err := os.Create(fullPath)
	if err != nil {
		c.JSON(500, gin.H{"error": "save failed"})
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	// Save to DB
	h.Store.SetClassroomTeacherPic(c.Request.Context(), classID, adminID(c), "/uploads/"+filePath)

	c.JSON(200, gin.H{"url": "/uploads/" + filePath})
}
