package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

)



func ns(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func ni(i int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}

func ni64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(i), Valid: true}
}
/*
func nf(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: true}
}
*/
func nb(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}
/*
func nt(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}

func nby(by byte) sql.NullByte {
	return sql.NullByte{Byte: by, Valid: true}
}
*/
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




