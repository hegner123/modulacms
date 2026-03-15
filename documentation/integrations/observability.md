# Observability

Connect ModulaCMS to external monitoring services to track errors, performance, and system health.

## Supported Providers

ModulaCMS forwards metrics and error reports to **Sentry**, **Datadog**, or **New Relic**. Configure one provider per instance in `modula.config.json`.

> **Good to know**: ModulaCMS collects metrics internally regardless of whether external forwarding is enabled. You can query the metrics endpoint even with `observability_enabled: false`.

## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `observability_enabled` | bool | `false` | Enable external metric forwarding |
| `observability_provider` | string | `"console"` | Provider: `sentry`, `datadog`, `newrelic`, `console` |
| `observability_dsn` | string | `""` | Connection string / DSN for the provider |
| `observability_environment` | string | `"development"` | Environment label (e.g., `production`, `staging`) |
| `observability_release` | string | `""` | Version or release identifier |
| `observability_sample_rate` | float | `1.0` | Event sample rate (0.0 to 1.0) |
| `observability_traces_rate` | float | `0.1` | Trace sample rate (0.0 to 1.0) |
| `observability_send_pii` | bool | `false` | Include personally identifiable information |
| `observability_debug` | bool | `false` | Enable debug logging for the observability client |
| `observability_server_name` | string | `""` | Instance identifier for multi-node deployments |
| `observability_flush_interval` | string | `"30s"` | How often metrics are flushed to the provider |
| `observability_tags` | object | `{}` | Global key-value tags added to all events |

All observability fields are hot-reloadable. Changes take effect without restarting the server.

## Set Up Sentry

```json
{
  "observability_enabled": true,
  "observability_provider": "sentry",
  "observability_dsn": "https://examplePublicKey@o0.ingest.sentry.io/0",
  "observability_environment": "production",
  "observability_release": "v1.2.3",
  "observability_sample_rate": 1.0,
  "observability_traces_rate": 0.1,
  "observability_server_name": "cms-prod-01"
}
```

Find your DSN in Sentry under **Settings > Projects > (your project) > Client Keys (DSN)**.

## Set Up Datadog

```json
{
  "observability_enabled": true,
  "observability_provider": "datadog",
  "observability_dsn": "https://http-intake.logs.datadoghq.com",
  "observability_environment": "production",
  "observability_release": "v1.2.3",
  "observability_server_name": "cms-prod-01",
  "observability_tags": {
    "service": "modulacms",
    "region": "us-east-1"
  }
}
```

Use `observability_tags` to attach metadata that your Datadog dashboards and alerts can filter on.

## Set Up New Relic

```json
{
  "observability_enabled": true,
  "observability_provider": "newrelic",
  "observability_dsn": "YOUR_NEW_RELIC_LICENSE_KEY",
  "observability_environment": "production",
  "observability_release": "v1.2.3",
  "observability_server_name": "cms-prod-01",
  "observability_tags": {
    "service": "modulacms"
  }
}
```

## Available Metrics

ModulaCMS collects metrics across three areas.

### HTTP Metrics

Every HTTP request records:

- **Request count** -- total requests by method, endpoint, and status code
- **Request latency** -- response time in milliseconds by method and endpoint
- **Error count** -- 4xx and 5xx responses by method, endpoint, and status code

### Database Metrics

Every database query records:

- **Query count** -- total queries by operation type (SELECT, INSERT, UPDATE, DELETE)
- **Query latency** -- execution time in milliseconds by operation
- **Error count** -- failed queries by operation

### Runtime Metrics

A background process samples system statistics every 15 seconds:

- **Memory usage** -- heap allocation in bytes
- **Goroutine count** -- number of active concurrent processes

## View the Metrics Endpoint

Retrieve a JSON snapshot of all collected metrics. Requires the `config:read` permission.

```bash
curl http://localhost:8080/api/v1/admin/metrics \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Use this endpoint for custom dashboards, alerting integrations, or periodic scraping.

## Sampling Rates

The `observability_sample_rate` and `observability_traces_rate` fields control what percentage of events and traces are forwarded to your provider:

- `1.0` sends everything (100%). Good for development and low-traffic staging environments.
- `0.1` sends 10%. A reasonable starting point for production.
- `0.0` disables forwarding for that category.

> **Good to know**: Sampling only affects what is sent to the external provider. Internal metric collection is always 100%.

## Health Check

`GET /api/v1/health` is a public endpoint (no authentication required) that reports the status of core subsystems:

```bash
curl http://localhost:8080/api/v1/health
```

```json
{
  "status": "ok",
  "checks": {
    "database": true,
    "storage": true,
    "plugins": true
  }
}
```

When a check fails, the endpoint returns HTTP 503 with details:

```json
{
  "status": "degraded",
  "checks": {
    "database": true,
    "storage": false,
    "plugins": true
  },
  "details": {
    "storage": "HeadBucket: connection refused"
  }
}
```

| Check | What it tests |
|-------|---------------|
| Database | Ping against the configured database (5-second timeout) |
| Storage | S3 HeadBucket call (only runs if S3 storage is configured) |
| Plugins | Plugin system health (only runs if plugins are enabled) |

Additional health endpoints for specific subsystems:

| Endpoint | Permission | Description |
|----------|------------|-------------|
| `GET /api/v1/media/health` | `media:admin` | Checks for orphaned files in storage |
| `GET /api/v1/deploy/health` | `deploy:read` | Reports deploy sync status |

## Request Tracking

ModulaCMS generates a unique ID for every HTTP request. The ID is:

1. Returned to the client as the `X-Request-ID` response header.
2. Included in audit trail records for change events triggered by that request.

Use the request ID to correlate log entries, audit events, and external monitoring traces for a single request.

## Global Tags

Use `observability_tags` to attach key-value metadata to all events and metrics. Tags are useful for filtering in multi-service dashboards:

```json
{
  "observability_tags": {
    "service": "modulacms",
    "region": "us-east-1",
    "team": "platform"
  }
}
```

## Next Steps

- [Configuration reference](../getting-started/configuration.md) -- all config fields
- [Deploy sync](deploy-sync.md) -- health checks for content sync between instances
- [S3 storage](s3-storage.md) -- configure media storage (affects the storage health check)
