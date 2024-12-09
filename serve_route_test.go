package main

import (
	"fmt"
	"os"
	"testing"
)

func TestServeTemplate(t *testing.T) {
	t1 := BuildTemplateStructFromRouteId(int64(1))
    fmt.Print("test serve template")
    templ,err := parseTemplateGlobs("./templates","*.html")
    if err != nil { 
        logError("failed to : ", err)
    }
	err = templ.Execute(os.Stdout, t1)
	if err != nil {
		t.FailNow()
	}
}
