package mTemplate

import (
	"fmt"
	"os"
	"testing"
)

func TestServeTemplate(t *testing.T) {
	t1, err := BuildTemplateStructFromRouteId(int64(1), "")
	templ, err := parseTemplateGlobs("./templates", "*.html")
	err = templ.Execute(os.Stdout, t1)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
