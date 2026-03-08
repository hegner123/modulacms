package plugin

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	db "github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/middleware"
	"golang.org/x/time/rate"
)

// RequestEngineConfig configures the outbound request engine.
type RequestEngineConfig struct {
	DefaultTimeoutSec       int   // default 10, max per-request timeout
	MaxResponseBytes        int64 // default 1MB (1_048_576)
	MaxRequestBodyBytes     int64 // default 1MB (1_048_576)
	MaxRequestsPerMin       int   // default 60, per plugin per domain
	GlobalMaxRequestsPerMin int   // default 600, aggregate; 0 = unlimited
	CBMaxFailures           int   // default 5, consecutive failures to trip
	CBResetIntervalSec      int   // default 60, seconds before half-open probe
	AllowLocalhost          bool  // default false; development only
}

// requestEngineDefaults fills zero-value config fields with defaults.
func requestEngineDefaults(cfg *RequestEngineConfig) {
	if cfg.DefaultTimeoutSec <= 0 {
		cfg.DefaultTimeoutSec = 10
	}
	if cfg.MaxResponseBytes <= 0 {
		cfg.MaxResponseBytes = 1_048_576
	}
	if cfg.MaxRequestBodyBytes <= 0 {
		cfg.MaxRequestBodyBytes = 1_048_576
	}
	if cfg.MaxRequestsPerMin <= 0 {
		cfg.MaxRequestsPerMin = 60
	}
	if cfg.GlobalMaxRequestsPerMin < 0 {
		cfg.GlobalMaxRequestsPerMin = 600
	}
	if cfg.CBMaxFailures <= 0 {
		cfg.CBMaxFailures = 5
	}
	if cfg.CBResetIntervalSec <= 0 {
		cfg.CBResetIntervalSec = 60
	}
}

// domainCB tracks per-plugin-per-domain circuit breaker state.
type domainCB struct {
	consecutiveFailures int
	disabled            bool
	lastFailure         time.Time
}

// rateLimiterEntry tracks per-plugin-per-domain rate limiter state.
type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RequestRegistration represents a registered outbound domain with its
// approval status. Used by ListRequests() for the admin API response.
type RequestRegistration struct {
	PluginName    string `json:"plugin_name"`
	Domain        string `json:"domain"`
	Description   string `json:"description"`
	Approved      bool   `json:"approved"`
	PluginVersion string `json:"plugin_version"`
}

// RequestEngine handles outbound HTTP requests from plugins with SSRF
// protection, per-domain rate limiting, circuit breaking, and admin approval.
//
// Key format invariant: all map keys use "plugin:domain" format. Safe because
// plugin names are [a-z0-9_] and domains are [a-zA-Z0-9.-] -- neither includes ":".
type RequestEngine struct {
	approved        map[string]bool // "plugin:domain" -> approved
	mu              sync.RWMutex    // protects approved map
	client          *http.Client
	rateLimiters    map[string]*rateLimiterEntry // "plugin:domain" -> token bucket
	rateMu          sync.Mutex
	globalLimiter   *rate.Limiter // aggregate cap; nil when disabled (0)
	circuitBreakers map[string]*domainCB
	cbMu            sync.RWMutex
	cleanupDone     chan struct{}
	pool            *sql.DB
	dialect         db.Dialect
	cfg             RequestEngineConfig
}

// NewRequestEngine creates a new outbound request engine.
func NewRequestEngine(pool *sql.DB, dialect db.Dialect, cfg RequestEngineConfig) *RequestEngine {
	requestEngineDefaults(&cfg)

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DialContext:         ssrfSafeDialer(cfg).DialContext,
	}

	client := &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // no redirect following
		},
	}

	var globalLimiter *rate.Limiter
	if cfg.GlobalMaxRequestsPerMin > 0 {
		globalRate := rate.Limit(float64(cfg.GlobalMaxRequestsPerMin) / 60.0)
		burst := cfg.GlobalMaxRequestsPerMin / 60
		if burst > 5 {
			burst = 5
		}
		if burst < 1 {
			burst = 1
		}
		globalLimiter = rate.NewLimiter(globalRate, burst)
	}

	e := &RequestEngine{
		approved:        make(map[string]bool),
		client:          client,
		rateLimiters:    make(map[string]*rateLimiterEntry),
		globalLimiter:   globalLimiter,
		circuitBreakers: make(map[string]*domainCB),
		cleanupDone:     make(chan struct{}),
		pool:            pool,
		dialect:         dialect,
		cfg:             cfg,
	}

	go e.cleanupRateLimiters()

	return e
}

// Execute performs an outbound HTTP request with all safety checks.
// Returns (response_map, nil) on success, or (nil, error) on operational failure.
// The error is converted to {error = "..."} by the Lua API layer.
func (e *RequestEngine) Execute(ctx context.Context, pluginName, method, urlStr string, opts OutboundRequestOpts) (map[string]any, error) {
	// 1. Parse and validate URL.
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %s", err.Error())
	}
	if parsedURL.User != nil {
		return nil, fmt.Errorf("URLs with userinfo are not allowed")
	}

	hostname := strings.ToLower(parsedURL.Hostname())

	// Validate scheme: HTTPS required; HTTP only for localhost when allowed.
	switch parsedURL.Scheme {
	case "https":
		// always allowed
	case "http":
		if !e.cfg.AllowLocalhost || (hostname != "127.0.0.1" && hostname != "::1") {
			return nil, fmt.Errorf("HTTP scheme requires HTTPS (HTTP only allowed to localhost when AllowLocalhost is enabled)")
		}
	default:
		return nil, fmt.Errorf("unsupported URL scheme %q (only https is allowed)", parsedURL.Scheme)
	}

	// 2. Domain approval check.
	key := pluginName + ":" + hostname
	e.mu.RLock()
	approved := e.approved[key]
	e.mu.RUnlock()
	if !approved {
		return nil, fmt.Errorf("domain not approved: %s", hostname)
	}

	// 3. Circuit breaker check.
	e.cbMu.RLock()
	cb := e.circuitBreakers[key]
	e.cbMu.RUnlock()

	if cb != nil && cb.disabled {
		resetInterval := time.Duration(e.cfg.CBResetIntervalSec) * time.Second
		if time.Since(cb.lastFailure) < resetInterval {
			return nil, fmt.Errorf("circuit breaker open for domain: %s", hostname)
		}
		// Half-open: allow one probe request (fall through to execute).
	}

	// 4. Per-domain rate limiter check.
	e.rateMu.Lock()
	rl, exists := e.rateLimiters[key]
	if !exists {
		perSecRate := rate.Limit(float64(e.cfg.MaxRequestsPerMin) / 60.0)
		burst := e.cfg.MaxRequestsPerMin / 60
		if burst > 5 {
			burst = 5
		}
		if burst < 1 {
			burst = 1
		}
		rl = &rateLimiterEntry{
			limiter:  rate.NewLimiter(perSecRate, burst),
			lastSeen: time.Now(),
		}
		e.rateLimiters[key] = rl
	}
	rl.lastSeen = time.Now()
	e.rateMu.Unlock()

	if !rl.limiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded for domain: %s", hostname)
	}

	// 5. Global rate limiter check.
	if e.globalLimiter != nil && !e.globalLimiter.Allow() {
		return nil, fmt.Errorf("global outbound request rate limit exceeded")
	}

	// 6. Serialize body.
	var bodyReader io.Reader
	var contentType string

	if opts.JSONBody != nil {
		jsonBytes, marshalErr := json.Marshal(opts.JSONBody)
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to marshal JSON body: %s", marshalErr.Error())
		}
		if int64(len(jsonBytes)) > e.cfg.MaxRequestBodyBytes {
			return nil, fmt.Errorf("request body exceeds maximum size (%d bytes)", e.cfg.MaxRequestBodyBytes)
		}
		bodyReader = bytes.NewReader(jsonBytes)
		contentType = "application/json"
	} else if opts.Body != "" {
		if int64(len(opts.Body)) > e.cfg.MaxRequestBodyBytes {
			return nil, fmt.Errorf("request body exceeds maximum size (%d bytes)", e.cfg.MaxRequestBodyBytes)
		}
		bodyReader = strings.NewReader(opts.Body)
	}

	// 7. Build request with timeout.
	timeoutSec := e.cfg.DefaultTimeoutSec
	if opts.Timeout > 0 && opts.Timeout < timeoutSec {
		timeoutSec = opts.Timeout
	}
	if timeoutSec < 1 {
		timeoutSec = 1
	}

	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	req, reqErr := http.NewRequestWithContext(reqCtx, method, urlStr, bodyReader)
	if reqErr != nil {
		return nil, fmt.Errorf("failed to create request: %s", reqErr.Error())
	}

	// 8. Set headers.
	req.Header.Set("User-Agent", "ModulaCMS-Plugin/1.0")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	// Forward correlation ID if present.
	if requestID := middleware.RequestIDFromContext(ctx); requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	// 9. Execute request.
	start := time.Now()
	resp, doErr := e.client.Do(req)
	durationMs := float64(time.Since(start).Milliseconds())

	if doErr != nil {
		// Record circuit breaker failure.
		e.recordCBFailure(key)

		// Categorize the error for a user-friendly message.
		errMsg := categorizeHTTPError(doErr, hostname, timeoutSec)

		RecordOutboundRequest(pluginName, method, hostname, 0, durationMs)
		return nil, fmt.Errorf("%s", errMsg)
	}
	defer resp.Body.Close()

	// 10. Read response body with size limit.
	limitedReader := io.LimitReader(resp.Body, e.cfg.MaxResponseBytes+1)
	bodyBytes, readErr := io.ReadAll(limitedReader)
	if readErr != nil {
		e.recordCBFailure(key)
		RecordOutboundRequest(pluginName, method, hostname, resp.StatusCode, durationMs)
		return nil, fmt.Errorf("failed to read response body: %s", readErr.Error())
	}
	if int64(len(bodyBytes)) > e.cfg.MaxResponseBytes {
		e.recordCBFailure(key)
		RecordOutboundRequest(pluginName, method, hostname, resp.StatusCode, durationMs)
		return nil, fmt.Errorf("response exceeded maximum size (%d bytes)", e.cfg.MaxResponseBytes)
	}

	// 11. Record circuit breaker result.
	if resp.StatusCode >= 500 {
		e.recordCBFailure(key)
	} else {
		e.recordCBSuccess(key)
	}

	// 12. Record metrics (MUST NOT log headers, body, or credentials).
	RecordOutboundRequest(pluginName, method, hostname, resp.StatusCode, durationMs)

	// 13. Build response map.
	result := map[string]any{
		"status": resp.StatusCode,
		"body":   string(bodyBytes),
	}

	// Flatten response headers.
	headers := make(map[string]string, len(resp.Header))
	for k := range resp.Header {
		headers[strings.ToLower(k)] = resp.Header.Get(k)
	}
	result["headers"] = headers

	// Opt-in JSON parsing.
	if opts.ParseJSON {
		ct := resp.Header.Get("Content-Type")
		if strings.HasPrefix(ct, "application/json") {
			var parsed any
			if jsonErr := json.Unmarshal(bodyBytes, &parsed); jsonErr == nil {
				result["json"] = parsed
			}
		}
	}

	return result, nil
}

// recordCBFailure increments the circuit breaker failure count for a key.
func (e *RequestEngine) recordCBFailure(key string) {
	e.cbMu.Lock()
	defer e.cbMu.Unlock()

	cb, exists := e.circuitBreakers[key]
	if !exists {
		cb = &domainCB{}
		e.circuitBreakers[key] = cb
	}

	cb.consecutiveFailures++
	cb.lastFailure = time.Now()

	if cb.consecutiveFailures >= e.cfg.CBMaxFailures {
		cb.disabled = true
	}
}

// recordCBSuccess resets the circuit breaker for a key.
func (e *RequestEngine) recordCBSuccess(key string) {
	e.cbMu.Lock()
	defer e.cbMu.Unlock()

	cb, exists := e.circuitBreakers[key]
	if !exists {
		return
	}

	cb.consecutiveFailures = 0
	cb.disabled = false
}

// categorizeHTTPError maps transport errors to user-friendly messages.
func categorizeHTTPError(err error, hostname string, timeoutSec int) string {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "context deadline exceeded"):
		return fmt.Sprintf("request timed out after %ds", timeoutSec)
	case strings.Contains(errStr, "connection refused"):
		return fmt.Sprintf("connection refused: %s", hostname)
	case strings.Contains(errStr, "no such host"):
		return fmt.Sprintf("dns lookup failed: %s", hostname)
	case strings.Contains(errStr, "tls"):
		return fmt.Sprintf("tls handshake failed: %s", hostname)
	case strings.Contains(errStr, "context canceled"):
		return "request canceled"
	case strings.Contains(errStr, "private/reserved IP blocked"):
		return "request to private/reserved IP address blocked"
	default:
		return fmt.Sprintf("request failed: %s", errStr)
	}
}

// cleanupRateLimiters removes idle rate limiter entries every 5 minutes.
// Entries not seen for 10+ minutes are removed.
func (e *RequestEngine) cleanupRateLimiters() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-e.cleanupDone:
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-10 * time.Minute)
			e.rateMu.Lock()
			for key, entry := range e.rateLimiters {
				if entry.lastSeen.Before(cutoff) {
					delete(e.rateLimiters, key)
				}
			}
			e.rateMu.Unlock()
		}
	}
}

// ---------- DB operations ----------

// CreatePluginRequestsTable creates the plugin_requests table using
// dialect-specific DDL. Idempotent (IF NOT EXISTS).
func (e *RequestEngine) CreatePluginRequestsTable(ctx context.Context) error {
	var ddl string

	switch e.dialect {
	case db.DialectSQLite:
		ddl = `CREATE TABLE IF NOT EXISTS plugin_requests (
    plugin_name    TEXT    NOT NULL,
    domain         TEXT    NOT NULL,
    description    TEXT    NOT NULL DEFAULT '',
    approved       INTEGER NOT NULL DEFAULT 0,
    approved_at    TEXT,
    approved_by    TEXT,
    plugin_version TEXT    NOT NULL DEFAULT '',
    created_at     TEXT    NOT NULL DEFAULT (datetime('now')),
    PRIMARY KEY (plugin_name, domain)
)`
	case db.DialectMySQL:
		ddl = `CREATE TABLE IF NOT EXISTS plugin_requests (
    plugin_name    VARCHAR(255) NOT NULL,
    domain         VARCHAR(253) NOT NULL,
    description    TEXT         NOT NULL,
    approved       TINYINT(1)   NOT NULL DEFAULT 0,
    approved_at    TIMESTAMP    NULL DEFAULT NULL,
    approved_by    VARCHAR(255),
    plugin_version VARCHAR(255) NOT NULL DEFAULT '',
    created_at     DATETIME     DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (plugin_name(191), domain(253))
)`
	case db.DialectPostgres:
		ddl = `CREATE TABLE IF NOT EXISTS plugin_requests (
    plugin_name    TEXT    NOT NULL,
    domain         TEXT    NOT NULL,
    description    TEXT    NOT NULL DEFAULT '',
    approved       BOOLEAN NOT NULL DEFAULT FALSE,
    approved_at    TIMESTAMPTZ,
    approved_by    TEXT,
    plugin_version TEXT    NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (plugin_name, domain)
)`
	default:
		return fmt.Errorf("unsupported dialect for plugin_requests table: %d", e.dialect)
	}

	_, err := e.pool.ExecContext(ctx, ddl)
	if err != nil {
		return fmt.Errorf("creating plugin_requests table: %w", err)
	}

	return nil
}

// UpsertRequestRegistrations records domain registrations in plugin_requests
// and loads approval state into memory. ON CONFLICT updates description and
// plugin_version only, does NOT revoke approval.
func (e *RequestEngine) UpsertRequestRegistrations(ctx context.Context, pluginName, pluginVersion string, requests []PendingRequest) error {
	for _, r := range requests {
		if err := e.upsertRequest(ctx, pluginName, pluginVersion, r.Domain, r.Description); err != nil {
			return fmt.Errorf("upserting request %s: %w", r.Domain, err)
		}

		approved, readErr := e.readRequestApproval(ctx, pluginName, r.Domain)
		if readErr != nil {
			return fmt.Errorf("reading request approval %s: %w", r.Domain, readErr)
		}

		key := pluginName + ":" + r.Domain
		e.mu.Lock()
		e.approved[key] = approved
		e.mu.Unlock()
	}

	return nil
}

// upsertRequest inserts or updates a single request registration.
func (e *RequestEngine) upsertRequest(ctx context.Context, pluginName, pluginVersion, domain, description string) error {
	switch e.dialect {
	case db.DialectSQLite:
		_, err := e.pool.ExecContext(ctx, `
			INSERT INTO plugin_requests (plugin_name, domain, description, approved, plugin_version)
			VALUES (?, ?, ?, 0, ?)
			ON CONFLICT(plugin_name, domain) DO UPDATE SET
				plugin_version = excluded.plugin_version,
				description = excluded.description
		`, pluginName, domain, description, pluginVersion)
		return err

	case db.DialectMySQL:
		_, err := e.pool.ExecContext(ctx, `
			INSERT INTO plugin_requests (plugin_name, domain, description, approved, plugin_version)
			VALUES (?, ?, ?, 0, ?)
			ON DUPLICATE KEY UPDATE
				plugin_version = VALUES(plugin_version),
				description = VALUES(description)
		`, pluginName, domain, description, pluginVersion)
		return err

	case db.DialectPostgres:
		_, err := e.pool.ExecContext(ctx, `
			INSERT INTO plugin_requests (plugin_name, domain, description, approved, plugin_version)
			VALUES ($1, $2, $3, FALSE, $4)
			ON CONFLICT (plugin_name, domain) DO UPDATE SET
				plugin_version = EXCLUDED.plugin_version,
				description = EXCLUDED.description
		`, pluginName, domain, description, pluginVersion)
		return err

	default:
		return fmt.Errorf("unsupported dialect: %d", e.dialect)
	}
}

// readRequestApproval reads the approval status from plugin_requests.
func (e *RequestEngine) readRequestApproval(ctx context.Context, pluginName, domain string) (bool, error) {
	var approvedInt int

	var query string
	var args []any
	switch e.dialect {
	case db.DialectPostgres:
		query = "SELECT CASE WHEN approved THEN 1 ELSE 0 END FROM plugin_requests WHERE plugin_name = $1 AND domain = $2"
		args = []any{pluginName, domain}
	default:
		query = "SELECT approved FROM plugin_requests WHERE plugin_name = ? AND domain = ?"
		args = []any{pluginName, domain}
	}

	err := e.pool.QueryRowContext(ctx, query, args...).Scan(&approvedInt)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return approvedInt != 0, nil
}

// ApproveRequest approves a domain for a plugin, updating DB and in-memory state.
func (e *RequestEngine) ApproveRequest(ctx context.Context, pluginName, domain, approvedBy string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	var query string
	var args []any
	switch e.dialect {
	case db.DialectPostgres:
		query = "UPDATE plugin_requests SET approved = TRUE, approved_at = $1, approved_by = $2 WHERE plugin_name = $3 AND domain = $4"
		args = []any{now, approvedBy, pluginName, domain}
	default:
		query = "UPDATE plugin_requests SET approved = 1, approved_at = ?, approved_by = ? WHERE plugin_name = ? AND domain = ?"
		args = []any{now, approvedBy, pluginName, domain}
	}

	_, err := e.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("approving request in DB: %w", err)
	}

	key := pluginName + ":" + domain
	e.mu.Lock()
	e.approved[key] = true
	e.mu.Unlock()

	return nil
}

// RevokeRequest revokes approval for a domain.
func (e *RequestEngine) RevokeRequest(ctx context.Context, pluginName, domain string) error {
	var query string
	var args []any
	switch e.dialect {
	case db.DialectPostgres:
		query = "UPDATE plugin_requests SET approved = FALSE, approved_at = NULL, approved_by = NULL WHERE plugin_name = $1 AND domain = $2"
		args = []any{pluginName, domain}
	default:
		query = "UPDATE plugin_requests SET approved = 0, approved_at = NULL, approved_by = NULL WHERE plugin_name = ? AND domain = ?"
		args = []any{pluginName, domain}
	}

	_, err := e.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("revoking request in DB: %w", err)
	}

	key := pluginName + ":" + domain
	e.mu.Lock()
	e.approved[key] = false
	e.mu.Unlock()

	return nil
}

// ListRequests returns all registered domains with approval status from DB.
// Sorted by (plugin_name, domain) for deterministic output.
func (e *RequestEngine) ListRequests(ctx context.Context) ([]RequestRegistration, error) {
	var query string
	switch e.dialect {
	case db.DialectPostgres:
		query = "SELECT plugin_name, domain, description, CASE WHEN approved THEN 1 ELSE 0 END, plugin_version FROM plugin_requests ORDER BY plugin_name, domain"
	default:
		query = "SELECT plugin_name, domain, description, approved, plugin_version FROM plugin_requests ORDER BY plugin_name, domain"
	}

	rows, err := e.pool.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing requests: %w", err)
	}
	defer rows.Close()

	var result []RequestRegistration
	for rows.Next() {
		var r RequestRegistration
		var approvedInt int
		if scanErr := rows.Scan(&r.PluginName, &r.Domain, &r.Description, &approvedInt, &r.PluginVersion); scanErr != nil {
			return nil, fmt.Errorf("scanning request row: %w", scanErr)
		}
		r.Approved = approvedInt != 0
		result = append(result, r)
	}

	if rowErr := rows.Err(); rowErr != nil {
		return nil, fmt.Errorf("iterating request rows: %w", rowErr)
	}

	return result, nil
}

// CleanupOrphanedRequests deletes rows from plugin_requests where plugin_name
// is not in the discovered set.
func (e *RequestEngine) CleanupOrphanedRequests(ctx context.Context, discoveredPlugins []string) error {
	if len(discoveredPlugins) == 0 {
		_, err := e.pool.ExecContext(ctx, "DELETE FROM plugin_requests")
		if err != nil {
			return fmt.Errorf("cleaning all orphaned requests: %w", err)
		}
		return nil
	}

	placeholders := make([]string, len(discoveredPlugins))
	args := make([]any, len(discoveredPlugins))
	for i, name := range discoveredPlugins {
		switch e.dialect {
		case db.DialectPostgres:
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		default:
			placeholders[i] = "?"
		}
		args[i] = name
	}

	query := fmt.Sprintf(
		"DELETE FROM plugin_requests WHERE plugin_name NOT IN (%s)",
		strings.Join(placeholders, ", "),
	)

	_, err := e.pool.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("cleaning orphaned requests: %w", err)
	}

	return nil
}

// UnregisterPlugin clears all in-memory state for a plugin.
// Called during reload before re-registering.
func (e *RequestEngine) UnregisterPlugin(pluginName string) {
	prefix := pluginName + ":"

	e.mu.Lock()
	for key := range e.approved {
		if strings.HasPrefix(key, prefix) {
			delete(e.approved, key)
		}
	}
	e.mu.Unlock()

	e.rateMu.Lock()
	for key := range e.rateLimiters {
		if strings.HasPrefix(key, prefix) {
			delete(e.rateLimiters, key)
		}
	}
	e.rateMu.Unlock()

	e.cbMu.Lock()
	for key := range e.circuitBreakers {
		if strings.HasPrefix(key, prefix) {
			delete(e.circuitBreakers, key)
		}
	}
	e.cbMu.Unlock()
}

// Close stops the rate limiter cleanup goroutine and closes idle connections.
// Does NOT drain in-flight requests -- those complete or timeout via their
// per-request context. Called AFTER hookEngine.Close() to ensure after-hook
// goroutines (which may call request.send()) have completed.
func (e *RequestEngine) Close() {
	close(e.cleanupDone)
	e.client.CloseIdleConnections()
}

// LoadApprovals reads all approved domains from the DB and populates the
// in-memory approved map. Called during startup after CreatePluginRequestsTable.
func (e *RequestEngine) LoadApprovals(ctx context.Context) error {
	var query string
	switch e.dialect {
	case db.DialectPostgres:
		query = "SELECT plugin_name, domain FROM plugin_requests WHERE approved = TRUE"
	default:
		query = "SELECT plugin_name, domain FROM plugin_requests WHERE approved = 1"
	}

	rows, err := e.pool.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("loading request approvals: %w", err)
	}
	defer rows.Close()

	e.mu.Lock()
	defer e.mu.Unlock()

	for rows.Next() {
		var pluginName, domain string
		if scanErr := rows.Scan(&pluginName, &domain); scanErr != nil {
			return fmt.Errorf("scanning approval row: %w", scanErr)
		}
		e.approved[pluginName+":"+domain] = true
	}

	return rows.Err()
}

// ApprovedDomains returns a sorted list of approved "plugin:domain" keys.
// Used for diagnostics and testing.
func (e *RequestEngine) ApprovedDomains() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []string
	for key, approved := range e.approved {
		if approved {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result
}
