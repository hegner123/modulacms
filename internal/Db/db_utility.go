package db

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func CopyDb(dbName string, useDefault bool) (string, error) {
	times := TimestampS()
	backup := "../../testdb/backups/"
	base := "../../testdb/"
	db := strings.TrimSuffix(dbName, ".db")
	srcSQLName := backup + db + ".sql"

	dstDbName := base + "testing" + times + dbName
	_, err := os.Create(dstDbName)
	if err != nil {
		fmt.Printf("Couldn't create file")
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
