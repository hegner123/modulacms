# Error Handling

Handle API errors from the Go SDK using the `*ApiError` type, classification helpers, and recovery patterns.

## The ApiError Type

The SDK wraps every non-2xx HTTP response in an `*ApiError`:

```go
type ApiError struct {
    StatusCode int    // HTTP status code (e.g., 404, 401, 500)
    Message    string // Server-provided error message from JSON response
    Body       string // Raw response body for debugging
}
```

The `Error()` method returns a formatted string:

```
modula: 404 content not found
modula: 401 Unauthorized
```

The SDK extracts `Message` from the JSON response body's `"message"` or `"error"` field. If the response is not JSON or lacks these fields, `Message` is empty and `Error()` falls back to the standard HTTP status text.

## Extracting ApiError

Use `errors.As` to extract the concrete error type:

```go
item, err := client.ContentData.Get(ctx, contentID)
if err != nil {
    var apiErr *modula.ApiError
    if errors.As(err, &apiErr) {
        fmt.Printf("HTTP %d: %s\n", apiErr.StatusCode, apiErr.Message)
        fmt.Printf("Raw body: %s\n", apiErr.Body)
    }
    return err
}
```

The SDK wraps errors with additional context (e.g., `"query blog-posts: modula: 404 Not Found"`), so always use `errors.As` rather than type assertion.

## Classification Helpers

The SDK provides four convenience functions that check the error type and status code. These work on wrapped errors.

### IsNotFound

```go
func IsNotFound(err error) bool
```

Returns `true` if the error is an `*ApiError` with HTTP status 404. Use after `Get` or `Delete` calls:

```go
item, err := client.ContentData.Get(ctx, id)
if modula.IsNotFound(err) {
    // The content item does not exist.
    return nil
}
if err != nil {
    return fmt.Errorf("fetch content: %w", err)
}
```

### IsUnauthorized

```go
func IsUnauthorized(err error) bool
```

Returns `true` if the error is an `*ApiError` with HTTP status 401. Indicates a missing, expired, or invalid API key or session:

```go
me, err := client.Auth.Me(ctx)
if modula.IsUnauthorized(err) {
    // Token expired or missing. Re-authenticate.
    return redirectToLogin()
}
```

### IsDuplicateMedia

```go
func IsDuplicateMedia(err error) bool
```

Returns `true` if the error is an `*ApiError` with HTTP status 409 (Conflict). The server returns this when a media upload duplicates an existing file:

```go
media, err := client.MediaUpload.Upload(ctx, file, "photo.jpg", nil)
if modula.IsDuplicateMedia(err) {
    // File already exists. Fetch the existing one instead.
    fmt.Println("Duplicate upload skipped")
    return nil
}
```

### IsInvalidMediaPath

```go
func IsInvalidMediaPath(err error) bool
```

Returns `true` if the error is an `*ApiError` with HTTP status 400 whose body mentions path traversal or invalid path characters. The server returns this when a media upload path contains `..` segments or disallowed characters:

```go
media, err := client.MediaUpload.Upload(ctx, file, "photo.jpg", &modula.MediaUploadOptions{
    Path: userProvidedPath,
})
if modula.IsInvalidMediaPath(err) {
    return fmt.Errorf("invalid upload path: %s", userProvidedPath)
}
```

## Error Handling Patterns

### Pattern: Classify Then Fall Through

Handle the most common cases explicitly and let everything else propagate:

```go
func getContent(ctx context.Context, client *modula.Client, id modula.ContentID) (*modula.ContentData, error) {
    item, err := client.ContentData.Get(ctx, id)
    if modula.IsNotFound(err) {
        return nil, nil // Caller checks for nil
    }
    if modula.IsUnauthorized(err) {
        return nil, fmt.Errorf("authentication required: %w", err)
    }
    if err != nil {
        return nil, fmt.Errorf("get content %s: %w", id, err)
    }
    return item, nil
}
```

### Pattern: Retry on Transient Errors

Server errors (5xx) are often transient. Retry with backoff:

```go
func getWithRetry(ctx context.Context, client *modula.Client, id modula.ContentID) (*modula.ContentData, error) {
    var lastErr error
    for attempt := range 3 {
        item, err := client.ContentData.Get(ctx, id)
        if err == nil {
            return item, nil
        }
        var apiErr *modula.ApiError
        if errors.As(err, &apiErr) && apiErr.StatusCode < 500 {
            return nil, err // Client errors are not retryable.
        }
        lastErr = err
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(time.Duration(attempt+1) * time.Second):
        }
    }
    return nil, fmt.Errorf("exhausted retries: %w", lastErr)
}
```

### Pattern: Context Cancellation

All SDK methods accept a `context.Context`. A cancelled context returns `context.Canceled`; a deadline-exceeded context returns `context.DeadlineExceeded`. These are not `*ApiError` values:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

items, err := client.ContentData.List(ctx)
if errors.Is(err, context.DeadlineExceeded) {
    log.Println("Request timed out")
}
```

### Pattern: Logging with Full Context

Use the `Body` field for detailed diagnostics:

```go
item, err := client.ContentData.Get(ctx, id)
if err != nil {
    var apiErr *modula.ApiError
    if errors.As(err, &apiErr) {
        slog.Error("API call failed",
            "status", apiErr.StatusCode,
            "message", apiErr.Message,
            "body", apiErr.Body,
            "content_id", id,
        )
    }
    return err
}
```

## Next Steps

- [Getting Started](/docs/sdks/go/getting-started) -- client setup and authentication
- [Content Operations](/docs/sdks/go/content-operations) -- CRUD, trees, publishing
- [Pagination](/docs/sdks/go/pagination) -- paginated listings
