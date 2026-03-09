package service_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// staticProvider implements config.Provider for tests.
type staticProvider struct {
	cfg *config.Config
}

func (p *staticProvider) Get() (*config.Config, error) {
	return p.cfg, nil
}

func testLocaleDB(t *testing.T, i18nEnabled bool) (db.Database, *service.LocaleService) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if _, err := conn.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if _, err := conn.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}

	cfg := config.Config{
		Node_ID:             types.NewNodeID().String(),
		I18n_Enabled:        i18nEnabled,
		I18n_Default_Locale: "en",
	}

	d := db.Database{
		Connection: conn,
		Context:    context.Background(),
		Config:     cfg,
	}

	if err := d.CreateAllTables(); err != nil {
		t.Fatalf("CreateAllTables: %v", err)
	}

	mgr := config.NewManager(&staticProvider{cfg: &cfg})
	if err := mgr.Load(); err != nil {
		t.Fatalf("mgr.Load: %v", err)
	}

	svc := service.NewLocaleService(d, mgr)
	return d, svc
}

func localeAuditCtx(d db.Database) audited.AuditContext {
	return audited.Ctx(types.NodeID(d.Config.Node_ID), types.UserID(""), "test", "127.0.0.1")
}

// ---------------------------------------------------------------------------
// 1. Create with valid BCP 47 code
// ---------------------------------------------------------------------------

func TestLocaleService_CreateValid(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	created, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "en",
		Label:     "English",
		IsDefault: true,
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale: %v", err)
	}
	if created.Code != "en" {
		t.Errorf("Code = %q, want %q", created.Code, "en")
	}
	if !created.IsDefault {
		t.Error("expected IsDefault = true")
	}
}

func TestLocaleService_CreateValidSubtag(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	created, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "en-US",
		Label:     "English (US)",
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale: %v", err)
	}
	if created.Code != "en-US" {
		t.Errorf("Code = %q, want %q", created.Code, "en-US")
	}
}

// ---------------------------------------------------------------------------
// 2. Create with invalid codes
// ---------------------------------------------------------------------------

func TestLocaleService_CreateInvalidCodes(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	tests := []struct {
		name string
		code string
	}{
		{"empty", ""},
		{"too short", "e"},
		{"underscore", "en_US"},
		{"trailing hyphen", "en-"},
		{"uppercase primary", "EN"},
		{"too long primary", "abcdef"},
		{"digit in primary", "e1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
				Code:  tt.code,
				Label: "test",
			})
			if err == nil {
				t.Fatalf("expected ValidationError for code %q, got nil", tt.code)
			}
			if !service.IsValidation(err) {
				t.Errorf("expected ValidationError, got %T: %v", err, err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. Create with empty label
// ---------------------------------------------------------------------------

func TestLocaleService_CreateEmptyLabel(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	_, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:  "en",
		Label: "",
	})
	if err == nil {
		t.Fatal("expected ValidationError for empty label, got nil")
	}
	if !service.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// 4. Create with is_default clears other defaults
// ---------------------------------------------------------------------------

func TestLocaleService_CreateDefaultClearsOthers(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	first, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "en",
		Label:     "English",
		IsDefault: true,
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale first: %v", err)
	}

	_, err = svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "fr",
		Label:     "French",
		IsDefault: true,
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale second: %v", err)
	}

	// First locale should no longer be default.
	refreshed, err := svc.GetLocale(ctx, first.LocaleID)
	if err != nil {
		t.Fatalf("GetLocale: %v", err)
	}
	if refreshed.IsDefault {
		t.Error("expected first locale IsDefault = false after creating second default")
	}
}

// ---------------------------------------------------------------------------
// 5. Create with duplicate code
// ---------------------------------------------------------------------------

func TestLocaleService_CreateDuplicateCode(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	_, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:  "en",
		Label: "English",
	})
	if err != nil {
		t.Fatalf("CreateLocale first: %v", err)
	}

	_, err = svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:  "en",
		Label: "English duplicate",
	})
	if err == nil {
		t.Fatal("expected ConflictError for duplicate code, got nil")
	}
	if !service.IsConflict(err) {
		t.Errorf("expected ConflictError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// 6. Create with fallback self-reference
// ---------------------------------------------------------------------------

func TestLocaleService_CreateFallbackSelfRef(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	_, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:         "en",
		Label:        "English",
		FallbackCode: "en",
	})
	if err == nil {
		t.Fatal("expected ValidationError for self-reference fallback, got nil")
	}
	if !service.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// 7. Create with fallback to non-existent locale
// ---------------------------------------------------------------------------

func TestLocaleService_CreateFallbackNonExistent(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	_, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:         "fr",
		Label:        "French",
		FallbackCode: "en",
	})
	if err == nil {
		t.Fatal("expected ValidationError for non-existent fallback, got nil")
	}
	if !service.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// 8. Create with fallback cycle
// ---------------------------------------------------------------------------

func TestLocaleService_CreateFallbackCycle(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	// Create "en" with no fallback.
	_, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:  "en",
		Label: "English",
	})
	if err != nil {
		t.Fatalf("CreateLocale en: %v", err)
	}

	// Create "fr" falling back to "en".
	_, err = svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:         "fr",
		Label:        "French",
		FallbackCode: "en",
	})
	if err != nil {
		t.Fatalf("CreateLocale fr: %v", err)
	}

	// Create "de" falling back to "fr" — this forms de -> fr -> en (OK, no cycle).
	_, err = svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:         "de",
		Label:        "German",
		FallbackCode: "fr",
	})
	if err != nil {
		t.Fatalf("CreateLocale de: %v", err)
	}

	// Now try to update "en" to fall back to "de" — this would create en -> de -> fr -> en (cycle via 2 hops).
	// But the service checks 2 hops from fallback, so en->de, de.fallback=fr, fr.fallback=en => cycle.
	// Actually, let's test create directly with cycle detection.
	// Create "es" with fallback "de" — es -> de -> fr. Not a cycle since fr->en doesn't equal es. This is OK.
	_, err = svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:         "es",
		Label:        "Spanish",
		FallbackCode: "de",
	})
	if err != nil {
		t.Fatalf("CreateLocale es (should succeed): %v", err)
	}

	// Now test actual cycle: update "en" to fallback to "fr".
	// en -> fr, fr.fallback = en => cycle at hop 2. en -> fr -> en.
	enLocale, err := svc.GetLocaleByCode(ctx, "en")
	if err != nil {
		t.Fatalf("GetLocaleByCode en: %v", err)
	}
	_, err = svc.UpdateLocale(ctx, ac, service.UpdateLocaleInput{
		LocaleID:     enLocale.LocaleID,
		Code:         "en",
		Label:        "English",
		FallbackCode: "fr",
	})
	if err == nil {
		t.Fatal("expected ValidationError for cycle, got nil")
	}
	if !service.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// 9. Update locale
// ---------------------------------------------------------------------------

func TestLocaleService_Update(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	created, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "en",
		Label:     "English",
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale: %v", err)
	}

	updated, err := svc.UpdateLocale(ctx, ac, service.UpdateLocaleInput{
		LocaleID:  created.LocaleID,
		Code:      "en",
		Label:     "English (Updated)",
		IsDefault: true,
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("UpdateLocale: %v", err)
	}
	if updated.Label != "English (Updated)" {
		t.Errorf("Label = %q, want %q", updated.Label, "English (Updated)")
	}
	if !updated.IsDefault {
		t.Error("expected IsDefault = true")
	}
	// DateCreated should be preserved.
	if updated.DateCreated != created.DateCreated {
		t.Errorf("DateCreated changed: %v -> %v", created.DateCreated, updated.DateCreated)
	}
}

// ---------------------------------------------------------------------------
// 10. Delete default locale
// ---------------------------------------------------------------------------

func TestLocaleService_DeleteDefault(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	created, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "en",
		Label:     "English",
		IsDefault: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale: %v", err)
	}

	err = svc.DeleteLocale(ctx, ac, created.LocaleID)
	if err == nil {
		t.Fatal("expected ForbiddenError when deleting default locale, got nil")
	}
	if !service.IsForbidden(err) {
		t.Errorf("expected ForbiddenError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// 11. Delete non-existent locale
// ---------------------------------------------------------------------------

func TestLocaleService_DeleteNonExistent(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	err := svc.DeleteLocale(ctx, ac, types.LocaleID("nonexistent"))
	if err == nil {
		t.Fatal("expected NotFoundError, got nil")
	}
	if !service.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// 12. Delete non-default locale
// ---------------------------------------------------------------------------

func TestLocaleService_DeleteNonDefault(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	created, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "fr",
		Label:     "French",
		IsDefault: false,
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale: %v", err)
	}

	err = svc.DeleteLocale(ctx, ac, created.LocaleID)
	if err != nil {
		t.Fatalf("DeleteLocale: %v", err)
	}

	// Verify it's gone.
	_, err = svc.GetLocale(ctx, created.LocaleID)
	if err == nil {
		t.Fatal("expected NotFoundError after delete, got nil")
	}
	if !service.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// 13. ListEnabledLocales
// ---------------------------------------------------------------------------

func TestLocaleService_ListEnabled(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	_, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "en",
		Label:     "English",
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale en: %v", err)
	}

	_, err = svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "fr",
		Label:     "French",
		IsEnabled: false,
	})
	if err != nil {
		t.Fatalf("CreateLocale fr: %v", err)
	}

	_, err = svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "de",
		Label:     "German",
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale de: %v", err)
	}

	enabled, err := svc.ListEnabledLocales(ctx)
	if err != nil {
		t.Fatalf("ListEnabledLocales: %v", err)
	}
	if len(enabled) != 2 {
		t.Errorf("ListEnabledLocales count = %d, want 2", len(enabled))
	}

	// All locales.
	all, err := svc.ListLocales(ctx)
	if err != nil {
		t.Fatalf("ListLocales: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("ListLocales count = %d, want 3", len(all))
	}
}

// ---------------------------------------------------------------------------
// 14. ResolveLocale
// ---------------------------------------------------------------------------

func TestLocaleService_ResolveLocale(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, true)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	// Seed some locales.
	_, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "en",
		Label:     "English",
		IsEnabled: true,
		IsDefault: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale en: %v", err)
	}

	_, err = svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "fr",
		Label:     "French",
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale fr: %v", err)
	}

	t.Run("query param takes priority", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/content/?locale=fr", nil)
		got := svc.ResolveLocale(r)
		if got != "fr" {
			t.Errorf("ResolveLocale = %q, want %q", got, "fr")
		}
	})

	t.Run("Accept-Language header", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/content/", nil)
		r.Header.Set("Accept-Language", "fr, en;q=0.9")
		got := svc.ResolveLocale(r)
		if got != "fr" {
			t.Errorf("ResolveLocale = %q, want %q", got, "fr")
		}
	})

	t.Run("falls back to default", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/content/", nil)
		r.Header.Set("Accept-Language", "es")
		got := svc.ResolveLocale(r)
		if got != "en" {
			t.Errorf("ResolveLocale = %q, want %q", got, "en")
		}
	})

	t.Run("no header, no param -> default", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/content/", nil)
		got := svc.ResolveLocale(r)
		if got != "en" {
			t.Errorf("ResolveLocale = %q, want %q", got, "en")
		}
	})
}

func TestLocaleService_ResolveLocale_I18nDisabled(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)

	r := httptest.NewRequest(http.MethodGet, "/api/v1/content/?locale=fr", nil)
	got := svc.ResolveLocale(r)
	if got != "" {
		t.Errorf("ResolveLocale with i18n disabled = %q, want empty string", got)
	}
}

// ---------------------------------------------------------------------------
// 15. CreateTranslation with i18n disabled
// ---------------------------------------------------------------------------

func TestLocaleService_CreateTranslation_I18nDisabled(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	_, err := svc.CreateTranslation(ctx, ac, types.ContentID("fake"), "fr", types.UserID("u"))
	if err == nil {
		t.Fatal("expected ValidationError for i18n disabled, got nil")
	}
	if !service.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

// ---------------------------------------------------------------------------
// 16. Validate fallback chain with valid chain
// ---------------------------------------------------------------------------

func TestLocaleService_ValidFallbackChain(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	// Create "en" locale.
	_, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:  "en",
		Label: "English",
	})
	if err != nil {
		t.Fatalf("CreateLocale en: %v", err)
	}

	// Create "fr" with fallback to "en" — valid chain.
	created, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:         "fr",
		Label:        "French",
		FallbackCode: "en",
	})
	if err != nil {
		t.Fatalf("CreateLocale fr with fallback: %v", err)
	}
	if created.FallbackCode != "en" {
		t.Errorf("FallbackCode = %q, want %q", created.FallbackCode, "en")
	}
}

// ---------------------------------------------------------------------------
// 17. Update set default clears others
// ---------------------------------------------------------------------------

func TestLocaleService_UpdateSetDefault(t *testing.T) {
	t.Parallel()
	_, svc := testLocaleDB(t, false)
	ctx := context.Background()
	ac := audited.Ctx(types.NewNodeID(), types.UserID(""), "test", "127.0.0.1")

	first, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "en",
		Label:     "English",
		IsDefault: true,
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale first: %v", err)
	}

	second, err := svc.CreateLocale(ctx, ac, service.CreateLocaleInput{
		Code:      "fr",
		Label:     "French",
		IsDefault: false,
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("CreateLocale second: %v", err)
	}

	// Set second as default via update.
	_, err = svc.UpdateLocale(ctx, ac, service.UpdateLocaleInput{
		LocaleID:  second.LocaleID,
		Code:      "fr",
		Label:     "French",
		IsDefault: true,
		IsEnabled: true,
	})
	if err != nil {
		t.Fatalf("UpdateLocale: %v", err)
	}

	// First should no longer be default.
	refreshed, err := svc.GetLocale(ctx, first.LocaleID)
	if err != nil {
		t.Fatalf("GetLocale: %v", err)
	}
	if refreshed.IsDefault {
		t.Error("expected first locale IsDefault = false after update")
	}
}
