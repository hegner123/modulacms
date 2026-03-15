# Deploy Sync

ModulaCMS can synchronize content between CMS instances. The deploy system exports content from a source environment as a JSON payload and imports it into a target environment. This enables workflows like dev-to-staging-to-production content promotion without manual re-entry.

## Concepts

**Deploy** -- The process of exporting content from one CMS instance and importing it into another. The export produces a self-contained JSON payload that can be transmitted to any reachable ModulaCMS instance.

**Sync payload** -- An opaque JSON document containing the exported tables and their rows. The payload is produced by the export endpoint and consumed by the import endpoint. You do not need to parse or modify it.

**Dry run** -- A preview import that reports what would change without writing anything to the database. Use this to verify the impact of an import before committing.

**Snapshot ID** -- Each import operation is tagged with a unique snapshot ID for auditing and potential rollback. The snapshot ID appears in the sync result.

## Workflow

A typical deploy follows these steps:

1. **Health check** -- Verify the target instance is reachable and compatible.
2. **Export** -- Extract content from the source instance.
3. **Dry run** -- Preview the import on the target to check for conflicts.
4. **Import** -- Apply the payload to the target instance.

## Health Check

Verify that the target CMS instance is reachable and report its version and node ID:

```bash
curl http://target-cms:8080/api/v1/deploy/health \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Response:

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
| `node_id` | Unique identifier of this CMS instance |

## Exporting Content

Export all tables from the source instance:

```bash
curl -X POST http://source-cms:8080/api/v1/deploy/export \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{}' \
  -o payload.json
```

To export only specific tables, provide a `tables` array:

```bash
curl -X POST http://source-cms:8080/api/v1/deploy/export \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"tables": ["datatypes", "fields", "content_data", "content_fields", "routes"]}' \
  -o payload.json
```

To include plugin table data in the export, set `include_plugins` to `true`:

```bash
curl -X POST http://source-cms:8080/api/v1/deploy/export \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"include_plugins": true}' \
  -o payload.json
```

Plugin tables are tables created by Lua plugins at runtime (named `plugin_<name>_<table>`). System plugin tables (`plugin_routes`, `plugin_hooks`, `plugin_requests`) are excluded automatically. You can combine `tables` and `include_plugins` in the same request.

The response is the raw sync payload JSON. Save it to a file for import.

## Dry Run Import

Preview what an import would change without writing to the database:

```bash
curl -X POST "http://target-cms:8080/api/v1/deploy/import?dry_run=true" \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d @payload.json
```

Response:

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

The `dry_run` field is `true`, confirming no data was written. Review `tables_affected`, `row_counts`, and any `warnings` before proceeding with the actual import.

## Importing Content

Apply the sync payload to the target instance:

```bash
curl -X POST http://target-cms:8080/api/v1/deploy/import \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d @payload.json
```

Response:

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

### Sync Result Fields

| Field | Description |
|-------|-------------|
| `success` | Whether the import completed without errors |
| `dry_run` | Whether this was a preview (true) or actual write (false) |
| `strategy` | Merge strategy used (`overwrite`) |
| `tables_affected` | List of database tables that were modified |
| `row_counts` | Number of rows written per table |
| `backup_path` | Path to the pre-import backup file (empty for dry runs) |
| `snapshot_id` | Unique ID for this sync operation (empty for dry runs) |
| `duration` | Elapsed time for the operation |
| `errors` | Per-table or per-row failures (see below) |
| `warnings` | Non-fatal issues encountered |

### Handling Errors

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
| `table` | Database table where the error occurred |
| `phase` | Stage of sync that failed (`validate`, `import`, `verify`) |
| `message` | Description of the error |
| `row_id` | ULID of the specific row that failed (when applicable) |

## SDK Examples

### Go

```go
import (
    "encoding/json"
    modula "github.com/hegner123/modulacms/sdks/go"
)

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

// Export specific tables only
payload, err = source.Deploy.Export(ctx, []string{"datatypes", "fields", "content_data"})

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

All deploy endpoints require authentication and `deploy:*` permissions.

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/deploy/health` | `deploy:read` | Check deploy subsystem health |
| POST | `/api/v1/deploy/export` | `deploy:read` | Export content as sync payload |
| POST | `/api/v1/deploy/import` | `deploy:create` | Import sync payload |
| POST | `/api/v1/deploy/import?dry_run=true` | `deploy:create` | Preview import without writing |

## Notes

- **Automatic backup.** Before writing, the import endpoint creates a backup of the affected tables. The backup path is included in the sync result for manual recovery if needed.
- **Overwrite strategy.** Import truncates each table and re-inserts all rows from the payload. This is a full replacement, not a merge.
- **Table dependencies.** When exporting specific tables, include dependency tables. For example, `content_data` requires `datatypes` and `routes` to satisfy foreign key constraints on the target. The `media_folders` table has a `parent_id` self-reference, so folder rows must be imported in dependency order (parents before children); the sync system handles this automatically.
- **Plugin tables.** When `include_plugins` is set, registered plugin tables (prefixed `plugin_`) are included in the export. On import, plugin tables that do not exist on the destination (plugin not installed) are skipped with a warning. Plugin tables with a schema mismatch (different columns) are also skipped with a warning.
- **Cross-version compatibility.** The health check reports the server version. Schema differences between versions may cause import errors. Keep source and target instances on the same major version.
- **Permissions.** Deploy operations require the `deploy:read` and `deploy:create` permissions, which are only assigned to the admin role by default.
