package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func apiCreateUser(w http.ResponseWriter, r *http.Request) string {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return "Couldn't parse form"
	}
	db, ctx, err := getDb(Database{})
	if err != nil {
		return "Couldn't get db"
	}
	username := r.FormValue("username")
	email := r.FormValue("email")
	name := r.FormValue("name")
	hash := r.FormValue("hash")
	role := r.FormValue("role")
	user := mdb.CreateUserParams{
		Datecreated:  ns(timestampS()),
		Datemodified: ns(timestampS()),
        Username: ns(username),
        Email: ns(email),
        Name: ns(name),
        Hash: ns(hash),
        Role: ns(role),
    }
	_ = dbCreateUser(db, ctx, user)
	message := "created successfully"
	return message
}

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
		Datecreated:  ns(timestampS()),
		Datemodified: ns(timestampS()),
		Content:      ns(content),
		Template:     ns("page.html")}
	_ = dbCreateRoute(db, ctx, route)
	message := "created successfully"
	return message
}

func apiCreateMedia(w http.ResponseWriter, r *http.Request) (mdb.Media, error) {
	m := mdb.Media{}
	return m, nil
}
func apiCreateField(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("admin create field\n")
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	var field = mdb.Field{}

	field.ID = 0
	field.Routeid = 0
	form := r.ParseForm()
	fmt.Print(form)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "wip"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}
