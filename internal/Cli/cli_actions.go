package cli

import "fmt"

func Cli_create(table string, columns string, values string) {
	q := fmt.Sprintf("INSERT INTO %s (%v) VALUES (%v);",table, columns, values)
	fmt.Print(q)
}

func Cli_update() {}

func Cli_read() {}

func Cli_delete() {}

func Cli_list() {}
