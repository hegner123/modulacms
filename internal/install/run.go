package install

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/utility"
)

const maxRetries = 3

// RunInstall runs the interactive installation process with retry support
func RunInstall(v *bool) error {
	return runInstallWithRetry(v, maxRetries)
}

func runInstallWithRetry(v *bool, retriesLeft int) error {
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
		return CreateDbSimple(iarg.ConfigPath, &iarg.Config)
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
			return runInstallWithRetry(v, retriesLeft-1)
		}
		return err
	}

	// Print final status
	PrintSuccess("Installation completed successfully!")

	if bucketStatus != "Connected" && bucketStatus != "" {
		PrintWarning(fmt.Sprintf("S3 bucket: %s (media storage will be unavailable)", bucketStatus))
	}

	return nil
}

func writeConfigFile(iarg *InstallArguments) error {
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
