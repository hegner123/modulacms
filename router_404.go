package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

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
