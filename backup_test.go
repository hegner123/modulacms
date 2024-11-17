package main

import (
	"testing"
)

func TestMakeBackup(t *testing.T) {
    
    err := createBackup("modula_test.db", "public/media", "plugins/", "backups/")
    if err!=nil {
        t.Errorf("%s\n",err)
    }

}

