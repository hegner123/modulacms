package cli

import (
	"database/sql"

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

func GetTables() []string {
	var (
		d      db.DbDriver
		labels []string
	)
	d = db.ConfigDB(config.Env)
	con, ctx, _ := d.GetConnection()
	q := "SELECT * FROM tables;"
	rows, err := con.QueryContext(ctx, q)
	if err != nil {
		utility.DefaultLogger.Error("", err)
	}

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
		utility.DefaultLogger.Error("ERROR: ", err)
	}
	return labels
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

func GetColumnsRows(t string) ([]string, [][]string, error) {
	dbt := db.StringDBTable(t)
	verbose := false
	query := "SELECT * FROM"
	c := config.LoadConfig(&verbose, "")
	d := db.ConfigDB(c)
	rows, err := d.ExecuteQuery(query, dbt)
	if err != nil {
		utility.DefaultLogger.Ferror( "", err)
		return nil, nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		utility.DefaultLogger.Ferror( "", err)
		return nil, nil, err
	}
	listRows, err := db.GenericList(dbt, d)
	if err != nil {
		utility.DefaultLogger.Ferror( "", err)
		return nil, nil, err
	}
	return columns, listRows, nil

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

