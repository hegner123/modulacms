package db
import (
	"fmt"
	"time"
    "strings"
    "os"
    "os/exec"
)
type ANSIColor string

const (
	RED    ANSIColor = "\033[31m"
	GREEN  ANSIColor = "\033[32m"
	YELLOW ANSIColor = "\033[33m"
	BLUE   ANSIColor = "\033[34m"
	RESET  ANSIColor = "\033[0m" // Reset
)

func timestampI() int64 {
	return time.Now().Unix()
}

func timestampS() string {
	return fmt.Sprint(time.Now().Unix())
}
func logError(message string, err error, args ...any) {
	var messageParts []string
	messageParts = append(messageParts, message)
	for _, arg := range args {
		switch v := arg.(type) {
		case fmt.Stringer: // If the type implements fmt.Stringer, use String()
			messageParts = append(messageParts, v.String())
		default:
			messageParts = append(messageParts, fmt.Sprintf("%+v", arg)) // Format structs nicely
		}
	}
	fullMessage := strings.Join(messageParts, " ")

	// Format the final error message
	er := fmt.Errorf("%sErr: %s\n%v\n%s", RED, fullMessage, err, RESET)
	if er != nil {
		fmt.Printf("%s\n", er)
	}
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
		logError("couldn't create file", err)
	}

	dstCmd := exec.Command("sqlite3", dstDbName, ".read " +srcSQLName )
    _, err = dstCmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Command failed: %s\n", err)
	}

	if err != nil {
		return "", err
	}

	return dstDbName, nil
}
