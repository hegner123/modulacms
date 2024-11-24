package main

import (
	"net/http"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)
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

func apiGetField(w http.ResponseWriter, r *http.Request) (mdb.Field, error) {
	f := mdb.Field{}
	return f, nil
}
func apiGetUser(w http.ResponseWriter, r *http.Request) (mdb.User, error) {
	u := mdb.User{}
	return u, nil
}


func apiGetMedia(w http.ResponseWriter, r *http.Request) (mdb.Media, error) {
	m := mdb.Media{}
	return m, nil
}
