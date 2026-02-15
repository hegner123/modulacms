package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/bucket"
	"github.com/hegner123/modulacms/internal/cli"
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

		cfg, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		// Initialize isolated plugin database pool
		pluginPool, pluginPoolCleanup, err := initPluginPool(cfg)
		if err != nil {
			return fmt.Errorf("plugin pool init failed: %w", err)
		}
		defer pluginPoolCleanup()
		pluginManager := initPluginManager(rootCtx, cfg, pluginPool)

		if cfg.Node_ID == "" {
			cfg.Node_ID = string(types.NewNodeID())
			utility.DefaultLogger.Finfo("No node_id configured, generated ephemeral node ID", "node_id", cfg.Node_ID)
		}

		logConfigSummary(cfg)

		// Start observability
		obsCleanup := initObservability(rootCtx, cfg)
		defer obsCleanup()

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
				for _, name := range []string{cfg.Bucket_Media, cfg.Bucket_Backup} {
					if name == "" {
						continue
					}
					if bErr := bucket.EnsureBucket(s3Session, name); bErr != nil {
						utility.DefaultLogger.Warn("Failed to ensure bucket", bErr, "bucket", name)
					}
				}
			}
		}

		// SSH server
		sshServer, err := wish.NewServer(
			wish.WithAddress(net.JoinHostPort(cfg.SSH_Host, cfg.SSH_Port)),
			wish.WithHostKeyPath(".ssh/id_ed25519"),
			wish.WithPublicKeyAuth(middleware.PublicKeyHandler(cfg)),
			wish.WithMiddleware(
				middleware.SSHSessionLoggingMiddleware(cfg),
				middleware.SSHAuthenticationMiddleware(cfg),
				middleware.SSHAuthorizationMiddleware(cfg),
				cli.CliMiddleware(&verbose, cfg, driver, utility.DefaultLogger),
				logging.Middleware(),
			),
		)
		if err != nil {
			return err
		}

		// HTTP router and middleware
		var bridge *plugin.HTTPBridge
		if pluginManager != nil {
			bridge = pluginManager.Bridge()
		}
		mux := router.NewModulacmsMux(*cfg, bridge)

		// Phase 3: Inject HookRunner into request context so that audited
		// operations can dispatch content lifecycle hooks. The HookRunnerMiddleware
		// is applied before the default chain so the runner is available to all
		// downstream middleware and handlers.
		var hookRunner audited.HookRunner
		if pluginManager != nil {
			hookRunner = pluginManager.HookEngine()
		}
		middlewareHandler := middleware.Chain(
			middleware.HookRunnerMiddleware(hookRunner),
		)(middleware.DefaultMiddlewareChain(cfg)(mux))

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

		httpServer := newHTTPServer(httpHost+cfg.Port, middlewareHandler, nil)
		httpsServer := newHTTPServer(httpHost+cfg.SSL_Port, middlewareHandler, tlsConfig)

		certDir, err := sanitizeCertDir(cfg.Cert_Dir)
		if err != nil {
			return err
		}

		// Start SSH server
		go func() {
			utility.DefaultLogger.Info("Starting SSH server",
				"address", net.JoinHostPort(cfg.SSH_Host, cfg.SSH_Port))

			if sshErr := sshServer.ListenAndServe(); sshErr != nil && !errors.Is(sshErr, ssh.ErrServerClosed) {
				utility.DefaultLogger.Error("SSH server error", sshErr)
				done <- syscall.SIGTERM
			}
		}()

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
