package cli

import (
	"github.com/charmbracelet/huh"
	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

var ErrLog utility.Logger = *utility.NewLogger(utility.ERROR)

func (m model) BuildForm(table db.DBTable) (*huh.Form, int) {
	columns, colType, err := GetColumns(string(table))
	if err != nil {
		return nil, 0
	}
	var fields []huh.Field
	for i, c := range *columns {
		t := *colType
		f, err := NewFieldFromType(c, t[i])
		if err != nil {
			return nil, 0
		}
        if f == nil {
            continue
        }
		fields = append(fields, f)

	}
	group := huh.NewGroup(fields...)
	form := huh.NewForm(group)
	return form, len(*columns)

}
func (m model) BuildUpdateForm(table db.DBTable) (*huh.Form, int) {
    row := *m.row
	columns, colType, err := GetColumns(string(table))
	if err != nil {
		return nil, 0
	}
	var fields []huh.Field
	for i, c := range *columns {
		t := *colType
		f, err := NewUpdateFieldFromType(c, t[i], row[i])
		if err != nil {
			return nil, 0
		}
        if f == nil {
            continue
        }
        
		fields = append(fields, f)

	}
	group := huh.NewGroup(fields...)
	form := huh.NewForm(group)
	return form, len(*columns)

}
