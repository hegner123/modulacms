package main

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
	"github.com/hegner123/modulacms/internal/cli"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/flags"
	"github.com/hegner123/modulacms/internal/install"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/router"
	"github.com/hegner123/modulacms/internal/update"
	"github.com/hegner123/modulacms/internal/utility"
	"golang.org/x/crypto/acme/autocert"
)

type ReturnCode int16

const (
	OKSIG ReturnCode = iota
	ERRSIG
)

func main() {

	code, err := run()
	if err != nil || code == ERRSIG {
		utility.DefaultLogger.Fatal("Root Return: ", err)
	}

}

func run() (ReturnCode, error) {
	var InitStatus install.ModulaInit
	var certDir string

	const updateUrl string = "https://modulacms.com/update"

	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	app := flags.ParseFlags()

	// Configure logger level based on verbose flag
	if *app.VerboseFlag {
		utility.DefaultLogger.SetLevel(utility.DEBUG)
	} else {
		utility.DefaultLogger.SetLevel(utility.INFO)
	}

	if *app.VersionFlag {
		HandleFlagVersion()
	}

	if *app.GenCertsFlag {
		HandleFlagGenCerts()
	}

	// Install flag bypasses all pre-launch checks (config, DB, etc.)
	if *app.InstallFlag {
		if installErr := install.RunInstall(app.VerboseFlag, app.YesFlag); installErr != nil {
			utility.DefaultLogger.Fatal("Installation failed", installErr)
		}
		os.Exit(0)
	}

	configProvider := config.NewFileProvider(*app.ConfigPath)
	configManager := config.NewManager(configProvider)

	if err := configManager.Load(); err != nil {
		utility.DefaultLogger.Fatal("Failed to load configuration", err)
	}

	configuration, err := configManager.Config()
	if err != nil {
		utility.DefaultLogger.Fatal("Failed to get configuration", err)
	}

	// Log loaded configuration
	utility.DefaultLogger.Info("Configuration loaded successfully")
	utility.DefaultLogger.Info("Database", "driver", configuration.Db_Driver, "url", configuration.Db_URL)
	utility.DefaultLogger.Info("Sites", "client", configuration.Client_Site, "admin", configuration.Admin_Site)
	utility.DefaultLogger.Info("Ports", "http", configuration.Port, "https", configuration.SSL_Port, "ssh", configuration.SSH_Port)
	utility.DefaultLogger.Info("Environment", "env", configuration.Environment, "host", configuration.Environment_Hosts[configuration.Environment])
	if configuration.Oauth_Provider_Name != "" {
		utility.DefaultLogger.Info("OAuth", "provider", configuration.Oauth_Provider_Name, "redirect", configuration.Oauth_Redirect_URL)
	}
	if configuration.Bucket_Endpoint != "" {
		utility.DefaultLogger.Info("Storage", "endpoint", configuration.Bucket_Endpoint, "media", configuration.Bucket_Media, "backup", configuration.Bucket_Backup)
	}
	utility.DefaultLogger.Info("CORS", "origins", configuration.Cors_Origins, "credentials", configuration.Cors_Credentials)

	if *app.InitDbFlag {
		HandleFlagInitDb(*app.ConfigPath, configuration)
	}

	// Initialize observability (metrics and error tracking)
	if configuration.Observability_Enabled {
		obsClient, err := utility.NewObservabilityClient(utility.ObservabilityConfig{
			Enabled:       configuration.Observability_Enabled,
			Provider:      configuration.Observability_Provider,
			DSN:           configuration.Observability_DSN,
			Environment:   configuration.Observability_Environment,
			Release:       configuration.Observability_Release,
			SampleRate:    configuration.Observability_Sample_Rate,
			TracesRate:    configuration.Observability_Traces_Rate,
			SendPII:       configuration.Observability_Send_PII,
			Debug:         configuration.Observability_Debug,
			ServerName:    configuration.Observability_Server_Name,
			FlushInterval: configuration.Observability_Flush_Interval,
			Tags:          configuration.Observability_Tags,
		})

		if err != nil {
			utility.DefaultLogger.Error("Failed to initialize observability", err)
		} else {
			utility.GlobalObservability = obsClient
			obsClient.Start(rootCtx)

			utility.DefaultLogger.Info("Observability started",
				"provider", configuration.Observability_Provider,
				"environment", configuration.Observability_Environment,
				"interval", configuration.Observability_Flush_Interval,
			)

			defer func() {
				utility.DefaultLogger.Info("Stopping observability...")
				if err := obsClient.Stop(); err != nil {
					utility.DefaultLogger.Error("Observability shutdown error", err)
				}
			}()
		}
	}

	InitStatus, err = install.CheckInstall(configuration, app.VerboseFlag)
	if err != nil {
		utility.DefaultLogger.Error("Installation check failed", err)
		if installErr := install.RunInstall(app.VerboseFlag, nil); installErr != nil {
			utility.DefaultLogger.Fatal("Installation failed", installErr)
		}
	}

	if *app.UpdateFlag {
		HandleFlagUpdate(updateUrl)
	}

	if *app.AuthFlag {
		HandleFlagAuth(*configuration)
	}

	if *app.CliFlag {
		HandleFlagCLI(app.VerboseFlag, configuration)
	}

	if *app.AlphaFlag {
		HandleFlagAlpha()
	}

	if *app.ResetFlag {
		utility.DefaultLogger.Info("Resetting database", "path", configuration.Db_URL)
		if err := os.Remove(configuration.Db_URL); err != nil {
			utility.DefaultLogger.Fatal("Error deleting database file", err)
		}
		utility.DefaultLogger.Info("Database reset complete")
	}

	// Initialize the singleton database connection pool.
	// This must happen before any handlers or middleware use db.ConfigDB().
	if _, err := db.InitDB(*configuration); err != nil {
		utility.DefaultLogger.Fatal("Failed to initialize database pool", err)
	}
	defer func() {
		if cerr := db.CloseDB(); cerr != nil {
			utility.DefaultLogger.Error("Database pool close error", cerr)
		}
	}()

	var host = configuration.SSH_Host
	sshServer, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, configuration.SSH_Port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithPublicKeyAuth(middleware.PublicKeyHandler(configuration)),
		wish.WithMiddleware(
			// Session logging - logs all connection attempts
			middleware.SSHSessionLoggingMiddleware(configuration),

			// Authentication - validates SSH keys and populates context
			middleware.SSHAuthenticationMiddleware(configuration),

			// Rate limiting - prevents brute force attacks (TODO: implement)
			// middleware.SSHRateLimitMiddleware(configuration),

			// Authorization - ensures user is authenticated or needs provisioning
			middleware.SSHAuthorizationMiddleware(configuration),

			// Application - launches the TUI
			cli.CliMiddleware(app.VerboseFlag, configuration),

			// Wish logging - framework-level logging
			logging.Middleware(),
		),
	)
	if err != nil {
		utility.DefaultLogger.Fatal("Failed to create SSH server", err)
	}

	mux := router.NewModulacmsMux(*configuration)
	middlewareHandler := middleware.DefaultMiddlewareChain(configuration)(mux)

	manager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(configuration.Environment_Hosts[configuration.Environment],
			configuration.Client_Site,
			configuration.Admin_Site,
		),
	}

	// Determine HTTP host based on environment
	httpHost := configuration.Client_Site
	if configuration.Environment == "local" {
		httpHost = "localhost"
	}

	// Create HTTP and HTTPS servers with consistent configuration
	// For local environment, don't use autocert (use local certs instead)
	var tlsConfig *tls.Config
	if configuration.Environment != "local" {
		tlsConfig = manager.TLSConfig()
	}

	httpServer := newHTTPServer(httpHost+configuration.Port, middlewareHandler, nil)
	httpsServer := newHTTPServer(httpHost+configuration.SSL_Port, middlewareHandler, tlsConfig)

	certDir, err = sanitizeCertDir(configuration.Cert_Dir)
	if err != nil {
		utility.DefaultLogger.Fatal("Certificate Directory path is invalid:", err)
	}

	// Run the SSH server concurrently
	go func() {
		utility.DefaultLogger.Info("Starting SSH server",
			"address", net.JoinHostPort(configuration.SSH_Host, configuration.SSH_Port))

		if sshErr := sshServer.ListenAndServe(); sshErr != nil && !errors.Is(sshErr, ssh.ErrServerClosed) {
			utility.DefaultLogger.Error("SSH server error", sshErr)
			done <- syscall.SIGTERM
		}
	}()

	// Run HTTPS server if configured
	if !InitStatus.UseSSL && configuration.Environment != "http-only" {
		go func() {
			utility.DefaultLogger.Info("Starting HTTPS server", "address", httpsServer.Addr)
			httpsErr := httpsServer.ListenAndServeTLS(
				filepath.Join(certDir, "localhost.crt"),
				filepath.Join(certDir, "localhost.key"),
			)
			if httpsErr != nil && httpsErr != http.ErrServerClosed {
				utility.DefaultLogger.Error("HTTPS server error", httpsErr)
				done <- syscall.SIGTERM
			}
		}()
	}

	// Run HTTP server
	go func() {
		utility.DefaultLogger.Info("Starting HTTP server", "address", httpServer.Addr)
		httpErr := httpServer.ListenAndServe()
		if httpErr != nil && httpErr != http.ErrServerClosed {
			utility.DefaultLogger.Error("HTTP server error", httpErr)
			done <- syscall.SIGTERM
		}
	}()

	<-done
	utility.DefaultLogger.Info("Shutting down servers...")

	// Cancel root context to stop all background workers (observability, etc.)
	rootCancel()

	// Create new context with timeout for graceful shutdown
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
		return ERRSIG, err
	}

	utility.DefaultLogger.Info("Servers gracefully stopped.")
	return OKSIG, nil
}

func HandleFlagAuth(c config.Config) {
	utility.DefaultLogger.Info("Auth flag handler not yet implemented")
	os.Exit(0)
}

func HandleFlagAlpha() {
	utility.DefaultLogger.Info("Alpha flag handler not yet implemented")
	os.Exit(0)
}
func HandleFlagVersion() {
	message, err := utility.GetVersion()
	if err != nil {
		utility.DefaultLogger.Fatal(*message, err)
	}
	utility.DefaultLogger.Info(*message)
	os.Exit(0)
}

func HandleFlagGenCerts() {
	utility.DefaultLogger.Info("Generating self-signed SSL certificates...")

	// Default cert directory
	certDir := "./certs"
	domain := "localhost"

	// Try to read config for cert_dir setting
	configProvider := config.NewFileProvider("config.json")
	configManager := config.NewManager(configProvider)
	if err := configManager.Load(); err == nil {
		if cfg, err := configManager.Config(); err == nil {
			if cfg.Cert_Dir != "" {
				certDir = cfg.Cert_Dir
			}
			if cfg.Client_Site != "" && cfg.Client_Site != "localhost" {
				domain = cfg.Client_Site
			}
		}
	}

	utility.DefaultLogger.Info("Certificate directory", certDir)
	utility.DefaultLogger.Info("Domain", domain)

	err := utility.GenerateSelfSignedCert(certDir, domain)
	if err != nil {
		utility.DefaultLogger.Fatal("Failed to generate certificates", err)
	}

	utility.DefaultLogger.Info("✓ Successfully generated SSL certificates!")
	utility.DefaultLogger.Info("  - Certificate: " + certDir + "/localhost.crt")
	utility.DefaultLogger.Info("  - Private Key: " + certDir + "/localhost.key")
	utility.DefaultLogger.Info("")

	// Offer to trust the certificate
	certPath := filepath.Join(certDir, "localhost.crt")
	if err := utility.TrustCertificate(certPath); err != nil {
		utility.DefaultLogger.Warn("Failed to trust certificate:", err)
		utility.DefaultLogger.Info("You can manually trust it later - see LOCAL_HTTPS_SETUP.md")
	}

	utility.DefaultLogger.Info("")
	utility.DefaultLogger.Info("To use HTTPS locally, set environment to 'local' in config.json")
	os.Exit(0)
}

func HandleFlagInitDb(configPath string, c *config.Config) {
	utility.DefaultLogger.Info("Initializing database tables and bootstrap data...")
	if err := install.CreateDbSimple(configPath, c); err != nil {
		utility.DefaultLogger.Fatal("Database initialization failed", err)
	}
	utility.DefaultLogger.Info("Database initialization complete")
	os.Exit(0)
}

func HandleFlagCLI(v *bool, c *config.Config) {
	model, _ := cli.InitialModel(v, c)
	if _, e := cli.CliRun(&model); !e {
		process, err := os.FindProcess(os.Getpid())
		if err != nil {
			utility.DefaultLogger.Error("", err)
			return
		}

		if err := process.Signal(syscall.SIGTERM); err != nil {
			utility.DefaultLogger.Error("", err)
		}
	}
}

func HandleFlagUpdate(url string) {
	// Note: url parameter is deprecated, now using GitHub Releases API
	utility.DefaultLogger.Info("Checking for updates...")

	currentVersion := utility.GetCurrentVersion()
	utility.DefaultLogger.Info("Current version", currentVersion)

	// Check for updates via GitHub API
	release, available, err := update.CheckForUpdates(currentVersion, "stable")
	if err != nil {
		utility.DefaultLogger.Error("Update check failed", err)
		os.Exit(1)
	}

	if !available {
		utility.DefaultLogger.Info("Already running latest version")
		os.Exit(0)
	}

	utility.DefaultLogger.Info("Update available", release.TagName)

	// Get download URL for current platform
	downloadURL, err := update.GetDownloadURL(release, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		utility.DefaultLogger.Error("No compatible binary found", err)
		os.Exit(1)
	}

	utility.DefaultLogger.Info("Downloading update...")
	tempPath, err := update.DownloadUpdate(downloadURL)
	if err != nil {
		utility.DefaultLogger.Error("Download failed", err)
		os.Exit(1)
	}

	utility.DefaultLogger.Info("Applying update...")
	err = update.ApplyUpdate(tempPath)
	if err != nil {
		utility.DefaultLogger.Error("Update failed", err)
		os.Exit(1)
	}

	utility.DefaultLogger.Info("Update complete! Please restart ModulaCMS.")
	os.Exit(0)
}

func newHTTPServer(addr string, handler http.Handler, tlsConfig *tls.Config) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		TLSConfig:    tlsConfig,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func sanitizeCertDir(configCertDir string) (string, error) {
	if strings.TrimSpace(configCertDir) == "" {
		return "", errors.New("certificate directory path cannot be empty")
	}

	certDir := filepath.Clean(configCertDir)

	absPath, err := filepath.Abs(certDir)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return "", errors.New("certificate path is not a directory")
	}

	return absPath, nil
}
