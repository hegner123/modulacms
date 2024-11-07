package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func adminRouter(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/admin/field/add":
		fmt.Print("/admin/field/add\n")
		AdminCreateField(w, r)
	}
}


func AdminCreateField(w http.ResponseWriter, r *http.Request) {
    fmt.Printf("admin create field\n")
	times := timestamp()
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	var field = Field{}
    
    field.ID = 0
	field.RouteID = 0
	field.Author = r.FormValue("author")
	field.AuthorID = r.FormValue("authorId")
	field.Key = r.FormValue("key")
	field.Data = r.FormValue("data")
	field.DateCreated = times
	field.DateModified = times
	field.Component = r.FormValue("component")
	field.Tags = r.FormValue("tags")
	field.Parent = r.FormValue("parent")
	res := dbCreateField(field)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"message": res})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}
