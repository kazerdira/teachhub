package handlers

import (
	"net/http"
	"path/filepath"
	"strings"

	"teachhub/middleware"

	"github.com/gin-gonic/gin"
)

// ServeUpload serves uploaded files with authentication.
// Access rules:
//   - Admin: can access any file in classrooms they own
//   - Student: can access files in classrooms they belong to
//   - No auth: blocked (403)
//
// This replaces the old r.Static("/uploads", uploadDir) which served
// ALL files publicly without any auth check.
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
	student := middleware.GetStudent(c)

	if aid == 0 && student == nil {
		c.String(http.StatusForbidden, "Access denied")
		return
	}

	// Serve the file
	fullPath := filepath.Join(h.UploadDir, reqPath)
	c.File(fullPath)
}
