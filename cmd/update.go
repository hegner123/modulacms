package main

import (
	"fmt"
	"runtime"

	"github.com/hegner123/modulacms/internal/update"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

// updateCmd provides the CLI command to check for and apply updates to ModulaCMS.
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for and apply updates",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		utility.DefaultLogger.Info("Checking for updates...")

		currentVersion := utility.GetCurrentVersion()
		utility.DefaultLogger.Info("Current version", currentVersion)

		release, available, err := update.CheckForUpdates(currentVersion, "stable")
		if err != nil {
			return fmt.Errorf("update check failed: %w", err)
		}

		if !available {
			utility.DefaultLogger.Info("Already running latest version")
			return nil
		}

		utility.DefaultLogger.Info("Update available", release.TagName)

		downloadURL, err := update.GetDownloadURL(release, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			return fmt.Errorf("no compatible binary found: %w", err)
		}

		utility.DefaultLogger.Info("Downloading update...")
		tempPath, err := update.DownloadUpdate(downloadURL)
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		utility.DefaultLogger.Info("Applying update...")
		if err := update.ApplyUpdate(tempPath); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		utility.DefaultLogger.Info("Update complete! Please restart ModulaCMS.")
		return nil
	},
}
