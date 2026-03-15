package db

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/db/types"
)

// nullableBoolString converts a NullableBool to a display string.
func nullableBoolString(nb types.NullableBool) string {
	if !nb.Valid {
		return ""
	}
	return fmt.Sprintf("%t", nb.Bool)
}

// This file imports all resources

// Resource types

// StringUsers represents user data as strings for TUI display.
type StringUsers struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Hash         string `json:"hash"`
	Role         string `json:"role"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// StringRoutes represents route data as strings for TUI display.
type StringRoutes struct {
	RouteID      string `json:"route_id"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

// StringFields represents field data as strings for TUI display.
type StringFields struct {
	FieldID      string `json:"field_id"`
	ParentID     string `json:"parent_id"`
	SortOrder    string `json:"sort_order"`
	Name         string `json:"name"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Validation   string `json:"validation"`
	UIConfig     string `json:"ui_config"`
	Type         string `json:"type"`
	Translatable string `json:"translatable"`
	Roles        string `json:"roles"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

// StringMedia represents media data as strings for TUI display.
type StringMedia struct {
	MediaID      string `json:"media_id"`
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Alt          string `json:"alt"`
	Caption      string `json:"caption"`
	Description  string `json:"description"`
	Class        string `json:"class"`
	Mimetype     string `json:"mimetype"`
	Dimensions   string `json:"dimensions"`
	URL          string `json:"url"`
	Srcset       string `json:"srcset"`
	FocalX       string `json:"focal_x"`
	FocalY       string `json:"focal_y"`
	AuthorID     string `json:"author_id"`
	FolderID     string `json:"folder_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// StringMediaDimensions represents media dimension data as strings for TUI display.
type StringMediaDimensions struct {
	MdID        string `json:"md_id"`
	Label       string `json:"label"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
}

// StringMediaFolder represents media folder data as strings for TUI display.
type StringMediaFolder struct {
	FolderID     string `json:"folder_id"`
	Name         string `json:"name"`
	ParentID     string `json:"parent_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// StringTokens represents token data as strings for TUI display.
type StringTokens struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	Token     string `json:"token"`
	IssuedAt  string `json:"issued_at"`
	ExpiresAt string `json:"expires_at"`
	Revoked   string `json:"revoked"`
}

// StringDatatypes represents datatype data as strings for TUI display.
type StringDatatypes struct {
	DatatypeID   string `json:"datatype_id"`
	ParentID     string `json:"parent_id"`
	SortOrder    string `json:"sort_order"`
	Name         string `json:"name"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

// StringSessions represents session data as strings for TUI display.
type StringSessions struct {
	SessionID   string `json:"session_id"`
	UserID      string `json:"user_id"`
	DateCreated string `json:"date_created"`
	ExpiresAt   string `json:"expires_at"`
	LastAccess  string `json:"last_access"`
	IpAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	SessionData string `json:"session_data"`
}

// StringRoles represents role data as strings for TUI display.
type StringRoles struct {
	RoleID          string `json:"role_id"`
	Label           string `json:"label"`
	SystemProtected string `json:"system_protected"`
}

// StringPermissions represents permission data as strings for TUI display.
type StringPermissions struct {
	PermissionID string `json:"permission_id"`
	Label        string `json:"label"`
}

// StringFieldTypes represents field type data as strings for TUI display.
type StringFieldTypes struct {
	FieldTypeID string `json:"field_type_id"`
	Type        string `json:"type"`
	Label       string `json:"label"`
}

// StringAdminFieldTypes represents admin field type data as strings for TUI display.
type StringAdminFieldTypes struct {
	AdminFieldTypeID string `json:"admin_field_type_id"`
	Type             string `json:"type"`
	Label            string `json:"label"`
}

// StringContentData represents content data as strings for TUI display.
type StringContentData struct {
	ContentDataID string `json:"content_data_id"`
	ParentID      string `json:"parent_id"`
	FirstChildID  string `json:"first_child_id"`
	NextSiblingID string `json:"next_sibling_id"`
	PrevSiblingID string `json:"prev_sibling_id"`
	RootID        string `json:"root_id"`
	RouteID       string `json:"route_id"`
	DatatypeID    string `json:"datatype_id"`
	AuthorID      string `json:"author_id"`
	Status        string `json:"status"`
	DateCreated   string `json:"date_created"`
	DateModified  string `json:"date_modified"`
	PublishedAt   string `json:"published_at"`
	PublishedBy   string `json:"published_by"`
	PublishAt     string `json:"publish_at"`
	Revision      string `json:"revision"`
	History       string `json:"history"`
}

// StringContentFields represents content field data as strings for TUI display.
type StringContentFields struct {
	ContentFieldID string `json:"content_field_id"`
	RouteID        string `json:"route_id"`
	RootID         string `json:"root_id"`
	ContentDataID  string `json:"content_data_id"`
	FieldID        string `json:"field_id"`
	FieldValue     string `json:"field_value"`
	Locale         string `json:"locale"`
	AuthorID       string `json:"author_id"`
	DateCreated    string `json:"date_created"`
	DateModified   string `json:"date_modified"`
	History        string `json:"history"`
}

// StringAdminRoutes represents admin route data as strings for TUI display.
type StringAdminRoutes struct {
	AdminRouteID string `json:"admin_route_id"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

// StringAdminFields represents admin field data as strings for TUI display.
type StringAdminFields struct {
	AdminFieldID string `json:"admin_field_id"`
	ParentID     string `json:"parent_id"`
	SortOrder    string `json:"sort_order"`
	Name         string `json:"name"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Validation   string `json:"validation"`
	UIConfig     string `json:"ui_config"`
	Type         string `json:"type"`
	Translatable string `json:"translatable"`
	Roles        string `json:"roles"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

// StringAdminDatatypes represents admin datatype data as strings for TUI display.
type StringAdminDatatypes struct {
	AdminDatatypeID string `json:"admin_datatype_id"`
	ParentID        string `json:"parent_id"`
	SortOrder       string `json:"sort_order"`
	Name            string `json:"name"`
	Label           string `json:"label"`
	Type            string `json:"type"`
	AuthorID        string `json:"author_id"`
	DateCreated     string `json:"date_created"`
	DateModified    string `json:"date_modified"`
	History         string `json:"history"`
}

// StringAdminContentData represents admin content data as strings for TUI display.
type StringAdminContentData struct {
	AdminContentDataID string `json:"admin_content_data_id"`
	ParentID           string `json:"parent_id"`
	FirstChildID       string `json:"first_child_id"`
	NextSiblingID      string `json:"next_sibling_id"`
	PrevSiblingID      string `json:"prev_sibling_id"`
	RootID             string `json:"root_id"`
	AdminRouteID       string `json:"admin_route_id"`
	AdminDatatypeID    string `json:"admin_datatype_id"`
	AuthorID           string `json:"author_id"`
	Status             string `json:"status"`
	DateCreated        string `json:"date_created"`
	DateModified       string `json:"date_modified"`
	PublishedAt        string `json:"published_at"`
	PublishedBy        string `json:"published_by"`
	PublishAt          string `json:"publish_at"`
	Revision           string `json:"revision"`
	History            string `json:"history"`
}

// StringAdminContentFields represents admin content field data as strings for TUI display.
type StringAdminContentFields struct {
	AdminContentFieldID string `json:"admin_content_field_id"`
	AdminRouteID        string `json:"admin_route_id"`
	RootID              string `json:"root_id"`
	AdminContentDataID  string `json:"admin_content_data_id"`
	AdminFieldID        string `json:"admin_field_id"`
	AdminFieldValue     string `json:"admin_field_value"`
	Locale              string `json:"locale"`
	AuthorID            string `json:"author_id"`
	DateCreated         string `json:"date_created"`
	DateModified        string `json:"date_modified"`
	History             string `json:"history"`
}

// StringTables represents table data as strings for TUI display.
type StringTables struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	AuthorID string `json:"author_id"`
}

// StringUserOauth represents user OAuth data as strings for TUI display.
type StringUserOauth struct {
	UserOauthID         string `json:"user_oauth_id"`
	UserID              string `json:"user_id"`
	OauthProvider       string `json:"oauth_provider"`
	OauthProviderUserID string `json:"oauth_provider_user_id"`
	AccessToken         string `json:"access_token"`
	RefreshToken        string `json:"refresh_token"`
	TokenExpiresAt      string `json:"token_expires_at"`
	DateCreated         string `json:"date_created"`
}

// StringBackup represents backup data as strings for TUI display.
type StringBackup struct {
	BackupID       string `json:"backup_id"`
	NodeID         string `json:"node_id"`
	BackupType     string `json:"backup_type"`
	Status         string `json:"status"`
	StartedAt      string `json:"started_at"`
	CompletedAt    string `json:"completed_at"`
	DurationMs     string `json:"duration_ms"`
	RecordCount    string `json:"record_count"`
	SizeBytes      string `json:"size_bytes"`
	ReplicationLsn string `json:"replication_lsn"`
	HlcTimestamp   string `json:"hlc_timestamp"`
	StoragePath    string `json:"storage_path"`
	Checksum       string `json:"checksum"`
	TriggeredBy    string `json:"triggered_by"`
	ErrorMessage   string `json:"error_message"`
	Metadata       string `json:"metadata"`
}

// StringBackupSet represents backup set data as strings for TUI display.
type StringBackupSet struct {
	BackupSetID    string `json:"backup_set_id"`
	DateCreated    string `json:"date_created"`
	HlcTimestamp   string `json:"hlc_timestamp"`
	Status         string `json:"status"`
	BackupIds      string `json:"backup_ids"`
	NodeCount      string `json:"node_count"`
	CompletedCount string `json:"completed_count"`
	ErrorMessage   string `json:"error_message"`
}

// StringBackupVerification represents backup verification data as strings for TUI display.
type StringBackupVerification struct {
	VerificationID   string `json:"verification_id"`
	BackupID         string `json:"backup_id"`
	VerifiedAt       string `json:"verified_at"`
	VerifiedBy       string `json:"verified_by"`
	RestoreTested    string `json:"restore_tested"`
	ChecksumValid    string `json:"checksum_valid"`
	RecordCountMatch string `json:"record_count_match"`
	Status           string `json:"status"`
	ErrorMessage     string `json:"error_message"`
	DurationMs       string `json:"duration_ms"`
}

// StringChangeEvent represents change event data as strings for TUI display.
type StringChangeEvent struct {
	EventID       string `json:"event_id"`
	HlcTimestamp  string `json:"hlc_timestamp"`
	WallTimestamp string `json:"wall_timestamp"`
	NodeID        string `json:"node_id"`
	TableName     string `json:"table_name"`
	RecordID      string `json:"record_id"`
	Operation     string `json:"operation"`
	Action        string `json:"action"`
	UserID        string `json:"user_id"`
	OldValues     string `json:"old_values"`
	NewValues     string `json:"new_values"`
	Metadata      string `json:"metadata"`
	RequestID     string `json:"request_id"`
	IP            string `json:"ip"`
	SyncedAt      string `json:"synced_at"`
	ConsumedAt    string `json:"consumed_at"`
}

// StringRolePermissions represents role-permission association data as strings for TUI display.
type StringRolePermissions struct {
	ID           string `json:"id"`
	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`
}

// StringUserSshKeys represents user SSH key data as strings for TUI display.
type StringUserSshKeys struct {
	SshKeyID    string `json:"ssh_key_id"`
	UserID      string `json:"user_id"`
	PublicKey   string `json:"public_key"`
	KeyType     string `json:"key_type"`
	Fingerprint string `json:"fingerprint"`
	Label       string `json:"label"`
	DateCreated string `json:"date_created"`
	LastUsed    string `json:"last_used"`
}

// StringWebhookDelivery represents webhook delivery data as strings for TUI display.
type StringWebhookDelivery struct {
	DeliveryID     string `json:"delivery_id"`
	WebhookID      string `json:"webhook_id"`
	Event          string `json:"event"`
	Payload        string `json:"payload"`
	Status         string `json:"status"`
	Attempts       string `json:"attempts"`
	LastStatusCode string `json:"last_status_code"`
	LastError      string `json:"last_error"`
	NextRetryAt    string `json:"next_retry_at"`
	CreatedAt      string `json:"created_at"`
	CompletedAt    string `json:"completed_at"`
}

// StringPipeline represents pipeline data as strings for TUI display.
type StringPipeline struct {
	PipelineID   string `json:"pipeline_id"`
	PluginID     string `json:"plugin_id"`
	TableName    string `json:"table_name"`
	Operation    string `json:"operation"`
	PluginName   string `json:"plugin_name"`
	Handler      string `json:"handler"`
	Priority     string `json:"priority"`
	Enabled      string `json:"enabled"`
	Config       string `json:"config"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// StringFieldPluginConfig represents field plugin config data as strings for TUI display.
type StringFieldPluginConfig struct {
	FieldID         string `json:"field_id"`
	PluginName      string `json:"plugin_name"`
	PluginInterface string `json:"plugin_interface"`
	PluginVersion   string `json:"plugin_version"`
	DateCreated     string `json:"date_created"`
	DateModified    string `json:"date_modified"`
}

// MapString functions

// MapStringBackup converts Backup to StringBackup for table display.
func MapStringBackup(a Backup) StringBackup {
	return StringBackup{
		BackupID:       a.BackupID.String(),
		NodeID:         a.NodeID.String(),
		BackupType:     a.BackupType.String(),
		Status:         a.Status.String(),
		StartedAt:      a.StartedAt.String(),
		CompletedAt:    a.CompletedAt.String(),
		DurationMs:     a.DurationMs.String(),
		RecordCount:    a.RecordCount.String(),
		SizeBytes:      a.SizeBytes.String(),
		ReplicationLsn: a.ReplicationLsn.String,
		HlcTimestamp:   fmt.Sprintf("%d", a.HlcTimestamp),
		StoragePath:    a.StoragePath,
		Checksum:       a.Checksum.String,
		TriggeredBy:    a.TriggeredBy.String,
		ErrorMessage:   a.ErrorMessage.String,
		Metadata:       a.Metadata.String(),
	}
}

// MapStringBackupSet converts BackupSet to StringBackupSet for table display.
func MapStringBackupSet(a BackupSet) StringBackupSet {
	return StringBackupSet{
		BackupSetID:    a.BackupSetID.String(),
		DateCreated:    a.DateCreated.String(),
		HlcTimestamp:   fmt.Sprintf("%d", a.HlcTimestamp),
		Status:         a.Status.String(),
		BackupIds:      a.BackupIds.String(),
		NodeCount:      fmt.Sprintf("%d", a.NodeCount),
		CompletedCount: a.CompletedCount.String(),
		ErrorMessage:   a.ErrorMessage.String,
	}
}

// MapStringBackupVerification converts BackupVerification to StringBackupVerification for table display.
func MapStringBackupVerification(a BackupVerification) StringBackupVerification {
	return StringBackupVerification{
		VerificationID:   a.VerificationID.String(),
		BackupID:         a.BackupID.String(),
		VerifiedAt:       a.VerifiedAt.String(),
		VerifiedBy:       a.VerifiedBy.String,
		RestoreTested:    nullableBoolString(a.RestoreTested),
		ChecksumValid:    nullableBoolString(a.ChecksumValid),
		RecordCountMatch: nullableBoolString(a.RecordCountMatch),
		Status:           a.Status.String(),
		ErrorMessage:     a.ErrorMessage.String,
		DurationMs:       a.DurationMs.String(),
	}
}

// MapStringChangeEvent converts ChangeEvent to StringChangeEvent for table display.
func MapStringChangeEvent(a ChangeEvent) StringChangeEvent {
	return StringChangeEvent{
		EventID:       a.EventID.String(),
		HlcTimestamp:  fmt.Sprintf("%d", a.HlcTimestamp),
		WallTimestamp: a.WallTimestamp.String(),
		NodeID:        a.NodeID.String(),
		TableName:     a.TableName,
		RecordID:      a.RecordID,
		Operation:     a.Operation.String(),
		Action:        a.Action.String(),
		UserID:        a.UserID.String(),
		OldValues:     a.OldValues.String(),
		NewValues:     a.NewValues.String(),
		Metadata:      a.Metadata.String(),
		RequestID:     a.RequestID.String,
		IP:            a.IP.String,
		SyncedAt:      a.SyncedAt.String(),
		ConsumedAt:    a.ConsumedAt.String(),
	}
}

// MapStringRolePermission converts RolePermissions to StringRolePermissions for table display.
func MapStringRolePermission(a RolePermissions) StringRolePermissions {
	return StringRolePermissions{
		ID:           a.ID.String(),
		RoleID:       a.RoleID.String(),
		PermissionID: a.PermissionID.String(),
	}
}

// MapStringUserSshKey converts UserSshKeys to StringUserSshKeys for table display.
func MapStringUserSshKey(a UserSshKeys) StringUserSshKeys {
	return StringUserSshKeys{
		SshKeyID:    a.SshKeyID,
		UserID:      a.UserID.String(),
		PublicKey:   a.PublicKey,
		KeyType:     a.KeyType,
		Fingerprint: a.Fingerprint,
		Label:       a.Label,
		DateCreated: a.DateCreated.String(),
		LastUsed:    a.LastUsed,
	}
}

// MapStringWebhookDelivery converts WebhookDelivery to StringWebhookDelivery for table display.
func MapStringWebhookDelivery(a WebhookDelivery) StringWebhookDelivery {
	return StringWebhookDelivery{
		DeliveryID:     a.DeliveryID.String(),
		WebhookID:      a.WebhookID.String(),
		Event:          a.Event,
		Payload:        a.Payload,
		Status:         a.Status,
		Attempts:       fmt.Sprintf("%d", a.Attempts),
		LastStatusCode: fmt.Sprintf("%d", a.LastStatusCode),
		LastError:      a.LastError,
		NextRetryAt:    a.NextRetryAt,
		CreatedAt:      a.CreatedAt.String(),
		CompletedAt:    a.CompletedAt,
	}
}

// MapStringPipeline converts Pipeline to StringPipeline for table display.
func MapStringPipeline(a Pipeline) StringPipeline {
	return StringPipeline{
		PipelineID:   a.PipelineID.String(),
		PluginID:     a.PluginID.String(),
		TableName:    a.TableName,
		Operation:    a.Operation,
		PluginName:   a.PluginName,
		Handler:      a.Handler,
		Priority:     fmt.Sprintf("%d", a.Priority),
		Enabled:      fmt.Sprintf("%t", a.Enabled),
		Config:       a.Config.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// MapStringFieldPluginConfig converts FieldPluginConfig to StringFieldPluginConfig for table display.
func MapStringFieldPluginConfig(a FieldPluginConfig) StringFieldPluginConfig {
	return StringFieldPluginConfig{
		FieldID:         a.FieldID.String(),
		PluginName:      a.PluginName,
		PluginInterface: a.PluginInterface,
		PluginVersion:   a.PluginVersion,
		DateCreated:     a.DateCreated.String(),
		DateModified:    a.DateModified.String(),
	}
}
