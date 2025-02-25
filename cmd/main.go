package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	cli "github.com/hegner123/modulacms/internal/Cli"
	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
	router "github.com/hegner123/modulacms/internal/Router"
	api_v1 "github.com/hegner123/modulacms/internal/Server"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

var useSSL, dbFileExists bool
var Env = config.Config{}

func main() {
	useSSL, dbFileExists = InitFileCheck()
	cliFlag := flag.Bool("cli", false, "Launch the Cli without the server.")
	versionFlag := flag.Bool("v", false, "Print version and exit")
	alphaFlag := flag.Bool("a", false, "including code for build purposes")
	verbose := flag.Bool("V", false, "Enable verbose mode")
	reset := flag.Bool("reset", false, "Delete Database and reinitialize")

	if *alphaFlag {
		_, err := os.Open("test.txt")
		if err != nil {
			log.Panic("failed to create database dump in archive: ", err)
		}
	}

	flag.Parse()
	if *versionFlag {
		message,err := utility.GetVersion()
        if err!=nil {
            return
        }
		log.Fatal(message)
	}
	config := config.LoadConfig(verbose, "")

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
	if !dbFileExists || *reset {
		dbc := db.GetDb(db.Database{})

		//err = (dbc, ctx, verbose, "modula.db")

		defer dbc.Connection.Close()
	}

	if *cliFlag {
		cli.CliRun()
		os.Exit(0)
	}

	mux := http.NewServeMux()
	api := api_v1.ApiServerV1{
        DeleteHandler: router.ApiDeleteHandler,
        GetHandler: router.ApiGetHandler,
        PutHandler: router.ApiPutHandler,
        PostHandler: router.ApiPostHandler,
    }

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		router.Router(w, r, &api)

	})

	if !useSSL {

		log.Printf("\n\nServer is running at https://localhost:%s\n", config.SSL_Port)
		err := http.ListenAndServeTLS(":"+config.SSL_Port, "./certs/localhost.crt", "./certs/localhost.key", mux)
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
	log.Printf("\n\nServer is running at http://localhost:%s\n", config.Port)
	err := http.ListenAndServe(":"+config.Port, mux)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func InitFileCheck() (bool, bool) {
	useSSL := true
	dbFileExists := true
	_, err := os.Open("modula.db")
	if err != nil {
		dbFileExists = false
	}
	var cert, key bool

	_, err = os.Open("./certs/localhost.crt")
	cert = true
	if err != nil {
		fmt.Printf("Error opening localhost.crt %s\n", err)
		cert = false
	}
	_, err = os.Open("./certs/localhost.key")
	key = true
	if err != nil {
		fmt.Printf("Error opening localhost.key %s\n", err)
		key = false
	}
	if !cert || !key {
		useSSL = false
	}
	return useSSL, dbFileExists
}
