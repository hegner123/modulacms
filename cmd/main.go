package main

import (
	"context"
	"errors"
	"flag"
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
	"golang.org/x/crypto/acme/autocert"

	auth "github.com/hegner123/modulacms/internal/Auth"
	cli "github.com/hegner123/modulacms/internal/Cli"
	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
	install "github.com/hegner123/modulacms/internal/Install"
	middleware "github.com/hegner123/modulacms/internal/Middleware"
	router "github.com/hegner123/modulacms/internal/Router"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

var InitStatus install.ModulaInit
var Env = config.Config{}
var certDir string

func main() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	authFlag := flag.Bool("auth", false, "Run oauth tests")
	updateFlag := flag.Bool("update", false, "Update binaries and plugins.")
	cliFlag := flag.Bool("cli", false, "Launch the Cli without the server.")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	alphaFlag := flag.Bool("a", false, "including code for build purposes")
	verbose := flag.Bool("v", false, "Enable verbose mode")
	reset := flag.Bool("reset", false, "Delete Database and reinitialize")
	installFlag := flag.Bool("i", false, "Run Installation UI")
	flag.Parse()
	if *versionFlag {
		proccessPrintVersion()
	}

	InstallStatus, err := install.CheckInstall()
	if err != nil {
		utility.DefaultLogger.Error("CheckInstall", err)
		ok := install.InstallUI(&InstallStatus)
		if !ok {
			os.Exit(1)
		}
	}

	if *updateFlag {
		proccessUpdateFlag()
	}
	if *authFlag {
		proccessAuthCheck()
	}
	if *cliFlag {
		proccessRunCli(verbose)
	}

	if *alphaFlag {
		proccessAlphaFlag()
	}

	if *reset {
		fmt.Println("Reset DB:")
		err := os.Remove("./modula.db")
		if err != nil {
			log.Fatal("Error deleting file:", err)
		}
	}

	if *installFlag {
		// check if installed, ask if you want to reinstall and lose content
		proccessRunInstall()
	}
	Env = config.LoadConfig(verbose, "")

	if !InitStatus.DbFileExists || *reset {
		dbc, _, _ := db.ConfigDB(Env).GetConnection()
		defer dbc.Close()
	}

	var host = Env.SSH_Host
	sshServer, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, Env.SSH_Port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			cli.CliMiddleware(verbose),
			logging.Middleware(),
		),
	)

    // Mux Routes

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		router.LoginHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		router.RegisterHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/auth/reset", func(w http.ResponseWriter, r *http.Request) {
		router.ResetPasswordHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/auth/oauth", func(w http.ResponseWriter, r *http.Request) {
		router.OauthCallbackHandler(Env, "")
	})
	mux.HandleFunc("/api/v1/admincontentdatas", func(w http.ResponseWriter, r *http.Request) {
		router.AdminContentDatasHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/admincontentdatas/", func(w http.ResponseWriter, r *http.Request) {
		router.AdminContentDataHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/admincontentfields", func(w http.ResponseWriter, r *http.Request) {
		router.AdminContentFieldsHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/admincontentfields/", func(w http.ResponseWriter, r *http.Request) {
		router.AdminContentFieldHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/admindatatypes", func(w http.ResponseWriter, r *http.Request) {
		router.AdminDatatypesHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/admindatatypes/", func(w http.ResponseWriter, r *http.Request) {
		router.AdminDatatypeHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/adminfields", func(w http.ResponseWriter, r *http.Request) {
		router.AdminFieldsHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/adminfields/", func(w http.ResponseWriter, r *http.Request) {
		router.AdminFieldHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/adminroutes", func(w http.ResponseWriter, r *http.Request) {
		router.AdminRoutesHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/adminroutes/", func(w http.ResponseWriter, r *http.Request) {
		router.AdminRouteHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/contentdata", func(w http.ResponseWriter, r *http.Request) {
		router.ContentDatasHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/contentdata/", func(w http.ResponseWriter, r *http.Request) {
		router.ContentDataHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/contentfields", func(w http.ResponseWriter, r *http.Request) {
		router.ContentFieldsHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/contentfields/", func(w http.ResponseWriter, r *http.Request) {
		router.ContentFieldHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/datatype", func(w http.ResponseWriter, r *http.Request) {
		router.DatatypesHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/datatype/", func(w http.ResponseWriter, r *http.Request) {
		router.DatatypeHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/fields", func(w http.ResponseWriter, r *http.Request) {
		router.FieldsHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/fields/", func(w http.ResponseWriter, r *http.Request) {
		router.FieldHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/media", func(w http.ResponseWriter, r *http.Request) {
		router.MediasHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/media/", func(w http.ResponseWriter, r *http.Request) {
		router.MediaHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/mediadimensions", func(w http.ResponseWriter, r *http.Request) {
		router.MediaDimensionsHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/mediadimensions/", func(w http.ResponseWriter, r *http.Request) {
		router.MediaDimensionHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/mediaupload/", func(w http.ResponseWriter, r *http.Request) {
		router.MediaUploadHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/routes", func(w http.ResponseWriter, r *http.Request) {
		router.RoutesHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/routes/", func(w http.ResponseWriter, r *http.Request) {
		router.RoutesHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/roles", func(w http.ResponseWriter, r *http.Request) {
		router.RolesHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/roles/", func(w http.ResponseWriter, r *http.Request) {
		router.RoleHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		router.SessionsHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/sessions/", func(w http.ResponseWriter, r *http.Request) {
		router.SessionHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/tables", func(w http.ResponseWriter, r *http.Request) {
		router.TablesHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/tables/", func(w http.ResponseWriter, r *http.Request) {
		router.TableHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/tokens", func(w http.ResponseWriter, r *http.Request) {
		router.TokensHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/tokens/", func(w http.ResponseWriter, r *http.Request) {
		router.TokenHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/usersoauth", func(w http.ResponseWriter, r *http.Request) {
		router.UserOauthsHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/usersoauth/", func(w http.ResponseWriter, r *http.Request) {
		router.UserOauthHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		router.UsersHandler(w, r, Env)
	})
	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		router.UserHandler(w, r, Env)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		router.SlugHandler(w, r, Env)
	})



    // Certificates

	manager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(Env.Cert_Dir),                         // Folder to store certificates
		HostPolicy: autocert.HostWhitelist(Env.Client_Site, Env.Admin_Site), // Your domain(s)
	}

	middlewareHandler := middleware.Serve(mux)
	var (
		// Define your HTTP server instance.
		httpServer = &http.Server{
			Addr:    Env.Port,
			Handler: middlewareHandler,
		}
		// Define your HTTPS server instance.
		httpsServer = &http.Server{
			Addr:      "localhost:" + Env.SSL_Port,
			TLSConfig: manager.TLSConfig(),
			Handler:   middlewareHandler,
		}
	)

	l := len(Env.Cert_Dir)
	c := Env.Cert_Dir[l-1]
	if string(c) != "/" {
		certDir = Env.Cert_Dir + "/"
	} else {
		certDir = Env.Cert_Dir
	}

	// Run the SSH server concurrently.
	go func() {

		utility.DefaultLogger.Info("Starting SSH server", "ssh "+Env.SSH_Host+" -p "+Env.SSH_Port)
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

		if !InitStatus.UseSSL {
			utility.DefaultLogger.Info("Server is running at https://localhost:", Env.SSL_Port)
			err = httpsServer.ListenAndServeTLS(certDir+"localhost.crt", certDir+"localhost.key")
			if err != nil {
				utility.DefaultLogger.Info("Shutting Down Server", err)
				done <- syscall.SIGTERM
			}
		}
		utility.DefaultLogger.Info("Server is running at http://localhost:", Env.Port)
		err = httpServer.ListenAndServe()
		if err != nil {
			utility.DefaultLogger.Info("Shutting Down Server", err)
			done <- syscall.SIGTERM
		}
	}()


	// Wait for an OS signal (e.g., Ctrl-C)
	<-done
	utility.DefaultLogger.Info("Shutting down servers...")

	// Create a context with a timeout for graceful shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server.
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if err := httpsServer.Shutdown(ctx); err != nil {
		log.Printf("HTTPS server shutdown error: %v", err)
	}

	// Shutdown SSH server.
	if err := sshServer.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Printf("SSH server shutdown error: %v", err)
	}

	utility.DefaultLogger.Info("Servers gracefully stopped.")
}


func proccessAuthCheck() {
	auth.OauthSettings(Env)
	os.Exit(0)
}

func proccessAlphaFlag() {
	_, err := os.Open("test.txt")
	if err != nil {
		log.Fatal("failed to create database dump in archive: ", err)
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

func proccessRunCli(v *bool) {
	m := cli.InitialModel(v)
	if _, e := cli.CliRun(&m); !e {
		//os.Exit(0)
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			fmt.Println("Error finding process:", err)
			return
		}

		// Send a SIGTERM to the process.
		if err := p.Signal(syscall.SIGTERM); err != nil {
			fmt.Println("Error sending signal:", err)
		}
	}
}

func proccessUpdateFlag() {
	fmt.Printf("TODO: update flag")
}

func proccessRunInstall() {
	fmt.Println("Run Install")
}
