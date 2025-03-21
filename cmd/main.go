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

	auth "github.com/hegner123/modulacms/internal/Auth"
	cli "github.com/hegner123/modulacms/internal/Cli"
	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
	install "github.com/hegner123/modulacms/internal/Install"
	middleware "github.com/hegner123/modulacms/internal/Middleware"
	router "github.com/hegner123/modulacms/internal/Router"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

type ModulaInit struct {
	UseSSL         bool
	DbFileExists   bool
	ContentVersion bool
	Certificates   bool
	Key            bool
}

var InitStatus ModulaInit
var Env = config.Config{}
var sshPort = "23234"

func main() {
	InitStatus := initFileCheck()
	authFlag := flag.Bool("auth", false, "Run oauth tests")
	updateFlag := flag.Bool("update", false, "Update binaries and plugins.")
	cliFlag := flag.Bool("cli", false, "Launch the Cli without the server.")
	versionFlag := flag.Bool("v", false, "Print version and exit")
	alphaFlag := flag.Bool("a", false, "including code for build purposes")
	verbose := flag.Bool("V", false, "Enable verbose mode")
	reset := flag.Bool("reset", false, "Delete Database and reinitialize")
	installFlag := flag.Bool("i", false, "Create tables in db driver")
	flag.Parse()

	err := install.CheckInstall()
	if err != nil {
		utility.LogError("CheckInstall", err)
		os.Exit(1)
	}

	Env = config.LoadConfig(verbose, "")
	var host = Env.SSH_Site

	if *versionFlag {
		proccessPrintVersion()
	}
	if *updateFlag {
		proccessUpdateFlag()
	}
	if *authFlag {
		proccessAuthCheck()
	}
	if *cliFlag {
		proccessRunCli()
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

	if !InitStatus.DbFileExists || *reset {
		dbc, _, _ := db.ConfigDB(Env).GetConnection()
		defer dbc.Close()
	}

	// Create the wish SSH server.
	sshServer, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, sshPort)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			cli.CliMiddleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start SSH server", "error", err)
		return
	}

	// Run the SSH server concurrently.
	go func() {

		done := make(chan os.Signal, 1)
		signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		log.Info("Starting SSH server", "host", host, "port", sshPort)
		go func() {
			if err = sshServer.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
				log.Error("Could not start server", "error", err)
				done <- nil
			}
		}()

		<-done
		log.Info("Stopping SSH server")
        os.Exit(1)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer func() { cancel() }()
		if err := sshServer.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not stop server", "error", err)
		}
	}()

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

	middlewareHandler := middleware.Serve(mux)

	if !InitStatus.UseSSL {

		log.Printf("\n\nServer is running at https://localhost:%s\n", Env.SSL_Port)
		err := http.ListenAndServeTLS(":"+Env.SSL_Port, "./certs/localhost.crt", "./certs/localhost.key", middlewareHandler)
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
	log.Printf("\n\nServer is running at http://localhost:%s\n", Env.Port)
	err = http.ListenAndServe(":"+Env.Port, middlewareHandler)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initFileCheck() ModulaInit {
	Status := ModulaInit{}
	//Check DB
	_, err := os.Open("modula.db")
	if err != nil {
		Status.DbFileExists = false
	}

	//Check for ssl certs
	_, err = os.Open("./certs/localhost.crt")
	Status.Certificates = true
	if err != nil {
		fmt.Printf("Error opening localhost.crt %s\n", err)
		Status.Certificates = false
	}
	_, err = os.Open("./certs/localhost.key")
	Status.Key = true
	if err != nil {
		fmt.Printf("Error opening localhost.key %s\n", err)
		Status.Key = false
	}
	if !Status.Certificates || !Status.Key {
		Status.UseSSL = false
	}

	//check for content version
	_, err = os.Stat("./content.version")
	if err != nil {
		Status.ContentVersion = false

	}

	return Status
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
		return
	}
	log.Fatal(message)
}

func proccessRunCli() {
	m := cli.InitialModel()
	cli.CliRun(&m)
}

func proccessUpdateFlag() {
	fmt.Printf("TODO: update flag")
}

func proccessRunInstall() {
	fmt.Println("Run Install")
}
