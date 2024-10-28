package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// PageData holds the dynamic data to be rendered in the HTML template

// Define the struct to match the JSON structure
type Config struct {
	DB_URL          string `json:"db_url"`
	DB_NAME         string `json:"db_name"`
	DB_PASSWORD     string `json:"db_password"`
	Bucket_URL      string `json:"bucket_url"`
	Bucket_PASSWORD string `json:"bucket_password"`
}
type User struct {
	ID       int
	UserName string
	Name     string
	Email    string
	Hash     string
	Role     string
}

type BlogPageData struct {
	Title       string
	Heading     string
	Description string
	Posts       []Post
}

type Page404 struct {
	Title   string
	Message string
}

type Routes struct {
	Title string
	Pages []Post
}


func main() {
	mux := http.NewServeMux()
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
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal("Error opening file:", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		log.Fatal("Error parsing JSON:", err)
	}
	if *verbose {
		fmt.Printf("%s\n", bytes)
		fmt.Printf(`
            DB URL:%s, 
            DB NAME: %s, 
            DB Password: %s,
            Bucket URL: %s, 
            Bucket Password: %s
            `, config.DB_URL, config.DB_NAME,
			config.DB_PASSWORD, config.Bucket_URL, config.Bucket_PASSWORD)

	}

	// verify connections
	//      - DB
	//      - Bucket
	//      - Environment

	db, err := initializeDatabase(*reset)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()


	// Handle any route starting with "/api/"

    http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("templates/js"))))
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request,){
        http.HandlerFunc(handlePageRoutes).ServeHTTP(w,r)
    })
    

	log.Println("\n\nServer is running at http://localhost:8080/blog")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
