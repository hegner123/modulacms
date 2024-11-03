package main

import (
	"database/sql"
	"fmt"
)

/*
const mediaTable string = `
    CREATE TABLE IF NOT EXISTS media (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    displayName TEXT,
    alt TEXT,
    caption TEXT,
    description TEXT,
    class TEXT,
    createdBy INTEGER,
    dateCreated TEXT,
    dateModified TEXT,
    url TEXT,
    mimeType TEXT,
    dimensions TEXT,
    optimizedMobile TEXT,
    optimizedTablet TEXT,
    optimizedDesktop TEXT,
    optimizedUltrawide TEXT);`
*/

func dbGetMediaByName(db *sql.DB, name string) (Media, error) {
	var media Media
	query := fmt.Sprintf(`SELECT * FROM media WHERE name = %s`, name)
	err := db.QueryRow(query).Scan(&media.Id, &media.Name, &media.DisplayName,
		&media.Alt, &media.Caption, &media.Description, &media.Class, &media.DateCreated,
		&media.DateModified, &media.Url, &media.MimeType, &media.Dimensions,
		&media.OptimizedMobile, &media.OptimizedTablet, &media.OptimizedDesktop, &media.OptimizedUltrawide)

	if err != nil {
		fmt.Printf("%s\n", err)
	}

	return media, nil
}

func createMedia(db *sql.DB, media Media) (int64, error) {
	result, err := db.Exec(queryCreateBuilder(media, "media"), media.Id, media.Name, media.DisplayName,
		media.Alt, media.Caption, media.Description, media.Class, media.DateCreated,
		media.DateModified, media.Url, media.MimeType, media.Dimensions,
		media.OptimizedMobile, media.OptimizedTablet, media.OptimizedDesktop, media.OptimizedUltrawide)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func deleteMediaByName(db *sql.DB, column string, value string) (int64, error) {
	query := fmt.Sprintf("DELETE FROM media WHERE %s='%s';", column, value)
	result, err := db.Exec(query)
	if err != nil {
		fmt.Printf("%s\n", err)
		return int64(0), err
	}
	ra, err := result.RowsAffected()
	return ra, err

}
