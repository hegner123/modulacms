# Webhooks

Send HTTP notifications to external systems when content changes, media uploads, or other CMS events occur.

## Concepts

**Webhook** -- A registered HTTP endpoint that receives POST requests when subscribed events occur. Each webhook has a URL, a list of events, an optional signing secret, and optional custom headers.

**Event** -- A named occurrence in the CMS lifecycle (e.g., `content.published`, `content.deleted`). Webhooks subscribe to specific events and only receive payloads for those events.

**Delivery** -- A single attempt to send an event payload to a webhook URL. Each delivery records the HTTP status code, response timing, and error information.

**Secret** -- An optional string used to sign webhook payloads with HMAC-SHA256. The signature is sent in the `X-ModulaCMS-Signature` header so the receiver can verify the payload came from ModulaCMS.

## Enable Webhooks

Webhooks require `webhook_enabled: true` in `modula.config.json`. Without it, all webhook events are silently ignored.

```json
{
  "webhook_enabled": true
}
```

### Webhook Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `webhook_enabled` | bool | `false` | Enable the webhook dispatcher |
| `webhook_allow_http` | bool | `false` | Allow HTTP (non-HTTPS) webhook URLs |
| `webhook_timeout` | int | `10` | Timeout in seconds for delivery requests |
| `webhook_max_retries` | int | `3` | Maximum delivery attempts before marking as failed |
| `webhook_workers` | int | `4` | Number of concurrent delivery workers |
| `webhook_delivery_retention_days` | int | `30` | Days to retain delivery history records |

## Event Types

| Event | Fires when |
|-------|-----------|
| `content.published` | Content transitions from draft to published |
| `content.unpublished` | Content reverts from published to draft |
| `content.updated` | Published content fields are updated |
| `content.scheduled` | Content is scheduled for future publication |
| `content.deleted` | Content is deleted |
| `locale.published` | Locale-specific content is published |
| `version.created` | A new version snapshot is created |
| `admin.content.published` | Admin content is published |
| `admin.content.unpublished` | Admin content is unpublished |
| `admin.content.updated` | Admin content is updated |
| `admin.content.deleted` | Admin content is deleted |
| `webhook.test` | Synthetic test event (sent by the test endpoint) |

Use `["*"]` as the events list to subscribe to all event types, including any added in future versions.

## Create a Webhook

```bash
curl -X POST http://localhost:8080/api/v1/admin/webhooks \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Deploy trigger",
    "url": "https://ci.example.com/hooks/modulacms",
    "secret": "whsec_your_secret_key",
    "events": ["content.published", "content.unpublished"],
    "is_active": true,
    "headers": {
      "X-Custom-Header": "my-value"
    }
  }'
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Human-readable name |
| `url` | string | Yes | HTTPS endpoint that receives POST requests |
| `secret` | string | No | Secret for HMAC-SHA256 signing. Auto-generated if omitted. |
| `events` | string[] | Yes | Event types to subscribe to (or `["*"]` for all) |
| `is_active` | bool | No | Whether the webhook is active (defaults to false) |
| `headers` | object | No | Custom HTTP headers sent with each delivery |

> **Good to know**: Webhook URLs are validated to prevent server-side request forgery (SSRF). Private IP ranges, loopback addresses, and non-HTTPS URLs are blocked by default. Set `webhook_allow_http: true` in config to allow HTTP URLs during development.

## Test a Webhook

Verify the endpoint is reachable before waiting for real events:

```bash
curl -X POST http://localhost:8080/api/v1/admin/webhooks/01JNRWDP6HMTY9S7Q1Z4B8K5FR/test \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

```json
{
  "status": "success",
  "status_code": 200,
  "duration": "142ms"
}
```

The test sends a `webhook.test` event synchronously and reports the result immediately.

## Payload Structure

When a subscribed event occurs, ModulaCMS POSTs a JSON payload to the webhook URL:

```json
{
  "id": "01JNRWEP7INUZ0T8R2A5C9L6GS",
  "event": "content.published",
  "occurred_at": "2026-03-07T14:30:00Z",
  "data": {
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "content_version_id": "01JNRWEP7HNTZ0T8R2A5C9L6GT",
    "version_number": 3,
    "locale": "",
    "published_by": "01JNRWAM3ENRZ7R5N9X4C6K8DL"
  }
}
```

### Request Headers

Each delivery includes these headers:

| Header | Description |
|--------|-------------|
| `Content-Type` | `application/json` |
| `X-ModulaCMS-Signature` | HMAC-SHA256 hex digest of the payload body |
| `X-ModulaCMS-Event` | The event type (e.g., `content.published`) |
| `User-Agent` | `ModulaCMS-Webhook/1.0` |

Custom headers configured on the webhook are also included.

## Verify Signatures

When a secret is configured, every payload is signed with HMAC-SHA256. Verify the signature on your receiving end to confirm the payload came from ModulaCMS.

1. Read the raw request body (before JSON parsing).
2. Compute `HMAC-SHA256(secret, body)` and hex-encode the result.
3. Compare the computed digest to the `X-ModulaCMS-Signature` header using a constant-time comparison.
4. Reject the request if they do not match.

### Go

```go
func verifySignature(body []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expected := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

### Node.js

```javascript
function verifySignature(body, signature, secret) {
  const expected = crypto
    .createHmac('sha256', secret)
    .update(body)
    .digest('hex')
  return crypto.timingSafeEqual(Buffer.from(signature), Buffer.from(expected))
}
```

### Python

```python
import hmac
import hashlib

def verify_signature(body: bytes, signature: str, secret: str) -> bool:
    expected = hmac.new(secret.encode(), body, hashlib.sha256).hexdigest()
    return hmac.compare_digest(signature, expected)
```

## Example Receiver: Go

A minimal HTTP server that receives webhook deliveries and verifies signatures:

```go
package main

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "log"
    "net/http"
)

const webhookSecret = "whsec_your_secret_key"

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "failed to read body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    // Verify signature
    signature := r.Header.Get("X-ModulaCMS-Signature")
    if signature != "" {
        mac := hmac.New(sha256.New, []byte(webhookSecret))
        mac.Write(body)
        expected := hex.EncodeToString(mac.Sum(nil))
        if !hmac.Equal([]byte(signature), []byte(expected)) {
            http.Error(w, "invalid signature", http.StatusUnauthorized)
            return
        }
    }

    event := r.Header.Get("X-ModulaCMS-Event")
    fmt.Printf("Received event: %s\n", event)
    fmt.Printf("Payload: %s\n", string(body))

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"received": true}`))
}

func main() {
    http.HandleFunc("/hooks/deploy", webhookHandler)
    log.Println("Webhook receiver listening on :9090")
    log.Fatal(http.ListenAndServe(":9090", nil))
}
```

## Example Receiver: Node.js

```javascript
const express = require('express')
const crypto = require('crypto')

const app = express()
const WEBHOOK_SECRET = 'whsec_your_secret_key'

// Use raw body for signature verification
app.use('/hooks/deploy', express.raw({ type: 'application/json' }))

app.post('/hooks/deploy', (req, res) => {
  const signature = req.headers['x-modulacms-signature']

  if (signature && WEBHOOK_SECRET) {
    const expected = crypto
      .createHmac('sha256', WEBHOOK_SECRET)
      .update(req.body)
      .digest('hex')

    if (!crypto.timingSafeEqual(Buffer.from(signature), Buffer.from(expected))) {
      return res.status(401).json({ error: 'invalid signature' })
    }
  }

  const event = req.headers['x-modulacms-event']
  const payload = JSON.parse(req.body.toString())

  console.log(`Received event: ${event}`)
  console.log('Payload:', payload)

  res.json({ received: true })
})

app.listen(9090, () => {
  console.log('Webhook receiver listening on :9090')
})
```

## Delivery and Retries

### Delivery Statuses

| Status | Description |
|--------|-------------|
| `pending` | Queued for delivery |
| `success` | Delivered successfully (2xx response) |
| `failed` | Permanently failed after exhausting all retry attempts |
| `retrying` | Queued for retry after a previous failure |

### Automatic Retries

Failed deliveries are retried with exponential backoff: 1 minute, 5 minutes, then 30 minutes. After `webhook_max_retries` attempts (default 3), the delivery is marked as permanently failed.

### View Delivery History

```bash
curl http://localhost:8080/api/v1/admin/webhooks/01JNRWDP6HMTY9S7Q1Z4B8K5FR/deliveries \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Each delivery record contains:

| Field | Description |
|-------|-------------|
| `delivery_id` | Unique ID of this delivery attempt |
| `event` | Event type that triggered the delivery |
| `status` | Delivery status |
| `attempts` | Number of delivery attempts made |
| `last_status_code` | HTTP status code from the most recent attempt |
| `last_error` | Error message from the most recent failed attempt |
| `next_retry_at` | Scheduled time for the next retry (if retrying) |
| `created_at` | When the delivery was created |
| `completed_at` | When the delivery succeeded or was abandoned |

### Retry a Failed Delivery

Re-enqueue a failed delivery for another attempt:

```bash
curl -X POST http://localhost:8080/api/v1/admin/webhooks/deliveries/01JNRWEP7INUZ0T8R2A5C9L6GS/retry \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Only failed deliveries can be retried. Attempting to retry a successful delivery returns an error.

## Manage Webhooks

### Update a Webhook

```bash
curl -X PUT http://localhost:8080/api/v1/admin/webhooks/01JNRWDP6HMTY9S7Q1Z4B8K5FR \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Deploy trigger (updated)",
    "url": "https://ci.example.com/hooks/modulacms",
    "events": ["content.published", "content.unpublished", "content.updated"],
    "is_active": true
  }'
```

### Delete a Webhook

```bash
curl -X DELETE http://localhost:8080/api/v1/admin/webhooks/01JNRWDP6HMTY9S7Q1Z4B8K5FR \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

## API Reference

All webhook endpoints require authentication and `webhook:*` permissions (admin-only by default).

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/admin/webhooks` | List all webhooks |
| POST | `/api/v1/admin/webhooks` | Create a webhook |
| GET | `/api/v1/admin/webhooks/{id}` | Get a webhook |
| PUT | `/api/v1/admin/webhooks/{id}` | Update a webhook |
| DELETE | `/api/v1/admin/webhooks/{id}` | Delete a webhook |
| POST | `/api/v1/admin/webhooks/{id}/test` | Send a test event |
| GET | `/api/v1/admin/webhooks/{id}/deliveries` | List delivery history |
| POST | `/api/v1/admin/webhooks/deliveries/{id}/retry` | Retry a failed delivery |

## Next Steps

- [Deploy sync](deploy-sync.md) -- export and import content between instances
- [Observability](observability.md) -- monitor webhook delivery health
- [Configuration reference](../getting-started/configuration.md) -- all config fields
