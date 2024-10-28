package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

func handlePageRoutes(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()


	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/api/", http.HandlerFunc(handleWildcard)).ServeHTTP(w, r)
	})

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
	fmt.Printf("page route")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		db, err := getDb()
		if err != nil {
			return
		}
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
}

func handleWildcard(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Matched route with wildcard: %s", r.URL.Path)

	http.HandleFunc("/add/page", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		db, err := getDb()
		if err != nil {
			return
		}

		// Retrieve other form fields (e.g., `title`)
		title := r.FormValue("title")
		slug := r.FormValue("slug")
		content := r.FormValue("content")
		now := time.Now().Unix()
		post := Post{Slug: slug, Title: title, Status: 0, DateCreated: now, DateModified: now, Content: content, Template: "Page"}
		_, err = createPost(db, post)
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
}
