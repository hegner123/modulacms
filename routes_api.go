package main

import (
	"net/http"
	"strconv"
	"time"
)

func apiCreateRoute(w http.ResponseWriter, r *http.Request) string {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return "Couldn't parse form"
	}
	db, err := getDb(Database{})
	if err != nil {
		return "Couldn't get db"
	}
	title := r.FormValue("title")
	slug := r.FormValue("slug")
	content := r.FormValue("content")
	now := time.Now().Unix()
	route := Routes{Slug: slug, Title: title, Status: 0, DateCreated: now, DateModified: now, Content: content, Template: "page.html"}
	_, err = createRoute(db, route)
	message := "created successfully"
	if err != nil {
		message = "error creating route"
	}
	return message
}
func apiGetAllRoutes() ([]Routes, error) {
	fetchedRoutes := []Routes{}
	db, err := getDb(Database{})
	if err != nil {
		return fetchedRoutes, err
	}

	fetchedRoutes, err = getAllRoutes(db)

	return fetchedRoutes, nil
}
func apiGetRoute(w http.ResponseWriter, r *http.Request) (Routes, error) {
	fetchedRoute := Routes{}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return fetchedRoute, err
	}
	db, err := getDb(Database{})
	if err != nil {
		return fetchedRoute, err
	}
	routeIdForm := r.FormValue("routeid")
	routeId, err := strconv.ParseInt(routeIdForm, 10, 32)
	if err != nil {
		return fetchedRoute, err
	}
	fetchedRoute, err = getRouteById(db, int(routeId))
	if err != nil {
		return fetchedRoute, err
	}
	return fetchedRoute, nil
}
func apiUpdateRoute() {}
func apiDeleteRoute() {}

func apiGetField()            {}
func apiGetAllFieldsForRoute() {}
func apiUpdateField()         {}

func apiGetUser()     {}
func apiAuthUser()    {}
func apiGetAllUsers() {}
func apiUpdateUser()  {}
func apiDeleteUser()  {}

func apiCreateMedia()  {}
func apiGetMedia()     {}
func apiGetAllMedias() {}
