package geo

import (
	"log"
	"net"
	"strings"

	"github.com/oschwald/maxminddb-golang"
)

var db *maxminddb.Reader

// Init loads the GeoLite2-Country database.
// Call once at startup. If the file doesn't exist, geolocation will be disabled.
func Init(path string) {
	var err error
	db, err = maxminddb.Open(path)
	if err != nil {
		log.Printf("⚠️  GeoLite2 DB not found (%s) — country detection disabled", path)
		db = nil
	} else {
		log.Printf("🌍 GeoLite2 loaded: %s", path)
	}
}

// Close releases the MaxMind DB resources.
func Close() {
	if db != nil {
		db.Close()
	}
}

type geoResult struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

// CountryFromIP returns the ISO 3166-1 alpha-2 country code (e.g. "DZ", "FR")
// for the given IP string. Returns "" if unknown.
func CountryFromIP(ipStr string) string {
	if db == nil {
		return ""
	}

	// Strip port if present
	if strings.Contains(ipStr, ":") {
		host, _, err := net.SplitHostPort(ipStr)
		if err == nil {
			ipStr = host
		}
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}

	var result geoResult
	err := db.Lookup(ip, &result)
	if err != nil {
		return ""
	}

	return result.Country.ISOCode
}

// CurrencyForCountry returns the display currency for a country code.
func CurrencyForCountry(country string) string {
	switch country {
	case "FR":
		return "€"
	default:
		return "DA"
	}
}
