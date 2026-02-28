package router

import (
	"context"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/utility"
)

// pruneInterval is the interval between periodic version retention sweeps.
// Runs independently from the publish schedule check.
const pruneInterval = 1 * time.Hour

// StartPublishScheduler runs a background loop that checks for content scheduled
// for publishing and publishes it when the scheduled time arrives. It also runs
// periodic retention cleanup to prune excess versions.
//
// The scheduler performs a catch-up pass on startup to handle any items whose
// publish_at time passed while the server was down.
//
// It respects ctx.Done() for graceful shutdown.
func StartPublishScheduler(ctx context.Context, driver db.DbDriver, cfg config.Config, interval time.Duration) {
	utility.DefaultLogger.Info("publish scheduler started", "interval", interval)

	// Catch-up pass: publish anything overdue from before server started.
	publishDueContent(ctx, driver, cfg)
	publishDueAdminContent(ctx, driver, cfg)

	publishTicker := time.NewTicker(interval)
	defer publishTicker.Stop()

	pruneTicker := time.NewTicker(pruneInterval)
	defer pruneTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			utility.DefaultLogger.Info("publish scheduler stopped")
			return
		case <-publishTicker.C:
			publishDueContent(ctx, driver, cfg)
			publishDueAdminContent(ctx, driver, cfg)
		case <-pruneTicker.C:
			pruneAllContentVersions(driver, cfg)
			pruneAllAdminContentVersions(driver, cfg)
		}
	}
}

// publishDueContent finds all content_data rows where publish_at <= now and
// status is 'draft', then publishes each one.
func publishDueContent(ctx context.Context, driver db.DbDriver, cfg config.Config) {
	now := types.TimestampNow()
	items, err := driver.ListContentDataDueForPublish(now)
	if err != nil {
		utility.DefaultLogger.Error("scheduler: list due content failed", err)
		return
	}
	if items == nil || len(*items) == 0 {
		return
	}

	retentionCap := cfg.VersionMaxPerContent()
	for _, item := range *items {
		ac := audited.Ctx(types.NewNodeID(), item.AuthorID, "scheduled-publish", "system")

		_, pubErr := publishing.PublishContent(ctx, driver, item.ContentDataID, item.AuthorID, ac, retentionCap)
		if pubErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("scheduler: publish content %s failed", item.ContentDataID), pubErr)
			continue
		}

		// Clear the publish_at field after successful publish.
		clearErr := driver.ClearContentDataSchedule(ctx, db.ClearContentDataScheduleParams{
			DateModified:  types.TimestampNow(),
			ContentDataID: item.ContentDataID,
		})
		if clearErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("scheduler: clear publish_at for %s failed", item.ContentDataID), clearErr)
		}

		utility.DefaultLogger.Info("scheduler: published content", "content_data_id", item.ContentDataID)
	}
}

// publishDueAdminContent finds all admin_content_data rows where publish_at <= now
// and status is 'draft', then publishes each one.
func publishDueAdminContent(ctx context.Context, driver db.DbDriver, cfg config.Config) {
	now := types.TimestampNow()
	items, err := driver.ListAdminContentDataDueForPublish(now)
	if err != nil {
		utility.DefaultLogger.Error("scheduler: list due admin content failed", err)
		return
	}
	if items == nil || len(*items) == 0 {
		return
	}

	retentionCap := cfg.VersionMaxPerContent()
	for _, item := range *items {
		ac := audited.Ctx(types.NewNodeID(), item.AuthorID, "scheduled-publish", "system")

		pubErr := publishing.PublishAdminContent(ctx, driver, item.AdminContentDataID, item.AuthorID, ac, retentionCap)
		if pubErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("scheduler: publish admin content %s failed", item.AdminContentDataID), pubErr)
			continue
		}

		// Clear the publish_at field after successful publish.
		clearErr := driver.ClearAdminContentDataSchedule(ctx, db.ClearAdminContentDataScheduleParams{
			DateModified:       types.TimestampNow(),
			AdminContentDataID: item.AdminContentDataID,
		})
		if clearErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("scheduler: clear publish_at for admin %s failed", item.AdminContentDataID), clearErr)
		}

		utility.DefaultLogger.Info("scheduler: published admin content", "admin_content_data_id", item.AdminContentDataID)
	}
}

// pruneAllContentVersions iterates all content data items and prunes excess
// unpublished, unlabeled versions that exceed the retention cap.
func pruneAllContentVersions(driver db.DbDriver, cfg config.Config) {
	retentionCap := cfg.VersionMaxPerContent()
	if retentionCap <= 0 {
		return
	}

	items, err := driver.ListContentData()
	if err != nil {
		utility.DefaultLogger.Error("scheduler: list content data for prune sweep failed", err)
		return
	}
	if items == nil || len(*items) == 0 {
		return
	}

	pruned := 0
	for _, item := range *items {
		publishing.PruneExcessVersions(driver, item.ContentDataID, "", retentionCap)
		pruned++
	}

	if pruned > 0 {
		utility.DefaultLogger.Info("scheduler: periodic prune sweep complete", "content_items_checked", pruned)
	}
}

// pruneAllAdminContentVersions iterates all admin content data items and prunes
// excess unpublished, unlabeled versions that exceed the retention cap.
func pruneAllAdminContentVersions(driver db.DbDriver, cfg config.Config) {
	retentionCap := cfg.VersionMaxPerContent()
	if retentionCap <= 0 {
		return
	}

	items, err := driver.ListAdminContentData()
	if err != nil {
		utility.DefaultLogger.Error("scheduler: list admin content data for prune sweep failed", err)
		return
	}
	if items == nil || len(*items) == 0 {
		return
	}

	pruned := 0
	for _, item := range *items {
		publishing.PruneExcessAdminVersions(driver, item.AdminContentDataID, "", retentionCap)
		pruned++
	}

	if pruned > 0 {
		utility.DefaultLogger.Info("scheduler: periodic admin prune sweep complete", "admin_content_items_checked", pruned)
	}
}
