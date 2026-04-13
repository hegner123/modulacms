package utility

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
)

// ObservabilityProvider defines the interface for different observability backends.
// Implementations handle error tracking, metrics, HTTP transaction tracing, and
// request-scoped error capture. The provider is selected via the config file's
// observability_provider field.
type ObservabilityProvider interface {
	// SendMetric sends a single metric to the provider.
	SendMetric(metric Metric) error

	// SendError sends an error event to the provider.
	SendError(err error, context map[string]any) error

	// Flush ensures all buffered data is sent.
	Flush(timeout time.Duration) error

	// Close shuts down the provider.
	Close() error

	// HTTPMiddleware returns provider-specific HTTP middleware for transaction
	// tracing and per-request context setup. Sentry creates transactions with
	// hub-per-request. Console is a pass-through.
	HTTPMiddleware() func(http.Handler) http.Handler

	// CaptureRequestError captures an error with request-scoped context. Unlike
	// SendError, this has access to the HTTP request so providers can extract
	// per-request state (e.g. Sentry hub from context) for enriched capture.
	CaptureRequestError(err error, r *http.Request, context map[string]any)
}

// ObservabilityClient manages metrics export to external providers
type ObservabilityClient struct {
	provider      ObservabilityProvider
	metrics       *Metrics
	flushInterval time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
	enabled       bool
}

// ObservabilityConfig holds configuration for observability
type ObservabilityConfig struct {
	Enabled       bool
	Provider      string
	DSN           string
	Environment   string
	Release       string
	SampleRate    float64
	TracesRate    float64
	SendPII       bool
	Debug         bool
	ServerName    string
	FlushInterval string
	Tags          map[string]string
}

// NewObservabilityClient creates a new observability client
func NewObservabilityClient(config ObservabilityConfig) (*ObservabilityClient, error) {
	if !config.Enabled {
		DefaultLogger.Debug("observability disabled")
		return &ObservabilityClient{
			enabled: false,
		}, nil
	}

	// Parse flush interval
	flushInterval, parseErr := time.ParseDuration(config.FlushInterval)
	if parseErr != nil {
		flushInterval = 30 * time.Second
		DefaultLogger.Warn("invalid observability flush interval, defaulting to 30s", parseErr, "raw", config.FlushInterval)
	}

	// Create provider based on config
	DefaultLogger.Info("initializing observability provider", "provider", config.Provider)
	var provider ObservabilityProvider
	var providerErr error
	switch config.Provider {
	case "sentry":
		provider, providerErr = NewSentryProvider(config)
	case "console":
		provider = NewConsoleProvider()
	default:
		return nil, fmt.Errorf("unsupported observability provider: %s", config.Provider)
	}

	if providerErr != nil {
		return nil, fmt.Errorf("failed to create provider: %w", providerErr)
	}

	DefaultLogger.Info("observability provider ready",
		"provider", config.Provider,
		"flush_interval", flushInterval.String(),
		"environment", config.Environment,
		"sample_rate", fmt.Sprintf("%.2f", config.SampleRate),
		"traces_rate", fmt.Sprintf("%.2f", config.TracesRate),
	)

	client := &ObservabilityClient{
		provider:      provider,
		metrics:       GlobalMetrics,
		flushInterval: flushInterval,
		stopChan:      make(chan struct{}),
		enabled:       true,
	}

	return client, nil
}

// Start begins periodic metric flushing
func (c *ObservabilityClient) Start(ctx context.Context) {
	if !c.enabled {
		return
	}

	DefaultLogger.Info("observability flush loop started", "interval", c.flushInterval.String())
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(c.flushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.flush()
			case <-c.stopChan:
				DefaultLogger.Debug("observability flush loop stopped")
				return
			case <-ctx.Done():
				DefaultLogger.Debug("observability flush loop stopped (context cancelled)")
				return
			}
		}
	}()
}

// Stop stops the observability client
func (c *ObservabilityClient) Stop() error {
	if !c.enabled {
		return nil
	}

	DefaultLogger.Info("observability shutting down")
	close(c.stopChan)
	c.wg.Wait()

	// Final flush
	DefaultLogger.Debug("observability final flush")
	if err := c.provider.Flush(5 * time.Second); err != nil {
		DefaultLogger.Error("observability final flush failed", err)
		return err
	}

	DefaultLogger.Info("observability provider closed")
	return c.provider.Close()
}

// flush sends all current metrics to the provider
func (c *ObservabilityClient) flush() {
	snapshot := c.metrics.GetSnapshot()
	DefaultLogger.Debug("observability flush", "metrics", fmt.Sprintf("%d", len(snapshot)))
	for _, metric := range snapshot {
		if err := c.provider.SendMetric(metric); err != nil {
			DefaultLogger.Error("failed to send metric", err, "metric", metric.Name)
		}
	}
}

// SendError sends an error to the observability provider
func (c *ObservabilityClient) SendError(err error, ctx map[string]any) error {
	if !c.enabled {
		return nil
	}
	DefaultLogger.Debug("observability capturing error", "error", err.Error())
	return c.provider.SendError(err, ctx)
}

// ConsoleProvider is a simple provider that logs metrics to console (useful for development)
type ConsoleProvider struct{}

// NewConsoleProvider creates a new console observability provider.
func NewConsoleProvider() *ConsoleProvider {
	return &ConsoleProvider{}
}

// SendMetric logs a metric to the console.
func (p *ConsoleProvider) SendMetric(metric Metric) error {
	DefaultLogger.Info("metric",
		"name", metric.Name,
		"type", metric.Type,
		"value", metric.Value,
		"labels", metric.Labels,
	)
	return nil
}

// SendError logs an error event to the console.
func (p *ConsoleProvider) SendError(err error, context map[string]any) error {
	DefaultLogger.Error("observability error", err, "context", context)
	return nil
}

// Flush is a no-op for the console provider.
func (p *ConsoleProvider) Flush(timeout time.Duration) error {
	return nil
}

// Close is a no-op for the console provider.
func (p *ConsoleProvider) Close() error {
	return nil
}

// HTTPMiddleware returns a pass-through middleware for the console provider.
func (p *ConsoleProvider) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next
	}
}

// CaptureRequestError logs the error with request details.
func (p *ConsoleProvider) CaptureRequestError(err error, r *http.Request, context map[string]any) {
	DefaultLogger.Error("request error captured",
		err,
		"method", r.Method,
		"path", r.URL.Path,
		"context", context,
	)
}

// logWriter is an io.Writer adapter that routes Sentry SDK debug output
// through DefaultLogger so all log output uses the same format and destination.
type logWriter struct{}

func (logWriter) Write(p []byte) (int, error) {
	msg := strings.TrimRight(string(p), "\n")
	if msg != "" {
		DefaultLogger.Debug(msg)
	}
	return len(p), nil
}

// SentryProvider integrates with Sentry for error tracking and performance monitoring.
type SentryProvider struct {
	config ObservabilityConfig
}

// NewSentryProvider initializes the Sentry SDK and returns a provider that
// sends errors and metrics to the configured Sentry project.
func NewSentryProvider(config ObservabilityConfig) (*SentryProvider, error) {
	opts := sentry.ClientOptions{
		Dsn:              config.DSN,
		Environment:      config.Environment,
		Release:          config.Release,
		SampleRate:       config.SampleRate,
		TracesSampleRate: config.TracesRate,
		EnableTracing:    config.TracesRate > 0,
		Debug:            config.Debug,
		DebugWriter:      &logWriter{},
		ServerName:       config.ServerName,
		AttachStacktrace: true,
	}

	if config.SendPII {
		opts.SendDefaultPII = true
	}

	if err := sentry.Init(opts); err != nil {
		return nil, fmt.Errorf("sentry.Init: %w", err)
	}
	dsnPreview := config.DSN
	if len(dsnPreview) > 20 {
		dsnPreview = dsnPreview[:20] + "..."
	}
	DefaultLogger.Info("sentry SDK initialized", "dsn", dsnPreview, "environment", config.Environment)

	// Apply global tags from config.
	if len(config.Tags) > 0 {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			for k, v := range config.Tags {
				scope.SetTag(k, v)
			}
		})
		DefaultLogger.Debug("sentry global tags applied", "count", fmt.Sprintf("%d", len(config.Tags)))
	}

	return &SentryProvider{config: config}, nil
}

// SendMetric records a metric as a Sentry breadcrumb so it is attached to
// subsequent error events for context. Sentry is primarily an error tracker;
// dedicated metrics export uses Prometheus or OTLP providers.
func (p *SentryProvider) SendMetric(metric Metric) error {
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: "metric",
		Message:  fmt.Sprintf("%s = %f", metric.Name, metric.Value),
		Level:    sentry.LevelInfo,
		Data: map[string]any{
			"name":   metric.Name,
			"type":   string(metric.Type),
			"value":  metric.Value,
			"labels": metric.Labels,
		},
	})
	return nil
}

// SendError captures an exception in Sentry with the provided context.
func (p *SentryProvider) SendError(err error, ctx map[string]any) error {
	DefaultLogger.Debug("sentry capturing exception", "error", err.Error())
	sentry.WithScope(func(scope *sentry.Scope) {
		for k, v := range ctx {
			scope.SetExtra(k, v)
		}
		sentry.CaptureException(err)
	})
	return nil
}

// Flush ensures all buffered events are sent to Sentry.
func (p *SentryProvider) Flush(timeout time.Duration) error {
	DefaultLogger.Debug("sentry flushing", "timeout", timeout.String())
	if !sentry.Flush(timeout) {
		return fmt.Errorf("sentry flush timed out after %s", timeout)
	}
	DefaultLogger.Debug("sentry flush complete")
	return nil
}

// Close flushes remaining events and releases Sentry resources.
func (p *SentryProvider) Close() error {
	DefaultLogger.Debug("sentry provider closing")
	sentry.Flush(2 * time.Second)
	return nil
}

// HTTPMiddleware returns Sentry-specific middleware that creates a per-request
// hub clone and starts an HTTP transaction for performance monitoring.
func (p *SentryProvider) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hub := sentry.GetHubFromContext(r.Context())
			if hub == nil {
				hub = sentry.CurrentHub().Clone()
			}
			ctx := sentry.SetHubOnContext(r.Context(), hub)

			transactionName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			options := []sentry.SpanOption{
				sentry.WithOpName("http.server"),
				sentry.WithTransactionSource(sentry.SourceURL),
				sentry.ContinueFromRequest(r),
			}
			transaction := sentry.StartTransaction(ctx, transactionName, options...)
			defer transaction.Finish()

			transaction.SetData("http.method", r.Method)
			transaction.SetData("http.url", r.URL.String())
			transaction.SetData("http.query", r.URL.RawQuery)
			transaction.SetTag("http.method", r.Method)
			transaction.SetTag("http.path", r.URL.Path)

			rw := &sentryResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r.WithContext(transaction.Context()))

			transaction.Status = sentry.HTTPtoSpanStatus(rw.statusCode)
			transaction.SetData("http.status_code", rw.statusCode)
			transaction.SetTag("http.status_code", fmt.Sprintf("%d", rw.statusCode))
		})
	}
}

// CaptureRequestError captures an exception using the per-request Sentry hub,
// preserving transaction context and scope set by HTTPMiddleware.
func (p *SentryProvider) CaptureRequestError(err error, r *http.Request, ctx map[string]any) {
	hub := sentry.GetHubFromContext(r.Context())
	if hub == nil {
		hub = sentry.CurrentHub()
	}
	hub.WithScope(func(scope *sentry.Scope) {
		for k, v := range ctx {
			scope.SetExtra(k, v)
		}
		hub.CaptureException(err)
	})
}

// sentryResponseWriter wraps http.ResponseWriter to capture the status code
// for Sentry transaction tagging.
type sentryResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *sentryResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// HTTPMiddleware returns the configured provider's HTTP middleware. Returns a
// pass-through when observability is disabled.
func (c *ObservabilityClient) HTTPMiddleware() func(http.Handler) http.Handler {
	if !c.enabled || c.provider == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return c.provider.HTTPMiddleware()
}

// CaptureRequestError delegates to the provider's request-aware error capture.
// No-op when observability is disabled.
func (c *ObservabilityClient) CaptureRequestError(err error, r *http.Request, ctx map[string]any) {
	if !c.enabled || c.provider == nil {
		return
	}
	c.provider.CaptureRequestError(err, r, ctx)
}

// CaptureError sends an error directly to the observability provider with
// enriched context. Use this at explicit error boundaries (panics, server
// crashes) where you want to attach extra metadata beyond what the logger
// provides. Normal handler errors are captured automatically via
// DefaultLogger.Error().
func CaptureError(err error, context map[string]any) {
	if GlobalObservability != nil {
		GlobalObservability.SendError(err, context)
	}
	// Log to console without re-triggering the observability send — use Warn
	// level so it still prints but does not recurse into SendError.
	DefaultLogger.Warn("error captured for observability", err, "context", context)
}

// NewObservabilityClientFromProvider creates an ObservabilityClient with an
// explicit provider, bypassing config-based selection. Useful for testing and
// for callers that construct their own provider.
func NewObservabilityClientFromProvider(provider ObservabilityProvider) *ObservabilityClient {
	return &ObservabilityClient{
		provider: provider,
		metrics:  GlobalMetrics,
		stopChan: make(chan struct{}),
		enabled:  true,
	}
}

// GlobalObservability is the global observability client instance
var GlobalObservability *ObservabilityClient
