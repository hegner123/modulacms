package main
import (
	"fmt"
	"reflect"
	"strings"
)


func formatInsertFields(input interface{}) (string, int64) {
	val := reflect.ValueOf(input)
	typ := val.Type()

	if val.Kind() != reflect.Struct {
		return "", 0
	}

	var fields []string
	for i := 0; i < val.NumField(); i++ {
		fieldName := typ.Field(i).Name
        if strings.ToLower(fieldName) == "id" {
            continue
        }
		fields = append(fields, strings.ToLower(fieldName))
	}

	return "(" + strings.Join(fields, ",") + ")", int64(len(fields))
}

func formatGetFields(input interface{}) (string) {
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
    sliced := strings.Split(holders,"")
	return fmt.Sprintf("(%s)", strings.Join(sliced, ","))
}




func queryCreateBuilder(structure interface{},table string) string {
	columns, fieldCount := formatInsertFields(structure)
	valuesHold := generatePlaceholders(int(fieldCount))
	return fmt.Sprintf(`INSERT INTO %s %s VALUES %s`,table, columns, valuesHold)
}


func queryGetFilteredBuilder(structure interface{}, returnColumns []string) string {
    // Columns to return
    return fmt.Sprintf(``)

}

