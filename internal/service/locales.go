package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// LocaleService manages locale CRUD, BCP 47 validation, fallback chain cycle
// detection, default locale atomicity, translation creation, and locale resolution.
type LocaleService struct {
	driver db.DbDriver
	mgr    *config.Manager
}

// NewLocaleService creates a LocaleService.
func NewLocaleService(driver db.DbDriver, mgr *config.Manager) *LocaleService {
	return &LocaleService{driver: driver, mgr: mgr}
}

// ---------------------------------------------------------------------------
// Param / result types
// ---------------------------------------------------------------------------

// CreateLocaleInput holds the caller-provided fields for creating a locale.
type CreateLocaleInput struct {
	Code         string `json:"code"`
	Label        string `json:"label"`
	IsDefault    bool   `json:"is_default"`
	IsEnabled    bool   `json:"is_enabled"`
	FallbackCode string `json:"fallback_code"`
	SortOrder    int64  `json:"sort_order"`
}

// UpdateLocaleInput holds the caller-provided fields for updating a locale.
type UpdateLocaleInput struct {
	LocaleID     types.LocaleID `json:"locale_id"`
	Code         string         `json:"code"`
	Label        string         `json:"label"`
	IsDefault    bool           `json:"is_default"`
	IsEnabled    bool           `json:"is_enabled"`
	FallbackCode string         `json:"fallback_code"`
	SortOrder    int64          `json:"sort_order"`
}

// TranslationResult is returned by CreateTranslation / CreateAdminTranslation.
type TranslationResult struct {
	Locale        string `json:"locale"`
	FieldsCreated int    `json:"fields_created"`
}

// LocaleVersionInfo holds per-locale version metadata.
type LocaleVersionInfo struct {
	Published   bool   `json:"published"`
	PublishedAt string `json:"published_at,omitempty"`
}

// LocaleMetadata holds lightweight locale availability info for ?locale=* responses.
type LocaleMetadata struct {
	Locales          map[string]LocaleVersionInfo `json:"locales"`
	AvailableLocales []string                     `json:"available_locales"`
}

// ---------------------------------------------------------------------------
// CRUD methods
// ---------------------------------------------------------------------------

// CreateLocale validates input, handles default locale atomicity, and creates
// a new locale.
func (s *LocaleService) CreateLocale(ctx context.Context, ac audited.AuditContext, input CreateLocaleInput) (*db.Locale, error) {
	if err := validateLocaleCode(input.Code); err != nil {
		return nil, err
	}

	if input.Label == "" {
		return nil, NewValidationError("label", "label must not be empty")
	}

	if input.FallbackCode != "" {
		if err := s.validateFallbackChain(input.FallbackCode, input.Code); err != nil {
			return nil, err
		}
	}

	if input.IsDefault {
		if err := s.driver.ClearDefaultLocale(ctx); err != nil {
			return nil, fmt.Errorf("clear default locale: %w", err)
		}
	}

	now := types.NewTimestamp(time.Now().UTC())
	created, err := s.driver.CreateLocale(ctx, ac, db.CreateLocaleParams{
		Code:         input.Code,
		Label:        input.Label,
		IsDefault:    input.IsDefault,
		IsEnabled:    input.IsEnabled,
		FallbackCode: input.FallbackCode,
		SortOrder:    input.SortOrder,
		DateCreated:  now,
	})
	if err != nil {
		// Check for unique constraint violation (duplicate code).
		errMsg := err.Error()
		if strings.Contains(errMsg, "UNIQUE") || strings.Contains(errMsg, "unique") || strings.Contains(errMsg, "Duplicate") {
			return nil, &ConflictError{Resource: "locale", ID: input.Code, Detail: "locale code already exists"}
		}
		return nil, fmt.Errorf("create locale: %w", err)
	}
	return created, nil
}

// UpdateLocale validates input, handles default locale atomicity, and updates
// an existing locale.
func (s *LocaleService) UpdateLocale(ctx context.Context, ac audited.AuditContext, input UpdateLocaleInput) (*db.Locale, error) {
	existing, err := s.driver.GetLocale(input.LocaleID)
	if err != nil {
		return nil, &NotFoundError{Resource: "locale", ID: string(input.LocaleID)}
	}

	if err := validateLocaleCode(input.Code); err != nil {
		return nil, err
	}

	if input.Label == "" {
		return nil, NewValidationError("label", "label must not be empty")
	}

	if input.FallbackCode != "" {
		if err := s.validateFallbackChain(input.FallbackCode, input.Code); err != nil {
			return nil, err
		}
	}

	if input.IsDefault && !existing.IsDefault {
		if err := s.driver.ClearDefaultLocale(ctx); err != nil {
			return nil, fmt.Errorf("clear default locale: %w", err)
		}
	}

	err = s.driver.UpdateLocale(ctx, ac, db.UpdateLocaleParams{
		LocaleID:     input.LocaleID,
		Code:         input.Code,
		Label:        input.Label,
		IsDefault:    input.IsDefault,
		IsEnabled:    input.IsEnabled,
		FallbackCode: input.FallbackCode,
		SortOrder:    input.SortOrder,
		DateCreated:  existing.DateCreated, // preserve original
	})
	if err != nil {
		return nil, fmt.Errorf("update locale: %w", err)
	}

	updated, err := s.driver.GetLocale(input.LocaleID)
	if err != nil {
		return nil, fmt.Errorf("fetch updated locale: %w", err)
	}
	return updated, nil
}

// DeleteLocale deletes a locale. Cannot delete the default locale.
func (s *LocaleService) DeleteLocale(ctx context.Context, ac audited.AuditContext, id types.LocaleID) error {
	existing, err := s.driver.GetLocale(id)
	if err != nil {
		return &NotFoundError{Resource: "locale", ID: string(id)}
	}
	if existing.IsDefault {
		return &ForbiddenError{Message: "cannot delete the default locale"}
	}
	if err := s.driver.DeleteLocale(ctx, ac, id); err != nil {
		return fmt.Errorf("delete locale: %w", err)
	}
	return nil
}

// GetLocale retrieves a locale by ID.
func (s *LocaleService) GetLocale(ctx context.Context, id types.LocaleID) (*db.Locale, error) {
	locale, err := s.driver.GetLocale(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "locale", ID: string(id)}
	}
	return locale, nil
}

// GetLocaleByCode retrieves a locale by its code.
func (s *LocaleService) GetLocaleByCode(ctx context.Context, code string) (*db.Locale, error) {
	locale, err := s.driver.GetLocaleByCode(code)
	if err != nil {
		return nil, &NotFoundError{Resource: "locale", ID: code}
	}
	return locale, nil
}

// GetDefaultLocale retrieves the default locale.
func (s *LocaleService) GetDefaultLocale(ctx context.Context) (*db.Locale, error) {
	locale, err := s.driver.GetDefaultLocale()
	if err != nil {
		return nil, &NotFoundError{Resource: "locale", ID: "default"}
	}
	return locale, nil
}

// ListLocales returns all locales (admin use).
func (s *LocaleService) ListLocales(ctx context.Context) ([]db.Locale, error) {
	locales, err := s.driver.ListLocales()
	if err != nil {
		return nil, fmt.Errorf("list locales: %w", err)
	}
	if locales == nil {
		return []db.Locale{}, nil
	}
	return *locales, nil
}

// ListEnabledLocales returns only enabled locales (public use).
func (s *LocaleService) ListEnabledLocales(ctx context.Context) ([]db.Locale, error) {
	locales, err := s.driver.ListEnabledLocales()
	if err != nil {
		return nil, fmt.Errorf("list enabled locales: %w", err)
	}
	if locales == nil {
		return []db.Locale{}, nil
	}
	return *locales, nil
}

// ListLocalesPaginated returns a page of locales with a total count.
func (s *LocaleService) ListLocalesPaginated(ctx context.Context, params db.PaginationParams) ([]db.Locale, int64, error) {
	locales, err := s.driver.ListLocalesPaginated(params)
	if err != nil {
		return nil, 0, fmt.Errorf("list locales paginated: %w", err)
	}
	count, err := s.driver.CountLocales()
	if err != nil {
		return nil, 0, fmt.Errorf("count locales: %w", err)
	}
	result := []db.Locale{}
	if locales != nil {
		result = *locales
	}
	return result, *count, nil
}

// ---------------------------------------------------------------------------
// Translation creation
// ---------------------------------------------------------------------------

// CreateTranslation creates locale-specific content field rows for a content
// data node. It copies the default locale's field values as starting content
// for each translatable field. Skips fields that already have rows for the
// target locale.
func (s *LocaleService) CreateTranslation(
	ctx context.Context,
	ac audited.AuditContext,
	contentDataID types.ContentID,
	locale string,
	authorID types.UserID,
) (*TranslationResult, error) {
	cfg, err := s.mgr.Config()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if !cfg.I18nEnabled() {
		return nil, NewValidationError("i18n", "i18n is not enabled")
	}

	if locale == "" {
		return nil, NewValidationError("locale", "locale is required")
	}

	// Verify the locale exists and is enabled.
	loc, err := s.driver.GetLocaleByCode(locale)
	if err != nil {
		return nil, NewValidationError("locale", fmt.Sprintf("locale %q not found", locale))
	}
	if !loc.IsEnabled {
		return nil, NewValidationError("locale", fmt.Sprintf("locale %q is not enabled", locale))
	}

	// Get the content data node to find its datatype.
	cd, err := s.driver.GetContentData(contentDataID)
	if err != nil {
		return nil, fmt.Errorf("get content data: %w", err)
	}

	// Get schema fields for the datatype.
	fields, err := s.driver.ListFieldsByDatatypeID(cd.DatatypeID)
	if err != nil {
		return nil, fmt.Errorf("list fields by datatype: %w", err)
	}

	// Get existing fields for the target locale to avoid duplicates.
	nullableContentDataID := types.NullableContentID{ID: contentDataID, Valid: true}
	existingFields, err := s.driver.ListContentFieldsByContentDataAndLocale(nullableContentDataID, locale)
	if err != nil {
		return nil, fmt.Errorf("list existing locale fields: %w", err)
	}

	existingFieldSet := make(map[types.FieldID]bool)
	if existingFields != nil {
		for _, ef := range *existingFields {
			if ef.FieldID.Valid {
				existingFieldSet[ef.FieldID.ID] = true
			}
		}
	}

	// Get default locale field values to copy as starting content.
	defaultLocale := cfg.I18nDefaultLocale()
	defaultFields, err := s.driver.ListContentFieldsByContentDataAndLocale(nullableContentDataID, defaultLocale)
	if err != nil {
		return nil, fmt.Errorf("list default locale fields: %w", err)
	}

	defaultValueMap := make(map[types.FieldID]string)
	if defaultFields != nil {
		for _, df := range *defaultFields {
			if df.FieldID.Valid {
				defaultValueMap[df.FieldID.ID] = df.FieldValue
			}
		}
	}

	now := types.TimestampNow()
	created := 0

	if fields != nil {
		for _, f := range *fields {
			if !f.Translatable {
				continue
			}
			if existingFieldSet[f.FieldID] {
				continue
			}

			fieldValue := defaultValueMap[f.FieldID]

			_, cfErr := s.driver.CreateContentField(ctx, ac, db.CreateContentFieldParams{
				RouteID:       cd.RouteID,
				ContentDataID: nullableContentDataID,
				FieldID:       types.NullableFieldID{ID: f.FieldID, Valid: true},
				FieldValue:    fieldValue,
				Locale:        locale,
				AuthorID:      authorID,
				DateCreated:   now,
				DateModified:  now,
			})
			if cfErr != nil {
				return nil, fmt.Errorf("create content field for field %s locale %s: %w", f.FieldID, locale, cfErr)
			}
			created++
		}
	}

	return &TranslationResult{Locale: locale, FieldsCreated: created}, nil
}

// CreateAdminTranslation creates locale-specific admin content field rows for
// an admin content data node. Same logic as CreateTranslation but uses admin types.
func (s *LocaleService) CreateAdminTranslation(
	ctx context.Context,
	ac audited.AuditContext,
	adminContentDataID types.AdminContentID,
	locale string,
	authorID types.UserID,
) (*TranslationResult, error) {
	cfg, err := s.mgr.Config()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if !cfg.I18nEnabled() {
		return nil, NewValidationError("i18n", "i18n is not enabled")
	}

	if locale == "" {
		return nil, NewValidationError("locale", "locale is required")
	}

	// Verify the locale exists and is enabled.
	loc, err := s.driver.GetLocaleByCode(locale)
	if err != nil {
		return nil, NewValidationError("locale", fmt.Sprintf("locale %q not found", locale))
	}
	if !loc.IsEnabled {
		return nil, NewValidationError("locale", fmt.Sprintf("locale %q is not enabled", locale))
	}

	// Get the admin content data node to find its datatype.
	cd, err := s.driver.GetAdminContentData(adminContentDataID)
	if err != nil {
		return nil, fmt.Errorf("get admin content data: %w", err)
	}

	// Get schema fields for the admin datatype.
	fields, err := s.driver.ListAdminFieldsByDatatypeID(cd.AdminDatatypeID)
	if err != nil {
		return nil, fmt.Errorf("list admin fields by datatype: %w", err)
	}

	// Get existing fields for the target locale to avoid duplicates.
	nullableAdminContentDataID := types.NullableAdminContentID{ID: adminContentDataID, Valid: true}
	existingFields, err := s.driver.ListAdminContentFieldsByContentDataAndLocale(nullableAdminContentDataID, locale)
	if err != nil {
		return nil, fmt.Errorf("list existing admin locale fields: %w", err)
	}

	existingFieldSet := make(map[types.AdminFieldID]bool)
	if existingFields != nil {
		for _, ef := range *existingFields {
			if ef.AdminFieldID.Valid {
				existingFieldSet[ef.AdminFieldID.ID] = true
			}
		}
	}

	// Get default locale field values to copy as starting content.
	defaultLocale := cfg.I18nDefaultLocale()
	defaultFields, err := s.driver.ListAdminContentFieldsByContentDataAndLocale(nullableAdminContentDataID, defaultLocale)
	if err != nil {
		return nil, fmt.Errorf("list default admin locale fields: %w", err)
	}

	defaultValueMap := make(map[types.AdminFieldID]string)
	if defaultFields != nil {
		for _, df := range *defaultFields {
			if df.AdminFieldID.Valid {
				defaultValueMap[df.AdminFieldID.ID] = df.AdminFieldValue
			}
		}
	}

	now := types.TimestampNow()
	created := 0

	if fields != nil {
		for _, f := range *fields {
			if !f.Translatable {
				continue
			}
			if existingFieldSet[f.AdminFieldID] {
				continue
			}

			fieldValue := defaultValueMap[f.AdminFieldID]

			_, cfErr := s.driver.CreateAdminContentField(ctx, ac, db.CreateAdminContentFieldParams{
				AdminRouteID:       cd.AdminRouteID,
				AdminContentDataID: nullableAdminContentDataID,
				AdminFieldID:       types.NullableAdminFieldID{ID: f.AdminFieldID, Valid: true},
				AdminFieldValue:    fieldValue,
				Locale:             locale,
				AuthorID:           authorID,
				DateCreated:        now,
				DateModified:       now,
			})
			if cfErr != nil {
				return nil, fmt.Errorf("create admin content field for field %s locale %s: %w", f.AdminFieldID, locale, cfErr)
			}
			created++
		}
	}

	return &TranslationResult{Locale: locale, FieldsCreated: created}, nil
}

// ---------------------------------------------------------------------------
// Locale resolution (HTTP-aware)
// ---------------------------------------------------------------------------

// ResolveLocale determines the locale for a delivery request.
// Priority: ?locale= query param > Accept-Language header > default locale.
// Returns "" when i18n is disabled or config is unavailable.
func (s *LocaleService) ResolveLocale(r *http.Request) string {
	cfg, err := s.mgr.Config()
	if err != nil {
		return ""
	}
	if !cfg.I18nEnabled() {
		return ""
	}

	// 1. Explicit query parameter takes highest priority.
	if qLocale := r.URL.Query().Get("locale"); qLocale != "" {
		return qLocale
	}

	// 2. Parse Accept-Language header and match against enabled locales.
	enabledLocales, err := s.driver.ListEnabledLocales()
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
func (s *LocaleService) WalkFallback(ctx context.Context, contentDataID types.ContentID, requestedLocale string) (string, *db.ContentVersion, error) {
	// Try the requested locale first.
	version, err := s.driver.GetPublishedSnapshot(contentDataID, requestedLocale)
	if err == nil {
		return requestedLocale, version, nil
	}

	// Walk the fallback chain (max 2 hops).
	currentLocale := requestedLocale
	for hop := range 2 {
		_ = hop
		locale, locErr := s.driver.GetLocaleByCode(currentLocale)
		if locErr != nil {
			break
		}
		if locale.FallbackCode == "" {
			break
		}

		version, err = s.driver.GetPublishedSnapshot(contentDataID, locale.FallbackCode)
		if err == nil {
			return locale.FallbackCode, version, nil
		}
		currentLocale = locale.FallbackCode
	}

	return "", nil, nil
}

// BuildLocaleMetadata builds a metadata response showing which locales
// have published snapshots for the given content data.
func (s *LocaleService) BuildLocaleMetadata(ctx context.Context, contentDataID types.ContentID) (*LocaleMetadata, error) {
	enabledLocales, err := s.driver.ListEnabledLocales()
	if err != nil {
		return nil, fmt.Errorf("list enabled locales: %w", err)
	}

	meta := &LocaleMetadata{
		Locales:          make(map[string]LocaleVersionInfo),
		AvailableLocales: make([]string, 0),
	}

	if enabledLocales == nil {
		return meta, nil
	}

	for _, loc := range *enabledLocales {
		version, verErr := s.driver.GetPublishedSnapshot(contentDataID, loc.Code)
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
	version, verErr := s.driver.GetPublishedSnapshot(contentDataID, "")
	if verErr == nil && version != nil {
		meta.Locales[""] = LocaleVersionInfo{
			Published:   true,
			PublishedAt: version.DateCreated.String(),
		}
	}

	return meta, nil
}

// ---------------------------------------------------------------------------
// Validation helpers (private)
// ---------------------------------------------------------------------------

// validateLocaleCode validates a locale code against a basic BCP 47 pattern:
// 2-5 lowercase letters optionally followed by a hyphen and 2-3 alphanumeric characters.
// Uses character-by-character validation -- no regex.
func validateLocaleCode(code string) *ValidationError {
	if len(code) == 0 {
		return NewValidationError("code", "locale code must not be empty")
	}

	// Find the hyphen position, if any.
	hyphenIdx := -1
	for i := range len(code) {
		if code[i] == '-' {
			hyphenIdx = i
			break
		}
	}

	var primary, subtag string
	if hyphenIdx == -1 {
		primary = code
	} else {
		primary = code[:hyphenIdx]
		subtag = code[hyphenIdx+1:]
	}

	// Primary subtag: 2-5 lowercase ASCII letters.
	if len(primary) < 2 || len(primary) > 5 {
		return NewValidationError("code", fmt.Sprintf("locale code primary subtag must be 2-5 lowercase letters, got %q", primary))
	}
	for i := range len(primary) {
		ch := primary[i]
		if ch < 'a' || ch > 'z' {
			return NewValidationError("code", fmt.Sprintf("locale code primary subtag must contain only lowercase letters, found %q at position %d", string(ch), i))
		}
	}

	// If there is a hyphen, the subtag is required: 2-3 alphanumeric characters.
	if hyphenIdx != -1 {
		if len(subtag) < 2 || len(subtag) > 3 {
			return NewValidationError("code", fmt.Sprintf("locale code subtag must be 2-3 alphanumeric characters, got %q", subtag))
		}
		for i := range len(subtag) {
			ch := subtag[i]
			isLower := ch >= 'a' && ch <= 'z'
			isUpper := ch >= 'A' && ch <= 'Z'
			isDigit := ch >= '0' && ch <= '9'
			if !isLower && !isUpper && !isDigit {
				return NewValidationError("code", fmt.Sprintf("locale code subtag must contain only alphanumeric characters, found %q at position %d", string(ch), i))
			}
		}
	}

	return nil
}

// validateFallbackChain verifies that fallbackCode references an existing locale
// and that the chain does not form a cycle. Walks at most 2 hops.
func (s *LocaleService) validateFallbackChain(fallbackCode, selfCode string) error {
	if fallbackCode == selfCode {
		return NewValidationError("fallback_code", "locale cannot fall back to itself")
	}

	// Hop 1: the direct fallback must exist.
	fb1, err := s.driver.GetLocaleByCode(fallbackCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NewValidationError("fallback_code", fmt.Sprintf("fallback locale %q does not exist", fallbackCode))
		}
		return NewValidationError("fallback_code", fmt.Sprintf("fallback locale %q does not exist", fallbackCode))
	}

	// Hop 2: if the direct fallback itself has a fallback, check it does not
	// point back to selfCode.
	// Note: the db layer returns "null" for NULL strings (via ReadNullString).
	if fb1.FallbackCode == "" || fb1.FallbackCode == "null" {
		return nil
	}
	if fb1.FallbackCode == selfCode {
		return NewValidationError("fallback_code", fmt.Sprintf("fallback chain cycle detected: %s -> %s -> %s", selfCode, fallbackCode, selfCode))
	}

	// Verify the second-hop locale exists.
	_, err = s.driver.GetLocaleByCode(fb1.FallbackCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NewValidationError("fallback_code", fmt.Sprintf("second-hop fallback locale %q (via %q) does not exist", fb1.FallbackCode, fallbackCode))
		}
		return NewValidationError("fallback_code", fmt.Sprintf("second-hop fallback locale %q (via %q) does not exist", fb1.FallbackCode, fallbackCode))
	}

	return nil
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
