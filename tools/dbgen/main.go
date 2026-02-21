package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"
)

func main() {
	var (
		outputDir  string
		entityName string
		dryRun     bool
		verify     bool
	)

	flag.StringVar(&outputDir, "output", "", "Output directory (default: ../../internal/db relative to tool)")
	flag.StringVar(&entityName, "entity", "", "Generate single entity by name (default: all)")
	flag.BoolVar(&dryRun, "dry-run", false, "Print file paths without writing")
	flag.BoolVar(&verify, "verify", false, "Check generated files match committed files (for CI)")
	flag.Parse()

	if outputDir == "" {
		// Default to ../../internal/db relative to the tool location
		exe, err := os.Executable()
		if err != nil {
			// Fall back to working directory
			outputDir = filepath.Join("internal", "db")
		} else {
			outputDir = filepath.Join(filepath.Dir(exe), "..", "..", "internal", "db")
		}
		// If running via go run, use relative path from module root
		if _, err := os.Stat(outputDir); err != nil {
			outputDir = filepath.Join("internal", "db")
		}
	}

	// Load template
	tmplPath := templatePath()
	tmpl, err := template.New("entity.go.tmpl").Funcs(templateFuncMap()).ParseFiles(tmplPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	// Filter entities if specific one requested
	entities := Entities
	if entityName != "" {
		entities = nil
		for _, e := range Entities {
			if e.Name == entityName {
				entities = append(entities, e)
				break
			}
		}
		if len(entities) == 0 {
			fmt.Fprintf(os.Stderr, "Entity %q not found. Available entities:\n", entityName)
			for _, e := range Entities {
				fmt.Fprintf(os.Stderr, "  %s\n", e.Name)
			}
			os.Exit(1)
		}
	}

	hasError := false
	for _, entity := range entities {
		outPath := filepath.Join(outputDir, entity.OutputFile)

		if dryRun {
			fmt.Printf("Would write: %s\n", outPath)
			continue
		}

		data := TemplateData{
			Entity:  entity,
			Drivers: DriverConfigs,
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing template for %s: %v\n", entity.Name, err)
			hasError = true
			continue
		}

		// Format with gofmt
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting generated code for %s: %v\n", entity.Name, err)
			fmt.Fprintf(os.Stderr, "Raw output:\n%s\n", buf.String())
			hasError = true
			continue
		}

		if verify {
			existing, err := os.ReadFile(outPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "VERIFY FAIL: cannot read %s: %v\n", outPath, err)
				hasError = true
				continue
			}
			if !bytes.Equal(existing, formatted) {
				fmt.Fprintf(os.Stderr, "VERIFY FAIL: %s is out of date. Run 'just dbgen' to regenerate.\n", outPath)
				hasError = true
				continue
			}
			fmt.Printf("VERIFY OK: %s\n", outPath)
			continue
		}

		if err := os.WriteFile(outPath, formatted, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outPath, err)
			hasError = true
			continue
		}
		fmt.Printf("Generated: %s\n", outPath)
	}

	if hasError {
		os.Exit(1)
	}
}

// templatePath returns the path to the template file.
// It checks multiple locations to support both go run and compiled binary.
func templatePath() string {
	candidates := []string{
		filepath.Join("tools", "dbgen", "templates", "entity.go.tmpl"),
		filepath.Join("templates", "entity.go.tmpl"),
	}

	// Also try relative to executable
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates,
			filepath.Join(filepath.Dir(exe), "templates", "entity.go.tmpl"),
		)
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// Default - will produce a clear error if missing
	return candidates[0]
}
