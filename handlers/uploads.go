package handlers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"teachhub/middleware"

	"github.com/gin-gonic/gin"
)

// ServeUpload serves uploaded files with authentication and scoping.
// Access rules:
//   - Admin: can access any file (classroom ownership checked elsewhere)
//   - Student: can only access resources in classrooms they belong to,
//     and submissions that belong to them
//   - No auth: blocked (403)
func (h *Handler) ServeUpload(c *gin.Context) {
	reqPath := c.Param("filepath")

	// Prevent directory traversal
	reqPath = filepath.Clean(reqPath)
	if strings.Contains(reqPath, "..") {
		c.String(http.StatusForbidden, "Access denied")
		return
	}

	// Must be logged in as either admin or student
	aid := adminID(c)
	// Also check admin session directly (since AdminRequired middleware is not on this route)
	if aid == 0 {
		session, _ := middleware.SessionStore.Get(c.Request, "teachhub-admin")
		if session.Values["admin_id"] != nil {
			aid = session.Values["admin_id"].(int)
		}
	}
	student := middleware.GetStudent(c)

	if aid == 0 && student == nil {
		c.String(http.StatusForbidden, "Access denied")
		return
	}

	// For students, scope access to their own classrooms/submissions
	if student != nil && aid == 0 {
		parts := strings.SplitN(filepath.ToSlash(reqPath), "/", 2)
		if len(parts) == 2 {
			folder := parts[0] // "resources", "submissions", "live"
			fname := parts[1]

			switch folder {
			case "resources":
				// Filename: <classID>_<timestamp>.<ext>
				if idx := strings.Index(fname, "_"); idx > 0 {
					if classID, err := strconv.Atoi(fname[:idx]); err == nil {
						ok, _ := h.Store.IsStudentInClassroom(c.Request.Context(), student.ID, classID)
						if !ok {
							c.String(http.StatusForbidden, "Access denied")
							return
						}
					}
				}
			case "submissions":
				// Regular: <assignID>_<studentID>_<ts>.<ext>
				// Quiz:    quiz_<quizID>_q<qID>_s<studentID>_<ts>.<ext>
				if strings.HasPrefix(fname, "quiz_") {
					// Extract studentID after "_s"
					if idx := strings.Index(fname, "_s"); idx > 0 {
						rest := fname[idx+2:]
						if uidx := strings.Index(rest, "_"); uidx > 0 {
							if ownerID, err := strconv.Atoi(rest[:uidx]); err == nil && ownerID != student.ID {
								c.String(http.StatusForbidden, "Access denied")
								return
							}
						}
					}
				} else {
					// Extract studentID (second segment)
					segs := strings.SplitN(fname, "_", 3)
					if len(segs) >= 2 {
						if ownerID, err := strconv.Atoi(segs[1]); err == nil && ownerID != student.ID {
							c.String(http.StatusForbidden, "Access denied")
							return
						}
					}
				}
			}
		}
	}

	// Serve the file
	fullPath := filepath.Join(h.UploadDir, reqPath)
	c.File(fullPath)
}
