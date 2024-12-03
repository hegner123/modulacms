package main

import (
	"fmt"
	"os"
	"testing"

	mdb "github.com/hegner123/modulacms/db-sqlite"
)





func TestServeTemplate(t *testing.T) {
	testTemplate := mdb.AdminRoute{Template: "modula_base.html"}
	testDT := mdb.AdminDatatype{Label: "TestNamedArg"}
	testData := TemplateDataTree{Label: "Modula Base Template"}
	testData.Datatypes[0] = testDT
	t1 := TestNested{Value: "t1 Parent"}
	t2 := TestNested{Parent: &t1, Value: "t2 Child"}
	t1.Child = &t2
	fmt.Println("TestServeTemplate")
	tmpl := servePageFromRoute(testTemplate.Template.(string))
	err := tmpl.Execute(os.Stdout, t1)
	if err != nil {
		logError("failed to : ", err)
		t.FailNow()
	}
}
