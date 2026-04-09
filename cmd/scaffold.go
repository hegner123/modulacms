package main

import (
	"fmt"
	"os"
	"path/filepath"

	"charm.land/huh/v2"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// scaffoldArgs holds user selections for the docker scaffold.
type scaffoldArgs struct {
	DBDriver string // "sqlite", "mysql", "postgres"
	Storage  string // "minio", "external"
	Proxy    string // "caddy", "nginx", "raw"
	OutDir   string
	Stdout   bool
}

var scaffoldFlags scaffoldArgs

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Generate project scaffolding",
	Long:  "Generate deployment files, plugin skeletons, and other project scaffolding.",
}

var scaffoldDockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Generate Dockerfile and docker-compose.yml for VPS deployment",
	Long: `Generate a Dockerfile, docker-compose.yml, config, and supporting files
for deploying ModulaCMS on a VPS.

Interactive mode (default): walks through a form to select database,
storage, and reverse proxy options.

Non-interactive: pass all three flags to skip the form.

Examples:
  modula scaffold docker
  modula scaffold docker --db postgres --storage minio --proxy caddy
  modula scaffold docker --db sqlite --storage external --proxy raw --stdout`,
	RunE: runScaffoldDocker,
}

func init() {
	scaffoldCmd.AddCommand(scaffoldDockerCmd)

	f := scaffoldDockerCmd.Flags()
	f.StringVar(&scaffoldFlags.DBDriver, "db", "", "database driver: sqlite, mysql, postgres")
	f.StringVar(&scaffoldFlags.Storage, "storage", "", "S3 storage: minio (self-hosted), external")
	f.StringVar(&scaffoldFlags.Proxy, "proxy", "", "Reverse proxy: caddy, nginx, raw")
	f.StringVar(&scaffoldFlags.OutDir, "out", "./modula-deploy", "Output directory")
	f.BoolVar(&scaffoldFlags.Stdout, "stdout", false, "Print files to stdout instead of writing")
}

func runScaffoldDocker(cmd *cobra.Command, args []string) error {
	a := scaffoldFlags

	// If all three flags are provided, skip the form.
	allFlagsSet := a.DBDriver != "" && a.Storage != "" && a.Proxy != ""

	if !allFlagsSet {
		if !isatty.IsTerminal(os.Stdout.Fd()) {
			return fmt.Errorf("non-interactive terminal: pass --db, --storage, and --proxy flags")
		}
		if err := runScaffoldForm(&a); err != nil {
			return err
		}
	}

	if err := validateScaffoldArgs(a); err != nil {
		return err
	}

	return writeScaffoldFiles(cmd, a)
}

func validateScaffoldArgs(a scaffoldArgs) error {
	switch a.DBDriver {
	case "sqlite", "mysql", "postgres":
	default:
		return fmt.Errorf("invalid --db value %q: must be sqlite, mysql, or postgres", a.DBDriver)
	}
	switch a.Storage {
	case "minio", "external":
	default:
		return fmt.Errorf("invalid --storage value %q: must be minio or external", a.Storage)
	}
	switch a.Proxy {
	case "caddy", "nginx", "raw":
	default:
		return fmt.Errorf("invalid --proxy value %q: must be caddy, nginx, or raw", a.Proxy)
	}
	return nil
}

func runScaffoldForm(a *scaffoldArgs) error {
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Database").
				Description("Which database engine will you use?").
				Options(
					huh.NewOption("SQLite (single file, no extra service)", "sqlite"),
					huh.NewOption("PostgreSQL", "postgres"),
					huh.NewOption("MySQL", "mysql"),
				).
				Value(&a.DBDriver),

			huh.NewSelect[string]().
				Title("Object Storage").
				Description("Where will media files be stored?").
				Options(
					huh.NewOption("Self-hosted MinIO (included in compose)", "minio"),
					huh.NewOption("External S3 service (AWS, Cloudflare R2, etc.)", "external"),
				).
				Value(&a.Storage),

			huh.NewSelect[string]().
				Title("Reverse Proxy").
				Description("How should traffic reach ModulaCMS?").
				Options(
					huh.NewOption("Caddy (automatic HTTPS)", "caddy"),
					huh.NewOption("Nginx", "nginx"),
					huh.NewOption("None (expose ports directly)", "raw"),
				).
				Value(&a.Proxy),
		),
	).Run()

	if err != nil {
		return fmt.Errorf("form cancelled: %w", err)
	}
	return nil
}

func writeScaffoldFiles(cmd *cobra.Command, a scaffoldArgs) error {
	data := newScaffoldData(a)
	out := cmd.OutOrStdout()

	type scaffoldFile struct {
		name    string
		render  func(scaffoldData) string
		cond    bool
	}

	files := []scaffoldFile{
		{"Dockerfile", renderDockerfile, true},
		{"docker-compose.yml", renderCompose, true},
		{"modula.config.json", renderConfig, true},
		{".env", renderEnv, true},
		{"Caddyfile", renderCaddyfile, data.IsCaddy},
		{"nginx.conf", renderNginxConf, data.IsNginx},
	}

	if a.Stdout {
		for _, f := range files {
			if !f.cond {
				continue
			}
			fmt.Fprintf(out, "--- %s ---\n", f.name)
			fmt.Fprintln(out, f.render(data))
		}
		return nil
	}

	dir := a.OutDir
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	for _, f := range files {
		if !f.cond {
			continue
		}
		path := filepath.Join(dir, f.name)
		if err := os.WriteFile(path, []byte(f.render(data)), 0640); err != nil {
			return fmt.Errorf("writing %s: %w", f.name, err)
		}
		fmt.Fprintf(out, "  Created %s\n", path)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Next steps:")
	fmt.Fprintf(out, "  1. cd %s\n", dir)
	fmt.Fprintln(out, "  2. Edit .env and replace CHANGE_ME values")
	if a.Proxy == "caddy" {
		fmt.Fprintln(out, "  3. Edit Caddyfile — replace :80 with your domain for auto-HTTPS")
	} else if a.Proxy == "nginx" {
		fmt.Fprintln(out, "  3. Edit nginx.conf — set server_name and configure TLS")
	}
	fmt.Fprintln(out, "  4. docker compose up -d --build")
	if a.Proxy == "raw" {
		fmt.Fprintln(out, "  5. Access the admin panel at http://your-server:8080/admin/")
	} else {
		fmt.Fprintln(out, "  5. Access the admin panel at https://your-domain/admin/")
	}
	fmt.Fprintf(out, "  6. SSH TUI: ssh -p 2233 admin@your-server\n")
	fmt.Fprintln(out)

	return nil
}
