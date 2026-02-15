// Package install provides the interactive installation workflow for ModulaCMS,
// including database setup, configuration file generation, and validation of required services.
package install

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
)

// maxRetries defines the maximum number of installation attempts before failure.
const maxRetries = 3

// RunInstall runs the interactive installation process with retry support.
// When yes is non-nil and true, all prompts are skipped and defaults are used.
// adminPassword provides the system admin password; required for non-interactive mode.
func RunInstall(v *bool, yes *bool, adminPassword *string) error {
	if yes != nil && *yes {
		return runInstallDefaults(v, adminPassword)
	}
	return runInstallWithRetry(v, adminPassword, maxRetries)
}

func runInstallDefaults(v *bool, adminPassword *string) error {
	if adminPassword == nil || *adminPassword == "" {
		return fmt.Errorf("admin password is required for non-interactive install (use --admin-password)")
	}

	if err := ValidatePassword(*adminPassword); err != nil {
		return fmt.Errorf("invalid admin password: %w", err)
	}

	adminHash, err := auth.HashPassword(*adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	iarg := &InstallArguments{
		UseDefaultConfig:  true,
		ConfigPath:        "config.json",
		Config:            config.DefaultConfig(),
		DB_Driver:         SQLITE,
		Create_Tables:     true,
		AdminPasswordHash: adminHash,
	}

	err = writeConfigFile(iarg)
	if err != nil {
		PrintError(err.Error())
		return err
	}

	progress := NewInstallProgress()

	progress.AddStep("Database connection", "Checking database connection", func() error {
		_, checkErr := CheckDb(v, iarg.Config)
		return checkErr
	})

	progress.AddStep("Database setup", "Setting up database tables", func() error {
		return CreateDbSimple(iarg.ConfigPath, &iarg.Config, iarg.AdminPasswordHash)
	})

	var bucketStatus string
	progress.AddStep("Bucket connection", "Checking S3 bucket connection", func() error {
		bucketStatus, _ = CheckBucket(v, &iarg.Config)
		return nil
	})

	err = progress.Run()
	if err != nil {
		PrintError(err.Error())
		return err
	}

	// Check backup tool availability (non-fatal warning)
	if _, toolErr := CheckBackupTools(iarg.Config.Db_Driver); toolErr != nil {
		PrintWarning(fmt.Sprintf("Backup tools: %v", toolErr))
	}

	PrintSuccess("Installation completed successfully!")
	printInstallSummary(iarg, bucketStatus)
	return nil
}

func runInstallWithRetry(v *bool, adminPassword *string, retriesLeft int) error {
	if retriesLeft <= 0 {
		PrintError("Installation failed after maximum retries")
		return ErrMaxRetries(maxRetries)
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	s := fmt.Sprintf("Would you like to install ModulaCMS at \n %s\n", dir)
	runInstall := false
	c := huh.NewConfirm().Title(s).Value(&runInstall)
	err = c.Run()
	if err != nil {
		return ErrUserAborted()
	}

	if !runInstall {
		return ErrUserAborted()
	}

	iarg, err := RunInstallIO()
	if err != nil {
		return err
	}

	// If admin password was provided via flag, hash it and skip the interactive prompt
	if adminPassword != nil && *adminPassword != "" {
		if err := ValidatePassword(*adminPassword); err != nil {
			return fmt.Errorf("invalid admin password: %w", err)
		}
		hash, err := auth.HashPassword(*adminPassword)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}
		iarg.AdminPasswordHash = hash
	}

	// Write config file
	err = writeConfigFile(iarg)
	if err != nil {
		PrintError(err.Error())
		return err
	}

	// Run installation checks with progress indicators
	progress := NewInstallProgress()

	progress.AddStep("Database connection", "Checking database connection", func() error {
		_, checkErr := CheckDb(v, iarg.Config)
		return checkErr
	})

	progress.AddStep("Database setup", "Setting up database tables", func() error {
		return CreateDbSimple(iarg.ConfigPath, &iarg.Config, iarg.AdminPasswordHash)
	})

	// Run with warnings for optional checks
	var bucketStatus string
	progress.AddStep("Bucket connection", "Checking S3 bucket connection", func() error {
		bucketStatus, _ = CheckBucket(v, &iarg.Config)
		return nil // Bucket is optional, don't fail on error
	})

	err = progress.Run()
	if err != nil {
		PrintError(err.Error())

		// Offer retry
		retry := false
		retryPrompt := huh.NewConfirm().
			Title("Would you like to try again?").
			Value(&retry)

		if promptErr := retryPrompt.Run(); promptErr != nil {
			return err
		}

		if retry {
			return runInstallWithRetry(v, adminPassword, retriesLeft-1)
		}
		return err
	}

	// Print final status
	PrintSuccess("Installation completed successfully!")

	if bucketStatus != "Connected" && bucketStatus != "" {
		PrintWarning(fmt.Sprintf("S3 bucket: %s (media storage will be unavailable)", bucketStatus))
	}

	// Check backup tool availability (non-fatal warning)
	if _, toolErr := CheckBackupTools(iarg.Config.Db_Driver); toolErr != nil {
		PrintWarning(fmt.Sprintf("Backup tools: %v", toolErr))
	}

	printInstallSummary(iarg, bucketStatus)

	return nil
}

func writeConfigFile(iarg *InstallArguments) error {
	// Back up existing config file before overwriting
	if _, statErr := os.Stat(iarg.ConfigPath); statErr == nil {
		existing, readErr := os.ReadFile(iarg.ConfigPath)
		if readErr == nil {
			bakPath := iarg.ConfigPath + ".bak"
			writeErr := os.WriteFile(bakPath, existing, 0644)
			if writeErr != nil {
				utility.DefaultLogger.Warn("Failed to create config backup: "+bakPath, writeErr)
			}
		}
	}

	f, err := os.OpenFile(iarg.ConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return ErrConfigWrite(err, iarg.ConfigPath)
	}
	defer f.Close()

	j, err := utility.FormatJSON(iarg.Config)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(f, j)
	if err != nil {
		return ErrConfigWrite(err, iarg.ConfigPath)
	}

	return nil
}

func printInstallSummary(iarg *InstallArguments, bucketStatus string) {
	var b strings.Builder
	b.WriteString("\n--- Installation Summary ---\n")
	b.WriteString(fmt.Sprintf("  Config file:   %s\n", iarg.ConfigPath))
	b.WriteString(fmt.Sprintf("  DB driver:     %s\n", iarg.Config.Db_Driver))
	b.WriteString(fmt.Sprintf("  DB URL:        %s\n", iarg.Config.Db_URL))
	b.WriteString(fmt.Sprintf("  DB name:       %s\n", iarg.Config.Db_Name))
	b.WriteString(fmt.Sprintf("  HTTP port:     %s\n", iarg.Config.Port))
	b.WriteString(fmt.Sprintf("  HTTPS port:    %s\n", iarg.Config.SSL_Port))
	b.WriteString(fmt.Sprintf("  SSH port:      %s\n", iarg.Config.SSH_Port))

	if iarg.Config.Client_Site != "" {
		b.WriteString(fmt.Sprintf("  Client site:   %s\n", iarg.Config.Client_Site))
	}
	if iarg.Config.Admin_Site != "" {
		b.WriteString(fmt.Sprintf("  Admin site:    %s\n", iarg.Config.Admin_Site))
	}

	if bucketStatus == "Connected" {
		b.WriteString("  S3 storage:    Configured\n")
	} else {
		b.WriteString("  S3 storage:    Not configured\n")
	}

	if iarg.Config.Oauth_Client_Id != "" {
		b.WriteString("  OAuth:         Configured\n")
	} else {
		b.WriteString("  OAuth:         Not configured\n")
	}

	b.WriteString("\n--- Admin Account ---\n")
	b.WriteString("  Email:         system@modulacms.local\n")
	b.WriteString("  Username:      system\n")

	b.WriteString("\n--- Next Steps ---\n")
	b.WriteString("  Gen certs:     ./modulacms-x86 cert generate\n")
	b.WriteString("  Start server:  ./modulacms-x86 serve\n")
	b.WriteString(fmt.Sprintf("  SSH access:    ssh localhost -p %s\n", iarg.Config.SSH_Port))
	b.WriteString("---\n")

	fmt.Print(b.String())
}
