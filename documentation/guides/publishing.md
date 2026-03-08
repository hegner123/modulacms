# Publishing and Versioning

ModulaCMS manages content through a lifecycle of statuses -- draft, published, and scheduled. Publishing creates an immutable version snapshot of the content's current field values. Versions form a history that you can browse, restore from, or delete. The same operations exist for both public content and admin content, accessed through separate endpoint paths.

## Concepts

**Content status** -- Every content data record has a status field. The two primary statuses are `draft` (editable, not visible through the public delivery API) and `published` (live, returned by the content delivery endpoint). A third status, `scheduled`, indicates the content will be automatically published at a future time.

**Version snapshot** -- An immutable, point-in-time copy of a content node's field values. Versions are created automatically when you publish and can also be created manually as checkpoints. Each version has a version number, a trigger (what caused the snapshot), and the serialized field values.

**Restore** -- Replaces the current draft field values with those from a previous version snapshot. Restoring does not change the content's publish status -- you must publish separately after restoring.

## Content Statuses

| Status | Description |
|--------|-------------|
| `draft` | Editable. Not visible through the public content delivery API. |
| `published` | Live. Returned by the content delivery endpoint (`/api/v1/content/{slug}`). |
| `scheduled` | Draft with a future `publish_at` time. Automatically transitions to `published` when the time arrives. |

## Publishing Content

Publishing transitions content from draft to published and creates a version snapshot.

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

## Unpublishing Content

Unpublishing reverts content from published to draft, removing it from the public delivery API. Existing version snapshots are preserved.

```bash
curl -X POST http://localhost:8080/api/v1/content/unpublish \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM"
  }'
```

## Scheduling Content

Scheduling sets a future publication time. The content remains in draft until the server automatically publishes it at the specified time. A version snapshot is created at the time of the schedule call.

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

## Version History

### Listing Versions

Retrieve all version snapshots for a content node, ordered newest first:

```bash
curl "http://localhost:8080/api/v1/content/versions?q=01JNRWBM4FNRZ7R5N9X4C6K8DM" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Each version contains:

| Field | Description |
|-------|-------------|
| `content_version_id` | ULID of this version snapshot |
| `content_data_id` | ULID of the content node |
| `version_number` | Sequential version number |
| `locale` | Locale code for this snapshot |
| `snapshot` | Serialized field values (JSON string) |
| `trigger` | What created this version (e.g., publish, manual) |
| `label` | Optional human-readable label |
| `published` | Whether the content was published at snapshot time |
| `published_by` | ULID of the user who triggered the snapshot |
| `date_created` | Timestamp of snapshot creation |

### Getting a Single Version

```bash
curl "http://localhost:8080/api/v1/content/versions/?q=01JNRWCP5GNSY8Q6P0Z3B7L9EN" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Creating a Manual Version

Create a checkpoint snapshot without changing publish status:

```bash
curl -X POST http://localhost:8080/api/v1/content/versions \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "label": "Before major rewrite"
  }'
```

### Deleting a Version

Remove a historical snapshot. This does not affect the current content:

```bash
curl -X DELETE "http://localhost:8080/api/v1/content/versions/?q=01JNRWCP5GNSY8Q6P0Z3B7L9EN" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

## Restoring a Version

Replace the current draft field values with those from a previous snapshot. A new version is created before the restore to preserve the current state.

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

`fields_restored` indicates how many fields were updated. `unmapped_fields` lists field names present in the snapshot but no longer defined on the datatype schema -- these are skipped during restore.

After restoring, the content remains in its current publish status. Call publish separately to make the restored content live.

## Admin Content Publishing

Admin content (content managed through the internal admin route system) uses a separate set of endpoints with identical behavior. Replace `/api/v1/content/` with `/api/v1/admin/content/` and use admin-prefixed request/response fields:

| Operation | Public Endpoint | Admin Endpoint |
|-----------|----------------|----------------|
| Publish | `POST /api/v1/content/publish` | `POST /api/v1/admin/content/publish` |
| Unpublish | `POST /api/v1/content/unpublish` | `POST /api/v1/admin/content/unpublish` |
| Schedule | `POST /api/v1/content/schedule` | `POST /api/v1/admin/content/schedule` |
| List versions | `GET /api/v1/content/versions` | `GET /api/v1/admin/content/versions` |
| Get version | `GET /api/v1/content/versions/` | `GET /api/v1/admin/content/versions/` |
| Create version | `POST /api/v1/content/versions` | `POST /api/v1/admin/content/versions` |
| Delete version | `DELETE /api/v1/content/versions/` | `DELETE /api/v1/admin/content/versions/` |
| Restore | `POST /api/v1/content/restore` | `POST /api/v1/admin/content/restore` |

Admin requests use `admin_content_data_id` and `admin_content_version_id` instead of `content_data_id` and `content_version_id`.

## SDK Examples

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

// Admin content uses the same methods via client.AdminPublishing
adminResp, err := client.AdminPublishing.AdminPublish(ctx, modula.AdminPublishRequest{
    AdminContentDataID: "01JNRXYZ...",
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

## Notes

- **Immutable snapshots.** Version snapshots cannot be edited after creation. They are permanent records of field values at a point in time.
- **Automatic backup on restore.** Before overwriting current field values during a restore, the system creates a new version snapshot of the current state. You can always undo a restore by restoring to the auto-created version.
- **Schema drift.** If the datatype schema has changed since a version was created (fields added, removed, or renamed), `unmapped_fields` in the restore response lists fields that could not be mapped. New fields not present in the snapshot are left unchanged.
- **Scheduling precision.** The server checks for scheduled content on a periodic interval. There may be a short delay between the `publish_at` time and actual publication.
