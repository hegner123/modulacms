package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
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

func createDbCopy(dbName string) (string, error) {
    times:= timestampS()
	backup := " testdb/backups/"
	base := "testdb/"
    db:=strings.TrimSuffix(dbName,".db")
	srcSQLName := backup + db + ".sql"

	dstDbName := base + "testing" + times + dbName
    _, err := os.Create(dstDbName)
	if err != nil {
		logError("failed to : ", err)
	}

	dstCmd := exec.Command("sqlite3", dstDbName, ".read " +srcSQLName )
    output, err := dstCmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Command failed: %s\n", err)
	}

	fmt.Printf("Command output:\n%s\n", output)
	if err != nil {
		return "", err
	}

	return dstDbName, nil
}
