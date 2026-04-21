package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"teachhub/geo"
	"time"

	"github.com/gin-gonic/gin"
)

// In-memory cache for IP-API lookups (TTL 24h, capped at 10k entries).
type ipCacheEntry struct {
	country string
	expires time.Time
}

var (
	ipCacheMu sync.RWMutex
	ipCache   = make(map[string]ipCacheEntry, 1024)
)

var ipApiClient = &http.Client{Timeout: 800 * time.Millisecond}

// lookupCountryAPI queries ip-api.com (free, no key, ~45 req/min per IP).
// Returns "" on failure or for private/loopback addresses. Cached for 24h.
func lookupCountryAPI(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsUnspecified() {
		return ""
	}

	ipCacheMu.RLock()
	if e, ok := ipCache[ip]; ok && time.Now().Before(e.expires) {
		ipCacheMu.RUnlock()
		return e.country
	}
	ipCacheMu.RUnlock()

	resp, err := ipApiClient.Get("http://ip-api.com/json/" + ip + "?fields=countryCode")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var body struct {
		CountryCode string `json:"countryCode"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return ""
	}

	country := strings.ToUpper(strings.TrimSpace(body.CountryCode))

	ipCacheMu.Lock()
	if len(ipCache) > 10000 {
		ipCache = make(map[string]ipCacheEntry, 1024)
	}
	ipCache[ip] = ipCacheEntry{country: country, expires: time.Now().Add(24 * time.Hour)}
	ipCacheMu.Unlock()

	return country
}

// inferCountry resolves the best country guess from request context.
// Order: local GeoLite DB -> ip-api.com -> language hint -> DZ fallback.
func inferCountry(c *gin.Context) string {
	ip := c.ClientIP()

	if country := strings.ToUpper(strings.TrimSpace(geo.CountryFromIP(ip))); country != "" {
		return country
	}

	if country := lookupCountryAPI(ip); country != "" {
		return country
	}

	if lang, _ := c.Cookie("lang"); strings.HasPrefix(strings.ToLower(lang), "fr") {
		return "FR"
	}
	if strings.HasPrefix(strings.ToLower(c.GetHeader("Accept-Language")), "fr") {
		return "FR"
	}

	return "DZ"
}
