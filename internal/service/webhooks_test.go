// Integration tests for WebhookService against a real SQLite database.
package service_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
)

// webhookConfigProvider satisfies config.Provider for webhook tests.
type webhookConfigProvider struct {
	cfg *config.Config
}

func (p *webhookConfigProvider) Get() (*config.Config, error) {
	return p.cfg, nil
}

// webhookTestSetup creates a fresh database, seeds a user, and returns a
// WebhookService wired to a test config.Manager.
func webhookTestSetup(t *testing.T) (db.Database, *service.WebhookService, types.UserID, audited.AuditContext) {
	t.Helper()
	d, _ := testDB(t)
	userID := seedUser(t, d)
	ac := testAuditCtx(d)

	cfg := d.Config
	cfg.Webhook_Allow_HTTP = true
	mgr := config.NewManager(&webhookConfigProvider{cfg: &cfg})

	svc := service.NewWebhookService(d, mgr, nil)
	return d, svc, userID, ac
}

// seedWebhook creates a webhook directly via the driver, bypassing SSRF validation.
// Use this for tests that need an existing webhook but aren't testing creation validation.
func seedWebhook(t *testing.T, d db.Database, ac audited.AuditContext, userID types.UserID, url, name, secret string) *db.Webhook {
	t.Helper()
	ctx := context.Background()
	now := types.TimestampNow()

	if secret == "" {
		secret = "test-secret-" + types.NewWebhookID().String()
	}
	if name == "" {
		name = "Test Webhook"
	}

	wh, err := d.CreateWebhook(ctx, ac, db.CreateWebhookParams{
		Name:         name,
		URL:          url,
		Secret:       secret,
		Events:       []string{"content.published"},
		IsActive:     true,
		Headers:      map[string]string{},
		AuthorID:     userID,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		t.Fatalf("seedWebhook: %v", err)
	}
	return wh
}

// --- Create tests ---

func TestWebhookService_CreateEmptyName(t *testing.T) {
	_, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	_, err := svc.CreateWebhook(ctx, ac, service.CreateWebhookInput{
		Name:   "",
		URL:    "https://example.com/webhook",
		Events: []string{"content.published"},
	}, userID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestWebhookService_CreateEmptyURL(t *testing.T) {
	_, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	_, err := svc.CreateWebhook(ctx, ac, service.CreateWebhookInput{
		Name:   "No URL",
		URL:    "",
		Events: []string{"content.published"},
	}, userID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestWebhookService_CreateSSRF(t *testing.T) {
	_, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	_, err := svc.CreateWebhook(ctx, ac, service.CreateWebhookInput{
		Name:   "SSRF Webhook",
		URL:    "http://127.0.0.1:8080/evil",
		Events: []string{"content.published"},
	}, userID)
	if err == nil {
		t.Fatal("expected error for SSRF URL, got nil")
	}
	if !service.IsValidation(err) {
		t.Errorf("expected ValidationError for SSRF URL, got %T: %v", err, err)
	}
}

func TestWebhookService_CreateEmptyEvents(t *testing.T) {
	_, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	_, err := svc.CreateWebhook(ctx, ac, service.CreateWebhookInput{
		Name:   "No Events",
		URL:    "https://example.com/webhook",
		Events: []string{},
	}, userID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

func TestWebhookService_CreateExplicitSecret(t *testing.T) {
	d, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Seed via driver (bypasses SSRF on loopback), then verify via service.
	wh := seedWebhook(t, d, ac, userID, ts.URL, "Explicit Secret", "my-custom-secret")

	got, err := svc.GetWebhook(ctx, wh.WebhookID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Secret != "my-custom-secret" {
		t.Errorf("Secret = %q, want %q", got.Secret, "my-custom-secret")
	}
}

// --- Update tests ---

func TestWebhookService_UpdateRename(t *testing.T) {
	d, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	// Use a real external URL that passes SSRF validation.
	wh := seedWebhook(t, d, ac, userID, "https://example.com/webhook", "Original", "original-secret")

	updated, err := svc.UpdateWebhook(ctx, ac, service.UpdateWebhookInput{
		WebhookID: wh.WebhookID,
		Name:      "Renamed",
		URL:       "https://example.com/webhook",
		Secret:    wh.Secret,
		Events:    []string{"content.published", "content.updated"},
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Renamed" {
		t.Errorf("Name = %q, want %q", updated.Name, "Renamed")
	}
	if updated.AuthorID != wh.AuthorID {
		t.Errorf("AuthorID changed: got %s, want %s", updated.AuthorID, wh.AuthorID)
	}
	if updated.DateCreated != wh.DateCreated {
		t.Errorf("DateCreated changed: got %v, want %v", updated.DateCreated, wh.DateCreated)
	}
}

func TestWebhookService_UpdateSSRF(t *testing.T) {
	d, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	wh := seedWebhook(t, d, ac, userID, "https://example.com/webhook", "Before SSRF", "")

	_, err := svc.UpdateWebhook(ctx, ac, service.UpdateWebhookInput{
		WebhookID: wh.WebhookID,
		Name:      "SSRF Update",
		URL:       "http://127.0.0.1:9999/evil",
		Secret:    wh.Secret,
		Events:    []string{"content.published"},
		IsActive:  true,
	})
	if err == nil {
		t.Fatal("expected error for SSRF URL, got nil")
	}
	if !service.IsValidation(err) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}
}

// --- Delete tests ---

func TestWebhookService_Delete(t *testing.T) {
	d, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	wh := seedWebhook(t, d, ac, userID, "https://example.com/hook", "To Delete", "")

	if err := svc.DeleteWebhook(ctx, ac, wh.WebhookID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, getErr := svc.GetWebhook(ctx, wh.WebhookID)
	if getErr == nil {
		t.Error("expected NotFoundError after delete, got nil")
	}
	if !service.IsNotFound(getErr) {
		t.Errorf("expected NotFoundError, got %T: %v", getErr, getErr)
	}
}

func TestWebhookService_DeleteNonExistent(t *testing.T) {
	_, svc, _, ac := webhookTestSetup(t)
	ctx := context.Background()

	fakeID := types.NewWebhookID()
	// DeleteWebhook may or may not error on non-existent -- verify no panic.
	_ = svc.DeleteWebhook(ctx, ac, fakeID)
}

// --- Get tests ---

func TestWebhookService_GetExisting(t *testing.T) {
	d, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	wh := seedWebhook(t, d, ac, userID, "https://example.com/hook", "Get Test", "get-secret")

	got, err := svc.GetWebhook(ctx, wh.WebhookID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Get Test" {
		t.Errorf("Name = %q, want %q", got.Name, "Get Test")
	}
	if got.WebhookID != wh.WebhookID {
		t.Errorf("WebhookID = %s, want %s", got.WebhookID, wh.WebhookID)
	}
	if got.Secret != "get-secret" {
		t.Errorf("Secret = %q, want %q", got.Secret, "get-secret")
	}
}

func TestWebhookService_GetNonExistent(t *testing.T) {
	_, svc, _, _ := webhookTestSetup(t)
	ctx := context.Background()

	fakeID := types.NewWebhookID()
	_, err := svc.GetWebhook(ctx, fakeID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- List paginated ---

func TestWebhookService_ListPaginated(t *testing.T) {
	d, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	for i := range 3 {
		name := "Webhook " + string(rune('A'+i))
		seedWebhook(t, d, ac, userID, "https://example.com/hook"+string(rune('A'+i)), name, "")
	}

	items, total, err := svc.ListWebhooksPaginated(ctx, 2, 0)
	if err != nil {
		t.Fatalf("list paginated: %v", err)
	}
	if total == nil {
		t.Fatal("total is nil")
	}
	if *total != 3 {
		t.Errorf("total = %d, want 3", *total)
	}
	if items == nil || len(*items) != 2 {
		count := 0
		if items != nil {
			count = len(*items)
		}
		t.Errorf("items count = %d, want 2", count)
	}
}

// --- TestWebhook ---
// Note: TestWebhook performs SSRF re-validation at test time. Since httptest
// servers bind to 127.0.0.1 (loopback), the SSRF check will correctly block
// the request and return a WebhookTestResult with Status "failed". This is
// correct behavior -- the SSRF check is working as designed.

func TestWebhookService_TestWebhookSSRF(t *testing.T) {
	d, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Seed a webhook pointing to the httptest server (loopback).
	wh := seedWebhook(t, d, ac, userID, ts.URL, "SSRF Test", "test-secret")

	// TestWebhook should return a result (not error) with "failed" status
	// because the SSRF re-check catches the loopback URL.
	result, err := svc.TestWebhook(ctx, wh.WebhookID)
	if err != nil {
		t.Fatalf("test webhook: %v", err)
	}
	if result.Status != "failed" {
		t.Errorf("Status = %q, want %q (SSRF should block loopback)", result.Status, "failed")
	}
	if result.Duration != "0s" {
		t.Errorf("Duration = %q, want %q (no HTTP request should be made)", result.Duration, "0s")
	}
}

func TestWebhookService_TestWebhookNotFound(t *testing.T) {
	_, svc, _, _ := webhookTestSetup(t)
	ctx := context.Background()

	fakeID := types.NewWebhookID()
	_, err := svc.TestWebhook(ctx, fakeID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !service.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

// --- Deliveries ---

func TestWebhookService_ListDeliveries(t *testing.T) {
	d, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	wh := seedWebhook(t, d, ac, userID, "https://example.com/hook", "Delivery List", "")

	_, delErr := d.CreateWebhookDelivery(ctx, db.CreateWebhookDeliveryParams{
		WebhookID: wh.WebhookID,
		Event:     "content.published",
		Payload:   `{"test": true}`,
		Status:    db.DeliveryStatusPending,
		Attempts:  0,
		CreatedAt: types.TimestampNow(),
	})
	if delErr != nil {
		t.Fatalf("create delivery: %v", delErr)
	}

	deliveries, err := svc.ListDeliveries(ctx, wh.WebhookID)
	if err != nil {
		t.Fatalf("list deliveries: %v", err)
	}
	if deliveries == nil || len(*deliveries) != 1 {
		count := 0
		if deliveries != nil {
			count = len(*deliveries)
		}
		t.Errorf("delivery count = %d, want 1", count)
	}
}

func TestWebhookService_RetryDelivery(t *testing.T) {
	d, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	wh := seedWebhook(t, d, ac, userID, "https://example.com/hook", "Retry Test", "")

	del, err := d.CreateWebhookDelivery(ctx, db.CreateWebhookDeliveryParams{
		WebhookID: wh.WebhookID,
		Event:     "content.published",
		Payload:   `{"retry": true}`,
		Status:    db.DeliveryStatusFailed,
		Attempts:  1,
		CreatedAt: types.TimestampNow(),
	})
	if err != nil {
		t.Fatalf("create delivery: %v", err)
	}

	result, retryErr := svc.RetryDelivery(ctx, del.DeliveryID)
	if retryErr != nil {
		t.Fatalf("retry delivery: %v", retryErr)
	}
	if result.Status != "queued" {
		t.Errorf("Status = %q, want %q", result.Status, "queued")
	}
	if result.DeliveryID != del.DeliveryID {
		t.Errorf("DeliveryID = %s, want %s", result.DeliveryID, del.DeliveryID)
	}

	updated, getErr := d.GetWebhookDelivery(del.DeliveryID)
	if getErr != nil {
		t.Fatalf("get updated delivery: %v", getErr)
	}
	if updated.Status != db.DeliveryStatusRetrying {
		t.Errorf("Status = %q, want %q", updated.Status, db.DeliveryStatusRetrying)
	}
}

func TestWebhookService_RetryNonExistent(t *testing.T) {
	_, svc, _, _ := webhookTestSetup(t)
	ctx := context.Background()

	fakeID := types.NewWebhookDeliveryID()
	_, err := svc.RetryDelivery(ctx, fakeID)
	if err == nil {
		t.Fatal("expected error for non-existent delivery, got nil")
	}
	if !service.IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestWebhookService_PruneDeliveries(t *testing.T) {
	d, svc, userID, ac := webhookTestSetup(t)
	ctx := context.Background()

	wh := seedWebhook(t, d, ac, userID, "https://example.com/hook", "Prune Test", "")

	oldTime := types.NewTimestamp(time.Now().UTC().Add(-48 * time.Hour))
	_, delErr := d.CreateWebhookDelivery(ctx, db.CreateWebhookDeliveryParams{
		WebhookID: wh.WebhookID,
		Event:     "content.published",
		Payload:   `{"old": true}`,
		Status:    db.DeliveryStatusSuccess,
		Attempts:  1,
		CreatedAt: oldTime,
	})
	if delErr != nil {
		t.Fatalf("create old delivery: %v", delErr)
	}

	pruneErr := svc.PruneDeliveries(ctx, time.Now().UTC().Add(-24*time.Hour))
	if pruneErr != nil {
		t.Fatalf("prune: %v", pruneErr)
	}

	deliveries, listErr := svc.ListDeliveries(ctx, wh.WebhookID)
	if listErr != nil {
		t.Fatalf("list after prune: %v", listErr)
	}
	if deliveries != nil && len(*deliveries) != 0 {
		t.Errorf("expected 0 deliveries after prune, got %d", len(*deliveries))
	}
}

// --- Error type verification ---

func TestWebhookErrorTypes(t *testing.T) {
	t.Run("NotFoundError", func(t *testing.T) {
		err := &service.NotFoundError{Resource: "webhook", ID: "test-id"}
		var target *service.NotFoundError
		if !errors.As(err, &target) {
			t.Error("errors.As should match NotFoundError")
		}
		if !service.IsNotFound(err) {
			t.Error("IsNotFound should return true")
		}
	})

	t.Run("ValidationError", func(t *testing.T) {
		err := service.NewValidationError("field", "message")
		if !service.IsValidation(err) {
			t.Error("IsValidation should return true")
		}
	})
}
