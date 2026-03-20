# Forms Plugin

Collect form submissions, validate input, and queue webhook deliveries from any frontend.

The forms plugin ships with ModulaCMS as a Lua plugin at `plugins/forms/`. It provides 21 REST API endpoints for building forms, managing fields, accepting submissions, and configuring webhook notifications. A companion web components package (`@modulacms/forms`) renders forms, displays entries, and provides a drag-and-drop form builder for admin panels.

> **Good to know**: The forms plugin requires `plugin_enabled: true` in `modula.config.json`. See [configuration](configuration.md) for plugin setup.

## Quick Start

### 1. Enable and Approve

```bash
# Start the server with plugins enabled
modulacms serve

# Verify the plugin loaded
modulacms plugin info forms

# Approve all routes
modulacms plugin approve forms --all-routes
```

### 2. Create a Form

```bash
curl -X POST http://localhost:8080/api/v1/plugins/forms/forms \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_KEY" \
  -d '{
    "name": "Contact",
    "description": "Contact form for the marketing site",
    "submit_label": "Send Message",
    "success_message": "Thanks for reaching out. We will get back to you within 24 hours."
  }'
```

Response:

```json
{
  "id": "01HXYZ...",
  "name": "Contact",
  "description": "Contact form for the marketing site",
  "submit_label": "Send Message",
  "success_message": "Thanks for reaching out. We will get back to you within 24 hours.",
  "redirect_url": null,
  "enabled": true,
  "version": 1,
  "rate_limit": 100,
  "created_at": "2026-03-19T12:00:00Z"
}
```

### 3. Add Fields

```bash
# Email field (required)
curl -X POST http://localhost:8080/api/v1/plugins/forms/forms/01HXYZ.../fields \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_KEY" \
  -d '{
    "name": "email",
    "label": "Email Address",
    "field_type": "email",
    "required": true,
    "placeholder": "you@example.com",
    "version": 1
  }'

# Message field (required, with length limits)
curl -X POST http://localhost:8080/api/v1/plugins/forms/forms/01HXYZ.../fields \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_KEY" \
  -d '{
    "name": "message",
    "label": "Message",
    "field_type": "textarea",
    "required": true,
    "help_text": "Tell us how we can help.",
    "validation_rules": {"min_length": 10, "max_length": 2000},
    "version": 2
  }'
```

> **Good to know**: Every field mutation requires a `version` field matching the current form version. The response includes the new version number. This prevents two people from editing the same form simultaneously without knowing about each other's changes.

### 4. Accept Submissions

The public submit endpoint requires no authentication:

```bash
curl -X POST http://localhost:8080/api/v1/plugins/forms/forms/01HXYZ.../submit \
  -H "Content-Type: application/json" \
  -d '{
    "email": "visitor@example.com",
    "message": "I am interested in your services."
  }'
```

Response:

```json
{
  "id": "01HABCD...",
  "message": "Thanks for reaching out. We will get back to you within 24 hours.",
  "redirect_url": null
}
```

## Concepts

### Form Versioning

Every form has a `version` number that increments on every change to the form or its fields. When someone submits an entry, the submission records which version it was submitted against (`form_version`). This lets you correlate entries with the field configuration that was active when they were submitted.

Endpoints that modify forms or fields require the current version in the request body. If the version does not match, the server returns `409 Conflict` with the current version number so the client can refresh and retry.

### Field Types

| Type | HTML Element | Validation |
|------|-------------|------------|
| `text` | `<input type="text">` | Optional min/max length |
| `textarea` | `<textarea>` | Optional min/max length |
| `email` | `<input type="email">` | Must contain `@` with valid local and domain parts |
| `number` | `<input type="number">` | Must be a valid number (integer or decimal) |
| `tel` | `<input type="tel">` | No format enforcement |
| `url` | `<input type="url">` | Must start with `http://` or `https://` |
| `date` | `<input type="date">` | No format enforcement |
| `time` | `<input type="time">` | No format enforcement |
| `datetime` | `<input type="datetime-local">` | No format enforcement |
| `select` | `<select>` | Value must be in the `options` array |
| `radio` | Radio button group | Value must be in the `options` array |
| `checkbox` | `<input type="checkbox">` | When required, must be checked |
| `hidden` | `<input type="hidden">` | No validation |
| `file` | `<input type="file">` | Base64 encoded, ~750KB max per submission |

Fields with `select` or `radio` types require an `options` array in the field definition:

```json
{
  "name": "department",
  "label": "Department",
  "field_type": "select",
  "options": ["Sales", "Support", "Engineering", "Other"],
  "required": true,
  "version": 3
}
```

### Validation Rules

Each field can include a `validation_rules` object:

| Rule | Type | Description |
|------|------|-------------|
| `min_length` | number | Minimum string length |
| `max_length` | number | Maximum string length |
| `max_file_size` | number | Maximum file size in bytes (file fields only, default 768000) |

Validation runs server-side on every submission. The web components also validate client-side before sending, using the same rules.

### Anti-Spam Protection

Three layers protect public submit endpoints:

1. **Honeypot field**: A hidden `_hp` field that bots tend to fill. If populated, the server returns a fake success response (identical to a real one) without creating an entry.

2. **Per-IP rate limiting**: The plugin HTTP bridge applies automatic per-IP throttling to all plugin routes.

3. **Per-form submission throttle**: Each form has a `rate_limit` (default 100 submissions per hour). When the limit is reached, the server returns `429 Too Many Requests`. The counter resets automatically after one hour.

### Webhook Queue

When a submission is created or deleted, the plugin inserts rows into a webhook delivery queue. A separate process (not part of the plugin) reads the queue and delivers HTTP requests. The plugin is responsible for queue population only.

Each queue row contains the full delivery configuration (URL, method, headers, secret) so it remains deliverable even if the webhook configuration is later deleted.

**Webhook events:**

| Event | Trigger |
|-------|---------|
| `entry.created` | New form submission |
| `entry.deleted` | Entry deleted via admin API |
| `form.deleted` | Form deleted (one event per active webhook, not per entry) |

## API Reference

All endpoints are under `/api/v1/plugins/forms/`.

### Public Endpoints

These require no authentication.

#### GET /forms/{id}/public

Retrieve an enabled form with its fields for rendering. Returns the form definition, field list, and CAPTCHA configuration (if any). Does not include admin-only fields like `captcha_secret` or rate limit counters.

#### POST /forms/{id}/submit

Submit a form entry. The request body is a JSON object mapping field names to values:

```json
{
  "email": "visitor@example.com",
  "message": "Hello",
  "_hp": ""
}
```

Returns `201` with `{id, message, redirect_url}` on success, `400` for validation errors, `429` when rate limited.

### Admin Endpoints

All admin endpoints require authentication via `X-API-Key` header or session cookie.

#### Forms

| Method | Path | Description |
|--------|------|-------------|
| GET | `/forms` | List forms (paginated: `?limit=`, `?offset=`) |
| POST | `/forms` | Create a form |
| GET | `/forms/{id}` | Get form with fields |
| PUT | `/forms/{id}` | Update form (requires `version` in body) |
| DELETE | `/forms/{id}` | Delete form and all related data |

**Updatable form fields**: `name`, `description`, `submit_label`, `success_message`, `redirect_url`, `enabled`, `captcha_config`, `captcha_secret`, `rate_limit`.

#### Fields

| Method | Path | Description |
|--------|------|-------------|
| GET | `/forms/{id}/fields` | List fields ordered by position |
| POST | `/forms/{id}/fields` | Add a field (requires `version`) |
| PUT | `/fields/{id}` | Update a field (requires `version`) |
| DELETE | `/fields/{id}` | Delete a field (requires `version` in body) |
| POST | `/forms/{id}/fields/reorder` | Reorder all fields |

**Reorder** requires `field_ids` (array of all field IDs in desired order) and `version`:

```json
{
  "field_ids": ["01FIELD_C", "01FIELD_A", "01FIELD_B"],
  "version": 5
}
```

#### Entries

| Method | Path | Description |
|--------|------|-------------|
| GET | `/forms/{id}/entries` | List entries (paginated) |
| GET | `/entries/{id}` | Get a single entry |
| DELETE | `/entries/{id}` | Delete an entry |
| GET | `/forms/{id}/export` | Export entries as JSON |

**Export** uses cursor-based pagination. Each request returns up to 10,000 entries. Pass `?after=<last_entry_id>` for the next page:

```bash
# First page
curl "http://localhost:8080/api/v1/plugins/forms/forms/01HXYZ.../export"

# Next page
curl "http://localhost:8080/api/v1/plugins/forms/forms/01HXYZ.../export?after=01LAST_ID"
```

Response:

```json
{
  "items": [...],
  "count": 10000,
  "after": "01LAST_ENTRY_ID",
  "has_more": true
}
```

#### Webhooks

| Method | Path | Description |
|--------|------|-------------|
| GET | `/forms/{id}/webhooks` | List webhooks for a form |
| POST | `/forms/{id}/webhooks` | Create a webhook |
| PUT | `/webhooks/{id}` | Update a webhook |
| DELETE | `/webhooks/{id}` | Delete a webhook |
| GET | `/forms/{id}/webhooks/queue` | Queue depth and recent failures |

**Create webhook example:**

```json
{
  "url": "https://hooks.slack.com/services/T00/B00/xxxx",
  "method": "POST",
  "events": "entry.created",
  "headers": {"Authorization": "Bearer token"},
  "secret": "whsec_signing_key"
}
```

**Queue info** returns pending count, failed count, and the 10 most recent failures:

```json
{
  "pending": 3,
  "failed": 1,
  "recent_failures": [
    {
      "id": "01QUEUE...",
      "webhook_id": "01WH...",
      "event": "entry.created",
      "last_error": "connection refused",
      "attempts": 3,
      "created_at": "2026-03-19T12:00:00Z"
    }
  ]
}
```

### Response Formats

**Success (create):** `201` with the created object including `id`.

**Success (update):** `200` with the full updated object.

**Success (delete):** `200` with `{"deleted": true}`.

**Success (list):** `200` with `{"items": [...], "total": N, "limit": N, "offset": N}`.

**Error:** `{"error": "description of what went wrong"}` with appropriate status code.

**Version conflict:** `409` with `{"error": "version conflict", "current_version": N}`.

## Configuration

Forms plugin behavior is controlled by global plugin settings in `modula.config.json`:

| Setting | Default | Description |
|---------|---------|-------------|
| `plugin_max_ops` | 1000 | Maximum database operations per request |
| `plugin_max_request_body` | 1MB | Maximum request body size (limits file upload size) |
| `plugin_rate_limit` | varies | Per-IP rate limiting for all plugin routes |

Per-form settings:

| Setting | Default | Description |
|---------|---------|-------------|
| `rate_limit` | 100 | Maximum submissions per hour per form |
| `enabled` | true | Whether the form accepts submissions |

### CAPTCHA

To add CAPTCHA to a form, update `captcha_config` with your provider details:

```json
{
  "captcha_config": {
    "provider": "recaptcha",
    "site_key": "6Lc..."
  },
  "captcha_secret": "6Lc..._secret",
  "version": 4
}
```

The web components read `captcha_config` and render the provider's challenge widget automatically. Server-side CAPTCHA validation is planned for a future release.

> **Good to know**: Store the `captcha_secret` separately from `captcha_config`. The public form endpoint returns `captcha_config` (with `site_key`) but never exposes `captcha_secret`.

## Known Limitations

- **File uploads** are stored as base64 in the database. The maximum total submission size is ~750KB due to the 1MB request body limit. For larger files, use the CMS media system instead.
- **Webhook delivery** is handled by a separate queue processor, not by the plugin itself. The plugin inserts queue rows only.
- **Server-side CAPTCHA validation** is not available yet. The CAPTCHA widget renders client-side for UX consistency.
- **Form deletion** queues a single `form.deleted` event per webhook, not per-entry events. Export entries before deleting if downstream systems need entry-level notifications.
- **Secrets are stored as plaintext.** Webhook signing secrets and CAPTCHA secret keys are stored unencrypted. Restrict database access and treat backups as sensitive.

## Next Steps

- [Embed forms on your site](../integrations/forms-components.md) using the `@modulacms/forms` web components
- [Lua API reference](lua-api.md) for extending plugin behavior
- [Plugin approval](approval.md) for managing route permissions
- [Webhook integration](../integrations/webhooks.md) for processing the delivery queue
