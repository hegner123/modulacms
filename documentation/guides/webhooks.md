# Webhooks

ModulaCMS supports event-driven webhooks that send HTTP POST notifications to external URLs when CMS events occur. Webhooks are useful for triggering static site rebuilds, syncing content to external systems, sending notifications, or integrating with CI/CD pipelines.

## Concepts

**Webhook** -- A registered endpoint that receives HTTP POST notifications when subscribed events occur. Each webhook has a URL, a list of events to subscribe to, an optional secret for signature verification, and optional custom headers.

**Event** -- A named occurrence in the CMS lifecycle (e.g., `content.published`, `content.deleted`). Webhooks subscribe to specific events and only receive payloads for those events.

**Delivery** -- A single attempt to send an event payload to a webhook URL. Each delivery records the HTTP status code, response, timing, and error information. Failed deliveries can be retried.

**Secret** -- An optional string used to sign webhook payloads with HMAC-SHA256. The signature is sent in the `X-ModulaCMS-Signature` header, allowing the receiver to verify that the payload came from ModulaCMS and was not tampered with.

## Event Types

| Event | Fires when |
|-------|-----------|
| `content.published` | Content is published (transitions from draft to published) |
| `content.unpublished` | Content is unpublished (reverts from published to draft) |
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

Use `*` as a wildcard to subscribe to all events.

## Creating a Webhook

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
| `name` | string | Yes | Human-readable name for the webhook |
| `url` | string | Yes | HTTPS endpoint that receives POST requests |
| `secret` | string | No | Secret for HMAC-SHA256 payload signing. Auto-generated if omitted. |
| `events` | string[] | Yes | List of event types to subscribe to (or `["*"]` for all) |
| `is_active` | bool | No | Whether the webhook is active. Defaults to false. |
| `headers` | object | No | Custom HTTP headers sent with each delivery |

The URL is validated against SSRF rules. By default, only HTTPS URLs are accepted. HTTP URLs can be allowed via the `webhook_allow_http` configuration option.

## Managing Webhooks

### Listing Webhooks

```bash
curl http://localhost:8080/api/v1/admin/webhooks \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Getting a Single Webhook

```bash
curl http://localhost:8080/api/v1/admin/webhooks/01JNRWDP6HMTY9S7Q1Z4B8K5FR \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Updating a Webhook

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

### Deleting a Webhook

```bash
curl -X DELETE http://localhost:8080/api/v1/admin/webhooks/01JNRWDP6HMTY9S7Q1Z4B8K5FR \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

## Webhook Delivery

When a subscribed event occurs, ModulaCMS POSTs a JSON payload to the webhook URL.

### Payload Structure

```json
{
  "id": "01JNRWEP7INUZ0T8R2A5C9L6GS",
  "event": "content.published",
  "occurred_at": "2026-03-07T14:30:00Z",
  "data": {
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "version_number": 3
  }
}
```

### Request Headers

Each delivery includes these headers:

| Header | Description |
|--------|-------------|
| `Content-Type` | `application/json` |
| `X-ModulaCMS-Signature` | HMAC-SHA256 hex digest of the payload body, signed with the webhook secret |
| `X-ModulaCMS-Event` | The event type (e.g., `content.published`) |
| `User-Agent` | `ModulaCMS-Webhook/1.0` |

Any custom headers configured on the webhook are also included.

### Delivery Statuses

| Status | Description |
|--------|-------------|
| `pending` | Queued for delivery |
| `success` | Delivered successfully (2xx response) |
| `failed` | Delivery failed (non-2xx response or network error) |
| `retrying` | Queued for retry after a previous failure |

## Delivery History

View the delivery history for a webhook:

```bash
curl http://localhost:8080/api/v1/admin/webhooks/01JNRWDP6HMTY9S7Q1Z4B8K5FR/deliveries \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Each delivery record contains:

| Field | Description |
|-------|-------------|
| `delivery_id` | ULID of this delivery attempt |
| `webhook_id` | ULID of the webhook |
| `event` | Event type that triggered this delivery |
| `payload` | The JSON payload that was sent |
| `status` | Delivery status (`pending`, `success`, `failed`, `retrying`) |
| `attempts` | Number of delivery attempts made |
| `last_status_code` | HTTP status code from the most recent attempt |
| `last_error` | Error message from the most recent failed attempt |
| `next_retry_at` | Scheduled time for the next retry (if retrying) |
| `created_at` | When the delivery was created |
| `completed_at` | When the delivery succeeded or was abandoned |

## Retrying Failed Deliveries

Re-enqueue a failed delivery for another attempt:

```bash
curl -X POST http://localhost:8080/api/v1/admin/webhooks/deliveries/01JNRWEP7INUZ0T8R2A5C9L6GS/retry \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Only failed deliveries can be retried. Attempting to retry a successful delivery returns an error.

## Testing Webhooks

Send a synthetic test event to verify the webhook endpoint is reachable and responds correctly:

```bash
curl -X POST http://localhost:8080/api/v1/admin/webhooks/01JNRWDP6HMTY9S7Q1Z4B8K5FR/test \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Response:

```json
{
  "status": "success",
  "status_code": 200,
  "duration": "142ms"
}
```

The test event uses the `webhook.test` event type. The payload includes the webhook ID and a test message. The test is synchronous -- it waits for the endpoint to respond and reports the result immediately.

## Securing Webhook Endpoints

When a secret is configured on a webhook, every payload is signed with HMAC-SHA256. To verify the signature on your receiving end:

1. Read the raw request body.
2. Compute the HMAC-SHA256 of the body using the webhook secret as the key.
3. Compare the hex-encoded result to the `X-ModulaCMS-Signature` header value.
4. Reject the request if the signatures do not match.

Example verification in Go:

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "io"
    "net/http"
)

func verifyWebhook(r *http.Request, secret string) bool {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        return false
    }

    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expected := hex.EncodeToString(mac.Sum(nil))

    signature := r.Header.Get("X-ModulaCMS-Signature")
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

Example verification in Node.js:

```javascript
const crypto = require('crypto')

function verifyWebhook(body, signature, secret) {
  const expected = crypto
    .createHmac('sha256', secret)
    .update(body)
    .digest('hex')
  return crypto.timingSafeEqual(
    Buffer.from(expected),
    Buffer.from(signature)
  )
}
```

## SDK Examples

### Go

```go
import modula "github.com/hegner123/modulacms/sdks/go"

client, _ := modula.NewClient(modula.ClientConfig{
    BaseURL: "http://localhost:8080",
    APIKey:  "mcms_YOUR_API_KEY",
})

// Create a webhook
webhook, err := client.Webhooks.Create(ctx, modula.CreateWebhookRequest{
    Name:     "Deploy trigger",
    URL:      "https://ci.example.com/hooks/modulacms",
    Events:   []string{"content.published", "content.unpublished"},
    IsActive: true,
})

// List all webhooks
webhooks, err := client.Webhooks.List(ctx, nil)

// Test a webhook
testResp, err := client.Webhooks.Test(ctx, webhook.WebhookID)

// View delivery history
deliveries, err := client.Webhooks.ListDeliveries(ctx, webhook.WebhookID)

// Retry a failed delivery
err = client.Webhooks.RetryDelivery(ctx, deliveryID)

// Delete a webhook
err = client.Webhooks.Delete(ctx, webhook.WebhookID)
```

### TypeScript

```typescript
import { ModulaCMSAdmin } from '@modulacms/admin-sdk'

const client = new ModulaCMSAdmin({
  baseUrl: 'http://localhost:8080',
  apiKey: 'mcms_YOUR_API_KEY',
})

// Create a webhook
const webhook = await client.webhooks.create({
  name: 'Deploy trigger',
  url: 'https://ci.example.com/hooks/modulacms',
  events: ['content.published', 'content.unpublished'],
  is_active: true,
})

// List all webhooks
const webhooks = await client.webhooks.list()

// Test a webhook
const testResp = await client.webhooks.test(webhook.webhook_id)

// View delivery history
const deliveries = await client.webhooks.listDeliveries(webhook.webhook_id)

// Retry a failed delivery
await client.webhooks.retryDelivery(deliveryId)
```

## API Reference

All webhook endpoints require authentication and `webhooks:*` permissions (admin-only by default).

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

## Configuration

Webhook-related configuration in `modula.config.json`:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `webhook_allow_http` | bool | `false` | Allow HTTP (non-HTTPS) webhook URLs |
| `webhook_timeout` | int | 10 | Timeout in seconds for webhook delivery requests |

## Notes

- **Auto-generated secrets.** If no secret is provided when creating a webhook, the server generates a 32-byte random hex string.
- **SSRF protection.** Webhook URLs are validated to prevent server-side request forgery. Private IP ranges, loopback addresses, and non-HTTPS URLs are blocked by default.
- **Wildcard events.** Use `["*"]` in the events list to subscribe to all event types, including any added in future versions.
- **No redirects.** Webhook delivery does not follow HTTP redirects. A redirect response is treated as a non-2xx status.
