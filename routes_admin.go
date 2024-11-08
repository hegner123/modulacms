package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func adminRouter(w http.ResponseWriter, r *http.Request) {
	fmt.Print(r.URL.Path)
	switch r.URL.Path {
	case "/admin/field/add":
		fmt.Print("/admin/field/add\n")
		AdminCreateField(w, r)
	}
}

func AdminCreateField(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("admin create field\n")
	//times := timestamp()
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}
	var field = Field{}

	field.ID = 0
	field.RouteID = 0
	form := r.ParseForm()
	fmt.Print(form)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "wip"})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
}
