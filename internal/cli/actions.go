package cli

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/backup"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/install"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/update"
	"github.com/hegner123/modulacms/internal/utility"
)

// ActionParams groups the context needed by action commands.
type ActionParams struct {
	Config         *config.Config
	UserID         types.UserID
	SSHFingerprint string
	SSHKeyType     string
	SSHPublicKey   string
}

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
		{Label: "Generate API Token", Description: "Create a new API token for the current user"},
		{Label: "Register SSH Key", Description: "Create a new user and register the current SSH key"},
		{Label: "Create Backup", Description: "Create a backup of database and configured paths"},
		{Label: "Restore Backup", Description: "Restore from a backup archive", Destructive: true},
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
	Title              string
	Message            string
	IsError            bool
	Width              int  // optional dialog width override (0 = default)
	ReloadPermissions  bool // signal serve to reload permission cache and start HTTP
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

// RunActionCmd creates a command to run a non-destructive action by index.
func RunActionCmd(p ActionParams, actionIndex int) tea.Cmd {
	switch actionIndex {
	case 0:
		return runDBInit(p.Config)
	case 4:
		return runDBExport(p.Config)
	case 5:
		return runGenerateCerts(p.Config)
	case 6:
		return runCheckForUpdates()
	case 7:
		return runValidateConfig(p.Config)
	case 8:
		return runGenerateAPIToken(p.Config, p.UserID)
	case 9:
		return runRegisterSSHKey(p)
	case 10:
		return runCreateBackup(p.Config)
	default:
		return nil
	}
}

// RunDestructiveActionCmd creates a command to run a destructive action by index.
func RunDestructiveActionCmd(p ActionParams, actionIndex int) tea.Cmd {
	switch actionIndex {
	case 1:
		return runDBWipe(p.Config)
	case 2:
		return runDBWipeRedeploy(p.Config)
	case 3:
		return runDBReset(p.Config)
	case 11:
		return func() tea.Msg {
			return OpenFilePickerForRestoreMsg{}
		}
	default:
		return nil
	}
}

// runDBInit creates a command to initialize the database with schema and bootstrap data.
func runDBInit(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		randomPassword, err := utility.MakeRandomString()
		if err != nil {
			return ActionResultMsg{
				Title:   "DB Init Failed",
				Message: fmt.Sprintf("Failed to generate admin password:\n%s", err),
				IsError: true,
			}
		}
		adminHash, err := auth.HashPassword(randomPassword)
		if err != nil {
			return ActionResultMsg{
				Title:   "DB Init Failed",
				Message: fmt.Sprintf("Failed to hash admin password:\n%s", err),
				IsError: true,
			}
		}

		if err := install.CreateDbSimple("config.json", cfg, adminHash); err != nil {
			return ActionResultMsg{
				Title:   "DB Init Failed",
				Message: fmt.Sprintf("Database initialization failed:\n%s", err),
				IsError: true,
			}
		}
		return ActionResultMsg{
			Title:             "DB Init Complete",
			Message:           fmt.Sprintf("Database tables created and bootstrap data loaded.\n\nSystem admin: system@modula.local\nTemporary password: %s\n\nPlease change this password immediately.", randomPassword),
			ReloadPermissions: true,
		}
	}
}

// runDBWipe creates a command to drop all database tables.
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

// runDBWipeRedeploy creates a command to drop, recreate, and bootstrap the database.
func runDBWipeRedeploy(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		// Generate a random password for the bootstrap admin user
		randomPassword, err := utility.MakeRandomString()
		if err != nil {
			return ActionResultMsg{
				Title:   "DB Wipe & Redeploy Failed",
				Message: fmt.Sprintf("Failed to generate admin password:\n%s", err),
				IsError: true,
			}
		}
		adminHash, err := auth.HashPassword(randomPassword)
		if err != nil {
			return ActionResultMsg{
				Title:   "DB Wipe & Redeploy Failed",
				Message: fmt.Sprintf("Failed to hash admin password:\n%s", err),
				IsError: true,
			}
		}

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
		if err := driver.CreateBootstrapData(adminHash); err != nil {
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
			Title:             "DB Wipe & Redeploy Complete",
			Message:           fmt.Sprintf("Database wiped and redeployed successfully.\n\nSystem admin: system@modula.local\nTemporary password: %s\n\nPlease change this password immediately.", randomPassword),
			ReloadPermissions: true,
		}
	}
}

// runDBReset creates a command to delete the SQLite database file.
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

// runDBExport creates a command to dump database schema and data to a SQL file.
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

// runGenerateCerts creates a command to generate self-signed SSL certificates.
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

// runCheckForUpdates creates a command to check for and apply application updates.
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
			Message: fmt.Sprintf("Updated to %s. Please restart Modula.", release.TagName),
		}
	}
}

// runValidateConfig creates a command to validate the configuration file.
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

// runGenerateAPIToken creates a command to generate a new API token for the current user.
func runGenerateAPIToken(cfg *config.Config, userID types.UserID) tea.Cmd {
	return func() tea.Msg {
		driver := db.ConfigDB(*cfg)

		// Use authenticated user if available, otherwise fall back to first admin
		ownerID := userID
		if ownerID.IsZero() {
			roles, err := driver.ListRoles()
			if err != nil {
				return ActionResultMsg{
					Title:   "Token Generation Failed",
					Message: fmt.Sprintf("Failed to list roles:\n%s", err),
					IsError: true,
				}
			}
			var adminRoleID string
			for _, r := range *roles {
				if r.Label == "admin" {
					adminRoleID = string(r.RoleID)
					break
				}
			}
			if adminRoleID == "" {
				return ActionResultMsg{
					Title:   "Token Generation Failed",
					Message: "No admin role found. Run DB Init first.",
					IsError: true,
				}
			}

			users, err := driver.ListUsers()
			if err != nil {
				return ActionResultMsg{
					Title:   "Token Generation Failed",
					Message: fmt.Sprintf("Failed to list users:\n%s", err),
					IsError: true,
				}
			}
			for _, u := range *users {
				if u.Role == adminRoleID {
					ownerID = u.UserID
					break
				}
			}
			if ownerID.IsZero() {
				return ActionResultMsg{
					Title:   "Token Generation Failed",
					Message: "No admin user found. Create one first.",
					IsError: true,
				}
			}
		}

		token, err := utility.MakeRandomString()
		if err != nil {
			return ActionResultMsg{
				Title:   "Token Generation Failed",
				Message: fmt.Sprintf("Failed to generate token:\n%s", err),
				IsError: true,
			}
		}

		now := time.Now().UTC()
		expiry := now.AddDate(0, 0, 90)

		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*cfg, userID)

		_, tokenErr := driver.CreateToken(ctx, ac, db.CreateTokenParams{
			UserID:    types.NullableUserID{ID: ownerID, Valid: true},
			TokenType: "api_key",
			Token:     token,
			IssuedAt:  now.Format(time.RFC3339),
			ExpiresAt: types.NewTimestamp(expiry),
			Revoked:   false,
		})
		if tokenErr != nil {
			return ActionResultMsg{
				Title:   "Token Generation Failed",
				Message: fmt.Sprintf("Failed to store token: %s", tokenErr),
				IsError: true,
			}
		}

		return ActionResultMsg{
			Title:   "API Token Generated",
			Message: fmt.Sprintf("Token: %s\n\nExpires: %s\n\nUse as: Authorization: Bearer <token>\n\nCopy this token now — it cannot be shown again.", token, expiry.Format(time.RFC3339)),
			Width:   70,
		}
	}
}

// runCreateBackup creates a command to create a full database and file backup.
func runCreateBackup(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		driver := db.ConfigDB(*cfg)
		backupID := types.NewBackupID()
		startTime := time.Now().UTC()

		_, recordErr := driver.CreateBackup(db.CreateBackupParams{
			BackupID:    backupID,
			NodeID:      types.NodeID(cfg.Node_ID),
			BackupType:  types.BackupTypeFull,
			Status:      types.BackupStatusInProgress,
			StartedAt:   types.NewTimestamp(startTime),
			StoragePath: "",
			TriggeredBy: types.NullableString{String: "tui", Valid: true},
			Metadata:    types.JSONData{Valid: false},
		})

		path, sizeBytes, err := backup.CreateFullBackup(*cfg)
		if err != nil {
			if recordErr == nil {
				updateErr := driver.UpdateBackupStatus(db.UpdateBackupStatusParams{
					BackupID:     backupID,
					Status:       types.BackupStatusFailed,
					CompletedAt:  types.NewTimestamp(time.Now().UTC()),
					DurationMs:   types.NullableInt64{Int64: time.Since(startTime).Milliseconds(), Valid: true},
					ErrorMessage: types.NullableString{String: err.Error(), Valid: true},
				})
				if updateErr != nil {
					utility.DefaultLogger.Warn("Failed to record backup failure", updateErr)
				}
			}
			return ActionResultMsg{
				Title:   "Backup Failed",
				Message: fmt.Sprintf("Failed to create backup:\n%s", err),
				IsError: true,
			}
		}

		if recordErr == nil {
			updateErr := driver.UpdateBackupStatus(db.UpdateBackupStatusParams{
				BackupID:    backupID,
				Status:      types.BackupStatusCompleted,
				CompletedAt: types.NewTimestamp(time.Now().UTC()),
				DurationMs:  types.NullableInt64{Int64: time.Since(startTime).Milliseconds(), Valid: true},
				SizeBytes:   types.NullableInt64{Int64: sizeBytes, Valid: true},
			})
			if updateErr != nil {
				utility.DefaultLogger.Warn("Failed to record backup completion", updateErr)
			}
		}

		sizeStr := formatBackupSize(sizeBytes)
		return ActionResultMsg{
			Title:   "Backup Complete",
			Message: fmt.Sprintf("Backup created successfully.\n\nPath: %s\nSize: %s", path, sizeStr),
		}
	}
}

// formatBackupSize formats bytes to human-readable size string.
func formatBackupSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// runRegisterSSHKey creates a command to register the current SSH key as a new user.
func runRegisterSSHKey(p ActionParams) tea.Cmd {
	return func() tea.Msg {
		if p.SSHFingerprint == "" {
			return ActionResultMsg{
				Title:   "Registration Failed",
				Message: "No SSH key detected. Connect via SSH with a public key.",
				IsError: true,
			}
		}

		driver := db.ConfigDB(*p.Config)

		// Check if key is already registered
		existing, err := driver.GetUserSshKeyByFingerprint(p.SSHFingerprint)
		if err == nil && existing != nil {
			return ActionResultMsg{
				Title:   "Registration Failed",
				Message: fmt.Sprintf("This SSH key is already registered.\n\nFingerprint: %s\nKey ID: %s", p.SSHFingerprint, existing.SshKeyID),
				IsError: true,
			}
		}

		// Generate a random password and hash it (user authenticates via SSH key)
		tempPassword, err := utility.MakeRandomString()
		if err != nil {
			return ActionResultMsg{
				Title:   "Registration Failed",
				Message: fmt.Sprintf("Failed to generate credentials:\n%s", err),
				IsError: true,
			}
		}

		hash, err := auth.HashPassword(tempPassword)
		if err != nil {
			return ActionResultMsg{
				Title:   "Registration Failed",
				Message: fmt.Sprintf("Failed to hash password:\n%s", err),
				IsError: true,
			}
		}

		now := time.Now().UTC()
		ts := types.NewTimestamp(now)

		// Look up the admin role by label
		roles, err := driver.ListRoles()
		if err != nil {
			return ActionResultMsg{
				Title:   "Registration Failed",
				Message: fmt.Sprintf("Failed to list roles:\n%s", err),
				IsError: true,
			}
		}

		var roleID string
		for _, r := range *roles {
			if r.Label == "admin" {
				roleID = string(r.RoleID)
				break
			}
		}
		if roleID == "" {
			return ActionResultMsg{
				Title:   "Registration Failed",
				Message: "No admin role found. Run DB Init first.",
				IsError: true,
			}
		}

		// Derive a short username from the fingerprint
		shortFP := p.SSHFingerprint
		if len(shortFP) > 15 {
			shortFP = shortFP[7:15] // Skip "SHA256:" prefix, take 8 chars
		}
		username := fmt.Sprintf("user-%s", shortFP)

		ctx := context.Background()
		ac := middleware.AuditContextFromCLI(*p.Config, p.UserID)

		user, err := driver.CreateUser(ctx, ac, db.CreateUserParams{
			Username:     username,
			Name:         "SSH User",
			Email:        types.Email(fmt.Sprintf("%s@modula.local", username)),
			Hash:         hash,
			Role:         roleID,
			DateCreated:  ts,
			DateModified: ts,
		})
		if err != nil {
			return ActionResultMsg{
				Title:   "Registration Failed",
				Message: fmt.Sprintf("Failed to create user:\n%s", err),
				IsError: true,
			}
		}

		// Register the SSH key to the new user
		_, err = driver.CreateUserSshKey(ctx, ac, db.CreateUserSshKeyParams{
			UserID:      types.NullableUserID{ID: user.UserID, Valid: true},
			PublicKey:   p.SSHPublicKey,
			KeyType:     p.SSHKeyType,
			Fingerprint: p.SSHFingerprint,
			Label:       "Registered via TUI",
			DateCreated: ts,
		})
		if err != nil {
			return ActionResultMsg{
				Title:   "Registration Partial",
				Message: fmt.Sprintf("User created (%s) but SSH key registration failed:\n%s\n\nRegister the key manually.", username, err),
				IsError: true,
			}
		}

		return ActionResultMsg{
			Title: "SSH Key Registered",
			Message: fmt.Sprintf("New user created and SSH key registered.\n\nUsername: %s\nUser ID: %s\nKey Type: %s\nFingerprint: %s\n\nReconnect via SSH to use the new account.",
				username, user.UserID, p.SSHKeyType, p.SSHFingerprint),
		}
	}
}
