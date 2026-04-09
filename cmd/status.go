package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/hegner123/modulacms/internal/registry"
	"github.com/hegner123/modulacms/internal/utility"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project status for the current directory",
	Long: `Display the registration status, config files, environments, and available
commands for the ModulaCMS project in the current directory.

If the current directory is not a registered project, suggests running 'modula init'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		reg, err := registry.Load()
		if err != nil {
			utility.DefaultLogger.Warn("failed to load project registry", err)
			reg = &registry.Registry{Projects: make(map[string]*registry.Project)}
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine working directory: %w", err)
		}

		name, proj := reg.FindByDir(cwd)
		if proj == nil {
			fmt.Fprintln(cmd.OutOrStdout(), "no ModulaCMS project found in this directory.")
			fmt.Fprintln(cmd.OutOrStdout(), "Run 'modula init' to set one up.")
			return nil
		}

		return showProjectStatus(cmd, reg, name, proj)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

// showProjectStatus displays the status of a registered project.
func showProjectStatus(cmd *cobra.Command, reg *registry.Registry, name string, proj *registry.Project) error {
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
