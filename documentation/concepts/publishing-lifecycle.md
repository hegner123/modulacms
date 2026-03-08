# Publishing Lifecycle

ModulaCMS content goes through a lifecycle from creation to publication. Content starts as a draft, can be published immediately or scheduled for a future time, and can be reverted to any previous version. Every publish operation creates an immutable version snapshot that preserves the content's field values at that point in time.

## Content Status

The `ContentStatus` type has two values:

| Status | Description |
|--------|-------------|
| `draft` | Content is being edited. Not visible through the public content delivery API. |
| `published` | Content is live and served by the public content delivery API. |

Content is created with `status: "draft"` by default. The `publish_at` timestamp field on content data supports scheduling.

## Publishing

Publishing transitions content from draft to published. The operation:

1. Sets the content node's status to `published`.
2. Records `published_at` (current timestamp) and `published_by` (the user performing the action).
3. Creates an immutable version snapshot of the current field values.
4. Increments the content's `revision` counter.

```go
resp, err := client.Publishing.Publish(ctx, modula.PublishRequest{
    ContentDataID: contentID,
    Locale:        "en",  // optional; default locale if omitted
})
// resp.Status == "published"
// resp.VersionNumber == 1
// resp.ContentVersionID == "01ABC..."
```

Published content is immediately available through the public delivery endpoint (`GET /api/v1/content/{slug}`).

## Unpublishing

Unpublishing reverts content back to draft status, removing it from public delivery. Existing version snapshots are preserved.

```go
resp, err := client.Publishing.Unpublish(ctx, modula.PublishRequest{
    ContentDataID: contentID,
})
// resp.Status == "unpublished"
```

The content's `published_at` and `published_by` fields are cleared. The content is no longer served by the delivery API until republished.

## Scheduling

Scheduling sets a future publication time. The content remains in draft status until the scheduled time, when the server automatically publishes it.

```go
resp, err := client.Publishing.Schedule(ctx, modula.ScheduleRequest{
    ContentDataID: contentID,
    PublishAt:      "2026-04-01T09:00:00Z",  // ISO 8601 timestamp
})
// resp.Status == "scheduled"
// resp.PublishAt == "2026-04-01T09:00:00Z"
```

The `publish_at` field on the content data node stores the scheduled time. The server checks for scheduled content and publishes it when the time arrives. Scheduling creates a version snapshot at the time of the schedule call.

To cancel a scheduled publish, update the content data node and clear the `publish_at` field.

## Version Snapshots

A version snapshot is an immutable record of a content node's field values at a specific point in time.

```go
type ContentVersion struct {
    ContentVersionID ContentVersionID `json:"content_version_id"`
    ContentDataID    ContentID        `json:"content_data_id"`
    VersionNumber    int64            `json:"version_number"`
    Locale           string           `json:"locale"`
    Snapshot         string           `json:"snapshot"`
    Trigger          string           `json:"trigger"`
    Label            string           `json:"label"`
    Published        bool             `json:"published"`
    PublishedBy      *UserID          `json:"published_by,omitempty"`
    DateCreated      Timestamp        `json:"date_created"`
}
```

| Field | Purpose |
|-------|---------|
| `VersionNumber` | Incrementing version counter per content node |
| `Locale` | The locale this snapshot applies to |
| `Snapshot` | JSON string containing all field values at the time of creation |
| `Trigger` | What created the version: `"publish"`, `"manual"`, `"restore"`, `"schedule"` |
| `Label` | Optional human-readable name (e.g., `"Pre-launch draft"`) |
| `Published` | True if this version is the currently live version for the given locale |

### Automatic Versioning

Versions are created automatically when you:

- **Publish** content -- trigger is `"publish"`
- **Schedule** content -- trigger is `"schedule"`
- **Restore** content -- a snapshot of the current state is saved before overwriting (trigger is `"restore"`)

### Manual Versioning

Create a version snapshot without changing publish status:

```go
version, err := client.Publishing.CreateVersion(ctx, modula.CreateVersionRequest{
    ContentDataID: contentID,
    Label:         "Before redesign",  // optional
})
```

This saves a checkpoint of the current field values that you can restore later.

### Listing Versions

```go
versions, err := client.Publishing.ListVersions(ctx, contentID.String())
// versions is ordered newest-first
```

### Deleting Versions

```go
err := client.Publishing.DeleteVersion(ctx, versionID)
```

Deleting a version removes the historical snapshot. It does not affect the current content.

## Restoring a Version

Restore replaces the current draft field values with those from a previous version snapshot. The content's publish status is not changed.

```go
resp, err := client.Publishing.Restore(ctx, modula.RestoreRequest{
    ContentDataID:    contentID,
    ContentVersionID: versionID,
})
// resp.FieldsRestored == 5
// resp.UnmappedFields == ["old_field_name"]  // fields in snapshot but not in current schema
```

The restore operation:

1. Creates a version snapshot of the current state (so you can undo the restore).
2. Reads the field values from the specified version's `Snapshot` JSON.
3. Overwrites the current content field values with the snapshot values.
4. Returns the count of restored fields and any unmapped fields.

**Unmapped fields** are field names present in the version snapshot but absent from the current datatype schema (because the schema changed after the snapshot was created). These values are skipped during restore.

After restoring, the content is still in its current publish status. Call `Publish` separately to make the restored content live.

## Admin Content Publishing

The admin content system has its own parallel publishing lifecycle with identical operations:

| User Content | Admin Content |
|-------------|---------------|
| `PublishRequest` | `AdminPublishRequest` |
| `PublishResponse` | `AdminPublishResponse` |
| `ScheduleRequest` | `AdminScheduleRequest` |
| `ContentVersion` | `AdminContentVersion` |
| `RestoreRequest` | `AdminRestoreRequest` |

Admin publishing is accessed via `client.AdminPublishing` and operates on the admin content tables.

## Lifecycle Summary

```
                    +-- Schedule ---> [publish_at set] ---> auto-Publish
                    |
Create (draft) ---> Publish ---> [published]
                    ^                  |
                    |                  +-- Unpublish ---> [draft]
                    |
                    +-- Restore (from version) ---> [draft, fields overwritten]
```

Every state transition that changes content visibility creates a version snapshot. The version history is append-only and survives unpublish, restore, and republish cycles.
