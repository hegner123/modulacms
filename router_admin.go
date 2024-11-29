package main

import (
	"fmt"
	"net/http"
)

func handleAdminRoutes(w http.ResponseWriter, r *http.Request, segments []string) {
	pLog("admin router")
	db, ctx, err := getDb(Database{})
	if err != nil {
		fmt.Printf("\nerror: %s", err)
		return
	}
	defer db.Close()
	s := r.URL.Path
	pLog(s)
	w.Header().Set("Content-Type", "text/html")
	route := dbGetAdminRoute(db, ctx, s)
	s, ok := route.Template.(string)
	if !ok {
		return
	} 
	res := servePageFromRoute(s)
	err = res.ExecuteTemplate(w, s, res)
	if err != nil {
		logError("failed to write response : ", err)
	}
}

func handleAdminAuth(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logError("failed to ParseForm ", err)
	}
	// status, err := handleAuth(r.Form)
	if err != nil {
		logError("failed to handle auth: ", err)
	}
	w.Header().Set("Content-Type", "application/json")
}
