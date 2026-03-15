# Reference

Quick reference for all resources available on the `*modula.Client` struct. Resources are grouped by category. Generic `Resource[E,C,U,ID]` types provide `List`, `Get`, `Create`, `Update`, `Delete`, `ListPaginated`, `Count`, and `RawList` methods unless noted otherwise.

## Content Resources

| Resource | Type | Key Methods | Description |
|----------|------|-------------|-------------|
| `ContentData` | `*Resource[ContentData, CreateContentDataParams, UpdateContentDataParams, ContentID]` | CRUD, ListPaginated, Count | Content data nodes in the published content tree. |
| `ContentFields` | `*Resource[ContentField, CreateContentFieldParams, UpdateContentFieldParams, ContentFieldID]` | CRUD, ListPaginated, Count | Field values attached to content data nodes. |
| `ContentRelations` | `*Resource[ContentRelation, CreateContentRelationParams, UpdateContentRelationParams, ContentRelationID]` | CRUD, ListPaginated, Count | Relations between content data nodes. |
| `ContentTree` | `*ContentTreeResource` | Save, GetByRoute | Bulk tree operations: create, update, delete nodes atomically. |
| `ContentReorder` | `*ContentReorderResource` | Reorder, Move | Reorder children within a parent or move nodes between parents. |
| `ContentBatch` | `*ContentBatchResource` | Update | Batch content updates in a single request. |
| `ContentHeal` | `*ContentHealResource` | Heal | Detect and repair broken sibling pointers in content trees. |
| `ContentVersions` | `*ContentVersionsResource` | ListByContent | List version snapshots for a content item. |
| `ContentComposite` | `*ContentCompositeResource` | CreateWithFields, DeleteRecursive | Atomic composite operations: create with fields, recursive delete. |
| `Content` | `*ContentDeliveryResource` | GetPage | Public content delivery by slug, format, and locale. |
| `Query` | `*QueryResource` | Query | Filtered, sorted, paginated content queries by datatype name. |

## Schema Resources

| Resource | Type | Key Methods | Description |
|----------|------|-------------|-------------|
| `Datatypes` | `*Resource[Datatype, CreateDatatypeParams, UpdateDatatypeParams, DatatypeID]` | CRUD, ListPaginated, Count | Content type definitions. |
| `Fields` | `*Resource[Field, CreateFieldParams, UpdateFieldParams, FieldID]` | CRUD, ListPaginated, Count | Field definitions on datatypes. |
| `FieldTypes` | `*Resource[FieldTypeInfo, CreateFieldTypeParams, UpdateFieldTypeParams, FieldTypeID]` | CRUD, ListPaginated, Count | Field type registry (text, number, boolean, etc.). |
| `DatatypeFields` | `*Resource[DatatypeField, CreateDatatypeFieldParams, UpdateDatatypeFieldParams, DatatypeFieldID]` | CRUD, ListPaginated, Count | Junction between datatypes and fields. |
| `DatatypesExtra` | `*DatatypesExtraResource` | UpdateSortOrder, MaxSortOrder | Sort order management for datatypes. |
| `FieldsExtra` | `*FieldsExtraResource` | UpdateSortOrder, MaxSortOrder | Sort order management for fields. |
| `DatatypeComposite` | `*DatatypeCompositeResource` | DeleteCascade | Cascade delete a datatype and all dependent content. |

## Media Resources

| Resource | Type | Key Methods | Description |
|----------|------|-------------|-------------|
| `Media` | `*Resource[Media, any, UpdateMediaParams, MediaID]` | Get, Update, Delete, List, ListPaginated, Count | Media file metadata. Create via MediaUpload. Responses include `DownloadURL` field (`/api/v1/media/{id}/download`). |
| `MediaDimensions` | `*Resource[MediaDimension, CreateMediaDimensionParams, UpdateMediaDimensionParams, MediaDimensionID]` | CRUD, ListPaginated, Count | Image dimension presets. |
| `MediaUpload` | `*MediaUploadResource` | Upload, UploadWithProgress | Multipart file upload with optional progress tracking. |
| `MediaAdmin` | `*MediaAdminResource` | Health, Cleanup | Media storage health check and orphan cleanup. |
| `MediaFolders` | `*MediaFoldersResource` | Tree, ListMedia, MoveMedia | Media folder hierarchy management. |
| `MediaComposite` | `*MediaCompositeResource` | GetReferences, DeleteWithCleanup | Reference scanning and safe delete with cleanup. |

## Auth and User Resources

| Resource | Type | Key Methods | Description |
|----------|------|-------------|-------------|
| `Auth` | `*AuthResource` | Login, Logout, Me, Register, RequestPasswordReset, ConfirmPasswordReset | Authentication and user registration. |
| `Users` | `*Resource[User, CreateUserParams, UpdateUserParams, UserID]` | CRUD, ListPaginated, Count | User accounts. |
| `UsersFull` | `*UsersFullResource` | List, Get | Users with expanded role and session data. |
| `Roles` | `*Resource[Role, CreateRoleParams, UpdateRoleParams, RoleID]` | CRUD, ListPaginated, Count | RBAC roles. |
| `Permissions` | `*Resource[Permission, CreatePermissionParams, UpdatePermissionParams, PermissionID]` | CRUD, ListPaginated, Count | RBAC permissions. |
| `RolePermissions` | `*RolePermissionsResource` | List, Get, Create, Delete, ListByRole | Role-permission junction management. |
| `Tokens` | `*Resource[Token, CreateTokenParams, UpdateTokenParams, TokenID]` | CRUD, ListPaginated, Count | API tokens. |
| `UsersOauth` | `*Resource[UserOauth, CreateUserOauthParams, UpdateUserOauthParams, UserOauthID]` | CRUD, ListPaginated, Count | OAuth provider links. |
| `SSHKeys` | `*SSHKeysResource` | List, Create, GetByFingerprint, Delete | SSH key management for TUI access. |
| `Sessions` | `*SessionsResource` | List, Get, Update, Remove | Active session management. |
| `UserComposite` | `*UserCompositeResource` | ReassignDelete | Reassign content ownership then delete user. |

## Admin Resources

Admin resources mirror their public counterparts but operate on draft/working content before publishing. They share the same method signatures.

| Resource | Type | Key Methods | Description |
|----------|------|-------------|-------------|
| `AdminContentData` | `*Resource[AdminContentData, ...]` | CRUD, ListPaginated, Count | Admin content data nodes. |
| `AdminContentFields` | `*Resource[AdminContentField, ...]` | CRUD, ListPaginated, Count | Admin content field values. |
| `AdminDatatypes` | `*Resource[AdminDatatype, ...]` | CRUD, ListPaginated, Count | Admin datatype definitions. |
| `AdminFields` | `*Resource[AdminField, ...]` | CRUD, ListPaginated, Count | Admin field definitions. |
| `AdminFieldTypes` | `*Resource[AdminFieldTypeInfo, ...]` | CRUD, ListPaginated, Count | Admin field type registry. |
| `AdminDatatypeFields` | `*Resource[AdminDatatypeField, ...]` | CRUD, ListPaginated, Count | Admin datatype-field junctions. |
| `AdminRoutes` | `*Resource[AdminRoute, ...]` | CRUD, ListPaginated, Count | Admin route definitions. |
| `AdminTree` | `*AdminTreeResource` | Get | Admin content tree by slug and format. |
| `AdminContentReorder` | `*AdminContentReorderResource` | Reorder, Move | Reorder and move admin content nodes. |
| `AdminPublishing` | `*PublishingResource` | Publish, Unpublish, Schedule, Restore, ListVersions, GetVersion, CreateVersion, DeleteVersion | Publishing lifecycle for admin content. |
| `AdminDatatypesExtra` | `*AdminDatatypesExtraResource` | UpdateSortOrder, MaxSortOrder | Sort order management for admin datatypes. |

## Publishing Resources

| Resource | Type | Key Methods | Description |
|----------|------|-------------|-------------|
| `Publishing` | `*PublishingResource` | Publish, Unpublish, Schedule, Restore, ListVersions, GetVersion, CreateVersion, DeleteVersion | Content publishing, versioning, and restore. |
| `AdminPublishing` | `*PublishingResource` | (same as above) | Admin content publishing. |

## Routing Resources

| Resource | Type | Key Methods | Description |
|----------|------|-------------|-------------|
| `Routes` | `*Resource[Route, CreateRouteParams, UpdateRouteParams, RouteID]` | CRUD, ListPaginated, Count | URL route definitions. |

## Operational Resources

| Resource | Type | Key Methods | Description |
|----------|------|-------------|-------------|
| `Health` | `*HealthResource` | Check | Server health check with subsystem status. |
| `Config` | `*ConfigResource` | Get, Update, Meta | Server configuration read/write and field metadata. |
| `Deploy` | `*DeployResource` | Health, Export, Import, DryRunImport | Content sync between environments. |
| `Import` | `*ImportResource` | Contentful, Sanity, Strapi, WordPress, Clean, Bulk | Import content from external CMS formats. |
| `Tables` | `*Resource[Table, CreateTableParams, UpdateTableParams, TableID]` | CRUD, ListPaginated, Count | Database table metadata. |

## Plugin Resources

| Resource | Type | Key Methods | Description |
|----------|------|-------------|-------------|
| `Plugins` | `*PluginsResource` | List, Get, Reload, Enable, Disable, CleanupDryRun, CleanupDrop | Plugin lifecycle management. |
| `PluginRoutes` | `*PluginRoutesResource` | List, Approve, Revoke | Plugin route approval management. |
| `PluginHooks` | `*PluginHooksResource` | List, Approve, Revoke | Plugin hook approval management. |

## Localization Resources

| Resource | Type | Key Methods | Description |
|----------|------|-------------|-------------|
| `Locales` | `*LocaleResource` | List, Get, Create, Update, Delete, ListPaginated, Count, ListEnabled, CreateTranslation | Locale management and content translation. |

## Webhook Resources

| Resource | Type | Key Methods | Description |
|----------|------|-------------|-------------|
| `Webhooks` | `*WebhookResource` | List, Get, Create, Update, Delete, ListPaginated, Count, Test, ListDeliveries, RetryDelivery | Webhook CRUD plus testing and delivery management. |

## Branded ID Types

All entity IDs are distinct `string`-based types. Each provides `String() string` and `IsZero() bool` methods.

| Type | Used by |
|------|---------|
| `ContentID` | ContentData |
| `ContentFieldID` | ContentFields |
| `ContentRelationID` | ContentRelations |
| `AdminContentID` | AdminContentData |
| `AdminContentFieldID` | AdminContentFields |
| `AdminContentRelationID` | Admin content relations |
| `DatatypeID` | Datatypes |
| `FieldID` | Fields |
| `FieldTypeID` | FieldTypes |
| `DatatypeFieldID` | DatatypeFields |
| `AdminDatatypeID` | AdminDatatypes |
| `AdminFieldID` | AdminFields |
| `AdminFieldTypeID` | AdminFieldTypes |
| `AdminDatatypeFieldID` | AdminDatatypeFields |
| `MediaID` | Media |
| `MediaDimensionID` | MediaDimensions |
| `MediaFolderID` | MediaFolders |
| `UserID` | Users |
| `RoleID` | Roles |
| `SessionID` | Sessions |
| `TokenID` | Tokens |
| `UserOauthID` | UsersOauth |
| `UserSshKeyID` | SSHKeys |
| `PermissionID` | Permissions |
| `RolePermissionID` | RolePermissions |
| `RouteID` | Routes |
| `AdminRouteID` | AdminRoutes |
| `TableID` | Tables |
| `EventID` | Change events (audit log) |
| `BackupID` | Backups |
| `LocaleID` | Locales |
| `ContentVersionID` | Content versions |
| `AdminContentVersionID` | Admin content versions |
| `WebhookID` | Webhooks |
| `WebhookDeliveryID` | Webhook deliveries |

### Value Types

These are branded string types for domain-specific values, not entity IDs:

| Type | Purpose |
|------|---------|
| `Slug` | URL path slugs |
| `Email` | Email addresses |
| `URL` | URLs |
| `Timestamp` | ISO 8601 timestamps. Provides `Time() (time.Time, error)`, `NewTimestamp(time.Time)`, and `TimestampNow()`. |

## Enums

### ContentStatus

| Constant | Value |
|----------|-------|
| `ContentStatusDraft` | `"draft"` |
| `ContentStatusPublished` | `"published"` |

### FieldType

| Constant | Value |
|----------|-------|
| `FieldTypeText` | `"text"` |
| `FieldTypeTextarea` | `"textarea"` |
| `FieldTypeNumber` | `"number"` |
| `FieldTypeDate` | `"date"` |
| `FieldTypeDatetime` | `"datetime"` |
| `FieldTypeBoolean` | `"boolean"` |
| `FieldTypeSelect` | `"select"` |
| `FieldTypeMedia` | `"media"` |
| `FieldTypeID` | `"_id"` |
| `FieldTypeJSON` | `"json"` |
| `FieldTypeRichtext` | `"richtext"` |
| `FieldTypeSlug` | `"slug"` |
| `FieldTypeEmail` | `"email"` |
| `FieldTypeURL` | `"url"` |

### RouteType

| Constant | Value |
|----------|-------|
| `RouteTypeStatic` | `"static"` |
| `RouteTypeDynamic` | `"dynamic"` |
| `RouteTypeAPI` | `"api"` |
| `RouteTypeRedirect` | `"redirect"` |

## Utility Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `NewClient` | `(ClientConfig) (*Client, error)` | Create a new API client. |
| `IsNotFound` | `(error) bool` | Check if error is HTTP 404. |
| `IsUnauthorized` | `(error) bool` | Check if error is HTTP 401. |
| `IsDuplicateMedia` | `(error) bool` | Check if error is HTTP 409 (duplicate upload). |
| `IsInvalidMediaPath` | `(error) bool` | Check if error is HTTP 400 (path traversal). |
| `StringPtr` | `(string) *string` | Create a `*string` from a string value. |
| `NewTimestamp` | `(time.Time) Timestamp` | Create a `Timestamp` from a `time.Time`. |
| `TimestampNow` | `() Timestamp` | Create a `Timestamp` for the current time. |
