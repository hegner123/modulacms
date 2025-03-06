package db

import (
	"fmt"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

type SqlLogicOperators string
type SqlMethods int

const (
	INSERT SqlMethods        = 1
	SELECT SqlMethods        = 2
	UPDATE SqlMethods        = 3
	DELETE SqlMethods        = 4
	AND    SqlLogicOperators = "AND"
	OR     SqlLogicOperators = "OR"
)

type SQLQuery struct {
	Method  int
	Table   string
	Columns []string
	Values  []string
	Filter  []WhereKeyValue
	Err     error
}

type WhereKeyValue struct {
	key    string
	value  string
	method *SqlLogicOperators
}


func FormatSqlColumns(columns []string, isValues bool) string {
	var v string
	var c string
	if isValues {
		values := columns
		for i := 0; i < len(values); i++ {
			if utility.IsInt(values[i]) {
				v += values[i] + ","
			} else {
				v += "'" + values[i] + "'" + ","
			}
		}
		v = utility.TrimStringEnd(v, 1)
		return v
	} else {
		for i := 0; i < len(columns); i++ {
			c += columns[i] + ","
		}
		c = utility.TrimStringEnd(c, 1)
		return c
	}
}

func FormatSqlFU(wkv []WhereKeyValue, isUpdate bool) string {
	var f string
	for i := 0; i < len(wkv); i++ {
		var v1 string
		if utility.IsInt(wkv[i].value) {
			v1 = wkv[i].value
		} else {
			v1 = "'" + wkv[i].value + "'"
		}
		if wkv[i].method != nil {
			f += " " + string(*wkv[i].method) + " " + wkv[i].key + "=" + v1
			if isUpdate {
				f += ","
			}
		} else {
			f += wkv[i].key + "=" + v1
			if isUpdate {
				f += ","
			}
		}
	}
	if isUpdate {
		f = utility.TrimStringEnd(f, 1)
	}
	return f
}

func FormatSqlSet(wkv []WhereKeyValue) string {
	return FormatSqlFU(wkv, true)
}

func InsertQuery(table string, columns string, values string) string {
	return fmt.Sprintf("INSERT INTO %s (%v) VALUES (%v);", table, columns, values)
}

func SelectQuery(table string, columns string, filter string) string {
	return fmt.Sprintf("SELECT (%s) FROM %s WHERE %s;", columns, table, filter)
}

func UpdateQuery(table string, set string, filter string) string {
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s; ", table, set, filter)
}

func DeleteQuery(table string, filter string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s;", table, filter)
}
