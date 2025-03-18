package cli

import (
	"database/sql"
	"fmt"
	"sort"

	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
	utility "github.com/hegner123/modulacms/internal/Utility"
)

// ForeignKeyReference holds the referenced table and column information.
type ForeignKeyReference struct {
	From   string
	Table  string // Referenced table name.
	Column string // Referenced column name.
}

func GetTables(dbName string) []string {
	var (
		d      db.DbDriver
		labels []string
	)

	if dbName == "" {
		d = db.ConfigDB(config.Env)
	} else {
		d = db.ConfigDB(config.Env)
	}
	con, ctx, _ := d.GetConnection()
	q := "SELECT * FROM tables;"
	rows, err := con.QueryContext(ctx, q)
	if err != nil {
		utility.LogError("", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id        int
			label     string
			author_id int
		)
		err = rows.Scan(&id, &label, &author_id)
		if err != nil {
			utility.LogError("", err)
		}
		labels = append(labels, label)
	}
	if err := rows.Err(); err != nil {
		utility.LogError("ERROR: ", err)
	}
	return labels
}

func GetFieldsString(table string, dbName string) string {
	var r string
	var (
		d db.DbDriver
	)

	if dbName == "" {
		d = db.ConfigDB(config.Env)
	} else {
		d = db.ConfigDB(config.Env)
	}
	con, ctx, _ := d.GetConnection()
	_, m, err := db.GetTableColumns(ctx, con, table)
	if err != nil {
		utility.LogError("failed to : ", err)
	}
	//fk := GetRelationships(table, dbc)
	//MapFields(m, fk, dbc)
	// Extract the keys into a slice
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// Sort the slice of keys
	sort.Ints(keys)

	// Iterate over the sorted keys and access the map's values
	for _, k := range keys {
		r += fmt.Sprintf("%d: %s\n", k, m[k])
	}

	return r

}

func GetColumns(t string) (*[]string, *[]*sql.ColumnType, error) {
	dbt := db.StringDBTable(t)
	verbose := false
	query := "SELECT * FROM"
	c := config.LoadConfig(&verbose, "")
	d := db.ConfigDB(c)
	rows, err := d.ExecuteQuery(query, dbt)
	if err != nil {
		return nil, nil, err
	}
	clm, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	ct, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, err

	}
	return &clm, &ct, nil
}

/*
func GetColumns(table string, dbName string) ([]Column, []string) {
	var columns []Column
	var headers []string
	var d db.DbDriver
	if dbName == "" {
		d = db.ConfigDB(config.Env)
	} else {
		d = db.ConfigDB(config.Env)
	}
	con, ctx, _ := d.GetConnection()
	t, m, err := db.GetTableColumns(ctx, con, table)
	if err != nil {
		utility.LogError("failed to : ", err)
	}
	//fk := GetRelationships(table, dbc)
	//MapFields(m, fk, dbc)
	// Extract the keys into a slice
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// Sort the slice of keys
	sort.Ints(keys)

	// Iterate over the sorted keys and access the map's values
	for _, k := range keys {
		c := Column{index: k, label: m[k], inputType: t[m[k]]}
		columns = append(columns, c)
		headers = append(headers, m[k])
	}

	return columns, headers

}
*/

func GetRelationships(tableName string, dbc db.Database) []ForeignKeyReference {
	if tableName == "" {
		err := fmt.Errorf("tableName is blank")
		utility.LogError("Error: ", err)
	}

	// Build the PRAGMA query.
	// Note: SQLite does not support parameter binding for table names,
	// so ensure that tableName is trusted or properly validated.
	query := fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)

	rows, err := dbc.Connection.QueryContext(dbc.Context, query)
	if err != nil {
		utility.LogError("Error: ", err)
	}
	fmt.Println(query)
	defer rows.Close()
	var references []ForeignKeyReference
	// PRAGMA foreign_key_list returns these columns:
	// id, seq, table, from, to, on_update, on_delete, match
	// We are interested in the 3rd column ("table") and 5th column ("to").
	for rows.Next() {
		var (
			id       int
			seq      int
			refTable string
			from     string
			refCol   sql.NullString
			onUpdate string
			onDelete string
			match    string
		)
		if err := rows.Scan(&id, &seq, &refTable, &from, &refCol, &onUpdate, &onDelete, &match); err != nil {
			utility.LogError("failed to scan row:", err)
		}

		// Append the referenced table and column to our slice.
		references = append(references, ForeignKeyReference{
			From:   from,
			Table:  refTable,
			Column: db.ReadNullString(refCol),
		})
	}

	if err := rows.Err(); err != nil {
		utility.LogError("ERROR: ", err)
	}
	utility.LogBody(references)
	return references
}

func GetColumnsRows(t string) (*[]string, *[][]string, error) {
	dbt := db.StringDBTable(t)
	verbose := false
	query := "SELECT * FROM"
	c := config.LoadConfig(&verbose, "")
	d := db.ConfigDB(c)
	rows, err := d.ExecuteQuery(query, dbt)
	if err != nil {
		return nil, nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	listRows, err := db.GenericList(dbt, d)
	if err != nil {
		return nil, nil, err
	}
	return &columns, listRows, nil

}
func (m model) GetSuggestionsString(column string) []string {
	d := db.ConfigDB(config.Env)
	con, ctx, _ := d.GetConnection()
	if column == "NIll" {
		return nil
	} else {
		r, err := db.GetColumnRowsString(con, ctx, m.table, column)
		if err != nil {
			utility.LogError("ERROR", err)
		}
		return r
	}

}


/*
func MapFields(m map[string]string, fk []ForeignKeyReference, dbc db.Database) {
	var fields []any
	var s []any
	fmt.Print(fields...)
	// Extract keys into a slice.
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	// Optionally sort the keys if you need a specific order.
	sort.Strings(keys)

	// Use a classic for loop with an index over the slice of keys.
	for i, v := range keys {
		key := v
		k := MatchFk(key, fk)
		if k != nil {
			s = GetSuggestions(dbc, k.Table, k.Column)
		}
		switch m[key] {
		case "TEXT":
		case "INTEGER":
		}
		b := fmt.Sprintf("Index: %d, Key: %s, Value: %s\n", i, key, m[key])
		utility.LogBody(b)
	}
	fmt.Print(s)
}
*/


