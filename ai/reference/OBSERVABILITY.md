# Observability Concurrency Guide

## Yes, It Already Runs Concurrently

The observability client is **designed to run in a background goroutine** from the start. Here's how it works:

### How It Works

```go
// 1. Create the client
obsClient, err := utility.NewObservabilityClient(config)

// 2. Start background goroutine
obsClient.Start(ctx)  // <-- Launches goroutine that runs until ctx is cancelled

// 3. Your application continues running...
// The client is now:
//   - Collecting metrics in a thread-safe manner (mutex protected)
//   - Flushing metrics every 30 seconds (or configured interval)
//   - Running independently without blocking your main thread

// 4. On shutdown
obsClient.Stop()  // <-- Stops goroutine, flushes final metrics, closes connection
```

### Internal Goroutine Structure

```go
func (c *ObservabilityClient) Start(ctx context.Context) {
    c.wg.Add(1)
    go func() {                           // <-- GOROUTINE LAUNCHED HERE
        defer c.wg.Done()
        ticker := time.NewTicker(c.flushInterval)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:              // Every 30s (default)
                c.flush()                 // Send metrics to Sentry/Datadog/etc
            case <-c.stopChan:            // Manual stop
                return
            case <-ctx.Done():            // Context cancelled
                return
            }
        }
    }()
}
```

### Thread Safety

The `Metrics` struct uses mutex locks for thread safety:

```go
type Metrics struct {
    mu         sync.RWMutex  // <-- Protects all data
    counters   map[string]float64
    gauges     map[string]float64
    histograms map[string][]float64
}

func (m *Metrics) Increment(name string, labels Labels) {
    m.mu.Lock()              // <-- Safe for concurrent access
    defer m.mu.Unlock()
    // ... update counter
}
```

**This means:**
- ✅ HTTP handlers can call `GlobalMetrics.Increment()` concurrently
- ✅ SSH middleware can call `GlobalMetrics.Timing()` concurrently
- ✅ Database layer can call `GlobalMetrics.Gauge()` concurrently
- ✅ Background goroutine flushes metrics concurrently
- ✅ No race conditions, no blocking

## Integration Pattern in main.go

### Recommended Approach

```go
func run() (ReturnCode, error) {
    // 1. Create root context that controls everything
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 2. Load config
    config, _ := loadConfig()

    // 3. Initialize observability
    obsClient, err := utility.NewObservabilityClient(...)
    if err != nil {
        log.Error("Observability init failed", err)
        // Continue without observability rather than crashing
    } else {
        utility.GlobalObservability = obsClient

        // 4. Start background goroutine
        obsClient.Start(ctx)  // Runs until ctx is cancelled

        // 5. Register shutdown
        defer func() {
            log.Info("Stopping observability...")
            if err := obsClient.Stop(); err != nil {
                log.Error("Observability shutdown error", err)
            }
        }()
    }

    // 6. Start your servers (they run in their own goroutines)
    go startSSHServer()
    go startHTTPServer()

    // 7. Wait for shutdown signal
    <-signalChan

    // 8. Cancel context - this stops observability goroutine
    cancel()

    // 9. Deferred obsClient.Stop() is called automatically
    //    - Stops the background goroutine
    //    - Flushes remaining metrics
    //    - Closes connection to Sentry

    return OKSIG, nil
}
```

### What Runs Concurrently

When your app is running, you have multiple goroutines:

```
main goroutine
├── Observability goroutine (flushing metrics every 30s)
├── SSH server goroutine (handling connections)
├── HTTP server goroutine (handling requests)
└── HTTPS server goroutine (handling requests)
```

Each HTTP/SSH request spawns more goroutines:
```
HTTP server goroutine
├── Request 1 goroutine → calls GlobalMetrics.Increment() (thread-safe)
├── Request 2 goroutine → calls GlobalMetrics.Timing() (thread-safe)
└── Request 3 goroutine → calls GlobalMetrics.Gauge() (thread-safe)
```

All can safely call metrics functions because of mutex protection.

## Shutdown Sequence

```
1. User presses Ctrl+C
2. Signal received → cancel context
3. Context cancellation propagates:
   ├── Observability goroutine exits
   ├── Stops accepting new requests
4. obsClient.Stop() called (deferred):
   ├── Waits for goroutine to finish (wg.Wait())
   ├── Flushes remaining metrics to Sentry
   ├── Closes provider connection
5. Shutdown HTTP/HTTPS/SSH servers
6. Exit cleanly
```

## Performance Considerations

### Why Concurrent Flushing?

**Without concurrency:**
```go
// BAD: Blocks request handling
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    metric := Metric{Name: "request", Value: 1}
    sendToSentry(metric)  // <-- BLOCKS for network I/O (10-100ms)
    // User waits for Sentry response before getting their data
}
```

**With concurrency:**
```go
// GOOD: Non-blocking
func HandleRequest(w http.ResponseWriter, r *http.Request) {
    GlobalMetrics.Increment("request", nil)  // <-- Just writes to map (< 1µs)
    // Background goroutine will flush later
    // User gets response immediately
}
```

### Batching Benefits

Instead of sending 1000 individual metrics:
```
Request 1 → Sentry API call (50ms)
Request 2 → Sentry API call (50ms)
Request 3 → Sentry API call (50ms)
...
Request 1000 → Sentry API call (50ms)
Total: 50 seconds of network time
```

We batch every 30 seconds:
```
Requests 1-1000 → collect in memory (fast)
After 30s → send 1000 metrics in batch (100ms)
Total: 100ms of network time
```

**Result:** 500x more efficient

## Advanced Patterns

### Multiple Background Workers

```go
// WorkerManager pattern for managing multiple background services
type WorkerManager struct {
    workers []BackgroundWorker
}

func main() {
    manager := NewWorkerManager()

    // Register all background workers
    manager.Register(observabilityClient)
    manager.Register(metricsCollector)
    manager.Register(healthChecker)
    manager.Register(cacheWarmer)

    // Start all concurrently
    manager.StartAll(ctx)

    // ... run application ...

    // Stop all with timeout
    manager.StopAll(30 * time.Second)
}
```

### System Metrics Collector

Run a separate goroutine to collect system metrics:

```go
collector := NewSystemMetricsCollector(10 * time.Second)
collector.Start(ctx)  // Goroutine that samples memory/goroutines every 10s

// Automatically sends to observability:
// - memory.usage
// - goroutines.count
// - cpu.percent
```

### Health Check Endpoint

Expose metrics via HTTP endpoint (doesn't need goroutine):

```go
http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    snapshot := GlobalMetrics.GetSnapshot()
    json.NewEncoder(w).Encode(snapshot)
})
```

## Configuration for Production

### Development (Console Provider)
```json
{
  "observability_enabled": true,
  "observability_provider": "console",
  "observability_flush_interval": "5s"
}
```
- Logs to console
- Fast flushing (5s) for debugging
- No external service needed

### Production (Sentry Provider)
```json
{
  "observability_enabled": true,
  "observability_provider": "sentry",
  "observability_dsn": "https://KEY@sentry.io/PROJECT",
  "observability_flush_interval": "30s",
  "observability_sample_rate": 1.0,
  "observability_traces_rate": 0.1
}
```
- Sends to Sentry
- Slower flushing (30s) to batch efficiently
- Sample traces at 10% to reduce cost

## Common Patterns

### Pattern 1: Measure HTTP Duration
```go
func HTTPMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)

        utility.GlobalMetrics.Timing(utility.MetricHTTPDuration,
            time.Since(start),
            utility.Labels{"method": r.Method, "path": r.URL.Path},
        )
    })
}
```

### Pattern 2: Track Errors
```go
if err := db.Query(); err != nil {
    utility.GlobalMetrics.Increment(utility.MetricDBErrors,
        utility.Labels{"operation": "query"},
    )
    utility.CaptureError(err, map[string]any{
        "query": "SELECT * FROM users",
    })
}
```

### Pattern 3: Monitor Active Connections
```go
func HandleConnection(conn net.Conn) {
    utility.GlobalMetrics.Counter(utility.MetricActiveConnections, 1, nil)
    defer utility.GlobalMetrics.Counter(utility.MetricActiveConnections, -1, nil)

    // Handle connection...
}
```

## Troubleshooting

### Goroutine Leak Detection

```go
import "runtime"

// Before starting servers
before := runtime.NumGoroutine()

// After shutdown
after := runtime.NumGoroutine()

if after > before {
    log.Warn("Goroutine leak detected", "before", before, "after", after)
}
```

### Memory Leak Detection

```go
import "runtime"

var m runtime.MemStats
runtime.ReadMemStats(&m)
utility.GlobalMetrics.Gauge(utility.MetricMemoryUsage, float64(m.Alloc), nil)
```

### Deadlock Detection

Go has built-in deadlock detection. If all goroutines are blocked:
```
fatal error: all goroutines are asleep - deadlock!
```

To prevent:
- Always use context with timeout
- Never hold locks while doing I/O
- Use buffered channels when appropriate

## Summary

| Aspect | Details |
|--------|---------|
| **Concurrency** | ✅ Runs in background goroutine automatically |
| **Thread Safety** | ✅ Mutex-protected, safe for concurrent calls |
| **Blocking** | ❌ Non-blocking - metrics collected in memory |
| **Flushing** | Every 30s (configurable) in background |
| **Shutdown** | Graceful with context cancellation |
| **Performance** | < 1µs to record metric, batched network calls |
| **Integration** | Call `Start(ctx)` once, use `GlobalMetrics` anywhere |

**Bottom line:** Your HTTP handlers, SSH middleware, and database calls can all safely call metrics functions concurrently. The background goroutine handles flushing to Sentry without blocking your application.
