package db

import (
	"fmt"
	"testing"

	config "github.com/hegner123/modulacms/internal/config"
)

func TestSqlite(t *testing.T) {
    p := config.NewFileProvider("")
    m := config.NewManager(p)
    c,err := m.Config()
    if err!=nil {
        t.Fatal(err)
    }
	d := ConfigDB(*c)
	args := []string{"admin_content_data"}
	rows := d.GetForeignKeys(args)
	if rows == nil {
		t.Fatal()
	}
    r:=d.ScanForeignKeyQueryRows(rows)
    for _,row:= range r {
		fmt.Printf("FK: id=%d, seq=%d, table=%s, from=%s, to=%s, on_update=%s, on_delete=%s, match=%s\n",
			row.id, row.seq, row.tableName, row.fromCol, row.toCol, row.onUpdate, row.onDelete, row.match)
        data,ok := d.(Database)
        if !ok{
            t.Fatal()
        } 

        data.SelectColumnFromTable(row.tableName,row.toCol)
        
    }
    

}
