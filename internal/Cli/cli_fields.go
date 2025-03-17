package cli

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

func NewFieldFromType(column string, colType *sql.ColumnType) (huh.Field, error) {
	if strings.Contains(column, "date_created") || strings.Contains(column, "date_modified") || strings.Contains(column, "history") {
		return nil, nil
	}
	var field huh.Field
	switch colType.DatabaseTypeName() {
	case "TEXT":
		if nullable, _ := colType.Nullable(); nullable {
			return huh.NewInput().Title(column).Validate(Required), nil
		} else {
			return huh.NewInput().Title(column), nil
		}
    case "INTEGER":
		if nullable, _ := colType.Nullable(); nullable {
			return huh.NewInput().Title(column).Validate(Required), nil
		} else {
			return huh.NewInput().Title(column), nil
		}

	}
	return field, nil
}
func NewUpdateFieldFromType(column string, colType *sql.ColumnType, value string) (huh.Field, error) {
	if strings.Contains(column, "DateCreated") || strings.Contains(column, "DateModified") || strings.Contains(column, "History") {
		return nil, nil
	}
	var field huh.Field
	switch colType.DatabaseTypeName() {
	case "TEXT":
		if nullable, _ := colType.Nullable(); nullable {
			return huh.NewInput().Title(column).Value(&value).Placeholder(value).Validate(Required), nil
		} else {
			return huh.NewInput().Title(column).Value(&value).Placeholder(value), nil
		}
    case "INTEGER":
		if nullable, _ := colType.Nullable(); nullable {
			return huh.NewInput().Title(column).Value(&value).Placeholder(value).Validate(Required), nil
		} else {
			return huh.NewInput().Title(column).Value(&value).Placeholder(value), nil
		}
	}
	return field, nil
}

func Required(s string) error {
	if len(s) < 1 {
		return fmt.Errorf("Input Cannot Be Null")
	} else {
		return nil
	}

}

// ColumnTyper defines the subset of sql.ColumnType methods needed.
type ColumnTyper interface {
	Name() string
	DatabaseTypeName() string
	Length() (int64, bool)
	Nullable() (bool, bool)
	DecimalSize() (int64, int64, bool)
}

// testColumnType is a stub implementation of the ColumnTyper interface.
type testColumnType struct {
	name           string
	dbType         string
	length         int64
	hasLength      bool
	nullable       bool
	hasNullable    bool
	precision      int64
	scale          int64
	hasDecimalSize bool
}

func (t testColumnType) Name() string {
	return t.name
}

func (t testColumnType) DatabaseTypeName() string {
	return t.dbType
}

func (t testColumnType) Length() (int64, bool) {
	return t.length, t.hasLength
}

func (t testColumnType) Nullable() (bool, bool) {
	return t.nullable, t.hasNullable
}

func (t testColumnType) DecimalSize() (int64, int64, bool) {
	return t.precision, t.scale, t.hasDecimalSize
}

func NewFieldFromTypeTest(column string, colType ColumnTyper) (huh.Field, error) {
	fmt.Println("Column", column)
	fmt.Println("ColType", colType.DatabaseTypeName())
	if strings.Contains(column, "DateCreated") || strings.Contains(column, "DateModified") || strings.Contains(column, "History") {
		return nil, nil
	}
	var field huh.Field
	switch colType.DatabaseTypeName() {
	case "TEXT":
		if nullable, _ := colType.Nullable(); nullable {
			return huh.NewInput().Title(column).Validate(Required), nil
		} else {
			return huh.NewInput().Title(column), nil
		}

	}
	return field, nil
}
