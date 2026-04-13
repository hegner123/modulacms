# Swift SDK -- Reference

Quick reference for all resources, branded ID types, enums, and utility types in the Modula Swift SDK.

## Client Resources

`ModulaClient` exposes the following resources. All are `Sendable`.

### Standard CRUD Resources

Generic `Resource<Entity, CreateParams, UpdateParams, ID>` instances. Each provides `list`, `get`, `create`, `update`, `delete`, `listPaginated`, `count`, and `rawList`.

| Property | Entity | Create Params | Update Params | ID Type |
|----------|--------|---------------|---------------|---------|
| `contentData` | `ContentData` | `CreateContentDataParams` | `UpdateContentDataParams` | `ContentID` |
| `contentFields` | `ContentField` | `CreateContentFieldParams` | `UpdateContentFieldParams` | `ContentFieldID` |
| `contentRelations` | `ContentRelation` | `CreateContentRelationParams` | `UpdateContentRelationParams` | `ContentRelationID` |
| `datatypes` | `Datatype` | `CreateDatatypeParams` | `UpdateDatatypeParams` | `DatatypeID` |
| `fields` | `Field` | `CreateFieldParams` | `UpdateFieldParams` | `FieldID` |
| `media` | `Media` (includes `downloadURL`) | `NoCreate` | `UpdateMediaParams` | `MediaID` |
| `mediaDimensions` | `MediaDimension` | `CreateMediaDimensionParams` | `UpdateMediaDimensionParams` | `MediaDimensionID` |
| `routes` | `Route` | `CreateRouteParams` | `UpdateRouteParams` | `RouteID` |
| `roles` | `Role` | `CreateRoleParams` | `UpdateRoleParams` | `RoleID` |
| `permissions` | `Permission` | `CreatePermissionParams` | `UpdatePermissionParams` | `PermissionID` |
| `users` | `User` | `CreateUserParams` | `UpdateUserParams` | `UserID` |
| `tokens` | `Token` | `CreateTokenParams` | `UpdateTokenParams` | `TokenID` |
| `usersOauth` | `UserOauth` | `CreateUserOauthParams` | `UpdateUserOauthParams` | `UserOauthID` |
| `tables` | `Table` | `CreateTableParams` | `UpdateTableParams` | `TableID` |

### Admin CRUD Resources

Same `Resource` interface as above, operating on admin-scoped entities.

| Property | Entity | Create Params | Update Params | ID Type |
|----------|--------|---------------|---------------|---------|
| `adminContentData` | `AdminContentData` | `CreateAdminContentDataParams` | `UpdateAdminContentDataParams` | `AdminContentID` |
| `adminContentFields` | `AdminContentField` | `CreateAdminContentFieldParams` | `UpdateAdminContentFieldParams` | `AdminContentFieldID` |
| `adminDatatypes` | `AdminDatatype` | `CreateAdminDatatypeParams` | `UpdateAdminDatatypeParams` | `AdminDatatypeID` |
| `adminFields` | `AdminField` | `CreateAdminFieldParams` | `UpdateAdminFieldParams` | `AdminFieldID` |
| `adminRoutes` | `AdminRoute` | `CreateAdminRouteParams` | `UpdateAdminRouteParams` | `AdminRouteID` |
| `fieldTypes` | `FieldTypeInfo` | `CreateFieldTypeParams` | `UpdateFieldTypeParams` | `FieldTypeID` |
| `adminFieldTypes` | `AdminFieldTypeInfo` | `CreateAdminFieldTypeParams` | `UpdateAdminFieldTypeParams` | `AdminFieldTypeID` |
| `adminMedia` | `AdminMedia` | `NoCreate` | `UpdateAdminMediaParams` | `AdminMediaID` |

### Specialized Resources

| Property | Type | Methods |
|----------|------|---------|
| `auth` | `AuthResource` | `login`, `logout`, `me`, `register`, `resetPassword`, `requestPasswordReset`, `confirmPasswordReset` |
| `mediaUpload` | `MediaUploadResource` | `upload(data:filename:options:)` |
| `adminMediaUpload` | `AdminMediaUploadResource` | `upload(data:filename:options:)` |
| `adminMediaFolders` | `AdminMediaFoldersResource` | `tree`, `listMedia`, `moveMedia` |
| `adminTree` | `AdminTreeResource` | `get(slug:format:)` |
| `content` | `ContentDeliveryResource` | `getPage(slug:format:locale:)` |
| `sshKeys` | `SSHKeysResource` | `list`, `create`, `delete` |
| `sessions` | `SessionsResource` | `update`, `remove` |
| `importResource` | `ImportResource` | `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `bulk` |
| `contentBatch` | `ContentBatchResource` | `update` |
| `contentTree` | `ContentTreeResource` | `save` |
| `contentHeal` | `ContentHealResource` | `heal(dryRun:)` |
| `deploy` | `DeployResource` | `health`, `export`, `importPayload`, `dryRunImport` |
| `rolePermissions` | `RolePermissionsResource` | `list`, `get`, `create`, `delete`, `listByRole` |
| `plugins` | `PluginsResource` | `list`, `get`, `reload`, `enable`, `disable`, `cleanupDryRun`, `cleanupDrop` |
| `pluginRoutes` | `PluginRoutesResource` | `list`, `approve`, `revoke` |
| `pluginHooks` | `PluginHooksResource` | `list`, `approve`, `revoke` |
| `publishing` | `PublishingResource` | `publish`, `unpublish`, `schedule`, `listVersions`, `getVersion`, `createVersion`, `deleteVersion`, `restore` |
| `adminPublishing` | `AdminPublishingResource` | Same as `publishing` with admin-prefixed types |
| `locales` | `LocaleResource` | `list`, `get`, `create`, `update`, `delete`, `createTranslation` |
| `webhooks` | `WebhookResource` | `list`, `get`, `create`, `update`, `delete`, `test`, `listDeliveries`, `retryDelivery` |
| `query` | `QueryResource` | `query(datatype:params:)` |
| `contentComposite` | `ContentCompositeResource` | `createWithFields`, `deleteRecursive` |
| `userComposite` | `UserCompositeResource` | `reassignDelete` |
| `mediaFolders` | `MediaFoldersResource` | `tree`, `listMedia`, `moveMedia` |
| `mediaComposite` | `MediaCompositeResource` | `getReferences`, `deleteWithCleanup` |
| `datatypeComposite` | `DatatypeCompositeResource` | `deleteCascade` |
| `config` | `ConfigResource` | `get(category:)`, `update(updates:)`, `meta` |

## Branded ID Types

All IDs conform to the `ResourceID` protocol:

```swift
public protocol ResourceID: Codable, Sendable, Hashable,
    CustomStringConvertible, ExpressibleByStringLiteral {
    var rawValue: String { get }
    init(_ rawValue: String)
}
```

Every `ResourceID` type provides:
- `description` -- returns `rawValue`
- `isZero` -- true if `rawValue` is empty
- String literal initialization: `let id: ContentID = "01HXYZ..."`
- Automatic `Codable` conformance (encodes/decodes as a plain string)

### Content IDs

| Type | Usage |
|------|-------|
| `ContentID` | Content data nodes |
| `ContentFieldID` | Content field values |
| `ContentRelationID` | Content relations |

### Admin Content IDs

| Type | Usage |
|------|-------|
| `AdminContentID` | Admin content data nodes |
| `AdminContentFieldID` | Admin content field values |
| `AdminContentRelationID` | Admin content relations |

### Schema IDs

| Type | Usage |
|------|-------|
| `DatatypeID` | Datatypes (content types) |
| `FieldID` | Field definitions |
| `AdminDatatypeID` | Admin datatypes |
| `AdminFieldID` | Admin field definitions |

### Field Type IDs

| Type | Usage |
|------|-------|
| `FieldTypeID` | Field type definitions |
| `AdminFieldTypeID` | Admin field type definitions |

### Media IDs

| Type | Usage |
|------|-------|
| `MediaID` | Media items |
| `MediaDimensionID` | Media dimension presets |
| `MediaFolderID` | Media folders |
| `AdminMediaID` | Admin media items |
| `AdminMediaFolderID` | Admin media folders |

### Auth IDs

| Type | Usage |
|------|-------|
| `UserID` | Users |
| `RoleID` | Roles |
| `SessionID` | Sessions |
| `TokenID` | API tokens |
| `UserOauthID` | OAuth connections |
| `UserSshKeyID` | SSH keys |
| `PermissionID` | Permissions |
| `RolePermissionID` | Role-permission assignments |

### Routing IDs

| Type | Usage |
|------|-------|
| `RouteID` | Routes |
| `AdminRouteID` | Admin routes |

### Locale IDs

| Type | Usage |
|------|-------|
| `LocaleID` | Locales |

### Version IDs

| Type | Usage |
|------|-------|
| `ContentVersionID` | Content version snapshots |
| `AdminContentVersionID` | Admin content version snapshots |

### Webhook IDs

| Type | Usage |
|------|-------|
| `WebhookID` | Webhooks |
| `WebhookDeliveryID` | Webhook delivery records |

### Other IDs

| Type | Usage |
|------|-------|
| `TableID` | Database tables |
| `EventID` | Change events |
| `BackupID` | Backups |

### Value Types

These also conform to `ResourceID` but represent values rather than entity identifiers:

| Type | Usage |
|------|-------|
| `Slug` | URL slugs |
| `Email` | Email addresses |
| `URLValue` | URL strings |

### Timestamp

`Timestamp` conforms to `ResourceID` and adds date conversion:

```swift
public struct Timestamp: ResourceID {
    public let rawValue: String

    /// Parse the ISO 8601 string into a Date.
    public func date() -> Date?

    /// Create a Timestamp from a Date.
    public init(date: Date)

    /// Create a Timestamp for the current moment.
    public static func now() -> Timestamp
}
```

Usage:

```swift
let content = try await client.contentData.get(id: someID)
if let created = content.dateCreated.date() {
    print("Created: \(created)")
}

let ts = Timestamp.now()
```

## Enums

### ContentStatus

```swift
public enum ContentStatus: String, Codable, Sendable {
    case draft
    case published
}
```

### FieldType

```swift
public enum FieldType: String, Codable, Sendable {
    case text
    case textarea
    case number
    case date
    case datetime
    case boolean
    case select
    case media
    case id = "_id"
    case json
    case richtext
    case slug
    case email
    case url
}
```

### RouteType

```swift
public enum RouteType: String, Codable, Sendable {
    case `static`
    case dynamic
    case api
    case redirect
}
```

## JSONValue

`JSONValue` is a recursive enum for representing arbitrary JSON:

```swift
public enum JSONValue: Codable, Sendable, Equatable {
    case string(String)
    case number(Double)
    case bool(Bool)
    case object([String: JSONValue])
    case array([JSONValue])
    case null
}
```

Used by `ConfigResource` for dynamic configuration values:

```swift
let response = try await client.config.get()
for (key, value) in response.config {
    switch value {
    case .string(let s): print("\(key) = \(s)")
    case .number(let n): print("\(key) = \(n)")
    case .bool(let b): print("\(key) = \(b)")
    case .null: print("\(key) = null")
    case .object, .array: print("\(key) = [complex]")
    }
}

// Update config with JSONValue
try await client.config.update(updates: [
    "port": .number(9090),
    "site_name": .string("My CMS"),
])
```

## JSON Encoding/Decoding

The SDK exposes shared encoder and decoder instances:

```swift
public enum JSON {
    public static let decoder: JSONDecoder
    public static let encoder: JSONEncoder
}
```

Use these when decoding raw `Data` responses from `getPage`, `contentBatch.update`, or `rawList`:

```swift
let data = try await client.content.getPage(slug: "blog/post")
let page = try JSON.decoder.decode(MyPageModel.self, from: data)
```

## NoCreate

`NoCreate` is an uninhabited enum used as the `CreateParams` type for resources that don't support JSON POST creation (e.g., `Media`, which uses multipart upload). Calling `create(params:)` on such a resource produces a compile error.

```swift
// This does not compile:
// try await client.media.create(params: ???)

// Use mediaUpload instead:
let media = try await client.mediaUpload.upload(data: fileData, filename: "photo.jpg")
```

## API Error Type

```swift
public struct APIError: Error, LocalizedError, Sendable {
    public let statusCode: Int
    public let message: String
    public let body: String

    public init(statusCode: Int, message: String = "", body: String = "")
    public var errorDescription: String?
}

public func isNotFound(_ error: Error) -> Bool
public func isUnauthorized(_ error: Error) -> Bool
public func isDuplicateMedia(_ error: Error) -> Bool
public func isInvalidMediaPath(_ error: Error) -> Bool
```

See [Error Handling](/docs/sdks/swift/error-handling) for usage patterns.

## Pagination Types

```swift
public struct PaginationParams: Sendable {
    public let limit: Int64
    public let offset: Int64
    public init(limit: Int64, offset: Int64)
}

public struct PaginatedResponse<T: Decodable & Sendable>: Decodable, Sendable {
    public let data: [T]
    public let total: Int64
    public let limit: Int64
    public let offset: Int64
}
```

## ClientConfig

```swift
public struct ClientConfig: Sendable {
    public let baseURL: String
    public let apiKey: String
    public let urlSession: URLSession?

    public init(baseURL: String, apiKey: String = "", urlSession: URLSession? = nil)
}
```

## Related Documentation

- [Getting Started](/docs/sdks/swift/getting-started) -- installation, client setup, first requests
- [Content Operations](/docs/sdks/swift/content-operations) -- CRUD, content delivery, trees, media
- [Error Handling](/docs/sdks/swift/error-handling) -- APIError, do/catch patterns
- [REST API Reference](/docs/api/rest-api) -- full HTTP endpoint documentation
