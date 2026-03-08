# Error Handling

The TypeScript SDKs use two error types: `ApiError` from `@modulacms/types` (used by the admin SDK) and `ModulaError` from `@modulacms/sdk` (used by the read-only SDK). Both represent server-side HTTP errors. Network failures surface as standard `TypeError` exceptions from the fetch API.

## ApiError (Admin SDK)

The admin SDK throws `ApiError` objects for non-2xx HTTP responses:

```typescript
type ApiError = {
  readonly _tag: 'ApiError'
  status: number
  message: string
  body?: unknown
}
```

| Field | Description |
|-------|-------------|
| `_tag` | Discriminant tag. Always `'ApiError'`. |
| `status` | HTTP status code (e.g. `404`, `403`, `500`). |
| `message` | HTTP status text (e.g. `'Not Found'`). |
| `body` | Parsed JSON response body, if the server returned `application/json`. |

`ApiError` is a plain object, not an `Error` subclass. Use the `isApiError` type guard to narrow caught values:

```typescript
import { isApiError } from '@modulacms/types'

try {
  const user = await client.users.get(userId)
} catch (err) {
  if (isApiError(err)) {
    console.error(`API error ${err.status}: ${err.message}`)
    if (err.body) {
      console.error('Response body:', err.body)
    }
  }
}
```

### isApiError

```typescript
function isApiError(err: unknown): err is ApiError
```

Returns `true` if `err` is an object with `_tag === 'ApiError'`. Returns `false` for network errors, timeouts, and non-SDK exceptions.

## ModulaError (Read-Only SDK)

The read-only SDK throws `ModulaError` instances, which extend the standard `Error` class:

```typescript
class ModulaError extends Error {
  status: number
  body: unknown
  get errorMessage(): string | undefined
}
```

| Field | Description |
|-------|-------------|
| `status` | HTTP status code. |
| `body` | Parsed JSON response body, if available. |
| `message` | Inherited from `Error`. Contains a description of the failure. |
| `errorMessage` | Getter that extracts a string message from `body`, if present. |

Use `instanceof` to catch:

```typescript
import { ModulaError } from '@modulacms/sdk'

try {
  const page = await client.getPage('about')
} catch (err) {
  if (err instanceof ModulaError) {
    console.error(`Status ${err.status}: ${err.errorMessage ?? err.message}`)
  }
}
```

## Error Patterns

### Checking Status Codes

```typescript
try {
  const item = await client.contentData.get(contentId)
} catch (err) {
  if (isApiError(err)) {
    switch (err.status) {
      case 404:
        // Resource not found
        break
      case 403:
        // Insufficient permissions
        break
      case 409:
        // Conflict (e.g. duplicate)
        break
      case 422:
        // Validation error -- check err.body for details
        break
      default:
        // Server error
        break
    }
  }
}
```

### Not Found Pattern

```typescript
async function findContent(id: ContentID): Promise<ContentData | null> {
  try {
    return await client.contentData.get(id)
  } catch (err) {
    if (isApiError(err) && err.status === 404) {
      return null
    }
    throw err
  }
}
```

### Network Errors

Network failures (DNS resolution, connection refused, TLS errors) throw standard `TypeError` from the fetch API. These are distinct from API errors:

```typescript
try {
  const data = await client.users.list()
} catch (err) {
  if (isApiError(err)) {
    // Server returned an error response
  } else if (err instanceof TypeError) {
    // Network failure -- server unreachable
  } else if (err instanceof DOMException && err.name === 'AbortError') {
    // Request was cancelled (timeout or manual abort)
  }
}
```

### Timeout Handling

The admin SDK applies `AbortSignal.timeout(defaultTimeout)` to every request (default 30 seconds). You can pass a custom signal per request:

```typescript
const controller = new AbortController()
setTimeout(() => controller.abort(), 5000) // 5 second timeout

try {
  const data = await client.users.list({ signal: controller.signal })
} catch (err) {
  if (err instanceof DOMException && err.name === 'AbortError') {
    console.error('Request timed out')
  }
}
```

When both the default timeout signal and a custom signal are provided, either one aborting cancels the request.

## Media Upload Errors

The admin SDK provides helper functions for common media upload errors:

```typescript
import {
  isDuplicateMedia,
  isFileTooLarge,
  isInvalidMediaPath,
} from '@modulacms/admin-sdk'

try {
  await client.mediaUpload.upload(file, { path: 'uploads' })
} catch (err) {
  if (isDuplicateMedia(err)) {
    // HTTP 409 -- file already exists
  } else if (isFileTooLarge(err)) {
    // HTTP 400 -- file exceeds size limit
  } else if (isInvalidMediaPath(err)) {
    // HTTP 400 -- path contains traversal or invalid characters
  }
}
```

## Void Responses

DELETE operations and some POST operations return `void` on success. On failure, they throw `ApiError` just like other methods:

```typescript
try {
  await client.users.remove(userId)
  // Success -- no return value
} catch (err) {
  if (isApiError(err) && err.status === 403) {
    console.error('Cannot delete system-protected user')
  }
}
```
