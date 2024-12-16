package mTemplate

import (
	"html/template"
)

func ParseTemplates(templateFiles []string) (*template.Template, error) {
	pageTemplate, err := template.ParseFiles(templateFiles...)
	if err != nil {
		logError("failed parsing files", err)
		return template.New(""), err
	}

	return pageTemplate, nil
}


