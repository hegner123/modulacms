package db

import "fmt"

func GetPluginSchema(table *string) {
	var dbc Database
	if table != nil {
		dbc = GetDb(Database{Src: *table})
	} else {
		dbc = GetDb(Database{})
	}
	s, err := GetTableColumns(dbc.Context, dbc.Connection, *table)
	if err != nil {
		return
	}
	fmt.Print(s)

}
