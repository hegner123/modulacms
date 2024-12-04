package main

import (
	"html/template"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

type TemplateDataTree struct {
	Label        string
	AdminRouteId int64
	Child        *TemplateDataTree
	Parent       *TemplateDataTree
	Datatypes    []mdb.AdminDatatype
	Fields       []mdb.AdminField
}
type TestNested struct {
	Child  *TestNested
    Children []TestNested
	Parent *TestNested
	Value  string
	Values []string
}

func servePageFromRoute(templatePath string) *template.Template {
	base := "./templates/"
	concat := base + templatePath
	t, err := template.ParseGlob(concat)
	if err != nil {
		logError("failed to parseTemplate", err)
	}

	return t
}

func CreateTemplateTree(){}


// Search an adminRouteId

// fetch Datatypes that match dynamic adminRouteId 

// fetch globalDataTypes.
