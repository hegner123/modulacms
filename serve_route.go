package main

import (
	"html/template"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)

func servePageFromRoute(route mdb.Adminroute) *template.Template {
    t,err:=template.ParseGlob("./templates")
    if err != nil { 
        logError("failed to parseTemplate", err)
    }
        
	return t
}
