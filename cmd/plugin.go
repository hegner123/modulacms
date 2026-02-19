package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/hegner123/modulacms/internal/plugin"
	"github.com/hegner123/modulacms/internal/utility"
)

// pluginTokenFlag holds the --token flag value for CI/CD use. When set, it
// overrides the token file read by pluginAPIClient().
var pluginTokenFlag string

// pluginCmd is the parent command for all plugin management operations.
var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Plugin management commands",
}

// pluginAPIClient reads config.json for the server port and the admin API token
// from <config_dir>/.plugin-api-token (or --token flag override). Returns a
// configured apiClient targeting localhost:<port> with Bearer token auth.
//
// Exits with code 1 (via cobra RunE error return) if the token file is missing
// or unreadable and no --token flag was provided.
type apiClient struct {
	baseURL string
	token   string
	http    *http.Client
}

// newPluginAPIClient creates an apiClient from config + token file/flag.
func newPluginAPIClient() (*apiClient, error) {
	cfg, err := loadConfigPtr()
	if err != nil {
		return nil, fmt.Errorf("loading configuration: %w", err)
	}

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	// Resolve token: --token flag takes precedence over token file.
	token := pluginTokenFlag
	if token == "" {
		// Resolve config_dir from the --config flag (cfgPath is the global).
		configDir := filepath.Dir(cfgPath)
		tokenPath := filepath.Join(configDir, ".plugin-api-token")

		tokenBytes, readErr := os.ReadFile(tokenPath)
		if readErr != nil {
			return nil, fmt.Errorf("server not running or token file missing: %w", readErr)
		}
		token = strings.TrimSpace(string(tokenBytes))
	}

	if token == "" {
		return nil, fmt.Errorf("server not running or token file missing")
	}

	return &apiClient{
		baseURL: "http://localhost:" + port,
		token:   token,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// do sends an HTTP request with the Bearer token header and returns the response.
// The caller is responsible for closing resp.Body.
func (c *apiClient) do(method, path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request to %s: %w", path, err)
	}
	return resp, nil
}

// pluginListCmd scans the plugin directory and prints a summary table.
// Offline: reads filesystem only, no database connection needed.
var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		cfg, err := loadConfigPtr()
		if err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}

		dir := cfg.Plugin_Directory
		if dir == "" {
			dir = "./plugins/"
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintln(cmd.OutOrStdout(), "No plugins directory found.")
				return nil
			}
			return fmt.Errorf("reading plugin directory %q: %w", dir, err)
		}

		// Collect results for each subdirectory.
		type pluginRow struct {
			name    string
			version string
			desc    string
			invalid bool
		}
		var rows []pluginRow

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			pluginDir := filepath.Join(dir, entry.Name())
			info, results, validateErr := plugin.ValidatePlugin(pluginDir)
			if validateErr != nil {
				// I/O error -- still show the entry as invalid.
				rows = append(rows, pluginRow{
					name:    entry.Name(),
					invalid: true,
				})
				continue
			}

			// Check if any result has error severity.
			hasErrors := false
			for _, r := range results {
				if r.Severity == plugin.SeverityError {
					hasErrors = true
					break
				}
			}

			if hasErrors {
				rows = append(rows, pluginRow{
					name:    entry.Name(),
					invalid: true,
				})
				continue
			}

			rows = append(rows, pluginRow{
				name:    info.Name,
				version: info.Version,
				desc:    info.Description,
			})
		}

		if len(rows) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No plugins found.")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tVERSION\tDESCRIPTION")
		for _, r := range rows {
			if r.invalid {
				fmt.Fprintf(w, "%s [invalid]\t\t\n", r.name)
				continue
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", r.name, r.version, r.desc)
		}
		return w.Flush()
	},
}

// Flags for plugin init (non-interactive mode).
var (
	pluginInitVersion     string
	pluginInitDescription string
	pluginInitAuthor      string
	pluginInitLicense     string
)

// pluginInitCmd scaffolds a new plugin directory with init.lua and lib/.
// Offline: writes to filesystem only, no database connection needed.
var pluginInitCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Create a new plugin scaffold",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		name := args[0]

		// Validate the name using the same rules as the Manager's manifest validation.
		// Construct a minimal PluginInfo with just the name and placeholder required fields
		// so ValidateManifest checks only the name rules. The version/description are
		// placeholders -- ValidateManifest requires them non-empty.
		nameCheck := &plugin.PluginInfo{
			Name:        name,
			Version:     "0.0.0",
			Description: "placeholder",
		}
		if err := plugin.ValidateManifest(nameCheck); err != nil {
			return fmt.Errorf("invalid plugin name: %w", err)
		}

		cfg, err := loadConfigPtr()
		if err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}

		dir := cfg.Plugin_Directory
		if dir == "" {
			dir = "./plugins/"
		}

		pluginDir := filepath.Join(dir, name)
		if _, statErr := os.Stat(pluginDir); statErr == nil {
			return fmt.Errorf("plugin directory already exists: %s", pluginDir)
		}

		// Determine whether to use interactive prompts.
		// Skip huh.Form when all flags are provided or stdout is not a terminal.
		allFlagsProvided := pluginInitVersion != "" &&
			pluginInitDescription != "" &&
			pluginInitAuthor != "" &&
			pluginInitLicense != ""
		interactive := !allFlagsProvided && isatty.IsTerminal(os.Stdout.Fd())

		version := pluginInitVersion
		description := pluginInitDescription
		author := pluginInitAuthor
		license := pluginInitLicense

		if interactive {
			// Set defaults for fields not provided via flags.
			if version == "" {
				version = "0.1.0"
			}
			if license == "" {
				license = "MIT"
			}

			form := huh.NewForm(huh.NewGroup(
				huh.NewInput().
					Title("Version").
					Value(&version).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return fmt.Errorf("version is required")
						}
						return nil
					}),
				huh.NewInput().
					Title("Description").
					Value(&description).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return fmt.Errorf("description is required")
						}
						return nil
					}),
				huh.NewInput().
					Title("Author").
					Value(&author),
				huh.NewInput().
					Title("License").
					Value(&license),
			))

			if formErr := form.Run(); formErr != nil {
				if errors.Is(formErr, huh.ErrUserAborted) {
					utility.DefaultLogger.Info("Plugin init cancelled")
					return nil
				}
				return fmt.Errorf("form error: %w", formErr)
			}
		}

		// In non-interactive mode with missing required fields, error.
		if version == "" {
			version = "0.1.0"
		}
		if description == "" && !interactive {
			return fmt.Errorf("--description is required in non-interactive mode")
		}

		// Create plugin directories.
		if err := os.MkdirAll(filepath.Join(pluginDir, "lib"), 0750); err != nil {
			return fmt.Errorf("creating plugin directories: %w", err)
		}

		// Write init.lua with manifest skeleton and lifecycle stubs.
		initContent := generateInitLua(name, version, description, author, license)
		initPath := filepath.Join(pluginDir, "init.lua")
		if err := os.WriteFile(initPath, []byte(initContent), 0644); err != nil {
			return fmt.Errorf("writing init.lua: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created plugin scaffold at %s\n", pluginDir)
		return nil
	},
}

// pluginValidateCmd validates a plugin directory without loading it into the runtime.
// Offline: parses Lua and checks manifest only, no database connection needed.
var pluginValidateCmd = &cobra.Command{
	Use:   "validate <path>",
	Short: "Validate a plugin without loading",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		pluginPath := args[0]
		info, results, err := plugin.ValidatePlugin(pluginPath)
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		hasErrors := false
		for _, r := range results {
			switch r.Severity {
			case plugin.SeverityError:
				hasErrors = true
				fmt.Fprintf(cmd.ErrOrStderr(), "  ERROR   [%s] %s\n", r.Field, r.Message)
			case plugin.SeverityWarning:
				fmt.Fprintf(cmd.ErrOrStderr(), "  WARN    [%s] %s\n", r.Field, r.Message)
			}
		}

		if hasErrors {
			fmt.Fprintln(cmd.ErrOrStderr(), "Validation failed.")
			os.Exit(1)
		}

		// Print summary on success.
		fmt.Fprintf(cmd.OutOrStdout(), "Plugin %q v%s is valid.\n", info.Name, info.Version)

		// Print any warnings that were found (already printed above to stderr).
		warningCount := 0
		for _, r := range results {
			if r.Severity == plugin.SeverityWarning {
				warningCount++
			}
		}
		if warningCount > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "  %d warning(s) found.\n", warningCount)
		}

		return nil
	},
}

// pluginInfoCmd retrieves detailed information about a plugin from the running server.
// Online: sends GET to /api/v1/admin/plugins/{name}.
var pluginInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show detailed plugin information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		client, err := newPluginAPIClient()
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		name := args[0]
		resp, err := client.do("GET", "/api/v1/admin/plugins/"+name, nil)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "request failed: %s\n", err.Error())
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			fmt.Fprintf(cmd.ErrOrStderr(), "plugin not found: %s\n", name)
			os.Exit(1)
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(cmd.ErrOrStderr(), "server error (%d): %s\n", resp.StatusCode, strings.TrimSpace(string(body)))
			os.Exit(1)
		}

		// Decode the response matching PluginInfoHandler's infoJSON structure.
		var info struct {
			Name         string `json:"name"`
			Version      string `json:"version"`
			Description  string `json:"description"`
			Author       string `json:"author"`
			License      string `json:"license"`
			State        string `json:"state"`
			FailedReason string `json:"failed_reason"`
			CBState      string `json:"circuit_breaker_state"`
			CBErrors     int    `json:"circuit_breaker_errors"`
			VMsAvailable int    `json:"vms_available"`
			VMsTotal     int    `json:"vms_total"`
			Dependencies []string `json:"dependencies"`
			SchemaDrift  []struct {
				Table  string `json:"table"`
				Kind   string `json:"kind"`
				Column string `json:"column"`
			} `json:"schema_drift"`
		}

		if decErr := json.NewDecoder(resp.Body).Decode(&info); decErr != nil {
			return fmt.Errorf("decoding response: %w", decErr)
		}

		// Print readable key: value output.
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Name:          %s\n", info.Name)
		fmt.Fprintf(out, "Version:       %s\n", info.Version)
		fmt.Fprintf(out, "Description:   %s\n", info.Description)
		if info.Author != "" {
			fmt.Fprintf(out, "Author:        %s\n", info.Author)
		}
		if info.License != "" {
			fmt.Fprintf(out, "License:       %s\n", info.License)
		}
		fmt.Fprintf(out, "State:         %s\n", info.State)
		if info.FailedReason != "" {
			fmt.Fprintf(out, "Failed Reason: %s\n", info.FailedReason)
		}
		if info.CBState != "" {
			fmt.Fprintf(out, "CB State:      %s\n", info.CBState)
			fmt.Fprintf(out, "CB Errors:     %d\n", info.CBErrors)
		}
		fmt.Fprintf(out, "VMs:           %d/%d available\n", info.VMsAvailable, info.VMsTotal)
		if len(info.Dependencies) > 0 {
			fmt.Fprintf(out, "Dependencies:  %s\n", strings.Join(info.Dependencies, ", "))
		}
		if len(info.SchemaDrift) > 0 {
			fmt.Fprintln(out, "Schema Drift:")
			for _, d := range info.SchemaDrift {
				fmt.Fprintf(out, "  - %s: %s column %q\n", d.Table, d.Kind, d.Column)
			}
		}

		return nil
	},
}

// pluginReloadCmd triggers a hot reload of a plugin on the running server.
// Online: sends POST to /api/v1/admin/plugins/{name}/reload.
var pluginReloadCmd = &cobra.Command{
	Use:   "reload <name>",
	Short: "Hot reload a plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		client, err := newPluginAPIClient()
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		name := args[0]
		resp, err := client.do("POST", "/api/v1/admin/plugins/"+name+"/reload", nil)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "request failed: %s\n", err.Error())
			os.Exit(1)
		}
		defer resp.Body.Close()

		return handleSimpleResponse(cmd, resp, name, "reloaded")
	},
}

// pluginEnableCmd re-enables a disabled plugin on the running server.
// Online: sends POST to /api/v1/admin/plugins/{name}/enable.
var pluginEnableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a disabled plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		client, err := newPluginAPIClient()
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		name := args[0]
		resp, err := client.do("POST", "/api/v1/admin/plugins/"+name+"/enable", nil)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "request failed: %s\n", err.Error())
			os.Exit(1)
		}
		defer resp.Body.Close()

		return handleSimpleResponse(cmd, resp, name, "enabled")
	},
}

// pluginDisableCmd disables a running plugin on the running server.
// Online: sends POST to /api/v1/admin/plugins/{name}/disable.
var pluginDisableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a running plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		client, err := newPluginAPIClient()
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		name := args[0]
		resp, err := client.do("POST", "/api/v1/admin/plugins/"+name+"/disable", nil)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "request failed: %s\n", err.Error())
			os.Exit(1)
		}
		defer resp.Body.Close()

		return handleSimpleResponse(cmd, resp, name, "disabled")
	},
}

// handleSimpleResponse handles the common pattern for reload/enable/disable responses.
// Prints success or error message and exits appropriately.
func handleSimpleResponse(cmd *cobra.Command, resp *http.Response, name, action string) error {
	if resp.StatusCode == http.StatusNotFound {
		fmt.Fprintf(cmd.ErrOrStderr(), "plugin not found: %s\n", name)
		os.Exit(1)
	}

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(cmd.ErrOrStderr(), "failed to %s plugin %q: %s\n", action, name, strings.TrimSpace(string(body)))
		os.Exit(1)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Plugin %q %s successfully.\n", name, action)
	return nil
}

// Flags for plugin approve/revoke.
var (
	pluginApproveRoute     string
	pluginApproveHook      string
	pluginApproveAllRoutes bool
	pluginApproveAllHooks  bool
	pluginApproveYes       bool

	pluginRevokeRoute     string
	pluginRevokeHook      string
	pluginRevokeAllRoutes bool
	pluginRevokeAllHooks  bool
	pluginRevokeYes       bool
)

// pluginApproveCmd approves routes and/or hooks for a plugin.
// Online: sends POST to routes/approve and/or hooks/approve endpoints.
var pluginApproveCmd = &cobra.Command{
	Use:   "approve <name>",
	Short: "Approve plugin routes or hooks",
	Long: `Approve plugin routes or hooks for execution.

Examples:
  modula plugin approve my_plugin --route "GET /tasks"
  modula plugin approve my_plugin --hook "after_insert:content_data"
  modula plugin approve my_plugin --all-routes
  modula plugin approve my_plugin --all-hooks --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		name := args[0]

		// Validate that at least one flag is provided.
		if pluginApproveRoute == "" && pluginApproveHook == "" && !pluginApproveAllRoutes && !pluginApproveAllHooks {
			return fmt.Errorf("at least one of --route, --hook, --all-routes, or --all-hooks is required")
		}

		client, err := newPluginAPIClient()
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		// Handle --route flag: approve a single route.
		if pluginApproveRoute != "" {
			method, path, parseErr := parseRouteFlag(pluginApproveRoute)
			if parseErr != nil {
				return parseErr
			}
			return approveRoutes(cmd, client, name, []routeItem{{Plugin: name, Method: method, Path: path}})
		}

		// Handle --hook flag: approve a single hook.
		if pluginApproveHook != "" {
			event, table, parseErr := parseHookFlag(pluginApproveHook)
			if parseErr != nil {
				return parseErr
			}
			return approveHooks(cmd, client, name, []hookItem{{Plugin: name, Event: event, Table: table}})
		}

		// Handle --all-routes: fetch unapproved routes, confirm, approve.
		if pluginApproveAllRoutes {
			return approveAllRoutes(cmd, client, name, pluginApproveYes)
		}

		// Handle --all-hooks: fetch unapproved hooks, confirm, approve.
		if pluginApproveAllHooks {
			return approveAllHooks(cmd, client, name, pluginApproveYes)
		}

		return nil
	},
}

// pluginRevokeCmd revokes approval for routes and/or hooks for a plugin.
// Online: sends POST to routes/revoke and/or hooks/revoke endpoints.
var pluginRevokeCmd = &cobra.Command{
	Use:   "revoke <name>",
	Short: "Revoke plugin route or hook approvals",
	Long: `Revoke approval for plugin routes or hooks.

Examples:
  modula plugin revoke my_plugin --route "GET /tasks"
  modula plugin revoke my_plugin --hook "after_insert:content_data"
  modula plugin revoke my_plugin --all-routes
  modula plugin revoke my_plugin --all-hooks --yes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		name := args[0]

		// Validate that at least one flag is provided.
		if pluginRevokeRoute == "" && pluginRevokeHook == "" && !pluginRevokeAllRoutes && !pluginRevokeAllHooks {
			return fmt.Errorf("at least one of --route, --hook, --all-routes, or --all-hooks is required")
		}

		client, err := newPluginAPIClient()
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
			os.Exit(1)
		}

		// Handle --route flag: revoke a single route.
		if pluginRevokeRoute != "" {
			method, path, parseErr := parseRouteFlag(pluginRevokeRoute)
			if parseErr != nil {
				return parseErr
			}
			return revokeRoutes(cmd, client, name, []routeItem{{Plugin: name, Method: method, Path: path}})
		}

		// Handle --hook flag: revoke a single hook.
		if pluginRevokeHook != "" {
			event, table, parseErr := parseHookFlag(pluginRevokeHook)
			if parseErr != nil {
				return parseErr
			}
			return revokeHooks(cmd, client, name, []hookItem{{Plugin: name, Event: event, Table: table}})
		}

		// Handle --all-routes: fetch approved routes, confirm, revoke.
		if pluginRevokeAllRoutes {
			return revokeAllRoutes(cmd, client, name, pluginRevokeYes)
		}

		// Handle --all-hooks: fetch approved hooks, confirm, revoke.
		if pluginRevokeAllHooks {
			return revokeAllHooks(cmd, client, name, pluginRevokeYes)
		}

		return nil
	},
}

// routeItem matches the JSON structure expected by the routes approve/revoke endpoints.
type routeItem struct {
	Plugin string `json:"plugin"`
	Method string `json:"method"`
	Path   string `json:"path"`
}

// hookItem matches the JSON structure expected by the hooks approve/revoke endpoints.
type hookItem struct {
	Plugin string `json:"plugin"`
	Event  string `json:"event"`
	Table  string `json:"table"`
}

// parseRouteFlag parses a route flag value like "GET /path" into method and path.
func parseRouteFlag(flag string) (method, path string, err error) {
	parts := strings.SplitN(flag, " ", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid --route format: expected \"METHOD /path\", got %q", flag)
	}
	method = strings.ToUpper(strings.TrimSpace(parts[0]))
	path = strings.TrimSpace(parts[1])
	if method == "" || path == "" {
		return "", "", fmt.Errorf("invalid --route format: method and path must be non-empty")
	}
	return method, path, nil
}

// parseHookFlag parses a hook flag value like "after_insert:content_data" into event and table.
func parseHookFlag(flag string) (event, table string, err error) {
	parts := strings.SplitN(flag, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid --hook format: expected \"event:table\", got %q", flag)
	}
	event = strings.TrimSpace(parts[0])
	table = strings.TrimSpace(parts[1])
	if event == "" || table == "" {
		return "", "", fmt.Errorf("invalid --hook format: event and table must be non-empty")
	}
	return event, table, nil
}

// fetchRoutesFiltered fetches routes and filters by plugin name and approval status.
func fetchRoutesFiltered(client *apiClient, pluginName string, wantApproved bool) ([]routeItem, error) {
	resp, err := client.do("GET", "/api/v1/admin/plugins/routes", nil)
	if err != nil {
		return nil, fmt.Errorf("fetching routes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		Routes []struct {
			Plugin   string `json:"plugin"`
			Method   string `json:"method"`
			Path     string `json:"path"`
			Approved bool   `json:"approved"`
		} `json:"routes"`
	}

	if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr != nil {
		return nil, fmt.Errorf("decoding routes: %w", decErr)
	}

	var items []routeItem
	for _, r := range result.Routes {
		if r.Plugin == pluginName && r.Approved == wantApproved {
			items = append(items, routeItem{Plugin: r.Plugin, Method: r.Method, Path: r.Path})
		}
	}
	return items, nil
}

// fetchHooksFiltered fetches hooks and filters by plugin name and approval status.
func fetchHooksFiltered(client *apiClient, pluginName string, wantApproved bool) ([]hookItem, error) {
	resp, err := client.do("GET", "/api/v1/admin/plugins/hooks", nil)
	if err != nil {
		return nil, fmt.Errorf("fetching hooks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		Hooks []struct {
			PluginName string `json:"plugin_name"`
			Event      string `json:"event"`
			Table      string `json:"table"`
			Approved   bool   `json:"approved"`
		} `json:"hooks"`
	}

	if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr != nil {
		return nil, fmt.Errorf("decoding hooks: %w", decErr)
	}

	var items []hookItem
	for _, h := range result.Hooks {
		if h.PluginName == pluginName && h.Approved == wantApproved {
			items = append(items, hookItem{Plugin: h.PluginName, Event: h.Event, Table: h.Table})
		}
	}
	return items, nil
}

// approveRoutes sends a POST to /api/v1/admin/plugins/routes/approve.
func approveRoutes(cmd *cobra.Command, client *apiClient, pluginName string, routes []routeItem) error {
	payload := map[string]any{"routes": routes}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encoding request: %w", err)
	}

	resp, err := client.do("POST", "/api/v1/admin/plugins/routes/approve", strings.NewReader(string(body)))
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "request failed: %s\n", err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(cmd.ErrOrStderr(), "failed to approve routes for %q: %s\n", pluginName, strings.TrimSpace(string(respBody)))
		os.Exit(1)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Approved %d route(s) for plugin %q.\n", len(routes), pluginName)
	return nil
}

// revokeRoutes sends a POST to /api/v1/admin/plugins/routes/revoke.
func revokeRoutes(cmd *cobra.Command, client *apiClient, pluginName string, routes []routeItem) error {
	payload := map[string]any{"routes": routes}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encoding request: %w", err)
	}

	resp, err := client.do("POST", "/api/v1/admin/plugins/routes/revoke", strings.NewReader(string(body)))
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "request failed: %s\n", err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(cmd.ErrOrStderr(), "failed to revoke routes for %q: %s\n", pluginName, strings.TrimSpace(string(respBody)))
		os.Exit(1)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Revoked %d route(s) for plugin %q.\n", len(routes), pluginName)
	return nil
}

// approveHooks sends a POST to /api/v1/admin/plugins/hooks/approve.
func approveHooks(cmd *cobra.Command, client *apiClient, pluginName string, hooks []hookItem) error {
	payload := map[string]any{"hooks": hooks}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encoding request: %w", err)
	}

	resp, err := client.do("POST", "/api/v1/admin/plugins/hooks/approve", strings.NewReader(string(body)))
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "request failed: %s\n", err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(cmd.ErrOrStderr(), "failed to approve hooks for %q: %s\n", pluginName, strings.TrimSpace(string(respBody)))
		os.Exit(1)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Approved %d hook(s) for plugin %q.\n", len(hooks), pluginName)
	return nil
}

// revokeHooks sends a POST to /api/v1/admin/plugins/hooks/revoke.
func revokeHooks(cmd *cobra.Command, client *apiClient, pluginName string, hooks []hookItem) error {
	payload := map[string]any{"hooks": hooks}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encoding request: %w", err)
	}

	resp, err := client.do("POST", "/api/v1/admin/plugins/hooks/revoke", strings.NewReader(string(body)))
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "request failed: %s\n", err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(cmd.ErrOrStderr(), "failed to revoke hooks for %q: %s\n", pluginName, strings.TrimSpace(string(respBody)))
		os.Exit(1)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Revoked %d hook(s) for plugin %q.\n", len(hooks), pluginName)
	return nil
}

// approveAllRoutes fetches unapproved routes, confirms with the user, then approves.
func approveAllRoutes(cmd *cobra.Command, client *apiClient, pluginName string, skipPrompt bool) error {
	pending, err := fetchRoutesFiltered(client, pluginName, false)
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No pending routes for %q.\n", pluginName)
		return nil
	}

	// List pending items to stderr.
	for _, r := range pending {
		fmt.Fprintf(cmd.ErrOrStderr(), "  %s %s\n", r.Method, r.Path)
	}

	if !confirmAction(cmd, skipPrompt, fmt.Sprintf("Approve all %d routes?", len(pending))) {
		fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
		return nil
	}

	return approveRoutes(cmd, client, pluginName, pending)
}

// revokeAllRoutes fetches approved routes, confirms with the user, then revokes.
func revokeAllRoutes(cmd *cobra.Command, client *apiClient, pluginName string, skipPrompt bool) error {
	approved, err := fetchRoutesFiltered(client, pluginName, true)
	if err != nil {
		return err
	}

	if len(approved) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No approved routes for %q.\n", pluginName)
		return nil
	}

	// List items to stderr.
	for _, r := range approved {
		fmt.Fprintf(cmd.ErrOrStderr(), "  %s %s\n", r.Method, r.Path)
	}

	if !confirmAction(cmd, skipPrompt, fmt.Sprintf("Revoke all %d routes?", len(approved))) {
		fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
		return nil
	}

	return revokeRoutes(cmd, client, pluginName, approved)
}

// approveAllHooks fetches unapproved hooks, confirms with the user, then approves.
func approveAllHooks(cmd *cobra.Command, client *apiClient, pluginName string, skipPrompt bool) error {
	pending, err := fetchHooksFiltered(client, pluginName, false)
	if err != nil {
		return err
	}

	if len(pending) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No pending hooks for %q.\n", pluginName)
		return nil
	}

	// List pending items to stderr.
	for _, h := range pending {
		fmt.Fprintf(cmd.ErrOrStderr(), "  %s:%s\n", h.Event, h.Table)
	}

	if !confirmAction(cmd, skipPrompt, fmt.Sprintf("Approve all %d hooks?", len(pending))) {
		fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
		return nil
	}

	return approveHooks(cmd, client, pluginName, pending)
}

// revokeAllHooks fetches approved hooks, confirms with the user, then revokes.
func revokeAllHooks(cmd *cobra.Command, client *apiClient, pluginName string, skipPrompt bool) error {
	approved, err := fetchHooksFiltered(client, pluginName, true)
	if err != nil {
		return err
	}

	if len(approved) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No approved hooks for %q.\n", pluginName)
		return nil
	}

	// List items to stderr.
	for _, h := range approved {
		fmt.Fprintf(cmd.ErrOrStderr(), "  %s:%s\n", h.Event, h.Table)
	}

	if !confirmAction(cmd, skipPrompt, fmt.Sprintf("Revoke all %d hooks?", len(approved))) {
		fmt.Fprintln(cmd.ErrOrStderr(), "Cancelled.")
		return nil
	}

	return revokeHooks(cmd, client, pluginName, approved)
}

// confirmAction prompts the user for confirmation. Returns true if the user
// confirms, the --yes flag is set, or stdout is not a terminal (CI/CD).
func confirmAction(cmd *cobra.Command, skipPrompt bool, question string) bool {
	if skipPrompt {
		return true
	}
	// Skip prompt in non-interactive environments (piped output, CI/CD).
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return true
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "%s [y/N]: ", question)
	reader := bufio.NewReader(os.Stdin)
	line, readErr := reader.ReadString('\n')
	if readErr != nil {
		return false
	}
	answer := strings.TrimSpace(strings.ToLower(line))
	return answer == "y" || answer == "yes"
}

// generateInitLua produces the init.lua content for a newly scaffolded plugin.
func generateInitLua(name, version, description, author, license string) string {
	var b strings.Builder

	b.WriteString("-- ")
	b.WriteString(name)
	b.WriteString(": Modula plugin\n")
	b.WriteString("--\n")
	b.WriteString("-- Generated by: modula plugin init\n\n")

	// Manifest.
	b.WriteString("plugin_info = {\n")
	b.WriteString(fmt.Sprintf("    name        = %q,\n", name))
	b.WriteString(fmt.Sprintf("    version     = %q,\n", version))
	b.WriteString(fmt.Sprintf("    description = %q,\n", description))
	if author != "" {
		b.WriteString(fmt.Sprintf("    author      = %q,\n", author))
	}
	if license != "" {
		b.WriteString(fmt.Sprintf("    license     = %q,\n", license))
	}
	b.WriteString("}\n\n")

	// on_init stub.
	b.WriteString("function on_init()\n")
	b.WriteString("    -- Called when the plugin is loaded.\n")
	b.WriteString("    -- Use db.define_table() to create tables.\n")
	b.WriteString("    -- Use http.handle() to register HTTP routes.\n")
	b.WriteString("    -- Use hooks.on() to register content lifecycle hooks.\n")
	b.WriteString("    log.info(\"")
	b.WriteString(name)
	b.WriteString(" initialized\")\n")
	b.WriteString("end\n\n")

	// on_shutdown stub.
	b.WriteString("function on_shutdown()\n")
	b.WriteString("    -- Called when the plugin is being unloaded.\n")
	b.WriteString("    -- Use this for cleanup tasks.\n")
	b.WriteString("    log.info(\"")
	b.WriteString(name)
	b.WriteString(" shutting down\")\n")
	b.WriteString("end\n")

	return b.String()
}

func init() {
	// Persistent --token flag on the parent command for CI/CD use.
	// Available to all subcommands (online commands read it via pluginTokenFlag).
	pluginCmd.PersistentFlags().StringVar(&pluginTokenFlag, "token", "", "Admin API token (overrides token file, for CI/CD)")

	// Non-interactive flags for plugin init.
	pluginInitCmd.Flags().StringVar(&pluginInitVersion, "version", "", "Plugin version (default: 0.1.0)")
	pluginInitCmd.Flags().StringVar(&pluginInitDescription, "description", "", "Plugin description")
	pluginInitCmd.Flags().StringVar(&pluginInitAuthor, "author", "", "Plugin author")
	pluginInitCmd.Flags().StringVar(&pluginInitLicense, "license", "", "Plugin license")

	// Flags for plugin approve.
	pluginApproveCmd.Flags().StringVar(&pluginApproveRoute, "route", "", `Route to approve (e.g., "GET /tasks")`)
	pluginApproveCmd.Flags().StringVar(&pluginApproveHook, "hook", "", `Hook to approve (e.g., "after_insert:content_data")`)
	pluginApproveCmd.Flags().BoolVar(&pluginApproveAllRoutes, "all-routes", false, "Approve all unapproved routes")
	pluginApproveCmd.Flags().BoolVar(&pluginApproveAllHooks, "all-hooks", false, "Approve all unapproved hooks")
	pluginApproveCmd.Flags().BoolVar(&pluginApproveYes, "yes", false, "Skip confirmation prompt")

	// Flags for plugin revoke.
	pluginRevokeCmd.Flags().StringVar(&pluginRevokeRoute, "route", "", `Route to revoke (e.g., "GET /tasks")`)
	pluginRevokeCmd.Flags().StringVar(&pluginRevokeHook, "hook", "", `Hook to revoke (e.g., "after_insert:content_data")`)
	pluginRevokeCmd.Flags().BoolVar(&pluginRevokeAllRoutes, "all-routes", false, "Revoke all approved routes")
	pluginRevokeCmd.Flags().BoolVar(&pluginRevokeAllHooks, "all-hooks", false, "Revoke all approved hooks")
	pluginRevokeCmd.Flags().BoolVar(&pluginRevokeYes, "yes", false, "Skip confirmation prompt")

	// Register all subcommands.
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInitCmd)
	pluginCmd.AddCommand(pluginValidateCmd)
	pluginCmd.AddCommand(pluginInfoCmd)
	pluginCmd.AddCommand(pluginReloadCmd)
	pluginCmd.AddCommand(pluginEnableCmd)
	pluginCmd.AddCommand(pluginDisableCmd)
	pluginCmd.AddCommand(pluginApproveCmd)
	pluginCmd.AddCommand(pluginRevokeCmd)
}
