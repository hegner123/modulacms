package main

import (
	"html/template"
	"os"
	"path/filepath"
)

func parseTemplateGlobs(rootDir string, pattern string) (*template.Template, error) {
	tmpl := template.New("")

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err // Return if an error occurs
		}

		if d.IsDir() {
			return nil
		}

		match, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return err // Return if there's a match error
		}

		if match {
			_, err := tmpl.ParseFiles(path)
			if err != nil {
				return err // Return if parsing fails
			}
		}

		return nil
	})

	if err != nil {
		return nil, err // Return if any error occurred during walk
	}

	return tmpl, nil
}
