package main

import (
	"database/sql"
	"fmt"
)

func seedDB(db *sql.DB) sql.Result {
	query := `INSERT INTO posts (slug,title,status,datecreated,datemodified,content,template)
              VALUES 
              ("/blog","Blog",0, 817236, 817236,"Placeholder content","blog.html"),
              ("/about","About",0, 817236, 817236,"Placeholder content","page.html"),
              ("/contact","Contact",0, 817236, 817236,"Placeholder content","page.html");`
	res, err := db.Exec(query)
	if err != nil {
        fmt.Printf("%s",err)
	}
	return res

}
