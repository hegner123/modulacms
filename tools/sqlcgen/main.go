package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

func main() {
	var (
		output string
		dryRun bool
		verify bool
	)

	flag.StringVar(&output, "output", "", "Output path (default: sql/sqlc.yml relative to repo root)")
	flag.BoolVar(&dryRun, "dry-run", false, "Print output to stdout without writing")
	flag.BoolVar(&verify, "verify", false, "Check generated file matches committed file (for CI)")
	flag.Parse()

	if output == "" {
		output = findOutputPath()
	}

	tmplPath := templatePath()
	funcMap := template.FuncMap{
		"derefBool": func(b *bool) bool {
			if b == nil {
				return false
			}
			return *b
		},
		"separator": func(i int) string {
			if i > 0 {
				return "\n"
			}
			return ""
		},
	}
	tmpl, err := template.New("sqlc.yml.tmpl").Funcs(funcMap).ParseFiles(tmplPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	data := TemplateData{
		Engines:   Engines,
		Renames:   Renames,
		Overrides: Overrides,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	generated := buf.Bytes()

	if dryRun {
		fmt.Print(string(generated))
		return
	}

	if verify {
		existing, err := os.ReadFile(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "VERIFY FAIL: cannot read %s: %v\n", output, err)
			os.Exit(1)
		}
		if !bytes.Equal(existing, generated) {
			fmt.Fprintf(os.Stderr, "VERIFY FAIL: %s is out of date. Run 'just sqlc-config' to regenerate.\n", output)
			os.Exit(1)
		}
		fmt.Printf("VERIFY OK: %s\n", output)
		return
	}

	if err := os.WriteFile(output, generated, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", output, err)
		os.Exit(1)
	}
	fmt.Printf("Generated: %s\n", output)
}

// findOutputPath resolves the default output path: sql/sqlc.yml relative to repo root.
func findOutputPath() string {
	candidates := []string{
		filepath.Join("sql", "sqlc.yml"),
	}
	for _, p := range candidates {
		dir := filepath.Dir(p)
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return p
		}
	}
	return candidates[0]
}

// templatePath returns the path to the template file.
func templatePath() string {
	candidates := []string{
		filepath.Join("tools", "sqlcgen", "templates", "sqlc.yml.tmpl"),
		filepath.Join("templates", "sqlc.yml.tmpl"),
	}

	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates,
			filepath.Join(filepath.Dir(exe), "templates", "sqlc.yml.tmpl"),
		)
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return candidates[0]
}
