package main

import (
	"encoding/json"
	"fmt"
	"html"
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
		res := apiCreateRoute(w, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	case matchesPath(apiRoute, "get/routes"):
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
	if r.Method == http.MethodPost {
		adminRouter(w, r)
		return
	}

	db, err := getDb(Database{})
	if err != nil {
		fmt.Printf("\nerror: %s", err)
		return
	}
	matchedRoute, err := matchAdminSlugToRoute(db, r.URL.Path)
	if err != nil {
		redirectTo404(w, r)
		fmt.Printf("\nerror: %s", r.URL.Path)
		fmt.Printf("\nerror: %s", err)
		return
	}
	// First we create a FuncMap with which to register the function.
	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"html": html.EscapeString,
	}
	/*
		adminPage := AdminPage{HtmlFirst: htmlFirst, Head: htmlHead, Body: matchedRoute.Template, HtmlLast: htmlLast}
		adminTemplate := buildAdminTemplate(adminPage)

				fields, err := getRouteFields(slugRoute, db)
				if err != nil {
					fmt.Printf("error: %s", err)
					return
				}
	*/
	tmp, err := template.ParseFiles("templates/" + matchedRoute.Template)
	if err != nil {
		fmt.Printf("\nerror: %s", err)
		return
	}
	fields, err := dbGetField(db, matchedRoute.ID)
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	if err := tmp.Funcs(funcMap).Execute(w, fields); err != nil {
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

	http.HandleFunc("/add/route", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		db, err := getDb(Database{})
		if err != nil {
			fmt.Printf("error: %s", err)
			return
		}

		title := r.FormValue("title")
		slug := r.FormValue("slug")
		content := r.FormValue("content")
		now := time.Now().Unix()
		Route := Routes{Slug: slug, Title: title, Status: 0, DateCreated: now, DateModified: now, Content: content, Template: "Page"}
		_, err = dbCreateRoute(db, Route)
		message := "created successfully"
		if err != nil {
			message = "error creating route"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"message": message})
		if err != nil {
			return
		}
	})
}
