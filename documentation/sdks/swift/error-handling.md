# Swift SDK -- Error Handling

Handle API errors and network failures from the Swift SDK's `async throws` methods.

## APIError

The SDK throws `APIError` for non-2xx HTTP responses:

```swift
public struct APIError: Error, LocalizedError, Sendable {
    public let statusCode: Int
    public let message: String
    public let body: String
}
```

| Field | Type | Description |
|-------|------|-------------|
| `statusCode` | `Int` | HTTP status code (e.g., 404, 401, 500). Set to 0 for client-side errors (invalid URL, empty base URL). |
| `message` | `String` | Extracted from the response JSON `message` or `error` field. Empty if the response is not JSON or has no message. |
| `body` | `String` | Raw response body as a string. Useful for debugging when `message` is empty. |

`APIError` conforms to `LocalizedError`. The `errorDescription` property returns a formatted string:

```
modula: 404 Not Found
modula: 401 Invalid credentials
```

## Error Helper Functions

The SDK provides free functions for common status checks:

```swift
public func isNotFound(_ error: Error) -> Bool
public func isUnauthorized(_ error: Error) -> Bool
public func isDuplicateMedia(_ error: Error) -> Bool
public func isInvalidMediaPath(_ error: Error) -> Bool
```

These safely downcast the error to `APIError` and check the status code (and body content for media-specific checks). They return `false` for non-`APIError` errors.

## Do/Catch Patterns

### Basic Error Handling

```swift
do {
    let datatype = try await client.datatypes.get(id: DatatypeID("01HXYZ..."))
    print(datatype.label)
} catch {
    print("Failed: \(error.localizedDescription)")
}
```

### Checking Specific Error Types

```swift
do {
    let datatype = try await client.datatypes.get(id: DatatypeID("nonexistent"))
} catch let error where isNotFound(error) {
    print("Datatype does not exist")
} catch let error where isUnauthorized(error) {
    print("Not authenticated -- redirect to login")
} catch let apiError as APIError {
    print("API error \(apiError.statusCode): \(apiError.message)")
    if !apiError.body.isEmpty {
        print("Response body: \(apiError.body)")
    }
} catch {
    print("Network or other error: \(error)")
}
```

### Pattern Matching on Status Code

```swift
do {
    try await client.datatypes.delete(id: datatypeID)
} catch let error as APIError {
    switch error.statusCode {
    case 400:
        print("Bad request: \(error.message)")
    case 403:
        print("Forbidden -- insufficient permissions")
    case 404:
        print("Already deleted or never existed")
    case 409:
        print("Conflict -- resource is in use")
    case 500:
        print("Server error: \(error.message)")
    default:
        print("Unexpected status \(error.statusCode): \(error.message)")
    }
} catch {
    print("Network error: \(error)")
}
```

## Network Errors vs API Errors

Non-`APIError` errors are typically `URLError` from the networking layer.

```swift
do {
    let datatypes = try await client.datatypes.list()
} catch is URLError {
    // Network-level failure: no connectivity, DNS resolution failed,
    // connection refused, TLS handshake failed, etc.
    print("Network error -- check connectivity")
} catch let apiError as APIError where apiError.statusCode == 0 {
    // Client-side error from the SDK itself:
    // invalid URL construction, empty base URL, invalid response type
    print("Client error: \(apiError.message)")
} catch let apiError as APIError {
    // Server responded with a non-2xx status
    print("Server error \(apiError.statusCode)")
} catch {
    // Decoding errors (DecodingError) or other unexpected failures
    print("Unexpected: \(error)")
}
```

Common `URLError` codes:

| Code | Meaning |
|------|---------|
| `.notConnectedToInternet` | No network connectivity |
| `.timedOut` | Request exceeded timeout |
| `.cannotFindHost` | DNS resolution failed |
| `.cannotConnectToHost` | Connection refused |
| `.secureConnectionFailed` | TLS/SSL failure |
| `.cancelled` | Request was cancelled (e.g., task cancellation) |

## Timeout Handling

The default `URLSession` configuration uses:
- 30 seconds for request timeout (`timeoutIntervalForRequest`)
- 300 seconds for resource timeout (`timeoutIntervalForResource`)

Timeouts throw `URLError(.timedOut)`:

```swift
do {
    let data = try await client.content.getPage(slug: "large-page")
} catch let error as URLError where error.code == .timedOut {
    print("Request timed out")
} catch {
    print("Other error: \(error)")
}
```

To customize timeouts, pass a configured `URLSession` to `ClientConfig`:

```swift
let config = URLSessionConfiguration.default
config.timeoutIntervalForRequest = 10  // 10 seconds

let client = try ModulaClient(config: ClientConfig(
    baseURL: "https://cms.example.com",
    apiKey: "token",
    urlSession: URLSession(configuration: config)
))
```

## Cancellation

Swift structured concurrency propagates cancellation to in-flight URL requests. Cancelling a `Task` causes the underlying `URLSession.data(for:)` call to throw `CancellationError` or `URLError(.cancelled)`.

```swift
let task = Task {
    do {
        let datatypes = try await client.datatypes.list()
        // process...
    } catch is CancellationError {
        print("Task was cancelled")
    } catch let error as URLError where error.code == .cancelled {
        print("Request was cancelled")
    }
}

task.cancel()
```

## Decoding Errors

When the server returns a 2xx status but the response body doesn't match the expected type, the SDK throws `DecodingError`:

```swift
do {
    let datatype = try await client.datatypes.get(id: someID)
} catch let error as DecodingError {
    switch error {
    case .keyNotFound(let key, _):
        print("Missing key: \(key.stringValue)")
    case .typeMismatch(let type, let context):
        print("Type mismatch for \(type) at \(context.codingPath)")
    default:
        print("Decoding failed: \(error)")
    }
}
```

This typically indicates a version mismatch between the SDK and the CMS server.

## Media-Specific Errors

Media upload has two additional helper functions:

```swift
do {
    let media = try await client.mediaUpload.upload(
        data: fileData,
        filename: "photo.jpg",
        options: .init(path: "blog/images")
    )
} catch let error where isDuplicateMedia(error) {
    // HTTP 409 -- a file with this name already exists at the path
    print("Duplicate media file")
} catch let error where isInvalidMediaPath(error) {
    // HTTP 400 with "path traversal" or "invalid character in path"
    print("Invalid upload path")
} catch {
    print("Upload failed: \(error)")
}
```

## Retry Strategy

The SDK doesn't include built-in retry logic. Implement retries in your application layer:

```swift
func withRetry<T>(
    maxAttempts: Int = 3,
    delay: Duration = .seconds(1),
    operation: () async throws -> T
) async throws -> T {
    var lastError: Error?
    for attempt in 1...maxAttempts {
        do {
            return try await operation()
        } catch let error as URLError where error.code == .timedOut || error.code == .notConnectedToInternet {
            lastError = error
            if attempt < maxAttempts {
                try await Task.sleep(for: delay * attempt)
            }
        } catch let error as APIError where error.statusCode >= 500 {
            lastError = error
            if attempt < maxAttempts {
                try await Task.sleep(for: delay * attempt)
            }
        }
    }
    throw lastError!
}

// Usage
let datatypes = try await withRetry {
    try await client.datatypes.list()
}
```

## Next Steps

- [Getting Started](/docs/sdks/swift/getting-started) -- installation, client setup, first requests
- [Content Operations](/docs/sdks/swift/content-operations) -- CRUD, content delivery, trees, media
- [Reference](/docs/sdks/swift/reference) -- all resources, ID types, enums
