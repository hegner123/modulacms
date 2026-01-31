package cli

import (
	"fmt"
	"os"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/install"
	"github.com/hegner123/modulacms/internal/update"
	"github.com/hegner123/modulacms/internal/utility"
)

// ActionItem describes a single action available on the Actions page.
type ActionItem struct {
	Label       string
	Description string
	Destructive bool
}

// ActionsMenu returns the ordered list of action items.
// The index in this slice matches the cursor position on the Actions page.
func ActionsMenu() []ActionItem {
	return []ActionItem{
		{Label: "DB Init", Description: "Create database tables and bootstrap data"},
		{Label: "DB Wipe", Description: "Drop ALL tables and delete ALL data", Destructive: true},
		{Label: "DB Wipe & Redeploy", Description: "Drop all tables, recreate schema, and bootstrap data", Destructive: true},
		{Label: "DB Reset", Description: "Delete the database file (SQLite only)", Destructive: true},
		{Label: "DB Export", Description: "Dump database SQL to file"},
		{Label: "Generate Certs", Description: "Generate self-signed SSL certificates"},
		{Label: "Check for Updates", Description: "Check for and apply updates"},
		{Label: "Validate Config", Description: "Validate the configuration file"},
	}
}

// ActionsMenuLabels returns just the label strings for menu rendering.
func ActionsMenuLabels() []string {
	items := ActionsMenu()
	labels := make([]string, len(items))
	for i, item := range items {
		labels[i] = item.Label
	}
	return labels
}

// ActionResultMsg is returned by action commands with the result to display.
type ActionResultMsg struct {
	Title   string
	Message string
	IsError bool
}

// ActionConfirmMsg is sent when a destructive action needs confirmation.
type ActionConfirmMsg struct {
	ActionIndex int
}

// ActionConfirmedMsg is sent when the user confirms a destructive action.
type ActionConfirmedMsg struct {
	ActionIndex int
}

// --- Action execution commands ---

func RunActionCmd(cfg *config.Config, actionIndex int) tea.Cmd {
	switch actionIndex {
	case 0:
		return runDBInit(cfg)
	case 4:
		return runDBExport(cfg)
	case 5:
		return runGenerateCerts(cfg)
	case 6:
		return runCheckForUpdates()
	case 7:
		return runValidateConfig(cfg)
	default:
		return nil
	}
}

func RunDestructiveActionCmd(cfg *config.Config, actionIndex int) tea.Cmd {
	switch actionIndex {
	case 1:
		return runDBWipe(cfg)
	case 2:
		return runDBWipeRedeploy(cfg)
	case 3:
		return runDBReset(cfg)
	default:
		return nil
	}
}

func runDBInit(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		if err := install.CreateDbSimple("config.json", cfg); err != nil {
			return ActionResultMsg{
				Title:   "DB Init Failed",
				Message: fmt.Sprintf("Database initialization failed:\n%s", err),
				IsError: true,
			}
		}
		return ActionResultMsg{
			Title:   "DB Init Complete",
			Message: "Database tables created and bootstrap data loaded.",
		}
	}
}

func runDBWipe(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		driver := db.ConfigDB(*cfg)
		if err := driver.DropAllTables(); err != nil {
			return ActionResultMsg{
				Title:   "DB Wipe Failed",
				Message: fmt.Sprintf("Failed to drop tables:\n%s", err),
				IsError: true,
			}
		}
		return ActionResultMsg{
			Title:   "DB Wipe Complete",
			Message: "All tables dropped successfully.",
		}
	}
}

func runDBWipeRedeploy(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		driver := db.ConfigDB(*cfg)
		if err := driver.DropAllTables(); err != nil {
			return ActionResultMsg{
				Title:   "DB Wipe & Redeploy Failed",
				Message: fmt.Sprintf("Failed to drop tables:\n%s", err),
				IsError: true,
			}
		}
		if err := driver.CreateAllTables(); err != nil {
			return ActionResultMsg{
				Title:   "DB Wipe & Redeploy Failed",
				Message: fmt.Sprintf("Tables dropped but failed to recreate:\n%s", err),
				IsError: true,
			}
		}
		if err := driver.CreateBootstrapData(); err != nil {
			return ActionResultMsg{
				Title:   "DB Wipe & Redeploy Failed",
				Message: fmt.Sprintf("Tables recreated but bootstrap data failed:\n%s", err),
				IsError: true,
			}
		}
		if err := driver.ValidateBootstrapData(); err != nil {
			return ActionResultMsg{
				Title:   "DB Wipe & Redeploy Warning",
				Message: fmt.Sprintf("Completed but validation failed:\n%s", err),
				IsError: true,
			}
		}
		return ActionResultMsg{
			Title:   "DB Wipe & Redeploy Complete",
			Message: "Database wiped and redeployed successfully.",
		}
	}
}

func runDBReset(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		if err := os.Remove(cfg.Db_URL); err != nil {
			return ActionResultMsg{
				Title:   "DB Reset Failed",
				Message: fmt.Sprintf("Error deleting database file:\n%s", err),
				IsError: true,
			}
		}
		return ActionResultMsg{
			Title:   "DB Reset Complete",
			Message: fmt.Sprintf("Database file deleted: %s\nRestart the application to recreate.", cfg.Db_URL),
		}
	}
}

func runDBExport(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		driver := db.ConfigDB(*cfg)
		if err := driver.DumpSql(*cfg); err != nil {
			return ActionResultMsg{
				Title:   "DB Export Failed",
				Message: fmt.Sprintf("Database export failed:\n%s", err),
				IsError: true,
			}
		}
		return ActionResultMsg{
			Title:   "DB Export Complete",
			Message: "Database exported successfully.",
		}
	}
}

func runGenerateCerts(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		certDir := "./certs"
		domain := "localhost"

		if cfg.Cert_Dir != "" {
			certDir = cfg.Cert_Dir
		}
		if cfg.Client_Site != "" && cfg.Client_Site != "localhost" {
			domain = cfg.Client_Site
		}

		if err := utility.GenerateSelfSignedCert(certDir, domain); err != nil {
			return ActionResultMsg{
				Title:   "Generate Certs Failed",
				Message: fmt.Sprintf("Failed to generate certificates:\n%s", err),
				IsError: true,
			}
		}

		certMsg := fmt.Sprintf("Certificates generated in %s for domain %s.\n\n  Certificate: %s/localhost.crt\n  Private Key: %s/localhost.key",
			certDir, domain, certDir, certDir)

		return ActionResultMsg{
			Title:   "Generate Certs Complete",
			Message: certMsg,
		}
	}
}

func runCheckForUpdates() tea.Cmd {
	return func() tea.Msg {
		currentVersion := utility.GetCurrentVersion()
		release, available, err := update.CheckForUpdates(currentVersion, "stable")
		if err != nil {
			return ActionResultMsg{
				Title:   "Update Check Failed",
				Message: fmt.Sprintf("Update check failed:\n%s", err),
				IsError: true,
			}
		}

		if !available {
			return ActionResultMsg{
				Title:   "Up to Date",
				Message: fmt.Sprintf("Already running latest version (%s).", currentVersion),
			}
		}

		downloadURL, err := update.GetDownloadURL(release, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			return ActionResultMsg{
				Title:   "Update Available",
				Message: fmt.Sprintf("Update %s available but no compatible binary found:\n%s", release.TagName, err),
				IsError: true,
			}
		}

		tempPath, err := update.DownloadUpdate(downloadURL)
		if err != nil {
			return ActionResultMsg{
				Title:   "Update Download Failed",
				Message: fmt.Sprintf("Failed to download update:\n%s", err),
				IsError: true,
			}
		}

		if err := update.ApplyUpdate(tempPath); err != nil {
			return ActionResultMsg{
				Title:   "Update Apply Failed",
				Message: fmt.Sprintf("Failed to apply update:\n%s", err),
				IsError: true,
			}
		}

		return ActionResultMsg{
			Title:   "Update Complete",
			Message: fmt.Sprintf("Updated to %s. Please restart ModulaCMS.", release.TagName),
		}
	}
}

func runValidateConfig(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		var errs []string

		if cfg.Db_Driver == "" {
			errs = append(errs, "db_driver is required")
		}
		if cfg.Db_URL == "" {
			errs = append(errs, "db_url is required")
		}
		if cfg.Port == "" {
			errs = append(errs, "port is required")
		}
		if cfg.SSH_Port == "" {
			errs = append(errs, "ssh_port is required")
		}

		if len(errs) > 0 {
			msg := fmt.Sprintf("Configuration has %d error(s):\n", len(errs))
			for _, e := range errs {
				msg += fmt.Sprintf("  - %s\n", e)
			}
			return ActionResultMsg{
				Title:   "Validation Failed",
				Message: msg,
				IsError: true,
			}
		}

		return ActionResultMsg{
			Title:   "Validation Passed",
			Message: "Configuration is valid.",
		}
	}
}
