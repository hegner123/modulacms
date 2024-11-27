package main

import (
	"fmt"
	"net/http"
)

func handleAdminRoutes(w http.ResponseWriter, r *http.Request) {
	db, ctx, err := getDb(Database{})
	if err != nil {
		fmt.Printf("\nerror: %s", err)
		return
	}
	defer db.Close()
	w.Header().Set("Content-Type", "text/html")
	route := dbGetAdminRoute(db, ctx, r.URL.Path)
	res := servePageFromRoute(route)
	_, err = w.Write(res)
	if err != nil {
		logError("failed to write response : ", err)
	}
}
