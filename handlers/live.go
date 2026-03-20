package handlers

import (
	"fmt"
	"net/http"
	"strconv"
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
