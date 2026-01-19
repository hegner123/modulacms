# Metrics & Observability

Basic metrics package for tracking application performance and integrating with external observability platforms.

## Quick Start

### 1. Configuration

Add to your `config.json`:

```json
{
  "observability_enabled": true,
  "observability_provider": "console",
  "observability_flush_interval": "30s",
  "observability_tags": {
    "service": "modulacms",
    "environment": "production"
  }
}
```

### 2. Initialize in main.go

```go
import "github.com/hegner123/modulacms/internal/utility"

func main() {
    // Load config
    cfg, _ := config.Load()

    // Initialize observability
    obsClient, err := utility.NewObservabilityClient(utility.ObservabilityConfig{
        Enabled:       cfg.Observability_Enabled,
        Provider:      cfg.Observability_Provider,
        DSN:           cfg.Observability_DSN,
        Environment:   cfg.Observability_Environment,
        Release:       cfg.Observability_Release,
        SampleRate:    cfg.Observability_Sample_Rate,
        TracesRate:    cfg.Observability_Traces_Rate,
        FlushInterval: cfg.Observability_Flush_Interval,
        Tags:          cfg.Observability_Tags,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Set global instance
    utility.GlobalObservability = obsClient

    // Start periodic flushing
    ctx := context.Background()
    obsClient.Start(ctx)

    // Ensure cleanup on shutdown
    defer obsClient.Stop()

    // Your application code...
}
```

### 3. Track Metrics

```go
// Counter - count events
utility.GlobalMetrics.Increment("user.login", utility.Labels{
    "method": "ssh",
})

// Gauge - track current values
utility.GlobalMetrics.Gauge("memory.usage", 512000000, utility.Labels{
    "type": "heap",
})

// Histogram/Timing - track durations
utility.GlobalMetrics.Timing("http.request", duration, utility.Labels{
    "method": "GET",
    "path": "/api/users",
})

// Easy timing with helper
utility.MeasureTime("database.query", utility.Labels{"table": "users"}, func() {
    // Your code here
})
```

### 4. Capture Errors

```go
if err != nil {
    utility.CaptureError(err, map[string]any{
        "user_id": userID,
        "action": "create_post",
    })
}
```

## Sentry Integration

### Configuration

```json
{
  "observability_enabled": true,
  "observability_provider": "sentry",
  "observability_dsn": "https://YOUR_KEY@o0.ingest.sentry.io/PROJECT_ID",
  "observability_environment": "production",
  "observability_release": "modulacms@1.0.0",
  "observability_sample_rate": 1.0,
  "observability_traces_rate": 0.1,
  "observability_send_pii": false,
  "observability_debug": false,
  "observability_server_name": "modulacms-prod-01",
  "observability_flush_interval": "30s",
  "observability_tags": {
    "service": "modulacms",
    "region": "us-east-1"
  }
}
```

### What Each Option Does

- **observability_enabled** - Master switch for all observability features
- **observability_provider** - Which platform to use (`sentry`, `datadog`, `console`)
- **observability_dsn** - Sentry Data Source Name (get from Sentry project settings)
- **observability_environment** - Environment name (production, staging, development)
- **observability_release** - Version/release identifier for tracking deployments
- **observability_sample_rate** - Percentage of errors to send (0.0 to 1.0)
  - `1.0` = send all errors
  - `0.1` = send 10% of errors (useful for high-traffic apps)
- **observability_traces_rate** - Percentage of performance traces to send
  - `0.1` = track 10% of requests (recommended for production)
  - `1.0` = track all requests (can be expensive)
- **observability_send_pii** - Send personally identifiable information
  - `false` = scrub email addresses, IP addresses, etc. (recommended)
  - `true` = include all data (useful for debugging)
- **observability_debug** - Enable debug logging from Sentry SDK
- **observability_server_name** - Name of this server instance
- **observability_flush_interval** - How often to batch-send metrics (`30s`, `1m`, etc.)
- **observability_tags** - Global tags added to all events (service, region, datacenter, etc.)

### Getting Your Sentry DSN

1. Sign up at https://sentry.io
2. Create a new project
3. Go to Settings → Client Keys (DSN)
4. Copy the DSN string: `https://examplePublicKey@o0.ingest.sentry.io/0`

### Installing Sentry SDK (Production)

To enable full Sentry integration:

```bash
go get github.com/getsentry/sentry-go
```

Then uncomment the Sentry initialization code in `observability.go`.

## Common Metrics

Pre-defined metric names:

```go
utility.MetricHTTPRequests      // HTTP request count
utility.MetricHTTPDuration      // HTTP request duration
utility.MetricHTTPErrors        // HTTP error count
utility.MetricDBQueries         // Database query count
utility.MetricDBDuration        // Database query duration
utility.MetricDBErrors          // Database error count
utility.MetricSSHConnections    // SSH connection count
utility.MetricSSHErrors         // SSH error count
utility.MetricCacheHits         // Cache hit count
utility.MetricCacheMisses       // Cache miss count
utility.MetricActiveConnections // Current active connections
utility.MetricMemoryUsage       // Memory usage in bytes
utility.MetricGoroutines        // Current goroutine count
```

## Middleware Integration

Add metrics to your HTTP middleware:

```go
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        utility.GlobalMetrics.Increment(utility.MetricHTTPRequests, utility.Labels{
            "method": r.Method,
            "path": r.URL.Path,
        })

        next.ServeHTTP(w, r)

        utility.GlobalMetrics.Timing(utility.MetricHTTPDuration, time.Since(start), utility.Labels{
            "method": r.Method,
            "path": r.URL.Path,
        })
    })
}
```

## Environment-Specific Settings

Development:
```json
{
  "observability_enabled": true,
  "observability_provider": "console",
  "observability_debug": true
}
```

Staging:
```json
{
  "observability_enabled": true,
  "observability_provider": "sentry",
  "observability_environment": "staging",
  "observability_sample_rate": 1.0,
  "observability_traces_rate": 0.5
}
```

Production:
```json
{
  "observability_enabled": true,
  "observability_provider": "sentry",
  "observability_environment": "production",
  "observability_sample_rate": 1.0,
  "observability_traces_rate": 0.1,
  "observability_send_pii": false
}
```

## Alternative Providers

### Datadog
```json
{
  "observability_provider": "datadog",
  "observability_dsn": "https://api.datadoghq.com",
  "observability_tags": {
    "dd.api_key": "YOUR_API_KEY"
  }
}
```

### New Relic
```json
{
  "observability_provider": "newrelic",
  "observability_dsn": "YOUR_LICENSE_KEY"
}
```

### Prometheus (self-hosted)
```json
{
  "observability_provider": "prometheus",
  "observability_dsn": "http://localhost:9090"
}
```

## Best Practices

1. **Use Labels Wisely** - Don't create unlimited label combinations (avoid user IDs, request IDs)
2. **Sample in Production** - Use `traces_rate: 0.1` to reduce costs
3. **Tag Everything** - Add environment, service, version tags
4. **Monitor Critical Paths** - Focus on HTTP, database, SSH performance
5. **Set Alerts** - Configure Sentry alerts for error rate spikes
6. **Review Weekly** - Check Sentry dashboard for trends

## See Also

- `metrics_example.go` - Complete integration examples
- `observability_example.json` - Full configuration example
