package main

import (
	"html/template"
	"log"
)

type AdminPage struct {
	HtmlFirst string
	Head      string
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
    `

func buildAdminTemplate(page AdminPage) template.Template {
	mainTemplate := template.New("main")

	mainTemplate, err := mainTemplate.Parse(page.HtmlFirst)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	mainTemplate, err = mainTemplate.Parse(page.Head)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	mainTemplate, err = mainTemplate.ParseFiles("/template/" + page.Body)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	mainTemplate, err = mainTemplate.Parse(page.HtmlLast)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}
	return *mainTemplate
}
