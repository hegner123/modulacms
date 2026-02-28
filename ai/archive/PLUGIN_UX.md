Plugin UX Implementation Plan
==============================

Context
-------

The plugin system is technically complete (19k LOC, 4 phases) but has no user-facing UX. Every other subsystem (db, config, backup, cert) has CLI commands, TUI screens, and justfile recipes. Plugins require raw HTTP API calls for everything. This plan closes 8 gaps.

Design Decisions
----------------

CLI command model: Two categories of commands:
- Offline (list, init, validate): filesystem/Lua VM only, no server needed, fast. If config.json is missing, error and exit.
- Online (info, reload, enable, disable, approve, revoke): send HTTP requests to the running server's existing admin API endpoints. State-changing operations (enable/disable/reload) are handled in-process by the Manager's existing methods (ReloadPlugin, EnablePlugin, DisablePlugin) -- no server restart needed.

Online CLI approach: Rather than spinning up a second independent plugin.Manager (which would have no shared state with the running server), online commands hit the HTTP admin API at localhost. This reuses the existing handlers in internal/router/mux.go and internal/plugin/cli_commands.go. The CLI reads config.json for the server port.

Online CLI authentication: The existing admin API endpoints require authChain(adminOnly(...)). The existing APIKeyAuth() middleware in internal/middleware/middleware.go already supports Bearer token auth via the tokens table. The CLI reuses this infrastructure:
- cmd/serve.go generates a random 32-byte crypto/rand token at startup, hex-encodes it, and inserts it into the tokens table via DbDriver.CreateToken() with type "api_key", tied to the existing system user (already used for CLI mode in internal/cli/model.go:218-224). ExpiresAt uses types.Timestamp{Valid: false} (NULL in DB = no expiry). On startup, stale api_key tokens for the system user are cleaned up before creating a new one (prevents accumulation from crashes).
- The token value is written to <config_dir>/.plugin-api-token (mode 0600) so the CLI can read it. <config_dir> is the directory containing config.json (resolved from the --config flag or default "./config.json" in cmd/root.go).
- pluginAPIClient() in cmd/plugin.go reads the token file and sends it as Authorization: Bearer <token> header. The existing APIKeyAuth middleware validates it against the tokens table -- no new middleware needed.
- For CI/CD: a --token flag on online commands allows passing the token directly (useful when the token file is on a different host).
- On graceful server shutdown, cmd/serve.go deletes the token row from the tokens table (via DbDriver) and removes the token file. On ungraceful shutdown (kill -9, OOM), the token row remains in the DB but is harmless -- the server is dead and can't serve requests. The stale file is overwritten on next startup.
- Audit trail: change_events record the system user (a real DB user), not a synthetic user. No special-case handling needed for audit consumers.

TUI integration: Pass *plugin.Manager from cmd/serve.go through SSH middleware into the TUI model. The Manager exposes Bridge() and HookEngine() accessors (manager.go:196, manager.go:202) which the TUI uses for route/hook approval operations. TUI provides view + management actions (enable/disable/reload/approve), not editing. SSH access is inherently admin-level (SSH auth middleware enforces this), so no additional role check is needed in TUI plugin commands.

Plugin directory auto-creation: os.MkdirAll in initPluginManager before NewManager call. Silent creation, INFO log.

Validation strategy: Export extractManifest() (manager.go:668) as ExtractManifest() and validateManifest() (manager.go:735) as ValidateManifest(). The validator calls ExtractManifest() directly -- no logic duplication. A syntax-only pre-flight check via a minimal lua.NewState() + L.LoadFile(initPath) (parse+compile without executing) is added before ExtractManifest() to catch obvious Lua syntax errors cheaply. gopher-lua has no standalone parse.Parse() function -- L.LoadFile is the idiomatic parse+compile API. Call L.Close() immediately after the check. The existing 2-second context.WithTimeout in ExtractManifest() prevents CPU-bound Lua loops from hanging the CLI. Separate errors (fail) from warnings (informational).

Approve/revoke idempotency: Approving an already-approved route/hook is a no-op (no error, no timestamp update). Revoking an already-revoked route/hook is likewise a no-op.

---
Implementation Steps
--------------------

Step 1: Export manifest functions + validator package

Modify: internal/plugin/manager.go
- Rename extractManifest() (line 668) to ExtractManifest() -- export it
- Rename validateManifest() (line 735) to ValidateManifest() -- export it
- Update all internal callers of both functions

New file: internal/plugin/validator.go
- ValidatePlugin(dir string) (PluginInfo, []ValidationResult, error)
- Check dir exists, init.lua exists
- Pre-flight: lua.NewState() + L.LoadFile(initPath) for syntax errors (parse+compile without executing, then L.Close()). gopher-lua has no standalone parse.Parse() -- L.LoadFile is the idiomatic API for parse+compile without execution.
- Call ExtractManifest() directly -- no logic duplication, no re-implementation of sandbox
- Call ValidateManifest() on the extracted info -- single source of truth for name rules
- Check on_init function presence: Add a HasOnInit bool field to PluginInfo. Inside ExtractManifest(), before L.Close(), check L.GetGlobal("on_init").Type() == lua.LTFunction and set info.HasOnInit accordingly. This avoids returning the VM or changing the function signature -- the check happens inside ExtractManifest before the defer L.Close() runs.
- Return structured results with severity (error vs warning)

New file: internal/plugin/validator_test.go
- Test against existing testdata/plugins/ fixtures
- Test invalid cases: invalid_no_manifest/, invalid_bad_name/
- Test valid case: test_bookmarks/

Reuse: PluginInfo struct from internal/plugin/plugin.go. Name validation MUST be the same exported function, not a copy.

Step 2: CLI parent + offline commands (cmd/plugin.go)

New file: cmd/plugin.go

Parent command:
modulacms plugin - Plugin management commands

Subcommands (offline, no DB). Offline commands operate on the filesystem regardless of the plugin_enabled config value -- they only need plugin_directory from config.json:

plugin list - Scan plugin_directory from config, parse each plugin's manifest, print table:
NAME              VERSION   DESCRIPTION
hello_world       1.0.0     Example greeting plugin
task_tracker      2.1.0     Task management system
Uses loadConfig() to get plugin_directory. If config.json missing, error and exit. Walks directories, calls ValidatePlugin() on each, prints summary. Invalid plugins shown with [invalid] state.

plugin init <name> - Scaffold a new plugin:
- Validate name via validateManifest() name rules (same function as manager.go, not re-implemented)
- Load config for plugin_directory. If config.json missing, error and exit.
- os.MkdirAll for <dir>/<name>/ and <dir>/<name>/lib/
- Write init.lua with manifest skeleton + on_init/on_shutdown
- Interactive prompts via huh.Form for version, description, author, license (with sensible defaults)
- Non-interactive alternative: --version, --description, --author, --license flags. If all required flags provided, skip huh.Form. Also skip huh.Form when !isatty.IsTerminal(os.Stdout.Fd()).

plugin validate <path> - Validate without loading:
- Call ValidatePlugin(path)
- Print errors/warnings
- Exit 0 if no errors, exit 1 if errors

Modify: cmd/root.go - add rootCmd.AddCommand(pluginCmd)

Step 3: Hook admin endpoints + route handler hardening

The HookEngine has ApproveHook/RevokeHook methods but no HTTP endpoints to list hooks or their approval status. The CLI online commands (Step 4) depend on these endpoints. This step also fixes missing body size limits on existing route handlers.

Modify: internal/plugin/hook_engine.go
- Add ListHooks() []HookRegistration method (returns all registered hooks with approval status)
- New type HookRegistration with PluginName, Event, Table, Priority, Approved, IsWildcard

New file: internal/plugin/hook_handlers.go
- PluginHooksListHandler(mgr *Manager) http.Handler -- returns all registered hooks with approval status via mgr.HookEngine().ListHooks()
- PluginHooksApproveHandler(mgr *Manager) http.Handler -- approves hooks. MUST use http.MaxBytesReader(w, r.Body, 1<<20) before JSON decoding. Request body format: {"hooks": [{"plugin": "...", "event": "...", "table": "..."}]}
- PluginHooksRevokeHandler(mgr *Manager) http.Handler -- revokes hooks. Same body limit and format as approve.
- These handlers follow the same pattern as PluginListHandler, PluginInfoHandler, etc. in internal/plugin/cli_commands.go

Modify: internal/plugin/http_bridge.go MountAdminEndpoints()
- Add hook endpoints to MountAdminEndpoints, matching the existing pattern where all plugin admin endpoints are registered:
    mux.Handle("GET /api/v1/admin/plugins/hooks", authChain(PluginHooksListHandler(mgr)))
    mux.Handle("POST /api/v1/admin/plugins/hooks/approve", authChain(adminOnlyFn(PluginHooksApproveHandler(mgr))))
    mux.Handle("POST /api/v1/admin/plugins/hooks/revoke", authChain(adminOnlyFn(PluginHooksRevokeHandler(mgr))))
- Register these as literal-path endpoints in section 1 (before wildcard {name} patterns), alongside the existing /cleanup endpoints
- NOTE: The existing route approve/revoke handlers in mux.go (pluginRoutesApproveHandler at mux.go:287, pluginRoutesRevokeHandler at mux.go:332) are a pre-existing inconsistency -- they live in mux.go rather than the plugin package. The hook handlers go in the plugin package to follow the established MountAdminEndpoints pattern.

Modify: internal/router/mux.go (bug fix only, separate from new hook endpoints)
- Add http.MaxBytesReader(w, r.Body, 1<<20) to the existing pluginRoutesApproveHandler (mux.go:287) and pluginRoutesRevokeHandler (mux.go:332) which currently decode JSON with no body limit

Step 4: CLI online commands (cmd/plugin.go continued)

Online commands send HTTP requests to the running server's admin API. No second Manager is created. The server handles state changes in-process via Manager.ReloadPlugin(), Manager.EnablePlugin(), Manager.DisablePlugin() -- no server restart is needed.

Helper in cmd/plugin.go: pluginAPIClient() - reads config.json for server port, reads admin API token from <config_dir>/.plugin-api-token, returns a configured HTTP client targeting localhost:<port> with Authorization: Bearer <token> header. <config_dir> is resolved from the --config flag or defaults to the directory containing "./config.json". If the token file does not exist or is unreadable, print "server not running or token file missing" and exit 1. All online commands also accept a --token flag that overrides the token file (for CI/CD or remote use).

plugin info <name> - Detailed plugin info:
- GET /api/v1/admin/plugins/{name}
- Print: name, version, description, author, license, state, failed reason, CB state, CB errors, VM available/total, dependencies, schema drift
- If server returns 404, print "plugin not found: <name>" and exit 1

plugin reload <name> - Hot reload:
- POST /api/v1/admin/plugins/{name}/reload
- Server-side reload also restarts file-polling watcher for that plugin (prevents double-reload)
- Print success/failure

plugin enable <name> - Re-enable disabled plugin:
- POST /api/v1/admin/plugins/{name}/enable
- Handled in-process by Manager.EnablePlugin() -- no server restart
- Print success/failure

plugin disable <name> - Disable plugin:
- POST /api/v1/admin/plugins/{name}/disable
- Handled in-process by Manager.DisablePlugin() -- no server restart
- Print success/failure

plugin approve <name> - Approve routes/hooks:
- Flags: --route "GET /path", --hook "after_insert:content_data", --all-routes, --all-hooks, --yes
- POST /api/v1/admin/plugins/routes/approve or POST /api/v1/admin/plugins/hooks/approve (endpoints created in Step 3)
- For --all-routes: GET /api/v1/admin/plugins/routes, filter unapproved, list each pending item to stderr, prompt "Approve all N routes? [y/N]:" before proceeding. Require --yes flag to skip prompt (for CI/CD and non-interactive use). Also skip prompt when !isatty.IsTerminal(os.Stdout.Fd()).
- For --all-hooks: same confirmation behavior as --all-routes
- Idempotent: approving already-approved items is a no-op
- When zero unapproved items exist, print "no pending routes/hooks for <name>" and exit 0

plugin revoke <name> - Revoke approvals:
- Same flags as approve (including --yes for non-interactive confirmation skip)
- POST /api/v1/admin/plugins/routes/revoke or POST /api/v1/admin/plugins/hooks/revoke
- For --all-routes/--all-hooks: list items and prompt before revoking, same as approve
- Idempotent: revoking already-revoked items is a no-op

Step 5: Auto-create plugins directory + admin API token

Modify: cmd/helpers.go initPluginManager() function
- Before plugin.NewManager(), call os.MkdirAll(dir, 0750)
- Log at INFO: "created plugin directory" with path
- Only log if directory was actually created (check os.Stat first)

Modify: cmd/serve.go
- After server starts, generate 32-byte crypto/rand token, hex-encode it
- Look up the system user: call driver.ListUsers(), find the user with Username "system" (same pattern as internal/cli/model.go:218-224). Store both the UserID and Username.
- Construct AuditContext: audited.Ctx(types.NodeID(""), systemUserID, "plugin-api-token-init", "127.0.0.1") -- empty NodeID is acceptable for system-internal operations, requestID is a descriptive string, IP is localhost.
- Construct CreateTokenParams:
    UserID:    types.NullableUserID{ID: systemUserID, Valid: true},
    TokenType: "api_key",
    Token:     hexEncodedToken,
    IssuedAt:  time.Now().UTC().Format(time.RFC3339),
    ExpiresAt: types.Timestamp{Valid: false},  // zero Timestamp = no expiry (Valid:false means NULL in DB)
    Revoked:   false,
- Insert via driver.CreateToken(ctx, auditCtx, params). Store the returned token ID for deletion on shutdown.
- Write token value to <config_dir>/.plugin-api-token with mode 0600 (<config_dir> = directory of the resolved config.json path)
- On startup, before creating a new token: clean up any stale api_key tokens for the system user left by ungraceful shutdowns. Call driver.GetTokenByUserId(systemNullableUserID), iterate results, delete any with TokenType "api_key" via driver.DeleteToken(ctx, auditCtx, id). This prevents token row accumulation from crashes.
- On graceful shutdown (after httpServer.Shutdown + httpsServer.Shutdown + sshServer.Shutdown, but BEFORE pluginManager.Shutdown -- insert between sshServer shutdown at line ~288 and plugin shutdown at line ~295): delete the token row via driver.DeleteToken(ctx, auditCtx, tokenID), then os.Remove(tokenPath). DB is still open at this point in the shutdown sequence.
- No new middleware needed -- the existing APIKeyAuth() in internal/middleware/middleware.go validates Bearer tokens against the tokens table

Step 6: TUI integration

Modify: internal/cli/pages.go
- Add PLUGINSPAGE PageIndex constant
- Add PLUGINDETAILPAGE PageIndex constant
- Register both in InitPages()

Modify: internal/cli/model.go
- Add fields: PluginManager *plugin.Manager, PluginsList []PluginDisplay, SelectedPlugin string
- New type PluginDisplay with Name, Version, State, CBState, Description
- Update InitialModel() signature to accept *plugin.Manager parameter

Modify: cmd/serve.go
- Pass pluginManager to cli.CliMiddleware() (update signature)

Modify: internal/cli/middleware.go
- Accept *plugin.Manager parameter, pass to InitialModel()

Modify: internal/cli/view.go
- Add PLUGINSPAGE case: TablePage showing plugin list (Name, Version, State, CB)
- Add PLUGINDETAILPAGE case: StaticPage with detailed info + action menu (Enable, Disable, Reload, Approve All Routes, Approve All Hooks)
- Approve All Routes/Hooks actions MUST show a confirmation dialog listing pending items before executing (SSH is interactive, no --yes flag equivalent)

Modify: internal/cli/commands.go
- Add LoadPluginsCmd() - calls manager.ListPlugins(), returns LoadPluginsMsg
- Add PluginEnableCmd(mgr, name, adminUsername), PluginDisableCmd(mgr, name), PluginReloadCmd(mgr, name)
  - EnablePlugin requires an adminUser string for circuit breaker audit trail (manager.go:1042). The TUI model has m.UserID (types.UserID from SSH session) but not the username. Resolve the username by looking up the user: call m.DB.GetUser(string(m.UserID)) to get the User struct, then pass user.Username to mgr.EnablePlugin(ctx, name, username). Store the resolved username on the Model at session start (in InitialModel or on first use) to avoid repeated DB lookups.
- Add PluginApproveRoutesCmd(), PluginApproveHooksCmd() -- use manager.Bridge().ApproveRoute() and manager.HookEngine().ApproveHook()

Modify: internal/cli/update.go
- Add message types: LoadPluginsMsg, PluginActionResultMsg
- Add handlers for plugin messages
- Wire navigation: select plugin from list -> detail page -> action menu

Step 7: Example plugins

New directory: examples/plugins/hello_world/
- init.lua: minimal plugin with one GET route returning {"message": "Hello from ModulaCMS!"}
- Heavily commented explaining manifest fields, on_init, http.handle

New directory: examples/plugins/task_tracker/
- init.lua: defines tasks and categories tables, CRUD routes, before-hook validation
- lib/validators.lua: input validation helpers
- Shows: db.define_table, db.insert/query/update/delete, db.transaction, http.handle (GET/POST/PUT/DELETE), hooks.on, log.*, require()

Step 8: Justfile recipes

Modify: justfile - add [Plugin] section:

# [Plugin] List installed plugins
plugin-list:
    ./{{x86_binary_name}} plugin list

# [Plugin] Create a new plugin scaffold
plugin-init name:
    ./{{x86_binary_name}} plugin init {{name}}

# [Plugin] Validate a plugin
plugin-validate path:
    ./{{x86_binary_name}} plugin validate {{path}}

# [Plugin] Show plugin details
plugin-info name:
    ./{{x86_binary_name}} plugin info {{name}}

# [Plugin] Reload a plugin
plugin-reload name:
    ./{{x86_binary_name}} plugin reload {{name}}

# [Plugin] Enable a plugin
plugin-enable name:
    ./{{x86_binary_name}} plugin enable {{name}}

# [Plugin] Disable a plugin
plugin-disable name:
    ./{{x86_binary_name}} plugin disable {{name}}

---
Files Summary

+---------+--------------------------------------------------+---------------------------------------------------+
| Action  |                       File                       |                      Purpose                      |
+---------+--------------------------------------------------+---------------------------------------------------+
| Modify  | internal/plugin/manager.go                       | Export ExtractManifest() + ValidateManifest()      |
| Create  | internal/plugin/validator.go                     | Offline plugin validation (thin wrapper)           |
| Create  | internal/plugin/validator_test.go                | Validator tests                                    |
| Create  | internal/plugin/hook_handlers.go                 | Hook list/approve/revoke HTTP handlers             |
| Create  | cmd/plugin.go                                    | All CLI plugin commands                            |
| Create  | examples/plugins/hello_world/init.lua            | Minimal example plugin                             |
| Create  | examples/plugins/task_tracker/init.lua           | Full-featured example                              |
| Create  | examples/plugins/task_tracker/lib/validators.lua | Example lib module                                 |
| Modify  | cmd/root.go                                      | Register pluginCmd                                 |
| Modify  | cmd/helpers.go                                   | Auto-create plugins/ dir (0750)                    |
| Modify  | cmd/serve.go                                     | Pass manager to TUI + generate API token via DB    |
| Modify  | internal/cli/pages.go                            | Add plugin page constants                          |
| Modify  | internal/cli/model.go                            | Add plugin state + InitialModel sig                |
| Modify  | internal/cli/middleware.go                       | Accept *plugin.Manager                             |
| Modify  | internal/cli/view.go                             | Render plugin pages                                |
| Modify  | internal/cli/commands.go                         | Plugin async commands                              |
| Modify  | internal/cli/update.go                           | Plugin message handlers                            |
| Modify  | internal/plugin/hook_engine.go                   | Add ListHooks() method                             |
| Modify  | internal/plugin/http_bridge.go                   | Mount hook endpoints in MountAdminEndpoints()      |
| Modify  | internal/router/mux.go                           | MaxBytesReader on existing route approve/revoke    |
| Modify  | justfile                                         | Plugin recipes                                     |
+---------+--------------------------------------------------+---------------------------------------------------+

21 files (7 create, 14 modify)

Parallel Execution Plan

Wave 1 (independent, all parallel):
- Step 1: Export manifest functions + validator package
- Step 3: Hook admin endpoints + route handler hardening
- Step 5: Auto-create plugins dir + admin API token via DB
- Step 7: Example plugins

Wave 2 (depends on Step 1):
- Step 2: CLI offline commands (uses validator)
- Step 8: Justfile recipes (recipe definitions, can stub)

Wave 3 (depends on Steps 2, 3, 5):
- Step 4: CLI online commands (uses HTTP admin API, needs hook endpoints from Step 3, needs token from Step 5)

Wave 4 (depends on Step 4):
- Step 6: TUI integration (needs manager pattern established, hook list from Step 3)

Testing Strategy

- Validator: unit tests in internal/plugin/validator_test.go (Step 1)
- CLI commands: integration tests via exec.Command, done as a separate follow-up task
- TUI: manual verification via SSH

Verification

1. just dev compiles without errors
2. ./modulacms-x86 plugin list shows empty list (no plugins dir yet)
3. ./modulacms-x86 plugin init hello_test creates scaffold in ./plugins/hello_test/
4. ./modulacms-x86 plugin init hello_test --version 1.0.0 --description "test" works non-interactively
5. ./modulacms-x86 plugin validate ./plugins/hello_test passes
6. Copy example plugin, start server with plugin_enabled: true, verify plugin list shows it
7. ./modulacms-x86 plugin approve hello_world --all-routes approves routes
8. SSH into TUI, navigate to Plugins page, see plugin list
9. just test passes (existing + new tests)

Resolved Design Questions
-------------------------

Q: Are online CLI commands intended to work while the server is running?
A: Yes. Online commands hit the running server's HTTP admin API. State changes (enable/disable/reload) are handled in-process by Manager.EnablePlugin(), Manager.DisablePlugin(), Manager.ReloadPlugin(). No server restart is needed.

Q: What happens if plugin init is run and config.json doesn't exist?
A: Error and exit.

Q: What is the test strategy for CLI commands?
A: CLI integration tests (exec.Command) will be done as a separate follow-up task.

Q: How does plugin reload via CLI interact with the file-polling hot reload watcher?
A: CLI reload goes through the server API, which restarts the file-polling watcher for that plugin. No double-reload.

Q: Approve/revoke idempotency?
A: No-op if already in target state. No error, no timestamp update.

Q: How do CLI online commands authenticate to the admin API?
A: cmd/serve.go inserts an API key token into the tokens table (via DbDriver.CreateToken()) tied to the system user at startup. CreateToken requires audited.AuditContext (constructed via audited.Ctx with system user ID, empty NodeID, descriptive requestID, localhost IP) and CreateTokenParams (ExpiresAt: types.Timestamp{Valid: false} for no expiry). Token value is written to <config_dir>/.plugin-api-token (mode 0600). The existing APIKeyAuth() middleware in internal/middleware/middleware.go validates Bearer tokens against the tokens table -- no new middleware needed. On startup, stale api_key tokens for the system user are cleaned up before inserting (prevents row accumulation from crashes). On graceful shutdown (after HTTP/SSH servers stop, before plugin manager shutdown), the token row is deleted and the file removed. On ungraceful shutdown, the stale token is harmless (server is dead) and cleaned up on next startup. For CI/CD, use --token flag.

Q: Is SSH access admin-only?
A: Yes. SSH auth middleware enforces admin-level access. TUI plugin commands do not need additional role checks.

Q: Do --all-routes and --all-hooks require confirmation?
A: Yes. CLI: bulk approve/revoke lists pending items and prompts for confirmation. Skip prompt with --yes flag or when stdout is not a terminal (CI/CD). TUI: confirmation dialog listing pending items before executing (interactive, no --yes equivalent).

Q: Do offline commands require plugin_enabled: true?
A: No. Offline commands (list, init, validate) operate on the filesystem regardless of the plugin_enabled config value. They only need plugin_directory from config.json.

Q: What happens when bulk approve/revoke finds zero pending items?
A: Print "no pending routes/hooks for <name>" and exit 0.

Q: What terminal detection library is used?
A: github.com/mattn/go-isatty (already a dependency). Use isatty.IsTerminal(os.Stdout.Fd()).
