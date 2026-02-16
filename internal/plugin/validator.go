package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
)

// ValidationSeverity indicates whether a validation result is a blocking error
// or an informational warning.
type ValidationSeverity int

const (
	// SeverityError indicates a problem that prevents the plugin from loading.
	SeverityError ValidationSeverity = iota
	// SeverityWarning indicates a non-blocking issue that should be addressed.
	SeverityWarning
)

// String returns the human-readable name of the severity level.
func (s ValidationSeverity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// ValidationResult represents a single validation finding for a plugin.
type ValidationResult struct {
	Field    string             // the manifest field or check area (e.g., "name", "syntax", "on_init")
	Message  string             // human-readable description of the issue
	Severity ValidationSeverity // error or warning
}

// ValidatePlugin performs offline validation of a plugin directory without
// loading it into the runtime. It checks:
//  1. Directory existence
//  2. init.lua file existence
//  3. Lua syntax validity (parse+compile without executing)
//  4. Manifest extraction and field validation via ExtractManifest/ValidateManifest
//
// Returns the extracted PluginInfo (zero value if extraction failed), a slice
// of validation results, and an error only for unexpected I/O failures (not
// validation failures, which are captured in the results slice).
func ValidatePlugin(dir string) (PluginInfo, []ValidationResult, error) {
	var results []ValidationResult

	// Check that the directory exists.
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return PluginInfo{}, nil, fmt.Errorf("checking plugin directory %q: %w", dir, err)
	}
	if !dirInfo.IsDir() {
		return PluginInfo{}, nil, fmt.Errorf("path %q is not a directory", dir)
	}

	// Check that init.lua exists.
	initPath := filepath.Join(dir, "init.lua")
	if _, err := os.Stat(initPath); err != nil {
		results = append(results, ValidationResult{
			Field:    "init.lua",
			Message:  "missing required init.lua file",
			Severity: SeverityError,
		})
		return PluginInfo{}, results, nil
	}

	// Pre-flight syntax check: parse+compile without executing.
	// Uses a minimal lua.LState and L.LoadFile which compiles but does not run
	// the code. This catches syntax errors cheaply before the full
	// ExtractManifest call (which executes the file with a 2-second timeout).
	if syntaxErr := checkSyntax(initPath); syntaxErr != nil {
		results = append(results, ValidationResult{
			Field:    "syntax",
			Message:  fmt.Sprintf("Lua syntax error: %s", syntaxErr.Error()),
			Severity: SeverityError,
		})
		return PluginInfo{}, results, nil
	}

	// Extract and validate manifest via the same path used by the Manager.
	// ExtractManifest internally calls ValidateManifest, so both extraction
	// and validation errors are captured here.
	info, extractErr := ExtractManifest(initPath)
	if extractErr != nil {
		results = append(results, ValidationResult{
			Field:    "manifest",
			Message:  extractErr.Error(),
			Severity: SeverityError,
		})
		return PluginInfo{}, results, nil
	}

	// Advisory warnings for missing optional fields.
	if info.Author == "" {
		results = append(results, ValidationResult{
			Field:    "author",
			Message:  "author field is empty (recommended for discoverability)",
			Severity: SeverityWarning,
		})
	}
	if info.License == "" {
		results = append(results, ValidationResult{
			Field:    "license",
			Message:  "license field is empty (recommended for distribution)",
			Severity: SeverityWarning,
		})
	}

	// Advisory: check on_init presence (already set in info.HasOnInit by ExtractManifest).
	if !info.HasOnInit {
		results = append(results, ValidationResult{
			Field:    "on_init",
			Message:  "no on_init() function defined (plugin will load but do nothing at startup)",
			Severity: SeverityWarning,
		})
	}

	return *info, results, nil
}

// checkSyntax uses a minimal Lua VM to parse and compile init.lua without
// executing it. Returns nil if the file compiles successfully, or the
// compilation error if syntax is invalid.
func checkSyntax(initPath string) error {
	L := lua.NewState(lua.Options{
		SkipOpenLibs: true,
	})
	defer L.Close()

	// L.LoadFile compiles the file into a Lua function without executing it.
	// If the file has syntax errors, LoadFile returns an error.
	_, err := L.LoadFile(initPath)
	return err
}
