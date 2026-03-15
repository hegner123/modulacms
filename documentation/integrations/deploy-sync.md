# Deploy Sync

Export content from one ModulaCMS instance and import it into another to promote content across environments.

## Concepts

**Deploy** -- The process of exporting content from a source CMS instance and importing it into a target instance. The export produces a self-contained JSON payload that any reachable ModulaCMS instance can consume.

**Sync payload** -- A JSON document containing the exported tables and their rows. The payload is produced by the export endpoint and consumed by the import endpoint. You do not need to parse or modify it.

**Dry run** -- A preview import that reports what would change without writing anything to the database.

**Snapshot ID** -- Each import operation is tagged with a unique identifier for auditing. The snapshot ID appears in the sync result.

## The Deploy Workflow

A typical deploy follows four steps:

1. **Health check** -- Verify the target instance is reachable and compatible.
2. **Export** -- Extract content from the source instance.
3. **Dry run** -- Preview the import on the target to check for conflicts.
4. **Import** -- Apply the payload to the target instance.

## Check Target Health

Verify that the target CMS instance is reachable and report its version:

```bash
curl http://target-cms:8080/api/v1/deploy/health \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

```json
{
  "status": "ok",
  "version": "0.42.0",
  "node_id": "01JMKW8N3QRYZ7T1B5K6F2P4HD"
}
```

| Field | Description |
|-------|-------------|
| `status` | Deploy subsystem state (`ok` or `degraded`) |
| `version` | ModulaCMS server version |
| `node_id` | Unique identifier of the CMS instance |

> **Good to know**: Keep source and target instances on the same major version. Schema differences between versions may cause import errors.

## Export Content

### Export All Tables

```bash
curl -X POST http://source-cms:8080/api/v1/deploy/export \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{}' \
  -o payload.json
```

### Export Specific Tables

Provide a `tables` array to export only the tables you need:

```bash
curl -X POST http://source-cms:8080/api/v1/deploy/export \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"tables": ["datatypes", "fields", "content_data", "content_fields", "routes"]}' \
  -o payload.json
```

### Include Plugin Tables

Set `include_plugins` to include data from plugin-created tables:

```bash
curl -X POST http://source-cms:8080/api/v1/deploy/export \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"include_plugins": true}' \
  -o payload.json
```

You can combine `tables` and `include_plugins` in the same request.

> **Good to know**: When exporting specific tables, include dependency tables. For example, `content_data` requires `datatypes` and `routes` to satisfy constraints on the target.

## Preview an Import (Dry Run)

Preview what the import would change without writing to the database:

```bash
curl -X POST "http://target-cms:8080/api/v1/deploy/import?dry_run=true" \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d @payload.json
```

```json
{
  "success": true,
  "dry_run": true,
  "strategy": "overwrite",
  "tables_affected": ["datatypes", "fields", "content_data", "content_fields", "routes"],
  "row_counts": {
    "datatypes": 5,
    "fields": 22,
    "content_data": 48,
    "content_fields": 192,
    "routes": 8
  },
  "backup_path": "",
  "snapshot_id": "",
  "duration": "1.2s",
  "errors": [],
  "warnings": []
}
```

Review `tables_affected`, `row_counts`, and any `warnings` before proceeding with the actual import.

## Import Content

Apply the sync payload to the target instance:

```bash
curl -X POST http://target-cms:8080/api/v1/deploy/import \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d @payload.json
```

```json
{
  "success": true,
  "dry_run": false,
  "strategy": "overwrite",
  "tables_affected": ["datatypes", "fields", "content_data", "content_fields", "routes"],
  "row_counts": {
    "datatypes": 5,
    "fields": 22,
    "content_data": 48,
    "content_fields": 192,
    "routes": 8
  },
  "backup_path": "/var/modulacms/backups/pre-import-20260307.zip",
  "snapshot_id": "01JNRWHSA1LQWZ3X5D8F2G9JKT",
  "duration": "3.8s",
  "errors": [],
  "warnings": []
}
```

> **Good to know**: Before writing, ModulaCMS creates a backup of the affected tables. The backup path is included in the sync result for manual recovery if needed.

## Sync Result Fields

| Field | Description |
|-------|-------------|
| `success` | Whether the import completed without errors |
| `dry_run` | Whether this was a preview (`true`) or actual write (`false`) |
| `strategy` | Merge strategy used (`overwrite`) |
| `tables_affected` | List of tables that were modified |
| `row_counts` | Number of rows written per table |
| `backup_path` | Path to the pre-import backup file (empty for dry runs) |
| `snapshot_id` | Unique ID for this sync operation (empty for dry runs) |
| `duration` | Elapsed time for the operation |
| `errors` | Per-table or per-row failures |
| `warnings` | Non-fatal issues encountered |

## Handle Import Errors

If errors occur during import, the `errors` array contains details:

```json
{
  "errors": [
    {
      "table": "content_data",
      "phase": "import",
      "message": "foreign key constraint failed: route_id references missing route",
      "row_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM"
    }
  ]
}
```

| Field | Description |
|-------|-------------|
| `table` | Table where the error occurred |
| `phase` | Stage of sync that failed (`validate`, `import`, `verify`) |
| `message` | Description of the error |
| `row_id` | ID of the specific row that failed (when applicable) |

## Import Strategy

The import uses an **overwrite** strategy. It truncates each affected table and re-inserts all rows from the payload. This is a full replacement, not a merge.

## Plugin Table Behavior

When `include_plugins` is set in the export:

- Registered plugin tables (prefixed `plugin_`) are included in the payload.
- On import, plugin tables that do not exist on the destination (plugin not installed) are skipped with a warning.
- Plugin tables with a schema mismatch (different columns) are also skipped with a warning.

## SDK Examples

### Go

```go
import modula "github.com/hegner123/modulacms/sdks/go"

source, _ := modula.NewClient(modula.ClientConfig{
    BaseURL: "http://source-cms:8080",
    APIKey:  "mcms_SOURCE_KEY",
})

target, _ := modula.NewClient(modula.ClientConfig{
    BaseURL: "http://target-cms:8080",
    APIKey:  "mcms_TARGET_KEY",
})

// 1. Health check
health, err := target.Deploy.Health(ctx)

// 2. Export from source
payload, err := source.Deploy.Export(ctx, nil) // nil exports all tables

// 3. Dry run on target
preview, err := target.Deploy.DryRunImport(ctx, payload)
if !preview.Success {
    // Handle errors
}

// 4. Import to target
result, err := target.Deploy.Import(ctx, payload)
```

### TypeScript

```typescript
import { ModulaCMSAdmin } from '@modulacms/admin-sdk'

const source = new ModulaCMSAdmin({
  baseUrl: 'http://source-cms:8080',
  apiKey: 'mcms_SOURCE_KEY',
})

const target = new ModulaCMSAdmin({
  baseUrl: 'http://target-cms:8080',
  apiKey: 'mcms_TARGET_KEY',
})

// 1. Health check
const health = await target.deploy.health()

// 2. Export from source
const payload = await source.deploy.export()

// 3. Dry run on target
const preview = await target.deploy.dryRunImport(payload)
if (!preview.success) {
  console.error('Dry run failed:', preview.errors)
}

// 4. Import to target
const result = await target.deploy.import(payload)
```

## API Reference

All deploy endpoints require authentication and `deploy:*` permissions (admin-only by default).

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/deploy/health` | `deploy:read` | Check deploy subsystem health |
| POST | `/api/v1/deploy/export` | `deploy:read` | Export content as sync payload |
| POST | `/api/v1/deploy/import` | `deploy:create` | Import sync payload |
| POST | `/api/v1/deploy/import?dry_run=true` | `deploy:create` | Preview import without writing |

## Next Steps

- [Webhooks](webhooks.md) -- trigger external actions when content changes
- [S3 storage](s3-storage.md) -- configure media storage (media files are not included in deploy sync)
- [Configuration reference](../getting-started/configuration.md) -- all config fields
