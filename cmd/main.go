package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	auth "github.com/hegner123/modulacms/internal/Auth"
	cli "github.com/hegner123/modulacms/internal/Cli"
	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
	middleware "github.com/hegner123/modulacms/internal/Middleware"
	router "github.com/hegner123/modulacms/internal/Router"
	api_v1 "github.com/hegner123/modulacms/internal/Server"
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

func main() {
	InitStatus := InitFileCheck()
	authFlag := flag.Bool("auth", false, "Run oauth tests")
	updateFlag := flag.Bool("update", false, "Update binaries and plugins.")
	cliFlag := flag.Bool("cli", false, "Launch the Cli without the server.")
	versionFlag := flag.Bool("v", false, "Print version and exit")
	alphaFlag := flag.Bool("a", false, "including code for build purposes")
	verbose := flag.Bool("V", false, "Enable verbose mode")
	reset := flag.Bool("reset", false, "Delete Database and reinitialize")
	flag.Parse()

	if *updateFlag {
		fmt.Printf("TODO: update flag")
	}

	if *alphaFlag {
		_, err := os.Open("test.txt")
		if err != nil {
			log.Panic("failed to create database dump in archive: ", err)
		}
	}

	if *versionFlag {
		message, err := utility.GetVersion()
		if err != nil {
			return
		}
		log.Fatal(message)
	}
	Env = config.LoadConfig(verbose, "")
	if *authFlag {
		auth.OauthSettings(Env)
		os.Exit(0)
	}

	if *reset {
		fmt.Println("Reset DB:")
		err := os.Remove("./modula.db")
		if err != nil {
			log.Fatal("Error deleting file:", err)
		}
	}

	/*if config.ClientSite != "" {
		clientDB, err := initClientDatabase(config.ClientSite, *reset)
		if err != nil {
			fmt.Printf("\nFailed to initialize database: %s", err)
			return
		}
		defer clientDB.Close()
	}*/
	if !InitStatus.DbFileExists || *reset {
		dbc, _, _ := db.ConfigDB(Env).GetConnection()
		defer dbc.Close()
	}

	if *cliFlag {
		r := cli.CliRun()
		if !r {
			os.Exit(0)
		}
	}

	mux := http.NewServeMux()
	api := api_v1.ApiServerV1{
		Config:        Env,
		DeleteHandler: router.ApiDeleteHandler,
		GetHandler:    router.ApiGetHandler,
		PutHandler:    router.ApiPutHandler,
		PostHandler:   router.ApiPostHandler,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		router.Router(w, r, &api)
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
	err := http.ListenAndServe(":"+Env.Port, middlewareHandler)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func InitFileCheck() ModulaInit {
	Status := ModulaInit{}
	_, err := os.Open("modula.db")
	if err != nil {
		Status.DbFileExists = false
	}

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
	_, err = os.Stat("./content.version")
	if err != nil {
		Status.ContentVersion = false

	}
	return Status
}
