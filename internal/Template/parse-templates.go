package mTemplate

import (
	"html/template"

	utility "github.com/hegner123/modulacms/internal/Utility"
)

func ParseTemplates(templateFiles []string) (*template.Template, error) {
	pageTemplate, err := template.ParseFiles(templateFiles...)
	if err != nil {
		utility.LogError("failed parsing files", err)
		return template.New(""), err
	}

	return pageTemplate, nil
}


