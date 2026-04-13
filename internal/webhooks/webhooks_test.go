package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ---------------------------------------------------------------------------
// Mock driver
// ---------------------------------------------------------------------------

// mockDriver implements only the db.DbDriver methods used by the webhook
// dispatcher. The embedded nil interface panics on any unimplemented method,
// ensuring tests fail loudly if unexpected methods are called.
type mockDriver struct {
	db.DbDriver

	mu sync.Mutex

	activeWebhooks    *[]db.Webhook
	activeWebhooksErr error

	createdDelivery   *db.WebhookDelivery
	createDeliveryErr error
	createCalls       []db.CreateWebhookDeliveryParams

	updateStatusErr error
	updateCalls     []db.UpdateWebhookDeliveryStatusParams

	webhook       *db.Webhook
	getWebhookErr error

	pendingRetries    *[]db.WebhookDelivery
	pendingRetriesErr error
}

func (m *mockDriver) ListActiveWebhooks() (*[]db.Webhook, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.activeWebhooks, m.activeWebhooksErr
}

func (m *mockDriver) CreateWebhookDelivery(_ context.Context, p db.CreateWebhookDeliveryParams) (*db.WebhookDelivery, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.createCalls = append(m.createCalls, p)
	if m.createDeliveryErr != nil {
		return nil, m.createDeliveryErr
	}
	return m.createdDelivery, nil
}

func (m *mockDriver) UpdateWebhookDeliveryStatus(_ context.Context, p db.UpdateWebhookDeliveryStatusParams) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateCalls = append(m.updateCalls, p)
	return m.updateStatusErr
}

func (m *mockDriver) GetWebhook(_ types.WebhookID) (*db.Webhook, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.webhook, m.getWebhookErr
}

func (m *mockDriver) ListPendingRetries(_ types.Timestamp, _ int64) (*[]db.WebhookDelivery, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.pendingRetries, m.pendingRetriesErr
}

func (m *mockDriver) getCreateCalls() []db.CreateWebhookDeliveryParams {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]db.CreateWebhookDeliveryParams, len(m.createCalls))
	copy(cp, m.createCalls)
	return cp
}

func (m *mockDriver) getUpdateCalls() []db.UpdateWebhookDeliveryStatusParams {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]db.UpdateWebhookDeliveryStatusParams, len(m.updateCalls))
	copy(cp, m.updateCalls)
	return cp
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func testConfig() config.Config {
	return config.Config{
		Webhook_Workers:     2,
		Webhook_Timeout:     5,
		Webhook_Max_Retries: 3,
		Webhook_Allow_HTTP:  true,
	}
}

// ---------------------------------------------------------------------------
// Signing
// ---------------------------------------------------------------------------

func TestSign(t *testing.T) {
	tests := []struct {
		name    string
		secret  string
		payload []byte
	}{
		{"basic", "my-secret", []byte(`{"event":"test"}`)},
		{"empty secret", "", []byte(`{"event":"test"}`)},
		{"empty payload", "secret", nil},
		{"both empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := Sign(tt.secret, tt.payload)
			if sig == "" {
				t.Error("Sign returned empty string")
			}
			sig2 := Sign(tt.secret, tt.payload)
			if sig != sig2 {
				t.Errorf("Sign not deterministic: %q != %q", sig, sig2)
			}
		})
	}
}

func TestSign_DifferentSecrets(t *testing.T) {
	payload := []byte(`{"event":"test"}`)
	sig1 := Sign("secret-a", payload)
	sig2 := Sign("secret-b", payload)
	if sig1 == sig2 {
		t.Error("different secrets produced identical signatures")
	}
}

func TestSign_DifferentPayloads(t *testing.T) {
	sig1 := Sign("secret", []byte(`{"event":"a"}`))
	sig2 := Sign("secret", []byte(`{"event":"b"}`))
	if sig1 == sig2 {
		t.Error("different payloads produced identical signatures")
	}
}

func TestVerify(t *testing.T) {
	secret := "webhook-secret"
	payload := []byte(`{"id":"123","event":"content.published"}`)
	sig := Sign(secret, payload)

	tests := []struct {
		name      string
		secret    string
		signature string
		payload   []byte
		want      bool
	}{
		{"valid", secret, sig, payload, true},
		{"wrong secret", "wrong", sig, payload, false},
		{"wrong signature", secret, "deadbeef", payload, false},
		{"wrong payload", secret, sig, []byte(`modified`), false},
		{"empty signature", secret, "", payload, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Verify(tt.secret, tt.signature, tt.payload)
			if got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Event matching
// ---------------------------------------------------------------------------

func TestMatchesEvent(t *testing.T) {
	tests := []struct {
		name   string
		events []string
		event  string
		want   bool
	}{
		{"exact match", []string{"content.published"}, "content.published", true},
		{"wildcard", []string{"*"}, "content.published", true},
		{"no match", []string{"content.updated"}, "content.published", false},
		{"empty events", []string{}, "content.published", false},
		{"nil events", nil, "content.published", false},
		{"match last in list", []string{"content.updated", "content.deleted", "content.published"}, "content.published", true},
		{"multiple no match", []string{"content.updated", "content.deleted"}, "content.published", false},
		{"wildcard among others", []string{"content.updated", "*"}, "anything", true},
		{"empty event string", []string{"content.published"}, "", false},
		{"empty event in list matches empty", []string{""}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesEvent(tt.events, tt.event)
			if got != tt.want {
				t.Errorf("matchesEvent(%v, %q) = %v, want %v", tt.events, tt.event, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Event constants
// ---------------------------------------------------------------------------

func TestEventConstants_NoDuplicates(t *testing.T) {
	events := []string{
		EventContentPublished,
		EventContentUnpublished,
		EventContentUpdated,
		EventContentScheduled,
		EventContentDeleted,
		EventLocalePublished,
		EventVersionCreated,
		EventAdminContentPublished,
		EventAdminContentUnpublished,
		EventAdminContentUpdated,
		EventAdminContentDeleted,
		EventUpdateAvailable,
	}
	seen := make(map[string]bool)
	for _, e := range events {
		if e == "" {
			t.Error("event constant is empty")
		}
		if seen[e] {
			t.Errorf("duplicate event constant: %q", e)
		}
		seen[e] = true
	}
}

// ---------------------------------------------------------------------------
// IP blocking (SSRF)
// ---------------------------------------------------------------------------

func TestIsBlockedIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		blocked bool
	}{
		{"loopback v4", "127.0.0.1", true},
		{"loopback v6", "::1", true},
		{"private 10.x", "10.0.0.1", true},
		{"private 172.16.x", "172.16.0.1", true},
		{"private 192.168.x", "192.168.1.1", true},
		{"link-local", "169.254.1.1", true},
		{"cloud metadata", "169.254.169.254", true},
		{"public 8.8.8.8", "8.8.8.8", false},
		{"public 1.1.1.1", "1.1.1.1", false},
		{"public 93.184.216.34", "93.184.216.34", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("failed to parse IP %q", tt.ip)
			}
			got := isBlockedIP(ip)
			if got != tt.blocked {
				t.Errorf("isBlockedIP(%s) = %v, want %v", tt.ip, got, tt.blocked)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// URL validation (SSRF)
// ---------------------------------------------------------------------------

func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		allowHTTP bool
		wantErr   bool
	}{
		{"http blocked when not allowed", "http://example.com/webhook", false, true},
		{"ftp unsupported", "ftp://example.com/file", false, true},
		{"javascript scheme", "javascript:alert(1)", false, true},
		{"data scheme", "data:text/html,hello", true, true},
		{"no scheme", "://missing", false, true},
		{"localhost blocked", "https://localhost/webhook", true, true},
		{"loopback blocked", "http://127.0.0.1:8080/webhook", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWebhookURL(tt.url, tt.allowHTTP)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWebhookURL(%q, %v) error = %v, wantErr %v", tt.url, tt.allowHTTP, err, tt.wantErr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestNew(t *testing.T) {
	mock := &mockDriver{}
	cfg := config.Config{
		Webhook_Workers: 3,
		Webhook_Timeout: 15,
	}
	d := New(mock, cfg)

	if d == nil {
		t.Fatal("New returned nil")
	}
	if cap(d.ch) != 30 {
		t.Errorf("channel buffer = %d, want 30 (3 workers * 10)", cap(d.ch))
	}
	if d.httpCli == nil {
		t.Fatal("httpCli is nil")
	}
	if d.httpCli.Timeout != 15*time.Second {
		t.Errorf("httpCli.Timeout = %v, want 15s", d.httpCli.Timeout)
	}
}

func TestNew_DefaultConfig(t *testing.T) {
	mock := &mockDriver{}
	cfg := config.Config{}
	d := New(mock, cfg)

	if cap(d.ch) != 40 {
		t.Errorf("channel buffer = %d, want 40 (4 default workers * 10)", cap(d.ch))
	}
	if d.httpCli.Timeout != 10*time.Second {
		t.Errorf("httpCli.Timeout = %v, want 10s (default)", d.httpCli.Timeout)
	}
}

func TestNew_NoRedirectFollowing(t *testing.T) {
	d := New(&mockDriver{}, testConfig())

	if d.httpCli.CheckRedirect == nil {
		t.Fatal("CheckRedirect is nil, redirects may be followed")
	}
	err := d.httpCli.CheckRedirect(nil, nil)
	if !errors.Is(err, http.ErrUseLastResponse) {
		t.Errorf("CheckRedirect returned %v, want http.ErrUseLastResponse", err)
	}
}

// ---------------------------------------------------------------------------
// Dispatch
// ---------------------------------------------------------------------------

func TestDispatch_NilReceiver(t *testing.T) {
	var d *Dispatcher
	d.Dispatch(context.Background(), EventContentPublished, map[string]any{"id": "123"})
}

func TestDispatch_NoActiveWebhooks(t *testing.T) {
	mock := &mockDriver{activeWebhooks: nil}
	d := New(mock, testConfig())

	d.Dispatch(context.Background(), EventContentPublished, map[string]any{"id": "123"})

	if len(mock.getCreateCalls()) != 0 {
		t.Error("expected 0 create calls when no webhooks registered")
	}
}

func TestDispatch_EmptyWebhooksList(t *testing.T) {
	mock := &mockDriver{activeWebhooks: &[]db.Webhook{}}
	d := New(mock, testConfig())

	d.Dispatch(context.Background(), EventContentPublished, map[string]any{"id": "123"})

	if len(mock.getCreateCalls()) != 0 {
		t.Error("expected 0 create calls for empty webhook list")
	}
}

func TestDispatch_ActiveWebhooksError(t *testing.T) {
	mock := &mockDriver{activeWebhooksErr: errors.New("db connection failed")}
	d := New(mock, testConfig())

	d.Dispatch(context.Background(), EventContentPublished, map[string]any{"id": "123"})

	if len(mock.getCreateCalls()) != 0 {
		t.Error("expected 0 create calls when ListActiveWebhooks fails")
	}
}

func TestDispatch_MatchingWebhook(t *testing.T) {
	whID := types.NewWebhookID()
	delID := types.NewWebhookDeliveryID()
	mock := &mockDriver{
		activeWebhooks: &[]db.Webhook{
			{
				WebhookID: whID,
				Name:      "test hook",
				URL:       "https://example.com/webhook",
				Secret:    "secret",
				Events:    []string{EventContentPublished},
				IsActive:  true,
			},
		},
		createdDelivery: &db.WebhookDelivery{
			DeliveryID: delID,
			WebhookID:  whID,
			Event:      EventContentPublished,
			Status:     db.DeliveryStatusPending,
		},
	}
	d := New(mock, testConfig())

	d.Dispatch(context.Background(), EventContentPublished, map[string]any{"id": "123"})

	calls := mock.getCreateCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 create call, got %d", len(calls))
	}
	if calls[0].Event != EventContentPublished {
		t.Errorf("event = %q, want %q", calls[0].Event, EventContentPublished)
	}
	if calls[0].WebhookID != whID {
		t.Errorf("webhook_id = %v, want %v", calls[0].WebhookID, whID)
	}
	if calls[0].Status != db.DeliveryStatusPending {
		t.Errorf("status = %q, want %q", calls[0].Status, db.DeliveryStatusPending)
	}
	if calls[0].Payload == "" {
		t.Error("payload is empty")
	}
}

func TestDispatch_NoMatchingEvents(t *testing.T) {
	mock := &mockDriver{
		activeWebhooks: &[]db.Webhook{
			{
				WebhookID: types.NewWebhookID(),
				Events:    []string{EventContentDeleted},
				IsActive:  true,
			},
		},
	}
	d := New(mock, testConfig())

	d.Dispatch(context.Background(), EventContentPublished, map[string]any{"id": "123"})

	if len(mock.getCreateCalls()) != 0 {
		t.Error("expected 0 create calls for non-matching event")
	}
}

func TestDispatch_WildcardWebhook(t *testing.T) {
	whID := types.NewWebhookID()
	mock := &mockDriver{
		activeWebhooks: &[]db.Webhook{
			{WebhookID: whID, Events: []string{"*"}, IsActive: true},
		},
		createdDelivery: &db.WebhookDelivery{
			DeliveryID: types.NewWebhookDeliveryID(),
			WebhookID:  whID,
			Status:     db.DeliveryStatusPending,
		},
	}
	d := New(mock, testConfig())

	d.Dispatch(context.Background(), EventContentPublished, map[string]any{"id": "123"})

	if len(mock.getCreateCalls()) != 1 {
		t.Error("wildcard webhook should match any event")
	}
}

func TestDispatch_MultipleWebhooks_OnlyMatchingCreated(t *testing.T) {
	mock := &mockDriver{
		activeWebhooks: &[]db.Webhook{
			{WebhookID: types.NewWebhookID(), Events: []string{EventContentPublished}, IsActive: true},
			{WebhookID: types.NewWebhookID(), Events: []string{EventContentUpdated}, IsActive: true},
			{WebhookID: types.NewWebhookID(), Events: []string{EventContentPublished, EventContentDeleted}, IsActive: true},
		},
		createdDelivery: &db.WebhookDelivery{
			DeliveryID: types.NewWebhookDeliveryID(),
			Status:     db.DeliveryStatusPending,
		},
	}
	d := New(mock, testConfig())

	d.Dispatch(context.Background(), EventContentPublished, map[string]any{"id": "123"})

	calls := mock.getCreateCalls()
	if len(calls) != 2 {
		t.Errorf("expected 2 create calls (webhooks 1 and 3 match), got %d", len(calls))
	}
}

func TestDispatch_CreateDeliveryError_ContinuesProcessing(t *testing.T) {
	mock := &mockDriver{
		activeWebhooks: &[]db.Webhook{
			{WebhookID: types.NewWebhookID(), Events: []string{"*"}, IsActive: true},
			{WebhookID: types.NewWebhookID(), Events: []string{"*"}, IsActive: true},
		},
		createDeliveryErr: errors.New("db write failed"),
	}
	d := New(mock, testConfig())

	d.Dispatch(context.Background(), EventContentPublished, map[string]any{"id": "123"})

	calls := mock.getCreateCalls()
	if len(calls) != 2 {
		t.Errorf("expected 2 create attempts despite errors, got %d", len(calls))
	}
}

func TestDispatch_ChannelFull_NonBlocking(t *testing.T) {
	whID := types.NewWebhookID()
	mock := &mockDriver{
		activeWebhooks: &[]db.Webhook{
			{WebhookID: whID, Events: []string{"*"}, IsActive: true},
		},
		createdDelivery: &db.WebhookDelivery{
			DeliveryID: types.NewWebhookDeliveryID(),
			WebhookID:  whID,
			Status:     db.DeliveryStatusPending,
		},
	}

	cfg := config.Config{Webhook_Workers: 1, Webhook_Max_Retries: 3, Webhook_Allow_HTTP: true}
	d := New(mock, cfg)

	// Fill the channel to capacity without starting workers.
	for range cap(d.ch) {
		d.ch <- dispatchJob{}
	}

	done := make(chan struct{})
	go func() {
		d.Dispatch(context.Background(), EventContentPublished, map[string]any{"id": "123"})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Dispatch blocked on full channel")
	}

	if len(mock.getCreateCalls()) != 1 {
		t.Error("delivery should still be created even when channel is full")
	}
}

func TestDispatch_PayloadStructure(t *testing.T) {
	whID := types.NewWebhookID()
	mock := &mockDriver{
		activeWebhooks: &[]db.Webhook{
			{WebhookID: whID, Events: []string{"*"}, IsActive: true},
		},
		createdDelivery: &db.WebhookDelivery{
			DeliveryID: types.NewWebhookDeliveryID(),
			WebhookID:  whID,
			Status:     db.DeliveryStatusPending,
		},
	}
	d := New(mock, testConfig())

	data := map[string]any{"content_id": "abc123", "title": "Hello"}
	d.Dispatch(context.Background(), EventContentPublished, data)

	calls := mock.getCreateCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 create call, got %d", len(calls))
	}

	var payload Payload
	if err := json.Unmarshal([]byte(calls[0].Payload), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload.Event != EventContentPublished {
		t.Errorf("payload.Event = %q, want %q", payload.Event, EventContentPublished)
	}
	if payload.ID == "" {
		t.Error("payload.ID is empty")
	}
	if payload.OccurredAt.IsZero() {
		t.Error("payload.OccurredAt is zero")
	}
	if payload.Data["content_id"] != "abc123" {
		t.Errorf("payload.Data[content_id] = %v, want %q", payload.Data["content_id"], "abc123")
	}
}

func TestDispatch_Concurrent(t *testing.T) {
	whID := types.NewWebhookID()
	mock := &mockDriver{
		activeWebhooks: &[]db.Webhook{
			{WebhookID: whID, Events: []string{"*"}, IsActive: true},
		},
		createdDelivery: &db.WebhookDelivery{
			DeliveryID: types.NewWebhookDeliveryID(),
			WebhookID:  whID,
			Status:     db.DeliveryStatusPending,
		},
	}
	d := New(mock, testConfig())

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.Dispatch(context.Background(), EventContentPublished, map[string]any{"id": "test"})
		}()
	}
	wg.Wait()

	calls := mock.getCreateCalls()
	if len(calls) != 10 {
		t.Errorf("expected 10 create calls from concurrent dispatch, got %d", len(calls))
	}
}

// ---------------------------------------------------------------------------
// Deliver (SSRF at delivery time)
// ---------------------------------------------------------------------------

func TestDeliver_SSRFBlocked(t *testing.T) {
	mock := &mockDriver{}
	d := New(mock, testConfig())

	job := dispatchJob{
		delivery: db.WebhookDelivery{
			DeliveryID: types.NewWebhookDeliveryID(),
			WebhookID:  types.NewWebhookID(),
			Event:      EventContentPublished,
			Payload:    `{"event":"content.published"}`,
		},
		webhook: db.Webhook{
			URL:    "http://127.0.0.1:9999/webhook",
			Secret: "secret",
		},
	}

	d.deliver(context.Background(), job)

	calls := mock.getUpdateCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 update call for SSRF block, got %d", len(calls))
	}
	if calls[0].Status != db.DeliveryStatusFailed {
		t.Errorf("status = %q, want %q", calls[0].Status, db.DeliveryStatusFailed)
	}
	if calls[0].LastError == "" {
		t.Error("expected LastError to describe URL validation failure")
	}
}

func TestDeliver_PrivateIPBlocked(t *testing.T) {
	mock := &mockDriver{}
	d := New(mock, testConfig())

	job := dispatchJob{
		delivery: db.WebhookDelivery{
			DeliveryID: types.NewWebhookDeliveryID(),
			Payload:    `{"event":"test"}`,
		},
		webhook: db.Webhook{
			URL:    "http://10.0.0.1/webhook",
			Secret: "secret",
		},
	}

	d.deliver(context.Background(), job)

	calls := mock.getUpdateCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 update for private IP block, got %d", len(calls))
	}
	if calls[0].Status != db.DeliveryStatusFailed {
		t.Errorf("status = %q, want %q", calls[0].Status, db.DeliveryStatusFailed)
	}
}

// ---------------------------------------------------------------------------
// Failure handling
// ---------------------------------------------------------------------------

func TestHandleFailure_Backoff(t *testing.T) {
	tests := []struct {
		name         string
		attempts     int64
		maxRetries   int
		expectStatus string
	}{
		{"first failure retries", 0, 3, db.DeliveryStatusRetrying},
		{"second failure retries", 1, 3, db.DeliveryStatusRetrying},
		{"third failure fails permanently", 2, 3, db.DeliveryStatusFailed},
		{"max retries 1 fails immediately", 0, 1, db.DeliveryStatusFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDriver{}
			cfg := config.Config{Webhook_Workers: 1, Webhook_Max_Retries: tt.maxRetries}
			d := New(mock, cfg)

			del := db.WebhookDelivery{
				DeliveryID: types.NewWebhookDeliveryID(),
				WebhookID:  types.NewWebhookID(),
				Attempts:   tt.attempts,
			}

			d.handleFailure(context.Background(), del, 500, "server error")

			calls := mock.getUpdateCalls()
			if len(calls) != 1 {
				t.Fatalf("expected 1 update call, got %d", len(calls))
			}
			if calls[0].Status != tt.expectStatus {
				t.Errorf("status = %q, want %q", calls[0].Status, tt.expectStatus)
			}
			if calls[0].Attempts != tt.attempts+1 {
				t.Errorf("attempts = %d, want %d", calls[0].Attempts, tt.attempts+1)
			}
		})
	}
}

func TestHandleFailure_BackoffTiming(t *testing.T) {
	tests := []struct {
		name       string
		attempts   int64
		minBackoff time.Duration
		maxBackoff time.Duration
	}{
		{"first retry 1min", 0, 50 * time.Second, 70 * time.Second},
		{"second retry 5min", 1, 290 * time.Second, 310 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDriver{}
			cfg := config.Config{Webhook_Workers: 1, Webhook_Max_Retries: 5}
			d := New(mock, cfg)

			del := db.WebhookDelivery{
				DeliveryID: types.NewWebhookDeliveryID(),
				Attempts:   tt.attempts,
			}

			before := time.Now().UTC()
			d.handleFailure(context.Background(), del, 500, "error")

			calls := mock.getUpdateCalls()
			if len(calls) != 1 {
				t.Fatalf("expected 1 update call, got %d", len(calls))
			}

			nextRetry, err := time.Parse(time.RFC3339, calls[0].NextRetryAt)
			if err != nil {
				t.Fatalf("failed to parse NextRetryAt %q: %v", calls[0].NextRetryAt, err)
			}

			backoff := nextRetry.Sub(before)
			if backoff < tt.minBackoff || backoff > tt.maxBackoff {
				t.Errorf("backoff = %v, want between %v and %v", backoff, tt.minBackoff, tt.maxBackoff)
			}
		})
	}
}

func TestHandleFailure_RetryingSetsNextRetryAt(t *testing.T) {
	mock := &mockDriver{}
	cfg := config.Config{Webhook_Workers: 1, Webhook_Max_Retries: 3}
	d := New(mock, cfg)

	del := db.WebhookDelivery{
		DeliveryID: types.NewWebhookDeliveryID(),
		Attempts:   0,
	}

	d.handleFailure(context.Background(), del, 503, "service unavailable")

	calls := mock.getUpdateCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 update call, got %d", len(calls))
	}
	if calls[0].NextRetryAt == "" {
		t.Error("NextRetryAt should be set for retrying status")
	}
	if calls[0].CompletedAt != "" {
		t.Error("CompletedAt should not be set for retrying status")
	}
}

func TestFailDelivery(t *testing.T) {
	mock := &mockDriver{}
	d := New(mock, testConfig())

	del := db.WebhookDelivery{
		DeliveryID: types.NewWebhookDeliveryID(),
		WebhookID:  types.NewWebhookID(),
		Attempts:   2,
	}

	d.failDelivery(context.Background(), del, 503, "service unavailable")

	calls := mock.getUpdateCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 update call, got %d", len(calls))
	}
	if calls[0].Status != db.DeliveryStatusFailed {
		t.Errorf("status = %q, want %q", calls[0].Status, db.DeliveryStatusFailed)
	}
	if calls[0].Attempts != 3 {
		t.Errorf("attempts = %d, want 3", calls[0].Attempts)
	}
	if calls[0].LastStatusCode != 503 {
		t.Errorf("last_status_code = %d, want 503", calls[0].LastStatusCode)
	}
	if calls[0].LastError != "service unavailable" {
		t.Errorf("last_error = %q, want %q", calls[0].LastError, "service unavailable")
	}
	if calls[0].CompletedAt == "" {
		t.Error("CompletedAt should be set for permanent failure")
	}
}

// ---------------------------------------------------------------------------
// Retries
// ---------------------------------------------------------------------------

func TestProcessRetries_NilReceiver(t *testing.T) {
	var d *Dispatcher
	d.ProcessRetries(context.Background())
}

func TestProcessRetries_NoPending(t *testing.T) {
	mock := &mockDriver{pendingRetries: nil}
	d := New(mock, testConfig())

	d.ProcessRetries(context.Background())

	if len(d.ch) != 0 {
		t.Error("no jobs should be enqueued when no retries pending")
	}
}

func TestProcessRetries_Error(t *testing.T) {
	mock := &mockDriver{pendingRetriesErr: errors.New("db error")}
	d := New(mock, testConfig())

	d.ProcessRetries(context.Background())

	if len(d.ch) != 0 {
		t.Error("no jobs should be enqueued when ListPendingRetries fails")
	}
}

func TestProcessRetries_ReEnqueues(t *testing.T) {
	whID := types.NewWebhookID()
	delID := types.NewWebhookDeliveryID()
	mock := &mockDriver{
		pendingRetries: &[]db.WebhookDelivery{
			{
				DeliveryID: delID,
				WebhookID:  whID,
				Event:      EventContentPublished,
				Attempts:   1,
			},
		},
		webhook: &db.Webhook{
			WebhookID: whID,
			URL:       "https://example.com/webhook",
			Secret:    "secret",
			Events:    []string{EventContentPublished},
		},
	}
	d := New(mock, testConfig())

	d.ProcessRetries(context.Background())

	if len(d.ch) != 1 {
		t.Errorf("expected 1 job in channel, got %d", len(d.ch))
	}

	job := <-d.ch
	if job.delivery.DeliveryID != delID {
		t.Errorf("delivery_id = %v, want %v", job.delivery.DeliveryID, delID)
	}
	if job.webhook.WebhookID != whID {
		t.Errorf("webhook_id = %v, want %v", job.webhook.WebhookID, whID)
	}
}

func TestProcessRetries_GetWebhookError_SkipsDelivery(t *testing.T) {
	mock := &mockDriver{
		pendingRetries: &[]db.WebhookDelivery{
			{
				DeliveryID: types.NewWebhookDeliveryID(),
				WebhookID:  types.NewWebhookID(),
				Event:      EventContentPublished,
			},
		},
		getWebhookErr: errors.New("webhook not found"),
	}
	d := New(mock, testConfig())

	d.ProcessRetries(context.Background())

	if len(d.ch) != 0 {
		t.Error("no jobs should be enqueued when GetWebhook fails")
	}
}

// ---------------------------------------------------------------------------
// Lifecycle
// ---------------------------------------------------------------------------

func TestStartShutdown(t *testing.T) {
	d := New(&mockDriver{}, testConfig())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d.Start(ctx)
	d.Shutdown()
}

func TestShutdown_ContextCancellation(t *testing.T) {
	d := New(&mockDriver{}, testConfig())

	ctx, cancel := context.WithCancel(context.Background())
	d.Start(ctx)
	cancel()
	d.Shutdown()
}
