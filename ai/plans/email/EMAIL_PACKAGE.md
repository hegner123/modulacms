 Plan: Email Package (internal/email/)

 Context

 ModulaCMS needs transactional email (password resets, invites, notifications). Config fields already exist (internal/config/config.go). This plan implements the email package in two
 commits: Part A (provider-agnostic message assembly) and Part B (provider connections and sending service).

 Two new config fields are needed for SES credentials (Part B scope):
 - Email_AWS_Access_Key_ID and Email_AWS_Secret_Access_Key

 ---
 Part A: Message Assembly (Commit 1)

 New files

 internal/email/message.go — package email. Types and validation, zero external dependencies.

 Types:
 - Address — wraps net/mail.Address, adds Validate() error. Constructor: NewAddress(name, email string) Address
   Validate() checks beyond net/mail.Address parsing:
   - Address string must contain "@"
   - Local part (before @) must be non-empty
   - Domain part (after @) must be non-empty
   - net/mail.ParseAddress must succeed (RFC 5322 syntax)
   These checks are explicit string operations, not regex.
   Note: domain is NOT required to contain a dot. Dotless domains like "localhost" and "intranet" are
   valid per RFC 5321 and commonly used with local SMTP test tools (Mailhog, MailPit, smtp4dev).
   Rejecting them would block local development on day one.
 - Attachment — Filename string, ContentType string, Data []byte, with Validate() error
 - Message — full email envelope:

 type Message struct {
     From        Address
     To          []Address
     CC          []Address
     BCC         []Address
     ReplyTo     *Address
     Subject     string
     PlainBody   string
     HTMLBody    string
     Attachments []Attachment
     Headers     map[string]string
 }

 - Message.Validate() error — checks all fields before any send attempt
 - Message.allRecipients() []string — unexported, flattens To+CC+BCC for SMTP RCPT TO

 Validation rules:
 - From must pass Address.Validate()
 - To must be non-empty, each address valid
 - CC, BCC each address valid if present
 - ReplyTo valid if non-nil
 - Subject non-blank
 - At least one of PlainBody or HTMLBody non-blank
 - Each Attachment valid (filename, content-type required, data non-empty)
 - Total message size check: sum of all attachment Data lengths must not exceed MaxMessageSize (7.5 MB).
   Set to 7.5 MB rather than 10 MB because base64 encoding inflates data by ~33%, and Postmark's 10 MB
   limit applies to the encoded API payload. 7.5 MB raw ensures the encoded payload stays under 10 MB.
   Defined as a package-level constant.
 - Total recipient count (len(To) + len(CC) + len(BCC)) must not exceed MaxRecipients (50).
   SES limits to 50 recipients per call; SMTP servers typically reject beyond 100 RCPT TO commands.
   50 is the safe lowest common denominator. Defined as a package-level constant.
 - Headers map validation (both keys and values):
   Keys: scanned byte-by-byte; must contain only printable ASCII (0x21-0x7E) excluding colon (0x3A).
   Reject keys containing colon, space, null byte, CR, LF, or any non-printable character with
   *InvalidMessageError{Field: "Headers", Problem: "header key contains invalid character"}.
   Per RFC 5322 Section 2.2, field names are printable US-ASCII excluding colon.
   Values: scanned byte-by-byte for CR (\r) and LF (\n) characters.
   If any value contains CR or LF, return *InvalidMessageError{Field: "Headers", Problem: "header value
   contains CR or LF (header injection)"} — this prevents header injection attacks where a CRLF in a
   header value allows arbitrary header/body manipulation (RFC 5322 Section 2.2).

 internal/email/errors.go — Named error types.

 // InvalidMessageError — validation failure, inspectable Field+Problem
 type InvalidMessageError struct {
     Field   string
     Problem string
 }

 // ProviderError — remote provider rejected, wraps upstream error
 type ProviderError struct {
     Provider string
     Code     int   // HTTP status, 0 for SMTP/dial
     Err      error
 }
 func (e *ProviderError) Unwrap() error { return e.Err }

 // DisabledError — email sending is disabled
 type DisabledError struct{}

 All implement error. Callers use errors.As to distinguish.

 internal/email/mime.go — MIME message building (unexported, used by SMTP sender in Part B but testable independently in Part A).

 - buildMIME(msg Message) ([]byte, error) — builds RFC 2045 MIME bytes
   Required headers generated automatically if not present in msg.Headers:
   - MIME-Version: 1.0 — required by RFC 2045 Section 4; without it some MTAs ignore Content-Type
     declarations and render HTML as source code or break attachments
   - Date: time.Now().UTC().Format(time.RFC1123Z) — required origination field per RFC 5322 Section 3.6.1;
     missing Date is a spam signal for corporate mail filters (Proofpoint, Mimecast)
   - Message-ID: <hex@from-domain> — generates 16 random bytes via crypto/rand, hex-encoded (32 chars),
     combined with the domain from msg.From. Guarantees uniqueness with zero internal dependencies.
     Does NOT import internal/db/types for ULID — avoids coupling the email package to database types.
     Required for threading, deduplication, and log correlation (RFC 5322 Section 3.6.4).
     Missing Message-ID is a spam signal.
   Body structure selection:
   - Plain only, no attachments: text/plain; charset=utf-8
   - HTML only, no attachments: text/html; charset=utf-8
   - Both bodies: multipart/alternative
   - Any attachments: multipart/mixed outer wrapping alternative inner + base64-encoded attachments
 - joinAddresses([]Address) string — comma-separated RFC 5322 addresses
 - Uses: mime/multipart, net/textproto, encoding/base64, encoding/hex, crypto/rand, mime, time (stdlib only)
 - No internal package imports — mime.go depends only on stdlib

 internal/email/message_test.go — Validation + MIME tests.

 ┌────────────────────────────────────┬───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │                Test                │                                                                   Cases                                                                   │
 ├────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ TestAddress_Validate               │ empty, invalid format, missing @, empty local part, empty domain, valid bare, valid with name, valid dotless domain (localhost)                                                                       │
 ├────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ TestAttachment_Validate            │ missing filename, missing content-type, empty data, valid                                                                                 │
 ├────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ TestMessage_Validate               │ missing from, empty to, invalid to addr, empty subject, no body, both bodies valid, plain only valid, HTML only valid, invalid attachment, attachments exceed 7.5MB, recipients exceed 50, header key with colon, header key with null byte, header value with CRLF │
 ├────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ TestMessage_AllRecipients          │ to+cc+bcc all included                                                                                                                    │
 ├────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ TestBuildMIME_PlainOnly            │ Content-Type: text/plain header present                                                                                                   │
 ├────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ TestBuildMIME_HTMLOnly             │ Content-Type: text/html header present                                                                                                    │
 ├────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ TestBuildMIME_MultipartAlternative │ both bodies, multipart/alternative boundary                                                                                               │
 ├────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ TestBuildMIME_WithAttachments      │ multipart/mixed wraps body + base64 attachment                                                                                            │
 ├────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ TestBuildMIME_SubjectEncoding      │ non-ASCII subject Q-encoded                                                                                                               │
 ├────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ TestBuildMIME_RequiredHeaders      │ MIME-Version, Date, Message-ID all present in output; Message-ID format is <hex@domain> (32 hex chars); Date parses as RFC 1123Z       │
 ├────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ TestBuildMIME_PreservesCustomHeaders│ Headers map entry not overwritten if Message-ID or Date already set by caller                                                          │
 └────────────────────────────────────┴───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

 All tests parallel, table-driven with t.Run().

 Part A verification

 1. go build ./internal/email/... — compiles
 2. go test ./internal/email/... -v — all tests pass
 3. go vet ./internal/email/... — clean

 ---
 Part B: Providers + Service (Commit 2)

 Config changes (new fields for SES)

 internal/config/config.go — Add to Config struct in the email section:
 Email_AWS_Access_Key_ID     string `json:"email_aws_access_key_id"`
 Email_AWS_Secret_Access_Key string `json:"email_aws_secret_access_key"`

 internal/config/field_meta.go — Add 2 entries to FieldRegistry in the Email section:
 - email_aws_access_key_id — HotReloadable: true, Sensitive: true
 - email_aws_secret_access_key — HotReloadable: true, Sensitive: true

 internal/config/default.go — Both fields default to ""

 internal/config/validate.go — Three changes:
 1. Exclude EmailSES from the existing email_api_key emptiness warning (line 54). Change:
      (c.Email_Provider == EmailSendGrid || c.Email_Provider == EmailSES || c.Email_Provider == EmailPostmark) && c.Email_API_Key == ""
    to:
      (c.Email_Provider == EmailSendGrid || c.Email_Provider == EmailPostmark) && c.Email_API_Key == ""
 2. Add new SES-specific validation after the existing email_api_key check: if provider is ses and both
    Email_AWS_Access_Key_ID and Email_AWS_Secret_Access_Key are empty, add a warning (not error) with text:
    "email_aws_access_key_id and email_aws_secret_access_key are empty; SES will use the default
    AWS credential chain (environment variables, IAM role, etc.) — this is expected on EC2/ECS/Lambda"
    This avoids confusing operators who are correctly using IAM roles.
 3. Add 2 cases to configFieldString():
      case "email_aws_access_key_id":
          return c.Email_AWS_Access_Key_ID
      case "email_aws_secret_access_key":
          return c.Email_AWS_Secret_Access_Key

 internal/config/redact.go — Add redaction for both new fields.

 New files

 internal/email/sender.go — Sender interface and factory.

 type Sender interface {
     Send(ctx context.Context, msg Message) error
     Close() error
 }

 Close() releases resources held by the sender (HTTP client idle connections, etc.).
 Called by Service.Reload on the old sender after swapping. noopSender and SMTPSender
 return nil. HTTP-based senders (SendGrid, Postmark, SES) call client.CloseIdleConnections().

 - noopSender — Send returns *DisabledError; Close returns nil
 - NewSender(cfg config.Config) (Sender, error) — switches on cfg.Email_Provider:
   - EmailDisabled or !Email_Enabled → noopSender{}
   - EmailSmtp → newSMTPSender(cfg)
   - EmailSendGrid → newSendGridSender(cfg)
   - EmailSES → newSESSender(cfg)
   - EmailPostmark → newPostmarkSender(cfg)
   - Unknown → error

 internal/email/service.go — Hot-reloadable service wrapper.

 type Service struct {
     mu      sync.RWMutex
     sender  Sender
     from    Address
     replyTo *Address
 }

 - NewService(cfg config.Config) (*Service, error) — constructs sender + validates default from/reply-to
 - Send(ctx context.Context, msg Message) error — injects default From if unset, default ReplyTo if nil, validates, delegates to sender under RLock
 - Reload(cfg config.Config) error — builds new sender, swaps under write lock, then calls Close() on the
   old sender to release HTTP idle connections. Called from config.Manager.OnChange.
   In-flight behavior: a Send that has already acquired RLock and is mid-transmission through the old sender
   will complete (or fail) using the old sender's connection. Close() only releases idle connections in the
   pool, not active ones — in-flight HTTP requests continue to completion on their existing connections.
   This is acceptable for a CMS -- at worst a single in-flight email during a config reload may fail.
   Note: there is no automatic retry mechanism in v1. If an in-flight send fails during reload, the
   calling handler receives the error and must decide how to respond (e.g., return 500 to the user).
   Adding retry/queuing is a future enhancement. This is documented in the Reload godoc.
 - Close() error — calls sender.Close() under write lock. Called from defer in serve.go shutdown.
 - Enabled() bool — reports whether sender is not noopSender.
   Implementation: type assertion `_, isNoop := s.sender.(*noopSender); return !isNoop`.
   Do NOT add an Enabled() method to the Sender interface.

 internal/email/smtp.go — SMTP sender using net/smtp stdlib.

 type SMTPSender struct {
     host, username, password string
     port                     int
     useTLS                   bool
 }

 Timeout is NOT stored on the struct. Send() computes it locally:
 30s for messages without attachments, 120s for messages with attachments.
 This avoids the problem of a single fixed timeout being too short for large attachment DATA transfers
 over high-latency links (7.5MB over a slow relay can take 60+ seconds for DATA alone).

 - newSMTPSender(cfg) (*SMTPSender, error) — validates Email_Host required, defaults port to 587 (STARTTLS) or 465 (TLS)
 - Send(ctx, msg) error — calls msg.Validate() as defense-in-depth (Service.Send also validates,
   but the SMTP sender validates independently so it is safe to use without the Service wrapper),
   then calls buildMIME(msg), then:
   - Computes timeout locally: 30s if len(msg.Attachments) == 0, else 120s.
   - Both paths use net.Dialer{}.DialContext(ctx, "tcp", host:port) for context-aware connection.
     After dial, conn.SetDeadline(time.Now().Add(timeout)) bounds the entire SMTP conversation.
   - TLS mode (port 465): tls.Client(conn, tlsConfig) → smtp.NewClient → Auth → Mail/Rcpt/Data
   - STARTTLS mode (port 587): smtp.NewClient(conn) → client.StartTLS(tlsConfig) → Auth → Mail/Rcpt/Data
   - Does NOT use smtp.SendMail — it ignores context.Context and has no timeout support, so a hung
     SMTP server would block the goroutine indefinitely.
 - Close() error — returns nil (no persistent connections; each Send dials fresh)
 - All errors wrapped in *ProviderError{Provider: "smtp"}
 - Known limitation — SMTP auth compatibility:
   PLAIN auth over TLS only. LOGIN and XOAUTH2 are not supported. smtp.PlainAuth also refuses to
   authenticate over non-TLS connections (Go stdlib safety check).
   Compatible: Amazon SES SMTP, Mailgun SMTP, Postfix, Sendmail, MailPit/Mailhog (no auth), any
   relay that accepts PLAIN over STARTTLS or implicit TLS.
   Incompatible: Microsoft 365 (deprecated PLAIN, requires XOAUTH2), Google Workspace (deprecated
   PLAIN, requires XOAUTH2 or app passwords).
   Operators using M365 or Google Workspace must use an HTTP API provider (SendGrid, SES, Postmark)
   instead of SMTP. Document this prominently in the godoc and in any operator-facing setup guide.

 internal/email/sendgrid.go — SendGrid v3 API via net/http.

 type SendGridSender struct {
     apiKey, endpoint string
     client           *http.Client
 }

 - Default endpoint: https://api.sendgrid.com/v3/mail/send
 - Auth: Authorization: Bearer {api_key}
 - JSON payload: personalizations, from, subject, content, attachments (base64)
 - Success: 202, Error: 4xx/5xx wrapped in *ProviderError
 - Response body always drained and closed for connection reuse
 - Close() error — calls client.CloseIdleConnections(); returns nil

 internal/email/ses.go — AWS SES v1 via aws-sdk-go v1 (already in go.mod for bucket package).

 Uses the v1 SES API (github.com/aws/aws-sdk-go/service/ses), NOT sesv2. The v1 ses package may not
 be vendored yet — run `go mod vendor` after adding the import. The v1 SES SendRawEmail API is simpler
 than sesv2 for raw MIME sending and avoids adding a new service package with different type conventions.

 type SESSender struct {
     sess   *session.Session
     client *ses.SES
     region string
 }

 - Uses github.com/aws/aws-sdk-go v1 which is already a project dependency
 - Credentials: Email_AWS_Access_Key_ID + Email_AWS_Secret_Access_Key via credentials.NewStaticCredentials.
   If both are empty, falls back to session.NewSession default credential chain (env vars, IAM role, etc.)
 - Region detection: if Email_API_Endpoint is non-empty and contains ".amazonaws.com",
   split on dots via strings.Split(host, ".") and take the second segment
   (e.g., "email.us-west-2.amazonaws.com" -> "us-west-2"). Otherwise default to "us-east-1".
   This is a simple strings.Split + index operation, not regex.
 - Custom endpoint: if Email_API_Endpoint is set, override the SDK endpoint resolver
 - Sends via ses.SendRawEmail with ses.SendRawEmailInput:
   - Calls buildMIME(msg) to produce RFC 2045 bytes (reuses Part A's MIME builder)
   - Wraps in ses.SendRawEmailInput{RawMessage: &ses.RawMessage{Data: mimeBytes}}
   - Sets Destinations to msg.allRecipients() (explicit envelope recipients)
   - This supports attachments, multipart/alternative, and custom headers uniformly
 - Close() error — calls client.Config.HTTPClient.CloseIdleConnections() if client set; returns nil otherwise
 - All errors wrapped in *ProviderError{Provider: "ses"}, extracting HTTP status from awserr.RequestFailure if available

 internal/email/postmark.go — Postmark API via net/http.

 type PostmarkSender struct {
     serverToken, endpoint string
     client                *http.Client
 }

 - Default endpoint: https://api.postmarkapp.com/email
 - Auth: X-Postmark-Server-Token: {api_key}
 - JSON payload: From, To (comma-separated), Subject, TextBody, HtmlBody, Attachments (base64)
 - Success: 200, Error: 422+ wrapped in *ProviderError
 - Close() error — calls client.CloseIdleConnections(); returns nil

 internal/email/helpers.go — Shared unexported utilities.

 - encodeBase64([]byte) string — used by sendgrid and postmark senders
 - newHTTPClient(timeout time.Duration) *http.Client — shared HTTP client constructor with timeout.
   MUST create its own *http.Transport (not http.DefaultTransport) to isolate connection pools from
   the S3/bucket HTTP clients and any other HTTP clients in the process. Use sensible defaults:
   MaxIdleConns: 10, MaxIdleConnsPerHost: 2, IdleConnTimeout: 90s. This also ensures that
   Close()/CloseIdleConnections() only affects the email sender's connections.
   All HTTP-based senders (SendGrid, Postmark, SES) call newHTTPClient(30 * time.Second).

 Test files

 internal/email/sender_test.go — Factory + noop tests:
 - TestNewSender_Disabled — Email_Enabled=false → noopSender
 - TestNewSender_DisabledProvider — provider "" → noopSender
 - TestNewSender_UnknownProvider — returns error
 - TestNoopSender_ReturnsDisabledError — errors.As check
 - TestNewSender_SMTPMissingHost — returns error
 - TestNewSender_SendGridMissingKey — returns error
 - TestNewSender_PostmarkMissingKey — returns error
 - TestNewSender_SESDefaultCredentialChain — both AWS fields empty, NewSender succeeds,
   session falls back to default credential chain (env vars, IAM role, etc.)

 internal/email/service_test.go — Service behavior:
 - TestNewService_DisabledConfig — no error, Enabled() false
 - TestNewService_InvalidFromAddress — returns error
 - TestService_Send_InjectsFrom — msg.From empty gets default from config
 - TestService_Send_InjectsReplyTo — msg.ReplyTo nil gets default from config
 - TestService_Send_ValidationFailure — returns *InvalidMessageError
 - TestService_Reload_SwapsSender — sender changes after reload
 - TestService_Reload_ConcurrentSafe — goroutines Send while Reload runs, use -race
 - TestService_Reload_InFlightCompletes — start a slow mock send, trigger reload mid-flight, verify send completes

 internal/email/smtp_test.go — SMTP tests (local net.Listener, no real server):
 - TestSMTPSender_ValidatesBeforeDial — invalid msg returns InvalidMessageError, no connection
 - TestSMTPSender_SendPlainAuth — local TCP listener speaking minimal SMTP (220/250/235/354/250/221)
 - TestSMTPSender_BuildMIMEIntegration — parse result with mail.ReadMessage
 - TestSMTPSender_RespectsContext — cancelled context causes Send to return before completing SMTP handshake
 - TestSMTPSender_TimeoutOnHungServer — listener accepts but never responds; verify Send returns within 2x timeout

 internal/email/http_senders_test.go — HTTP provider tests using httptest.NewServer:
 - TestSendGridSender_CorrectPayload — verify auth header, JSON shape
 - TestSendGridSender_4xxReturnsProviderError — verify ProviderError.Code
 - TestPostmarkSender_CorrectHeaders — X-Postmark-Server-Token present
 - TestPostmarkSender_AttachmentBase64 — verify attachment encoding
 - TestSESSender_SendsRawMIME — verify ses.SendRawEmail receives RawMessage with MIME bytes (mock AWS client)
 - TestSESSender_WithAttachments — attachments included in MIME bytes (not rejected)
 - TestSESSender_ProviderErrorWrapping — awserr mapped to *ProviderError with HTTP status
 - TestSESSender_DefaultCredentialChain — both AWS key fields empty, session falls back to default chain

 Startup wiring (cmd/serve.go + cmd/helpers.go)

 Add initEmailService helper to cmd/helpers.go following the initObservability pattern:

 func initEmailService(cfg *config.Config) (*email.Service, error) {
     svc, err := email.NewService(*cfg)
     if err != nil {
         if cfg.Email_Enabled {
             // Email was explicitly enabled but config is invalid -- fail startup so the
             // operator notices immediately rather than discovering silent failures later
             // (e.g., password resets that never arrive).
             return nil, fmt.Errorf("email service init failed (email_enabled=true): %w", err)
         }
         utility.DefaultLogger.Warn("email service init failed, sending disabled", "error", err)
         disabled := config.Config{Email_Enabled: false}
         svc, _ = email.NewService(disabled)
     }
     if svc.Enabled() {
         utility.DefaultLogger.Info("Email service started", "provider", cfg.Email_Provider)
     }
     return svc, nil
 }

 In cmd/serve.go, insert the emailSvc block immediately after the obsCleanup block
 (after line 121: `defer obsCleanup()`) and before the install.CheckInstall call (line 124).
 This ensures email is available before any handler that might need it, and the defer
 ordering is: db.CloseDB > pluginPoolCleanup > obsCleanup > emailSvc.Close
 (innermost defer runs first).

 emailSvc, err := initEmailService(cfg)
 if err != nil {
     return fmt.Errorf("email: %w", err)
 }
 defer emailSvc.Close()
 mgr.OnChange(func(newCfg config.Config) {
     if err := emailSvc.Reload(newCfg); err != nil {
         utility.DefaultLogger.Error("email service hot-reload failed", "error", err)
     }
 })

 emailSvc is not injected into the router yet (no endpoints send email yet). When password reset or invite endpoints are added, pass emailSvc into router.NewModulacmsMux.

 Part B verification

 1. go mod vendor — vendor the ses service package (github.com/aws/aws-sdk-go/service/ses)
 2. Verify vendor/github.com/aws/aws-sdk-go/service/ses/ directory exists
 3. go build ./internal/config/... — config changes compile
 4. go test ./internal/config/... — existing config tests pass
 5. go build ./internal/email/... — email package compiles
 6. go test ./internal/email/... -v -race — all tests pass with race detector
 7. go build ./cmd/... — serve.go wiring compiles
 8. go vet ./internal/email/... — clean

 ---
 File summary

 ┌─────────────────────────────────────┬──────┬───────────────────────────────────┐
 │                File                 │ Part │              Action               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/message.go           │ A    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/errors.go            │ A    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/mime.go              │ A    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/message_test.go      │ A    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/sender.go            │ B    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/service.go           │ B    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/smtp.go              │ B    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/sendgrid.go          │ B    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/ses.go               │ B    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/postmark.go          │ B    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/helpers.go           │ B    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/sender_test.go       │ B    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/service_test.go      │ B    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/smtp_test.go         │ B    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/email/http_senders_test.go │ B    │ new                               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/config/config.go           │ B    │ modify (2 fields)                 │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/config/field_meta.go       │ B    │ modify (2 entries)                │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/config/default.go          │ B    │ modify (2 defaults)               │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/config/validate.go         │ B    │ modify (SES exclusion + SES AWS warning + 2 configFieldString cases) │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ internal/config/redact.go           │ B    │ modify (2 redactions)             │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ cmd/helpers.go                      │ B    │ modify (initEmailService)         │
 ├─────────────────────────────────────┼──────┼───────────────────────────────────┤
 │ cmd/serve.go                        │ B    │ modify (wiring + OnChange)        │
 └─────────────────────────────────────┴──────┴───────────────────────────────────┘

 Known v1 limitations (deferred to future work)

 - No retry/queuing: Send is fire-and-forget. If a send fails, the caller receives the error and must
   decide how to respond. There is no automatic retry, dead letter queue, or delivery confirmation.
   The audited command pattern (internal/db/audited/) could be extended to log email send attempts
   and outcomes for audit trail. Retry with exponential backoff could wrap the Sender interface.
 - No rate limiting: SES defaults to 1/second in sandbox, 14/second in production. SendGrid has daily
   limits. If email is hooked into content lifecycle events, provider throttling will hit. A token-bucket
   rate limiter in the Service layer would address this.
 - No email templates: PlainBody and HTMLBody are raw strings. Every caller builds its own content.
   A template layer (html/template with embedded FS for built-in templates) should be added before
   the second email-sending feature to prevent inconsistent, broken HTML emails across callers.
 - No config test command: operators have no way to verify email configuration works without triggering
   a real flow (password reset). A `modulacms config test-email` CLI command or admin API endpoint
   that sends a test message to a specified address should be added alongside the first email-sending
   feature (password reset).
 - SMTP PLAIN auth only: see smtp.go compatibility notes. XOAUTH2 support is a future enhancement.
 - aws-sdk-go v1 deprecation: AWS marked v1 as maintenance mode (Nov 2023) with end-of-support
   July 2025. When the project migrates S3/bucket code to v2, the SES sender must move too. This is
   tracked as migration debt, not a v1 blocker.

 Caller error handling contract

 When Service.Send returns an error, the calling handler (password reset, invite) is responsible for:
 - *InvalidMessageError: programming error, log and return 500 (should not happen with well-formed callers)
 - *ProviderError: transient or permanent provider failure. The handler should return a generic success
   response to the user (e.g., "if that email exists, you'll receive a reset link") to avoid leaking
   whether the email address is registered. Log the error for operator visibility.
 - *DisabledError: email is disabled. The handler should return a user-facing error explaining that
   email is not configured (this is an operator misconfiguration, not a user error).
 The password reset endpoint plan (separate from this plan) must implement this contract.

 Key patterns to reuse

 - config.Manager.OnChange(func(Config)) — internal/config/manager.go:196 — hot-reload callback
 - middleware.PermissionCache — internal/middleware/authorization.go — RWMutex + build-then-swap pattern for Service.Reload
 - initObservability — cmd/helpers.go:72 — startup helper returning cleanup/noop pattern
 - media.DuplicateMediaError / media.FileTooLargeError — internal/media/media_service.go — named error type pattern
 - net/mail.Address — stdlib RFC 5322 address type
