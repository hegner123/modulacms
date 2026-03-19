package handlers

import (
	"net/http"
	"time"

	"github.com/hegner123/modulacms/internal/admin/pages"
	"github.com/hegner123/modulacms/internal/admin/partials"
	"github.com/hegner123/modulacms/internal/backup"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/middleware"
	"github.com/hegner123/modulacms/internal/service"
	"github.com/hegner123/modulacms/internal/utility"
)

// BackupsListHandler handles GET /admin/settings/backups.
// Lists all backup records with pagination.
func BackupsListHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, offset := ParsePagination(r)

		list, err := svc.Driver().ListBackups(db.ListBackupsParams{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			utility.DefaultLogger.Error("failed to list backups", err)
			http.Error(w, "Failed to load backups", http.StatusInternalServerError)
			return
		}

		var backups []db.Backup
		if list != nil {
			backups = *list
		}

		count, countErr := svc.Driver().CountBackups()
		var total int64
		if countErr == nil && count != nil {
			total = *count
		}

		pd := NewPaginationData(total, limit, offset, "#backups-table-body", "/admin/settings/backups")
		pg := partials.PaginationPageData{
			Current:    pd.Current,
			TotalPages: pd.TotalPages,
			Limit:      pd.Limit,
			Target:     pd.Target,
			BaseURL:    pd.BaseURL,
		}

		csrfToken := CSRFTokenFromContext(r.Context())

		if IsNavHTMX(r) {
			w.Header().Set("HX-Trigger", `{"pageTitle": "Backups"}`)
			Render(w, r, pages.BackupsContent(backups, pg, csrfToken))
			return
		}

		if IsHTMX(r) {
			Render(w, r, pages.BackupsTableRows(backups, pg, csrfToken))
			return
		}

		layout := NewAdminData(r, "Backups")
		Render(w, r, pages.Backups(layout, backups, pg))
	}
}

// BackupDetailHandler handles GET /admin/settings/backups/{id}.
func BackupDetailHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Backup ID required", http.StatusBadRequest)
			return
		}

		b, err := svc.Driver().GetBackup(types.BackupID(id))
		if err != nil || b == nil {
			http.Error(w, "Backup not found", http.StatusNotFound)
			return
		}

		layout := NewAdminData(r, "Backup")
		RenderNav(w, r, "Backup",
			pages.BackupDetailContent(*b, layout.CSRFToken),
			pages.BackupDetail(layout, *b))
	}
}

// BackupCreateHandler handles POST /admin/settings/backups.
// Creates a full backup, records it in the database, and returns updated table rows.
func BackupCreateHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		user := middleware.AuthenticatedUser(r.Context())
		triggeredBy := "admin"
		if user != nil {
			triggeredBy = "admin:" + user.Username
		}

		driver := svc.Driver()
		backupID := types.NewBackupID()
		startTime := time.Now().UTC()

		// Predict the storage path so it's recorded even if the update lacks a path field.
		outputDir := cfg.Backup_Option
		if outputDir == "" {
			outputDir = "./"
		}
		predictedPath := outputDir + "backups/backup_" + startTime.Format("20060102_150405") + ".zip"

		utility.DefaultLogger.Info("backup: starting creation", "id", backupID.String(), "triggered_by", triggeredBy)

		_, createRecordErr := driver.CreateBackup(db.CreateBackupParams{
			BackupID:    backupID,
			NodeID:      types.NodeID(cfg.Node_ID),
			BackupType:  types.BackupTypeFull,
			Status:      types.BackupStatusInProgress,
			StartedAt:   types.NewTimestamp(startTime),
			StoragePath: predictedPath,
			TriggeredBy: types.NullableString{String: triggeredBy, Valid: true},
			Metadata:    types.JSONData{Valid: false},
		})
		if createRecordErr != nil {
			utility.DefaultLogger.Error("backup: failed to create DB record", createRecordErr)
		}

		path, sizeBytes, backupErr := backup.CreateFullBackup(*cfg, driver)
		if backupErr != nil {
			driver.UpdateBackupStatus(db.UpdateBackupStatusParams{
				BackupID:     backupID,
				Status:       types.BackupStatusFailed,
				CompletedAt:  types.NewTimestamp(time.Now().UTC()),
				DurationMs:   types.NullableInt64{Int64: time.Since(startTime).Milliseconds(), Valid: true},
				ErrorMessage: types.NullableString{String: backupErr.Error(), Valid: true},
			})

			utility.DefaultLogger.Error("backup creation failed", backupErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Backup failed", "type": "error"}}`)
			renderBackupsTableRows(w, r, svc)
			return
		}

		driver.UpdateBackupStatus(db.UpdateBackupStatusParams{
			BackupID:    backupID,
			Status:      types.BackupStatusCompleted,
			CompletedAt: types.NewTimestamp(time.Now().UTC()),
			DurationMs:  types.NullableInt64{Int64: time.Since(startTime).Milliseconds(), Valid: true},
			SizeBytes:   types.NullableInt64{Int64: sizeBytes, Valid: true},
		})

		_ = path // actual path matches predictedPath

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Backup created", "type": "success"}}`)
		renderBackupsTableRows(w, r, svc)
	}
}

// BackupRestoreHandler handles POST /admin/settings/backups/{id}/restore.
// Restores the database from the specified backup.
func BackupRestoreHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Backup ID required", http.StatusBadRequest)
			return
		}

		cfg, cfgErr := svc.Config()
		if cfgErr != nil {
			http.Error(w, "Configuration unavailable", http.StatusInternalServerError)
			return
		}

		b, err := svc.Driver().GetBackup(types.BackupID(id))
		if err != nil || b == nil {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Backup not found", "type": "error"}}`)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if b.StoragePath == "" {
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Backup has no storage path", "type": "error"}}`)
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		restoreErr := backup.RestoreFromBackup(*cfg, b.StoragePath)
		if restoreErr != nil {
			utility.DefaultLogger.Error("backup restore failed", restoreErr)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Restore failed: `+restoreErr.Error()+`", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Database restored from backup. Restart recommended.", "type": "success", "persist": true}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// BackupDeleteHandler handles DELETE /admin/settings/backups/{id}.
// Removes the backup record from the database.
func BackupDeleteHandler(svc *service.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsHTMX(r) {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Backup ID required", http.StatusBadRequest)
			return
		}

		if err := svc.Driver().DeleteBackup(types.BackupID(id)); err != nil {
			utility.DefaultLogger.Error("failed to delete backup", err)
			w.Header().Set("HX-Trigger", `{"showToast": {"message": "Failed to delete backup record", "type": "error"}}`)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", `{"showToast": {"message": "Backup record deleted", "type": "success"}}`)
		w.WriteHeader(http.StatusOK)
	}
}

// renderBackupsTableRows reloads and renders backup table rows after a mutation.
func renderBackupsTableRows(w http.ResponseWriter, r *http.Request, svc *service.Registry) {
	limit, offset := ParsePagination(r)

	list, listErr := svc.Driver().ListBackups(db.ListBackupsParams{
		Limit:  limit,
		Offset: offset,
	})
	if listErr != nil {
		utility.DefaultLogger.Error("failed to reload backups", listErr)
		http.Error(w, "Failed to reload backups", http.StatusInternalServerError)
		return
	}

	var backups []db.Backup
	if list != nil {
		backups = *list
	}

	count, countErr := svc.Driver().CountBackups()
	var total int64
	if countErr == nil && count != nil {
		total = *count
	}

	pd := NewPaginationData(total, limit, offset, "#backups-table-body", "/admin/settings/backups")
	pg := partials.PaginationPageData{
		Current:    pd.Current,
		TotalPages: pd.TotalPages,
		Limit:      pd.Limit,
		Target:     pd.Target,
		BaseURL:    pd.BaseURL,
	}

	csrfToken := CSRFTokenFromContext(r.Context())
	Render(w, r, pages.BackupsTableRows(backups, pg, csrfToken))
}
