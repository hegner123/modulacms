package utility

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ObservabilityProvider defines the interface for different observability backends
type ObservabilityProvider interface {
	// SendMetric sends a single metric to the provider
	SendMetric(metric Metric) error

	// SendError sends an error event to the provider
	SendError(err error, context map[string]any) error

	// Flush ensures all buffered data is sent
	Flush(timeout time.Duration) error

	// Close shuts down the provider
	Close() error
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
		return &ObservabilityClient{
			enabled: false,
		}, nil
	}

	// Parse flush interval
	flushInterval, err := time.ParseDuration(config.FlushInterval)
	if err != nil {
		flushInterval = 30 * time.Second // Default to 30 seconds
	}

	// Create provider based on config
	var provider ObservabilityProvider
	switch config.Provider {
	case "sentry":
		provider, err = NewSentryProvider(config)
	case "console":
		provider = NewConsoleProvider()
	default:
		return nil, fmt.Errorf("unsupported observability provider: %s", config.Provider)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

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
				return
			case <-ctx.Done():
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

	close(c.stopChan)
	c.wg.Wait()

	// Final flush
	if err := c.provider.Flush(5 * time.Second); err != nil {
		return err
	}

	return c.provider.Close()
}

// flush sends all current metrics to the provider
func (c *ObservabilityClient) flush() {
	snapshot := c.metrics.GetSnapshot()
	for _, metric := range snapshot {
		if err := c.provider.SendMetric(metric); err != nil {
			DefaultLogger.Error("Failed to send metric", err, "metric", metric.Name)
		}
	}
}

// SendError sends an error to the observability provider
func (c *ObservabilityClient) SendError(err error, context map[string]any) error {
	if !c.enabled {
		return nil
	}
	return c.provider.SendError(err, context)
}

// ConsoleProvider is a simple provider that logs metrics to console (useful for development)
type ConsoleProvider struct{}

func NewConsoleProvider() *ConsoleProvider {
	return &ConsoleProvider{}
}

func (p *ConsoleProvider) SendMetric(metric Metric) error {
	DefaultLogger.Info("METRIC",
		"name", metric.Name,
		"type", metric.Type,
		"value", metric.Value,
		"labels", metric.Labels,
	)
	return nil
}

func (p *ConsoleProvider) SendError(err error, context map[string]any) error {
	DefaultLogger.Error("OBSERVABILITY ERROR", err, "context", context)
	return nil
}

func (p *ConsoleProvider) Flush(timeout time.Duration) error {
	return nil
}

func (p *ConsoleProvider) Close() error {
	return nil
}

// SentryProvider integrates with Sentry for error tracking and performance monitoring
// This is a placeholder - actual implementation would use github.com/getsentry/sentry-go
type SentryProvider struct {
	config ObservabilityConfig
	// In production: add sentry.Client here
}

func NewSentryProvider(config ObservabilityConfig) (*SentryProvider, error) {
	// In production, initialize Sentry SDK here:
	// err := sentry.Init(sentry.ClientOptions{
	//     Dsn:              config.DSN,
	//     Environment:      config.Environment,
	//     Release:          config.Release,
	//     SampleRate:       config.SampleRate,
	//     TracesSampleRate: config.TracesRate,
	//     Debug:            config.Debug,
	//     ServerName:       config.ServerName,
	// })

	return &SentryProvider{
		config: config,
	}, nil
}

func (p *SentryProvider) SendMetric(metric Metric) error {
	// In production, send to Sentry:
	// sentry.CaptureMessage(fmt.Sprintf("%s: %f", metric.Name, metric.Value))
	//
	// Or for custom metrics:
	// transaction := sentry.StartTransaction(...)
	// transaction.SetData(metric.Name, metric.Value)
	// transaction.Finish()

	DefaultLogger.Debug("Would send to Sentry",
		"metric", metric.Name,
		"value", metric.Value,
	)
	return nil
}

func (p *SentryProvider) SendError(err error, context map[string]any) error {
	// In production:
	// sentry.WithScope(func(scope *sentry.Scope) {
	//     for k, v := range context {
	//         scope.SetContext(k, v)
	//     }
	//     sentry.CaptureException(err)
	// })

	DefaultLogger.Error("Would send error to Sentry", err, "context", context)
	return nil
}

func (p *SentryProvider) Flush(timeout time.Duration) error {
	// In production:
	// return sentry.Flush(timeout)
	return nil
}

func (p *SentryProvider) Close() error {
	// In production: cleanup Sentry client
	return nil
}

// Helper function to capture errors with context
func CaptureError(err error, context map[string]any) {
	// This would be called throughout the app to send errors to observability
	if GlobalObservability != nil {
		GlobalObservability.SendError(err, context)
	}
	DefaultLogger.Error("Error captured", err, "context", context)
}

// GlobalObservability is the global observability client instance
var GlobalObservability *ObservabilityClient
