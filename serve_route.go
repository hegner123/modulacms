package main

import (
	"fmt"
	"html/template"
	"os"

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

func servePageFromRoute(templatePaths []string) *template.Template {
	base := "./templates/"
	rf, err := os.ReadFile(base + templatePaths[0])
	if err != nil {
		logError("failed to find file ", err)
	}
	baseT := template.New("base")
	s := string(rf)
    fmt.Println(s)
	baseT, err = baseT.Parse(s)
	if err != nil {
		logError("failed to parse template ", err)
	}
	rf1, err := os.ReadFile(base + templatePaths[1])
	if err != nil {
		logError("failed to find file ", err)
	}
    menuT := baseT.New("menu")
	if err != nil {
		logError("failed to make menu Template", err)
	}
	s1 := string(rf1)
	_, err = menuT.Parse(s1)
	if err != nil {
		logError("failed to parse template ", err)
	}

	return baseT
}

func CreateTemplateTree() {}

// Search an adminRouteId

// fetch Datatypes that match dynamic adminRouteId

// fetch globalDataTypes.
