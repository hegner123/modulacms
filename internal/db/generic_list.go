package db

import "fmt"

// GenericHeaders returns the column header names for the given table.
func GenericHeaders(t DBTable) []string {
	switch t {
	case Admin_content_data:
		return []string{
			"admin_content_data_id",
			"parent_id",
			"first_child_id",
			"next_sibling_id",
			"prev_sibling_id",
			"root_id",
			"admin_route_id",
			"admin_datatype_id",
			"author_id",
			"status",
			"date_created",
			"date_modified",
			"published_at",
			"published_by",
			"publish_at",
			"revision",
			"history",
		}
	case Admin_content_fields:
		return []string{
			"admin_content_field_id",
			"admin_route_id",
			"root_id",
			"admin_content_data_id",
			"admin_field_id",
			"admin_field_value",
			"locale",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Admin_content_relations:
		return []string{
			"admin_content_relation_id",
			"source_content_id",
			"target_content_id",
			"admin_field_id",
			"sort_order",
			"date_created",
		}
	case Admin_content_versions:
		return []string{
			"admin_content_version_id",
			"admin_content_data_id",
			"version_number",
			"locale",
			"snapshot",
			"trigger",
			"label",
			"published",
			"published_by",
			"date_created",
		}
	case Admin_datatype:
		return []string{
			"admin_datatype_id",
			"parent_id",
			"sort_order",
			"name",
			"label",
			"type",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Admin_field:
		return []string{
			"admin_field_id",
			"parent_id",
			"sort_order",
			"name",
			"label",
			"data",
			"validation",
			"ui_config",
			"type",
			"translatable",
			"roles",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Admin_field_types:
		return []string{
			"admin_field_type_id",
			"type",
			"label",
		}
	case Admin_route:
		return []string{
			"admin_route_id",
			"slug",
			"title",
			"status",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case BackupT:
		return []string{
			"backup_id",
			"node_id",
			"backup_type",
			"status",
			"started_at",
			"completed_at",
			"duration_ms",
			"record_count",
			"size_bytes",
			"replication_lsn",
			"hlc_timestamp",
			"storage_path",
			"checksum",
			"triggered_by",
			"error_message",
			"metadata",
		}
	case Backup_set:
		return []string{
			"backup_set_id",
			"date_created",
			"hlc_timestamp",
			"status",
			"backup_ids",
			"node_count",
			"completed_count",
			"error_message",
		}
	case Backup_verification:
		return []string{
			"verification_id",
			"backup_id",
			"verified_at",
			"verified_by",
			"restore_tested",
			"checksum_valid",
			"record_count_match",
			"status",
			"error_message",
			"duration_ms",
		}
	case Change_event:
		return []string{
			"event_id",
			"hlc_timestamp",
			"wall_timestamp",
			"node_id",
			"table_name",
			"record_id",
			"operation",
			"action",
			"user_id",
			"old_values",
			"new_values",
			"metadata",
			"request_id",
			"ip",
			"synced_at",
			"consumed_at",
		}
	case Content_data:
		return []string{
			"content_data_id",
			"parent_id",
			"first_child_id",
			"next_sibling_id",
			"prev_sibling_id",
			"root_id",
			"route_id",
			"datatype_id",
			"author_id",
			"status",
			"date_created",
			"date_modified",
			"published_at",
			"published_by",
			"publish_at",
			"revision",
			"history",
		}
	case Content_fields:
		return []string{
			"content_field_id",
			"route_id",
			"root_id",
			"content_data_id",
			"field_id",
			"field_value",
			"locale",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Content_relations:
		return []string{
			"content_relation_id",
			"source_content_id",
			"target_content_id",
			"field_id",
			"sort_order",
			"date_created",
		}
	case Content_versions:
		return []string{
			"content_version_id",
			"content_data_id",
			"version_number",
			"locale",
			"snapshot",
			"trigger",
			"label",
			"published",
			"published_by",
			"date_created",
		}
	case Datatype:
		return []string{
			"datatype_id",
			"parent_id",
			"sort_order",
			"name",
			"label",
			"type",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Field:
		return []string{
			"field_id",
			"parent_id",
			"sort_order",
			"name",
			"label",
			"data",
			"validation",
			"ui_config",
			"type",
			"translatable",
			"roles",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Field_plugin_config:
		return []string{
			"field_id",
			"plugin_name",
			"plugin_interface",
			"plugin_version",
			"date_created",
			"date_modified",
		}
	case Field_types:
		return []string{
			"field_type_id",
			"type",
			"label",
		}
	case LocaleT:
		return []string{
			"locale_id",
			"code",
			"label",
			"is_default",
			"is_enabled",
			"fallback_code",
			"sort_order",
			"date_created",
		}
	case MediaT:
		return []string{
			"media_id",
			"name",
			"display_name",
			"alt",
			"caption",
			"description",
			"class",
			"mimetype",
			"dimensions",
			"url",
			"srcset",
			"focal_x",
			"focal_y",
			"author_id",
			"folder_id",
			"date_created",
			"date_modified",
		}
	case Media_folder:
		return []string{
			"folder_id",
			"name",
			"parent_id",
			"date_created",
			"date_modified",
		}
	case Media_dimension:
		return []string{
			"md_id",
			"label",
			"width",
			"height",
			"aspect_ratio",
		}
	case Permission:
		return []string{
			"permission_id",
			"label",
		}
	case PipelineT:
		return []string{
			"pipeline_id",
			"plugin_id",
			"table_name",
			"operation",
			"plugin_name",
			"handler",
			"priority",
			"enabled",
			"config",
			"date_created",
			"date_modified",
		}
	case Role:
		return []string{
			"role_id",
			"label",
			"system_protected",
		}
	case Role_permissions:
		return []string{
			"id",
			"role_id",
			"permission_id",
		}
	case Route:
		return []string{
			"route_id",
			"slug",
			"title",
			"status",
			"author_id",
			"date_created",
			"date_modified",
			"history",
		}
	case Session:
		return []string{
			"session_id",
			"user_id",
			"date_created",
			"expires_at",
			"last_access",
			"ip_address",
			"user_agent",
			"session_data",
		}
	case Table:
		return []string{
			"id",
			"label",
			"author_id",
		}
	case Token:
		return []string{
			"id",
			"user_id",
			"token_type",
			"token",
			"issued_at",
			"expires_at",
			"revoked",
		}
	case User:
		return []string{
			"user_id",
			"username",
			"name",
			"email",
			"hash",
			"role",
			"date_created",
			"date_modified",
		}
	case User_oauth:
		return []string{
			"user_oauth_id",
			"user_id",
			"oauth_provider",
			"oauth_provider_user_id",
			"access_token",
			"refresh_token",
			"token_expires_at",
			"date_created",
		}
	case User_ssh_keys:
		return []string{
			"ssh_key_id",
			"user_id",
			"public_key",
			"key_type",
			"fingerprint",
			"label",
			"date_created",
			"last_used",
		}
	case WebhookT:
		return []string{
			"webhook_id",
			"name",
			"url",
			"is_active",
			"author_id",
			"date_created",
			"date_modified",
		}
	case Webhook_deliveries:
		return []string{
			"delivery_id",
			"webhook_id",
			"event",
			"payload",
			"status",
			"attempts",
			"last_status_code",
			"last_error",
			"next_retry_at",
			"created_at",
			"completed_at",
		}
	}
	return nil
}

// GenericList retrieves all rows from the given table as string slices for TUI display.
func GenericList(t DBTable, d DbDriver) ([][]string, error) {
	switch t {
	case Admin_content_data:
		a, err := d.ListAdminContentData()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminContentData(row)
			r := []string{
				s.AdminContentDataID,
				s.ParentID,
				s.FirstChildID,
				s.NextSiblingID,
				s.PrevSiblingID,
				s.RootID,
				s.AdminRouteID,
				s.AdminDatatypeID,
				s.AuthorID,
				s.Status,
				s.DateCreated,
				s.DateModified,
				s.PublishedAt,
				s.PublishedBy,
				s.PublishAt,
				s.Revision,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Admin_content_fields:
		a, err := d.ListAdminContentFields()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminContentField(row)
			r := []string{
				s.AdminContentFieldID,
				s.AdminRouteID,
				s.RootID,
				s.AdminContentDataID,
				s.AdminFieldID,
				s.AdminFieldValue,
				s.Locale,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Admin_content_relations:
		a, err := d.ListAdminContentRelations()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminContentRelation(row)
			r := []string{
				s.AdminContentRelationID,
				s.SourceContentID,
				s.TargetContentID,
				s.AdminFieldID,
				s.SortOrder,
				s.DateCreated,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Admin_content_versions:
		// No parameterless ListAdminContentVersions method exists;
		// versions are queried by content data ID.
		return nil, fmt.Errorf("table %q requires content data ID parameter for listing", t)
	case Admin_datatype:
		a, err := d.ListAdminDatatypes()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminDatatype(row)
			r := []string{
				s.AdminDatatypeID,
				s.ParentID,
				s.SortOrder,
				s.Name,
				s.Label,
				s.Type,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Admin_field:
		a, err := d.ListAdminFields()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminField(row)
			r := []string{
				s.AdminFieldID,
				s.ParentID,
				s.SortOrder,
				s.Name,
				s.Label,
				s.Data,
				s.Validation,
				s.UIConfig,
				s.Type,
				s.Translatable,
				s.Roles,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Admin_field_types:
		a, err := d.ListAdminFieldTypes()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminFieldType(row)
			r := []string{
				s.AdminFieldTypeID,
				s.Type,
				s.Label,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Admin_route:
		a, err := d.ListAdminRoutes()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringAdminRoute(row)
			r := []string{
				s.AdminRouteID,
				s.Slug,
				s.Title,
				s.Status,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case BackupT:
		a, err := d.ListBackups(ListBackupsParams{Limit: 1000, Offset: 0})
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringBackup(row)
			r := []string{
				s.BackupID,
				s.NodeID,
				s.BackupType,
				s.Status,
				s.StartedAt,
				s.CompletedAt,
				s.DurationMs,
				s.RecordCount,
				s.SizeBytes,
				s.ReplicationLsn,
				s.HlcTimestamp,
				s.StoragePath,
				s.Checksum,
				s.TriggeredBy,
				s.ErrorMessage,
				s.Metadata,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Backup_set:
		// No parameterless ListBackupSets method exists.
		return nil, fmt.Errorf("table %q has no parameterless list method", t)
	case Backup_verification:
		// No parameterless ListBackupVerifications method exists.
		return nil, fmt.Errorf("table %q has no parameterless list method", t)
	case Change_event:
		a, err := d.ListChangeEvents(ListChangeEventsParams{Limit: 1000, Offset: 0})
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringChangeEvent(row)
			r := []string{
				s.EventID,
				s.HlcTimestamp,
				s.WallTimestamp,
				s.NodeID,
				s.TableName,
				s.RecordID,
				s.Operation,
				s.Action,
				s.UserID,
				s.OldValues,
				s.NewValues,
				s.Metadata,
				s.RequestID,
				s.IP,
				s.SyncedAt,
				s.ConsumedAt,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Content_data:
		a, err := d.ListContentData()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringContentData(row)
			r := []string{
				s.ContentDataID,
				s.ParentID,
				s.FirstChildID,
				s.NextSiblingID,
				s.PrevSiblingID,
				s.RootID,
				s.RouteID,
				s.DatatypeID,
				s.AuthorID,
				s.Status,
				s.DateCreated,
				s.DateModified,
				s.PublishedAt,
				s.PublishedBy,
				s.PublishAt,
				s.Revision,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Content_fields:
		a, err := d.ListContentFields()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringContentField(row)
			r := []string{
				s.ContentFieldID,
				s.RouteID,
				s.RootID,
				s.ContentDataID,
				s.FieldID,
				s.FieldValue,
				s.Locale,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Content_relations:
		a, err := d.ListContentRelations()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringContentRelation(row)
			r := []string{
				s.ContentRelationID,
				s.SourceContentID,
				s.TargetContentID,
				s.FieldID,
				s.SortOrder,
				s.DateCreated,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Content_versions:
		// No parameterless ListContentVersions method exists;
		// versions are queried by content data ID.
		return nil, fmt.Errorf("table %q requires content data ID parameter for listing", t)
	case Datatype:
		a, err := d.ListDatatypes()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringDatatype(row)
			r := []string{
				s.DatatypeID,
				s.ParentID,
				s.SortOrder,
				s.Name,
				s.Label,
				s.Type,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Field:
		a, err := d.ListFields()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringField(row)
			r := []string{
				s.FieldID,
				s.ParentID,
				s.SortOrder,
				s.Name,
				s.Label,
				s.Data,
				s.Validation,
				s.UIConfig,
				s.Type,
				s.Translatable,
				s.Roles,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Field_plugin_config:
		// No parameterless ListFieldPluginConfig method exists.
		return nil, fmt.Errorf("table %q has no parameterless list method", t)
	case Field_types:
		a, err := d.ListFieldTypes()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringFieldType(row)
			r := []string{
				s.FieldTypeID,
				s.Type,
				s.Label,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case LocaleT:
		a, err := d.ListLocales()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringLocale(row)
			r := []string{
				s.LocaleID,
				s.Code,
				s.Label,
				s.IsDefault,
				s.IsEnabled,
				s.FallbackCode,
				s.SortOrder,
				s.DateCreated,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case MediaT:
		a, err := d.ListMedia()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringMedia(row)
			r := []string{
				s.MediaID,
				s.Name,
				s.DisplayName,
				s.Alt,
				s.Caption,
				s.Description,
				s.Class,
				s.Mimetype,
				s.Dimensions,
				s.URL,
				s.Srcset,
				s.FocalX,
				s.FocalY,
				s.AuthorID,
				s.FolderID,
				s.DateCreated,
				s.DateModified,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Media_folder:
		a, err := d.ListMediaFolders()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringMediaFolder(row)
			r := []string{
				s.FolderID,
				s.Name,
				s.ParentID,
				s.DateCreated,
				s.DateModified,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Media_dimension:
		a, err := d.ListMediaDimensions()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringMediaDimension(row)
			r := []string{
				s.MdID,
				s.Label,
				s.Width,
				s.Height,
				s.AspectRatio,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Permission:
		a, err := d.ListPermissions()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringPermission(row)
			r := []string{
				s.PermissionID,
				s.Label,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case PipelineT:
		a, err := d.ListPipelines()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringPipeline(row)
			r := []string{
				s.PipelineID,
				s.PluginID,
				s.TableName,
				s.Operation,
				s.PluginName,
				s.Handler,
				s.Priority,
				s.Enabled,
				s.Config,
				s.DateCreated,
				s.DateModified,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Role:
		a, err := d.ListRoles()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringRole(row)
			r := []string{
				s.RoleID,
				s.Label,
				s.SystemProtected,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Role_permissions:
		a, err := d.ListRolePermissions()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringRolePermission(row)
			r := []string{
				s.ID,
				s.RoleID,
				s.PermissionID,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Route:
		a, err := d.ListRoutes()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringRoute(row)
			r := []string{
				s.RouteID,
				s.Slug,
				s.Title,
				s.Status,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
				s.History,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Session:
		a, err := d.ListSessions()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringSession(row)
			r := []string{
				s.SessionID,
				s.UserID,
				s.DateCreated,
				s.ExpiresAt,
				s.LastAccess,
				s.IpAddress,
				s.UserAgent,
				s.SessionData,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Table:
		a, err := d.ListTables()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringTable(row)
			r := []string{
				s.ID,
				s.Label,
				s.AuthorID,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Token:
		a, err := d.ListTokens()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringToken(row)
			r := []string{
				s.ID,
				s.UserID,
				s.TokenType,
				s.Token,
				s.IssuedAt,
				s.ExpiresAt,
				s.Revoked,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case User:
		a, err := d.ListUsers()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringUser(row)
			r := []string{
				s.UserID,
				s.Username,
				s.Name,
				s.Email,
				s.Hash,
				s.Role,
				s.DateCreated,
				s.DateModified,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case User_oauth:
		a, err := d.ListUserOauths()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringUserOauth(row)
			r := []string{
				s.UserOauthID,
				s.UserID,
				s.OauthProvider,
				s.OauthProviderUserID,
				s.AccessToken,
				s.RefreshToken,
				s.TokenExpiresAt,
				s.DateCreated,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case User_ssh_keys:
		// ListUserSshKeys requires a user ID parameter; list all not supported.
		return nil, fmt.Errorf("table %q requires user ID parameter for listing", t)
	case WebhookT:
		a, err := d.ListWebhooks()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringWebhook(row)
			r := []string{
				s.WebhookID,
				s.Name,
				s.URL,
				s.IsActive,
				s.AuthorID,
				s.DateCreated,
				s.DateModified,
			}
			collection = append(collection, r)
		}
		return collection, nil
	case Webhook_deliveries:
		a, err := d.ListWebhookDeliveries()
		if err != nil {
			return nil, err
		}
		var collection [][]string
		for i := range len(*a) {
			rows := *a
			row := rows[i]
			s := MapStringWebhookDelivery(row)
			r := []string{
				s.DeliveryID,
				s.WebhookID,
				s.Event,
				s.Payload,
				s.Status,
				s.Attempts,
				s.LastStatusCode,
				s.LastError,
				s.NextRetryAt,
				s.CreatedAt,
				s.CompletedAt,
			}
			collection = append(collection, r)
		}
		return collection, nil
	}
	return nil, nil
}
