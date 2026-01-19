package utility

import (
	"context"
	"time"
)

// Commented imports for example code:
// import (
// 	"os"
// 	"os/signal"
// 	"syscall"
// )

// Example showing how to integrate observability into main.go
// This demonstrates concurrent execution with proper shutdown

/*
func main() {
	code, err := run()
	if err != nil || code == ERRSIG {
		DefaultLogger.Fatal("Root Return: ", err)
	}
}

func run() (ReturnCode, error) {
	// Create root context that controls all background workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal handling for graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		return ERRSIG, err
	}

	// Initialize observability client
	obsClient, err := NewObservabilityClient(ObservabilityConfig{
		Enabled:       config.Observability_Enabled,
		Provider:      config.Observability_Provider,
		DSN:           config.Observability_DSN,
		Environment:   config.Observability_Environment,
		Release:       config.Observability_Release,
		SampleRate:    config.Observability_Sample_Rate,
		TracesRate:    config.Observability_Traces_Rate,
		SendPII:       config.Observability_Send_PII,
		Debug:         config.Observability_Debug,
		ServerName:    config.Observability_Server_Name,
		FlushInterval: config.Observability_Flush_Interval,
		Tags:          config.Observability_Tags,
	})
	if err != nil {
		DefaultLogger.Error("Failed to initialize observability", err)
		// Continue without observability rather than failing
	} else {
		// Set global instance
		GlobalObservability = obsClient

		// Start background metric flushing (runs in goroutine)
		obsClient.Start(ctx)

		DefaultLogger.Info("Observability started",
			"provider", config.Observability_Provider,
			"environment", config.Observability_Environment,
		)

		// Register shutdown handler
		defer func() {
			DefaultLogger.Info("Stopping observability...")
			if err := obsClient.Stop(); err != nil {
				DefaultLogger.Error("Observability shutdown error", err)
			} else {
				DefaultLogger.Info("Observability stopped")
			}
		}()
	}

	// Start your servers (SSH, HTTP, HTTPS)
	// They run in their own goroutines
	startServers(ctx, config)

	// Wait for shutdown signal
	<-done
	DefaultLogger.Info("Shutdown signal received")

	// Cancel context - this stops all background workers including observability
	cancel()

	// Give background workers time to clean up
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown servers gracefully
	shutdownServers(shutdownCtx)

	return OKSIG, nil
}
*/

// BackgroundWorker represents a service that runs in the background
type BackgroundWorker interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Name() string
}

// WorkerManager manages multiple background workers
type WorkerManager struct {
	workers []BackgroundWorker
}

// NewWorkerManager creates a new worker manager
func NewWorkerManager() *WorkerManager {
	return &WorkerManager{
		workers: make([]BackgroundWorker, 0),
	}
}

// Register adds a worker to be managed
func (m *WorkerManager) Register(worker BackgroundWorker) {
	m.workers = append(m.workers, worker)
}

// StartAll starts all registered workers concurrently
func (m *WorkerManager) StartAll(ctx context.Context) {
	for _, worker := range m.workers {
		w := worker // Capture for goroutine
		go func() {
			DefaultLogger.Info("Starting worker", "name", w.Name())
			if err := w.Start(ctx); err != nil {
				DefaultLogger.Error("Worker error", err, "name", w.Name())
			}
		}()
	}
}

// StopAll stops all workers with timeout
func (m *WorkerManager) StopAll(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errChan := make(chan error, len(m.workers))

	for _, worker := range m.workers {
		w := worker // Capture for goroutine
		go func() {
			DefaultLogger.Info("Stopping worker", "name", w.Name())
			if err := w.Stop(ctx); err != nil {
				errChan <- err
			}
		}()
	}

	// Wait for all to complete or timeout
	for i := 0; i < len(m.workers); i++ {
		select {
		case err := <-errChan:
			if err != nil {
				DefaultLogger.Error("Worker stop error", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// Example: Integrating observability as a background worker
/*
func runWithWorkerManager() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create worker manager
	manager := NewWorkerManager()

	// Register observability as a worker
	obsClient, _ := NewObservabilityClient(config)
	manager.Register(obsClient)

	// Register other background services
	// manager.Register(metricsCollector)
	// manager.Register(healthChecker)
	// manager.Register(cacheWarmer)

	// Start all workers concurrently
	manager.StartAll(ctx)

	// Wait for shutdown signal
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done

	// Stop all workers with 30 second timeout
	return manager.StopAll(30 * time.Second)
}
*/

// Example: Observability client as a BackgroundWorker
// This shows how ObservabilityClient could implement the interface
/*
func (c *ObservabilityClient) Name() string {
	return "observability"
}

func (c *ObservabilityClient) Start(ctx context.Context) error {
	if !c.enabled {
		return nil
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

	return nil
}

func (c *ObservabilityClient) Stop(ctx context.Context) error {
	if !c.enabled {
		return nil
	}

	close(c.stopChan)
	c.wg.Wait()

	// Final flush with timeout from context
	if err := c.provider.Flush(5 * time.Second); err != nil {
		return err
	}

	return c.provider.Close()
}
*/

// Example: System metrics collector that runs in background
type SystemMetricsCollector struct {
	interval  time.Duration
	stopChan  chan struct{}
	enabled   bool
}

// NewSystemMetricsCollector creates a collector that periodically captures system metrics
func NewSystemMetricsCollector(interval time.Duration) *SystemMetricsCollector {
	return &SystemMetricsCollector{
		interval: interval,
		stopChan: make(chan struct{}),
		enabled:  true,
	}
}

func (s *SystemMetricsCollector) Name() string {
	return "system-metrics-collector"
}

func (s *SystemMetricsCollector) Start(ctx context.Context) error {
	if !s.enabled {
		return nil
	}

	go func() {
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.collectMetrics()
			case <-s.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (s *SystemMetricsCollector) Stop(ctx context.Context) error {
	close(s.stopChan)
	return nil
}

func (s *SystemMetricsCollector) collectMetrics() {
	// In production, use runtime.ReadMemStats() and runtime.NumGoroutine()
	// For now, just示例
	GlobalMetrics.Gauge(MetricMemoryUsage, 512000000, Labels{
		"type": "heap",
	})
	GlobalMetrics.Gauge(MetricGoroutines, 100, nil)
}
