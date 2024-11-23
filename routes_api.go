package main

import (
	"net/http"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func apiCreateRoute(w http.ResponseWriter, r *http.Request) string {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return "Couldn't parse form"
	}
	db, ctx, err := getDb(Database{})
	if err != nil {
		return "Couldn't get db"
	}
	title := r.FormValue("title")
	slug := r.FormValue("slug")
	content := r.FormValue("content")
	route := mdb.CreateRouteParams{
		Slug:         ns(slug),
		Title:        ns(title),
		Status:       ni(0),
		Datecreated:  ni(int(timestampI())),
		Datemodified: ni(int(timestampI())),
		Content:      ns(content),
		Template:     ns("page.html")}
	_ = dbCreateRoute(db, ctx, route)
	message := "created successfully"
	return message
}

func apiGetAllRoutes() ([]mdb.Route, error) {
	fetchedRoutes := []mdb.Route{}
	db, ctx, err := getDb(Database{})
	if err != nil {
		return fetchedRoutes, err
	}

	fetchedRoutes = dbListRoute(db, ctx)
	return fetchedRoutes, nil
}

func apiGetRoute(w http.ResponseWriter, r *http.Request) (mdb.Route, error) {
	fetchedRoute := mdb.Route{}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return fetchedRoute, err
	}
	db, ctx, err := getDb(Database{})
	if err != nil {
		return fetchedRoute, err
	}
	routeSlug := r.FormValue("slug")
	if err != nil {
		return fetchedRoute, err
	}
	fetchedRoute = dbGetRoute(db, ctx, routeSlug)
	return fetchedRoute, nil
}
func apiUpdateRoute() {}
func apiDeleteRoute() {}

func apiGetField()             {}
func apiGetAllFieldsForRoute() {}
func apiUpdateField()          {}

func apiGetUser()     {}
func apiAuthUser()    {}
func apiGetAllUsers() {}
func apiUpdateUser()  {}
func apiDeleteUser()  {}

func apiCreateMedia()  {}
func apiGetMedia()     {}
func apiGetAllMedias() {}
