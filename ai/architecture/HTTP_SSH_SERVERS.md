# HTTP_SSH_SERVERS.md

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/HTTP_SSH_SERVERS.md`
**Purpose:** Documents ModulaCMS's triple-server architecture with HTTP, HTTPS, and SSH servers running concurrently.
**Last Updated:** 2026-01-12

---

## Overview

ModulaCMS runs three concurrent servers in a single Go binary:

1. **HTTP Server** - Fallback server for non-TLS connections
2. **HTTPS Server** - Primary server with automatic Let's Encrypt TLS certificates
3. **SSH Server** - Terminal User Interface (TUI) access for developer/ops management

All three servers start concurrently using goroutines and support graceful shutdown with proper cleanup. This architecture provides multiple access methods to the CMS: HTTP/HTTPS for API and content delivery, and SSH for interactive management.

**Key Benefits:**
- Single binary deployment (no separate processes to manage)
- Automatic HTTPS with Let's Encrypt (no manual certificate management)
- Development-friendly local mode (no SSL certificates needed)
- Interactive SSH management interface
- Graceful shutdown across all servers

---

## Server Architecture

### Entry Point

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go`

The `main()` function calls `run()` which:
1. Parses command-line flags
2. Loads configuration
3. Initializes database connection
4. Creates all three servers
5. Launches them concurrently in goroutines
6. Waits for shutdown signal
7. Performs graceful shutdown

### Server Lifecycle

```
main() → run()
  ↓
  ├─ Parse flags and load config
  ├─ Initialize database
  ├─ Create SSH server
  ├─ Create HTTP server
  ├─ Create HTTPS server
  ↓
  Launch goroutines:
  ├─ SSH server goroutine
  └─ HTTP/HTTPS server goroutine
  ↓
  Wait for signal (SIGINT/SIGTERM)
  ↓
  Graceful shutdown (30s timeout):
  ├─ HTTP server shutdown
  ├─ HTTPS server shutdown
  └─ SSH server shutdown
```

---

## HTTP Server

### Purpose

The HTTP server provides:
- Fallback access when HTTPS is not available
- Local development access without TLS certificates
- API endpoint access for client applications

### Configuration

HTTP server settings from **Config struct** (`internal/config/config.go`):

```go
type Config struct {
    Port        string  `json:"port"`         // HTTP port (e.g., "8080")
    Client_Site string  `json:"client_site"`  // Client domain
    Admin_Site  string  `json:"admin_site"`   // Admin domain
    Environment string  `json:"environment"`  // "local", "staging", "production"
    // ... other fields
}
```

### Server Creation

**Location:** `cmd/main.go:148-154`

```go
httpServer = &http.Server{
    Addr:         configuration.Client_Site + configuration.Port,
    Handler:      middlewareHandler,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

**Key Settings:**
- **Addr:** Combines client site domain and port (e.g., `example.com:8080`)
- **Handler:** HTTP multiplexer wrapped with middleware (CORS, auth, session)
- **ReadTimeout:** 15 seconds (prevents slow client attacks)
- **WriteTimeout:** 15 seconds (prevents slow response attacks)
- **IdleTimeout:** 60 seconds (keeps connections alive for performance)

### Local Environment Override

**Location:** `cmd/main.go:166-173`

When `Environment == "local"`, HTTP server binds to localhost:

```go
if configuration.Environment == "local" {
    httpServer = &http.Server{
        Addr:         "localhost:" + configuration.SSL_Port,
        Handler:      middlewareHandler,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
}
```

**Rationale:** Local development doesn't need external domain binding or TLS certificates. Uses SSL_Port (e.g., 443) for consistency with production routing.

### HTTP Server Launch

**Location:** `cmd/main.go:208-224`

```go
go func() {
    if !InitStatus.UseSSL && configuration.Environment != "local" {
        utility.DefaultLogger.Info("Server is running at https://localhost:", configuration.SSL_Port)
        err = httpsServer.ListenAndServeTLS(certDir+"localhost.crt", certDir+"localhost.key")
        if err != nil {
            utility.DefaultLogger.Info("Shutting Down Server", err)
            done <- syscall.SIGTERM
        }
    }
    utility.DefaultLogger.Info("Server is running at http://localhost:", configuration.Port)
    err = httpServer.ListenAndServe()
    if err != nil {
        utility.DefaultLogger.Info("Shutting Down Server", err)
        done <- syscall.SIGTERM
    }
}()
```

**Flow:**
1. Try HTTPS first (if SSL certificates exist and not local environment)
2. Fall back to HTTP
3. On error, send shutdown signal to `done` channel

---

## HTTPS Server

### Purpose

The HTTPS server provides:
- Secure TLS-encrypted API and content delivery
- Automatic certificate management via Let's Encrypt
- Production-ready hosting

### Let's Encrypt Integration

**Location:** `cmd/main.go:139-145`

ModulaCMS uses **autocert** from `golang.org/x/crypto/acme/autocert` for automatic certificate provisioning:

```go
manager := autocert.Manager{
    Prompt: autocert.AcceptTOS,
    HostPolicy: autocert.HostWhitelist(
        configuration.Environment_Hosts[configuration.Environment],
        configuration.Client_Site,
        configuration.Admin_Site,
    ),
}
```

**How It Works:**
1. **Prompt:** Automatically accepts Let's Encrypt Terms of Service
2. **HostPolicy:** Whitelist of allowed domains (prevents unauthorized certificate requests)
3. **Certificate Cache:** Autocert caches certificates on disk (default: current directory `.cache`)
4. **Automatic Renewal:** Certificates auto-renew before expiration

**Allowed Domains:**
- Environment-specific host (e.g., `production.example.com`)
- Client site domain (e.g., `www.example.com`)
- Admin site domain (e.g., `admin.example.com`)

### HTTPS Server Creation

**Location:** `cmd/main.go:156-163`

```go
httpsServer = &http.Server{
    Addr:         configuration.Client_Site + configuration.SSL_Port,
    TLSConfig:    manager.TLSConfig(),
    Handler:      middlewareHandler,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

**Key Difference from HTTP:**
- **TLSConfig:** Uses autocert manager's TLS configuration
- **SSL_Port:** Uses secure port (e.g., `:443`)

### Certificate Directory

**Location:** `cmd/main.go:183-186, 287-309`

```go
certDir, err = sanitizeCertDir(configuration.Cert_Dir)
if err != nil {
    utility.DefaultLogger.Fatal("Certificate Directory path is invalid:", err)
}
```

**sanitizeCertDir()** validates:
1. Path is not empty
2. Path is clean (no `../` traversal)
3. Path is absolute
4. Path exists and is a directory

**Certificate Files:**
- `localhost.crt` - TLS certificate for local development
- `localhost.key` - Private key for local development

### HTTPS Launch

Same goroutine as HTTP server (see HTTP Server Launch section). HTTPS is attempted first, with HTTP as fallback.

---

## SSH Server

### Purpose

The SSH server provides:
- Terminal User Interface (TUI) for content management
- Secure remote access for developers and operators
- Interactive database operations
- Real-time content tree navigation

### SSH Framework: Charmbracelet Wish

ModulaCMS uses **Wish** (https://github.com/charmbracelet/wish), a Charmbracelet library for building SSH servers with rich TUIs.

**Dependencies:**
- `github.com/charmbracelet/ssh` - SSH server implementation
- `github.com/charmbracelet/wish` - SSH server framework
- `github.com/charmbracelet/wish/logging` - SSH request logging
- `github.com/charmbracelet/wish/bubbletea` - Bubbletea integration for TUI

### SSH Server Creation

**Location:** `cmd/main.go:127-135`

```go
var host = configuration.SSH_Host
sshServer, err := wish.NewServer(
    wish.WithAddress(net.JoinHostPort(host, configuration.SSH_Port)),
    wish.WithHostKeyPath(".ssh/id_ed25519"),
    wish.WithMiddleware(
        cli.CliMiddleware(app.VerboseFlag, configuration),
        logging.Middleware(),
    ),
)
```

**Configuration:**

1. **Address:** `net.JoinHostPort(host, SSH_Port)`
   - Example: `0.0.0.0:22` or `localhost:2222`
   - Binds to SSH_Host and SSH_Port from configuration

2. **Host Key Path:** `.ssh/id_ed25519`
   - Ed25519 SSH host key for server authentication
   - Clients verify server identity via this key
   - If missing, Wish generates a new key automatically

3. **Middleware Stack:**
   - **cli.CliMiddleware:** Launches Bubbletea TUI application (see below)
   - **logging.Middleware:** Logs SSH connection events

### CLI Middleware

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/middleware.go`

The CLI middleware bridges SSH sessions to Bubbletea TUI programs:

```go
func CliMiddleware(v *bool, c *config.Config) wish.Middleware {
    newProg := func(m tea.Model, opts ...tea.ProgramOption) *tea.Program {
        p := tea.NewProgram(m, opts...)
        go func() {
            for {
                <-time.After(1 * time.Second)
                p.Send(timeMsg(time.Now()))
            }
        }()
        return p
    }
    teaHandler := func(s ssh.Session) *tea.Program {
        pty, _, active := s.Pty()
        if !active {
            wish.Fatalln(s, "no active terminal, skipping")
            return nil
        }
        m, _ := InitialModel(v, c)
        m.Term = pty.Term
        m.Width = pty.Window.Width
        m.Height = pty.Window.Height
        m.Time = time.Now()
        return newProg(&m, append(bubbletea.MakeOptions(s), tea.WithAltScreen())...)
    }
    return bubbletea.MiddlewareWithProgramHandler(teaHandler, termenv.ANSI256)
}
```

**Flow:**
1. **SSH Session Start:** Wish receives SSH connection
2. **PTY Check:** Verify client has active pseudo-terminal
3. **Initialize TUI Model:** Create Bubbletea model with terminal dimensions
4. **Launch Program:** Start Bubbletea program in alternate screen mode
5. **Time Updates:** Send time messages every second (for UI updates)

**Bubbletea Integration:**
- **InitialModel:** Creates the TUI application model (from `internal/cli/`)
- **MakeOptions:** Creates tea.ProgramOptions from SSH session (handles input/output)
- **tea.WithAltScreen():** Uses alternate screen buffer (preserves terminal history)
- **termenv.ANSI256:** Use 256-color ANSI color profile

### SSH Server Launch

**Location:** `cmd/main.go:189-206`

```go
go func() {
    utility.DefaultLogger.Info("Starting SSH server", "ssh "+configuration.SSH_Host+" -p "+configuration.SSH_Port)
    go func() {
        if err = sshServer.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
            utility.DefaultLogger.Error("Could not start server", err)
            done <- nil
        }
    }()

    <-done
    utility.DefaultLogger.Info("Stopping SSH Server")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer func() { cancel() }()
    if err := sshServer.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
        utility.DefaultLogger.Error("Could not stop server", err)
    }
}()
```

**Flow:**
1. **Nested Goroutine:** Inner goroutine runs `sshServer.ListenAndServe()`
2. **Error Handling:** Sends signal to `done` channel if startup fails
3. **Wait for Shutdown:** Outer goroutine waits for `done` signal
4. **Graceful Shutdown:** 30-second timeout for active connections to close
5. **Ignore ErrServerClosed:** Normal shutdown error is ignored

---

## Routing and Middleware

### HTTP/HTTPS Routing

**Router:** `internal/router/mux.go`

ModulaCMS uses **standard library `net/http.ServeMux`** for HTTP routing:

```go
func NewModulacmsMux(c config.Config) *http.ServeMux {
    mux := http.NewServeMux()

    // API endpoints
    mux.HandleFunc("/api/v1/contentdata", ContentDatasHandler)
    mux.HandleFunc("/api/v1/contentdata/", ContentDataHandler)
    // ... many more endpoints

    // Catch-all for content slugs
    mux.HandleFunc("/", SlugHandler)

    return mux
}
```

**Endpoint Categories:**
- `/api/v1/auth/*` - Authentication and OAuth
- `/api/v1/admin*` - Admin-specific API endpoints
- `/api/v1/content*` - Content data and fields
- `/api/v1/datatype*` - Datatype definitions
- `/api/v1/media*` - Media upload and management
- `/api/v1/users*` - User management
- `/` - Catch-all slug handler for content routing

### HTTP Middleware

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware/middleware.go`

The `Serve()` function wraps the router with middleware:

```go
func Serve(next http.Handler, c *config.Config) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        Cors(w, r, c)  // CORS headers

        u, user := AuthRequest(w, r, c)  // Authentication
        if u != nil {
            ctx := context.WithValue(r.Context(), u, user)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }

        // Block unauthenticated API requests
        if strings.Contains(r.URL.Path, "api") {
            w.WriteHeader(http.StatusUnauthorized)
            w.Write([]byte("Unauthorized Request"))
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

**Middleware Chain:**
1. **CORS:** Set CORS headers from configuration
2. **Authentication:** Extract and validate auth cookie
3. **Context Injection:** Add authenticated user to request context
4. **Authorization:** Block unauthenticated API requests
5. **Pass Through:** Allow unauthenticated non-API requests (for public content)

**Usage in main.go:**
```go
mux := router.NewModulacmsMux(*configuration)
middlewareHandler := middleware.Serve(mux, configuration)
```

---

## Configuration

### Server Configuration Fields

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/config/config.go`

```go
type Config struct {
    Environment       string              `json:"environment"`        // "local", "staging", "production"
    Environment_Hosts map[string]string   `json:"environment_hosts"`  // Mapping of env to domain
    Port              string              `json:"port"`               // HTTP port (e.g., "8080")
    SSL_Port          string              `json:"ssl_port"`           // HTTPS port (e.g., "443")
    Cert_Dir          string              `json:"cert_dir"`           // Certificate directory path
    Client_Site       string              `json:"client_site"`        // Client domain
    Admin_Site        string              `json:"admin_site"`         // Admin domain
    SSH_Host          string              `json:"ssh_host"`           // SSH bind address
    SSH_Port          string              `json:"ssh_port"`           // SSH port (e.g., "22")
    // ... other fields
}
```

### Example Configuration

**config.json:**
```json
{
  "environment": "production",
  "environment_hosts": {
    "production": "prod.example.com",
    "staging": "staging.example.com",
    "local": "localhost"
  },
  "port": "8080",
  "ssl_port": "443",
  "cert_dir": "/etc/modulacms/certs/",
  "client_site": "www.example.com",
  "admin_site": "admin.example.com",
  "ssh_host": "0.0.0.0",
  "ssh_port": "2222"
}
```

### Environment-Specific Behavior

**Local Environment:**
- HTTP/HTTPS bind to `localhost` instead of configured domains
- Uses local certificate files (`localhost.crt`, `localhost.key`)
- No Let's Encrypt integration

**Production Environment:**
- HTTP/HTTPS bind to configured domains
- Automatic Let's Encrypt certificates
- Multi-domain support (client site, admin site, environment host)

---

## Graceful Shutdown

### Signal Handling

**Location:** `cmd/main.go:64-65, 226-243`

ModulaCMS listens for OS signals to trigger graceful shutdown:

```go
done := make(chan os.Signal, 1)
signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
```

**Signals:**
- **os.Interrupt:** User presses Ctrl+C
- **syscall.SIGINT:** Interrupt signal (INT)
- **syscall.SIGTERM:** Termination signal (TERM)

### Shutdown Process

```go
<-done  // Block until signal received
utility.DefaultLogger.Info("Shutting down servers...")

// Create 30-second timeout context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Shutdown HTTP server
if err := httpServer.Shutdown(ctx); err != nil {
    utility.DefaultLogger.Error("HTTP server shutdown error:", err)
}

// Shutdown HTTPS server
if err := httpsServer.Shutdown(ctx); err != nil {
    utility.DefaultLogger.Error("HTTPS server shutdown error:", err)
}

// Shutdown SSH server
if err := sshServer.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
    utility.DefaultLogger.Error("SSH server shutdown error:", err)
    return ERRSIG, err
}

utility.DefaultLogger.Info("Servers gracefully stopped.")
return OKSIG, nil
```

**Shutdown Flow:**
1. **Receive Signal:** `done` channel receives OS signal
2. **Create Timeout:** 30-second context for graceful shutdown
3. **Stop Accepting Connections:** All servers stop accepting new connections
4. **Wait for Active Requests:** Servers wait for active requests/sessions to complete
5. **Force Close:** After 30 seconds, forcefully close remaining connections
6. **Exit:** Return success status

**Error Handling:**
- SSH `ErrServerClosed` is ignored (normal shutdown)
- HTTP/HTTPS errors are logged but don't block other shutdowns
- SSH errors trigger `ERRSIG` return code

---

## Port Configuration

### Standard Ports

**HTTP:**
- Development: `8080`
- Production: `80` (requires root/elevated permissions)

**HTTPS:**
- Development: `8443`
- Production: `443` (requires root/elevated permissions)

**SSH:**
- Development: `2222` (non-privileged port)
- Production: `22` (standard SSH port, requires root)

### Running on Privileged Ports

**Option 1: Run as root (not recommended)**
```bash
sudo ./modulacms-x86
```

**Option 2: Use setcap (Linux)**
```bash
sudo setcap 'cap_net_bind_service=+ep' ./modulacms-x86
./modulacms-x86
```

**Option 3: Reverse Proxy**
Use nginx or Caddy to proxy ports 80/443 to higher ports (8080/8443):

```nginx
# nginx example
server {
    listen 80;
    listen 443 ssl;
    server_name example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## TLS Certificate Management

### Let's Encrypt (Production)

**Automatic Certificate Provisioning:**
1. Client makes HTTPS request to domain
2. autocert Manager checks for cached certificate
3. If missing/expired, requests certificate from Let's Encrypt
4. Completes ACME HTTP-01 challenge
5. Receives and caches certificate
6. Serves request with new certificate

**Certificate Storage:**
- Default cache directory: `./.cache/`
- Can be configured via autocert.Manager.Cache field
- Certificates are renewed automatically 30 days before expiration

**Requirements:**
- Domain must resolve to server IP address
- Port 80 must be accessible for ACME challenge
- Server must accept Let's Encrypt Terms of Service

### Local Certificates (Development)

**Self-Signed Certificate Generation:**
```bash
# Create certificate directory
mkdir -p .ssh

# Generate self-signed certificate
openssl req -x509 -newkey rsa:4096 -keyout .ssh/localhost.key -out .ssh/localhost.crt -days 365 -nodes -subj "/CN=localhost"

# Set certificate directory in config
{
  "cert_dir": "./.ssh/",
  ...
}
```

**Certificate Files:**
- `localhost.crt` - Public certificate
- `localhost.key` - Private key

**Browser Warning:**
Browsers will show security warning for self-signed certificates. Click "Advanced" → "Proceed to localhost" to bypass.

---

## SSH Key Management

### Host Key

**Location:** `.ssh/id_ed25519`

The SSH server uses an Ed25519 host key for server authentication:

```bash
# Generate host key (if missing)
ssh-keygen -t ed25519 -f .ssh/id_ed25519 -N ""
```

**Key Fingerprint:**
Clients verify server identity using this fingerprint. On first connection:
```
The authenticity of host '[localhost]:2222 ([127.0.0.1]:2222)' can't be established.
ED25519 key fingerprint is SHA256:abc123...
Are you sure you want to continue connecting (yes/no)?
```

**Key Rotation:**
If host key changes, clients will see "WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!" error. Update `~/.ssh/known_hosts` to resolve.

### Client Authentication

ModulaCMS SSH server currently uses **password-free access** (authentication handled via Wish middleware). For production deployments, consider adding:

**Public Key Authentication:**
```go
wish.WithPublicKeyAuth(func(ctx wish.Context, key wish.PublicKey) bool {
    // Verify public key against authorized keys
    return isAuthorized(key)
})
```

**Password Authentication:**
```go
wish.WithPasswordAuth(func(ctx wish.Context, password string) bool {
    // Verify username/password
    return validateCredentials(ctx.User(), password)
})
```

---

## Connection Flow Examples

### HTTP Request Flow

```
1. Client makes HTTP request to http://example.com:8080/api/v1/contentdata
   ↓
2. HTTP server receives request on port 8080
   ↓
3. Middleware chain executes:
   ├─ CORS headers added
   ├─ Authentication cookie checked
   └─ User context injected (if authenticated)
   ↓
4. Router matches /api/v1/contentdata
   ↓
5. ContentDatasHandler executes
   ↓
6. Response sent to client
```

### HTTPS Request Flow

```
1. Client makes HTTPS request to https://example.com/api/v1/contentdata
   ↓
2. TLS handshake:
   ├─ Client requests certificate
   ├─ autocert Manager provides cached or new certificate
   └─ Encrypted connection established
   ↓
3. HTTPS server receives decrypted request
   ↓
4. Same as HTTP flow (middleware → router → handler)
   ↓
5. Encrypted response sent to client
```

### SSH Connection Flow

```
1. Client connects via SSH: ssh localhost -p 2222
   ↓
2. SSH handshake:
   ├─ Server presents host key (.ssh/id_ed25519)
   ├─ Client verifies fingerprint
   └─ Encrypted session established
   ↓
3. Wish middleware executes:
   ├─ logging.Middleware logs connection
   └─ cli.CliMiddleware launches TUI
   ↓
4. Bubbletea TUI program starts:
   ├─ InitialModel creates application state
   ├─ Terminal dimensions captured
   └─ Alternate screen mode activated
   ↓
5. User interacts with TUI (Model-Update-View cycle)
   ↓
6. User exits TUI (Ctrl+C or quit command)
   ↓
7. SSH session closes gracefully
```

---

## Timeouts and Limits

### HTTP/HTTPS Server Timeouts

**ReadTimeout: 15 seconds**
- Maximum time to read request headers and body
- Prevents slow client attacks (Slowloris)
- Triggers if client doesn't send complete request in 15s

**WriteTimeout: 15 seconds**
- Maximum time to write response
- Prevents slow write attacks
- Triggers if server can't send complete response in 15s

**IdleTimeout: 60 seconds**
- Maximum time between requests on keep-alive connections
- Allows connection reuse for performance
- Closes idle connections after 60s

### Shutdown Timeout

**30-second graceful shutdown:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

- Active requests/sessions have 30 seconds to complete
- After timeout, connections forcefully closed
- Balance between data safety and shutdown speed

---

## Monitoring and Logging

### Server Startup Logs

```
INFO: Starting SSH server ssh 0.0.0.0 -p 2222
INFO: Server is running at http://localhost:8080
```

### Connection Logs (SSH)

```
INFO: New SSH connection from 192.168.1.100:54321
INFO: SSH session started for user: admin
```

### Shutdown Logs

```
INFO: Shutting down servers...
INFO: Stopping SSH Server
INFO: Servers gracefully stopped.
```

### Error Logs

```
ERROR: Could not start server: bind: address already in use
ERROR: HTTP server shutdown error: context deadline exceeded
ERROR: SSH server shutdown error: connection refused
```

### Logging Configuration

**Logger:** `utility.DefaultLogger` (Charmbracelet Log)
- Configured in `internal/utility/`
- Outputs to `debug.log` (configurable via `config.Log_Path`)
- Structured logging with key-value pairs

---

## Common Workflows

### Starting the Server

**Development (local environment):**
```bash
# Start with default config
./modulacms-x86

# Start with custom config
./modulacms-x86 --config=./config.dev.json

# Start in CLI-only mode (no HTTP/HTTPS servers)
./modulacms-x86 --cli
```

**Production:**
```bash
# Run with systemd
sudo systemctl start modulacms

# Run directly (with port permissions)
sudo setcap 'cap_net_bind_service=+ep' ./modulacms-amd64
./modulacms-amd64 --config=/etc/modulacms/config.json
```

### Stopping the Server

**Graceful shutdown (30-second timeout):**
```bash
# Send SIGTERM
kill -TERM <pid>

# Or Ctrl+C if running in foreground
^C
```

**Force shutdown (immediate):**
```bash
# Send SIGKILL
kill -9 <pid>
```

### Connecting via SSH

```bash
# Connect to SSH server
ssh localhost -p 2222

# With specific user (if authentication enabled)
ssh admin@example.com -p 22

# First connection (verify fingerprint)
The authenticity of host '[localhost]:2222 ([127.0.0.1]:2222)' can't be established.
ED25519 key fingerprint is SHA256:abc123...
Are you sure you want to continue connecting (yes/no)? yes
```

### Testing HTTPS Locally

```bash
# Generate self-signed certificate
openssl req -x509 -newkey rsa:4096 -keyout .ssh/localhost.key -out .ssh/localhost.crt -days 365 -nodes -subj "/CN=localhost"

# Update config.json
{
  "environment": "local",
  "cert_dir": "./.ssh/",
  ...
}

# Start server
./modulacms-x86

# Test with curl (ignore self-signed warning)
curl -k https://localhost:443/api/v1/contentdata
```

---

## Troubleshooting

### Port Already in Use

**Error:**
```
ERROR: Could not start server: bind: address already in use
```

**Solution:**
```bash
# Find process using port
sudo lsof -i :8080  # HTTP
sudo lsof -i :443   # HTTPS
sudo lsof -i :2222  # SSH

# Kill process
kill -9 <pid>

# Or use different port in config.json
```

### Permission Denied on Privileged Ports

**Error:**
```
ERROR: Could not start server: bind: permission denied
```

**Solution:**
```bash
# Option 1: Use setcap (Linux)
sudo setcap 'cap_net_bind_service=+ep' ./modulacms-x86

# Option 2: Use non-privileged ports (>1024)
{
  "port": "8080",
  "ssl_port": "8443",
  "ssh_port": "2222"
}

# Option 3: Run as root (not recommended)
sudo ./modulacms-x86
```

### Let's Encrypt Certificate Failure

**Error:**
```
ERROR: autocert: unable to satisfy acme challenge
```

**Causes:**
1. Domain doesn't resolve to server IP
2. Port 80 not accessible (firewall blocking)
3. Rate limit exceeded (5 certs/week per domain)
4. Domain not in HostWhitelist

**Solution:**
```bash
# Verify DNS resolution
dig example.com  # Should match server IP

# Test port 80 accessibility
curl http://example.com/.well-known/acme-challenge/test

# Check firewall rules
sudo ufw status
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Wait if rate limited (check: https://crt.sh/)
```

### SSH Host Key Warning

**Error:**
```
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
```

**Cause:** SSH host key changed (`.ssh/id_ed25519` regenerated)

**Solution:**
```bash
# Remove old host key from known_hosts
ssh-keygen -R "[localhost]:2222"

# Reconnect and verify new fingerprint
ssh localhost -p 2222
```

### Graceful Shutdown Timeout

**Error:**
```
ERROR: HTTP server shutdown error: context deadline exceeded
```

**Cause:** Active requests/connections didn't complete within 30 seconds

**Solution:**
- Increase shutdown timeout in `cmd/main.go`:
  ```go
  ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
  ```
- Or force shutdown with SIGKILL: `kill -9 <pid>`

### TUI Rendering Issues

**Problem:** Garbled text or incorrect layout in SSH TUI

**Causes:**
1. Terminal doesn't support alternate screen
2. Terminal size not detected correctly
3. Color profile incompatibility

**Solution:**
```bash
# Check terminal capabilities
echo $TERM  # Should be xterm-256color or similar

# Force terminal type
TERM=xterm-256color ssh localhost -p 2222

# Resize detection (press Ctrl+L to refresh)
```

---

## Performance Considerations

### Connection Limits

**HTTP/HTTPS:**
- No explicit connection limit (OS default)
- `IdleTimeout: 60s` closes idle connections
- Consider using reverse proxy (nginx, Caddy) for production load balancing

**SSH:**
- Wish handles multiple concurrent sessions
- Each session = separate Bubbletea program instance
- Memory usage scales with concurrent sessions

### Timeout Tuning

**Short Timeouts (current: 15s):**
- Pros: Better protection against slow attacks, faster error detection
- Cons: May timeout legitimate slow clients/responses

**Long Timeouts (30s+):**
- Pros: More tolerant of slow networks, large uploads
- Cons: More vulnerable to resource exhaustion attacks

**Recommendation:**
- Keep default 15s for most deployments
- Increase WriteTimeout if serving large files/responses
- Increase ReadTimeout if accepting large uploads

### TLS Performance

**Autocert Caching:**
- Cached certificates served instantly
- No Let's Encrypt API call on cached cert
- Cache persists across restarts

**TLS Handshake Overhead:**
- ~1-2ms per connection
- Mitigated by connection keep-alive (IdleTimeout: 60s)
- Consider TLS session resumption for high-traffic sites

---

## Security Considerations

### TLS/HTTPS

**Strengths:**
- Automatic certificate management (no expired certs)
- Modern TLS configuration via autocert
- Host whitelist prevents unauthorized certificate requests

**Weaknesses:**
- HTTP port remains open (no redirect to HTTPS)
- Consider adding HTTP → HTTPS redirect middleware

**Recommendation:**
```go
// Add to middleware.Serve()
if r.TLS == nil && configuration.Environment != "local" {
    httpsUrl := "https://" + r.Host + r.RequestURI
    http.Redirect(w, r, httpsUrl, http.StatusMovedPermanently)
    return
}
```

### SSH

**Strengths:**
- Ed25519 host key (modern, secure algorithm)
- Encrypted session transport

**Weaknesses:**
- No authentication required (open access)
- No authorization/role-based access control

**Recommendation:**
```go
// Add public key authentication
wish.WithPublicKeyAuth(func(ctx wish.Context, key wish.PublicKey) bool {
    authorizedKeys := loadAuthorizedKeys("./authorized_keys")
    return isKeyAuthorized(key, authorizedKeys)
})
```

### Port Exposure

**Current Setup:**
- All three servers bind to configured host (e.g., `0.0.0.0`)
- Exposed to external network

**Hardening:**
```json
{
  "ssh_host": "127.0.0.1",  // SSH only on localhost
  "port": "8080",           // HTTP behind reverse proxy
  "ssl_port": "443"         // HTTPS public
}
```

### Rate Limiting

**Missing:** No built-in rate limiting on HTTP/HTTPS endpoints

**Recommendation:**
- Add rate limiting middleware
- Or use reverse proxy (nginx limit_req, Caddy rate_limit)

---

## Related Documentation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TUI_ARCHITECTURE.md` - Bubbletea/Elm Architecture details
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/DATABASE_LAYER.md` - Database connection management

**Packages:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/CLI_PACKAGE.md` - TUI implementation details
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/MIDDLEWARE_PACKAGE.md` - HTTP middleware patterns

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/DEPLOYMENT.md` - Production deployment guide
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/DEBUGGING.md` - Debugging server issues

**Reference:**
- `/Users/home/Documents/Code/Go_dev/modulacms/CLAUDE.md` - Build commands and development guidelines
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/FILE_TREE.md` - Project structure

---

## Quick Reference

### Key Files

```
cmd/main.go                         - Server initialization and launch
internal/router/mux.go              - HTTP routing
internal/middleware/middleware.go   - HTTP middleware
internal/cli/middleware.go          - SSH/TUI middleware
internal/config/config.go           - Configuration structure
```

### Server Ports

| Server | Development | Production |
|--------|-------------|------------|
| HTTP   | 8080        | 80         |
| HTTPS  | 8443        | 443        |
| SSH    | 2222        | 22         |

### Server Lifecycle Commands

```bash
# Start
./modulacms-x86
./modulacms-x86 --config=./config.json
./modulacms-x86 --cli  # CLI-only mode

# Stop
kill -TERM <pid>       # Graceful (30s timeout)
kill -9 <pid>          # Force

# Connect
curl http://localhost:8080/api/v1/contentdata
curl -k https://localhost:8443/api/v1/contentdata
ssh localhost -p 2222
```

### Configuration Keys

```json
{
  "port": "8080",              // HTTP port
  "ssl_port": "443",           // HTTPS port
  "cert_dir": "/path/to/certs/",
  "client_site": "example.com",
  "admin_site": "admin.example.com",
  "ssh_host": "0.0.0.0",       // SSH bind address
  "ssh_port": "2222",          // SSH port
  "environment": "production"  // "local", "staging", "production"
}
```

### Timeouts

- **ReadTimeout:** 15 seconds (request reading)
- **WriteTimeout:** 15 seconds (response writing)
- **IdleTimeout:** 60 seconds (keep-alive connections)
- **ShutdownTimeout:** 30 seconds (graceful shutdown)

---

**Last Updated:** 2026-01-12
**Status:** Complete
