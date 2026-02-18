package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/bucket"
	"github.com/hegner123/modulacms/internal/cli"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/install"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/plugin"
	"github.com/hegner123/modulacms/internal/router"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/acme/autocert"
)

// wizard indicates whether to run the interactive configuration wizard before starting.
var wizard bool

func init() {
	serveCmd.Flags().BoolVar(&wizard, "wizard", false, "Run interactive configuration wizard before starting")
}

// serveCmd starts the HTTP, HTTPS, and SSH servers for ModulaCMS.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP, HTTPS, and SSH servers",
	Long:  "Start all ModulaCMS servers. Use --wizard for interactive setup.",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		rootCtx, rootCancel := context.WithCancel(context.Background())
		defer rootCancel()

		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		// Second signal forces immediate exit
		go func() {
			<-done
			utility.DefaultLogger.Info("Shutting down servers...")
			// Drain done so the main select sees it
			select {
			case done <- syscall.SIGTERM:
			default:
			}
		}()

		// Setup: wizard mode runs interactive install, otherwise auto-create defaults if config missing
		if wizard {
			if err := install.RunInstall(&verbose, nil, nil); err != nil {
				return fmt.Errorf("wizard setup failed: %w", err)
			}
		} else {
			if _, statErr := os.Stat(cfgPath); errors.Is(statErr, os.ErrNotExist) {
				utility.DefaultLogger.Info("No config found, creating defaults...", "path", cfgPath)
				yes := true
				autoPassword, err := utility.MakeRandomString()
				if err != nil {
					return fmt.Errorf("failed to generate admin password: %w", err)
				}
				utility.DefaultLogger.Finfo("Generated system admin password", "email", "system@modulacms.local", "password", autoPassword)
				if err := install.RunInstall(&verbose, &yes, &autoPassword); err != nil {
					return fmt.Errorf("auto-setup failed: %w", err)
				}
			}
		}

		mgr, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		cfg, err := mgr.Config()
		if err != nil {
			return err
		}

		// Initialize isolated plugin database pool
		pluginPool, pluginPoolCleanup, err := initPluginPool(cfg)
		if err != nil {
			return fmt.Errorf("plugin pool init failed: %w", err)
		}
		defer pluginPoolCleanup()
		pluginManager := initPluginManager(rootCtx, cfg, pluginPool, driver)

		if cfg.Node_ID == "" {
			cfg.Node_ID = string(types.NewNodeID())
			utility.DefaultLogger.Finfo("No node_id configured, generated ephemeral node ID", "node_id", cfg.Node_ID)
		}

		logConfigSummary(cfg)

		// Start observability
		obsCleanup := initObservability(rootCtx, cfg)
		defer obsCleanup()

		// Start email service
		emailSvc, err := initEmailService(cfg)
		if err != nil {
			return fmt.Errorf("email: %w", err)
		}
		defer emailSvc.Close()
		mgr.OnChange(func(newCfg config.Config) {
			if err := emailSvc.Reload(newCfg); err != nil {
				utility.DefaultLogger.Error("email service hot-reload failed", err)
			}
		})

		// Run install check — if DB isn't connected, attempt table creation and bootstrap
		initStatus, err := install.CheckInstall(cfg, &verbose)
		if err != nil {
			utility.DefaultLogger.Warn("Installation check reported issues, attempting DB setup", err)
			fallbackPassword, pwErr := utility.MakeRandomString()
			if pwErr != nil {
				return fmt.Errorf("failed to generate admin password: %w", pwErr)
			}
			fallbackHash, pwErr := auth.HashPassword(fallbackPassword)
			if pwErr != nil {
				return fmt.Errorf("failed to hash admin password: %w", pwErr)
			}
			utility.DefaultLogger.Finfo("Generated system admin password for DB setup", "email", "system@modulacms.local", "password", fallbackPassword)
			if setupErr := install.CreateDbSimple(cfgPath, cfg, fallbackHash); setupErr != nil {
				return fmt.Errorf("database setup failed: %w", setupErr)
			}
		}

		// dbReadyCh: TUI signals this after DB init so we can reload the
		// permission cache and start HTTP/HTTPS servers.
		dbReadyCh := make(chan struct{}, 1)

		// SSH server — started before permission cache so it remains available
		// even when the HTTP stack cannot initialize (e.g. missing bootstrap data).
		sshServer, err := wish.NewServer(
			wish.WithAddress(net.JoinHostPort(cfg.SSH_Host, cfg.SSH_Port)),
			wish.WithHostKeyPath(".ssh/id_ed25519"),
			wish.WithPublicKeyAuth(middleware.PublicKeyHandler(cfg)),
			wish.WithMiddleware(
				middleware.SSHSessionLoggingMiddleware(cfg),
				middleware.SSHAuthenticationMiddleware(cfg),
				middleware.SSHAuthorizationMiddleware(cfg),
				cli.CliMiddleware(&verbose, cfg, driver, utility.DefaultLogger, pluginManager, mgr, dbReadyCh),
				logging.Middleware(),
			),
		)
		if err != nil {
			return err
		}

		go func() {
			utility.DefaultLogger.Info("Starting SSH server",
				"address", net.JoinHostPort(cfg.SSH_Host, cfg.SSH_Port))

			if sshErr := sshServer.ListenAndServe(); sshErr != nil && !errors.Is(sshErr, ssh.ErrServerClosed) {
				utility.DefaultLogger.Error("SSH server error", sshErr)
				done <- syscall.SIGTERM
			}
		}()

		// Initialize permission cache after install check (bootstrap data must exist).
		// On failure the SSH server keeps running so operators can diagnose via TUI;
		// HTTP starts with a placeholder handler until DB is initialized.
		sshOnly := false
		utility.DefaultLogger.Info("Loading permission cache...")
		pc := middleware.NewPermissionCache()
		if pcErr := pc.Load(driver); pcErr != nil {
			utility.DefaultLogger.Error("Permission cache load failed — placeholder HTTP mode until DB init via SSH", pcErr)
			sshOnly = true
		} else {
			utility.DefaultLogger.Info("Permission cache loaded")
			pc.StartPeriodicRefresh(rootCtx, driver, 60*time.Second)
		}

		var bridge *plugin.HTTPBridge
		var tokenMu sync.Mutex
		var pluginAPITokenID string
		var pluginAPITokenPath string

		if pluginManager != nil {
			bridge = pluginManager.Bridge()
		}

		// buildRealHandler creates the full router + middleware stack.
		buildRealHandler := func() http.Handler {
			// Ensure S3 buckets exist (media + backup)
			if cfg.BucketEndpointURL() != "" {
				s3Creds := bucket.S3Credentials{
					AccessKey:      cfg.Bucket_Access_Key,
					SecretKey:      cfg.Bucket_Secret_Key,
					URL:            cfg.BucketEndpointURL(),
					Region:         cfg.Bucket_Region,
					ForcePathStyle: cfg.Bucket_Force_Path_Style,
				}
				s3Session, s3Err := s3Creds.GetBucket()
				if s3Err != nil {
					utility.DefaultLogger.Warn("S3 connection failed, media uploads will be unavailable", s3Err)
				} else {
					if cfg.Bucket_Media != "" {
						if bErr := bucket.EnsureBucket(s3Session, cfg.Bucket_Media); bErr != nil {
							utility.DefaultLogger.Warn("Failed to ensure media bucket", bErr, "bucket", cfg.Bucket_Media)
						} else if pErr := bucket.SetPublicReadPolicy(s3Session, cfg.Bucket_Media); pErr != nil {
							utility.DefaultLogger.Warn("Failed to set public-read policy on media bucket", pErr, "bucket", cfg.Bucket_Media)
						}
					}
					if cfg.Bucket_Backup != "" {
						if bErr := bucket.EnsureBucket(s3Session, cfg.Bucket_Backup); bErr != nil {
							utility.DefaultLogger.Warn("Failed to ensure backup bucket", bErr, "bucket", cfg.Bucket_Backup)
						}
					}
				}
			}

			mux := router.NewModulacmsMux(mgr, bridge, driver, pc)

			var hookRunner audited.HookRunner
			if pluginManager != nil {
				hookRunner = pluginManager.HookEngine()
			}
			return middleware.Chain(
				middleware.HookRunnerMiddleware(hookRunner),
			)(middleware.DefaultMiddlewareChain(mgr, pc)(mux))
		}

		// Swappable handler: starts as placeholder when DB is missing,
		// swapped to the real handler after TUI DB init.
		handler := &handlerSwap{}
		if sshOnly {
			sshAddr := net.JoinHostPort(cfg.SSH_Host, cfg.SSH_Port)
			handler.set(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintf(w,
					`{"error":"database not initialized","message":"SSH into %s to initialize the database and set up your API key"}`,
					sshAddr,
				)
			}))
			utility.DefaultLogger.Info("HTTP serving placeholder (DB not initialized)",
				"ssh", sshAddr)
		} else {
			handler.set(buildRealHandler())
		}

		manager := autocert.Manager{
			Prompt: autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(
				cfg.Environment_Hosts[cfg.Environment],
				cfg.Client_Site,
				cfg.Admin_Site,
			),
		}

		// Determine HTTP host based on environment
		httpHost := cfg.Client_Site
		switch cfg.Environment {
		case "local":
			httpHost = "localhost"
		case "docker":
			httpHost = "0.0.0.0"
		}

		// TLS config: use autocert for production, nil for local/docker (self-signed)
		var tlsConfig *tls.Config
		if cfg.Environment != "local" && cfg.Environment != "docker" {
			tlsConfig = manager.TLSConfig()
		}

		httpServer := newHTTPServer(httpHost+cfg.Port, handler, nil)
		httpsServer := newHTTPServer(httpHost+cfg.SSL_Port, handler, tlsConfig)

		certDir, certErr := sanitizeCertDir(cfg.Cert_Dir)
		if certErr != nil {
			return certErr
		}

		// Start HTTPS server if configured
		if !initStatus.UseSSL && cfg.Environment != "http-only" {
			go func() {
				utility.DefaultLogger.Info("Starting HTTPS server", "address", httpsServer.Addr)
				httpsErr := httpsServer.ListenAndServeTLS(
					filepath.Join(certDir, "localhost.crt"),
					filepath.Join(certDir, "localhost.key"),
				)
				if httpsErr != nil && !errors.Is(httpsErr, http.ErrServerClosed) {
					utility.DefaultLogger.Error("HTTPS server error", httpsErr)
					done <- syscall.SIGTERM
				}
			}()
		}

		// Start HTTP server
		go func() {
			utility.DefaultLogger.Info("Starting HTTP server", "address", httpServer.Addr)
			httpErr := httpServer.ListenAndServe()
			if httpErr != nil && !errors.Is(httpErr, http.ErrServerClosed) {
				utility.DefaultLogger.Error("HTTP server error", httpErr)
				done <- syscall.SIGTERM
			}
		}()

		// Generate admin API token for CLI plugin commands.
		if !sshOnly && cfg.Plugin_Enabled {
			tokenID, tokenPath, tokenErr := generatePluginAPIToken(rootCtx, driver, cfg.Node_ID)
			if tokenErr != nil {
				utility.DefaultLogger.Error("failed to generate plugin API token", tokenErr)
			} else {
				tokenMu.Lock()
				pluginAPITokenID = tokenID
				pluginAPITokenPath = tokenPath
				tokenMu.Unlock()
				utility.DefaultLogger.Info("plugin API token created", "path", tokenPath)
			}
		}

		// When DB was not ready at startup, wait for TUI to signal DB init
		// then reload the permission cache and swap to the real handler.
		if sshOnly {
			go func() {
				select {
				case <-dbReadyCh:
					utility.DefaultLogger.Info("DB initialized via TUI, reloading permission cache...")
					if pcErr := pc.Load(driver); pcErr != nil {
						utility.DefaultLogger.Error("Permission cache reload after DB init failed", pcErr)
						return
					}
					utility.DefaultLogger.Info("Permission cache loaded, switching to real HTTP handler")
					pc.StartPeriodicRefresh(rootCtx, driver, 60*time.Second)
					handler.set(buildRealHandler())

					if cfg.Plugin_Enabled {
						tokenID, tokenPath, tokenErr := generatePluginAPIToken(rootCtx, driver, cfg.Node_ID)
						if tokenErr != nil {
							utility.DefaultLogger.Error("failed to generate plugin API token", tokenErr)
						} else {
							tokenMu.Lock()
							pluginAPITokenID = tokenID
							pluginAPITokenPath = tokenPath
							tokenMu.Unlock()
							utility.DefaultLogger.Info("plugin API token created", "path", tokenPath)
						}
					}
				case <-rootCtx.Done():
				}
			}()
		}

		// Wait for shutdown signal
		<-done
		utility.DefaultLogger.Info("Shutting down servers...")

		rootCancel()

		// Graceful shutdown with 30s timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			utility.DefaultLogger.Error("HTTP server shutdown error:", err)
		}

		if err := httpsServer.Shutdown(shutdownCtx); err != nil {
			utility.DefaultLogger.Error("HTTPS server shutdown error:", err)
		}

		if err := sshServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			utility.DefaultLogger.Error("SSH server shutdown error:", err)
			return err
		}

		// Clean up the plugin API token before shutting down the plugin
		// subsystems. The DB is still open at this point, so the token row
		// can be deleted cleanly.
		tokenMu.Lock()
		tokID := pluginAPITokenID
		tokPath := pluginAPITokenPath
		tokenMu.Unlock()
		if tokID != "" {
			cleanupPluginAPIToken(shutdownCtx, driver, tokID, tokPath, cfg.Node_ID)
		}

		// Shut down plugin subsystems after HTTP servers are drained so
		// in-flight plugin requests can complete.
		// S11 shutdown order: StopWatcher (stop file polling) -> bridge
		// (stop accepting new plugin HTTP requests) -> hook engine (drain
		// after-hooks) -> manager (close VM pools and DB).
		if pluginManager != nil {
			pluginManager.StopWatcher()
		}
		if bridge != nil {
			bridge.Close(shutdownCtx)
		}
		if pluginManager != nil {
			if hookEngine := pluginManager.HookEngine(); hookEngine != nil {
				hookEngine.Close(shutdownCtx)
			}
			pluginManager.Shutdown(shutdownCtx)
		}

		utility.DefaultLogger.Info("Servers gracefully stopped.")
		return nil
	},
}

// generatePluginAPIToken creates a short-lived API token for CLI plugin commands.
// It looks up the system user, cleans up any stale tokens from previous runs,
// generates a new token, inserts it into the database, and writes it to a file.
// Returns the token row ID and file path on success.
func generatePluginAPIToken(ctx context.Context, driver db.DbDriver, nodeID string) (string, string, error) {
	// Look up the system user (same pattern as internal/cli/model.go:218-224).
	users, err := driver.ListUsers()
	if err != nil {
		return "", "", fmt.Errorf("list users for plugin API token: %w", err)
	}
	if users == nil {
		return "", "", fmt.Errorf("no users found for plugin API token")
	}

	var systemUserID types.UserID
	var found bool
	for _, u := range *users {
		if strings.EqualFold(u.Username, "system") {
			systemUserID = u.UserID
			found = true
			break
		}
	}
	if !found {
		return "", "", fmt.Errorf("system user not found for plugin API token")
	}

	auditCtx := audited.Ctx(types.NodeID(nodeID), systemUserID, "plugin-api-token-init", "127.0.0.1")
	systemNullableID := types.NullableUserID{ID: systemUserID, Valid: true}

	// Clean up stale api_key tokens from previous ungraceful shutdowns.
	existingTokens, err := driver.GetTokenByUserId(systemNullableID)
	if err != nil {
		utility.DefaultLogger.Warn("failed to query existing tokens for stale cleanup", err)
	} else if existingTokens != nil {
		for _, tok := range *existingTokens {
			if tok.TokenType != "api_key" {
				continue
			}
			if delErr := driver.DeleteToken(ctx, auditCtx, tok.ID); delErr != nil {
				utility.DefaultLogger.Warn("failed to delete stale plugin API token", delErr, "token_id", tok.ID)
			} else {
				utility.DefaultLogger.Info("deleted stale plugin API token", "token_id", tok.ID)
			}
		}
	}

	// Generate a new 32-byte random token, hex-encoded to 64 characters.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("generate random token: %w", err)
	}
	tokenValue := hex.EncodeToString(tokenBytes)

	// Insert the token into the database.
	created, err := driver.CreateToken(ctx, auditCtx, db.CreateTokenParams{
		UserID:    systemNullableID,
		TokenType: "api_key",
		Token:     tokenValue,
		IssuedAt:  time.Now().UTC().Format(time.RFC3339),
		ExpiresAt: types.NewTimestamp(time.Now().UTC().Add(24 * time.Hour)),
		Revoked:   false,
	})
	if err != nil {
		return "", "", fmt.Errorf("insert plugin API token: %w", err)
	}
	if created == nil {
		return "", "", fmt.Errorf("insert plugin API token returned nil")
	}

	// Write the token value to <config_dir>/.plugin-api-token (mode 0600).
	configDir := filepath.Dir(cfgPath)
	tokenPath := filepath.Join(configDir, ".plugin-api-token")
	if err := os.WriteFile(tokenPath, []byte(tokenValue), 0600); err != nil {
		// Token is in the DB but file write failed. Attempt cleanup of the
		// DB row to avoid an orphaned token, then return the error.
		if delErr := driver.DeleteToken(ctx, auditCtx, created.ID); delErr != nil {
			utility.DefaultLogger.Warn("failed to clean up token after file write failure", delErr)
		}
		return "", "", fmt.Errorf("write plugin API token file: %w", err)
	}

	return created.ID, tokenPath, nil
}

// cleanupPluginAPIToken deletes the token row from the database and removes the
// token file. Called during graceful shutdown while the DB is still open.
func cleanupPluginAPIToken(ctx context.Context, driver db.DbDriver, tokenID, tokenPath, nodeID string) {
	auditCtx := audited.Ctx(types.NodeID(nodeID), types.UserID(""), "plugin-api-token-shutdown", "127.0.0.1")

	if err := driver.DeleteToken(ctx, auditCtx, tokenID); err != nil {
		utility.DefaultLogger.Warn("failed to delete plugin API token on shutdown", err, "token_id", tokenID)
	} else {
		utility.DefaultLogger.Info("deleted plugin API token on shutdown", "token_id", tokenID)
	}

	if err := os.Remove(tokenPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		utility.DefaultLogger.Warn("failed to remove plugin API token file on shutdown", err, "path", tokenPath)
	}
}
