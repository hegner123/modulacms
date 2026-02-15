package utility

import (
	"fmt"
	"sync"
	"time"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// Labels are key-value pairs for dimensional metrics
type Labels map[string]string

// Metric represents a single metric data point
type Metric struct {
	Name      string
	Type      MetricType
	Value     float64
	Labels    Labels
	Timestamp time.Time
}

// Metrics holds all application metrics with thread-safe operations
type Metrics struct {
	mu         sync.RWMutex
	counters   map[string]float64
	gauges     map[string]float64
	histograms map[string][]float64
	labels     map[string]Labels
}

// GlobalMetrics is the default metrics instance
var GlobalMetrics = NewMetrics()

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		counters:   make(map[string]float64),
		gauges:     make(map[string]float64),
		histograms: make(map[string][]float64),
		labels:     make(map[string]Labels),
	}
}

// Counter increments a counter metric
func (m *Metrics) Counter(name string, value float64, labels Labels) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.keyWithLabels(name, labels)
	m.counters[key] += value
	if labels != nil {
		m.labels[key] = labels
	}
}

// Increment increments a counter by 1
func (m *Metrics) Increment(name string, labels Labels) {
	m.Counter(name, 1, labels)
}

// Gauge sets a gauge metric to a specific value
func (m *Metrics) Gauge(name string, value float64, labels Labels) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.keyWithLabels(name, labels)
	m.gauges[key] = value
	if labels != nil {
		m.labels[key] = labels
	}
}

// Histogram records a value in a histogram
func (m *Metrics) Histogram(name string, value float64, labels Labels) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.keyWithLabels(name, labels)
	m.histograms[key] = append(m.histograms[key], value)
	if labels != nil {
		m.labels[key] = labels
	}
}

// Timing records a duration as a histogram value in milliseconds
func (m *Metrics) Timing(name string, duration time.Duration, labels Labels) {
	m.Histogram(name, float64(duration.Milliseconds()), labels)
}

// GetSnapshot returns a snapshot of all metrics
func (m *Metrics) GetSnapshot() map[string]Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := make(map[string]Metric)
	now := time.Now()

	for key, value := range m.counters {
		snapshot[key] = Metric{
			Name:      key,
			Type:      MetricTypeCounter,
			Value:     value,
			Labels:    m.labels[key],
			Timestamp: now,
		}
	}

	for key, value := range m.gauges {
		snapshot[key] = Metric{
			Name:      key,
			Type:      MetricTypeGauge,
			Value:     value,
			Labels:    m.labels[key],
			Timestamp: now,
		}
	}

	for key, values := range m.histograms {
		// For histograms, return the average
		avg := 0.0
		if len(values) > 0 {
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			avg = sum / float64(len(values))
		}
		snapshot[key] = Metric{
			Name:      key,
			Type:      MetricTypeHistogram,
			Value:     avg,
			Labels:    m.labels[key],
			Timestamp: now,
		}
	}

	return snapshot
}

// Reset clears all metrics (useful for testing or periodic resets)
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counters = make(map[string]float64)
	m.gauges = make(map[string]float64)
	m.histograms = make(map[string][]float64)
	m.labels = make(map[string]Labels)
}

// keyWithLabels creates a unique key combining name and labels
func (m *Metrics) keyWithLabels(name string, labels Labels) string {
	if labels == nil || len(labels) == 0 {
		return name
	}
	// Simple key generation - in production, use stable sorting
	key := name
	for k, v := range labels {
		key += fmt.Sprintf(",%s=%s", k, v)
	}
	return key
}

// MeasureTime executes a function and records its duration
func MeasureTime(name string, labels Labels, fn func()) {
	start := time.Now()
	fn()
	GlobalMetrics.Timing(name, time.Since(start), labels)
}

// MeasureTimeCtx executes a function and records its duration, returns any error
func MeasureTimeCtx(name string, labels Labels, fn func() error) error {
	start := time.Now()
	err := fn()
	GlobalMetrics.Timing(name, time.Since(start), labels)
	return err
}

// Standard metric names for application monitoring.
const (
	MetricHTTPRequests      = "http.requests"
	MetricHTTPDuration      = "http.duration"
	MetricHTTPErrors        = "http.errors"
	MetricDBQueries         = "db.queries"
	MetricDBDuration        = "db.duration"
	MetricDBErrors          = "db.errors"
	MetricSSHConnections    = "ssh.connections"
	MetricSSHErrors         = "ssh.errors"
	MetricCacheHits         = "cache.hits"
	MetricCacheMisses       = "cache.misses"
	MetricActiveConnections = "connections.active"
	MetricMemoryUsage       = "memory.usage"
	MetricGoroutines        = "goroutines.count"
)
