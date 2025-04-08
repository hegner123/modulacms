package media

import (
	config "github.com/hegner123/modulacms/internal/config"
	db "github.com/hegner123/modulacms/internal/db"
)

func CreateMedia(name string, c config.Config) string {
	d := db.ConfigDB(c)
	params := db.CreateMediaParams{
		Name: db.StringToNullString(name),
	}
	mediaRow := d.CreateMedia(params)
    return mediaRow.Name.String
}
