package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/huh/v2"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/registry"
	"github.com/hegner123/modulacms/internal/utility"
)

// runDefaultCommand runs when `modula` is invoked with no subcommand.
// Phase 1: if the cwd matches a registered project, show its status.
// Phase 2: if not, offer interactive setup.
func runDefaultCommand(cmd *cobra.Command, args []string) error {
	configureLogger()

	reg, err := registry.Load()
	if err != nil {
		utility.DefaultLogger.Warn("Failed to load project registry", err)
		reg = &registry.Registry{Projects: make(map[string]*registry.Project)}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}

	name, proj := reg.FindByDir(cwd)
	if proj != nil {
		return runProjectStatus(cmd, reg, name, proj)
	}

	return runProjectSetup(cmd, reg, cwd)
}

// runProjectStatus displays the status of a registered project (Phase 1).
func runProjectStatus(cmd *cobra.Command, reg *registry.Registry, name string, proj *registry.Project) error {
	out := cmd.OutOrStdout()
	hasMissing := false

	fmt.Fprintf(out, "ModulaCMS project: %s\n\n", name)

	// Verify base config
	if proj.Base != "" {
		if _, err := os.Stat(proj.Base); err == nil {
			fmt.Fprintf(out, "  base config: %s  [ok]\n", proj.Base)
		} else {
			fmt.Fprintf(out, "  base config: %s  [missing]\n", proj.Base)
			hasMissing = true
		}
	} else {
		fmt.Fprintf(out, "  base config: (not set)\n")
	}

	// Verify project assets
	baseDir := filepath.Dir(proj.Base)
	certsDir := filepath.Join(baseDir, "certs")
	searchIdx := filepath.Join(baseDir, "search.idx")

	if _, err := os.Stat(certsDir); err == nil {
		fmt.Fprintf(out, "  certificates: %s  [ok]\n", certsDir)
	} else {
		fmt.Fprintf(out, "  certificates: %s  [missing]\n", certsDir)
		hasMissing = true
	}

	if _, err := os.Stat(searchIdx); err == nil {
		fmt.Fprintf(out, "  search index: %s  [ok]\n", searchIdx)
	} else {
		fmt.Fprintf(out, "  search index: (created on first run)\n")
	}

	// Verify environment overlays
	envNames := reg.EnvNames(name)
	if len(envNames) > 0 {
		fmt.Fprintf(out, "\n  Environments:\n")
		for _, env := range envNames {
			path := proj.Envs[env]
			marker := ""
			if env == proj.DefaultEnv {
				marker = " *"
			}

			if _, err := os.Stat(path); err == nil {
				fmt.Fprintf(out, "    %-12s %s  [ok]%s\n", env, path, marker)
			} else {
				fmt.Fprintf(out, "    %-12s %s  [missing]%s\n", env, path, marker)
				hasMissing = true
			}
		}
	}

	// Available commands
	fmt.Fprintf(out, "\n  Commands:\n")
	for _, env := range envNames {
		fmt.Fprintf(out, "    modula serve %s %s\n", name, env)
	}
	if len(envNames) > 0 {
		fmt.Fprintln(out)
		for _, env := range envNames {
			fmt.Fprintf(out, "    modula connect %s %s\n", name, env)
		}
		fmt.Fprintln(out)
		for _, env := range envNames {
			fmt.Fprintf(out, "    modula tui %s %s\n", name, env)
		}
	}

	// Troubleshooting
	if hasMissing {
		fmt.Fprintf(out, "\n  Some files are missing. To fix:\n")
		fmt.Fprintf(out, "    modula config overlay --env <name>   # scaffold an overlay\n")
		fmt.Fprintf(out, "    modula cert generate                 # generate localhost certificates\n")
		fmt.Fprintf(out, "    modula connect set --base %s <path>  # update the base config path\n", name)
		fmt.Fprintf(out, "    modula connect set %s <env> <path>   # update an overlay path\n", name)
	}

	fmt.Fprintln(out)
	return nil
}

// runProjectSetup offers interactive project setup (Phase 2).
func runProjectSetup(cmd *cobra.Command, reg *registry.Registry, cwd string) error {
	out := cmd.OutOrStdout()

	suggested := filepath.Base(cwd)
	projectName := suggested

	if yesFlag {
		// Auto-accept: use defaults, skip all prompts
		fmt.Fprintln(out, "No ModulaCMS project found. Setting up with defaults (--yes).")
		fmt.Fprintln(out)
	} else {
		// Non-interactive: just print a message
		if !isatty.IsTerminal(os.Stdout.Fd()) {
			fmt.Fprintln(out, "No ModulaCMS project found in this directory.")
			fmt.Fprintln(out, "Run 'modula' in an interactive terminal to set one up, or use 'modula --yes'.")
			return nil
		}

		fmt.Fprintln(out, "No ModulaCMS project found in this directory.")
		fmt.Fprintln(out)

		// Prompt 1: confirm setup
		var setupConfirm bool
		err := huh.NewConfirm().
			Title("Set up this directory as a ModulaCMS project?").
			Affirmative("Yes").
			Negative("No").
			Value(&setupConfirm).Run()
		if err != nil || !setupConfirm {
			return nil
		}

		// Prompt 2: project name
		err = huh.NewInput().
			Title("Project name").
			Value(&projectName).
			Placeholder(suggested).Run()
		if err != nil {
			return nil
		}
		projectName = strings.TrimSpace(projectName)
		if projectName == "" {
			projectName = suggested
		}

		// Prompt 3: confirm config creation
		var createConfigs bool
		err = huh.NewConfirm().
			Title("Create config files in modula/?").
			Affirmative("Yes").
			Negative("No").
			Value(&createConfigs).Run()
		if err != nil || !createConfigs {
			return nil
		}
	}

	// Create .modula/ directory
	modulaDir := filepath.Join(cwd, "modula")
	if mkErr := os.MkdirAll(modulaDir, 0750); mkErr != nil {
		return fmt.Errorf("creating .modula directory: %w", mkErr)
	}

	// Write base config
	basePath := filepath.Join(modulaDir, "modula.config.json")
	baseConfig := buildBaseConfig()
	if wErr := writeConfigFile(basePath, baseConfig); wErr != nil {
		return fmt.Errorf("writing base config: %w", wErr)
	}
	fmt.Fprintf(out, "  Created %s\n", basePath)

	// Write overlays
	overlays := map[string]map[string]any{
		"local": buildLocalOverlay(),
		"dev":   buildDevOverlay(),
		"prod":  buildProdOverlay(),
	}

	for env, data := range overlays {
		overlayFile := filepath.Join(modulaDir, fmt.Sprintf("modula.%s.config.json", env))
		if wErr := writeOverlayFile(overlayFile, data); wErr != nil {
			utility.DefaultLogger.Warn(fmt.Sprintf("Failed to write %s overlay", env), wErr)
			continue
		}
		fmt.Fprintf(out, "  Created %s\n", overlayFile)
	}

	// Register in ~/.modula/configs.json
	if err := reg.SetBase(projectName, basePath); err != nil {
		utility.DefaultLogger.Warn("Failed to register base config", err)
	}

	envOrder := []string{"local", "dev", "prod"}
	for _, env := range envOrder {
		overlayFile := filepath.Join(modulaDir, fmt.Sprintf("modula.%s.config.json", env))
		if _, statErr := os.Stat(overlayFile); statErr != nil {
			continue // overlay write failed earlier
		}
		if err := reg.Set(projectName, env, overlayFile); err != nil {
			utility.DefaultLogger.Warn(fmt.Sprintf("Failed to register %s environment", env), err)
		}
	}

	if err := reg.SetDefaultEnv(projectName, "local"); err != nil {
		utility.DefaultLogger.Warn("Failed to set default environment", err)
	}

	if reg.Default == "" {
		if err := reg.SetDefault(projectName); err != nil {
			utility.DefaultLogger.Warn("Failed to set default project", err)
		}
	}

	// Create project directories (certs, search index is created at runtime)
	certsDir := filepath.Join(modulaDir, "certs")
	createAssets := yesFlag
	if !yesFlag {
		if promptErr := huh.NewConfirm().
			Title("Create certificates directory and generate localhost certs?").
			Affirmative("Yes").
			Negative("No").
			Value(&createAssets).Run(); promptErr != nil {
			createAssets = false
		}
	}
	if createAssets {
		if mkErr := os.MkdirAll(certsDir, 0750); mkErr != nil {
			utility.DefaultLogger.Warn("Failed to create certs directory", mkErr)
		} else {
			fmt.Fprintf(out, "  Created %s\n", certsDir)
			if genErr := utility.GenerateSelfSignedCert(certsDir, "localhost"); genErr != nil {
				utility.DefaultLogger.Warn("Failed to generate certificates", genErr)
				fmt.Fprintf(out, "  Run 'modula cert generate' later to create certificates\n")
			} else {
				fmt.Fprintf(out, "  Generated localhost certificates in %s\n", certsDir)
			}
		}
	}

	// Success output
	fmt.Fprintf(out, "\nProject %q created successfully!\n\n", projectName)
	fmt.Fprintf(out, "  Config files:\n")
	fmt.Fprintf(out, "    modula/modula.config.json          (base)\n")
	fmt.Fprintf(out, "    modula/modula.local.config.json    (local)\n")
	fmt.Fprintf(out, "    modula/modula.dev.config.json      (dev)\n")
	fmt.Fprintf(out, "    modula/modula.prod.config.json     (prod)\n")
	fmt.Fprintf(out, "\n  Project assets:\n")
	fmt.Fprintf(out, "    modula/certs/                      (localhost TLS certificates)\n")
	fmt.Fprintf(out, "    modula/search.idx                  (created on first run)\n")
	fmt.Fprintf(out, "\n  Available commands:\n")
	for _, env := range envOrder {
		fmt.Fprintf(out, "    modula serve %s %s\n", projectName, env)
	}
	fmt.Fprintln(out)
	for _, env := range envOrder {
		fmt.Fprintf(out, "    modula connect %s %s\n", projectName, env)
	}
	fmt.Fprintln(out)
	for _, env := range envOrder {
		fmt.Fprintf(out, "    modula tui %s %s\n", projectName, env)
	}
	fmt.Fprintln(out)

	return nil
}

// buildBaseConfig returns a Config with project-level defaults.
// Environment-specific fields (db, credentials, ports) are left at
// zero values since they belong in overlays.
func buildBaseConfig() config.Config {
	c := config.DefaultConfig()

	// Clear environment-specific fields — these go in overlays
	c.Environment = ""
	c.Db_Driver = ""
	c.Db_URL = ""
	c.Db_Name = ""
	c.Db_User = ""
	c.Db_Password = ""
	c.Port = ""
	c.SSL_Port = ""
	c.SSH_Port = ""
	c.SSH_Host = ""
	c.Client_Site = ""
	c.Admin_Site = ""
	c.Auth_Salt = ""
	c.Cert_Dir = ""
	c.Log_Path = ""
	c.Bucket_Endpoint = ""
	c.Bucket_Access_Key = ""
	c.Bucket_Secret_Key = ""
	c.Bucket_Public_URL = ""

	return c
}

// buildLocalOverlay returns the local development overlay.
func buildLocalOverlay() map[string]any {
	return map[string]any{
		"environment": "http-only",
		"db_driver":   "sqlite",
		"db_url":      "./modula.db",
		"db_name":     "modula.db",
		"port":        ":8080",
		"ssl_port":    ":4000",
		"ssh_host":    "localhost",
		"ssh_port":    "2233",
		"client_site": "localhost",
		"admin_site":  "localhost",
		"cert_dir":    "./certs",
		"log_path":    "",
	}
}

// buildDevOverlay returns the dev environment overlay.
func buildDevOverlay() map[string]any {
	return map[string]any{
		"environment": "development",
		"db_driver":   "sqlite",
		"db_url":      "./modula-dev.db",
		"db_name":     "modula-dev.db",
		"port":        ":8080",
		"ssl_port":    ":4000",
		"ssh_host":    "localhost",
		"ssh_port":    "2233",
		"client_site": "localhost",
		"admin_site":  "localhost",
		"cert_dir":    "./certs",
		"log_path":    "",
	}
}

// buildProdOverlay returns the production overlay with env var placeholders.
func buildProdOverlay() map[string]any {
	return map[string]any{
		"environment":    "production",
		"db_driver":      "postgres",
		"db_url":         "${DB_HOST}:5432",
		"db_name":        "${DB_NAME}",
		"db_username":    "${DB_USER}",
		"db_password":    "${DB_PASSWORD}",
		"port":           ":8080",
		"ssl_port":       ":443",
		"ssh_host":       "0.0.0.0",
		"ssh_port":       "2233",
		"client_site":    "${CLIENT_SITE}",
		"admin_site":     "${ADMIN_SITE}",
		"cert_dir":       "${CERT_DIR}",
		"auth_salt":      "${AUTH_SALT}",
		"cookie_secure":  true,
		"bucket_endpoint":   "${BUCKET_ENDPOINT}",
		"bucket_access_key": "${BUCKET_ACCESS_KEY}",
		"bucket_secret_key": "${BUCKET_SECRET_KEY}",
		"bucket_public_url": "${BUCKET_PUBLIC_URL}",
		"log_path":       "/var/log/modulacms",
	}
}

// writeConfigFile marshals a Config struct to indented JSON and writes it.
func writeConfigFile(path string, c config.Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0640)
}

// writeOverlayFile marshals a map to indented JSON and writes it.
func writeOverlayFile(path string, data map[string]any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0640)
}
