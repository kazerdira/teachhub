package handlers

import (
	"strings"
	"teachhub/geo"

	"github.com/gin-gonic/gin"
)

// inferCountry resolves the best country guess from request context.
// Order: IP geolocation -> language hint -> DZ fallback.
func inferCountry(c *gin.Context) string {
	if country := strings.ToUpper(strings.TrimSpace(geo.CountryFromIP(c.ClientIP()))); country != "" {
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
