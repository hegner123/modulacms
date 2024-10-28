package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
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
	verbose := flag.Bool("v", false, "Enable verbose mode")
	reset := flag.Bool("r", false, "Delete Database and reinitialize")
	flag.Parse()
	fmt.Println("Verbose mode:", *verbose)
	if *reset {
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

	tmpl, err := template.ParseFiles("templates/blog.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	tmpl404, err := template.ParseFiles("templates/404.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	//pageTempl, err := template.ParseFiles("templates/page.html")
	//if err != nil {
	//	log.Fatalf("Failed to parse template: %v", err)
	//}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		/*	exists := postExists(db, r.PathValue("*"))
			w.Header().Set("Content-Type", "text/html")
			if !exists {
				data := Page404{Title: "404 Not Found", Message: "Test the chakifbakj abrfiouhbau awkenvjbaovi"}
				if err := tmpl404.Execute(w, data); err != nil {
					http.Error(w, "Failed to render template", http.StatusInternalServerError)
					log.Printf("Template execution error: %v", err)
				}
			} else {*/
		w.Header().Set("Content-Type", "text/html")
		allRoutes, err := getAllPosts(db)
        fmt.Print(allRoutes)
		if err != nil {
            log.Fatal("HEEEEEEELLLLPPPPPP")
			return
		}
		ro := Routes{Title: "404", Pages: allRoutes}
		if err := tmpl404.Execute(w, ro); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template execution error: %v", err)
		}
		//}
	})
	http.HandleFunc("/blog", func(w http.ResponseWriter, r *http.Request) {
		data := BlogPageData{
			Title:       "My Blog",
			Heading:     "Welcome to My Blog",
			Description: "This is a simple blog page served by a Go server.",
			Posts: []Post{
				{Title: "First Post"},
				{Title: "Second Post"},
			},
		}

		// Render the template with data
		fmt.Print(r.PathValue("*"))
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template execution error: %v", err)
		}

	})

	http.HandleFunc("/add/page", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Retrieve other form fields (e.g., `title`)
		title := r.FormValue("title")
		slug := r.FormValue("slug")
		content := r.FormValue("content")
		now := time.Now().Unix()
		post := Post{Slug: slug, Title: title, Status: 0, DateCreated: now, DateModified: now, Content: content, Template: "Page"}
		_, err := createPost(db, post)
		message := "created successfully"
		if err != nil {
			message = "error creating post"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"message": message})
		if err != nil {
			return
		}
	})
	log.Println("\n\nServer is running at http://localhost:8080/blog")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
