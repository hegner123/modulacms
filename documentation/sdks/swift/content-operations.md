# Swift SDK -- Content Operations

CRUD operations, content delivery, content trees, media upload, the query API, and SwiftUI integration patterns.

## Generic Resource CRUD

Most entities on `ModulaClient` are exposed as `Resource<Entity, CreateParams, UpdateParams, ID>`. Every `Resource` provides the same seven methods:

| Method | Signature | Description |
|--------|-----------|-------------|
| `list()` | `async throws -> [Entity]` | List all entities |
| `get(id:)` | `async throws -> Entity` | Get one by ID |
| `create(params:)` | `async throws -> Entity` | Create and return the new entity |
| `update(params:)` | `async throws -> Entity` | Update and return the modified entity |
| `delete(id:)` | `async throws` | Delete by ID |
| `listPaginated(params:)` | `async throws -> PaginatedResponse<Entity>` | List with pagination |
| `count()` | `async throws -> Int64` | Count all entities |

There is also `rawList(queryItems:)` which returns raw `Data` for custom query parameters.

### List

```swift
let datatypes = try await client.datatypes.list()
```

### Get by ID

```swift
let datatype = try await client.datatypes.get(id: DatatypeID("01HXYZ..."))
```

### Create

```swift
let newDatatype = try await client.datatypes.create(params: CreateDatatypeParams(
    name: "blog_post",
    label: "Blog Post"
))
```

### Update

```swift
let updated = try await client.datatypes.update(params: UpdateDatatypeParams(
    datatypeID: DatatypeID("01HXYZ..."),
    name: "blog_post",
    label: "Blog Post (Updated)"
))
```

### Delete

```swift
try await client.datatypes.delete(id: DatatypeID("01HXYZ..."))
```

### Paginated List

```swift
let page = try await client.datatypes.listPaginated(params: PaginationParams(
    limit: 20,
    offset: 0
))
print("Page has \(page.data.count) items out of \(page.total) total")
```

`PaginatedResponse<T>` contains:

| Field | Type | Description |
|-------|------|-------------|
| `data` | `[T]` | Items in the current page |
| `total` | `Int64` | Total count across all pages |
| `limit` | `Int64` | Requested page size |
| `offset` | `Int64` | Requested offset |

### Count

```swift
let total = try await client.datatypes.count()
```

### NoCreate Resources

`Media` uses `NoCreate` as its create params type because you create media via multipart upload, not JSON POST. Calling `create(params:)` on `client.media` produces a compile error.

```swift
// Upload media instead:
let media = try await client.mediaUpload.upload(
    data: imageData,
    filename: "photo.jpg"
)

// List and get work normally:
let allMedia = try await client.media.list()
let one = try await client.media.get(id: MediaID("01HXYZ..."))
```

## Content Delivery

`ContentDeliveryResource` retrieves published content by slug. This is the primary read path for frontend applications.

```swift
let pageData = try await client.content.getPage(slug: "blog/hello-world")
```

`getPage` returns raw `Data` because the response shape varies based on the `format` parameter.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `slug` | `String` | required | URL path to the content. Leading `/` is stripped. |
| `format` | `String` | `""` | Response format: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `raw`. Empty uses server default. |
| `locale` | `String` | `""` | Locale code for localized content. Empty uses default locale. |

```swift
let data = try await client.content.getPage(
    slug: "blog/hello-world",
    format: "clean",
    locale: "en"
)

// Decode the response based on your chosen format
let page = try JSON.decoder.decode(YourPageModel.self, from: data)
```

## Query API

`QueryResource` provides structured content queries by datatype name with sorting, filtering, pagination, and locale support.

```swift
let result = try await client.query.query(
    datatype: "blog_post",
    params: QueryParams(
        sort: "-date_created",
        limit: 10,
        offset: 0,
        locale: "en",
        status: "published",
        filters: ["category": "tech"]
    )
)
```

`QueryParams` fields:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sort` | `String` | `""` | Sort field. Prefix with `-` for descending. |
| `limit` | `Int` | `0` | Max items to return. 0 uses server default. |
| `offset` | `Int` | `0` | Number of items to skip. |
| `locale` | `String` | `""` | Locale code filter. |
| `status` | `String` | `""` | Content status filter (`draft`, `published`). |
| `filters` | `[String: String]` | `[:]` | Field-level filters passed as query parameters. |

`QueryResult` contains:

| Field | Type | Description |
|-------|------|-------------|
| `data` | `[QueryItem]` | Matching content items |
| `total` | `Int` | Total matching count |
| `limit` | `Int` | Applied limit |
| `offset` | `Int` | Applied offset |
| `datatype` | `QueryDatatype` | Datatype metadata (name, label) |

Each `QueryItem` provides `contentDataID`, `datatypeID`, `authorID`, `status`, dates, and a `fields` dictionary mapping field names to string values.

## Content Trees

Content in ModulaCMS is organized as hierarchical trees that you can read, save, and restructure.

### Reading Trees

Use `adminTree` to fetch a full tree by slug:

```swift
let treeData = try await client.adminTree.get(slug: "blog", format: "clean")
```

Returns raw `Data` that you decode based on your format.

### Saving Tree Changes

`ContentTreeResource` atomically applies creates, deletes, and pointer-field updates in a single request:

```swift
let response = try await client.contentTree.save(TreeSaveRequest(
    contentID: rootContentID,
    creates: [
        TreeNodeCreate(
            clientID: UUID().uuidString,
            datatypeID: datatypeID.rawValue,
            parentID: rootContentID.rawValue
        )
    ],
    updates: [
        TreeNodeUpdate(
            contentDataID: existingNodeID,
            firstChildID: "client-uuid-from-creates"
        )
    ],
    deletes: [removedNodeID]
))

print("Created: \(response.created), Updated: \(response.updated), Deleted: \(response.deleted)")

// Map client IDs to server-generated ULIDs
if let idMap = response.idMap {
    for (clientID, serverID) in idMap {
        print("\(clientID) -> \(serverID)")
    }
}

// Check for partial failures
if let errors = response.errors, !errors.isEmpty {
    for err in errors {
        print("Error: \(err)")
    }
}
```

Key types:

- `TreeNodeCreate` -- new node with a `clientID` you generate. The server returns a mapping in `idMap`. Pointer fields can reference other client IDs.
- `TreeNodeUpdate` -- updates only the four pointer fields on an existing node. All other fields are preserved.
- `TreeSaveRequest` -- wraps creates, updates, and deletes. `contentID` identifies the root node for route inheritance.
- `TreeSaveResponse` -- counts of successful operations, the `idMap`, and any per-node `errors`.

> **Good to know**: The server processes deletes before updates, so removed nodes don't interfere with tree restructuring.

### Composite Content Operations

`ContentCompositeResource` provides higher-level operations:

```swift
// Create a content node with field values in one request
let result = try await client.contentComposite.createWithFields(params: ContentCreateParams(
    datatypeID: DatatypeID("01HXYZ..."),
    fields: [
        ContentCreateFieldValue(fieldID: FieldID("01HABC..."), value: "Hello World"),
        ContentCreateFieldValue(fieldID: FieldID("01HDEF..."), value: "Post body text"),
    ]
))
print("Created content: \(result.contentData.contentDataID)")

// Delete a content node and all descendants
let deleteResult = try await client.contentComposite.deleteRecursive(id: ContentID("01HXYZ..."))
print("Deleted \(deleteResult.totalDeleted) nodes")
```

## Batch Updates

`ContentBatchResource` applies multiple content changes in a single request:

```swift
let responseData = try await client.contentBatch.update(request: yourBatchPayload)
```

The `update` method accepts any `Encodable & Sendable` request body and returns raw `Data`. Structure your batch payload according to the batch API contract.

## Media Upload

`MediaUploadResource` handles multipart file uploads:

```swift
let imageData = try Data(contentsOf: imageURL)
let media = try await client.mediaUpload.upload(
    data: imageData,
    filename: "hero-image.jpg"
)
print("Uploaded: \(media.mediaID) at \(media.s3Key)")
```

### Upload with Path

Organize uploaded files into S3 key prefixes:

```swift
let media = try await client.mediaUpload.upload(
    data: imageData,
    filename: "hero.jpg",
    options: MediaUploadResource.UploadOptions(path: "blog/headers")
)
```

> **Good to know**: When `path` is nil, the server organizes files by date (`YYYY/M`).

### Media Composite Operations

Scan for references before deleting:

```swift
// Check which content fields reference a media item
let refs = try await client.mediaComposite.getReferences(id: MediaID("01HXYZ..."))
print("\(refs.references.count) content fields reference this media")

// Delete media and clean up all references
try await client.mediaComposite.deleteWithCleanup(id: MediaID("01HXYZ..."))
```

### Admin Media

Admin media items are stored separately and power the admin panel UI. The API mirrors the public media resources:

```swift
// List and get admin media
let adminItems = try await client.adminMedia.list()
let item = try await client.adminMedia.get(id: AdminMediaID("01HXYZ..."))

// Upload a file to admin media
let uploaded = try await client.adminMediaUpload.upload(
    data: imageData,
    filename: "admin-logo.png"
)

// Admin media folders
let tree = try await client.adminMediaFolders.tree()
let folderMedia = try await client.adminMediaFolders.listMedia(
    folderID: AdminMediaFolderID("01HXYZ..."),
    params: PaginationParams(limit: 20, offset: 0)
)
```

## Publishing

`PublishingResource` manages content lifecycle:

```swift
// Publish content (creates a version snapshot)
let result = try await client.publishing.publish(req: PublishRequest(
    contentDataID: "01HXYZ..."
))

// Unpublish
let result = try await client.publishing.unpublish(req: PublishRequest(
    contentDataID: "01HXYZ..."
))

// Schedule future publication
let result = try await client.publishing.schedule(req: ScheduleRequest(
    contentDataID: "01HXYZ...",
    publishAt: "2026-04-01T00:00:00Z"
))
```

### Versions

```swift
// List version history
let versions = try await client.publishing.listVersions(contentDataID: "01HXYZ...")

// Get a specific version
let version = try await client.publishing.getVersion(versionID: "01HVER...")

// Manually create a version snapshot
let version = try await client.publishing.createVersion(req: CreateVersionRequest(
    contentDataID: "01HXYZ..."
))

// Restore content to a previous version
let result = try await client.publishing.restore(req: RestoreRequest(
    contentDataID: "01HXYZ...",
    versionID: "01HVER..."
))

// Delete a version
try await client.publishing.deleteVersion(versionID: "01HVER...")
```

> **Good to know**: Admin content uses the parallel `adminPublishing` resource with identical methods and admin-prefixed types.

## Authentication

```swift
// Login
let response = try await client.auth.login(params: LoginParams(
    email: "user@example.com",
    password: "password"
))
print("Token: \(response.token)")

// Get current user
let user = try await client.auth.me()

// Register
let newUser = try await client.auth.register(params: CreateUserParams(
    email: "new@example.com",
    displayName: "New User",
    password: "secure-password",
    roleID: RoleID("01HROLE...")
))

// Password reset flow
let msg = try await client.auth.requestPasswordReset(params: RequestPasswordResetParams(
    email: "user@example.com"
))
let confirm = try await client.auth.confirmPasswordReset(params: ConfirmPasswordResetParams(
    token: "reset-token",
    password: "new-password"
))

// Logout
try await client.auth.logout()
```

## SwiftUI Integration

### Observable ViewModel

```swift
import SwiftUI
import Modula

@Observable
final class ContentViewModel {
    var datatypes: [Datatype] = []
    var isLoading = false
    var error: Error?

    private let client: ModulaClient

    init(client: ModulaClient) {
        self.client = client
    }

    func loadDatatypes() async {
        isLoading = true
        defer { isLoading = false }

        do {
            datatypes = try await client.datatypes.list()
            error = nil
        } catch {
            self.error = error
        }
    }
}
```

### View with Task Loading

```swift
struct DatatypeListView: View {
    let viewModel: ContentViewModel

    var body: some View {
        List(viewModel.datatypes, id: \.datatypeID) { dt in
            VStack(alignment: .leading) {
                Text(dt.label)
                    .font(.headline)
                Text(dt.name)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
        .overlay {
            if viewModel.isLoading {
                ProgressView()
            }
        }
        .task {
            await viewModel.loadDatatypes()
        }
    }
}
```

### Paginated Loading

```swift
@Observable
final class PaginatedViewModel<T: Decodable & Sendable> {
    var items: [T] = []
    var total: Int64 = 0
    var isLoading = false
    private var offset: Int64 = 0
    private let pageSize: Int64 = 20

    func loadPage<ID: ResourceID>(
        from resource: Resource<T, some Encodable & Sendable, some Encodable & Sendable, ID>
    ) async throws {
        isLoading = true
        defer { isLoading = false }

        let page = try await resource.listPaginated(params: PaginationParams(
            limit: pageSize,
            offset: offset
        ))
        items.append(contentsOf: page.data)
        total = page.total
        offset += pageSize
    }

    var hasMore: Bool { Int64(items.count) < total }
}
```

## Next Steps

- [Error Handling](error-handling.md) -- APIError, do/catch patterns, network vs API errors
- [Reference](reference.md) -- all resources, ID types, enums, JSONValue
