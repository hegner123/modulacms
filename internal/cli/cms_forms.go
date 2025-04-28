package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
)

/*
type Datatypes struct {
    DatatypeID   int64          `json:"datatype_id"`
    ParentID     sql.NullInt64  `json:"parent_id"`
    Label        string         `json:"label"`
    Type         string         `json:"type"`
    AuthorID     int64          `json:"author_id"`
    DateCreated  sql.NullString `json:"date_created"`
    DateModified sql.NullString `json:"date_modified"`
    History      sql.NullString `json:"history"`
}
*/

func (m *Model) CreateDatatypeForm() *huh.Form {
	var (
		parentID int
		label    string
		dType    string
	)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Parent").
				OptionsFunc(
					func() []huh.Option[int] {
						options := make([]huh.Option[int], 0)
						d := db.ConfigDB(*m.config)
						r, err := d.ListDatatypes()
						if err != nil {
							e := fmt.Errorf("error listing datatypes %w", err)
							m.err = e
						}
						for _, v := range *r {
							option := huh.NewOption(v.Label, int(v.DatatypeID))
							options = append(options, option)
						}
						return options
					},
					nil,
				).
				Value(&parentID), // st
			huh.NewInput().
				Title("Label").
				Value(&label),
			huh.NewInput().
				Title("Type").
				Value(&dType),
		))

	return form
}
func (m *Model) CreateFieldForm() *huh.Form {
	var (
		parentID int
		label    string
		data     *[]string
		dType    string
	)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Parent").
				OptionsFunc(
					func() []huh.Option[int] {
						options := make([]huh.Option[int], 0)
						d := db.ConfigDB(*m.config)
						r, err := d.ListDatatypes()
						if err != nil {
							e := fmt.Errorf("error listing datatypes %w", err)
							m.err = e
						}
						for _, v := range *r {
							option := huh.NewOption(v.Label, int(v.DatatypeID))
							options = append(options, option)
						}
						return options
					},
					nil,
				).
				Value(&parentID), 
			huh.NewMultiSelect[string]().Title("Options").Options(
                huh.NewOption("Required", "required"),
                huh.NewOption("Validation","Validation"),
                ).Value(data),
			huh.NewInput().
				Title("Label").
				Value(&label),
			huh.NewInput().
				Title("Type").
				Value(&dType),
		))

	return form
}
