# Publishing

Publish content to make it live, schedule future publications, track version history, and restore any previous version.

## Content status

Every content node has a status that controls its visibility:

| Status | Visibility |
|--------|------------|
| `draft` | Visible in the admin panel and TUI. Not served by the public delivery API. |
| `published` | Live and served by the public delivery API. |
| `scheduled` | Draft with a future publication time. Automatically transitions to `published` when the time arrives. |

ModulaCMS creates content with `draft` status by default.

> **Good to know**: `scheduled` appears in API responses for clarity, but the content is stored as `draft` internally until the scheduled time arrives.

## Publish content

Publishing transitions content from draft to published and makes it available through the delivery endpoint (`GET /api/v1/content/{slug}`).

When you publish, ModulaCMS sets the content to `published`, records who published it and when, creates a version snapshot of the current field values, and increments the revision counter.

```bash
curl -X POST http://localhost:8080/api/v1/content/publish \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM"
  }'
```

Response:

```json
{
  "status": "published",
  "version_number": 1,
  "content_version_id": "01JNRWCP5GNSY8Q6P0Z3B7L9EN",
  "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM"
}
```

To publish locale-specific content, include the `locale` field:

```bash
curl -X POST http://localhost:8080/api/v1/content/publish \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "locale": "fr-FR"
  }'
```

## Unpublish content

Unpublishing reverts content to draft status, removing it from the public delivery API. Existing version snapshots are preserved -- you don't lose history by unpublishing.

```bash
curl -X POST http://localhost:8080/api/v1/content/unpublish \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM"
  }'
```

## Schedule content

Set a future publication time. The content stays in draft until ModulaCMS automatically publishes it at the specified time.

```bash
curl -X POST http://localhost:8080/api/v1/content/schedule \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "publish_at": "2026-04-01T09:00:00Z"
  }'
```

Response:

```json
{
  "status": "scheduled",
  "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
  "publish_at": "2026-04-01T09:00:00Z"
}
```

> **Good to know**: The server checks for scheduled content on a periodic interval. There may be a short delay between the `publish_at` time and actual publication.

## Version history

Every publish and restore operation creates a **version snapshot** -- a frozen record of the content's field values at that point in time. Versions form a complete history that you can browse, restore from, or delete.

### List versions

Retrieve all version snapshots for a content node, ordered newest first:

```bash
curl "http://localhost:8080/api/v1/content/versions?q=01JNRWBM4FNRZ7R5N9X4C6K8DM" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Each version contains:

| Field | Description |
|-------|-------------|
| `content_version_id` | ID of this version snapshot |
| `version_number` | Sequential version number |
| `locale` | Locale code for this snapshot |
| `snapshot` | Serialized field values (JSON string) |
| `trigger` | What created this version: `publish`, `manual`, `restore`, or `schedule` |
| `label` | Optional human-readable name |
| `published` | Whether the content was published at snapshot time |
| `published_by` | ID of the user who triggered the snapshot |
| `date_created` | Timestamp of snapshot creation |

### Get a single version

```bash
curl "http://localhost:8080/api/v1/content/versions/?q=01JNRWCP5GNSY8Q6P0Z3B7L9EN" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Create a manual version

Save a checkpoint of the current field values without changing publish status:

```bash
curl -X POST http://localhost:8080/api/v1/content/versions \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "label": "Before major rewrite"
  }'
```

Use manual versions as save points before large edits. They work the same as publish-triggered versions -- you can restore from them at any time.

### Delete a version

Remove a historical snapshot. This doesn't affect the current content.

```bash
curl -X DELETE "http://localhost:8080/api/v1/content/versions/?q=01JNRWCP5GNSY8Q6P0Z3B7L9EN" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

> **Good to know**: Version snapshots are immutable after creation. They survive unpublish, restore, and republish cycles.

## Restore a version

Replace the current draft field values with those from a previous version snapshot:

```bash
curl -X POST http://localhost:8080/api/v1/content/restore \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "content_version_id": "01JNRWCP5GNSY8Q6P0Z3B7L9EN"
  }'
```

Response:

```json
{
  "status": "restored",
  "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
  "restored_version_id": "01JNRWCP5GNSY8Q6P0Z3B7L9EN",
  "fields_restored": 5,
  "unmapped_fields": []
}
```

Before overwriting the current field values, ModulaCMS automatically creates a snapshot of the current state. You can always undo a restore by restoring to this auto-created version.

After restoring, the content remains in its current publish status. Publish separately to make the restored content live.

### Unmapped fields

If the datatype schema changed after a version was created (fields added, removed, or renamed), some snapshot fields may not match the current schema. These appear in the `unmapped_fields` array and are skipped during restore. Fields in the current schema that don't exist in the snapshot are left unchanged.

## The publishing workflow

```
                    +-- Schedule ---> [scheduled] ---> auto-Publish
                    |
Create (draft) ---> Publish ---> [published]
                    ^                  |
                    |                  +-- Unpublish ---> [draft]
                    |
                    +-- Restore (from version) ---> [draft, fields overwritten]
```

Every state transition that changes content visibility creates a version snapshot. The version history is append-only.

## SDK examples

### Go

```go
import modula "github.com/hegner123/modulacms/sdks/go"

client, _ := modula.NewClient(modula.ClientConfig{
    BaseURL: "http://localhost:8080",
    APIKey:  "mcms_YOUR_API_KEY",
})

// Publish content
resp, err := client.Publishing.Publish(ctx, modula.PublishRequest{
    ContentDataID: "01JNRWBM4FNRZ7R5N9X4C6K8DM",
})

// Schedule for future publication
schedResp, err := client.Publishing.Schedule(ctx, modula.ScheduleRequest{
    ContentDataID: "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    PublishAt:     "2026-04-01T09:00:00Z",
})

// List version history
versions, err := client.Publishing.ListVersions(ctx, "01JNRWBM4FNRZ7R5N9X4C6K8DM")

// Create a manual checkpoint
version, err := client.Publishing.CreateVersion(ctx, modula.CreateVersionRequest{
    ContentDataID: "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    Label:         "Before major rewrite",
})

// Restore from a previous version
restoreResp, err := client.Publishing.Restore(ctx, modula.RestoreRequest{
    ContentDataID:    "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    ContentVersionID: "01JNRWCP5GNSY8Q6P0Z3B7L9EN",
})

// Unpublish
unpubResp, err := client.Publishing.Unpublish(ctx, modula.PublishRequest{
    ContentDataID: "01JNRWBM4FNRZ7R5N9X4C6K8DM",
})
```

### TypeScript

```typescript
import { ModulaCMSAdmin } from '@modulacms/admin-sdk'

const client = new ModulaCMSAdmin({
  baseUrl: 'http://localhost:8080',
  apiKey: 'mcms_YOUR_API_KEY',
})

// Publish content
const resp = await client.publishing.publish({
  content_data_id: '01JNRWBM4FNRZ7R5N9X4C6K8DM',
})

// Schedule for future publication
const schedResp = await client.publishing.schedule({
  content_data_id: '01JNRWBM4FNRZ7R5N9X4C6K8DM',
  publish_at: '2026-04-01T09:00:00Z',
})

// List version history
const versions = await client.publishing.listVersions('01JNRWBM4FNRZ7R5N9X4C6K8DM')

// Restore from a previous version
const restoreResp = await client.publishing.restore({
  content_data_id: '01JNRWBM4FNRZ7R5N9X4C6K8DM',
  content_version_id: '01JNRWCP5GNSY8Q6P0Z3B7L9EN',
})
```

## Admin content publishing

The admin content system has its own parallel publishing lifecycle with identical operations. Replace `/api/v1/content/` with `/api/v1/admin/content/` in the endpoint paths:

| Operation | Public Endpoint | Admin Endpoint |
|-----------|----------------|----------------|
| Publish | `POST /api/v1/content/publish` | `POST /api/v1/admin/content/publish` |
| Unpublish | `POST /api/v1/content/unpublish` | `POST /api/v1/admin/content/unpublish` |
| Schedule | `POST /api/v1/content/schedule` | `POST /api/v1/admin/content/schedule` |
| List versions | `GET /api/v1/content/versions` | `GET /api/v1/admin/content/versions` |
| Create version | `POST /api/v1/content/versions` | `POST /api/v1/admin/content/versions` |
| Delete version | `DELETE /api/v1/content/versions/` | `DELETE /api/v1/admin/content/versions/` |
| Restore | `POST /api/v1/content/restore` | `POST /api/v1/admin/content/restore` |

Admin requests use `admin_content_data_id` and `admin_content_version_id` instead of `content_data_id` and `content_version_id`.

## Next steps

Learn how to [fetch and display your published content](/docs/building-content/serving-your-frontend) in a frontend application.
