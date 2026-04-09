package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"teachhub/middleware"
	"teachhub/store"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
	middleware.Init("test-secret-for-handlers", false)
}

// newTestHandler creates a Handler with a temporary upload directory.
func newTestHandler(t *testing.T) (*Handler, string) {
	t.Helper()
	uploadDir := t.TempDir()
	h := &Handler{
		UploadDir: uploadDir,
	}
	return h, uploadDir
}

// setAdmin injects admin_id into the Gin context via middleware.
func setAdmin(adminID int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("admin_id", adminID)
		c.Next()
	}
}

// setStudent injects a student into the Gin context.
func setStudent(student *store.Student) gin.HandlerFunc {
	return func(c *gin.Context) {
		if student != nil {
			c.Set("student", student)
			c.Set("student_id", student.ID)
		}
		c.Next()
	}
}

// ─── Upload Auth Tests ─────────────────────────────────

func TestServeUpload_NoAuth_Forbidden(t *testing.T) {
	h, uploadDir := newTestHandler(t)

	// Create a test file
	os.WriteFile(filepath.Join(uploadDir, "test.txt"), []byte("secret"), 0644)

	r := gin.New()
	r.GET("/uploads/*filepath", h.ServeUpload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/uploads/test.txt", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("unauthenticated request should be 403, got %d", w.Code)
	}
	if w.Body.String() == "secret" {
		t.Error("SECURITY: file content was served without auth!")
	}
}

func TestServeUpload_AdminAuth_Allowed(t *testing.T) {
	h, uploadDir := newTestHandler(t)

	os.WriteFile(filepath.Join(uploadDir, "doc.pdf"), []byte("pdf-content"), 0644)

	r := gin.New()
	r.Use(setAdmin(1))
	r.GET("/uploads/*filepath", h.ServeUpload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/uploads/doc.pdf", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("admin should access uploads, got %d", w.Code)
	}
}

func TestServeUpload_StudentAuth_Allowed(t *testing.T) {
	h, uploadDir := newTestHandler(t)

	os.WriteFile(filepath.Join(uploadDir, "image.png"), []byte("png-data"), 0644)

	r := gin.New()
	r.Use(setStudent(&store.Student{ID: 42, Name: "Test Student"}))
	r.GET("/uploads/*filepath", h.ServeUpload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/uploads/image.png", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("student should access uploads, got %d", w.Code)
	}
}

func TestServeUpload_DirectoryTraversal_Blocked(t *testing.T) {
	h, _ := newTestHandler(t)

	r := gin.New()
	r.Use(setAdmin(1))
	r.GET("/uploads/*filepath", h.ServeUpload)

	tests := []struct {
		name string
		path string
	}{
		{"dot-dot-slash", "/uploads/../../../etc/passwd"},
		{"encoded-dots", "/uploads/..%2F..%2Fetc/passwd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			r.ServeHTTP(w, req)

			// Should be either 403 (blocked) or 404 (cleaned path doesn't exist), never 200
			if w.Code == http.StatusOK {
				t.Errorf("SECURITY: directory traversal via %s returned 200!", tt.path)
			}
		})
	}
}

func TestServeUpload_NonexistentFile_404(t *testing.T) {
	h, _ := newTestHandler(t)

	r := gin.New()
	r.Use(setAdmin(1))
	r.GET("/uploads/*filepath", h.ServeUpload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/uploads/does-not-exist.txt", nil)
	r.ServeHTTP(w, req)

	// Gin's c.File returns 404 for missing files
	if w.Code == http.StatusOK {
		t.Errorf("nonexistent file should not return 200, got %d", w.Code)
	}
}

func TestServeUpload_NestedPath_Works(t *testing.T) {
	h, uploadDir := newTestHandler(t)

	// Create nested directory structure
	nested := filepath.Join(uploadDir, "classrooms", "1", "assignments")
	os.MkdirAll(nested, 0755)
	os.WriteFile(filepath.Join(nested, "homework.pdf"), []byte("hw"), 0644)

	r := gin.New()
	r.Use(setAdmin(1))
	r.GET("/uploads/*filepath", h.ServeUpload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/uploads/classrooms/1/assignments/homework.pdf", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("nested path should work for admin, got %d", w.Code)
	}
}
