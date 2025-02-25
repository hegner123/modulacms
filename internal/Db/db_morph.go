package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

func Ns(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func Ni(i int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func Ni64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func Nb(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

func ns(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func ni(i int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func ni64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func nb(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

func getColumnValue(column string, s interface{}) string {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if strings.EqualFold(field.Name, column) {
			fieldValue := v.Field(i)
			if fieldValue.Kind() == reflect.String {
				return fieldValue.String()
			}
			return fmt.Sprintf("%v", fieldValue.Interface())
		}
	}

	return ""
}
