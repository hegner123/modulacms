package cli

import (
	"database/sql"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	utility "github.com/hegner123/modulacms/internal/utility"
)

func (m *Model) NewFieldFromType(c *config.Config, column string, colType *sql.ColumnType, value *string) (huh.Field, error) {
	if strings.Contains(column, "date_created") || strings.Contains(column, "date_modified") || strings.Contains(column, "history") {
		switch column {
		case "date_created":
			ts := utility.TimestampReadable()
			*value = ts
			m.FormValues = append(m.FormValues, value)
		case "date_modified":
			ts := utility.TimestampReadable()
			*value = ts
			m.FormValues = append(m.FormValues, value)
		case "history":
			h := ""
			*value = h
			m.FormValues = append(m.FormValues, value)
		}
		return nil, nil
	}
	var field huh.Field
	res := m.GetSuggestionsString(c, column)
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

func (m *Model) NewUpdateFieldFromType(c *config.Config, column string, colType *sql.ColumnType, value *string, prevValue string) (huh.Field, error) {
	if strings.Contains(column, "date_created") || strings.Contains(column, "date_modified") || strings.Contains(column, "history") {
		switch column {
		case "date_created":
			pv := prevValue
			m.FormValues = append(m.FormValues, &pv)
		case "date_modified":
			ts := utility.TimestampReadable()
			m.FormValues = append(m.FormValues, &ts)
		case "history":
			pv := prevValue
			m.FormValues = append(m.FormValues, &pv)
		}
		return nil, nil
	}
	var field huh.Field
	res := m.GetSuggestionsString(c, column)
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

func (m Model) GetSuggestionsString(c *config.Config, column string) []string {
	d := db.ConfigDB(*c)
	con, ctx, _ := d.GetConnection()
	if column == "NIll" {
		return nil
	} else {
		r, err := db.GetColumnRowsString(con, ctx, m.Table, column)
		if err != nil {
			utility.DefaultLogger.Error("ERROR", err)
			return nil
		}
		return r

	}
}
