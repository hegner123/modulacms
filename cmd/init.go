package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/huh/v2"
	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/install"
	"github.com/hegner123/modulacms/internal/registry"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var (
	initAdminPassword string
	initProjectName   string
	initMode          string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a ModulaCMS project (idempotent)",
	Long: `Initialize a ModulaCMS project in the current directory. Each step checks
whether its output already exists and skips if so. Safe to run multiple times.

Steps performed:
  1. Load or create the project registry (~/.modula/configs.json)
  2. Create modula/ project directory (if not present)
  3. Write base config and environment overlays (if not present)
  4. Register any unregistered config files in the registry
  5. Create certificates directory and generate localhost certs (if not present)
  6. Create and seed the database (only for SQLite when .db file is missing;
     skipped entirely when the config specifies an external database)

Modes:
  interactive   Prompt for project name and admin password (default)
  ci            No prompts; requires --admin-password
  container     No prompts; skips database creation entirely

Examples:
  modula init                                       # interactive
  modula init --mode ci --admin-password s3cret!    # CI pipeline
  modula init --mode container                      # Docker entrypoint
  modula init --name my-site                        # custom project name`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVar(&initMode, "mode", "interactive", "Init mode: interactive, ci, container")
	initCmd.Flags().StringVar(&initAdminPassword, "admin-password", "", "System admin password (required for ci mode)")
	initCmd.Flags().StringVar(&initProjectName, "name", "", "Project name (default: current directory name)")
}

// initResult tracks what each step did.
type initResult struct {
	name   string
	action string // "created", "skipped", "registered"
	detail string
}

func runInit(cmd *cobra.Command, args []string) error {
	configureLogger()

	out := cmd.OutOrStdout()
	mode := initMode

	// --yes global flag maps to ci mode for backward compatibility
	if yesFlag && mode == "interactive" {
		mode = "ci"
	}

	switch mode {
	case "interactive", "ci", "container":
	default:
		return fmt.Errorf("unknown mode %q: must be interactive, ci, or container", mode)
	}

	interactive := mode == "interactive" && isatty.IsTerminal(os.Stdout.Fd())

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}

	projectName := initProjectName
	if projectName == "" {
		projectName = filepath.Base(cwd)
	}

	// Interactive: prompt for project name
	if interactive && initProjectName == "" {
		suggested := projectName
		promptErr := huh.NewInput().
			Title("Project name").
			Value(&projectName).
			Placeholder(suggested).Run()
		if promptErr != nil {
			return nil
		}
		projectName = strings.TrimSpace(projectName)
		if projectName == "" {
			projectName = suggested
		}
	}

	var results []initResult

	// Step 1: Registry
	reg, err := registry.Load()
	if err != nil {
		// File doesn't exist yet, create empty registry
		reg = &registry.Registry{Projects: make(map[string]*registry.Project)}
		if saveErr := reg.Save(); saveErr != nil {
			return fmt.Errorf("creating registry: %w", saveErr)
		}
		results = append(results, initResult{"registry", "created", "~/.modula/configs.json"})
	} else {
		results = append(results, initResult{"registry", "skipped", "already exists"})
	}

	// Step 2: Project directory
	// If cwd already has modula.config.json, treat cwd as the project dir.
	// Otherwise use modula/ subdirectory.
	projectDir := filepath.Join(cwd, "modula")
	if _, statErr := os.Stat(filepath.Join(cwd, "modula.config.json")); statErr == nil {
		projectDir = cwd
	}

	if _, statErr := os.Stat(projectDir); statErr != nil {
		if mkErr := os.MkdirAll(projectDir, 0750); mkErr != nil {
			return fmt.Errorf("creating project directory: %w", mkErr)
		}
		results = append(results, initResult{"project dir", "created", projectDir})
	} else {
		results = append(results, initResult{"project dir", "skipped", projectDir})
	}

	// Step 3: Base config
	basePath := filepath.Join(projectDir, "modula.config.json")
	if _, statErr := os.Stat(basePath); statErr != nil {
		baseConfig := buildBaseConfig()
		if wErr := writeConfigFile(basePath, baseConfig); wErr != nil {
			return fmt.Errorf("writing base config: %w", wErr)
		}
		results = append(results, initResult{"base config", "created", basePath})
	} else {
		results = append(results, initResult{"base config", "skipped", "already exists"})
	}

	// Step 4: Overlay configs
	overlayDefs := []struct {
		env   string
		build func() map[string]any
	}{
		{"local", buildLocalOverlay},
		{"dev", buildDevOverlay},
		{"prod", buildProdOverlay},
	}

	for _, od := range overlayDefs {
		overlayFile := filepath.Join(projectDir, fmt.Sprintf("modula.%s.config.json", od.env))
		if _, statErr := os.Stat(overlayFile); statErr != nil {
			if wErr := writeOverlayFile(overlayFile, od.build()); wErr != nil {
				utility.DefaultLogger.Warn(fmt.Sprintf("failed to write %s overlay", od.env), wErr)
				results = append(results, initResult{fmt.Sprintf("overlay: %s", od.env), "error", wErr.Error()})
				continue
			}
			results = append(results, initResult{fmt.Sprintf("overlay: %s", od.env), "created", overlayFile})
		} else {
			results = append(results, initResult{fmt.Sprintf("overlay: %s", od.env), "skipped", "already exists"})
		}
	}

	// Step 5: Register configs
	// Set base if not registered
	registered := 0
	if err := reg.SetBase(projectName, basePath); err != nil {
		utility.DefaultLogger.Warn("failed to register base config", err)
	}

	envOrder := []string{"local", "dev", "prod"}
	for _, env := range envOrder {
		overlayFile := filepath.Join(projectDir, fmt.Sprintf("modula.%s.config.json", env))
		if _, statErr := os.Stat(overlayFile); statErr != nil {
			continue
		}
		// Check if already registered
		_, proj := reg.FindByDir(cwd)
		if proj != nil && proj.Envs[env] == overlayFile {
			continue
		}
		if err := reg.Set(projectName, env, overlayFile); err != nil {
			utility.DefaultLogger.Warn(fmt.Sprintf("failed to register %s environment", env), err)
			continue
		}
		registered++
	}

	if err := reg.SetDefaultEnv(projectName, "local"); err != nil {
		utility.DefaultLogger.Warn("failed to set default environment", err)
	}
	if reg.Default == "" {
		if err := reg.SetDefault(projectName); err != nil {
			utility.DefaultLogger.Warn("failed to set default project", err)
		}
	}

	if registered > 0 {
		results = append(results, initResult{"register", "registered", fmt.Sprintf("%d configs registered", registered)})
	} else {
		results = append(results, initResult{"register", "skipped", "all configs already registered"})
	}

	// Step 6: Certificates
	certsDir := filepath.Join(projectDir, "certs")
	if _, statErr := os.Stat(certsDir); statErr != nil {
		if mkErr := os.MkdirAll(certsDir, 0750); mkErr != nil {
			utility.DefaultLogger.Warn("failed to create certs directory", mkErr)
			results = append(results, initResult{"certificates", "error", mkErr.Error()})
		} else {
			if genErr := utility.GenerateSelfSignedCert(certsDir, "localhost"); genErr != nil {
				utility.DefaultLogger.Warn("failed to generate certificates", genErr)
				results = append(results, initResult{"certificates", "error", genErr.Error()})
			} else {
				results = append(results, initResult{"certificates", "created", certsDir})
			}
		}
	} else {
		results = append(results, initResult{"certificates", "skipped", "already exists"})
	}

	// Step 7: Database (single decision)
	dbResult := stepDatabase(mode, interactive, projectDir, basePath)
	results = append(results, dbResult)

	// Print report
	fmt.Fprintf(out, "\nModulaCMS init complete: %s\n\n", projectName)
	for _, r := range results {
		marker := "  "
		switch r.action {
		case "created", "registered":
			marker = "+ "
		case "skipped":
			marker = "  "
		case "error":
			marker = "! "
		}
		fmt.Fprintf(out, "  %s%-16s %s\n", marker, r.name, r.detail)
	}
	fmt.Fprintln(out)

	return nil
}

// stepDatabase handles the single database decision:
//   - container mode: always skip
//   - non-sqlite driver: skip
//   - sqlite .db exists: skip
//   - sqlite .db missing: create + seed
func stepDatabase(mode string, interactive bool, projectDir, basePath string) initResult {
	if mode == "container" {
		return initResult{"database", "skipped", "container mode (external database)"}
	}

	// Load the local overlay to determine DB config
	localOverlay := filepath.Join(projectDir, "modula.local.config.json")
	var cfg *config.Config

	if _, statErr := os.Stat(localOverlay); statErr == nil {
		lp := config.NewLayeredFileProvider(basePath, localOverlay)
		mgr := config.NewManager(lp)
		if loadErr := mgr.Load(); loadErr != nil {
			return initResult{"database", "error", fmt.Sprintf("loading config: %s", loadErr)}
		}
		var cfgErr error
		cfg, cfgErr = mgr.Config()
		if cfgErr != nil {
			return initResult{"database", "error", fmt.Sprintf("reading config: %s", cfgErr)}
		}
	} else {
		// No local overlay, load base only
		fp := config.NewFileProvider(basePath)
		mgr := config.NewManager(fp)
		if loadErr := mgr.Load(); loadErr != nil {
			return initResult{"database", "error", fmt.Sprintf("loading config: %s", loadErr)}
		}
		var cfgErr error
		cfg, cfgErr = mgr.Config()
		if cfgErr != nil {
			return initResult{"database", "error", fmt.Sprintf("reading config: %s", cfgErr)}
		}
	}

	// Non-sqlite driver means external DB, skip entirely
	if cfg.Db_Driver != "" && cfg.Db_Driver != config.Sqlite {
		return initResult{"database", "skipped", fmt.Sprintf("external database (%s)", cfg.Db_Driver)}
	}

	// Resolve the DB file path
	dbURL := cfg.Db_URL
	if dbURL == "" {
		dbURL = "modula.db"
	}
	dbPath := dbURL
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(projectDir, dbPath)
	}

	// If DB file exists, skip
	if _, statErr := os.Stat(dbPath); statErr == nil {
		return initResult{"database", "skipped", fmt.Sprintf("%s already exists", dbPath)}
	}

	// Need admin password to seed bootstrap data
	adminHash, err := resolveAdminPassword(mode, interactive)
	if err != nil {
		return initResult{"database", "error", err.Error()}
	}

	// Temporarily set the DB URL to the absolute path so CreateDbSimple
	// can find it regardless of working directory.
	originalURL := cfg.Db_URL
	cfg.Db_URL = dbPath
	if cfg.Db_Driver == "" {
		cfg.Db_Driver = config.Sqlite
	}
	if cfg.Db_Name == "" {
		cfg.Db_Name = "modula_db"
	}

	if setupErr := install.CreateDbSimple(basePath, cfg, adminHash); setupErr != nil {
		cfg.Db_URL = originalURL
		return initResult{"database", "error", fmt.Sprintf("setup failed: %s", setupErr)}
	}
	cfg.Db_URL = originalURL

	return initResult{"database", "created", dbPath}
}

// resolveAdminPassword gets the admin password hash based on the init mode.
func resolveAdminPassword(mode string, interactive bool) (string, error) {
	if initAdminPassword != "" {
		if err := install.ValidatePassword(initAdminPassword); err != nil {
			return "", fmt.Errorf("invalid admin password: %w", err)
		}
		return auth.HashPassword(initAdminPassword)
	}

	if mode == "ci" {
		return "", fmt.Errorf("--admin-password is required in ci mode")
	}

	if interactive {
		password := ""
		confirm := ""

		f1 := huh.NewInput().
			Title("System admin password (min 8 characters)").
			Value(&password).
			EchoMode(huh.EchoModePassword).
			Validate(install.ValidatePassword)

		f2 := huh.NewInput().
			Title("Confirm admin password").
			Value(&confirm).
			EchoMode(huh.EchoModePassword).
			Validate(install.ValidatePassword)

		g := huh.NewGroup(f1, f2)
		f := huh.NewForm(g)
		if err := f.Run(); err != nil {
			return "", fmt.Errorf("password prompt: %w", err)
		}

		if password != confirm {
			return "", fmt.Errorf("passwords do not match")
		}

		return auth.HashPassword(password)
	}

	// Non-interactive, no password provided: generate one
	autoPassword, err := utility.MakeRandomString()
	if err != nil {
		return "", fmt.Errorf("generating password: %w", err)
	}
	utility.DefaultLogger.Finfo("generated system admin password", "email", "system@modula.local", "password", autoPassword)
	return auth.HashPassword(autoPassword)
}
