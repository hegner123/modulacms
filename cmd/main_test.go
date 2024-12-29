package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	utility "github.com/hegner123/modulacms/internal/Utility"
)

type GlobalTestingState struct {
	Initialized bool
}

var globalTestingState GlobalTestingState

func setup() {
	fmt.Printf("TestMain setup\n")
	globalTestingState.Initialized = true
}

func teardown() {
	fmt.Printf("TestMain teardown\n")
	globalTestingState.Initialized = false
}

func TestMain(m *testing.M) {
	fmt.Printf("TestMain init\n")
	globalTestingState.Initialized = false
	setup()
	code := m.Run()
	teardown()
	fmt.Printf("TestMain exit\n")
	os.Exit(code)
}

func CreateDbCopy(dbName string, useDefault bool) (string, error) {
	times := utility.TimestampS()
	backup := " testdb/backups/"
	base := "testdb/"
	db := strings.TrimSuffix(dbName, ".db")
	srcSQLName := backup + db + ".sql"

	dstDbName := base + "testing" + times + dbName
	_, err := os.Create(dstDbName)
	if err != nil {
		utility.LogError("couldn't create file", err)
	}
	if useDefault {
		srcSQLName = backup + "test.sql"
	}

	dstCmd := exec.Command("sqlite3", dstDbName, ".read "+srcSQLName)
	_, err = dstCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Command failed: %s\n", err)
	}

	if err != nil {
		return "", err
	}

	return dstDbName, nil
}
