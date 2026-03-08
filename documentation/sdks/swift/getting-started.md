# Swift SDK -- Getting Started

The Modula Swift SDK provides a type-safe async/await client for the ModulaCMS REST API. It targets Apple platforms with zero external dependencies.

## Platform Requirements

| Requirement | Minimum |
|-------------|---------|
| iOS | 16.0+ |
| macOS | 13.0+ |
| tvOS | 16.0+ |
| watchOS | 9.0+ |
| Swift | 5.9+ |
| Xcode | 15.0+ |

The SDK uses Swift concurrency (`async`/`await`), `Sendable` conformance throughout, and `URLSession` for networking. No third-party dependencies.

## Installation

Add the package to your `Package.swift`:

```swift
dependencies: [
    .package(url: "https://github.com/hegner123/modulacms.git", from: "0.1.0"),
]
```

Then add the `Modula` target to your dependency list:

```swift
.target(
    name: "YourApp",
    dependencies: [
        .product(name: "Modula", package: "modulacms"),
    ]
)
```

In Xcode: File > Add Package Dependencies, paste the repository URL, and select the `Modula` library.

## Creating a Client

All API access goes through `ModulaClient`. Initialize it with a `ClientConfig`:

```swift
import Modula

let client = try ModulaClient(config: ClientConfig(
    baseURL: "https://cms.example.com",
    apiKey: "your-api-token"
))
```

`ClientConfig` fields:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `baseURL` | `String` | Yes | CMS server URL. Trailing slashes are stripped automatically. |
| `apiKey` | `String` | No | Bearer token for API authentication. Defaults to empty string. |
| `urlSession` | `URLSession?` | No | Custom URL session. Defaults to a session with 30s request timeout, 300s resource timeout, and cookies disabled. |

The initializer throws if `baseURL` is empty.

### Custom URLSession

Pass your own `URLSession` for proxy configuration, certificate pinning, or custom caching:

```swift
let sessionConfig = URLSessionConfiguration.default
sessionConfig.timeoutIntervalForRequest = 10
sessionConfig.waitsForConnectivity = true

let client = try ModulaClient(config: ClientConfig(
    baseURL: "https://cms.example.com",
    apiKey: "your-api-token",
    urlSession: URLSession(configuration: sessionConfig)
))
```

## First Requests

All SDK methods are `async throws`. Use them inside async contexts.

### List Datatypes

```swift
let datatypes = try await client.datatypes.list()
for dt in datatypes {
    print("\(dt.name): \(dt.label)")
}
```

### Get a Single Datatype

```swift
let id = DatatypeID("01HXYZ...")
let datatype = try await client.datatypes.get(id: id)
print(datatype.label)
```

### Authenticate and Get Current User

```swift
let loginResponse = try await client.auth.login(params: LoginParams(
    email: "admin@example.com",
    password: "your-password"
))
print("Token: \(loginResponse.token)")

let me = try await client.auth.me()
print("Logged in as: \(me.email)")
```

### Fetch Content by Slug

```swift
let pageData = try await client.content.getPage(slug: "blog/hello-world")
// pageData is raw Data -- decode as needed for your frontend
```

### Query Content by Datatype

```swift
let result = try await client.query.query(
    datatype: "blog_post",
    params: QueryParams(sort: "-date_created", limit: 10, status: "published")
)
print("Found \(result.total) posts")
for item in result.data {
    print(item.fields["title"] ?? "untitled")
}
```

## Async/Await Patterns

Every method on `ModulaClient` is `async throws`. Standard Swift concurrency patterns apply.

### Sequential Requests

```swift
let datatypes = try await client.datatypes.list()
let fields = try await client.fields.list()
```

### Concurrent Requests

Use `async let` for independent requests:

```swift
async let datatypes = client.datatypes.list()
async let users = client.users.list()
async let routes = client.routes.list()

let (dt, u, r) = try await (datatypes, users, routes)
```

### TaskGroup for Dynamic Concurrency

```swift
let ids: [ContentID] = [...]

let results = try await withThrowingTaskGroup(of: ContentData.self) { group in
    for id in ids {
        group.addTask {
            try await client.contentData.get(id: id)
        }
    }
    var items: [ContentData] = []
    for try await item in group {
        items.append(item)
    }
    return items
}
```

### Cancellation

All SDK methods respect Swift's cooperative cancellation. Cancel a parent `Task` and in-flight `URLSession` requests are cancelled automatically.

```swift
let task = Task {
    let data = try await client.datatypes.list()
    // ...
}

// Later:
task.cancel()
```

## Sendable Safety

`ModulaClient` and all resource types are `Sendable`. The client is safe to share across actors and tasks without additional synchronization.

```swift
actor ContentManager {
    let client: ModulaClient

    init(client: ModulaClient) {
        self.client = client
    }

    func fetchDatatypes() async throws -> [Datatype] {
        try await client.datatypes.list()
    }
}
```

## Next Steps

- [Content Operations](content-operations.md) -- CRUD, content delivery, trees, media upload, query API
- [Error Handling](error-handling.md) -- APIError, do/catch patterns, network vs API errors
- [Reference](reference.md) -- all resources, ID types, enums, JSONValue
