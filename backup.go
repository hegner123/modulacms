package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func createBackup(dbFile, mediaDir, pluginDir, output string) error {

	dbF := FileExists(dbFile)
	mediaD := DirExists(mediaDir)
	pluginD := DirExists(pluginDir)
	outD := DirExists(output)

	if !dbF || !mediaD || !pluginD || !outD {
		return fmt.Errorf("dbFile Exists: %v, mediaDir exists: %v, pluginDir exists: %v,outputDir exists: %v\n", dbF, mediaD, pluginD, outD)
	}

	// Create a zip file
	backupName := fmt.Sprintf("%s_%s.zip", output, timestamp())
	backupFile, err := os.Create(backupName)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer backupFile.Close()

	zipWriter := zip.NewWriter(backupFile)
	defer zipWriter.Close()

	db, err := getDb(Database{DB: dbFile})
	if err != nil {
		logError("db failed to open", err)
	}

	dbDumpFile, err := zipWriter.Create("database.sql")
	if err != nil {
		logError("failed to create database dump in archive: ", err)
	}

	rows, err := db.Query("SELECT sql FROM sqlite_master WHERE type='table'")
	if err != nil {
		return fmt.Errorf("failed to dump database: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sqlText string
		if err := rows.Scan(&sqlText); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		_, err = dbDumpFile.Write([]byte(sqlText + ";\n"))
		if err != nil {
			return fmt.Errorf("failed to write to archive: %w", err)
		}
	}

	// Add media files
	err = addFilesToZip(zipWriter, mediaDir, "media")
	if err != nil {
		return fmt.Errorf("failed to add media files: %w", err)
	}

	// Add plugin files
	err = addFilesToZip(zipWriter, pluginDir, "plugins")
	if err != nil {
		return fmt.Errorf("failed to add plugin files: %w", err)
	}

	return nil
}

func addFilesToZip(zipWriter *zip.Writer, dir, baseInZip string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		zipPath := filepath.Join(baseInZip, relPath)
		fileWriter, err := zipWriter.Create(zipPath)
		if err != nil {
			return fmt.Errorf("failed to add file %s to zip: %w", path, err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", path, err)
		}
		defer file.Close()

		_, err = io.Copy(fileWriter, file)
		if err != nil {
			return fmt.Errorf("failed to copy file %s to zip: %w", path, err)
		}

		return nil
	})
}
