# Webhook Integration

Recipes for setting up webhooks to receive HTTP notifications when CMS events occur. Webhooks enable external systems to react to content changes, media uploads, user actions, and other CMS events in real time.

## Webhook Concepts

**Webhook** -- A registered HTTP endpoint that receives POST requests when subscribed events occur in the CMS.

**Events** -- Strings identifying what happened (e.g., `content.published`, `media.created`, `user.created`). A webhook can subscribe to multiple event types.

**Secret** -- An HMAC-SHA256 signing key. When a secret is set, every delivery includes a signature header so the receiver can verify the payload was sent by the CMS and was not tampered with.

**Delivery** -- A single HTTP POST to the webhook URL for one event. Failed deliveries are retried with exponential backoff.

## Create a Webhook

**curl:**

```bash
curl -X POST http://localhost:8080/api/v1/admin/webhooks \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Deploy on publish",
    "url": "https://builds.example.com/hooks/deploy",
    "secret": "whsec_my-signing-secret",
    "events": ["content.published", "content.unpublished"],
    "is_active": true,
    "headers": {
      "X-Source": "modulacms"
    }
  }'
```

Response (201):

```json
{
  "webhook_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
  "name": "Deploy on publish",
  "url": "https://builds.example.com/hooks/deploy",
  "secret": "whsec_my-signing-secret",
  "events": ["content.published", "content.unpublished"],
  "is_active": true,
  "headers": {"X-Source": "modulacms"},
  "author_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD",
  "date_created": "2026-01-15T10:00:00Z",
  "date_modified": "2026-01-15T10:00:00Z"
}
```

**Go SDK:**

```go
webhook, err := client.Webhooks.Create(ctx, modula.CreateWebhookRequest{
    Name:     "Deploy on publish",
    URL:      "https://builds.example.com/hooks/deploy",
    Secret:   "whsec_my-signing-secret",
    Events:   []string{"content.published", "content.unpublished"},
    IsActive: true,
    Headers:  map[string]string{"X-Source": "modulacms"},
})
if err != nil {
    // handle error
}

fmt.Printf("Webhook created: %s\n", webhook.WebhookID)
```

**TypeScript SDK (admin):**

```typescript
const webhook = await admin.webhooks.create({
  name: 'Deploy on publish',
  url: 'https://builds.example.com/hooks/deploy',
  secret: 'whsec_my-signing-secret',
  events: ['content.published', 'content.unpublished'],
  is_active: true,
  headers: { 'X-Source': 'modulacms' },
})

console.log(`Webhook created: ${webhook.webhook_id}`)
```

## Test the Webhook

Send a test event to the webhook URL without creating a persistent delivery record. Returns the HTTP status code from the target.

**curl:**

```bash
curl -X POST "http://localhost:8080/api/v1/admin/webhooks/01HXK4N2F8RJZGP6VTQY3MCSW9/test" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{}'
```

Response:

```json
{
  "status": "success",
  "status_code": 200
}
```

**Go SDK:**

```go
result, err := client.Webhooks.Test(ctx, modula.WebhookID("01HXK4N2F8RJZGP6VTQY3MCSW9"))
if err != nil {
    // handle error
}

fmt.Printf("Test result: %s (HTTP %d)\n", result.Status, result.StatusCode)
```

**TypeScript SDK (admin):**

```typescript
const result = await admin.webhooks.test('01HXK4N2F8RJZGP6VTQY3MCSW9' as WebhookID)
console.log(`Test result: ${result.status} (HTTP ${result.status_code})`)
```

## Check Delivery History

List all delivery attempts for a webhook. Each delivery records the event, payload, status, attempt count, and any errors.

**curl:**

```bash
curl "http://localhost:8080/api/v1/admin/webhooks/01HXK4N2F8RJZGP6VTQY3MCSW9/deliveries" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

Response:

```json
[
  {
    "delivery_id": "01HXK8D2...",
    "webhook_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
    "event": "content.published",
    "payload": "{\"content_data_id\":\"01HXK5E3...\",\"slug\":\"homepage\"}",
    "status": "success",
    "attempts": 1,
    "last_status_code": 200,
    "last_error": "",
    "next_retry_at": "",
    "created_at": "2026-01-15T10:30:00Z",
    "completed_at": "2026-01-15T10:30:01Z"
  },
  {
    "delivery_id": "01HXK9F4...",
    "webhook_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
    "event": "content.published",
    "payload": "{\"content_data_id\":\"01HXK6G5...\",\"slug\":\"about\"}",
    "status": "failed",
    "attempts": 3,
    "last_status_code": 502,
    "last_error": "Bad Gateway",
    "next_retry_at": "2026-01-15T11:00:00Z",
    "created_at": "2026-01-15T10:45:00Z",
    "completed_at": ""
  }
]
```

**Go SDK:**

```go
deliveries, err := client.Webhooks.ListDeliveries(ctx, modula.WebhookID("01HXK4N2F8RJZGP6VTQY3MCSW9"))
if err != nil {
    // handle error
}

for _, d := range deliveries {
    fmt.Printf("[%s] %s: %s (attempts: %d, HTTP %d)\n",
        d.Status, d.Event, d.DeliveryID, d.Attempts, d.LastStatusCode)
}
```

**TypeScript SDK (admin):**

```typescript
const deliveries = await admin.webhooks.listDeliveries('01HXK4N2F8RJZGP6VTQY3MCSW9' as WebhookID)

for (const d of deliveries) {
  console.log(`[${d.status}] ${d.event}: ${d.delivery_id} (attempts: ${d.attempts}, HTTP ${d.last_status_code})`)
}
```

## Delivery Statuses

| Status | Description |
|--------|-------------|
| `pending` | Queued, not yet attempted |
| `success` | Delivered successfully (2xx response) |
| `failed` | All retry attempts exhausted |
| `retrying` | Failed, will be retried at `next_retry_at` |

## Retry a Failed Delivery

Manually retry a specific delivery. The CMS re-sends the original payload to the webhook URL.

**curl:**

```bash
curl -X POST "http://localhost:8080/api/v1/admin/webhooks/deliveries/01HXK9F4.../retry" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Go SDK:**

```go
err := client.Webhooks.RetryDelivery(ctx, modula.WebhookDeliveryID("01HXK9F4..."))
if err != nil {
    // handle error
}
```

**TypeScript SDK (admin):**

```typescript
await admin.webhooks.retryDelivery('01HXK9F4...' as WebhookDeliveryID)
```

## Example Webhook Receiver: Go

A minimal HTTP handler that receives webhook deliveries and verifies the HMAC-SHA256 signature.

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

const webhookSecret = "whsec_my-signing-secret"

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "failed to read body", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    // Verify signature
    signature := r.Header.Get("X-Webhook-Signature")
    if signature != "" {
        mac := hmac.New(sha256.New, []byte(webhookSecret))
        mac.Write(body)
        expected := hex.EncodeToString(mac.Sum(nil))
        if !hmac.Equal([]byte(signature), []byte(expected)) {
            http.Error(w, "invalid signature", http.StatusUnauthorized)
            return
        }
    }

    event := r.Header.Get("X-Webhook-Event")
    fmt.Printf("Received event: %s\n", event)
    fmt.Printf("Payload: %s\n", string(body))

    // Process the event
    switch event {
    case "content.published":
        // Trigger a site rebuild, invalidate cache, etc.
        log.Printf("Content published, triggering rebuild...")
    case "media.created":
        log.Printf("New media uploaded")
    default:
        log.Printf("Unhandled event: %s", event)
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"received": true}`))
}

func main() {
    http.HandleFunc("/hooks/deploy", webhookHandler)
    log.Println("Webhook receiver listening on :9090")
    log.Fatal(http.ListenAndServe(":9090", nil))
}
```

## Example Webhook Receiver: Node.js (Express)

```javascript
const express = require('express')
const crypto = require('crypto')

const app = express()
const WEBHOOK_SECRET = 'whsec_my-signing-secret'

// Use raw body for signature verification
app.use('/hooks/deploy', express.raw({ type: 'application/json' }))

app.post('/hooks/deploy', (req, res) => {
  const signature = req.headers['x-webhook-signature']

  // Verify signature
  if (signature && WEBHOOK_SECRET) {
    const expected = crypto
      .createHmac('sha256', WEBHOOK_SECRET)
      .update(req.body)
      .digest('hex')

    if (!crypto.timingSafeEqual(Buffer.from(signature), Buffer.from(expected))) {
      return res.status(401).json({ error: 'invalid signature' })
    }
  }

  const event = req.headers['x-webhook-event']
  const payload = JSON.parse(req.body.toString())

  console.log(`Received event: ${event}`)
  console.log('Payload:', payload)

  switch (event) {
    case 'content.published':
      console.log('Content published, triggering rebuild...')
      // Trigger site rebuild
      break
    case 'media.created':
      console.log('New media uploaded')
      break
    default:
      console.log(`Unhandled event: ${event}`)
  }

  res.json({ received: true })
})

app.listen(9090, () => {
  console.log('Webhook receiver listening on :9090')
})
```

## Verify Webhook Signatures

When a webhook has a secret configured, every delivery includes an `X-Webhook-Signature` header containing the HMAC-SHA256 hex digest of the request body, keyed with the webhook secret.

Verification steps:

1. Read the raw request body (before JSON parsing).
2. Compute `HMAC-SHA256(secret, body)` and hex-encode the result.
3. Compare the computed digest to the `X-Webhook-Signature` header using a constant-time comparison.
4. Reject the request if they do not match.

**Go:**

```go
func verifySignature(body []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expected := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expected))
}
```

**Node.js:**

```javascript
function verifySignature(body, signature, secret) {
  const expected = crypto
    .createHmac('sha256', secret)
    .update(body)
    .digest('hex')
  return crypto.timingSafeEqual(Buffer.from(signature), Buffer.from(expected))
}
```

**Python:**

```python
import hmac
import hashlib

def verify_signature(body: bytes, signature: str, secret: str) -> bool:
    expected = hmac.new(secret.encode(), body, hashlib.sha256).hexdigest()
    return hmac.compare_digest(signature, expected)
```

## Update a Webhook

**curl:**

```bash
curl -X PUT "http://localhost:8080/api/v1/admin/webhooks/01HXK4N2F8RJZGP6VTQY3MCSW9" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_id": "01HXK4N2F8RJZGP6VTQY3MCSW9",
    "name": "Deploy on publish (v2)",
    "url": "https://builds.example.com/hooks/deploy-v2",
    "events": ["content.published", "content.unpublished", "media.created"],
    "is_active": true
  }'
```

**Go SDK:**

```go
updated, err := client.Webhooks.Update(ctx, modula.UpdateWebhookRequest{
    WebhookID: modula.WebhookID("01HXK4N2F8RJZGP6VTQY3MCSW9"),
    Name:      "Deploy on publish (v2)",
    URL:       "https://builds.example.com/hooks/deploy-v2",
    Events:    []string{"content.published", "content.unpublished", "media.created"},
    IsActive:  true,
})
```

**TypeScript SDK (admin):**

```typescript
const updated = await admin.webhooks.update({
  webhook_id: '01HXK4N2F8RJZGP6VTQY3MCSW9' as WebhookID,
  name: 'Deploy on publish (v2)',
  url: 'https://builds.example.com/hooks/deploy-v2',
  events: ['content.published', 'content.unpublished', 'media.created'],
  is_active: true,
})
```

## Delete a Webhook

**curl:**

```bash
curl -X DELETE "http://localhost:8080/api/v1/admin/webhooks/01HXK4N2F8RJZGP6VTQY3MCSW9" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

**Go SDK:**

```go
err := client.Webhooks.Delete(ctx, modula.WebhookID("01HXK4N2F8RJZGP6VTQY3MCSW9"))
```

**TypeScript SDK (admin):**

```typescript
await admin.webhooks.remove('01HXK4N2F8RJZGP6VTQY3MCSW9' as WebhookID)
```

## Next Steps

- [Fetching Content](fetching-content.md) -- content retrieval recipes
- [User Management](user-management.md) -- user and role management
