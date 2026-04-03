package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	plugintesting "github.com/hegner123/modulacms/internal/plugin/testing"
)

var pluginTestCmd = &cobra.Command{
	Use:   "test <path>",
	Short: "Run plugin tests",
	Long: `Run automated tests for a Lua plugin.

Discovers test/*.test.lua files in the plugin directory, loads the plugin into
an isolated in-memory SQLite environment, and executes all test_* functions
with assertion support and SAVEPOINT-based isolation.

Exit codes:
  0  All tests passed
  1  One or more tests failed
  2  Plugin load error

Examples:
  modula plugin test ./plugins/x402
  modula plugin test ./plugins/x402 --verbose
  modula plugin test ./plugins/x402 --filter test_create
  modula plugin test ./plugins/x402 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginTest,
}

func runPluginTest(cmd *cobra.Command, args []string) error {
	pluginDir := args[0]
	verbose, _ := cmd.Flags().GetBool("verbose")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	filter, _ := cmd.Flags().GetString("filter")
	timeoutSec, _ := cmd.Flags().GetInt("timeout")

	opts := plugintesting.HarnessOpts{
		Verbose: verbose,
		Timeout: time.Duration(timeoutSec) * time.Second,
		Filter:  filter,
	}

	harness, err := plugintesting.NewHarness(pluginDir, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}
	defer harness.Close()

	files, err := harness.DiscoverTests()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}

	ctx := context.Background()
	report := harness.RunAll(ctx, files)

	if jsonOutput {
		if err := report.WriteNDJSON(os.Stdout); err != nil {
			return fmt.Errorf("write JSON output: %w", err)
		}
	} else {
		report.WriteHuman(os.Stdout, verbose)
	}

	os.Exit(report.ExitCode())
	return nil
}
