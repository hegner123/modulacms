package main

import (
	"html/template"
	"os"
	"path/filepath"
)

// parseTemplateGlobs recursively parses all templates matching the provided glob pattern into a single *template.Template.
func parseTemplateGlobs(rootDir string, pattern string) (*template.Template, error) {
	// Create a new template instance
	tmpl := template.New("")

	// Walk the directory to find all matching files
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err // Return if an error occurs
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if the file matches the pattern
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
