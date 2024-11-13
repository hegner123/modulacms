package main

import (
	"encoding/json"
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
	fields, fieldTypes := formatGetStructFields(input) // Ensure this function's output suits your needs

	for i := 0; i < len(fields); i++ {
		switch fieldTypes[i].Kind() { // Using reflect.Kind for type checks
		case reflect.Int, reflect.Int32, reflect.Int64:
			if fields[i] == "id" {
				columns = append(columns, fields[i]+" INTEGER PRIMARY KEY")
				continue
			}
			columns = append(columns, fields[i]+" "+intType)
		case reflect.String:
			columns = append(columns, fields[i]+" "+stringType)
        case reflect.Struct:
            columns = append(columns, fields[i]+" "+stringType)
		}
	}

	return create + table + " (" + strings.Join(columns, ", ") + end
}

func formatGetStructFields(input interface{}) ([]string, []reflect.Type) {

	interfaceValue := reflect.ValueOf(input)
	interfaceType := reflect.TypeOf(input)
	var fieldNames []string
	var fieldTypes []reflect.Type

	if interfaceValue.Kind() != reflect.Struct {
		fmt.Printf("%v isn't a struct", input)
	}

	for i := 0; i < interfaceValue.NumField(); i++ {
		field := interfaceType.Field(i)
		if strings.ToLower(field.Name) == "id" {
			continue
		}
		fieldNames = append(fieldNames, strings.ToLower(field.Name))
		fieldTypes = append(fieldTypes, field.Type)
	}
	return fieldNames, fieldTypes
}

func formatSQLColumns(input interface{}) (string, int64) {
	fields, _ := formatGetStructFields(input)

	return "(" + strings.Join(fields, ", ") + ")", int64(len(fields))
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

func formatGetStructValues(s interface{}) string {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Struct {
		return ""
	}
	var values []string
	for i := 1; i < val.NumField(); i++ {
		fieldValue := val.Field(i)
		if fieldValue.Kind() == reflect.Struct {
			fmt.Printf("internal struct")

			j, err := json.Marshal(fieldValue.Interface())
			if err != nil {
				fmt.Printf("%s\n", err)
			}
			values = append(values, fmt.Sprintf("'%s'", string(j)))
			continue
		}
		newValue := fieldValue.Interface()
		if _, ok := newValue.(int32); ok {
			values = append(values, fmt.Sprintf("'%v'", fieldValue.Interface()))
		} else {
			values = append(values, fmt.Sprintf("'%v'", fieldValue.Interface()))
		}

	}

	return fmt.Sprintf("(%s);", strings.Join(values, ", "))
}

func FormatSqlInsertStatement(structure interface{}, table string) string {
	columns, _ := formatSQLColumns(structure)
	valuesHold := formatGetStructValues(structure)
	query := fmt.Sprintf(`INSERT INTO %s %s VALUES %s`, table, columns, valuesHold)
	return query
}

func queryGetFilteredBuilder(structure interface{}, returnColumns []string) string {
	table := reflect.ValueOf(structure).Type()
	fmt.Print(table)
	//columns,_ := formatSQLColumns(structure)
	// Columns to return
	return fmt.Sprintf(`SELECT %s FROM `, table)

}

func RemoveTLD(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ".")
	}
	return domain // return as-is if no TLD is found
}
