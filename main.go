package main

import (
	"flag"
	"fmt"
    "path/filepath"
	"log"
	"net/http"
	"os"
)



func main() {
	verbose := flag.Bool("v", false, "Enable verbose mode")
	reset := flag.Bool("r", false, "Delete Database and reinitialize")
	flag.Parse()
	if *reset {
		fmt.Println("Verbose mode:")
		err := os.Remove("./modula.db")
		if err != nil {
			log.Fatal("Error deleting file:", err)
			fmt.Printf("error deleting file\n")
		}
	}
    loadConfig(verbose)
	db, err := initializeDatabase(*reset)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()


	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
            fmt.Print("page route\n")
			handlePageRoutes(w, r)
		} else {
            fmt.Print("static route\n")
			staticFileHandler(w, r)
		}
	})

	mux.HandleFunc("/api", apiRoutes)

	fs := http.FileServer(http.Dir("static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fs))


	log.Println("\n\nServer is running at http://localhost:8080/blog")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}


func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join("public", r.URL.Path)
    fmt.Print(filePath)
	http.ServeFile(w, r, filePath)
}
