package main

import (
	"fmt"
	"reflect"
	"strings"
)

func formatCreateTable(input interface{}, table string) string {
	stringType := "TEXT"
	intType := "INTEGER"
	create := "CREATE TABLE IF NOT EXISTS "
	end := ");"
	var columns []string
	fields, fieldTypes := GetStructFields(input)
	for i := 0; i < len(fields); i++ {
		if fields[i] == "id" {
			columns = append(columns, fields[i]+" INTEGER PRIMARY KEY,")
            continue
		}
		switch fieldTypes[i].(type) {
		case int:
			columns = append(columns, fields[i]+intType)
		case string:
			columns = append(columns, fields[i]+stringType)
		}
	}

	return create + table + "(" + strings.Join(columns, " ,") + end
}

func GetStructFields(input interface{}) ([]string, []interface{}) {

	val := reflect.ValueOf(input)
	typ := val.Type()
	var fields []string
	var fieldTypes []interface{}

	if val.Kind() != reflect.Struct {
		fmt.Printf("%v isn't a struct", input)
	}

	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		fieldTypes = append(fieldTypes, typ.Field(i).Type)
		fields = append(fields, strings.ToLower(fieldName))
	}
	return fields, fieldTypes
}

func formatInsertColumns(input interface{}) (string, int64) {
	fields, _ := GetStructFields(input)

	return "(" + strings.Join(fields, ", ") + ")", int64(len(fields))
}

func formatGetFields(input interface{}) string {
	val := reflect.ValueOf(input)
	typ := val.Type()

	if val.Kind() != reflect.Struct {
		return ""
	}

	var fields []string
	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		fields = append(fields, strings.ToLower(fieldName))
	}

	return "(" + strings.Join(fields, ",") + ")"
}

func formatUpdateFields(input interface{}) string {
	val := reflect.ValueOf(input)
	typ := val.Type()

	if val.Kind() != reflect.Struct {
		return ""
	}

	var fields []string
	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
		fields = append(fields, fmt.Sprintf("%s = ?", strings.ToLower(fieldName)))
	}

	return strings.Join(fields, ", ")
}

func generatePlaceholders(n int) string {
	if n <= 0 {
		return "()"
	}
	holders := strings.Repeat("?", n)
	sliced := strings.Split(holders, "")
	return fmt.Sprintf("(%s)", strings.Join(sliced, ","))
}
func FormatStructValues(s interface{}) string {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Struct {
		return ""
	}
	var values []string
	for i := 1; i < val.NumField(); i++ {
		fieldValue := val.Field(i)
		newValue := fieldValue.Interface()
		if _, ok := newValue.(int32); ok {
			values = append(values, fmt.Sprintf("%v", fieldValue.Interface()))
		} else {
			values = append(values, fmt.Sprintf("\"%v\"", fieldValue.Interface()))
		}

	}

	return fmt.Sprintf("(%s)", strings.Join(values, ", "))
}

func queryCreateBuilder(structure interface{}, table string) string {
	columns, _ := formatInsertColumns(structure)
	valuesHold := FormatStructValues(structure)
	query := fmt.Sprintf(`INSERT INTO %s %s VALUES %s`, table, columns, valuesHold)
	return query
}

func queryGetFilteredBuilder(structure interface{}, returnColumns []string) string {
	// Columns to return
	return fmt.Sprintf(``)

}

func RemoveTLD(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ".")
	}
	return domain // return as-is if no TLD is found
}
