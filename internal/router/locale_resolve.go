package router

import (
	"net/http"
	"strings"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ResolveLocale determines the locale for a delivery request.
// Priority: ?locale= query param > Accept-Language header > default locale.
// Returns "" when i18n is disabled.
func ResolveLocale(r *http.Request, cfg *config.Config, d db.DbDriver) string {
	if !cfg.I18nEnabled() {
		return ""
	}

	// 1. Explicit query parameter takes highest priority.
	if qLocale := r.URL.Query().Get("locale"); qLocale != "" {
		return qLocale
	}

	// 2. Parse Accept-Language header and match against enabled locales.
	enabledLocales, err := d.ListEnabledLocales()
	if err != nil || enabledLocales == nil || len(*enabledLocales) == 0 {
		return cfg.I18nDefaultLocale()
	}

	// Build a set of enabled locale codes for O(1) lookup.
	enabled := make(map[string]struct{}, len(*enabledLocales))
	for _, loc := range *enabledLocales {
		enabled[loc.Code] = struct{}{}
	}

	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang != "" {
		// Parse the Accept-Language header and find the first match.
		parsed := parseAcceptLanguage(acceptLang)
		for _, lang := range parsed {
			if _, ok := enabled[lang]; ok {
				return lang
			}
			// Try base language (e.g., "en" from "en-US").
			if idx := strings.IndexByte(lang, '-'); idx > 0 {
				base := lang[:idx]
				if _, ok := enabled[base]; ok {
					return base
				}
			}
		}
	}

	// 3. Fall back to default locale.
	return cfg.I18nDefaultLocale()
}

// WalkFallback tries to find a published snapshot by walking the locale
// fallback chain. Max 2 hops. Returns the resolved locale and snapshot
// version, or empty string and nil if none found.
func WalkFallback(d db.DbDriver, contentDataID types.ContentID, requestedLocale string) (string, *db.ContentVersion, error) {
	// Try the requested locale first.
	version, err := d.GetPublishedSnapshot(contentDataID, requestedLocale)
	if err == nil {
		return requestedLocale, version, nil
	}

	// Walk the fallback chain (max 2 hops).
	currentLocale := requestedLocale
	for hop := 0; hop < 2; hop++ {
		locale, locErr := d.GetLocaleByCode(currentLocale)
		if locErr != nil {
			break
		}
		if locale.FallbackCode == "" {
			break
		}

		version, err = d.GetPublishedSnapshot(contentDataID, locale.FallbackCode)
		if err == nil {
			return locale.FallbackCode, version, nil
		}
		currentLocale = locale.FallbackCode
	}

	return "", nil, nil
}

// LocaleMetadata holds lightweight locale availability info for ?locale=* responses.
type LocaleMetadata struct {
	Locales          map[string]LocaleVersionInfo `json:"locales"`
	AvailableLocales []string                     `json:"available_locales"`
}

// LocaleVersionInfo holds per-locale version metadata.
type LocaleVersionInfo struct {
	Published   bool   `json:"published"`
	PublishedAt string `json:"published_at,omitempty"`
}

// BuildLocaleMetadata builds a metadata response showing which locales
// have published snapshots for the given content data.
func BuildLocaleMetadata(d db.DbDriver, contentDataID types.ContentID) (*LocaleMetadata, error) {
	enabledLocales, err := d.ListEnabledLocales()
	if err != nil {
		return nil, err
	}

	meta := &LocaleMetadata{
		Locales:          make(map[string]LocaleVersionInfo),
		AvailableLocales: make([]string, 0),
	}

	if enabledLocales == nil {
		return meta, nil
	}

	for _, loc := range *enabledLocales {
		version, verErr := d.GetPublishedSnapshot(contentDataID, loc.Code)
		if verErr == nil && version != nil {
			meta.Locales[loc.Code] = LocaleVersionInfo{
				Published:   true,
				PublishedAt: version.DateCreated.String(),
			}
			meta.AvailableLocales = append(meta.AvailableLocales, loc.Code)
		} else {
			meta.Locales[loc.Code] = LocaleVersionInfo{Published: false}
		}
	}

	// Also check the empty-locale snapshot (pre-i18n content).
	version, verErr := d.GetPublishedSnapshot(contentDataID, "")
	if verErr == nil && version != nil {
		meta.Locales[""] = LocaleVersionInfo{
			Published:   true,
			PublishedAt: version.DateCreated.String(),
		}
	}

	return meta, nil
}

// parseAcceptLanguage parses the Accept-Language header into an ordered list
// of language tags sorted by quality (highest first). Does not handle quality
// weighting beyond preserving the original order — the first matching language
// in the header wins.
func parseAcceptLanguage(header string) []string {
	// Lightweight parser: split on comma, trim, extract language tag before ";".
	// Quality values are ignored — we use position order (first = highest priority).
	parts := strings.Split(header, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "*" {
			continue
		}
		// Strip quality value: "en-US;q=0.9" -> "en-US"
		if idx := strings.IndexByte(part, ';'); idx > 0 {
			part = strings.TrimSpace(part[:idx])
		}
		if part != "" && part != "*" {
			result = append(result, strings.ToLower(part))
		}
	}
	return result
}
