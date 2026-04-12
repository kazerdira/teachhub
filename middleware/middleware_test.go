package middleware

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
	// Initialize session store with a test secret
	Init("test-secret-key-for-unit-tests", false)
}

// rateLimitRouter creates a Gin engine with a minimal template for rate-limit error rendering.
func rateLimitRouter() *gin.Engine {
	r := gin.New()
	tmpl := template.Must(template.New("error_rate_limit.html").Parse(`Too many requests. Wait {{.Minutes}} min.`))
	r.SetHTMLTemplate(tmpl)
	return r
}

// ─── CSRF Tests ─────────────────────────────────────────

func TestCSRF_GetRequest_AlwaysAllowed(t *testing.T) {
	r := gin.New()
	r.Use(CSRFProtection())
	r.GET("/some-page", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/some-page", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET request should pass CSRF, got status %d", w.Code)
	}
}

func TestCSRF_PostWithoutToken_Blocked(t *testing.T) {
	r := gin.New()
	r.Use(CSRFProtection())
	r.POST("/submit", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/submit", strings.NewReader("data=value"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("POST without CSRF token should be 403, got %d", w.Code)
	}
}

func TestCSRF_PostWithValidToken_Allowed(t *testing.T) {
	r := gin.New()
	r.Use(CSRFProtection())

	var csrfToken string
	r.GET("/get-token", func(c *gin.Context) {
		csrfToken = GetCSRFToken(c)
		c.String(http.StatusOK, csrfToken)
	})
	r.POST("/submit", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Step 1: GET to obtain a CSRF token and session cookie
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/get-token", nil)
	r.ServeHTTP(w1, req1)

	cookies := w1.Result().Cookies()

	// Step 2: POST with valid token and session cookie
	body := "_csrf=" + csrfToken
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/submit", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("POST with valid CSRF token should be 200, got %d", w2.Code)
	}
}

func TestCSRF_PostWithHeaderToken_Allowed(t *testing.T) {
	r := gin.New()
	r.Use(CSRFProtection())

	var csrfToken string
	r.GET("/get-token", func(c *gin.Context) {
		csrfToken = GetCSRFToken(c)
		c.String(http.StatusOK, csrfToken)
	})
	r.POST("/submit", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/get-token", nil)
	r.ServeHTTP(w1, req1)
	cookies := w1.Result().Cookies()

	// Use X-CSRF-Token header instead of form field
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/submit", nil)
	req2.Header.Set("X-CSRF-Token", csrfToken)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("POST with X-CSRF-Token header should be 200, got %d", w2.Code)
	}
}

func TestCSRF_PostWithWrongToken_Blocked(t *testing.T) {
	r := gin.New()
	r.Use(CSRFProtection())

	r.GET("/get-token", func(c *gin.Context) {
		c.String(http.StatusOK, GetCSRFToken(c))
	})
	r.POST("/submit", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Get a session cookie
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/get-token", nil)
	r.ServeHTTP(w1, req1)
	cookies := w1.Result().Cookies()

	// POST with wrong token
	body := "_csrf=totally-wrong-token"
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/submit", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusForbidden {
		t.Errorf("POST with wrong CSRF token should be 403, got %d", w2.Code)
	}
}

// ─── CSRF Exempt Path Tests ─────────────────────────────

func TestCSRF_ExemptPath_ClassroomLiveLeave(t *testing.T) {
	r := gin.New()
	r.Use(CSRFProtection())
	r.POST("/classroom/:id/live/leave", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/classroom/5/live/leave", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/classroom/5/live/leave should be CSRF exempt, got %d", w.Code)
	}
}

func TestCSRF_ExemptPath_AdminClassroomLiveImage(t *testing.T) {
	r := gin.New()
	r.Use(CSRFProtection())
	r.POST("/admin/classroom/:id/live/image", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/admin/classroom/3/live/image", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/admin/classroom/3/live/image should be CSRF exempt, got %d", w.Code)
	}
}

func TestCSRF_ExemptPath_AdminClassroomLivePdf(t *testing.T) {
	r := gin.New()
	r.Use(CSRFProtection())
	r.POST("/admin/classroom/:id/live/pdf", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/admin/classroom/7/live/pdf", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("/admin/classroom/7/live/pdf should be CSRF exempt, got %d", w.Code)
	}
}

// ─── CSRF Bypass Prevention ─────────────────────────────

func TestCSRF_CraftedPath_EvilLiveLeave_Blocked(t *testing.T) {
	r := gin.New()
	r.Use(CSRFProtection())
	r.POST("/evil/live/leave", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/evil/live/leave", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("SECURITY: /evil/live/leave should NOT bypass CSRF, got %d", w.Code)
	}
}

func TestCSRF_CraftedPath_RootLiveLeave_Blocked(t *testing.T) {
	r := gin.New()
	r.Use(CSRFProtection())
	r.POST("/live/leave", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/live/leave", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("SECURITY: /live/leave at root should NOT bypass CSRF, got %d", w.Code)
	}
}

func TestCSRF_CraftedPath_NestedEvil_Blocked(t *testing.T) {
	r := gin.New()
	r.Use(CSRFProtection())
	r.POST("/api/delete-account/live/leave", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/delete-account/live/leave", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("SECURITY: /api/delete-account/live/leave should NOT bypass CSRF, got %d", w.Code)
	}
}

// ─── Rate Limiter Tests ─────────────────────────────────

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(5, 1*time.Minute)

	r := rateLimitRouter()
	r.Use(RateLimit(rl))
	r.POST("/login", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 4; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d should be allowed (under limit), got %d", i+1, w.Code)
		}
	}
}

func TestRateLimiter_BlocksAtLimit(t *testing.T) {
	rl := NewRateLimiter(3, 1*time.Minute)

	r := rateLimitRouter()
	r.Use(RateLimit(rl))
	r.POST("/login", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Exhaust the limit
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "5.6.7.8:1234"
		r.ServeHTTP(w, req)
	}

	// Next request should be blocked
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "5.6.7.8:1234"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("request after exceeding limit should be 429, got %d", w.Code)
	}
}

func TestRateLimiter_GETPassesThrough(t *testing.T) {
	rl := NewRateLimiter(1, 1*time.Minute)

	r := rateLimitRouter()
	r.Use(RateLimit(rl))
	r.GET("/login", func(c *gin.Context) {
		c.String(http.StatusOK, "login page")
	})

	// GET should always pass, even after many requests
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/login", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("GET request %d should always pass rate limiter, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimiter_DifferentIPs_Independent(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Minute)

	r := rateLimitRouter()
	r.Use(RateLimit(rl))
	r.POST("/login", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// IP1: exhaust limit
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		r.ServeHTTP(w, req)
	}

	// IP1 should be blocked
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/login", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusTooManyRequests {
		t.Errorf("IP1 should be blocked, got %d", w1.Code)
	}

	// IP2 should still be fine
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/login", nil)
	req2.RemoteAddr = "10.0.0.2:5678"
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("IP2 should still be allowed, got %d", w2.Code)
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(2, 1*time.Minute)

	// Register some attempts
	r := rateLimitRouter()
	r.Use(RateLimit(rl))
	r.POST("/login", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "11.11.11.11:1234"
		r.ServeHTTP(w, req)
	}

	// Should be blocked
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "11.11.11.11:1234"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("should be blocked before reset, got %d", w.Code)
	}

	// Reset
	rl.Reset("11.11.11.11")

	// Should be allowed again
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/login", nil)
	req2.RemoteAddr = "11.11.11.11:1234"
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("should be allowed after reset, got %d", w2.Code)
	}
}

// ─── Security Headers Tests ────────────────────────────

func TestSecurityHeaders_SetCorrectly(t *testing.T) {
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	expected := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "SAMEORIGIN",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}
	for header, want := range expected {
		got := w.Header().Get(header)
		if got != want {
			t.Errorf("header %s: want %q, got %q", header, want, got)
		}
	}
}

func TestSecurityHeaders_HSTSWithHTTPS(t *testing.T) {
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	r.ServeHTTP(w, req)

	hsts := w.Header().Get("Strict-Transport-Security")
	if !strings.Contains(hsts, "max-age=") {
		t.Errorf("expected HSTS header with max-age when behind HTTPS, got %q", hsts)
	}
}

func TestSecurityHeaders_NoHSTSWithoutHTTPS(t *testing.T) {
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	hsts := w.Header().Get("Strict-Transport-Security")
	if hsts != "" {
		t.Errorf("expected no HSTS header on plain HTTP, got %q", hsts)
	}
}
