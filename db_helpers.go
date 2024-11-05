package main

import (
	"database/sql"
	"fmt"
)

func forEachStatement(db *sql.DB, statements []string, label string) error {
	for i := 0; i < len(statements); i++ {
		_, err := db.Exec(statements[i])
		if err != nil {
			fmt.Print(err)
			return err
		}
	}
    fmt.Printf("%s inserted successfully\n",label)
	return nil
}
