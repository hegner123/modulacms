package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join("public", r.URL.Path)
	fmt.Print(filePath)
	if filepath.Ext(filePath) == ".js" {

		w.Header().Set("Content-Type", "text/javascript")
	}
	http.ServeFile(w, r, filePath)
}

func checkAPIPath(rawURL string) (bool, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false, err
	}
	return strings.HasPrefix(parsedURL.Path, "/api/"), nil
}

func matchesPath(text, searchTerm string) bool {
	return strings.Contains(text, searchTerm)
}

func stripAPIPath(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	parsedURL.Path = strings.TrimPrefix(parsedURL.Path, "/api/")
	return parsedURL.String(), nil
}

func apiRoutes(w http.ResponseWriter, r *http.Request) {
	fmt.Print("api Route\n")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	apiRoute, err := stripAPIPath(r.URL.Path)
	if err != nil {
		fmt.Print("UM, this ain't a url bud.")
		fmt.Printf("\nerror: %s", err)
		return
	}
	switch {
	case matchesPath(apiRoute, "add/page"):
		res := apiCreatePost(w, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	case matchesPath(apiRoute, "get/posts"):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"message": "boom"})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	}
}

func handlePageRoutes(w http.ResponseWriter, r *http.Request) {

	db, err := getDb(Database{})
	if err != nil {
		fmt.Printf("\nerror: %s", err)
		return
	}
	matchedPost, err := matchAdminSlugToRoute(db, r.URL.Path)
	if err != nil {
		redirectTo404(w, r)
		fmt.Printf("\nerror: %s", r.URL.Path)
		fmt.Printf("\nerror: %s", err)
		return
	}
	/*
		adminPage := AdminPage{HtmlFirst: htmlFirst, Head: htmlHead, Body: matchedPost.Template, HtmlLast: htmlLast}
		adminTemplate := buildAdminTemplate(adminPage)

				fields, err := getPostFields(slugRoute, db)
				if err != nil {
					fmt.Printf("error: %s", err)
					return
				}
	*/
	tmp, err := template.ParseFiles("templates/" + matchedPost.Template)
	if err != nil {
		fmt.Printf("\nerror: %s", err)
		return
	}
	if err := tmp.Execute(w, nil); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
	/*
		if err := adminTemplate.Execute(w, nil); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			log.Printf("Template execution error: %v", err)
		}
	*/
}
func redirectTo404(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/404", http.StatusNotFound)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)

	tmp, err := template.ParseFiles("templates/404.html")
	if err != nil {
		fmt.Printf("\nerror: %s", err)
		return
	}
	if err := tmp.Execute(w, nil); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}
func handleWildcard(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Matched route with wildcard: %s", r.URL.Path)

	http.HandleFunc("/add/page", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		db, err := getDb(Database{})
		if err != nil {
			fmt.Printf("error: %s", err)
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

/*
func blogpage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/blog.html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	data := BlogPageData{
		Title:       r.URL.RawPath,
		Heading:     "Welcome to My Blog",
		Description: "This is a simple blog page served by a Go server.",
		Posts: []Post{
			{Title: r.Pattern},
			{Title: r.URL.Path},
		},
	}
	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}
*/
