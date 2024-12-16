package main

import (
	"fmt"
	"testing"
)

func makeTestBackup(output string, timestamp string) string {
	return fmt.Sprintf("%s_%s_TEST.zip", output, timestampS())
}

func TestMakeBackup(t *testing.T) {
	err := createBackup("modula.db", "public/media", "plugins/", "backups/", makeTestBackup)
	if err != nil {
		t.Errorf("%s\n", err)
	}
}
