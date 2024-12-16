package mTemplate

import (
	"html/template"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

type TemplateDataTree struct {
	Label        string
	AdminRouteId int64
	Child        *TemplateDataTree
	Parent       *TemplateDataTree
	Datatypes    []mdb.AdminDatatypes
	Fields       []mdb.AdminFields
}
type TestNested struct {
	Child    *TestNested
	Children []TestNested
	Parent   *TestNested
	Value    string
	Values   []string
}

func servePageFromRoute() (*template.Template, error) {
	// base := "./templates"
	// baseT, err := parseTemplateGlobs(base, "*.html")
	baseT := template.New("")
    baseT,err := baseT.ParseGlob("./templates/*.html")
	baseT,err = baseT.ParseGlob("./templates/partials/*.html")
	baseT,err = baseT.ParseGlob("./templates/ui/*.html")
	if err != nil {
		return nil, err
	}

	return baseT, nil
}

func CreateTemplateTree() {}

// Search an adminRouteId

// fetch Datatypes that match dynamic adminRouteId

// fetch globalDataTypes.
