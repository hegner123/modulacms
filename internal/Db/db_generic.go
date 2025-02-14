package db

import (
	"fmt"

	utility "github.com/hegner123/modulacms/internal/Utility"
)

func FormatSqlColumns(columns []string) string {
	var c string
	for i := 0; i < len(columns); i++ {
		c += columns[i] + ", "
	}
	c = utility.TrimStringEnd(c, 1)
	return c
}

func FormatSqlFilter(kv map[string]string) string {
	var f string
	for k, v := range kv {
		f += k + "=" + v + ","
	}
	f = utility.TrimStringEnd(f, 1)
	return f
}

func FormatSqlSet(kv map[string]string) string {
	return FormatSqlFilter(kv)
}

func InsertQuery(table string, columns string, values string) string {
	return fmt.Sprintf("INSERT INTO %s (%v) VALUES (%v);", table, columns, values)
}

func SelectQuery(table string, columns string, filter string) string {
	return fmt.Sprintf("SELECT %s FROM %s WHERE %s;", table, columns, filter)
}

func UpdateQuery(table string, set string, filter string) string {
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s; ", table, set, filter)

}

func DeleteQuery(table string, filter string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s;", table, filter)
}
