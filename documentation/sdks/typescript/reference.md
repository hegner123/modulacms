# Reference

Quick reference for all resources, methods, and types across the TypeScript SDK packages.

## Read-Only SDK (`@modulacms/sdk`)

### ModulaClient Methods

| Method | Return Type | Description |
|--------|------------|-------------|
| `getPage(slug, options?)` | `Promise<T>` | Fetch rendered content tree by route slug |
| `listRoutes()` | `Promise<Route[]>` | List all public routes |
| `getRoute(id)` | `Promise<Route>` | Get a route by ID |
| `listContentData()` | `Promise<ContentData[]>` | List all content data nodes |
| `getContentData(id)` | `Promise<ContentData>` | Get a content data node |
| `listContentFields()` | `Promise<ContentField[]>` | List all content field values |
| `getContentField(id)` | `Promise<ContentField>` | Get a content field value |
| `listMedia()` | `Promise<Media[]>` | List all media assets |
| `getMedia(id)` | `Promise<Media>` | Get a media asset |
| `listMediaPaginated(params)` | `Promise<PaginatedResponse<Media>>` | List media with pagination |
| `listMediaDimensions()` | `Promise<MediaDimension[]>` | List media dimension presets |
| `getMediaDimension(id)` | `Promise<MediaDimension>` | Get a media dimension preset |
| `listDatatypes()` | `Promise<Datatype[]>` | List all datatypes |
| `getDatatype(id)` | `Promise<Datatype>` | Get a datatype |
| `listFields()` | `Promise<Field[]>` | List all field definitions |
| `getField(id)` | `Promise<Field>` | Get a field definition |
| `queryContent(datatype, params?)` | `Promise<QueryResult>` | Query content by datatype name |

## Admin SDK (`@modulacms/admin-sdk`)

### CRUD Resources

Every CRUD resource provides these standard methods:

| Method | Return Type |
|--------|------------|
| `list(opts?)` | `Promise<Entity[]>` |
| `get(id, opts?)` | `Promise<Entity>` |
| `create(params, opts?)` | `Promise<Entity>` |
| `update(params, opts?)` | `Promise<Entity>` |
| `remove(id, opts?)` | `Promise<void>` |
| `listPaginated(params, opts?)` | `Promise<PaginatedResponse<Entity>>` |
| `count(opts?)` | `Promise<number>` |

### ModulaCMSAdminClient Resources

| Resource | Entity Type | ID Type | Extra Methods |
|----------|------------|---------|---------------|
| `contentData` | `ContentData` | `ContentID` | `reorder`, `move`, `batch`, `createWithFields`, `deleteRecursive` |
| `contentFields` | `ContentField` | `ContentFieldID` | -- |
| `datatypes` | `Datatype` | `DatatypeID` | `getFull`, `deleteCascade`, `updateSortOrder`, `maxSortOrder` |
| `fields` | `Field` | `FieldID` | -- |
| `fieldTypes` | `FieldTypeInfo` | `FieldTypeID` | -- |
| `routes` | `Route` | `RouteID` | -- |
| `media` | `Media` | `MediaID` | `health`, `cleanup`, `getReferences`, `deleteWithCleanup` |
| `mediaFolders` | `MediaFolder` | `MediaFolderID` | `tree`, `listMedia`, `moveMedia` |
| `mediaDimensions` | `MediaDimension` | -- | -- |
| `users` | `User` | `UserID` | `listFull`, `getFull`, `reassignDelete` |
| `roles` | `Role` | `RoleID` | -- |
| `permissions` | `Permission` | `PermissionID` | -- |
| `tokens` | `Token` | -- | -- |
| `usersOauth` | `UserOauth` | `UserOauthID` | -- |
| `tables` | `Table` | -- | -- |
| `adminRoutes` | `AdminRoute` | `Slug` (get/list), `AdminRouteID` (remove) | `listOrdered` |
| `adminContentData` | `AdminContentData` | `AdminContentID` | `reorder`, `move` |
| `adminContentFields` | `AdminContentField` | `AdminContentFieldID` | -- |
| `adminDatatypes` | `AdminDatatype` | `AdminDatatypeID` | `updateSortOrder`, `maxSortOrder` |
| `adminFields` | `AdminField` | `AdminFieldID` | -- |
| `adminFieldTypes` | `AdminFieldTypeInfo` | `AdminFieldTypeID` | -- |

### Specialized Resources

| Resource | Methods |
|----------|---------|
| `auth` | `login`, `logout`, `me`, `register`, `reset` |
| `adminTree` | `get(slug, format?)` |
| `mediaUpload` | `upload(file, opts?)` |
| `sessions` | `update`, `remove` |
| `sshKeys` | `list`, `create`, `remove` |
| `rolePermissions` | `list`, `get`, `create`, `remove`, `listByRole` |
| `contentTree` | `save(params)` |
| `contentHeal` | `heal(dryRun?)` |
| `contentDelivery` | `getPage(slug, format?, locale?)` |
| `publishing` | `publish`, `unpublish`, `schedule`, `listVersions`, `getVersion`, `createVersion`, `deleteVersion`, `restore` |
| `adminPublishing` | `publish`, `unpublish`, `schedule`, `listVersions`, `getVersion`, `createVersion`, `deleteVersion`, `restore` |
| `plugins` | `list`, `get`, `reload`, `enable`, `disable`, `cleanupDryRun`, `cleanupDrop` |
| `pluginRoutes` | `list`, `approve`, `revoke` |
| `pluginHooks` | `list`, `approve`, `revoke` |
| `config` | `get`, `update`, `meta` |
| `locales` | `list`, `get`, `create`, `update`, `remove`, `createTranslation` |
| `webhooks` | `list`, `get`, `create`, `update`, `remove`, `test`, `listDeliveries`, `retryDelivery` |
| `query` | `query(datatype, params?)` |
| `deploy` | `health`, `export`, `importPayload` |
| `import` | `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `bulk` |

## Type Index

### From `@modulacms/types`

**IDs:** `UserID`, `ContentID`, `ContentFieldID`, `ContentRelationID`, `ContentVersionID`, `DatatypeID`, `FieldID`, `MediaID`, `MediaFolderID`, `RoleID`, `PermissionID`, `RolePermissionID`, `FieldTypeID`, `RouteID`, `SessionID`, `UserOauthID`, `AdminContentID`, `AdminContentFieldID`, `AdminContentRelationID`, `AdminContentVersionID`, `AdminDatatypeID`, `AdminFieldID`, `AdminRouteID`, `AdminFieldTypeID`, `LocaleID`, `WebhookID`, `WebhookDeliveryID`, `Slug`, `Email`, `URL`

**Entities:** `ContentData`, `ContentField`, `ContentRelation`, `ContentVersion`, `AdminContentVersion`, `Datatype`, `Field`, `FieldTypeInfo`, `AdminFieldTypeInfo`, `Route`, `Media` (includes `download_url` field), `MediaDimension`, `MediaFolder`, `Locale`, `Webhook`, `WebhookDelivery`

**Content tree:** `ContentTree`, `ContentNode`, `NodeDatatype`, `NodeField`

**Query:** `QueryParams`, `QueryResult`, `QueryItem`, `QueryDatatype`

**Enums:** `ContentStatus`, `FieldType`, `ContentFormat`, `CONTENT_FORMATS`

**Pagination:** `PaginationParams`, `PaginatedResponse<T>`

**Errors:** `ApiError`, `isApiError`

**Utilities:** `ULID`, `Timestamp`, `NullableString`, `NullableNumber`, `Brand<T, B>`, `RequestOptions`

### From `@modulacms/admin-sdk`

**Client:** `ClientConfig`, `ModulaCMSAdminClient`, `createAdminClient`

**Resource types:** `CrudResource<E, C, U, ID>`, `PublishingResource`, `AdminPublishingResource`, `PluginsResource`, `PluginRoutesResource`, `PluginHooksResource`, `ConfigResource`, `ContentDeliveryResource`, `LocalesResource`, `WebhooksResource`, `QueryResource`

**Admin entities:** `AdminRoute`, `AdminContentData`, `AdminContentField`, `AdminDatatype`, `AdminField`, `AdminContentRelation`, `User`, `Role`, `Permission`, `RolePermission`, `Token`, `UserOauth`, `Session`, `SshKey`, `SshKeyListItem`, `Table`

**View types:** `UserWithRoleLabel`, `UserFullView`, `UserOauthView`, `UserSshKeyView`, `SessionView`, `TokenView`, `DatatypeFullView`

**Auth:** `LoginRequest`, `LoginResponse`, `MeResponse`

**Deploy:** `DeployHealthResponse`, `DeploySyncPayload`, `DeploySyncResult`, `DeploySyncError`

**Plugins:** `PluginListItem`, `PluginInfo`, `PluginActionResponse`, `PluginStateResponse`, `PluginRoute`, `PluginHook`, `CleanupDryRunResponse`, `CleanupDropResponse`

**Config:** `ConfigFieldMeta`, `ConfigGetResponse`, `ConfigUpdateResponse`, `ConfigMetaResponse`

**Publishing:** `PublishRequest`, `PublishResponse`, `ScheduleRequest`, `ScheduleResponse`, `RestoreRequest`, `RestoreResponse`, `CreateVersionRequest` (and admin equivalents)

**Content operations:** `TreeSaveRequest`, `TreeSaveResponse`, `TreeNodeCreate`, `TreeNodeUpdate`, `BatchContentUpdateParams`, `BatchContentUpdateResponse`, `ContentCreateParams`, `ContentCreateResponse`, `RecursiveDeleteResponse`, `HealReport`, `HealRepair`

**Media:** `MediaUploadOptions`, `MediaHealthResponse`, `MediaCleanupResponse`, `MediaReferenceScanResponse`

**Webhooks:** `CreateWebhookParams`, `UpdateWebhookParams`, `WebhookTestResponse`

**Locales:** `CreateLocaleParams`, `UpdateLocaleParams`, `CreateTranslationRequest`, `CreateTranslationResponse`

**Import:** `ImportFormat`, `ImportResponse`

### From `@modulacms/sdk`

**Client:** `ModulaClient`, `ModulaClientConfig`, `GetPageOptions<T>`, `Validator<T>`

**Errors:** `ModulaError`

## Package Dependency Diagram

```
@modulacms/types
    ^          ^
    |          |
@modulacms/sdk   @modulacms/admin-sdk
(read-only)      (full CRUD)
```

Both SDKs depend on `@modulacms/types` for shared entity types, branded IDs, and enums. The two SDKs are independent of each other -- install whichever one your application needs, or both.
