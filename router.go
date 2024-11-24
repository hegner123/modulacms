package main

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

)



func apiRoutes(w http.ResponseWriter, r *http.Request) {
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
	case matchesPath(apiRoute, "create/route"):
		res := apiCreateRoute(w, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	case matchesPath(apiRoute, "create/user"):
		res := apiCreateUser(w, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(map[string]string{"result": res})
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
	case matchesPath(apiRoute, "list/routes"):
        routes,err := apiListRoutes()
        if err != nil { 
            logError("failed to list Routes: ", err)
        }
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(routes)
		if err != nil {
			fmt.Printf("\nerror: %s", err)
			return
		}
    case matchesPath(apiRoute,"list/fieldsbyroute"):
        fields ,err := apiListFieldsForRoute(w,r)
        if err != nil { 
            logError("failed to get fields : ", err)
        }
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(fields)
	}
}

func handlePageRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		return
	}

	db, ctx, err := getDb(Database{})
	if err != nil {
		fmt.Printf("\nerror: %s", err)
		return
	}
	matchedRoute := dbGetRoute(db, ctx, r.URL.Path)
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
	tmp, err := template.ParseFiles("templates/" + matchedRoute.Template.String)
	if err != nil {
		fmt.Printf("\nerror: %s", err)
		return
	}
	fields := dbListFieldsByRoute(db, ctx, matchedRoute.ID)

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

func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	filePath := filepath.Join("public", r.URL.Path)
	fmt.Print(filePath)
	if filepath.Ext(filePath) == ".js" {

		w.Header().Set("Content-Type", "text/javascript")
	}
	http.ServeFile(w, r, filePath)
}
