package update

import (
	"context"
	"runtime"
	"time"

	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/hegner123/modulacms/internal/webhooks"
)

// SchedulerConfig holds the settings that control the update scheduler's
// behavior. All fields map directly to modula.config.json keys.
type SchedulerConfig struct {
	// Interval is the check frequency: "startup" for a single boot-time check,
	// or a time.Duration string like "24h".
	Interval string

	// Channel filters releases: "stable" skips prereleases, "beta" includes them.
	Channel string

	// NotifyOnly when true limits the scheduler to logging and firing a webhook.
	// When false the scheduler also downloads and applies the binary (a restart
	// is still required to run the new version).
	NotifyOnly bool
}

// StartScheduler runs a background goroutine that periodically checks GitHub
// for new releases. Depending on the configuration it will:
//
//  1. Always fire an update.available webhook event when a newer version exists.
//  2. If NotifyOnly is false, download the platform binary and replace the
//     current executable (the process must still be restarted externally).
func StartScheduler(ctx context.Context, cfg SchedulerConfig, dispatcher publishing.WebhookDispatcher) {
	channel := cfg.Channel
	if channel == "" {
		channel = "stable"
	}

	check := func() {
		currentVersion := utility.GetCurrentVersion()
		release, available, err := CheckForUpdates(currentVersion, channel)
		if err != nil {
			utility.DefaultLogger.Warn("scheduled update check failed", err)
			return
		}
		if !available || release == nil {
			return
		}

		utility.DefaultLogger.Info("new release detected", "current", currentVersion, "latest", release.TagName)

		// Fire webhook regardless of notify-only setting.
		if dispatcher != nil {
			dispatcher.Dispatch(ctx, webhooks.EventUpdateAvailable, map[string]any{
				"current_version": currentVersion,
				"latest_version":  release.TagName,
				"published_at":    release.PublishedAt,
				"channel":         channel,
				"release_name":    release.Name,
			})
		}

		if cfg.NotifyOnly {
			utility.DefaultLogger.Info("update available (notify-only mode, skipping auto-apply)",
				"latest", release.TagName)
			return
		}

		// Auto-apply: download platform binary and replace the running executable.
		downloadURL, err := GetDownloadURL(release, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			utility.DefaultLogger.Error("no compatible binary for auto-update", err,
				"os", runtime.GOOS, "arch", runtime.GOARCH)
			return
		}

		utility.DefaultLogger.Info("downloading update", "version", release.TagName, "url", downloadURL)
		tempPath, err := DownloadUpdate(downloadURL)
		if err != nil {
			utility.DefaultLogger.Error("auto-update download failed", err)
			return
		}

		if err := ApplyUpdate(tempPath); err != nil {
			utility.DefaultLogger.Error("auto-update apply failed", err)
			return
		}

		utility.DefaultLogger.Info("update applied — restart the process to run the new version",
			"from", currentVersion, "to", release.TagName)
	}

	// "startup" means fire once and exit.
	if cfg.Interval == "" || cfg.Interval == "startup" {
		go check()
		return
	}

	dur, err := time.ParseDuration(cfg.Interval)
	if err != nil {
		utility.DefaultLogger.Warn("invalid update_check_interval, falling back to single check",
			err, "interval", cfg.Interval)
		go check()
		return
	}

	go func() {
		// Initial check on startup.
		check()

		ticker := time.NewTicker(dur)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				check()
			}
		}
	}()
}
