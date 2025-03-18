package cli

import (
	"database/sql"
	"strings"

	"github.com/charmbracelet/huh"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

func (m model) NewFieldFromType(column string, colType *sql.ColumnType, value *string) (huh.Field, error) {
	if strings.Contains(column, "date_created") || strings.Contains(column, "date_modified") || strings.Contains(column, "history") {
		switch column {
		case "date_created":
			m.formMap[column] = utility.TimestampReadable()
		case "date_modified":
			m.formMap[column] = utility.TimestampReadable()
		case "history":
			m.formMap[column] = ""
		}
		return nil, nil
	}
	var field huh.Field
	res := m.GetSuggestionsString(column)
	switch colType.DatabaseTypeName() {
	case "TEXT":
		if nullable, _ := colType.Nullable(); nullable {
			i := huh.NewInput().Title(column).Key(column).Value(value)
			if res != nil {
				i.Suggestions(res)
			}
			return i, nil
		} else {
			i := huh.NewInput().Title(column).Key(column).Value(value).Validate(Required)
			if res != nil {
				i.Suggestions(res)
			}
			return i, nil
		}
	case "INTEGER":
		if nullable, _ := colType.Nullable(); nullable {
			i := huh.NewInput().Title(column).Key(column).Value(value)
			if res != nil {
				i.Suggestions(res)
			}
			return i, nil
		} else {
			i := huh.NewInput().Title(column).Key(column).Value(value).Validate(Required)
			if res != nil {
				i.Suggestions(res)
			}
			return i, nil
		}

	}
	return field, nil
}
func (m model) NewUpdateFieldFromType(column string, colType *sql.ColumnType, value *string, prevValue string) (huh.Field, error) {
	if strings.Contains(column, "date_created") || strings.Contains(column, "date_modified") || strings.Contains(column, "history") {
		switch column {
		case "date_created":
			m.formMap[column] = prevValue
		case "date_modified":
			m.formMap[column] = utility.TimestampReadable()
		case "history":
			m.formMap[column] = prevValue
		}
		return nil, nil
	}
	var field huh.Field
	res := m.GetSuggestionsString(column)
	switch colType.DatabaseTypeName() {
	case "TEXT":
		if nullable, _ := colType.Nullable(); nullable {
			i := huh.NewInput().Title(column).Key(column).Value(value).Placeholder(prevValue)
			if res != nil {
				i = i.Suggestions(res)
			}
			return i, nil
		} else {
			i := huh.NewInput().Title(column).Key(column).Value(value).Placeholder(prevValue).Validate(Required)
			if res != nil {
				i = i.Suggestions(res)
			}
			return i, nil
		}
	case "INTEGER":
		if nullable, _ := colType.Nullable(); nullable {
			return huh.NewInput().Title(column).Key(column).Value(value).Placeholder(prevValue).Suggestions(res), nil
		} else {
			return huh.NewInput().Title(column).Key(column).Value(value).Placeholder(prevValue).Suggestions(res).Validate(Required), nil
		}
	}
	return field, nil
}

