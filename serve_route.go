package main

import (
	"html/template"
)

func servePageFromRoute(templatePath string) *template.Template {
	base := "./templates/"
	concat := base + templatePath
	t, err := template.ParseGlob(concat)
	if err != nil {
		logError("failed to parseTemplate", err)
	}

	return t
}
