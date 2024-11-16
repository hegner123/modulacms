package main

import (
	"database/sql"
	"fmt"
	"log"
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
	var columns = "id, name, displayname, alt, caption, description, class, author, authorid, datecreated, datemodified, url, mimeType, dimensions, optimizedmobile, optimizedtablet, optimizeddesktop, optimizedultrawide "
	query := fmt.Sprintf(`SELECT %s FROM media WHERE name = ?`,columns)
    fmt.Print(query)
	err := db.QueryRow(query, name).Scan(&media.ID, &media.Name, &media.DisplayName,
		&media.Alt, &media.Caption, &media.Description, &media.Class, &media.Author, &media.AuthorID, &media.DateCreated,
		&media.DateModified, &media.Url, &media.MimeType, &media.Dimensions,
		&media.OptimizedMobile, &media.OptimizedTablet, &media.OptimizedDesktop, &media.OptimizedUltrawide)

	if err != nil {
		fmt.Printf("%s\n", err)
	}

	return media, nil
}

func dbCreateMedia(db *sql.DB, media Media) (int64, error) {
	result, err := db.Exec(FormatSqlInsertStatement(media, "media"), media.ID, media.Name, media.DisplayName,
		media.Alt, media.Caption, media.Description, media.Class, media.DateCreated,
		media.DateModified, media.Url, media.MimeType, media.Dimensions,
		media.OptimizedMobile, media.OptimizedTablet, media.OptimizedDesktop, media.OptimizedUltrawide)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func dbDeleteMediaByName(db *sql.DB, column string, value string) (int64, error) {
	query := fmt.Sprintf(`DELETE FROM media WHERE %s="%s";`, column, value)
	result, err := db.Exec(query)
	if err != nil {
		fmt.Printf("%s\n", err)
		return int64(0), err
	}
	ra, err := result.RowsAffected()
	return ra, err

}

func dbGetMediaDimensions(dbName string) []MediaDimension {
	db, err := getDb(Database{})
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	i := 0
	ds := []MediaDimension{}
	query := `SELECT label, width, height FROM media_dimensions`
	rows, err := db.Query(query)
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	for rows.Next() {
		err := rows.Scan(&ds[i].Label, &ds[i].Width, &ds[i].Height)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		i++
	}
	return ds
}

func dbCreateMediaDimensions(db *sql.DB, d MediaDimension) (int64, error) {
	result, err := db.Exec(FormatSqlInsertStatement(d, "media_dimensions"), d.Label, d.Width, d.Height)

	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func dbUpdateMediaDimensions(db *sql.DB, field map[string]string) bool {
	log.Panic("not yet implemented")
	return false
}

func dbDeleteMediaDimensionByName(db *sql.DB, column string, value string) (int64, error) {
	query := fmt.Sprintf(`DELETE FROM media_dimensions WHERE %s="%s";`, column, value)
	result, err := db.Exec(query)
	if err != nil {
		fmt.Printf("%s\n", err)
		return int64(0), err
	}
	ra, err := result.RowsAffected()
	return ra, err

}
