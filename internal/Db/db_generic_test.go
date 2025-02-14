package db

import (
	"fmt"
	"strings"
	"testing"
)

func TestFormatSqlItems(t *testing.T) {
	source1 := []string{"test1", "test2", "test3"}
	source2 := []string{"test1"}
	s1 := FormatSqlColumns(source1)
	s2 := FormatSqlColumns(source2)
	e1 := "test1, test2, test3"
	e2 := "test1"

	if s1 != e1 {
		fmt.Printf("\nSource1 test: %s\n", s1)
		t.Fatal("s1 fail")
	}
	if s2 != e2 {
		fmt.Printf("\nSource2 test: %s\n", s2)
		t.Fatal("s2 fail")
	}

}
func TestFormatSqlFilter(t *testing.T) {
	var (
		k1 string = "key1"
		k2 string = "key2"
		k3 string = "key3"
		v1 string = "value1"
		v2 string = "value2"
		v3 string = "value3"
	)
	source1 := map[string]string{k1: v1, k2: v2, k3: v3}
	source2 := map[string]string{k1: v1}
	s1 := FormatSqlFilter(source1)
	s2 := FormatSqlFilter(source2)
	e11 := "key1=value1"
	e12 := "key2=value2"
	e13 := "key3=value3"
	e14 := ","

	r1 := strings.Contains(s1, e11)
	r2 := strings.Contains(s1, e12)
	r3 := strings.Contains(s1, e13)
	r4 := string(s1[len(s1)-1]) != e14
	r5 := string(s1[11]) == e14
	r6 := string(s1[23]) == e14
	if !r1 || !r2 || !r3 || !r4 || !r5 || !r6 {
		fmt.Printf("\nr1: %v\n", r1)
		fmt.Printf("\nr2: %v\n", r2)
		fmt.Printf("\nr3: %v\n", r3)
		fmt.Printf("\nr4: %v\n", r4)
		fmt.Printf("\nr5: %v\n", r5)
		fmt.Printf("\nr6: %v\n", r6)

		t.Fatal("s1 fail")
	}

	e2 := "key1=value1"

	if s2 != e2 {
		fmt.Printf("\nSource2 test: %s\n", s2)
		t.Fatal("s2 fail")
	}

}


func TestInsertQuery(t *testing.T){
    table:="users"
    s1:=[]string{"id","username","email","hash"}
}
