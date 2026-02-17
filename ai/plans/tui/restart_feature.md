    ╭───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
     │ Plan: Add "Restart Server" action to TUI                                                                                                                                              │
     │                                                                                                                                                                                       │
     │ Context                                                                                                                                                                               │
     │                                                                                                                                                                                       │
     │ When the Docker container starts with an empty database (no bootstrap data), the permission cache load fails and the server enters SSH-only mode. The operator connects via SSH TUI   │
     │ and runs "DB Init" to populate tables. After that, they need to restart the server so HTTP/HTTPS come up — but currently the only way is to restart the Docker container externally.  │
     │ Adding a "Restart Server" action in the TUI lets operators trigger a graceful restart from within the SSH session.                                                                    │
     │                                                                                                                                                                                       │
     │ Approach                                                                                                                                                                              │
     │                                                                                                                                                                                       │
     │ Thread a restartCh chan struct{} from serve.go through CliMiddleware into the TUI Model. Add a new destructive action "Restart Server" that closes this channel. In serve.go, select  │
     │ on both done (OS signals) and restartCh to trigger graceful shutdown — Docker's restart policy brings the process back up.                                                            │
     │                                                                                                                                                                                       │
     │ Changes                                                                                                                                                                               │
     │                                                                                                                                                                                       │
     │ 1. internal/cli/model.go — Add restart channel to Model                                                                                                                               │
     │                                                                                                                                                                                       │
     │ Add field to Model struct:                                                                                                                                                            │
     │ RestartCh chan struct{}                                                                                                                                                               │
     │                                                                                                                                                                                       │
     │ 2. internal/cli/middleware.go — Thread channel through CliMiddleware                                                                                                                  │
     │                                                                                                                                                                                       │
     │ Update CliMiddleware signature to accept restartCh chan struct{}:                                                                                                                     │
     │ func CliMiddleware(v *bool, c *config.Config, driver db.DbDriver, logger Logger, pluginMgr *plugin.Manager, mgr *config.Manager, restartCh chan struct{}) wish.Middleware             │
     │                                                                                                                                                                                       │
     │ Pass to model in InitialModel or assign directly after model creation:                                                                                                                │
     │ m.RestartCh = restartCh                                                                                                                                                               │
     │                                                                                                                                                                                       │
     │ 3. internal/cli/init.go (or wherever InitialModel lives) — Accept channel                                                                                                             │
     │                                                                                                                                                                                       │
     │ Update InitialModel to accept and store restartCh. If InitialModel doesn't take it, just assign on the Model directly in middleware.go after InitialModel returns (simpler).          │
     │                                                                                                                                                                                       │
     │ 4. internal/cli/actions.go — Add Restart Server action                                                                                                                                │
     │                                                                                                                                                                                       │
     │ Add to ActionsMenu() at index 12:                                                                                                                                                     │
     │ {Label: "Restart Server", Description: "Gracefully restart all servers", Destructive: true},                                                                                          │
     │                                                                                                                                                                                       │
     │ Add case to RunDestructiveActionCmd:                                                                                                                                                  │
     │ case 12:                                                                                                                                                                              │
     │     return runRestartServer(p)                                                                                                                                                        │
     │                                                                                                                                                                                       │
     │ Add runRestartServer function:                                                                                                                                                        │
     │ func runRestartServer(p ActionParams) tea.Cmd {                                                                                                                                       │
     │     return func() tea.Msg {                                                                                                                                                           │
     │         if p.RestartCh == nil {                                                                                                                                                       │
     │             return ActionResultMsg{                                                                                                                                                   │
     │                 Title:   "Restart Failed",                                                                                                                                            │
     │                 Message: "Restart channel not available.",                                                                                                                            │
     │                 IsError: true,                                                                                                                                                        │
     │             }                                                                                                                                                                         │
     │         }                                                                                                                                                                             │
     │         close(p.RestartCh)                                                                                                                                                            │
     │         return ActionResultMsg{                                                                                                                                                       │
     │             Title:   "Server Restarting",                                                                                                                                             │
     │             Message: "Graceful shutdown initiated. The server will restart momentarily.",                                                                                             │
     │         }                                                                                                                                                                             │
     │     }                                                                                                                                                                                 │
     │ }                                                                                                                                                                                     │
     │                                                                                                                                                                                       │
     │ Add RestartCh to ActionParams:                                                                                                                                                        │
     │ type ActionParams struct {                                                                                                                                                            │
     │     Config         *config.Config                                                                                                                                                     │
     │     UserID         types.UserID                                                                                                                                                       │
     │     SSHFingerprint string                                                                                                                                                             │
     │     SSHKeyType     string                                                                                                                                                             │
     │     SSHPublicKey   string                                                                                                                                                             │
     │     RestartCh      chan struct{}                                                                                                                                                      │
     │ }                                                                                                                                                                                     │
     │                                                                                                                                                                                       │
     │ 5. cmd/serve.go — Create restart channel, listen on it                                                                                                                                │
     │                                                                                                                                                                                       │
     │ Create the channel near the done channel:                                                                                                                                             │
     │ restartCh := make(chan struct{})                                                                                                                                                      │
     │                                                                                                                                                                                       │
     │ Pass to CliMiddleware:                                                                                                                                                                │
     │ cli.CliMiddleware(&verbose, cfg, driver, utility.DefaultLogger, pluginManager, mgr, restartCh)                                                                                        │
     │                                                                                                                                                                                       │
     │ Replace <-done with a select:                                                                                                                                                         │
     │ select {                                                                                                                                                                              │
     │ case <-done:                                                                                                                                                                          │
     │ case <-restartCh:                                                                                                                                                                     │
     │     utility.DefaultLogger.Info("Restart requested via TUI")                                                                                                                           │
     │ }                                                                                                                                                                                     │
     │                                                                                                                                                                                       │
     │ 6. deploy/docker/docker-compose.postgres.yml — Add restart policy                                                                                                                     │
     │                                                                                                                                                                                       │
     │ Add restart: unless-stopped to the modulacms service so the container comes back up after the graceful shutdown.                                                                      │
     │                                                                                                                                                                                       │
     │ 7. Find where ActionParams is constructed and add RestartCh                                                                                                                           │
     │                                                                                                                                                                                       │
     │ Search for where ActionParams{...} is built (likely in update_controls.go or similar) and pass m.RestartCh into it.                                                                   │
     │                                                                                                                                                                                       │
     │ Files to modify                                                                                                                                                                       │
     │                                                                                                                                                                                       │
     │ - internal/cli/model.go — Add RestartCh field to Model struct                                                                                                                         │
     │ - internal/cli/middleware.go — Thread restartCh through CliMiddleware                                                                                                                 │
     │ - internal/cli/actions.go — Add action item, handler, RestartCh to ActionParams                                                                                                       │
     │ - cmd/serve.go — Create channel, pass to middleware, select on it                                                                                                                     │
     │ - deploy/docker/docker-compose.postgres.yml — Add restart policy                                                                                                                      │
     │ - File where ActionParams is constructed — pass m.RestartCh                                                                                                                           │
     │                                                                                                                                                                                       │
     │ Verification                                                                                                                                                                          │
     │                                                                                                                                                                                       │
     │ 1. go vet ./cmd/ ./internal/cli/ — compile check                                                                                                                                      │
     │ 2. just test — run tests                                                                                                                                                              │
     │ 3. Docker: just docker-postgres-up, verify SSH-only mode on permission failure, run "DB Init" via TUI, then "Restart Server" — container should come back up with HTTP/HTTPS running
