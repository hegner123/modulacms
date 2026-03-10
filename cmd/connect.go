package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/registry"
	"github.com/hegner123/modulacms/internal/remote"
	"github.com/hegner123/modulacms/internal/tui"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect [name] [env]",
	Short: "Launch TUI for a registered project",
	Long: `Open the SSH TUI connected to a registered project's config file.

Each project can have multiple environments (local, dev, staging, prod, etc.),
each pointing to a different modula.config.json. The command resolves which config to
use based on the project name and environment you provide.

Resolution order:
  1. If both name and env are given, use that exact project + environment.
  2. If only name is given, use that project's default environment.
  3. If neither is given, use the default project's default environment.
  4. If the registry is empty, look for modula.config.json in the current directory.

Examples:
  modula connect                      # default project, default env
  modula connect mysite               # "mysite" project, its default env
  modula connect mysite prod          # "mysite" project, "prod" env
  modula connect mysite local         # "mysite" project, "local" env`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		reg, err := registry.Load()
		if err != nil {
			return fmt.Errorf("loading registry: %w", err)
		}

		var name, env string
		if len(args) > 0 {
			name = args[0]
		}
		if len(args) > 1 {
			env = args[1]
		}

		configPath, err := reg.Resolve(name, env)
		if err != nil {
			// Auto-detect: when no name given and registry has no default,
			// check for modula.config.json in the current working directory
			if name == "" {
				cwd, wdErr := os.Getwd()
				if wdErr != nil {
					return err // return original registry error
				}
				localCfg := filepath.Join(cwd, config.DefaultConfigFilename)
				if _, statErr := os.Stat(localCfg); statErr == nil {
					configPath = localCfg
				} else {
					return fmt.Errorf("no project specified and no %s in current directory", config.DefaultConfigFilename)
				}
			} else {
				return err
			}
		}

		// chdir to project root so relative paths in config resolve correctly
		projectDir := filepath.Dir(configPath)
		if err := os.Chdir(projectDir); err != nil {
			return fmt.Errorf("changing to project directory %s: %w", projectDir, err)
		}

		// Override the global cfgPath so loadConfig reads the right file
		cfgPath = filepath.Base(configPath)

		mgr, err := loadConfig()
		if err != nil {
			return err
		}

		cfg, err := mgr.Config()
		if err != nil {
			return err
		}

		// Route driver: remote (SDK over HTTPS) or local (database)
		var driver db.DbDriver
		if cfg.Remote_URL != "" {
			driver, err = remote.NewDriver(cfg.Remote_URL, cfg.Remote_API_Key)
			if err != nil {
				return fmt.Errorf("remote connection failed: %w", err)
			}
			// Set singleton so the 83 db.ConfigDB(*cfg) calls in TUI code
			// return this driver via the existing dbInstance fast path
			db.SetInstance(driver)
		} else if cfg.Db_Driver != "" {
			_, driver, err = loadConfigAndDB()
			if err != nil {
				return err
			}
			defer closeDBWithLog()
		} else {
			return fmt.Errorf("modula.config.json must have either remote_url or db_driver")
		}

		model, _ := tui.InitialModel(&verbose, cfg, driver, utility.DefaultLogger, nil, mgr, nil, nil)
		model.IsRemote = cfg.Remote_URL != ""
		model.RemoteURL = cfg.Remote_URL
		if model.IsRemote {
			model.PageMenu = model.HomepageMenuInit()
		}
		if _, ok := tui.CliRun(&model); !ok {
			process, err := os.FindProcess(os.Getpid())
			if err != nil {
				utility.DefaultLogger.Error("", err)
				return err
			}
			if err := process.Signal(syscall.SIGTERM); err != nil {
				utility.DefaultLogger.Error("", err)
				return err
			}
		}

		return nil
	},
}

var connectSetCmd = &cobra.Command{
	Use:   "set <name> <env> <config-path>",
	Short: "Register or update an environment for a project",
	Long: `Set the config path for a project environment in the registry.

Creates the project if it does not exist. Creates the environment if it does not
exist. Overwrites the config path if the environment already exists. The first
environment added to a new project becomes its default.

The config-path must point to an existing modula.config.json file. Relative paths are
resolved against the current working directory and stored as absolute paths in
the registry, so "./modula.config.json" and "/full/path/modula.config.json" are equivalent.

Arguments:
  name         Project name (e.g. mysite, blog, docs)
  env          Environment label (e.g. local, dev, staging, prod)
  config-path  Path to the environment's modula.config.json (relative or absolute)

Examples:
  modula connect set mysite local ./modula.config.json           # register new project + env
  modula connect set mysite prod /srv/mysite/modula.config.json  # add another env
  modula connect set mysite local /new/path/modula.config.json   # update existing env path`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := registry.Load()
		if err != nil {
			return fmt.Errorf("loading registry: %w", err)
		}

		if err := reg.Set(args[0], args[1], args[2]); err != nil {
			return err
		}

		abs, _ := filepath.Abs(args[2])
		fmt.Printf("Set project %q env %q -> %s\n", args[0], args[1], abs)
		return nil
	},
}

var connectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered projects and environments",
	Long: `Show all registered projects and their environments.

The default project is marked with "(default)". Each project's default
environment is marked with an asterisk (*).

Example output:
  mysite (default)
    local -> /home/user/mysite/modula.config.json *
    prod  -> /srv/mysite/prod-config.json
  blog
    dev   -> /home/user/blog/dev-config.json *`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := registry.Load()
		if err != nil {
			return fmt.Errorf("loading registry: %w", err)
		}

		if len(reg.Projects) == 0 {
			fmt.Println("No projects registered. Use: modula connect set <name> <env> <config-path>")
			return nil
		}

		for name, proj := range reg.Projects {
			projectMarker := ""
			if name == reg.Default {
				projectMarker = " (default)"
			}
			fmt.Printf("  %s%s\n", name, projectMarker)

			for _, env := range reg.EnvNames(name) {
				envMarker := ""
				if env == proj.DefaultEnv {
					envMarker = " *"
				}
				fmt.Printf("    %s -> %s%s\n", env, proj.Envs[env], envMarker)
			}
		}
		return nil
	},
}

var connectRemoveCmd = &cobra.Command{
	Use:   "remove <name> [--env <env>]",
	Short: "Remove a project or single environment from the registry",
	Long: `Delete a project and all its environments from the registry.

By default, removes the entire project. Use --env to remove only one
environment while keeping the rest. If the removed environment was the
project's default, the default is cleared. If no environments remain
after removal, the project itself is removed.

If the removed project was the default project, the default is cleared.

Arguments:
  name   Project name to remove

Flags:
  --env  Remove only this environment instead of the whole project

Examples:
  modula connect remove mysite              # remove mysite and all its envs
  modula connect remove mysite --env dev    # remove only the "dev" env
  modula connect remove mysite --env prod   # remove only the "prod" env`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := registry.Load()
		if err != nil {
			return fmt.Errorf("loading registry: %w", err)
		}

		envFlag, _ := cmd.Flags().GetString("env")

		if envFlag != "" {
			if err := reg.RemoveEnv(args[0], envFlag); err != nil {
				return err
			}
			fmt.Printf("Removed environment %q from project %q\n", envFlag, args[0])
			return nil
		}

		if err := reg.Remove(args[0]); err != nil {
			return err
		}

		fmt.Printf("Removed project %q\n", args[0])
		return nil
	},
}

var connectDefaultCmd = &cobra.Command{
	Use:   "default <name> [env]",
	Short: "Set the default project or default environment",
	Long: `Configure which project or environment is used when arguments are omitted.

With one argument: sets the default project. Running "modula connect" with
no arguments will use this project.

With two arguments: sets the default environment for a project. Running
"modula connect <name>" without an env will use this environment.

The project (and environment, if given) must already be registered.

Arguments:
  name   Project name (required)
  env    Environment name (optional — sets the default env for the project)

Examples:
  modula connect default mysite            # "modula connect" now uses mysite
  modula connect default mysite prod       # "modula connect mysite" now uses prod
  modula connect default blog local        # "modula connect blog" now uses local`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		reg, err := registry.Load()
		if err != nil {
			return fmt.Errorf("loading registry: %w", err)
		}

		if len(args) == 2 {
			if err := reg.SetDefaultEnv(args[0], args[1]); err != nil {
				return err
			}
			fmt.Printf("Default environment for %q set to %q\n", args[0], args[1])
			return nil
		}

		if err := reg.SetDefault(args[0]); err != nil {
			return err
		}

		fmt.Printf("Default project set to %q\n", args[0])
		return nil
	},
}

func init() {
	connectCmd.AddCommand(connectSetCmd)
	connectCmd.AddCommand(connectListCmd)
	connectCmd.AddCommand(connectRemoveCmd)
	connectCmd.AddCommand(connectDefaultCmd)

	connectRemoveCmd.Flags().String("env", "", "remove a single environment instead of the entire project")
}
