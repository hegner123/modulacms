package cli

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
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
		dbc    db.Database
		labels []string
	)

	if dbName == "" {
		dbc = db.GetDb(db.Database{})
	} else {
		dbc = db.GetDb(db.Database{Src: dbName})
	}
	q := "SELECT * FROM tables;"
	rows, err := dbc.Connection.QueryContext(dbc.Context, q)
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

func GetFields(table string, dbName string) {
	var dbc db.Database
	if dbName == "" {
		dbc = db.GetDb(db.Database{})
	} else {
		dbc = db.GetDb(db.Database{Src: dbName})
	}
	m, err := db.GetTableColumns(dbc.Context, dbc.Connection, table)
	if err != nil {
		utility.LogError("failed to : ", err)
	}
	fk := GetRelationships(table, dbc)
	MapFields(m, fk, dbc)

}

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

func MapFields(m map[string]string, fk []ForeignKeyReference, dbc db.Database) []huh.Field {
	var fields []huh.Field
	var s []any
	// Extract keys into a slice.
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	// Optionally sort the keys if you need a specific order.
	sort.Strings(keys)

	// Use a classic for loop with an index over the slice of keys.
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		k := MatchFk(key, fk)
		if k != nil {
			s = GetSuggestions(dbc, k.Table, k.Column)
		}
		switch m[key] {
		case "TEXT":
			fields = append(fields, MakeTextInput(key, s))
		case "INTEGER":
			fields = append(fields, MakeIntInput(key, s))
		}
		b := fmt.Sprintf("Index: %d, Key: %s, Value: %s\n", i, key, m[key])
		utility.LogBody(b)
	}
	return fields
}

func MakeTextInput(name string, s []any) huh.Field {
	var ss []string
	for _, suggestion := range s {
		sp := fmt.Sprint(suggestion)
		ss = append(ss, sp)

	}
	var t huh.Field
	if s != nil {
		t = huh.NewInput().Title(name).Key(name).Suggestions(ss).Description("Accepts Strings")
	} else {
		t = huh.NewText().Title(name).Key(name).Description("Accepts Strings")
	}
	return t
}

func MakeIntInput(name string, s []any) huh.Field {
	var i huh.Field
	if s != nil {
		var ss []string
		for _, suggestion := range s {
			sp := fmt.Sprint(suggestion)
			ss = append(ss, sp)

			i = huh.NewInput().Title(name).Key(name).Suggestions(ss).Description("Accepts Integers")
		}
	} else {
		i = huh.NewInput().Title(name).Key(name).Description("Accepts Integers")
	}
	return i
}

func MatchFk(name string, references []ForeignKeyReference) *ForeignKeyReference {
	for _, r := range references {
		if r.From == name {
			return &r
		}
	}
	return nil
}

func GetSuggestions(dbc db.Database, table string, column string) []any {
	if column == "NIll" {
		return nil
	} else {
		r, err := db.GetColumnRows(dbc, table, column)
		if err != nil {
			utility.LogError("ERROR", err)
		}
		fmt.Println(r)
		return r
	}

}
