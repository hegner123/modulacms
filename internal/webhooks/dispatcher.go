package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/oklog/ulid/v2"
)

type dispatchJob struct {
	delivery db.WebhookDelivery
	webhook  db.Webhook
}

// Dispatcher manages webhook delivery workers and retry processing.
type Dispatcher struct {
	driver  db.DbDriver
	cfg     config.Config
	ch      chan dispatchJob
	done    chan struct{}
	wg      sync.WaitGroup
	httpCli *http.Client
}

// New creates a new Dispatcher. Does not start workers until Start is called.
func New(driver db.DbDriver, cfg config.Config) *Dispatcher {
	workers := cfg.WebhookWorkers()
	bufSize := workers * 10

	return &Dispatcher{
		driver: driver,
		cfg:    cfg,
		ch:     make(chan dispatchJob, bufSize),
		done:   make(chan struct{}),
		httpCli: &http.Client{
			Timeout: time.Duration(cfg.WebhookTimeout()) * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse // do not follow redirects
			},
		},
	}
}

// Start launches worker goroutines and the internal retry ticker.
func (d *Dispatcher) Start(ctx context.Context) {
	workers := d.cfg.WebhookWorkers()
	for range workers {
		d.wg.Add(1)
		go d.worker(ctx)
	}

	// Internal retry goroutine — runs every 60 seconds.
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				d.ProcessRetries(ctx)
			case <-d.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	utility.DefaultLogger.Info("Webhook dispatcher started", "workers", workers)
}

// Shutdown stops the dispatcher. In-flight deliveries are allowed to complete
// with a 10-second timeout. Unprocessed channel items remain in DB for retry.
func (d *Dispatcher) Shutdown() {
	close(d.done)

	shutdownDone := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		utility.DefaultLogger.Info("Webhook dispatcher shut down cleanly")
	case <-time.After(10 * time.Second):
		utility.DefaultLogger.Warn("Webhook dispatcher shutdown timed out after 10s", nil)
	}
}

// Dispatch creates a delivery record for each active webhook matching the event
// and enqueues them for async delivery. Safe to call on a nil receiver.
func (d *Dispatcher) Dispatch(ctx context.Context, event string, data map[string]any) {
	if d == nil {
		return
	}

	utility.DefaultLogger.Info("Webhook event triggered", "event", event)

	webhooks, err := d.driver.ListActiveWebhooks()
	if err != nil {
		utility.DefaultLogger.Error("Failed to list active webhooks", err)
		return
	}
	if webhooks == nil {
		utility.DefaultLogger.Debug("No active webhooks registered", "event", event)
		return
	}

	payloadBytes, err := json.Marshal(Payload{
		ID:         ulid.Make().String(),
		Event:      event,
		OccurredAt: time.Now().UTC(),
		Data:       data,
	})
	if err != nil {
		utility.DefaultLogger.Error("Failed to marshal webhook payload", err)
		return
	}

	matched := 0
	for _, wh := range *webhooks {
		if !matchesEvent(wh.Events, event) {
			continue
		}
		matched++

		delivery, createErr := d.driver.CreateWebhookDelivery(ctx, db.CreateWebhookDeliveryParams{
			WebhookID: wh.WebhookID,
			Event:     event,
			Payload:   string(payloadBytes),
			Status:    db.DeliveryStatusPending,
			Attempts:  0,
			CreatedAt: types.TimestampNow(),
		})
		if createErr != nil {
			utility.DefaultLogger.Error("Failed to create webhook delivery", createErr, "webhook_id", wh.WebhookID)
			continue
		}

		utility.DefaultLogger.Info("Webhook delivery enqueued", "event", event, "webhook_id", wh.WebhookID, "delivery_id", delivery.DeliveryID, "url", wh.URL)

		// Non-blocking send — if the channel is full, the retry processor will pick it up.
		select {
		case d.ch <- dispatchJob{delivery: *delivery, webhook: wh}:
		default:
			utility.DefaultLogger.Warn("Webhook dispatch channel full, delivery will be retried", nil, "delivery_id", delivery.DeliveryID)
		}
	}

	if matched == 0 {
		utility.DefaultLogger.Debug("No webhooks matched event", "event", event, "total_webhooks", len(*webhooks))
	}
}

// ProcessRetries queries for deliveries with status='retrying' whose next_retry_at
// has passed, and re-enqueues them for delivery.
func (d *Dispatcher) ProcessRetries(ctx context.Context) {
	if d == nil {
		return
	}

	deliveries, err := d.driver.ListPendingRetries(types.TimestampNow(), 100)
	if err != nil {
		utility.DefaultLogger.Error("Failed to list pending retries", err)
		return
	}
	if deliveries == nil {
		return
	}

	utility.DefaultLogger.Info("Processing webhook retries", "count", len(*deliveries))

	for _, del := range *deliveries {
		wh, whErr := d.driver.GetWebhook(del.WebhookID)
		if whErr != nil {
			utility.DefaultLogger.Error("Failed to get webhook for retry", whErr, "webhook_id", del.WebhookID)
			continue
		}

		utility.DefaultLogger.Info("Retrying webhook delivery", "event", del.Event, "webhook_id", del.WebhookID, "delivery_id", del.DeliveryID, "attempt", del.Attempts+1)

		select {
		case d.ch <- dispatchJob{delivery: del, webhook: *wh}:
		default:
			utility.DefaultLogger.Warn("Webhook dispatch channel full during retry", nil, "delivery_id", del.DeliveryID)
		}
	}
}

// worker processes dispatch jobs from the channel.
func (d *Dispatcher) worker(ctx context.Context) {
	defer d.wg.Done()
	for {
		select {
		case job, ok := <-d.ch:
			if !ok {
				return
			}
			d.deliver(ctx, job)
		case <-d.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

// deliver executes a single webhook HTTP POST.
func (d *Dispatcher) deliver(ctx context.Context, job dispatchJob) {
	wh := job.webhook
	del := job.delivery

	// SSRF: re-validate URL at delivery time (DNS rebinding defense).
	if err := ValidateWebhookURL(wh.URL, d.cfg.WebhookAllowHTTP()); err != nil {
		d.failDelivery(ctx, del, 0, fmt.Sprintf("URL validation failed: %v", err))
		return
	}

	signature := Sign(wh.Secret, []byte(del.Payload))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader([]byte(del.Payload)))
	if err != nil {
		d.failDelivery(ctx, del, 0, fmt.Sprintf("request creation failed: %v", err))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ModulaCMS-Signature", signature)
	req.Header.Set("X-ModulaCMS-Event", del.Event)
	req.Header.Set("User-Agent", "ModulaCMS-Webhook/1.0")

	// Add custom headers from webhook config.
	for k, v := range wh.Headers {
		req.Header.Set(k, v)
	}

	resp, err := d.httpCli.Do(req)
	if err != nil {
		d.handleFailure(ctx, del, 0, fmt.Sprintf("HTTP request failed: %v", err))
		return
	}
	resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		utility.DefaultLogger.Info("Webhook delivered successfully", "event", del.Event, "webhook_id", wh.WebhookID, "delivery_id", del.DeliveryID, "url", wh.URL, "status_code", resp.StatusCode)
		now := time.Now().UTC().Format(time.RFC3339)
		updateErr := d.driver.UpdateWebhookDeliveryStatus(ctx, db.UpdateWebhookDeliveryStatusParams{
			Status:         db.DeliveryStatusSuccess,
			Attempts:       del.Attempts + 1,
			LastStatusCode: int64(resp.StatusCode),
			CompletedAt:    now,
			DeliveryID:     del.DeliveryID,
		})
		if updateErr != nil {
			utility.DefaultLogger.Error("Failed to update delivery status to success", updateErr, "delivery_id", del.DeliveryID)
		}
		return
	}

	utility.DefaultLogger.Warn("Webhook delivery failed", nil, "event", del.Event, "webhook_id", wh.WebhookID, "delivery_id", del.DeliveryID, "url", wh.URL, "status_code", resp.StatusCode)
	d.handleFailure(ctx, del, int64(resp.StatusCode), fmt.Sprintf("HTTP %d", resp.StatusCode))
}

// handleFailure increments attempts and either marks as retrying or failed.
func (d *Dispatcher) handleFailure(ctx context.Context, del db.WebhookDelivery, statusCode int64, errMsg string) {
	attempts := del.Attempts + 1
	maxRetries := int64(d.cfg.WebhookMaxRetries())

	if attempts < maxRetries {
		// Exponential backoff: 1min, 5min, 30min.
		var backoff time.Duration
		switch attempts {
		case 1:
			backoff = 1 * time.Minute
		case 2:
			backoff = 5 * time.Minute
		default:
			backoff = 30 * time.Minute
		}
		nextRetry := time.Now().UTC().Add(backoff).Format(time.RFC3339)

		updateErr := d.driver.UpdateWebhookDeliveryStatus(ctx, db.UpdateWebhookDeliveryStatusParams{
			Status:         db.DeliveryStatusRetrying,
			Attempts:       attempts,
			LastStatusCode: statusCode,
			LastError:      errMsg,
			NextRetryAt:    nextRetry,
			DeliveryID:     del.DeliveryID,
		})
		if updateErr != nil {
			utility.DefaultLogger.Error("Failed to update delivery for retry", updateErr, "delivery_id", del.DeliveryID)
		}
		return
	}

	d.failDelivery(ctx, del, statusCode, errMsg)
}

// failDelivery marks a delivery as permanently failed.
func (d *Dispatcher) failDelivery(ctx context.Context, del db.WebhookDelivery, statusCode int64, errMsg string) {
	now := time.Now().UTC().Format(time.RFC3339)
	updateErr := d.driver.UpdateWebhookDeliveryStatus(ctx, db.UpdateWebhookDeliveryStatusParams{
		Status:         db.DeliveryStatusFailed,
		Attempts:       del.Attempts + 1,
		LastStatusCode: statusCode,
		LastError:      errMsg,
		CompletedAt:    now,
		DeliveryID:     del.DeliveryID,
	})
	if updateErr != nil {
		utility.DefaultLogger.Error("Failed to update delivery status to failed", updateErr, "delivery_id", del.DeliveryID)
	}
}

// matchesEvent checks if the webhook's event list includes the given event.
// A wildcard "*" matches all events.
func matchesEvent(events []string, event string) bool {
	for _, e := range events {
		if e == "*" || e == event {
			return true
		}
	}
	return false
}

// ValidateWebhookURL checks that a URL is safe for webhook delivery (SSRF protection).
// Blocks loopback, private, link-local, and cloud metadata addresses.
func ValidateWebhookURL(rawURL string, allowHTTP bool) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Scheme check.
	if parsed.Scheme == "http" && !allowHTTP {
		return fmt.Errorf("HTTP URLs are not allowed (set webhook_allow_http=true for dev)")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported scheme %q, must be http or https", parsed.Scheme)
	}

	// Resolve hostname to IP addresses.
	hostname := parsed.Hostname()
	ips, err := net.LookupHost(hostname)
	if err != nil {
		return fmt.Errorf("DNS resolution failed for %q: %w", hostname, err)
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if isBlockedIP(ip) {
			return fmt.Errorf("IP %s for host %q is blocked (private/loopback/link-local/metadata)", ipStr, hostname)
		}
	}

	return nil
}

// isBlockedIP returns true if the IP is loopback, private, link-local, or a cloud metadata address.
func isBlockedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Cloud metadata endpoint: 169.254.169.254
	metadata := net.ParseIP("169.254.169.254")
	if ip.Equal(metadata) {
		return true
	}

	return false
}
