package media

import (
	config "github.com/hegner123/modulacms/internal/Config"
	db "github.com/hegner123/modulacms/internal/Db"
)

func CreateMedia(name string, c config.Config) string {
	d := db.ConfigDB(c)
	params := db.CreateMediaParams{
		Name: db.Ns(name),
	}
	mediaRow := d.CreateMedia(params)
    return mediaRow.Name.String
}
