package middleware

import (
	"context"
	"net/http"
	"strings"
)

// DeviceType represents the class of device making the request.
type DeviceType string

const (
	DeviceDesktop DeviceType = "desktop"
	DeviceMobile  DeviceType = "mobile"
	DeviceTablet  DeviceType = "tablet"
	DeviceBot     DeviceType = "bot"
	DeviceUnknown DeviceType = "unknown"
)

// BrandVersion represents a browser brand and its version from Client Hints.
type BrandVersion struct {
	Brand   string
	Version string
}

// UserAgentInfo holds parsed User-Agent data and Client Hints.
type UserAgentInfo struct {
	Raw            string
	BrowserName    string
	BrowserVersion string
	OSName         string
	OSVersion      string
	Device         DeviceType
	IsBot          bool
	BotName        string
	// Client Hints fields (populated when browser sends them)
	Brands          []BrandVersion
	Platform        string // overrides OSName when present
	PlatformVersion string // overrides OSVersion when present
	Mobile          bool   // from Sec-CH-UA-Mobile
	Model           string // device model
	Architecture    string // CPU architecture
	HintsAvailable  bool   // true if any Client Hints headers were present
}

type userAgentKey struct{}

// acceptCHValue is the value for the Accept-CH response header.
const acceptCHValue = "Sec-CH-UA, Sec-CH-UA-Mobile, Sec-CH-UA-Platform, Sec-CH-UA-Platform-Version, Sec-CH-UA-Model, Sec-CH-UA-Arch"

// UserAgentMiddleware parses the User-Agent header and Client Hints,
// stores the result in context, and sets Accept-CH to opt in to Client Hints.
func UserAgentMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info := parseUserAgent(r.Header.Get("User-Agent"))
			parseClientHints(r.Header, &info)
			w.Header().Set("Accept-CH", acceptCHValue)
			ctx := context.WithValue(r.Context(), userAgentKey{}, info)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserAgentInfoFromContext extracts UserAgentInfo from the context.
// Returns a zero-value UserAgentInfo if not present (safe, no nil checks).
func UserAgentInfoFromContext(ctx context.Context) UserAgentInfo {
	info, _ := ctx.Value(userAgentKey{}).(UserAgentInfo)
	return info
}

// parseUserAgent extracts browser, OS, device type, and bot info from a UA string.
func parseUserAgent(ua string) UserAgentInfo {
	info := UserAgentInfo{
		Raw:    ua,
		Device: DeviceUnknown,
	}

	if ua == "" {
		info.IsBot = true
		info.BotName = "unknown"
		info.Device = DeviceBot
		return info
	}

	// Bot detection (checked first, short-circuits)
	if detectBot(ua, &info) {
		info.Device = DeviceBot
		return info
	}

	detectBrowser(ua, &info)
	detectOS(ua, &info)
	detectDevice(ua, &info)

	return info
}

// botSignatures maps bot identifiers to their display names.
var botSignatures = []struct {
	pattern string
	name    string
}{
	{"Googlebot", "Googlebot"},
	{"bingbot", "Bingbot"},
	{"Baiduspider", "Baiduspider"},
	{"YandexBot", "YandexBot"},
	{"DuckDuckBot", "DuckDuckBot"},
	{"Slurp", "Yahoo! Slurp"},
	{"facebookexternalhit", "Facebook"},
	{"Twitterbot", "Twitterbot"},
	{"LinkedInBot", "LinkedInBot"},
}

// botGenericPatterns are lowercased substrings that indicate a generic bot.
var botGenericPatterns = []string{"bot", "crawler", "spider"}

func detectBot(ua string, info *UserAgentInfo) bool {
	for _, sig := range botSignatures {
		if strings.Contains(ua, sig.pattern) {
			info.IsBot = true
			info.BotName = sig.name
			return true
		}
	}
	lower := strings.ToLower(ua)
	for _, pat := range botGenericPatterns {
		if strings.Contains(lower, pat) {
			info.IsBot = true
			info.BotName = "generic"
			return true
		}
	}
	return false
}

func detectBrowser(ua string, info *UserAgentInfo) {
	// Order matters: Edge and Opera include "Chrome" in their UA strings.
	switch {
	case strings.Contains(ua, "Edg/") || strings.Contains(ua, "Edge/"):
		info.BrowserName = "Edge"
		if v := extractVersionAfter(ua, "Edg/"); v != "" {
			info.BrowserVersion = v
		} else {
			info.BrowserVersion = extractVersionAfter(ua, "Edge/")
		}
	case strings.Contains(ua, "OPR/") || strings.Contains(ua, "Opera/"):
		info.BrowserName = "Opera"
		if v := extractVersionAfter(ua, "OPR/"); v != "" {
			info.BrowserVersion = v
		} else {
			info.BrowserVersion = extractVersionAfter(ua, "Opera/")
		}
	case strings.Contains(ua, "Chrome/"):
		info.BrowserName = "Chrome"
		info.BrowserVersion = extractVersionAfter(ua, "Chrome/")
	case strings.Contains(ua, "Chromium/"):
		info.BrowserName = "Chromium"
		info.BrowserVersion = extractVersionAfter(ua, "Chromium/")
	case strings.Contains(ua, "Safari/") && strings.Contains(ua, "Version/"):
		info.BrowserName = "Safari"
		info.BrowserVersion = extractVersionAfter(ua, "Version/")
	case strings.Contains(ua, "Firefox/"):
		info.BrowserName = "Firefox"
		info.BrowserVersion = extractVersionAfter(ua, "Firefox/")
	}
}

func detectOS(ua string, info *UserAgentInfo) {
	switch {
	case strings.Contains(ua, "iPhone OS") || strings.Contains(ua, "iPad") || strings.Contains(ua, "iPod"):
		// iOS must be checked before macOS because iOS UAs contain "Mac OS X"
		info.OSName = "iOS"
		if raw := extractVersionAfter(ua, "iPhone OS "); raw != "" {
			info.OSVersion = strings.ReplaceAll(raw, "_", ".")
		} else if raw := extractVersionAfter(ua, "CPU OS "); raw != "" {
			// iPad uses "CPU OS 17_2" instead of "iPhone OS 17_2"
			info.OSVersion = strings.ReplaceAll(raw, "_", ".")
		}
	case strings.Contains(ua, "Windows NT"):
		info.OSName = "Windows"
		info.OSVersion = extractVersionAfter(ua, "Windows NT ")
	case strings.Contains(ua, "Android"):
		info.OSName = "Android"
		info.OSVersion = extractVersionAfter(ua, "Android ")
	case strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Mac OS X"):
		info.OSName = "macOS"
		raw := extractVersionAfter(ua, "Mac OS X ")
		// macOS versions use underscores in UA strings: 10_15_7 → 10.15.7
		info.OSVersion = strings.ReplaceAll(raw, "_", ".")
	case strings.Contains(ua, "CrOS"):
		info.OSName = "ChromeOS"
	case strings.Contains(ua, "Linux"):
		info.OSName = "Linux"
	}
}

func detectDevice(ua string, info *UserAgentInfo) {
	switch {
	case strings.Contains(ua, "iPad") || strings.Contains(ua, "Tablet"):
		info.Device = DeviceTablet
	case strings.Contains(ua, "Mobile") || strings.Contains(ua, "iPhone") ||
		(strings.Contains(ua, "Android") && !strings.Contains(ua, "Tablet")):
		info.Device = DeviceMobile
	default:
		info.Device = DeviceDesktop
	}
}

// extractVersionAfter finds prefix in ua and scans the version string that follows.
// A version consists of digits, dots, and underscores (e.g. "10_15_7", "131.0.6778.86").
// Returns empty string if prefix is not found.
func extractVersionAfter(ua string, prefix string) string {
	idx := strings.Index(ua, prefix)
	if idx < 0 {
		return ""
	}
	start := idx + len(prefix)
	end := start
	for end < len(ua) {
		c := ua[end]
		if (c >= '0' && c <= '9') || c == '.' || c == '_' {
			end++
			continue
		}
		break
	}
	// Trim trailing dots/underscores
	for end > start && (ua[end-1] == '.' || ua[end-1] == '_') {
		end--
	}
	return ua[start:end]
}

// parseClientHints reads Client Hints headers and updates info accordingly.
func parseClientHints(h http.Header, info *UserAgentInfo) {
	secUA := h.Get("Sec-CH-UA")
	mobile := h.Get("Sec-CH-UA-Mobile")
	platform := h.Get("Sec-CH-UA-Platform")
	platformVer := h.Get("Sec-CH-UA-Platform-Version")
	model := h.Get("Sec-CH-UA-Model")
	arch := h.Get("Sec-CH-UA-Arch")

	if secUA == "" && mobile == "" && platform == "" && platformVer == "" && model == "" && arch == "" {
		return
	}
	info.HintsAvailable = true

	if secUA != "" {
		info.Brands = parseSecCHUA(secUA)
	}
	if mobile == "?1" {
		info.Mobile = true
	}
	if platform != "" {
		info.Platform = unquoteHeader(platform)
	}
	if platformVer != "" {
		info.PlatformVersion = unquoteHeader(platformVer)
	}
	if model != "" {
		info.Model = unquoteHeader(model)
	}
	if arch != "" {
		info.Architecture = unquoteHeader(arch)
	}
}

// parseSecCHUA parses a Sec-CH-UA structured header into brand-version pairs.
// Filters out GREASE entries that contain "Not" and "Brand" together.
// Example input: `"Chromium";v="131", "Google Chrome";v="131", "Not_A Brand";v="24"`
func parseSecCHUA(header string) []BrandVersion {
	var brands []BrandVersion
	for _, entry := range strings.Split(header, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// Split on ;v=
		parts := strings.SplitN(entry, ";v=", 2)
		if len(parts) != 2 {
			continue
		}
		brand := unquoteHeader(strings.TrimSpace(parts[0]))
		version := unquoteHeader(strings.TrimSpace(parts[1]))

		// Skip GREASE entries: brands containing both "Not" and "Brand"
		if strings.Contains(brand, "Not") && strings.Contains(brand, "Brand") {
			continue
		}
		brands = append(brands, BrandVersion{Brand: brand, Version: version})
	}
	return brands
}

// unquoteHeader removes surrounding double quotes from a header value.
func unquoteHeader(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
