package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"log"
)

type AdminPage struct {
	HtmlFirst string
	Head      string
	Menu      string
	Body      string
	HtmlLast  string
	Template  string
	Fields    []Field
}

const htmlFirst string = `
<!DOCTYPE html>
<html lang="en">
    `
const htmlLast string = `
    </body>
    </html>
    `

const htmlHead string = `
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <meta http-equiv="Content-Security-Policy" content="
                default-src 'self'; 
                script-src 'self' 'wasm-unsafe-eval' 'unsafe-inline' 'unsafe-eval' chrome-extension:;
                style-src 'self' 'unsafe-inline'; 
                connect-src 'self' chrome-extension:; 
                img-src 'self' chrome-extension:;
                font-src 'self' chrome-extension:;
                object-src 'none'; 
                frame-ancestors 'self'; 
                base-uri 'self';
              ">

            <style>
                html {
                    font-family: sans-serif, system-ui;
                    color: white;
                }

                body {
                    background: #333;
                    padding-left: 2rem;
                    padding-right: 2rem;

                }

                .form-row {
                    display: flex;
                }
            </style>
            <title>{{.Title}}</title>
        </head>
    <body>
    `

func buildAdminTemplate(page AdminPage) template.Template {
	db, err := getDb(Database{})
	if err != nil {
		fmt.Printf("\n%s", err)
	}
	mainTemplate := template.New("main")

	mainTemplate, err = mainTemplate.Parse(page.HtmlFirst)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	mainTemplate, err = mainTemplate.Parse(page.Head)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	menu := buildAdminMenu(db)
	fmt.Printf("\n%s", menu)
	mainTemplate, err = mainTemplate.Parse(menu)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
    /*
	mainTemplate, err = mainTemplate.ParseFiles("templates/" + page.Body)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
*/
	mainTemplate, err = mainTemplate.Parse(page.HtmlLast)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	return *mainTemplate
}

type Link struct {
	Slug  string
	Title string
}

type MenuLinks struct {
	Links []Link
}

func buildAdminMenu(db *sql.DB) string {
	var menuBuffer = bytes.Buffer{}
	var menu MenuLinks
	const link string = `<a href="%s">%s</a>`
	posts, err := getAllAdminRoutes(db)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	for i := 0; i < len(posts); i++ {
        link := Link{Slug: posts[i].Slug,Title:  posts[i].Title}
		menu.Links = append(menu.Links, link)
	}
	menuTemplate, err := template.ParseFiles("templates/menu.html")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	err = menuTemplate.Execute(&menuBuffer, menu)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	return menuBuffer.String()
}
