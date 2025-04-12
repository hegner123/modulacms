// Package main is the entry point for ModulaCMS, a flexible content management system.
//
// It handles server initialization, command-line arguments, installation processes,
// and sets up HTTP, HTTPS, and SSH servers. The main package coordinates the various
// components of the CMS including authentication, routing, middleware, and database
// connections. It supports multiple operational modes including CLI-only, installation,
// and update functionalities.
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
	"github.com/hegner123/modulacms/internal/auth"
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

type AppFlags struct {
	AuthFlag    *bool
	UpdateFlag  *bool
	CliFlag     *bool
	VersionFlag *bool
	AlphaFlag   *bool
	VerboseFlag *bool
	ResetFlag   *bool
	InstallFlag *bool
	ConfigPath  *string
}

var InitStatus install.ModulaInit
var certDir string

const updateUrl string = "https://modulacms.com/update"

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Create a context with a timeout for graceful shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	app := flags.ParseFlags()

	if *app.VersionFlag {
		proccessPrintVersion()
	}

	configProvider := config.NewFileProvider(*app.ConfigPath)
	configManager := config.NewManager(configProvider)

	// Load config
	if err := configManager.Load(); err != nil {
		utility.DefaultLogger.Fatal("Failed to load configuration", err)
	}

	// Get the config
	cfg, _ := configManager.Config()

	// If install check fails, and install flag is present
	// disable flag to prevent multiple calls to install
	_, err := install.CheckInstall(cfg, app.VerboseFlag)
	if err != nil {
		if *app.InstallFlag {
			*app.InstallFlag = false
		}
		utility.DefaultLogger.Error("", err)
		install.RunInstall(app.VerboseFlag)
	}

	if *app.UpdateFlag {
		proccessUpdateFlag()
	}
	if *app.AuthFlag {
		proccessAuthCheck(*cfg)
	}
	if *app.CliFlag {
		proccessRunCli(app.VerboseFlag, cfg)
	}

	if *app.AlphaFlag {
		proccessAlphaFlag()
	}

	if *app.ResetFlag {
		fmt.Println("Reset DB:")
		err := os.Remove("./modula.db")
		if err != nil {
			log.Fatal("Error deleting file:", err)
		}
	}

	if *app.InstallFlag {
		install.RunInstall(app.VerboseFlag)
	}

	if !InitStatus.DbFileExists || *app.ResetFlag {
		dbc, _, _ := db.ConfigDB(*cfg).GetConnection()

		defer func() {
			if closeErr := dbc.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
		}()

	}

	var host = cfg.SSH_Host
	sshServer, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, cfg.SSH_Port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			cli.CliMiddleware(app.VerboseFlag, cfg),
			logging.Middleware(),
		),
	)

	mux := router.NewModulacmsMux(*cfg)
	middlewareHandler := middleware.Serve(mux, cfg)
	manager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.Environment_Hosts[cfg.Environment], cfg.Client_Site, cfg.Admin_Site), // Your domain(s)
	}
	var (
		// Define your HTTP server instance.
		httpServer = &http.Server{
			Addr:         cfg.Client_Site + cfg.Port,
			Handler:      middlewareHandler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		// Define your HTTPS server instance.
		httpsServer = &http.Server{
			Addr:         cfg.Client_Site + cfg.SSL_Port,
			TLSConfig:    manager.TLSConfig(),
			Handler:      middlewareHandler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
	)
	if cfg.Environment == "local" {
		httpServer = &http.Server{
			Addr:         "localhost:" + cfg.SSL_Port,
			Handler:      middlewareHandler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
		httpsServer = &http.Server{
			Addr:         "localhost:" + cfg.SSL_Port,
			Handler:      middlewareHandler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}
	}

	l := len(cfg.Cert_Dir)
	c := cfg.Cert_Dir[l-1]
	if string(c) != "/" {
		certDir = cfg.Cert_Dir + "/"
	} else {
		certDir = cfg.Cert_Dir
	}

	// Run the SSH server concurrently.
	go func() {

		utility.DefaultLogger.Info("Starting SSH server", "ssh "+cfg.SSH_Host+" -p "+cfg.SSH_Port)
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

	go func() {

		if !InitStatus.UseSSL && cfg.Environment != "local" {
			utility.DefaultLogger.Info("Server is running at https://localhost:", cfg.SSL_Port)
			err = httpsServer.ListenAndServeTLS(certDir+"localhost.crt", certDir+"localhost.key")
			if err != nil {
				utility.DefaultLogger.Info("Shutting Down Server", err)
				done <- syscall.SIGTERM
			}
		}
		utility.DefaultLogger.Info("Server is running at http://localhost:", cfg.Port)
		err = httpServer.ListenAndServe()
		if err != nil {
			utility.DefaultLogger.Info("Shutting Down Server", err)
			done <- syscall.SIGTERM
		}
	}()

	// Wait for an OS signal (e.g., Ctrl-C)
	<-done
	utility.DefaultLogger.Info("Shutting down servers...")

	// Shutdown HTTP server.
	if err := httpServer.Shutdown(ctx); err != nil {
		utility.DefaultLogger.Error("HTTP server shutdown error: %v", err)
	}

	if err := httpsServer.Shutdown(ctx); err != nil {
		utility.DefaultLogger.Error("HTTPS server shutdown error: %v", err)
	}

	// Shutdown SSH server.
	if err := sshServer.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		utility.DefaultLogger.Error("SSH server shutdown error: %v", err)
	}

	utility.DefaultLogger.Info("Servers gracefully stopped.")
}

func proccessAuthCheck(c config.Config) {
	auth.OauthSettings(c)
	os.Exit(0)
}

func proccessAlphaFlag() {
	_, err := os.Open("test.txt")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to create database dump in archive: ", err)
	}
}
func proccessPrintVersion() {
	message, err := utility.GetVersion()
	if err != nil {
		utility.DefaultLogger.Fatal(*message, err)
	}
	utility.DefaultLogger.Info(*message)
	os.Exit(0)
}

func proccessRunCli(v *bool, c *config.Config) {
	m, _ := cli.InitialModel(v, c)
	if _, e := cli.CliRun(&m); !e {
		//os.Exit(0)
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			utility.DefaultLogger.Error("", err)
			return
		}

		// Send a SIGTERM to the process.
		if err := p.Signal(syscall.SIGTERM); err != nil {
			utility.DefaultLogger.Error("", err)
		}
	}
}

func proccessUpdateFlag() {
	err := update.Fetch(updateUrl)
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
}
