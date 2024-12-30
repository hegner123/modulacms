package cli

import (
	"fmt"
	"log"

	huh "github.com/charmbracelet/huh"
	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

var (
	menu   string
	table  string
)

func TableOptions() []huh.Option[string] {
	var options []huh.Option[string]
	dbc := db.GetDb(db.Database{})
	tables, err := db.ListTable(dbc.Connection, dbc.Context)
	if err != nil {
		utility.LogError("failed to : ", err)
	}
	for _, table := range *tables {
		options = append(options, huh.NewOption(table.Label.String, table.Label.String))

	}
	return options

}

func Form() {
	form := huh.NewForm(
		huh.NewGroup(
			// Show main menu
			huh.NewSelect[string]().
				Title("Welcome to ModulaCMS").
				Options(
					huh.NewOption("Create", "0"),
					huh.NewOption("Read", "1"),
					huh.NewOption("Update", "2"),
					huh.NewOption("Delete", "3"),
				).
				Value(&menu),

			huh.NewSelect[string]().
				Title("Which Table do you want to use?").
				OptionsFunc(
					func() []huh.Option[string] {
						return TableOptions()

					},
					&table,
				).
				Value(&table),
		),
	)
	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(table)

}
