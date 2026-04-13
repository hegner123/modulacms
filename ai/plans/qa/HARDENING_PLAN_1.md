 Hardening Plan: ModulaCMS

 Context

 Audit identified 14 hardening gaps across data integrity, concurrency, security, and error handling. All fixes target existing code with minimal new abstractions. The project is greenfield
 with no active users, so breaking internal signatures is acceptable.

 Four independent workstreams (A-D) can execute in parallel. Each step maintains compilation via just check.

 ---
 Group A: Data Integrity (Publishing Transactions)

 Problem: Publish metadata updates and descendant publishing happen outside the version-creation transaction. Crash between commit and metadata update leaves inconsistent state. Descendant
 errors swallowed with _ =.

 A1. Create InTx variants for publish metadata

 New file: internal/db/content_data_publish_tx.go

 Follow pattern from internal/db/content_version_tx.go (type-switch on DbDriver, mdb.New(tx)):

 func UpdateContentDataPublishMetaInTx(d DbDriver, ctx context.Context, tx *sql.Tx,
     p UpdateContentDataPublishMetaParams) error

 Same for admin variant:
 func UpdateAdminContentDataPublishMetaInTx(d DbDriver, ctx context.Context, tx *sql.Tx,
     p UpdateAdminContentDataPublishMetaParams) error

 Each has 3 driver branches mirroring internal/db/content_data_publish.go:53,114,175 and admin_content_data_publish.go:53,114,175, but using mdb.New(tx) instead of mdb.New(d.Connection).

 Verify: just check

 A2. Move metadata + descendants inside PublishContent transaction

 File: internal/publishing/publishing.go:240-305

 Current: tx ends at line 270, metadata at 276, descendants at 288-304.

 Change: Expand WithTransactionResult closure (line 240) to include:
 1. Root metadata update via UpdateContentDataPublishMetaInTx (replacing line 276)
 2. Descendant loop via UpdateContentDataPublishMetaInTx (replacing lines 288-304)
 3. Propagate descendant errors (remove _ = on line 296)

 Fetch descendants before tx starts (they're read-only context, already fetched for snapshot). Pass descendant list into closure.

 Remove lines 275-305 (post-tx metadata block).

 Verify: just check && just test

 A3. Same fix for PublishAdminContent

 File: internal/publishing/publishing_admin.go:222-283

 Mirror A2: expand WithTransaction closure to include admin metadata + descendant updates. Use UpdateAdminContentDataPublishMetaInTx. Remove lines 253-283.

 Verify: just check && just test

 A4. Wrap UnpublishContent in transaction

 File: internal/publishing/publishing.go:340-357

 Currently lines 342 and 348 are separate non-tx calls. Wrap both in db.WithTransaction:
 - ClearPublishedFlagInTx (already exists in content_version_tx.go:38)
 - UpdateContentDataPublishMetaInTx (created in A1)

 Webhook dispatch and search index update remain outside tx (side effects).

 Verify: just check && just test

 A5. Wrap UnpublishAdminContent in transaction

 File: internal/publishing/publishing_admin.go (UnpublishAdminContent function)

 Same pattern as A4 using admin InTx variants.

 Verify: just check && just test

 ---
 Group B: Concurrency Safety

 B1. Fix search service race + double-close

 File: internal/search/service.go

 Changes to struct (lines 23-31):
 - Replace stopped bool (line 30) with stopped atomic.Bool
 - Add stopOnce sync.Once field

 Stop() (lines 68-76):
 func (s *Service) Stop() error {
     s.stopOnce.Do(func() {
         s.stopped.Store(true)
         close(s.updates)
     })
     <-s.done
     return nil
 }

 OnPublish (line 164) and OnUnpublish (line 177): change s.stopped to s.stopped.Load().

 Verify: just check && just test

 B2. Fix TUI time loop goroutine leak

 File: internal/tui/middleware.go:26-35

 Move goroutine from newProg into teaHandler where s.Context() is available. Use time.NewTicker instead of time.After:

 teaHandler := func(s ssh.Session) *tea.Program {
     // ... existing session setup (lines 37-67) ...
     ctx := s.Context()
     p := newProg(&m, bubbletea.MakeOptions(s)...)
     go func() {
         ticker := time.NewTicker(1 * time.Second)
         defer ticker.Stop()
         for {
             select {
             case <-ticker.C:
                 p.Send(timeMsg(time.Now()))
             case <-ctx.Done():
                 return
             }
         }
     }()
     return p
 }

 Remove the goroutine from newProg (lines 28-33). newProg becomes just tea.NewProgram(m, opts...).

 Verify: just check

 B3. Fix RequestEngine cleanup goroutine race

 File: internal/plugin/request_engine.go

 Add wg sync.WaitGroup field to struct (after line 96).

 Constructor (line 146):
 e.wg.Add(1)
 go func() {
     defer e.wg.Done()
     e.cleanupRateLimiters()
 }()

 Close() (lines 729-732):
 func (e *RequestEngine) Close() {
     close(e.cleanupDone)
     e.wg.Wait()
     e.client.CloseIdleConnections()
 }

 Verify: just check

 B4. Add shutdown tracking to PermissionCache refresh

 File: internal/middleware/authorization.go:142-171

 Change return type to <-chan struct{}:
 func (pc *PermissionCache) StartPeriodicRefresh(ctx context.Context, driver db.RBACRepository, interval time.Duration) <-chan struct{} {
     done := make(chan struct{})
     go func() {
         defer close(done)
         // ... existing goroutine body (lines 144-170) ...
     }()
     return done
 }

 File: cmd/serve.go — update callers (grep found lines 362, 548) to capture return value. During shutdown, select on the done channel with a timeout to confirm goroutine exit.

 File: internal/middleware/client_ip_test.go — update test if it calls StartPeriodicRefresh (check).

 Verify: just check && just test

 ---
 Group C: Security / Input Validation

 C1. Trusted proxy configuration for IP resolution

 File: internal/config/config.go — add field:
 Trusted_Proxies []string `json:"trusted_proxies"`

 File: internal/middleware/client_ip.go

 Change signature to accept parsed CIDRs:
 func ClientIPMiddleware(trustedProxies []net.IPNet) func(http.Handler) http.Handler

 Add ParseTrustedProxies([]string) ([]net.IPNet, error) helper.

 Update resolveClientIP(r, trusted): only check X-Forwarded-For/X-Real-IP when RemoteAddr matches a trusted CIDR. Empty trusted list = RemoteAddr only (secure default).

 File: internal/middleware/ratelimit.go:90-108 — simplify getIP() to only use RemoteAddr (proxy trust already handled upstream by ClientIPMiddleware in context). The getIP fallback path only
  fires when context IP is missing.

 File: internal/middleware/http_chain.go — update 4 call sites (lines 28, 53, 69, 82) to pass parsed proxies.

 File: cmd/serve.go — parse cfg.Trusted_Proxies at startup, pass to middleware chain constructors.

 File: internal/middleware/client_ip_test.go — update tests for new signature, add trusted/untrusted proxy cases.

 Verify: just check && just test

 C2. Upload filename sanitization

 File: internal/media/media_service.go

 Add sanitizeFilename(name string) (string, error) function:
 - filepath.Base() to strip path components
 - Remove null bytes and control chars (0x00-0x1F, 0x7F)
 - Strip leading dots
 - Return error if result empty

 Call before deduplication (before the existing loop around line 121).

 Apply same to admin media upload path (internal/admin/handlers/admin_media.go).

 Verify: just check && just test

 C3. Upload MIME whitelist

 File: internal/config/config.go — add field:
 Allowed_Upload_Types []string `json:"allowed_upload_types"`

 File: internal/media/media_service.go — after http.DetectContentType() (line 116 area), check against whitelist if non-empty. Add isAllowedMIME(contentType string, allowed []string) bool.

 Thread allowedMIMETypes parameter through ProcessMediaUpload (or use config reference).

 Verify: just check && just test

 ---
 Group D: Error Handling

 D1. Create writeJSONResponse helper

 New file: internal/router/json_response.go

 func writeJSONResponse(w http.ResponseWriter, status int, v any) {
     w.Header().Set("Content-Type", "application/json")
     w.WriteHeader(status)
     if err := json.NewEncoder(w).Encode(v); err != nil {
         utility.DefaultLogger.Error("failed to write json response", err)
     }
 }

 Named writeJSONResponse to avoid conflict with existing writeJSON in contentBatch.go.

 Verify: just check

 D2. Migrate router files to writeJSONResponse

 ~206 occurrences across internal/router/ files. Replace patterns:
 w.Header().Set("Content-Type", "application/json")
 w.WriteHeader(http.StatusOK)
 json.NewEncoder(w).Encode(val)
 with:
 writeJSONResponse(w, http.StatusOK, val)

 Process in batches of 5-10 files, just check after each batch. Use checkfor + repfor for mechanical replacement where patterns are consistent.

 Also update writeJSON in contentBatch.go to delegate to writeJSONResponse.

 Verify: just check && just test after all batches.

 D3. Fix OAuth response body read errors

 File: internal/auth/user_provision.go

 Lines 66 and 128: change body, _ := io.ReadAll(...) to check error:
 body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
 if readErr != nil {
     return nil, fmt.Errorf("... (body unreadable: %w)", readErr)
 }

 Verify: just check

 ---
 Execution Order

 Phase 1 (parallel foundations):
   A1: Create InTx methods
   B1: Search service atomic.Bool + sync.Once
   B2: TUI goroutine leak fix
   B3: RequestEngine WaitGroup
   B4: PermissionCache done channel
   C2: Filename sanitization
   D1: writeJSONResponse helper
   D3: OAuth error handling

 Phase 2 (depends on Phase 1):
   A2-A5: Publishing transaction expansion (depends on A1)
   C1: Trusted proxies (touches middleware chain, config, serve.go)
   C3: MIME whitelist (depends on understanding upload flow)
   D2: Router migration (depends on D1)

 Verification

 After all groups complete:
 1. just check — compile verification
 2. just test — full unit test suite
 3. go vet ./... — static analysis
 4. Manual: connect via SSH, verify TUI session creates/destroys goroutine cleanly
 5. Manual: publish/unpublish content, verify atomic metadata state
 6. Manual: upload file with path traversal filename, verify rejection
