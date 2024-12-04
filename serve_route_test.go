package main

import (
	"os"
	"testing"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)


func TestServeTemplate(t *testing.T) {
	testTemplate := mdb.AdminRoutes{Template: "modula_base_test.html"}
	t1 := TestNested{Value: "t1 Parent"}
	t2 := TestNested{Parent: &t1, Value: "t2 Child"}
	t1.Child = &t2
	tmpl := servePageFromRoute(testTemplate.Template.(string))
	err := tmpl.Execute(os.Stdout, t1)
	if err != nil {
		t.FailNow()
	}
}
