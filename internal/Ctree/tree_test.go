package ctree

import (
	"fmt"
	"testing"
)

var TreeTestTable string

func TestTreeParse(t *testing.T) {
	b, err := BuildTree("tree_test.db")
	if err != nil {
		t.Fatal(err)
	}
    fmt.Println(string(b))

}
