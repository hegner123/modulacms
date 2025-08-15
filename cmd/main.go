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

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	app := flags.ParseFlags()

	if *app.VersionFlag {
		HandleFlagVersion()
	}

	configProvider := config.NewFileProvider(*app.ConfigPath)
	configManager := config.NewManager(configProvider)

	if err := configManager.Load(); err != nil {
		utility.DefaultLogger.Fatal("Failed to load configuration", err)
	}

	cfg, _ := configManager.Config()

	_, err := install.CheckInstall(cfg, app.VerboseFlag)
	if err != nil {
		if *app.InstallFlag {
			*app.InstallFlag = false
		}
		utility.DefaultLogger.Error("", err)
		install.RunInstall(app.VerboseFlag)
	}

	if *app.UpdateFlag {
		HandleFlagUpdate(updateUrl)
	}
	if *app.AuthFlag {
		HandleFlagAuth(*cfg)
	}
	if *app.CliFlag {
		HandleFlagCLI(app.VerboseFlag, cfg)
	}

	if *app.AlphaFlag {
		HandleFlagAlpha()
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

		defer utility.HandleConnectionCloseDeferErr(dbc)
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
		return ERRSIG, err
	}

	utility.DefaultLogger.Info("Servers gracefully stopped.")
	return OKSIG, nil
}

func HandleFlagAuth(c config.Config) {
	os.Exit(0)
}

func HandleFlagAlpha() {
	_, err := os.Open("test.txt")
	if err != nil {
		utility.DefaultLogger.Fatal("failed to create database dump in archive: ", err)
	}
}
func HandleFlagVersion() {
	message, err := utility.GetVersion()
	if err != nil {
		utility.DefaultLogger.Fatal(*message, err)
	}
	utility.DefaultLogger.Info(*message)
	os.Exit(0)
}

func HandleFlagCLI(v *bool, c *config.Config) {
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

func HandleFlagUpdate(url string) {
	err := update.Fetch(url)
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}
}
