# Go SDK -- Getting Started

The ModulaCMS Go SDK provides typed access to the entire CMS API from any Go application. Zero dependencies, compile-time type safety via branded IDs and generics, and a single client instance that is safe for concurrent use.

## Installation

```bash
go get github.com/hegner123/modulacms/sdks/go
```

Import path:

```go
import modula "github.com/hegner123/modulacms/sdks/go"
```

The package name is `modula`. All types and functions are accessed as `modula.Client`, `modula.NewClient`, etc.

Requires Go 1.25 or later.

## Creating a Client

All API access goes through a `*modula.Client` returned by `modula.NewClient`:

```go
client, err := modula.NewClient(modula.ClientConfig{
    BaseURL: "https://cms.example.com",
    APIKey:  "your-api-key",
})
if err != nil {
    log.Fatal(err)
}
```

### ClientConfig Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `BaseURL` | `string` | Yes | Root URL of the ModulaCMS server, including scheme. Trailing slash is stripped automatically. |
| `APIKey` | `string` | No | Bearer token sent in the `Authorization` header on every request. Leave empty for unauthenticated access to public endpoints. |
| `HTTPClient` | `*http.Client` | No | Custom HTTP client for TLS settings, proxies, or transport middleware. When nil, a default client with a 30-second timeout is used. |

`NewClient` validates the configuration and returns an error if `BaseURL` is empty or malformed.

## First Requests

### Health Check

Verify the server is reachable:

```go
ctx := context.Background()

health, err := client.Health.Check(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Status: %s\n", health.Status)
```

### List Datatypes

Fetch all datatypes defined in the CMS:

```go
datatypes, err := client.Datatypes.List(ctx)
if err != nil {
    log.Fatal(err)
}
for _, dt := range datatypes {
    fmt.Printf("%s (%s)\n", dt.Label, dt.Name)
}
```

### Get a Single Item

Fetch a content item by its typed ID:

```go
item, err := client.ContentData.Get(ctx, contentID)
if err != nil {
    if modula.IsNotFound(err) {
        fmt.Println("Content not found")
        return
    }
    log.Fatal(err)
}
fmt.Printf("Status: %s, Created: %s\n", item.Status, item.DateCreated)
```

### Paginated Listing

```go
page, err := client.Users.ListPaginated(ctx, modula.PaginationParams{
    Limit:  25,
    Offset: 0,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Showing %d of %d users\n", len(page.Data), page.Total)
```

## Authentication

### API Key

The simplest approach. Set `APIKey` in `ClientConfig` and all requests include a `Bearer` token:

```go
client, err := modula.NewClient(modula.ClientConfig{
    BaseURL: "https://cms.example.com",
    APIKey:  "tok_abc123",
})
```

API keys are created via the Tokens resource or the admin panel. They inherit the permissions of the user they belong to.

### Session-Based Login

For interactive applications, authenticate with username and password:

```go
// Create an unauthenticated client first.
client, err := modula.NewClient(modula.ClientConfig{
    BaseURL: "https://cms.example.com",
})
if err != nil {
    log.Fatal(err)
}

// Log in. The server returns a session cookie handled by the HTTP client.
resp, err := client.Auth.Login(ctx, modula.LoginParams{
    Email:    "admin@example.com",
    Password: "secret",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Logged in as user %s\n", resp.UserID)

// Subsequent requests use the session cookie.
me, err := client.Auth.Me(ctx)
```

To use session cookies, supply a custom `http.Client` with a cookie jar:

```go
jar, _ := cookiejar.New(nil)
client, err := modula.NewClient(modula.ClientConfig{
    BaseURL:    "https://cms.example.com",
    HTTPClient: &http.Client{Jar: jar, Timeout: 30 * time.Second},
})
```

### Checking Current User

```go
user, err := client.Auth.Me(ctx)
if modula.IsUnauthorized(err) {
    fmt.Println("Not authenticated")
    return
}
fmt.Printf("Authenticated as %s (%s)\n", user.Username, user.Email)
```

## Next Steps

- [Content Operations](content-operations.md) -- CRUD, trees, publishing, queries
- [Error Handling](error-handling.md) -- error classification patterns
- [Pagination](pagination.md) -- paginated listings and iteration
- [Reference](reference.md) -- full resource index
