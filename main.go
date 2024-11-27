package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var useSSL, dbFileExists bool = initFileCheck()

func hasFileExtension(path string) bool {
	ext := filepath.Ext(path)
	return ext != ""
}

func main() {
	versionFlag := flag.Bool("v", false, "Print version and exit")
	alphaFlag := flag.Bool("a", false, "including code for build purposes")
	verbose := flag.Bool("V", false, "Enable verbose mode")
	reset := flag.Bool("reset", false, "Delete Database and reinitialize")

	if *alphaFlag {
		_, err := os.Open("test.txt")
		if err != nil {
			logError("failed to create database dump in archive: ", err)
		}
	}

	flag.Parse()
	if *versionFlag {
		message := logGetVersion()
		log.Fatal(message)
	}
	config := loadConfig(verbose)

	if *reset {
		fmt.Println("Verbose mode:")
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
		db, ctx, err := getDb(Database{DB: "modula.db"})
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		err = initDb(db, ctx)
		if err != nil {
			fmt.Printf("\nFailed to initialize database: %s", err)
		}
		createSetupInserts(db, ctx, "")
		defer db.Close()
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		router(w, r)
	})

	if useSSL {

		log.Printf("\n\nServer is running at https://localhost:%s", config.SSL_Port)
		err := http.ListenAndServeTLS(":"+config.SSL_Port, "localhost.crt", "localhost.key", mux)
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
	log.Printf("\n\nServer is running at localhost:%s", config.Port)
	err := http.ListenAndServe(":"+config.Port, mux)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initFileCheck() (bool, bool) {
	useSSL := true
	dbFileExists := true
	_, err := os.Open("modula.db")
	if err != nil {
		dbFileExists = false
	}
	var cert, key bool
	_, err = os.Open("cert.pem")
	cert = true
	if err != nil {
		cert = false
	}
	_, err = os.Open("key.pem")
	key = true
	if err != nil {
		key = false
	}
	if !cert || !key {
		useSSL = false
	}
	return useSSL, dbFileExists
}
