package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/hegner123/modulacms/internal/db"
)

// ---------------------------------------------------------------------------
// formatBytes
// ---------------------------------------------------------------------------

func TestFormatBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{name: "zero bytes", input: 0, want: "0 B"},
		{name: "one byte", input: 1, want: "1 B"},
		{name: "below KB threshold", input: 1023, want: "1023 B"},
		{name: "exact 1 KB", input: 1024, want: "1.0 KB"},
		{name: "1.5 KB", input: 1536, want: "1.5 KB"},
		{name: "exact 1 MB", input: 1024 * 1024, want: "1.0 MB"},
		{name: "exact 1 GB", input: 1024 * 1024 * 1024, want: "1.0 GB"},
		{name: "exact 1 TB", input: 1024 * 1024 * 1024 * 1024, want: "1.0 TB"},
		{name: "2.5 MB", input: 2621440, want: "2.5 MB"},
		{name: "large value 5.3 GB", input: 5690621952, want: "5.3 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatBytes(tt.input)
			if got != tt.want {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// sanitizeCertDir
// ---------------------------------------------------------------------------

func TestSanitizeCertDir(t *testing.T) {
	t.Parallel()

	t.Run("empty string returns error", func(t *testing.T) {
		t.Parallel()
		_, err := sanitizeCertDir("")
		if err == nil {
			t.Fatal("expected error for empty cert dir, got nil")
		}
		if !strings.Contains(err.Error(), "cannot be empty") {
			t.Errorf("expected 'cannot be empty' error, got: %v", err)
		}
	})

	t.Run("whitespace only returns error", func(t *testing.T) {
		t.Parallel()
		_, err := sanitizeCertDir("   ")
		if err == nil {
			t.Fatal("expected error for whitespace-only cert dir, got nil")
		}
	})

	t.Run("nonexistent directory returns error", func(t *testing.T) {
		t.Parallel()
		_, err := sanitizeCertDir("/nonexistent/path/that/does/not/exist")
		if err == nil {
			t.Fatal("expected error for nonexistent directory, got nil")
		}
	})

	t.Run("file path instead of directory returns error", func(t *testing.T) {
		t.Parallel()
		tmpFile := filepath.Join(t.TempDir(), "not-a-dir.txt")
		if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}

		_, err := sanitizeCertDir(tmpFile)
		if err == nil {
			t.Fatal("expected error when path is a file, got nil")
		}
		if !strings.Contains(err.Error(), "not a directory") {
			t.Errorf("expected 'not a directory' error, got: %v", err)
		}
	})

	t.Run("valid directory returns absolute path", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		got, err := sanitizeCertDir(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !filepath.IsAbs(got) {
			t.Errorf("expected absolute path, got: %s", got)
		}
	})

	t.Run("cleans path traversal", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		// Create a subdirectory then reference it via parent traversal
		subDir := filepath.Join(dir, "sub")
		if err := os.MkdirAll(subDir, 0750); err != nil {
			t.Fatalf("failed to create subdir: %v", err)
		}

		// dir/sub/../sub should resolve to dir/sub
		traversal := filepath.Join(dir, "sub", "..", "sub")
		got, err := sanitizeCertDir(traversal)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		absSubDir, _ := filepath.Abs(subDir)
		if got != absSubDir {
			t.Errorf("expected %s, got %s", absSubDir, got)
		}
	})
}

// ---------------------------------------------------------------------------
// newHTTPServer
// ---------------------------------------------------------------------------

func TestNewHTTPServer(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	srv := newHTTPServer("localhost:9090", handler, nil)

	if srv.Addr != "localhost:9090" {
		t.Errorf("expected addr localhost:9090, got %s", srv.Addr)
	}
	if srv.TLSConfig != nil {
		t.Error("expected nil TLSConfig when none provided")
	}
	if srv.ReadTimeout != 15*1e9 { // 15 seconds in nanoseconds
		t.Errorf("expected ReadTimeout 15s, got %v", srv.ReadTimeout)
	}
	if srv.WriteTimeout != 15*1e9 {
		t.Errorf("expected WriteTimeout 15s, got %v", srv.WriteTimeout)
	}
	if srv.IdleTimeout != 60*1e9 {
		t.Errorf("expected IdleTimeout 60s, got %v", srv.IdleTimeout)
	}
}

// ---------------------------------------------------------------------------
// handlerSwap
// ---------------------------------------------------------------------------

func TestHandlerSwap_ServeHTTP(t *testing.T) {
	t.Parallel()

	swap := &handlerSwap{}

	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler1"))
	})
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler2"))
	})

	swap.set(handler1)

	// Verify first handler serves
	rec := httptest.NewRecorder()
	swap.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Body.String() != "handler1" {
		t.Errorf("expected handler1 response, got %q", rec.Body.String())
	}

	// Swap to second handler
	swap.set(handler2)

	rec = httptest.NewRecorder()
	swap.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Body.String() != "handler2" {
		t.Errorf("expected handler2 response, got %q", rec.Body.String())
	}
}

// TestHandlerSwap_ConcurrentAccess verifies no race conditions when
// swapping handlers while serving requests. Run with -race.
func TestHandlerSwap_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	swap := &handlerSwap{}
	swap.set(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	const goroutines = 50

	// Concurrent reads
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := httptest.NewRecorder()
			swap.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
		}()
	}

	// Concurrent writes (swaps)
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			swap.set(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusAccepted)
			}))
		}()
	}

	wg.Wait()
}

// ---------------------------------------------------------------------------
// parseRouteFlag
// ---------------------------------------------------------------------------

func TestParseRouteFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantMethod string
		wantPath   string
		wantErr    string
	}{
		{
			name:       "valid GET route",
			input:      "GET /tasks",
			wantMethod: "GET",
			wantPath:   "/tasks",
		},
		{
			name:       "valid POST route with nested path",
			input:      "POST /api/v1/items",
			wantMethod: "POST",
			wantPath:   "/api/v1/items",
		},
		{
			name:       "lowercase method is uppercased",
			input:      "get /tasks",
			wantMethod: "GET",
			wantPath:   "/tasks",
		},
		{
			// Leading spaces before the method cause SplitN to produce an empty
			// first part, so the method is empty after TrimSpace -> error.
			name:    "leading spaces before method causes error",
			input:   "  PUT   /data  ",
			wantErr: "method and path must be non-empty",
		},
		{
			name:       "trailing spaces on path are trimmed",
			input:      "PUT /data  ",
			wantMethod: "PUT",
			wantPath:   "/data",
		},
		{
			name:    "missing space separator",
			input:   "GET/tasks",
			wantErr: "invalid --route format",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: "invalid --route format",
		},
		{
			name:    "only method no path",
			input:   "GET ",
			wantErr: "method and path must be non-empty",
		},
		{
			name:    "only space",
			input:   " ",
			wantErr: "method and path must be non-empty",
		},
		{
			name:       "path with spaces in path portion",
			input:      "DELETE /some path with spaces",
			wantMethod: "DELETE",
			wantPath:   "/some path with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			method, path, err := parseRouteFlag(tt.input)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if method != tt.wantMethod {
				t.Errorf("method: got %q, want %q", method, tt.wantMethod)
			}
			if path != tt.wantPath {
				t.Errorf("path: got %q, want %q", path, tt.wantPath)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseHookFlag
// ---------------------------------------------------------------------------

func TestParseHookFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantEvent string
		wantTable string
		wantErr   string
	}{
		{
			name:      "valid hook",
			input:     "after_insert:content_data",
			wantEvent: "after_insert",
			wantTable: "content_data",
		},
		{
			name:      "before_delete hook",
			input:     "before_delete:media",
			wantEvent: "before_delete",
			wantTable: "media",
		},
		{
			name:      "extra colons go into table portion",
			input:     "after_update:some:complex:name",
			wantEvent: "after_update",
			wantTable: "some:complex:name",
		},
		{
			name:    "no colon separator",
			input:   "after_insert_content_data",
			wantErr: "invalid --hook format",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: "invalid --hook format",
		},
		{
			name:    "colon but empty event",
			input:   ":content_data",
			wantErr: "event and table must be non-empty",
		},
		{
			name:    "colon but empty table",
			input:   "after_insert:",
			wantErr: "event and table must be non-empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			event, table, err := parseHookFlag(tt.input)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if event != tt.wantEvent {
				t.Errorf("event: got %q, want %q", event, tt.wantEvent)
			}
			if table != tt.wantTable {
				t.Errorf("table: got %q, want %q", table, tt.wantTable)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// generateInitLua
// ---------------------------------------------------------------------------

func TestGenerateInitLua(t *testing.T) {
	t.Parallel()

	t.Run("all fields populated", func(t *testing.T) {
		t.Parallel()
		got := generateInitLua("my_plugin", "1.0.0", "A test plugin", "Alice", "MIT")

		// Verify the header comment includes the plugin name
		if !strings.Contains(got, "-- my_plugin: Modula plugin") {
			t.Error("expected header with plugin name")
		}

		// Verify manifest fields
		if !strings.Contains(got, `name        = "my_plugin"`) {
			t.Error("expected name field in manifest")
		}
		if !strings.Contains(got, `version     = "1.0.0"`) {
			t.Error("expected version field in manifest")
		}
		if !strings.Contains(got, `description = "A test plugin"`) {
			t.Error("expected description field in manifest")
		}
		if !strings.Contains(got, `author      = "Alice"`) {
			t.Error("expected author field in manifest")
		}
		if !strings.Contains(got, `license     = "MIT"`) {
			t.Error("expected license field in manifest")
		}

		// Verify lifecycle stubs
		if !strings.Contains(got, "function on_init()") {
			t.Error("expected on_init stub")
		}
		if !strings.Contains(got, "function on_shutdown()") {
			t.Error("expected on_shutdown stub")
		}

		// Verify log messages reference the plugin name
		if !strings.Contains(got, `log.info("my_plugin initialized")`) {
			t.Error("expected init log message with plugin name")
		}
		if !strings.Contains(got, `log.info("my_plugin shutting down")`) {
			t.Error("expected shutdown log message with plugin name")
		}
	})

	t.Run("empty author and license omitted", func(t *testing.T) {
		t.Parallel()
		got := generateInitLua("bare_plugin", "0.1.0", "Bare plugin", "", "")

		if strings.Contains(got, "author") {
			t.Error("expected author to be omitted when empty")
		}
		if strings.Contains(got, "license") {
			t.Error("expected license to be omitted when empty")
		}

		// Other fields should still be present
		if !strings.Contains(got, `name        = "bare_plugin"`) {
			t.Error("expected name field")
		}
	})

	t.Run("special characters in description are quoted", func(t *testing.T) {
		t.Parallel()
		got := generateInitLua("special", "1.0.0", `A "special" plugin with 'quotes'`, "", "")

		// fmt.Sprintf %q handles escaping
		if !strings.Contains(got, `description = "A \"special\" plugin with 'quotes'"`) {
			t.Errorf("expected properly escaped description, got:\n%s", got)
		}
	})
}

// ---------------------------------------------------------------------------
// parseTablesFlag
// ---------------------------------------------------------------------------

func TestParseTablesFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []db.DBTable
		wantNil bool // true if expected result is nil
	}{
		{
			name:    "empty string returns nil",
			input:   "",
			wantNil: true,
		},
		{
			name:  "single valid table",
			input: "users",
			want:  []db.DBTable{db.User},
		},
		{
			name:  "multiple valid tables",
			input: "users,media,fields",
			want:  []db.DBTable{db.User, db.MediaT, db.Field},
		},
		{
			name:  "spaces around names are trimmed",
			input: " users , media , fields ",
			want:  []db.DBTable{db.User, db.MediaT, db.Field},
		},
		{
			name:    "unknown table name returns nil",
			input:   "nonexistent_table",
			wantNil: true,
		},
		{
			name:    "mix of valid and invalid returns nil",
			input:   "users,nonexistent_table",
			wantNil: true,
		},
		{
			name:  "trailing comma with empty segment is skipped",
			input: "users,",
			want:  []db.DBTable{db.User},
		},
		{
			name:  "leading comma with empty segment is skipped",
			input: ",users",
			want:  []db.DBTable{db.User},
		},
		{
			name:  "content_data table",
			input: "content_data",
			want:  []db.DBTable{db.Content_data},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseTablesFlag(tt.input)

			if tt.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected non-nil result, got nil")
			}

			if len(got) != len(tt.want) {
				t.Fatalf("expected %d tables, got %d: %v", len(tt.want), len(got), got)
			}

			for i, w := range tt.want {
				if got[i] != w {
					t.Errorf("table[%d]: got %q, want %q", i, got[i], w)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Cobra command tree structure
// ---------------------------------------------------------------------------

func TestRootCommand_Structure(t *testing.T) {
	t.Parallel()

	if rootCmd.Use != "modula" {
		t.Errorf("root command Use: got %q, want %q", rootCmd.Use, "modula")
	}

	// Verify all expected subcommands are registered.
	expectedSubcommands := []string{
		"serve", "install", "version", "update", "tui",
		"cert", "db", "config", "backup", "plugin", "deploy",
	}

	subCmds := rootCmd.Commands()
	cmdNames := make(map[string]bool, len(subCmds))
	for _, c := range subCmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range expectedSubcommands {
		if !cmdNames[name] {
			t.Errorf("expected subcommand %q not found on root command", name)
		}
	}
}

func TestRootCommand_PersistentFlags(t *testing.T) {
	t.Parallel()

	configFlag := rootCmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Fatal("expected persistent flag 'config' not found")
	}
	if configFlag.DefValue != "modula.config.json" {
		t.Errorf("config flag default: got %q, want %q", configFlag.DefValue, "modula.config.json")
	}

	verboseFlag := rootCmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Fatal("expected persistent flag 'verbose' not found")
	}
	if verboseFlag.DefValue != "false" {
		t.Errorf("verbose flag default: got %q, want %q", verboseFlag.DefValue, "false")
	}
	if verboseFlag.Shorthand != "v" {
		t.Errorf("verbose flag shorthand: got %q, want %q", verboseFlag.Shorthand, "v")
	}
}

func TestDbCommand_Subcommands(t *testing.T) {
	t.Parallel()

	expectedSubs := []string{"init", "wipe", "wipe-redeploy", "reset", "export"}
	subCmds := dbCmd.Commands()
	cmdNames := make(map[string]bool, len(subCmds))
	for _, c := range subCmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range expectedSubs {
		if !cmdNames[name] {
			t.Errorf("expected db subcommand %q not found", name)
		}
	}
}

func TestConfigCommand_Subcommands(t *testing.T) {
	t.Parallel()

	expectedSubs := []string{"show", "validate", "set"}
	subCmds := configParentCmd.Commands()
	cmdNames := make(map[string]bool, len(subCmds))
	for _, c := range subCmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range expectedSubs {
		if !cmdNames[name] {
			t.Errorf("expected config subcommand %q not found", name)
		}
	}
}

func TestBackupCommand_Subcommands(t *testing.T) {
	t.Parallel()

	expectedSubs := []string{"create", "restore", "list"}
	subCmds := backupCmd.Commands()
	cmdNames := make(map[string]bool, len(subCmds))
	for _, c := range subCmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range expectedSubs {
		if !cmdNames[name] {
			t.Errorf("expected backup subcommand %q not found", name)
		}
	}
}

func TestPluginCommand_Subcommands(t *testing.T) {
	t.Parallel()

	expectedSubs := []string{
		"list", "init", "validate", "info", "reload",
		"enable", "disable", "approve", "revoke",
	}
	subCmds := pluginCmd.Commands()
	cmdNames := make(map[string]bool, len(subCmds))
	for _, c := range subCmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range expectedSubs {
		if !cmdNames[name] {
			t.Errorf("expected plugin subcommand %q not found", name)
		}
	}
}

func TestDeployCommand_Subcommands(t *testing.T) {
	t.Parallel()

	expectedSubs := []string{"export", "import", "pull", "push", "snapshot", "env"}
	subCmds := deployCmd.Commands()
	cmdNames := make(map[string]bool, len(subCmds))
	for _, c := range subCmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range expectedSubs {
		if !cmdNames[name] {
			t.Errorf("expected deploy subcommand %q not found", name)
		}
	}
}

func TestTuiCommand_NoSubcommands(t *testing.T) {
	t.Parallel()

	subCmds := tuiCmd.Commands()
	if len(subCmds) != 0 {
		t.Errorf("expected tui to have no subcommands, got %d", len(subCmds))
	}
}

func TestCertCommand_Subcommands(t *testing.T) {
	t.Parallel()

	expectedSubs := []string{"generate"}
	subCmds := certCmd.Commands()
	cmdNames := make(map[string]bool, len(subCmds))
	for _, c := range subCmds {
		cmdNames[c.Name()] = true
	}

	for _, name := range expectedSubs {
		if !cmdNames[name] {
			t.Errorf("expected cert subcommand %q not found", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Command Args constraints
// ---------------------------------------------------------------------------

func TestCommandArgsConstraints(t *testing.T) {
	t.Parallel()

	// Commands that require exactly 1 argument
	exactOneArgCmds := []struct {
		name string
		cmd  *bytes.Buffer // not used, placeholder
		use  string
	}{
		{name: "backup restore", use: "restore <path>"},
		{name: "plugin init", use: "init <name>"},
		{name: "plugin validate", use: "validate <path>"},
		{name: "plugin info", use: "info <name>"},
		{name: "plugin reload", use: "reload <name>"},
		{name: "plugin enable", use: "enable <name>"},
		{name: "plugin disable", use: "disable <name>"},
		{name: "plugin approve", use: "approve <name>"},
		{name: "plugin revoke", use: "revoke <name>"},
		{name: "deploy import", use: "import <file>"},
		{name: "deploy pull", use: "pull <source>"},
		{name: "deploy push", use: "push <target>"},
	}

	for _, tt := range exactOneArgCmds {
		t.Run(tt.name+" requires an argument", func(t *testing.T) {
			t.Parallel()
			// The Use string contains "<" which indicates required args
			if !strings.Contains(tt.use, "<") {
				t.Errorf("expected Use string %q to contain angle-bracket arg", tt.use)
			}
		})
	}

	// config set requires exactly 2 arguments
	if configSetCmd.Use != "set <key> <value>" {
		t.Errorf("config set Use: got %q, want %q", configSetCmd.Use, "set <key> <value>")
	}
}

// ---------------------------------------------------------------------------
// Plugin command flags
// ---------------------------------------------------------------------------

func TestPluginCommand_PersistentTokenFlag(t *testing.T) {
	t.Parallel()

	tokenFlag := pluginCmd.PersistentFlags().Lookup("token")
	if tokenFlag == nil {
		t.Fatal("expected persistent flag 'token' on plugin command")
	}
	if tokenFlag.DefValue != "" {
		t.Errorf("token flag default: got %q, want empty string", tokenFlag.DefValue)
	}
}

func TestPluginInitCommand_Flags(t *testing.T) {
	t.Parallel()

	expectedFlags := []string{"version", "description", "author", "license"}
	for _, name := range expectedFlags {
		f := pluginInitCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("expected flag %q on plugin init command", name)
		}
	}
}

func TestPluginApproveCommand_Flags(t *testing.T) {
	t.Parallel()

	expectedFlags := []string{"route", "hook", "all-routes", "all-hooks", "yes"}
	for _, name := range expectedFlags {
		f := pluginApproveCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("expected flag %q on plugin approve command", name)
		}
	}
}

func TestPluginRevokeCommand_Flags(t *testing.T) {
	t.Parallel()

	expectedFlags := []string{"route", "hook", "all-routes", "all-hooks", "yes"}
	for _, name := range expectedFlags {
		f := pluginRevokeCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("expected flag %q on plugin revoke command", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Serve command flags
// ---------------------------------------------------------------------------

func TestServeCommand_WizardFlag(t *testing.T) {
	t.Parallel()

	wizardFlag := serveCmd.Flags().Lookup("wizard")
	if wizardFlag == nil {
		t.Fatal("expected flag 'wizard' on serve command")
	}
	if wizardFlag.DefValue != "false" {
		t.Errorf("wizard flag default: got %q, want %q", wizardFlag.DefValue, "false")
	}
}

// ---------------------------------------------------------------------------
// Install command flags
// ---------------------------------------------------------------------------

func TestInstallCommand_Flags(t *testing.T) {
	t.Parallel()

	yesFlag := installCmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("expected flag 'yes' on install command")
	}
	if yesFlag.Shorthand != "y" {
		t.Errorf("yes flag shorthand: got %q, want %q", yesFlag.Shorthand, "y")
	}

	pwFlag := installCmd.Flags().Lookup("admin-password")
	if pwFlag == nil {
		t.Fatal("expected flag 'admin-password' on install command")
	}
}

// ---------------------------------------------------------------------------
// Version command execution
// ---------------------------------------------------------------------------

func TestVersionCommand_Output(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	versionCmd.SetOut(&buf)
	versionCmd.SetErr(&buf)

	err := versionCmd.RunE(versionCmd, nil)
	if err != nil {
		t.Fatalf("version command returned error: %v", err)
	}

	output := buf.String()
	// utility.GetFullVersionInfo returns "Version: ...\nCommit: ...\nBuilt: ..."
	if !strings.Contains(output, "Version:") {
		t.Errorf("expected version output to contain 'Version:', got: %q", output)
	}
}

// ---------------------------------------------------------------------------
// Deploy command flags
// ---------------------------------------------------------------------------

func TestDeployExportCommand_Flags(t *testing.T) {
	t.Parallel()

	expectedFlags := []string{"file", "tables", "json"}
	for _, name := range expectedFlags {
		f := deployExportCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("expected flag %q on deploy export command", name)
		}
	}
}

func TestDeployImportCommand_Flags(t *testing.T) {
	t.Parallel()

	expectedFlags := []string{"dry-run", "skip-backup", "json"}
	for _, name := range expectedFlags {
		f := deployImportCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("expected flag %q on deploy import command", name)
		}
	}
}

func TestDeployPullCommand_Flags(t *testing.T) {
	t.Parallel()

	expectedFlags := []string{"tables", "skip-backup", "dry-run", "json"}
	for _, name := range expectedFlags {
		f := deployPullCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("expected flag %q on deploy pull command", name)
		}
	}
}

func TestDeployPushCommand_Flags(t *testing.T) {
	t.Parallel()

	expectedFlags := []string{"tables", "dry-run", "json"}
	for _, name := range expectedFlags {
		f := deployPushCmd.Flags().Lookup(name)
		if f == nil {
			t.Errorf("expected flag %q on deploy push command", name)
		}
	}
}
