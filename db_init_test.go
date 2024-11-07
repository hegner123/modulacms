package main

import (
	"fmt"
	"testing"
)


func TestDbInit(t *testing.T) {
    _,err := getDb(Database{DB: "modula_test.db"})
    if err!=nil {
        fmt.Printf("%s\n",err)
    }
    
}

