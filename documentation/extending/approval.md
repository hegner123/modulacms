# Route and Hook Approval

Approve plugin routes and hooks before they can execute, and manage approvals through CLI, API, or TUI.

## Why Approval Exists

Plugins run Lua code that registers HTTP endpoints and injects logic into content lifecycle events. The approval gate ensures an admin reviews what a plugin does before it affects the system.

- **Unapproved routes** return 404 as if they don't exist.
- **Unapproved hooks** are silently skipped during content operations.

There is no partial execution. A route is either fully approved and serving traffic, or completely invisible.

## Approval States

Each route and hook has one of three states:

| State | Description |
|-------|-------------|
| Unapproved | Default for new routes/hooks. No code executes. |
| Approved | Admin has explicitly approved. Route serves traffic / hook fires on events. |
| Revoked | Previously approved, then explicitly revoked. Same behavior as unapproved. |

There is no functional difference between "unapproved" and "revoked" -- both prevent execution. The distinction exists for audit trail clarity.

## Version-Change Behavior

When a plugin's `version` field in `plugin_info` changes, **all** route and hook approvals for that plugin are automatically revoked. This happens during plugin loading, before any code executes.

This prevents a scenario where:
1. Admin approves plugin v1.0.0 routes after reviewing the code.
2. Developer updates the plugin to v2.0.0 with different behavior.
3. The updated routes serve traffic without review.

After a version bump, the admin must re-approve all routes and hooks. Use `--all-routes` and `--all-hooks` flags for bulk re-approval.

Version-change revocation does not fire if only the Lua code changes without a version bump. To enforce re-approval on code changes without version bumps, use the revoke commands manually.

## Approving via CLI

> **Good to know**: The CLI uses a Bearer token auto-generated at server startup (written to `<config_dir>/.plugin-api-token`). Pass `--token <value>` for CI/CD environments.

### Approve Routes

```bash
# Approve all routes for a plugin
modulacms plugin approve my_plugin --all-routes

# Approve a specific route
modulacms plugin approve my_plugin --route "GET /tasks"
modulacms plugin approve my_plugin --route "POST /tasks"
modulacms plugin approve my_plugin --route "GET /tasks/{id}"

# Skip confirmation prompt (for CI/CD)
modulacms plugin approve my_plugin --all-routes --yes
```

### Approve Hooks

```bash
# Approve all hooks for a plugin
modulacms plugin approve my_plugin --all-hooks

# Approve a specific hook (format: "event:table")
modulacms plugin approve my_plugin --hook "before_create:content_data"
modulacms plugin approve my_plugin --hook "after_update:content_data"

# Wildcard hooks use "*" as the table
modulacms plugin approve my_plugin --hook "after_delete:*"

# Skip confirmation prompt
modulacms plugin approve my_plugin --all-hooks --yes
```

### Revoke Routes

```bash
# Revoke all routes
modulacms plugin revoke my_plugin --all-routes

# Revoke a specific route
modulacms plugin revoke my_plugin --route "GET /tasks"
```

### Revoke Hooks

```bash
# Revoke all hooks
modulacms plugin revoke my_plugin --all-hooks

# Revoke a specific hook
modulacms plugin revoke my_plugin --hook "before_create:content_data"
```

### Combined Approval

Approve both routes and hooks in one command:

```bash
modulacms plugin approve my_plugin --all-routes --all-hooks --yes
```

## Approving via API

All approval endpoints require authentication with `plugins:admin` permission. Read endpoints require `plugins:read`.

### List Routes

```bash
curl http://localhost:8080/api/v1/admin/plugins/routes \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

The response includes each route's plugin name, method, path, public flag, and approval status.

### Approve Routes

```bash
curl -X POST http://localhost:8080/api/v1/admin/plugins/routes/approve \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "routes": [
      {"plugin": "task_tracker", "method": "GET", "path": "/tasks"},
      {"plugin": "task_tracker", "method": "POST", "path": "/tasks"},
      {"plugin": "task_tracker", "method": "GET", "path": "/tasks/{id}"},
      {"plugin": "task_tracker", "method": "PUT", "path": "/tasks/{id}"},
      {"plugin": "task_tracker", "method": "DELETE", "path": "/tasks/{id}"}
    ]
  }'
```

### Revoke Routes

```bash
curl -X POST http://localhost:8080/api/v1/admin/plugins/routes/revoke \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "routes": [
      {"plugin": "task_tracker", "method": "GET", "path": "/tasks"}
    ]
  }'
```

### List Hooks

```bash
curl http://localhost:8080/api/v1/admin/plugins/hooks \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

The response includes each hook's plugin name, event, table, priority, wildcard flag, and approval status.

### Approve Hooks

```bash
curl -X POST http://localhost:8080/api/v1/admin/plugins/hooks/approve \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "hooks": [
      {"plugin": "task_tracker", "event": "before_create", "table": "content_data"},
      {"plugin": "task_tracker", "event": "after_update", "table": "content_data"}
    ]
  }'
```

### Revoke Hooks

```bash
curl -X POST http://localhost:8080/api/v1/admin/plugins/hooks/revoke \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "hooks": [
      {"plugin": "task_tracker", "event": "before_create", "table": "content_data"}
    ]
  }'
```

## Approving via TUI

The SSH TUI includes a Plugins page accessible from the homepage menu:

1. Navigate to the **Plugins** screen from the home menu.
2. Select a plugin from the list.
3. The plugin detail view shows all registered routes and hooks with their approval status.
4. Use the approve action to approve individual routes or hooks through a confirmation dialog.

## CI/CD Patterns

For automated deployments, script the approval after deploying the plugin:

```bash
#!/usr/bin/env bash

PLUGIN_NAME="task_tracker"
TOKEN=$(cat /opt/modulacms/.plugin-api-token)

# Wait for plugin to load
sleep 2

# Approve all routes and hooks
modulacms plugin approve "$PLUGIN_NAME" \
  --all-routes --all-hooks \
  --yes \
  --token "$TOKEN"

# Verify
modulacms plugin info "$PLUGIN_NAME" --token "$TOKEN"
```

For API-based CI/CD without the CLI:

```bash
#!/usr/bin/env bash

BASE="http://localhost:8080/api/v1/admin/plugins"
SESSION="YOUR_CI_SESSION_COOKIE"

# Approve routes
curl -X POST "$BASE/routes/approve" \
  -H "Cookie: session=$SESSION" \
  -H "Content-Type: application/json" \
  -d '{
    "routes": [
      {"plugin": "task_tracker", "method": "GET", "path": "/tasks"},
      {"plugin": "task_tracker", "method": "POST", "path": "/tasks"}
    ]
  }'

# Approve hooks
curl -X POST "$BASE/hooks/approve" \
  -H "Cookie: session=$SESSION" \
  -H "Content-Type: application/json" \
  -d '{
    "hooks": [
      {"plugin": "task_tracker", "event": "after_create", "table": "content_data"}
    ]
  }'
```

## Idempotent Operations

All approval and revocation operations are idempotent. Approving an already-approved route or revoking a never-approved hook produces no error. Redundant operations are no-ops, making CI/CD scripts safe to re-run without conditional logic.
