# TUI QA Phase 5: Extensions

Covers: Plugins (list + detail + actions), Deploy (test/pull/push), Pipelines (read-only inspection), Webhooks (read-only — known gap).

## Prerequisites

- DB Wipe & Redeploy + Modula Default schema installed
- Plugin tests require at least one Lua plugin installed (may need manual setup)
- Deploy tests require a deploy environment configured in modula.config.json

## 5.1 Plugins Screen

### 5.1.1 Plugin list renders
- Navigate to Plugins, wait for stable
- Verify either:
  - "No plugins installed" message, OR
  - Plugin list with state indicators (running/stopped/failed)
- Verify Info panel shows counts: total, running, failed, stopped
- Snapshot: `goldens/5.1.1_plugins.txt`

### 5.1.2 Plugin with drift detection (if plugins exist)
- If a plugin has manifest drift: verify drift warning indicator
- If a plugin has circuit breaker tripped: verify circuit breaker state shown

### 5.1.3 Navigate to plugin detail
- Select a plugin, press `enter`
- Verify transition to PluginDetailScreen
- Verify info panel: name, version, state, circuit breaker status, description

### 5.1.4 Plugin detail — actions list
- Verify 6 actions visible: Enable, Disable, Reload, Approve Routes, Approve Hooks, Sync Capabilities
- Verify action descriptions update on cursor movement

### 5.1.5 Enable/Disable plugin
- Select Disable, press `enter` (if plugin is enabled)
- Verify plugin state changes to disabled
- Select Enable, press `enter`
- Verify plugin state changes to enabled/running

### 5.1.6 Reload plugin
- Select Reload, press `enter`
- Verify reload result (success or error)

### 5.1.7 Approve routes
- Select "Approve Routes", press `enter`
- If unapproved routes exist: verify approval dialog with route list
- Confirm → verify routes approved
- If no unapproved routes: verify appropriate message

### 5.1.8 Approve hooks
- Select "Approve Hooks", press `enter`
- Same pattern as routes approval

### 5.1.9 Sync capabilities
- Select "Sync Capabilities", press `enter`
- Verify sync result

### 5.1.10 Back to plugins list
- Press `h` → verify return to Plugins list

## 5.2 Deploy Screen

### 5.2.1 Deploy screen renders
- Navigate to Deploy, wait for stable
- Verify environments list or empty state (if none configured)
- If environments exist: verify detail panel shows name, URL, API key (masked), health status
- Snapshot: `goldens/5.2.1_deploy.txt`

### 5.2.2 Test connection
- Select an environment, press `t`
- Wait for result
- Verify connection test result (success with status or error)
- Verify detail panel updates with health status

### 5.2.3 Pull (remote → local)
- Press `p` → verify confirmation dialog
- Confirm → wait for pull result
- Verify result shows: duration, tables, rows, warnings, errors
- Cancel → verify no pull executed

### 5.2.4 Push (local → remote)
- Press `s` → verify confirmation dialog
- Confirm → wait for push result
- Verify result details
- Cancel → verify no push

### 5.2.5 Dry-run pull
- Press `P` (shift+p) → verify dry-run result (no confirmation needed, or lighter confirmation)
- Verify result shows what WOULD be pulled without executing

### 5.2.6 Dry-run push
- Press `S` (shift+s) → verify dry-run push result

### 5.2.7 Concurrent operation guard
- Start a pull operation
- Immediately try another operation (push, test)
- Verify guard prevents concurrent operations (or queues them)

### 5.2.8 No environments configured
- With empty deploy config: navigate to Deploy
- Verify appropriate empty state message

## 5.3 Pipelines Screen

### 5.3.1 Pipeline list renders
- Navigate to Pipelines, wait for stable
- If plugins with pipelines installed: verify chain list with stats
- If no plugins: verify empty state
- Verify Info panel: registry stats (total, entries, before, after), by-table grouping
- Snapshot: `goldens/5.3.1_pipelines.txt`

### 5.3.2 Navigate to pipeline detail
- Select a pipeline chain, press `enter`
- Verify PipelineDetailScreen loads
- Verify: all chains (context panel), entries list, entry detail
- Verify chain status: enabled/disabled counts, execution order

### 5.3.3 Pipeline entry inspection
- In detail screen, move cursor through entries
- Verify detail panel updates for each entry
- Verify execution order display

### 5.3.4 Back to pipeline list
- Press `h` → verify return to pipelines list

## 5.4 Webhooks Screen (Read-Only — Known Gap)

### 5.4.1 Webhook list renders
- Navigate to Webhooks, wait for stable
- If webhooks configured: verify list with active/off status
- If none: verify empty state
- Verify Detail panel shows: name, URL, active status, events, dates
- Verify Info panel shows: total, active counts
- Snapshot: `goldens/5.4.1_webhooks.txt`

### 5.4.2 Cursor navigation
- Move through webhook list
- Verify detail updates for each

### 5.4.3 Verify no CUD operations
- Press `n` → verify no create dialog (no handler)
- Press `e` → verify no edit dialog
- Press `d` → verify no delete dialog
- Verify key hints only show: nav, panel, back, quit (no n/e/d)
- **Document this as a gap for future implementation**

### 5.4.4 Panel focus
- Press `tab` → verify panel focus cycling works
- Press `shift+tab` → verify reverse
