# DB Package Struct Reference

All wrapper structs defined in `internal/db/`. These are database-agnostic types that abstract over the three sqlc-generated driver packages (SQLite, MySQL, PostgreSQL).

Types prefixed with `types.` come from `internal/db/types/`.

---

## Database Connections (db.go)

### Database
SQLite database connection.

| Field | Type |
|-------|------|
| Src | string |
| Status | DbStatus |
| Connection | *sql.DB |
| LastConnection | string |
| Err | error |
| Config | config.Config |
| Context | context.Context |

### MysqlDatabase
MySQL database connection. Same fields as Database.

### PsqlDatabase
PostgreSQL database connection. Same fields as Database.

---

## Entity Structs

### AdminContentData (admin_content_data.go)

| Field | Type |
|-------|------|
| AdminContentDataID | types.AdminContentID |
| ParentID | types.NullableContentID |
| FirstChildID | sql.NullString |
| NextSiblingID | sql.NullString |
| PrevSiblingID | sql.NullString |
| AdminRouteID | string |
| AdminDatatypeID | types.NullableAdminDatatypeID |
| AuthorID | types.NullableUserID |
| Status | types.ContentStatus |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### AdminContentFields (admin_content_field.go)

| Field | Type |
|-------|------|
| AdminContentFieldID | types.AdminContentFieldID |
| AdminRouteID | sql.NullString |
| AdminContentDataID | string |
| AdminFieldID | types.NullableAdminFieldID |
| AdminFieldValue | string |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### AdminContentRelations (admin_content_relation.go)

| Field | Type |
|-------|------|
| AdminContentRelationID | types.AdminContentRelationID |
| SourceContentID | types.AdminContentID |
| TargetContentID | types.AdminContentID |
| AdminFieldID | types.AdminFieldID |
| SortOrder | int64 |
| DateCreated | types.Timestamp |

### AdminDatatypes (admin_datatype.go)

| Field | Type |
|-------|------|
| AdminDatatypeID | types.AdminDatatypeID |
| ParentID | types.NullableContentID |
| Label | string |
| Type | string |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### AdminDatatypeFields (admin_datatype_field.go)

| Field | Type |
|-------|------|
| ID | string |
| AdminDatatypeID | types.AdminDatatypeID |
| AdminFieldID | types.AdminFieldID |

### AdminFields (admin_field.go)

| Field | Type |
|-------|------|
| AdminFieldID | types.AdminFieldID |
| ParentID | types.NullableAdminDatatypeID |
| Label | string |
| Data | string |
| Validation | string |
| UIConfig | string |
| Type | types.FieldType |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### AdminRoutes (admin_route.go)

| Field | Type |
|-------|------|
| AdminRouteID | types.AdminRouteID |
| Slug | types.Slug |
| Title | string |
| Status | int64 |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### Backup (backup.go)

| Field | Type |
|-------|------|
| BackupID | types.BackupID |
| NodeID | types.NodeID |
| BackupType | types.BackupType |
| Status | types.BackupStatus |
| StartedAt | types.Timestamp |
| CompletedAt | types.Timestamp |
| DurationMs | types.NullableInt64 |
| RecordCount | types.NullableInt64 |
| SizeBytes | types.NullableInt64 |
| ReplicationLsn | types.NullableString |
| HlcTimestamp | types.HLC |
| StoragePath | string |
| Checksum | types.NullableString |
| TriggeredBy | types.NullableString |
| ErrorMessage | types.NullableString |
| Metadata | types.JSONData |

### BackupSet (backup.go)

| Field | Type |
|-------|------|
| BackupSetID | types.BackupSetID |
| CreatedAt | types.Timestamp |
| HlcTimestamp | types.HLC |
| Status | types.BackupSetStatus |
| BackupIds | types.JSONData |
| NodeCount | int64 |
| CompletedCount | types.NullableInt64 |
| ErrorMessage | types.NullableString |

### BackupVerification (backup.go)

| Field | Type |
|-------|------|
| VerificationID | types.VerificationID |
| BackupID | types.BackupID |
| VerifiedAt | types.Timestamp |
| VerifiedBy | types.NullableString |
| RestoreTested | types.NullableBool |
| ChecksumValid | types.NullableBool |
| RecordCountMatch | types.NullableBool |
| Status | types.VerificationStatus |
| ErrorMessage | types.NullableString |
| DurationMs | types.NullableInt64 |

### ChangeEvent (change_event.go)

| Field | Type |
|-------|------|
| EventID | types.EventID |
| HlcTimestamp | types.HLC |
| WallTimestamp | types.Timestamp |
| NodeID | types.NodeID |
| TableName | string |
| RecordID | string |
| Operation | types.Operation |
| Action | types.Action |
| UserID | types.NullableUserID |
| OldValues | types.JSONData |
| NewValues | types.JSONData |
| Metadata | types.JSONData |
| RequestID | types.NullableString |
| IP | types.NullableString |
| SyncedAt | types.Timestamp |
| ConsumedAt | types.Timestamp |

### ContentData (content_data.go)

| Field | Type |
|-------|------|
| ContentDataID | types.ContentID |
| ParentID | types.NullableContentID |
| FirstChildID | sql.NullString |
| NextSiblingID | sql.NullString |
| PrevSiblingID | sql.NullString |
| RouteID | types.NullableRouteID |
| DatatypeID | types.NullableDatatypeID |
| AuthorID | types.NullableUserID |
| Status | types.ContentStatus |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### ContentFields (content_field.go)

| Field | Type |
|-------|------|
| ContentFieldID | types.ContentFieldID |
| RouteID | types.NullableRouteID |
| ContentDataID | types.NullableContentID |
| FieldID | types.NullableFieldID |
| FieldValue | string |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### ContentRelations (content_relation.go)

| Field | Type |
|-------|------|
| ContentRelationID | types.ContentRelationID |
| SourceContentID | types.ContentID |
| TargetContentID | types.ContentID |
| FieldID | types.FieldID |
| SortOrder | int64 |
| DateCreated | types.Timestamp |

### Datatypes (datatype.go)

| Field | Type |
|-------|------|
| DatatypeID | types.DatatypeID |
| ParentID | types.NullableContentID |
| Label | string |
| Type | string |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### DatatypeFields (datatype_field.go)

| Field | Type |
|-------|------|
| ID | string |
| DatatypeID | types.DatatypeID |
| FieldID | types.FieldID |
| SortOrder | int64 |

### Fields (field.go)

| Field | Type |
|-------|------|
| FieldID | types.FieldID |
| ParentID | types.NullableDatatypeID |
| Label | string |
| Data | string |
| Validation | string |
| UIConfig | string |
| Type | types.FieldType |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### Media (media.go)

| Field | Type |
|-------|------|
| MediaID | types.MediaID |
| Name | sql.NullString |
| DisplayName | sql.NullString |
| Alt | sql.NullString |
| Caption | sql.NullString |
| Description | sql.NullString |
| Class | sql.NullString |
| Mimetype | sql.NullString |
| Dimensions | sql.NullString |
| URL | types.URL |
| Srcset | sql.NullString |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### MediaDimensions (media_dimension.go)

| Field | Type |
|-------|------|
| MdID | string |
| Label | sql.NullString |
| Width | sql.NullInt64 |
| Height | sql.NullInt64 |
| AspectRatio | sql.NullString |

### Permissions (permission.go)

| Field | Type |
|-------|------|
| PermissionID | types.PermissionID |
| TableID | string |
| Mode | int64 |
| Label | string |

### Roles (role.go)

| Field | Type |
|-------|------|
| RoleID | types.RoleID |
| Label | string |
| Permissions | string |

### Routes (route.go)

| Field | Type |
|-------|------|
| RouteID | types.RouteID |
| Slug | types.Slug |
| Title | string |
| Status | int64 |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### Sessions (session.go)

| Field | Type |
|-------|------|
| SessionID | types.SessionID |
| UserID | types.NullableUserID |
| CreatedAt | types.Timestamp |
| ExpiresAt | types.Timestamp |
| LastAccess | sql.NullString |
| IpAddress | sql.NullString |
| UserAgent | sql.NullString |
| SessionData | sql.NullString |

### Tables (table.go)

| Field | Type |
|-------|------|
| ID | string |
| Label | string |
| AuthorID | types.NullableUserID |

### Tokens (token.go)

| Field | Type |
|-------|------|
| ID | string |
| UserID | types.NullableUserID |
| TokenType | string |
| Token | string |
| IssuedAt | string |
| ExpiresAt | types.Timestamp |
| Revoked | bool |

### Users (user.go)

| Field | Type |
|-------|------|
| UserID | types.UserID |
| Username | string |
| Name | string |
| Email | types.Email |
| Hash | string |
| Role | string |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### UserOauth (user_oauth.go)

| Field | Type |
|-------|------|
| UserOauthID | types.UserOauthID |
| UserID | types.NullableUserID |
| OauthProvider | string |
| OauthProviderUserID | string |
| AccessToken | string |
| RefreshToken | string |
| TokenExpiresAt | string |
| DateCreated | types.Timestamp |

### UserSshKeys (user_ssh_keys.go)

| Field | Type |
|-------|------|
| SshKeyID | string |
| UserID | types.NullableUserID |
| PublicKey | string |
| KeyType | string |
| Fingerprint | string |
| Label | string |
| DateCreated | types.Timestamp |
| LastUsed | string |

---

## Create Params

### CreateAdminContentDataParams

| Field | Type |
|-------|------|
| ParentID | types.NullableContentID |
| FirstChildID | sql.NullString |
| NextSiblingID | sql.NullString |
| PrevSiblingID | sql.NullString |
| AdminRouteID | string |
| AdminDatatypeID | types.NullableAdminDatatypeID |
| AuthorID | types.NullableUserID |
| Status | types.ContentStatus |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### CreateAdminContentFieldParams

| Field | Type |
|-------|------|
| AdminRouteID | sql.NullString |
| AdminContentDataID | string |
| AdminFieldID | types.NullableAdminFieldID |
| AdminFieldValue | string |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### CreateAdminContentRelationParams

| Field | Type |
|-------|------|
| SourceContentID | types.AdminContentID |
| TargetContentID | types.AdminContentID |
| AdminFieldID | types.AdminFieldID |
| SortOrder | int64 |
| DateCreated | types.Timestamp |

### CreateAdminDatatypeParams

| Field | Type |
|-------|------|
| ParentID | types.NullableContentID |
| Label | string |
| Type | string |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### CreateAdminDatatypeFieldParams

| Field | Type |
|-------|------|
| AdminDatatypeID | types.AdminDatatypeID |
| AdminFieldID | types.AdminFieldID |

### CreateAdminFieldParams

| Field | Type |
|-------|------|
| ParentID | types.NullableAdminDatatypeID |
| Label | string |
| Data | string |
| Validation | string |
| UIConfig | string |
| Type | types.FieldType |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### CreateAdminRouteParams

| Field | Type |
|-------|------|
| Slug | types.Slug |
| Title | string |
| Status | int64 |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### CreateBackupParams

| Field | Type |
|-------|------|
| BackupID | types.BackupID |
| NodeID | types.NodeID |
| BackupType | types.BackupType |
| Status | types.BackupStatus |
| StartedAt | types.Timestamp |
| StoragePath | string |
| TriggeredBy | types.NullableString |
| Metadata | types.JSONData |

### CreateBackupSetParams

| Field | Type |
|-------|------|
| BackupSetID | types.BackupSetID |
| CreatedAt | types.Timestamp |
| HlcTimestamp | types.HLC |
| Status | types.BackupSetStatus |
| BackupIds | types.JSONData |
| NodeCount | int64 |
| CompletedCount | types.NullableInt64 |
| ErrorMessage | types.NullableString |

### CreateVerificationParams

| Field | Type |
|-------|------|
| VerificationID | types.VerificationID |
| BackupID | types.BackupID |
| VerifiedAt | types.Timestamp |
| VerifiedBy | types.NullableString |
| RestoreTested | types.NullableBool |
| ChecksumValid | types.NullableBool |
| RecordCountMatch | types.NullableBool |
| Status | types.VerificationStatus |
| ErrorMessage | types.NullableString |
| DurationMs | types.NullableInt64 |

### CreateContentDataParams

| Field | Type |
|-------|------|
| RouteID | types.NullableRouteID |
| ParentID | types.NullableContentID |
| FirstChildID | sql.NullString |
| NextSiblingID | sql.NullString |
| PrevSiblingID | sql.NullString |
| DatatypeID | types.NullableDatatypeID |
| AuthorID | types.NullableUserID |
| Status | types.ContentStatus |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### CreateContentFieldParams

| Field | Type |
|-------|------|
| RouteID | types.NullableRouteID |
| ContentDataID | types.NullableContentID |
| FieldID | types.NullableFieldID |
| FieldValue | string |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### CreateContentRelationParams

| Field | Type |
|-------|------|
| SourceContentID | types.ContentID |
| TargetContentID | types.ContentID |
| FieldID | types.FieldID |
| SortOrder | int64 |
| DateCreated | types.Timestamp |

### CreateDatatypeParams

| Field | Type |
|-------|------|
| DatatypeID | types.DatatypeID |
| ParentID | types.NullableContentID |
| Label | string |
| Type | string |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### CreateDatatypeFieldParams

| Field | Type |
|-------|------|
| ID | string |
| DatatypeID | types.DatatypeID |
| FieldID | types.FieldID |
| SortOrder | int64 |

### CreateFieldParams

| Field | Type |
|-------|------|
| FieldID | types.FieldID |
| ParentID | types.NullableDatatypeID |
| Label | string |
| Data | string |
| Validation | string |
| UIConfig | string |
| Type | types.FieldType |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### CreateMediaParams

| Field | Type |
|-------|------|
| Name | sql.NullString |
| DisplayName | sql.NullString |
| Alt | sql.NullString |
| Caption | sql.NullString |
| Description | sql.NullString |
| Class | sql.NullString |
| Mimetype | sql.NullString |
| Dimensions | sql.NullString |
| URL | types.URL |
| Srcset | sql.NullString |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### CreateMediaDimensionParams

| Field | Type |
|-------|------|
| Label | sql.NullString |
| Width | sql.NullInt64 |
| Height | sql.NullInt64 |
| AspectRatio | sql.NullString |

### CreatePermissionParams

| Field | Type |
|-------|------|
| TableID | string |
| Mode | int64 |
| Label | string |

### CreateRoleParams

| Field | Type |
|-------|------|
| Label | string |
| Permissions | string |

### CreateRouteParams

| Field | Type |
|-------|------|
| RouteID | types.RouteID |
| Slug | types.Slug |
| Title | string |
| Status | int64 |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### CreateSessionParams

| Field | Type |
|-------|------|
| UserID | types.NullableUserID |
| CreatedAt | types.Timestamp |
| ExpiresAt | types.Timestamp |
| LastAccess | sql.NullString |
| IpAddress | sql.NullString |
| UserAgent | sql.NullString |
| SessionData | sql.NullString |

### CreateTableParams

| Field | Type |
|-------|------|
| Label | string |

### CreateTokenParams

| Field | Type |
|-------|------|
| UserID | types.NullableUserID |
| TokenType | string |
| Token | string |
| IssuedAt | string |
| ExpiresAt | types.Timestamp |
| Revoked | bool |

### CreateUserParams

| Field | Type |
|-------|------|
| Username | string |
| Name | string |
| Email | types.Email |
| Hash | string |
| Role | string |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### CreateUserOauthParams

| Field | Type |
|-------|------|
| UserID | types.NullableUserID |
| OauthProvider | string |
| OauthProviderUserID | string |
| AccessToken | string |
| RefreshToken | string |
| TokenExpiresAt | string |
| DateCreated | types.Timestamp |

### CreateUserSshKeyParams

| Field | Type |
|-------|------|
| UserID | types.NullableUserID |
| PublicKey | string |
| KeyType | string |
| Fingerprint | string |
| Label | string |
| DateCreated | types.Timestamp |

### RecordChangeEventParams (change_event.go)

| Field | Type |
|-------|------|
| EventID | types.EventID |
| HlcTimestamp | types.HLC |
| NodeID | types.NodeID |
| TableName | string |
| RecordID | string |
| Operation | types.Operation |
| Action | types.Action |
| UserID | types.NullableUserID |
| OldValues | types.JSONData |
| NewValues | types.JSONData |
| Metadata | types.JSONData |
| RequestID | types.NullableString |
| IP | types.NullableString |

---

## Update Params

### UpdateAdminContentDataParams
Same as AdminContentData entity fields. ID field: `AdminContentDataID`.

### UpdateAdminContentFieldParams
Same as AdminContentFields entity fields. ID field: `AdminContentFieldID`.

### UpdateAdminContentRelationSortOrderParams

| Field | Type |
|-------|------|
| AdminContentRelationID | types.AdminContentRelationID |
| SortOrder | int64 |

### UpdateAdminDatatypeParams
Same as AdminDatatypes entity fields. ID field: `AdminDatatypeID`.

### UpdateAdminDatatypeFieldParams

| Field | Type |
|-------|------|
| AdminDatatypeID | types.AdminDatatypeID |
| AdminFieldID | types.AdminFieldID |
| ID | string |

### UpdateAdminFieldParams
Same as AdminFields entity fields. ID field: `AdminFieldID`.

### UpdateAdminRouteParams
Same as AdminRoutes entity fields (minus AdminRouteID). Uses `Slug_2 types.Slug` as WHERE clause identifier.

### UpdateBackupStatusParams

| Field | Type |
|-------|------|
| Status | types.BackupStatus |
| CompletedAt | types.Timestamp |
| DurationMs | types.NullableInt64 |
| RecordCount | types.NullableInt64 |
| SizeBytes | types.NullableInt64 |
| Checksum | types.NullableString |
| ErrorMessage | types.NullableString |
| BackupID | types.BackupID |

### UpdateContentDataParams
Same as ContentData entity fields. ID field: `ContentDataID`.

### UpdateContentFieldParams
Same as ContentFields entity fields. ID field: `ContentFieldID`.

### UpdateContentRelationSortOrderParams

| Field | Type |
|-------|------|
| ContentRelationID | types.ContentRelationID |
| SortOrder | int64 |

### UpdateDatatypeParams
Same as Datatypes entity fields. ID field: `DatatypeID`.

### UpdateDatatypeFieldParams

| Field | Type |
|-------|------|
| DatatypeID | types.DatatypeID |
| FieldID | types.FieldID |
| SortOrder | int64 |
| ID | string |

### UpdateFieldParams
Same as Fields entity fields. ID field: `FieldID`.

### UpdateMediaParams
Same as Media entity fields. ID field: `MediaID`.

### UpdateMediaDimensionParams
Same as MediaDimensions entity fields. ID field: `MdID`.

### UpdatePermissionParams
Same as Permissions entity fields. ID field: `PermissionID`.

### UpdateRoleParams
Same as Roles entity fields. ID field: `RoleID`.

### UpdateRouteParams
Same as Routes entity fields (minus RouteID). Uses `Slug_2 types.Slug` as WHERE clause identifier.

### UpdateSessionParams
Same as Sessions entity fields. ID field: `SessionID`.

### UpdateTableParams

| Field | Type |
|-------|------|
| Label | string |
| ID | string |

### UpdateTokenParams

| Field | Type |
|-------|------|
| Token | string |
| IssuedAt | string |
| ExpiresAt | types.Timestamp |
| Revoked | bool |
| ID | string |

### UpdateUserParams
Same as Users entity fields. ID field: `UserID`.

### UpdateUserOauthParams

| Field | Type |
|-------|------|
| AccessToken | string |
| RefreshToken | string |
| TokenExpiresAt | string |
| UserOauthID | types.UserOauthID |

---

## Query Result Row Structs

### ListAdminDatatypeByRouteIdRow (admin_datatype.go)

| Field | Type |
|-------|------|
| AdminDatatypeID | types.AdminDatatypeID |
| AdminRouteID | types.NullableRouteID |
| ParentID | types.NullableContentID |
| Label | string |
| Type | string |

### UtilityGetAdminDatatypesRow (admin_datatype.go)

| Field | Type |
|-------|------|
| AdminDatatypeID | types.AdminDatatypeID |
| Label | string |

### ListAdminFieldByRouteIdRow (admin_field.go)

| Field | Type |
|-------|------|
| AdminFieldID | types.AdminFieldID |
| ParentID | types.NullableAdminDatatypeID |
| Label | string |
| Data | string |
| Validation | string |
| UIConfig | string |
| Type | types.FieldType |

### ListAdminFieldsByDatatypeIDRow (admin_field.go)

| Field | Type |
|-------|------|
| AdminFieldID | types.AdminFieldID |
| ParentID | types.NullableAdminDatatypeID |
| Label | string |
| Data | string |
| Validation | string |
| UIConfig | string |
| Type | types.FieldType |

### UtilityGetAdminfieldsRow (admin_field.go)

| Field | Type |
|-------|------|
| AdminFieldID | types.AdminFieldID |
| Label | string |

### UtilityGetAdminRoutesRow (admin_route.go)

| Field | Type |
|-------|------|
| AdminRouteID | types.AdminRouteID |
| Slug | types.Slug |

### RootContentSummary (content_data.go)

| Field | Type |
|-------|------|
| ContentDataID | types.ContentID |
| RouteID | types.NullableRouteID |
| DatatypeID | types.NullableDatatypeID |
| RouteSlug | types.Slug |
| RouteTitle | string |
| DatatypeLabel | string |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |

### GetRouteTreeByRouteIDRow (getTree.go)

| Field | Type |
|-------|------|
| ContentDataID | types.ContentID |
| ParentID | types.NullableContentID |
| FirstChildID | sql.NullString |
| NextSiblingID | sql.NullString |
| PrevSiblingID | sql.NullString |
| DatatypeLabel | string |
| DatatypeType | string |
| FieldLabel | string |
| FieldType | types.FieldType |
| FieldValue | sql.NullString |

### GetContentTreeByRouteRow (getTree.go)

| Field | Type |
|-------|------|
| ContentDataID | types.ContentID |
| ParentID | types.NullableContentID |
| FirstChildID | sql.NullString |
| NextSiblingID | sql.NullString |
| PrevSiblingID | sql.NullString |
| DatatypeID | types.NullableDatatypeID |
| RouteID | types.NullableRouteID |
| AuthorID | types.NullableUserID |
| DateCreated | types.Timestamp |
| DateModified | types.Timestamp |
| Status | types.ContentStatus |
| DatatypeLabel | string |
| DatatypeType | string |

### GetContentFieldsByRouteRow (getTree.go)

| Field | Type |
|-------|------|
| ContentDataID | types.NullableContentID |
| FieldID | types.NullableFieldID |
| FieldValue | string |

### GetFieldDefinitionsByRouteRow (getTree.go)

| Field | Type |
|-------|------|
| FieldID | types.FieldID |
| Label | string |
| Type | types.FieldType |
| DatatypeID | types.DatatypeID |

---

## List/Query Params

### ListBackupsParams (backup.go)

| Field | Type |
|-------|------|
| Limit | int64 |
| Offset | int64 |

### ListChangeEventsParams (change_event.go)

| Field | Type |
|-------|------|
| Limit | int64 |
| Offset | int64 |

### ListChangeEventsByUserParams (change_event.go)

| Field | Type |
|-------|------|
| UserID | types.NullableUserID |
| Limit | int64 |
| Offset | int64 |

### ListChangeEventsByActionParams (change_event.go)

| Field | Type |
|-------|------|
| Action | types.Action |
| Limit | int64 |
| Offset | int64 |

---

## Media Dimension Form Params

### CreateMediaDimensionFormParams (media_dimension.go)
All-string variant for form input.

| Field | Type |
|-------|------|
| Label | string |
| Width | string |
| Height | string |
| AspectRatio | string |

### UpdateMediaDimensionFormParams (media_dimension.go)
All-string variant for form input.

| Field | Type |
|-------|------|
| Label | string |
| Width | string |
| Height | string |
| AspectRatio | string |
| MdID | string |

### MediaDimensionsHistoryEntry (media_dimension.go)
Same fields as MediaDimensions.
