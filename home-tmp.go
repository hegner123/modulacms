package main

import (
	"html/template"
    "net/http"
)

type PageData struct {
	Title   string
	Content string
}

// Handler to serve HTML directly without creating a file
func serveHTML(w http.ResponseWriter, r *http.Request) {
	// Define the HTML template
	const tmpl = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
</head>
<body>
    <h1>{{.Content}}</h1>
</body>
</html>`

	// Parse the template
	t, err := template.New("webpage").Parse(tmpl)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Define the data for the HTML template
	data := PageData{
		Title:   "Sample Page",
		Content: "Hello, World!",
	}

	// Set the content type to HTML
	w.Header().Set("Content-Type", "text/html")

	// Execute the template and write it to the response writer
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

