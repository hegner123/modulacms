# Deploy Sync

Export content from one ModulaCMS instance and import it into another to promote content across environments.

## Concepts

**Deploy** -- The process of exporting content from a source CMS instance and importing it into a target instance. The export produces a self-contained JSON payload that any reachable ModulaCMS instance can consume.

**Sync payload** -- A JSON document containing the exported tables and their rows. The payload is produced by the export endpoint and consumed by the import endpoint. You do not need to parse or modify it.

**Dry run** -- A preview import that reports what would change without writing anything to the database.

**Snapshot ID** -- Each import operation is tagged with a unique identifier for auditing. The snapshot ID appears in the sync result.

## Configuration

Deploy sync requires configuration on both the source and target instances.

### Minimum Config (API-Only Sync)

If you only use the export/import API endpoints directly (curl, SDK), no special config is needed beyond authentication. The export and import endpoints work with any authenticated user that has `deploy:read` and `deploy:create` permissions (admin role by default).

### Environment Config (Push/Pull via Admin Panel or CLI)

To use push/pull from the admin panel or CLI, configure named environments in `modula.config.json`:

```json
{
  "deploy_environments": [
    {
      "name": "production",
      "url": "https://production-cms.example.com",
      "api_key": "mcms_PRODUCTION_API_KEY"
    },
    {
      "name": "staging",
      "url": "https://staging-cms.example.com",
      "api_key": "mcms_STAGING_API_KEY"
    }
  ],
  "deploy_snapshot_dir": "./deploy/snapshots"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `deploy_environments[].name` | Yes | Environment label used in push/pull commands |
| `deploy_environments[].url` | Yes | Base URL of the target CMS instance (with scheme, without trailing slash) |
| `deploy_environments[].api_key` | Yes | API token for authenticating to the target. Generate one via the admin panel or `POST /api/v1/tokens` |
| `deploy_snapshot_dir` | No | Directory for pre-import snapshots. Defaults to `./deploy/snapshots` |

### Target Instance Requirements

The target CMS instance must:

1. **Be running and reachable** from the source at the configured URL
2. **Have the same major version** as the source (schema differences cause import errors)
3. **Have a valid API token** with `deploy:read` and `deploy:create` permissions
4. **Have the same database tables** -- the import truncates and replaces table contents

### Permissions

All deploy endpoints require authentication. The default permission mapping:

| Endpoint | Permission | Default Roles |
|----------|------------|---------------|
| `GET /api/v1/deploy/health` | `deploy:read` | admin |
| `POST /api/v1/deploy/export` | `deploy:read` | admin |
| `POST /api/v1/deploy/import` | `deploy:create` | admin |

### Snapshot Directory

Before each import, ModulaCMS saves a snapshot of the affected tables. Configure the directory:

```json
{
  "deploy_snapshot_dir": "/var/modulacms/snapshots"
}
```

If unset, snapshots are saved to `./deploy/snapshots` relative to the working directory. Ensure the directory exists and the CMS process has write permission.

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

## Troubleshooting

### unknown deploy environment "production"

The environment name in your push/pull command doesn't match any entry in `deploy_environments`. Check your config:

```bash
# View configured environments
cat modula.config.json | grep -A 5 deploy_environments
```

Environment names are case-sensitive. The name must match exactly.

### environment "production" has no URL configured

The environment entry exists but `url` is empty. Add the target URL:

```json
{
  "deploy_environments": [
    {
      "name": "production",
      "url": "https://production-cms.example.com",
      "api_key": "mcms_YOUR_KEY"
    }
  ]
}
```

### environment "production" has no api_key configured

Same as above, but the `api_key` field is empty. Generate a token on the target instance and add it to the config.

### remote health check failed

The source cannot reach the target CMS. Verify:

1. The URL is correct and includes the scheme (`https://` or `http://`)
2. The target CMS is running and its HTTP port is accessible from the source
3. No firewall or network policy blocks the connection
4. The API key is valid -- test manually:

```bash
curl -H "Authorization: Bearer mcms_YOUR_KEY" \
  https://target-cms.example.com/api/v1/deploy/health
```

If the health check returns `401 Unauthorized`, the API key is invalid or expired.

### import already in progress (409 Conflict)

Another import is running on the target. Only one import can run at a time. Wait for it to finish, or restart the target if it's stuck. Imports time out after 5 minutes.

### foreign key constraint failed during import

The payload references records that don't exist on the target. Common causes:

- **Missing dependency tables:** You exported `content_data` without including `datatypes` or `routes`. Re-export with all dependency tables, or export all tables (empty `tables` array).
- **Version mismatch:** Source and target are on different schema versions. Upgrade both to the same version.
- **Partial import from a previous failure:** Run the import again -- it truncates tables before inserting, so stale data from a failed import is cleared.

### payload hash mismatch

The payload was modified after export (corrupted in transit or manually edited). Re-export from the source and import the fresh payload.

### schema version mismatch

The source and target databases have different table schemas (different columns). This happens when instances are on different CMS versions. Upgrade both to the same version and re-export.

### Would create N placeholder user(s) (warning)

The payload references user IDs that don't exist on the target. During import, ModulaCMS creates placeholder user records so that foreign key constraints are satisfied. This is a warning, not an error -- the import proceeds. After import, you can update placeholder users with real credentials.

### Plugin table skipped (warning)

A plugin table in the payload doesn't exist on the target (the plugin isn't installed there). The table is skipped with a warning. Install the plugin on the target first if you need its data.

### import failed: snapshot directory not writable

The configured `deploy_snapshot_dir` doesn't exist or the CMS process can't write to it:

```bash
mkdir -p ./deploy/snapshots
chmod 755 ./deploy/snapshots
```

### Request body too large (import)

The import payload exceeds the 100 MB limit. For very large datasets:

1. Export specific tables instead of all tables: `{"tables": ["content_data", "content_fields"]}`
2. Split the import into multiple smaller payloads by table group
3. Use gzip compression -- the import endpoint accepts `Content-Encoding: gzip`

### Recovering from a bad import

Every import creates a pre-import snapshot. The `backup_path` in the sync result tells you where it's saved. To restore:

1. Find the snapshot ID from the import result or list snapshots in the snapshot directory
2. Use the CLI or API to restore from the snapshot

If the snapshot directory is empty or the backup was skipped, you'll need to re-import from a known-good source.

## Next Steps

- [Webhooks](webhooks.md) -- trigger external actions when content changes
- [S3 storage](s3-storage.md) -- configure media storage (media files are not included in deploy sync)
- [Configuration reference](../getting-started/configuration.md) -- all config fields
