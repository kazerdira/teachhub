package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/gob"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"teachhub/store"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

var SessionStore *sessions.CookieStore

func Init(secret string, isProduction bool) {
	gob.Register(0)  // register int type for session storage
	gob.Register("") // register string type for CSRF token
	SessionStore = sessions.NewCookieStore([]byte(secret))
	SessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
	}
}

// ─── Admin Auth ─────────────────────────────────────────

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, _ := SessionStore.Get(c.Request, "teachhub-admin")
		if session.Values["admin_id"] == nil {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}
		c.Set("admin_id", session.Values["admin_id"])
		c.Next()
	}
}

// AdminSubscriptionCheck verifies subscription on every request for platform-created teachers.
func AdminSubscriptionCheck(db *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminID, _ := c.Get("admin_id")
		id, ok := adminID.(int)
		if !ok {
			c.Next()
			return
		}
		admin, err := db.GetAdminByID(c.Request.Context(), id)
		if err != nil {
			ClearAdminSession(c)
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}
		if admin.CreatedByPlatform {
			// Auto-expire if subscription_end has passed
			if admin.SubscriptionEnd != nil && admin.SubscriptionEnd.Before(time.Now()) && admin.SubscriptionStatus == "active" {
				db.UpdateTeacherSubscription(c.Request.Context(), admin.ID, "expired")
				admin.SubscriptionStatus = "expired"
			}
			if admin.SubscriptionStatus != "active" {
				ClearAdminSession(c)
				errKey := "suspended"
				if admin.SubscriptionStatus == "expired" {
					errKey = "expired"
				}
				c.Redirect(http.StatusFound, "/admin/login?error="+errKey)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func SetAdminSession(c *gin.Context, adminID int) error {
	session, _ := SessionStore.Get(c.Request, "teachhub-admin")
	session.Values["admin_id"] = adminID
	return session.Save(c.Request, c.Writer)
}

func ClearAdminSession(c *gin.Context) error {
	session, _ := SessionStore.Get(c.Request, "teachhub-admin")
	session.Options.MaxAge = -1
	return session.Save(c.Request, c.Writer)
}

// ─── Student Session ────────────────────────────────────

func StudentFromSession(db *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, _ := SessionStore.Get(c.Request, "teachhub-student")
		if sid, ok := session.Values["student_id"].(int); ok {
			student, err := db.GetStudent(c.Request.Context(), sid)
			if err == nil {
				c.Set("student", student)
				c.Set("student_id", sid)
			}
		}
		c.Next()
	}
}

func StudentRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("student"); !exists {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}

func SetStudentSession(c *gin.Context, studentID int) error {
	session, _ := SessionStore.Get(c.Request, "teachhub-student")
	session.Values["student_id"] = studentID
	return session.Save(c.Request, c.Writer)
}

func GetStudent(c *gin.Context) *store.Student {
	if s, exists := c.Get("student"); exists {
		return s.(*store.Student)
	}
	return nil
}

// ─── Platform Admin Auth ────────────────────────────────

func PlatformAdminRequired(platformPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, _ := SessionStore.Get(c.Request, "teachhub-platform")
		if session.Values["platform_admin_id"] == nil {
			c.Redirect(http.StatusFound, platformPath+"/login")
			c.Abort()
			return
		}
		c.Set("platform_admin_id", session.Values["platform_admin_id"])
		c.Next()
	}
}

func SetPlatformSession(c *gin.Context, adminID int) error {
	session, _ := SessionStore.Get(c.Request, "teachhub-platform")
	session.Values["platform_admin_id"] = adminID
	return session.Save(c.Request, c.Writer)
}

func ClearPlatformSession(c *gin.Context) error {
	session, _ := SessionStore.Get(c.Request, "teachhub-platform")
	session.Options.MaxAge = -1
	return session.Save(c.Request, c.Writer)
}

func GetPlatformAdminID(c *gin.Context) int {
	if id, exists := c.Get("platform_admin_id"); exists {
		if intID, ok := id.(int); ok {
			return intID
		}
	}
	return 0
}

// ─── CSRF Protection ───────────────────────────────────

func generateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// CSRFProtection generates a CSRF token on every request and validates it on POST.
// The token is stored in a dedicated "teachhub-csrf" session cookie.
func CSRFProtection() gin.HandlerFunc {
	// Routes that are exempt from CSRF (e.g. sendBeacon, AJAX uploads with session auth)
	exempt := map[string]bool{
		"/live/leave": true,
		"/live/image": true,
	}

	return func(c *gin.Context) {
		session, _ := SessionStore.Get(c.Request, "teachhub-csrf")

		token, ok := session.Values["csrf_token"].(string)
		if !ok || token == "" {
			token = generateCSRFToken()
			session.Values["csrf_token"] = token
			session.Save(c.Request, c.Writer)
		}

		c.Set("csrf_token", token)

		if c.Request.Method == "POST" {
			// Check if path ends with an exempt suffix
			skip := false
			for suffix := range exempt {
				if strings.HasSuffix(c.Request.URL.Path, suffix) {
					skip = true
					break
				}
			}
			if !skip {
				formToken := c.PostForm("_csrf")
				if formToken == "" {
					formToken = c.GetHeader("X-CSRF-Token")
				}
				if subtle.ConstantTimeCompare([]byte(formToken), []byte(token)) != 1 {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			}
		}

		c.Next()
	}
}

func GetCSRFToken(c *gin.Context) string {
	if t, exists := c.Get("csrf_token"); exists {
		if s, ok := t.(string); ok {
			return s
		}
	}
	return ""
}

// ─── Rate Limiting ──────────────────────────────────────

type rateLimitEntry struct {
	attempts  int
	blockedAt time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*rateLimitEntry
	max      int           // max attempts before blocking
	window   time.Duration // how long to block
}

func NewRateLimiter(maxAttempts int, blockDuration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		attempts: make(map[string]*rateLimitEntry),
		max:      maxAttempts,
		window:   blockDuration,
	}
	// Cleanup old entries every 10 minutes
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			rl.mu.Lock()
			now := time.Now()
			for ip, entry := range rl.attempts {
				if !entry.blockedAt.IsZero() && now.Sub(entry.blockedAt) > rl.window {
					delete(rl.attempts, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

// RateLimit returns a middleware that blocks an IP after maxAttempts failed POSTs
// within the window. GET requests always pass through (to show the login page).
func RateLimit(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only rate-limit POST (the actual login attempt)
		if c.Request.Method != "POST" {
			c.Next()
			return
		}

		ip := c.ClientIP()
		rl.mu.Lock()
		entry, exists := rl.attempts[ip]
		if !exists {
			entry = &rateLimitEntry{}
			rl.attempts[ip] = entry
		}

		// If currently blocked, check if window has passed
		if !entry.blockedAt.IsZero() {
			if time.Since(entry.blockedAt) < rl.window {
				rl.mu.Unlock()
				remaining := rl.window - time.Since(entry.blockedAt)
				mins := int(remaining.Minutes()) + 1
				c.HTML(http.StatusTooManyRequests, "error_rate_limit.html", gin.H{
					"Minutes": mins,
				})
				c.Abort()
				return
			}
			// Window passed, reset
			entry.attempts = 0
			entry.blockedAt = time.Time{}
		}

		entry.attempts++
		if entry.attempts >= rl.max {
			entry.blockedAt = time.Now()
		}
		rl.mu.Unlock()

		c.Next()

		// If login succeeded (redirect to dashboard, not back to login), reset counter
		if c.Writer.Status() == http.StatusFound {
			location := c.Writer.Header().Get("Location")
			if !strings.Contains(location, "login") && !strings.Contains(location, "error") {
				rl.mu.Lock()
				delete(rl.attempts, ip)
				rl.mu.Unlock()
			}
		}
	}
}

// ResetRateLimit clears rate limit for an IP (call on successful login)
func (rl *RateLimiter) Reset(ip string) {
	rl.mu.Lock()
	delete(rl.attempts, ip)
	rl.mu.Unlock()
}

// ─── Security Headers ──────────────────────────────────

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "camera=(self), microphone=(self), geolocation=()")
		// HSTS only makes sense behind HTTPS — reverse proxy will handle it,
		// but we set it here so it's ready when deployed behind TLS.
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}
