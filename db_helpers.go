package main

import (
	"database/sql"
	"fmt"
)

func forEachStatement(db *sql.DB, statements []string) error {
	for i := 0; i < len(statements); i++ {
		_, err := db.Exec(statements[i])
		fmt.Printf("\n%d", i)
		if err != nil {
			fmt.Print(err)
			return err
		}
	}
	return nil
}
