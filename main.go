package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)


func hasFileExtension(path string) bool {
	ext := filepath.Ext(path)
	return ext != ""
}

func main() {
	//add function to check for existance of keys
	//if not found log error w/ bash to create cert and link to docs
	//if found continue
	var useSSL = true
	var dbFileExists = true
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

	verbose := flag.Bool("v", false, "Enable verbose mode")
	reset := flag.Bool("r", false, "Delete Database and reinitialize")

	flag.Parse()
	config := loadConfig(verbose)

	if *reset {
		fmt.Println("Verbose mode:")
		err := os.Remove("./modula.db")  
		if err != nil {
			log.Fatal("Error deleting file:", err)
		}
	}

	if config.ClientSite != "" {
		clientDB, err := initializeClientDatabase(config.ClientSite, *reset)
		if err != nil {
			fmt.Printf("\nFailed to initialize database: %s", err)
			return
		}
		defer clientDB.Close()
	}
	if !dbFileExists || *reset {
		db, err := initializeDatabase(*reset)
		if err != nil {
			fmt.Printf("\nFailed to initialize database: %s", err)
		}
		defer db.Close()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if hasFileExtension(r.URL.Path) {
			fmt.Print("static route\n")
			staticFileHandler(w, r)
		} else {
			fmt.Print("page route\n")
			handlePageRoutes(w, r)
		}
	})

	mux.HandleFunc("/api", apiRoutes)
	if useSSL {

		log.Printf("\n\nServer is running at https://localhost:%s", config.SSLPort)
		err = http.ListenAndServeTLS(":"+config.SSLPort, "cert.pem", "./key.pem", mux)
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
	log.Printf("\n\nServer is running at localhost:%s", config.Port)
	err = http.ListenAndServe(":"+config.Port, mux)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
