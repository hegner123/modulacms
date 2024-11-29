package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)


func handleClientRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		return
	}

	db, ctx, err := getDb(Database{})
	if err != nil {
		fmt.Printf("\nerror: %s", err)
		return
	}
    defer db.Close()
	matchedRoute := dbGetRoute(db, ctx, r.URL.Path)
	if err != nil {
		redirectTo404(w, r)
		fmt.Printf("\nerror: %s", r.URL.Path)
		fmt.Printf("\nerror: %s", err)
		return
	}
	fields := dbJoinDatatypeByRoute(db, ctx, matchedRoute.ID)

		err = json.NewEncoder(w).Encode(fields)
		http.Error(w, "Failed to encode json", http.StatusInternalServerError)
	

}
