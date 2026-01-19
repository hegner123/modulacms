# main.go Improvements

## Critical Issues to Fix

### 1. Ignored Errors

**Line 95** - Ignoring config error:
```go
// BAD
configuration, _ := configManager.Config()

// GOOD
configuration, err := configManager.Config()
if err != nil {
    utility.DefaultLogger.Fatal("Failed to get configuration", err)
}
```

**Line 135** - Ignoring InitStatus:
```go
// BAD
_, err := install.CheckInstall(configuration, app.VerboseFlag)

// GOOD
InitStatus, err := install.CheckInstall(configuration, app.VerboseFlag)
```

**Line 173** - Ignoring all database errors:
```go
// BAD
databaseConnection, _, _ := db.ConfigDB(*configuration).GetConnection()

// GOOD
databaseConnection, dbDriver, err := db.ConfigDB(*configuration).GetConnection()
if err != nil {
    utility.DefaultLogger.Fatal("Failed to connect to database", err)
}
utility.DefaultLogger.Info("Database connected", "driver", dbDriver)
```

**Line 178** - SSH server error not checked:
```go
// BAD
sshServer, err := wish.NewServer(...)
// err never checked!

// GOOD
sshServer, err := wish.NewServer(...)
if err != nil {
    utility.DefaultLogger.Fatal("Failed to create SSH server", err)
}
```

### 2. Duplicate HTTP Server Configuration

**Lines 214-247** - Hardcoded timeout values repeated:
```go
// BAD - Duplicated configuration
httpServer = &http.Server{
    Addr:         configuration.Client_Site + configuration.Port,
    Handler:      middlewareHandler,
    ReadTimeout:  15 * time.Second,  // Repeated
    WriteTimeout: 15 * time.Second,  // Repeated
    IdleTimeout:  60 * time.Second,  // Repeated
}
// Then duplicated again for local environment

// GOOD - Extract to function
func newHTTPServer(addr string, handler http.Handler) *http.Server {
    return &http.Server{
        Addr:         addr,
        Handler:      handler,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
}

// Usage
httpServer := newHTTPServer(configuration.Client_Site+configuration.Port, middlewareHandler)
if configuration.Environment == "local" {
    httpServer.Addr = "localhost" + configuration.Port
}
```

### 3. Double Nested Goroutine (SSH Server)

**Lines 255-272** - Goroutine inside goroutine:
```go
// BAD - Nested goroutines
go func() {
    utility.DefaultLogger.Info("Starting SSH server", ...)
    go func() {  // <-- NESTED!
        if err = sshServer.ListenAndServe(); ...
    }()
    <-done
    // shutdown...
}()

// GOOD - Single goroutine
go func() {
    utility.DefaultLogger.Info("Starting SSH server",
        "address", net.JoinHostPort(configuration.SSH_Host, configuration.SSH_Port))

    if err := sshServer.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
        utility.DefaultLogger.Error("SSH server error", err)
        done <- syscall.SIGTERM
    }
}()
```

### 4. Shared Error Variable in Goroutines

**Lines 259, 278, 288** - All goroutines share `err`:
```go
// BAD - Race condition
go func() {
    if err = sshServer.ListenAndServe(); ... // Writes to err
}()
go func() {
    err = httpsServer.ListenAndServeTLS(...) // Writes to err
}()

// GOOD - Local error variables
go func() {
    if sshErr := sshServer.ListenAndServe(); sshErr != nil && !errors.Is(sshErr, ssh.ErrServerClosed) {
        utility.DefaultLogger.Error("SSH server error", sshErr)
        done <- syscall.SIGTERM
    }
}()
```

### 5. Empty/Placeholder Flag Handlers

**Line 322-324** - Empty handler:
```go
func HandleFlagAuth(c config.Config) {
    os.Exit(0)  // Does nothing!
}
```

**Line 326-331** - Wrong error message:
```go
func HandleFlagAlpha() {
    _, err := os.Open("test.txt")
    if err != nil {
        utility.DefaultLogger.Fatal("failed to create database dump in archive: ", err)
        // ^ Wrong message for opening test.txt
    }
}
```

**Fix:** Remove these or implement properly.

### 6. Context Misuse in SSH Shutdown

**Line 267** - Creating new context instead of using existing:
```go
// BAD - Creates new context
<-done
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer func() { cancel() }()
if err := sshServer.Shutdown(ctx); ...

// GOOD - Use the shutdown context
// SSH shutdown happens automatically when rootCtx is cancelled
// No need for separate shutdown goroutine
```

### 7. Inconsistent Signal Handling

**Line 261** - Sending `nil` to signal channel:
```go
// BAD
done <- nil  // nil is not a valid os.Signal

// GOOD
done <- syscall.SIGTERM
```

### 8. Flag Handling Pattern

**Lines 80-170** - Scattered flag handling:
```go
// BAD - Mixed together
if *app.VersionFlag { HandleFlagVersion() }
if *app.GenCertsFlag { HandleFlagGenCerts() }
// ... config loading ...
if *app.UpdateFlag { HandleFlagUpdate(updateUrl) }
if *app.AuthFlag { HandleFlagAuth(*configuration) }

// GOOD - Group early-exit flags, then config-dependent flags
// Early exits (no config needed)
if *app.VersionFlag {
    HandleFlagVersion()
}
if *app.GenCertsFlag {
    HandleFlagGenCerts()
}

// Load config
configManager.Load()
configuration, _ := configManager.Config()

// Config-dependent flags
if *app.UpdateFlag {
    HandleFlagUpdate(updateUrl)
}
if *app.InstallFlag {
    install.RunInstall(app.VerboseFlag)
}
// ... continue with server startup
```

## Recommended Refactoring

### Extract Server Startup

```go
type Servers struct {
    SSH   *ssh.Server
    HTTP  *http.Server
    HTTPS *http.Server
}

func (s *Servers) Start(ctx context.Context, config *config.Config) error {
    // Start SSH
    go func() {
        utility.DefaultLogger.Info("Starting SSH server")
        if err := s.SSH.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
            utility.DefaultLogger.Error("SSH server error", err)
        }
    }()

    // Start HTTP
    go func() {
        utility.DefaultLogger.Info("Starting HTTP server", "addr", s.HTTP.Addr)
        if err := s.HTTP.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            utility.DefaultLogger.Error("HTTP server error", err)
        }
    }()

    // Start HTTPS if configured
    if config.Environment != "http-only" {
        go func() {
            utility.DefaultLogger.Info("Starting HTTPS server", "addr", s.HTTPS.Addr)
            if err := s.HTTPS.ListenAndServeTLS(...); err != nil && err != http.ErrServerClosed {
                utility.DefaultLogger.Error("HTTPS server error", err)
            }
        }()
    }

    return nil
}

func (s *Servers) Shutdown(ctx context.Context) error {
    // Shutdown all servers concurrently
    errChan := make(chan error, 3)
    var wg sync.WaitGroup

    wg.Add(3)
    go func() {
        defer wg.Done()
        if err := s.HTTP.Shutdown(ctx); err != nil {
            errChan <- fmt.Errorf("HTTP shutdown: %w", err)
        }
    }()
    go func() {
        defer wg.Done()
        if err := s.HTTPS.Shutdown(ctx); err != nil {
            errChan <- fmt.Errorf("HTTPS shutdown: %w", err)
        }
    }()
    go func() {
        defer wg.Done()
        if err := s.SSH.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
            errChan <- fmt.Errorf("SSH shutdown: %w", err)
        }
    }()

    wg.Wait()
    close(errChan)

    for err := range errChan {
        utility.DefaultLogger.Error("Shutdown error", err)
    }

    return nil
}
```

### Simplified run() Function Structure

```go
func run() (ReturnCode, error) {
    // 1. Setup
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 2. Parse flags and handle early exits
    app := flags.ParseFlags()
    if err := handleEarlyExitFlags(app); err != nil {
        return OKSIG, nil // Version/GenCerts exit normally
    }

    // 3. Load configuration
    config, err := loadConfiguration(app.ConfigPath)
    if err != nil {
        return ERRSIG, err
    }

    // 4. Initialize services
    if err := initializeServices(ctx, config); err != nil {
        return ERRSIG, err
    }

    // 5. Handle config-dependent flags
    if err := handleConfigFlags(app, config); err != nil {
        return ERRSIG, err
    }

    // 6. Create and start servers
    servers, err := createServers(config)
    if err != nil {
        return ERRSIG, err
    }

    if err := servers.Start(ctx, config); err != nil {
        return ERRSIG, err
    }

    // 7. Wait for shutdown signal
    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
    <-done

    // 8. Graceful shutdown
    utility.DefaultLogger.Info("Shutting down...")
    cancel() // Stop background workers

    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()

    if err := servers.Shutdown(shutdownCtx); err != nil {
        return ERRSIG, err
    }

    utility.DefaultLogger.Info("Shutdown complete")
    return OKSIG, nil
}
```

## Quick Wins (Fix These First)

1. **Check all errors** - Lines 95, 135, 173, 178
2. **Fix SSH goroutine nesting** - Lines 255-272
3. **Use local error variables** - Lines 259, 278, 288
4. **Remove/implement empty handlers** - Lines 322-331
5. **Extract server timeout constants** - Lines 214-247

## Medium Priority

6. **Refactor server creation** - Extract to builder function
7. **Consolidate flag handling** - Group by dependency
8. **Add panic recovery** - Wrap goroutines
9. **Use errgroup for coordinated shutdown**

## Low Priority

10. **Extract to separate files** - `cmd/server.go`, `cmd/flags.go`
11. **Add startup timing metrics**
12. **Implement health checks**

## Performance Issues

### Hardcoded Path in ResetFlag
**Line 162** - Hardcoded database path:
```go
// BAD
err := os.Remove("./modula.db")

// GOOD
err := os.Remove(configuration.Db_URL)
```

### Unused Variable InitStatus
**Line 60** - Declared but never properly used:
```go
// BAD
var InitStatus install.ModulaInit  // Declared
_, err := install.CheckInstall(...) // Ignored

// GOOD
InitStatus, err := install.CheckInstall(configuration, app.VerboseFlag)
if err != nil {
    utility.DefaultLogger.Error("Install check failed", err)
}
```

## Security Concerns

### 1. SSH Host Key Path Hardcoded
**Line 180** - Hardcoded path:
```go
// BAD
wish.WithHostKeyPath(".ssh/id_ed25519")

// GOOD - Make it configurable
wish.WithHostKeyPath(configuration.SSH_Host_Key_Path)
```

### 2. Certificate Paths Not Validated
**Line 278-280** - No validation before use:
```go
// Add validation
if !utility.FileExists(filepath.Join(certDir, "localhost.crt")) {
    utility.DefaultLogger.Fatal("Certificate file not found", nil)
}
```

## Example: Fixed Server Creation

```go
func createHTTPServers(config *config.Config, handler http.Handler) (*http.Server, *http.Server, error) {
    newServer := func(addr string, tlsConfig *tls.Config) *http.Server {
        return &http.Server{
            Addr:         addr,
            Handler:      handler,
            TLSConfig:    tlsConfig,
            ReadTimeout:  15 * time.Second,
            WriteTimeout: 15 * time.Second,
            IdleTimeout:  60 * time.Second,
        }
    }

    host := config.Client_Site
    if config.Environment == "local" {
        host = "localhost"
    }

    manager := autocert.Manager{
        Prompt:     autocert.AcceptTOS,
        HostPolicy: autocert.HostWhitelist(
            config.Environment_Hosts[config.Environment],
            config.Client_Site,
            config.Admin_Site,
        ),
    }

    httpServer := newServer(host+config.Port, nil)
    httpsServer := newServer(host+config.SSL_Port, manager.TLSConfig())

    return httpServer, httpsServer, nil
}
```

## Summary of Critical Fixes

| Line | Issue | Priority | Fix |
|------|-------|----------|-----|
| 95 | Ignored config error | **Critical** | Check error |
| 135 | Ignored InitStatus | **Critical** | Use return value |
| 173 | Ignored DB errors | **Critical** | Check all 3 returns |
| 178 | Ignored SSH server error | **Critical** | Check before use |
| 255-272 | Nested goroutines | **High** | Flatten to single goroutine |
| 259,278,288 | Shared error variable | **High** | Use local vars |
| 214-247 | Duplicate config | **Medium** | Extract to function |
| 322-331 | Empty handlers | **Medium** | Remove or implement |

Fix these in order, test after each fix.
