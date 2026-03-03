package remote

import (
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// Backup: SDK <-> db (write-only for TUI: create, update status)
// ---------------------------------------------------------------------------

// backupToDb converts a SDK Backup to a db Backup.
func backupToDb(s *modula.Backup) db.Backup {
	return db.Backup{
		BackupID:       types.BackupID(string(s.BackupID)),
		NodeID:         types.NodeID(s.NodeID),
		BackupType:     types.BackupType(s.BackupType),
		Status:         types.BackupStatus(s.Status),
		StartedAt:      sdkTimestampToDb(s.StartedAt),
		CompletedAt:    sdkTimestampToDb(s.CompletedAt),
		DurationMs:     nullInt64(s.DurationMs),
		RecordCount:    nullInt64(s.RecordCount),
		SizeBytes:      nullInt64(s.SizeBytes),
		ReplicationLsn: nullNullableString(s.ReplicationLsn),
		HlcTimestamp:   types.HLC(s.HlcTimestamp),
		StoragePath:    s.StoragePath,
		Checksum:       nullNullableString(s.Checksum),
		TriggeredBy:    nullNullableString(s.TriggeredBy),
		ErrorMessage:   nullNullableString(s.ErrorMessage),
		Metadata:       rawToJSONData(s.Metadata),
	}
}

// backupFromDb converts a db Backup to a SDK Backup.
func backupFromDb(d db.Backup) modula.Backup {
	return modula.Backup{
		BackupID:       modula.BackupID(string(d.BackupID)),
		NodeID:         string(d.NodeID),
		BackupType:     string(d.BackupType),
		Status:         string(d.Status),
		StartedAt:      dbTimestampToSdk(d.StartedAt),
		CompletedAt:    dbTimestampToSdk(d.CompletedAt),
		DurationMs:     int64Ptr(d.DurationMs),
		RecordCount:    int64Ptr(d.RecordCount),
		SizeBytes:      int64Ptr(d.SizeBytes),
		ReplicationLsn: nullableStringPtr(d.ReplicationLsn),
		HlcTimestamp:   int64(d.HlcTimestamp),
		StoragePath:    d.StoragePath,
		Checksum:       nullableStringPtr(d.Checksum),
		TriggeredBy:    nullableStringPtr(d.TriggeredBy),
		ErrorMessage:   nullableStringPtr(d.ErrorMessage),
		Metadata:       jsonDataToRaw(d.Metadata),
	}
}

// ---------------------------------------------------------------------------
// Locale: SDK <-> db (read-only for TUI)
// ---------------------------------------------------------------------------

// localeToDb converts a SDK Locale to a db Locale.
func localeToDb(s *modula.Locale) db.Locale {
	return db.Locale{
		LocaleID:     types.LocaleID(string(s.LocaleID)),
		Code:         s.Code,
		Label:        s.Label,
		IsDefault:    s.IsDefault,
		IsEnabled:    s.IsEnabled,
		FallbackCode: s.FallbackCode,
		SortOrder:    s.SortOrder,
		DateCreated:  types.Timestamp{}, // SDK Locale uses string for DateCreated
	}
}

// localeFromDb converts a db Locale to a SDK Locale.
func localeFromDb(d db.Locale) modula.Locale {
	return modula.Locale{
		LocaleID:     modula.LocaleID(string(d.LocaleID)),
		Code:         d.Code,
		Label:        d.Label,
		IsDefault:    d.IsDefault,
		IsEnabled:    d.IsEnabled,
		FallbackCode: d.FallbackCode,
		SortOrder:    d.SortOrder,
		DateCreated:  d.DateCreated.String(),
	}
}

// ---------------------------------------------------------------------------
// Webhook: SDK <-> db (read-only for TUI)
// ---------------------------------------------------------------------------

// webhookToDb converts a SDK Webhook to a db Webhook.
func webhookToDb(s *modula.Webhook) db.Webhook {
	return db.Webhook{
		WebhookID:    types.WebhookID(string(s.WebhookID)),
		Name:         s.Name,
		URL:          s.URL,
		Secret:       s.Secret,
		Events:       s.Events,
		IsActive:     s.IsActive,
		Headers:      s.Headers,
		AuthorID:     types.UserID(string(s.AuthorID)),
		DateCreated:  sdkTimestampToDb(s.DateCreated),
		DateModified: sdkTimestampToDb(s.DateModified),
	}
}

// webhookFromDb converts a db Webhook to a SDK Webhook.
func webhookFromDb(d db.Webhook) modula.Webhook {
	return modula.Webhook{
		WebhookID:    modula.WebhookID(string(d.WebhookID)),
		Name:         d.Name,
		URL:          d.URL,
		Secret:       d.Secret,
		Events:       d.Events,
		IsActive:     d.IsActive,
		Headers:      d.Headers,
		AuthorID:     modula.UserID(string(d.AuthorID)),
		DateCreated:  dbTimestampToSdk(d.DateCreated),
		DateModified: dbTimestampToSdk(d.DateModified),
	}
}

// ---------------------------------------------------------------------------
// Table: SDK <-> db (read-only for TUI)
// ---------------------------------------------------------------------------

// tableToDb converts a SDK Table to a db Tables.
func tableToDb(s *modula.Table) db.Tables {
	return db.Tables{
		ID:       string(s.ID),
		Label:    s.Label,
		AuthorID: nullUserID(s.AuthorID),
	}
}

// tableFromDb converts a db Tables to a SDK Table.
func tableFromDb(d db.Tables) modula.Table {
	return modula.Table{
		ID:       modula.TableID(d.ID),
		Label:    d.Label,
		AuthorID: userIDPtr(d.AuthorID),
	}
}
