# Media Download Endpoint Plan

Add a download endpoint that serves media files with the correct filename derived from the editable `display_name` field. Frontend developers use a `download_url` from the API response in an `<a href>`. CMS users control the download filename by editing `display_name` in any interface (TUI, admin panel, API, MCP).

## How It Works

```
1. CMS user uploads "92y3414168198264.pdf", sets display_name to "User Manual.pdf"
2. Frontend dev renders: <a href="{download_url}">Download</a>
3. User clicks link → GET /api/v1/media/{id}/download
4. Handler looks up media record, gets display_name "User Manual.pdf"
5. Handler generates pre-signed S3 URL with Content-Disposition: attachment; filename="User Manual.pdf"
6. Handler responds 302 redirect to pre-signed URL
7. Browser follows redirect, S3 serves file, save dialog shows "User Manual.pdf"
```

ModulaCMS never proxies the file bytes. The redirect is transparent to the user.

## Implementation

### 1. Download Endpoint

**Route:** `GET /api/v1/media/{id}/download`

**Permission:** `media:read` (viewing media implies downloading it)

**File:** `internal/router/media_download.go`

```go
func apiDownloadMedia(w http.ResponseWriter, r *http.Request, c config.Config) {
    d := db.ConfigDB(c)

    // Extract media ID from URL path
    rawID := r.PathValue("id")
    mediaID := types.MediaID(rawID)
    if err := mediaID.Validate(); err != nil {
        http.Error(w, "invalid media ID", http.StatusBadRequest)
        return
    }

    // Fetch media record
    m, err := d.GetMedia(mediaID)
    if err != nil {
        http.Error(w, "media not found", http.StatusNotFound)
        return
    }

    // Determine download filename: display_name > name > fallback from URL
    filename := filenameFromMedia(m)

    // Extract S3 key from stored URL
    // URL format: {BucketPublicURL}/{Bucket_Media}/{s3Key}
    s3Key := extractS3Key(string(m.URL), c)
    if s3Key == "" {
        http.Error(w, "unable to resolve storage key", http.StatusInternalServerError)
        return
    }

    // Generate pre-signed URL with Content-Disposition override
    creds := bucket.GetS3Creds(&c)
    s3Client, err := creds.GetBucket()
    if err != nil {
        http.Error(w, "storage unavailable", http.StatusServiceUnavailable)
        return
    }

    disposition := fmt.Sprintf(`attachment; filename="%s"`, sanitizeFilename(filename))

    req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
        Bucket:                     aws.String(c.Bucket_Media),
        Key:                        aws.String(s3Key),
        ResponseContentDisposition: aws.String(disposition),
    })

    presignedURL, err := req.Presign(15 * time.Minute)
    if err != nil {
        http.Error(w, "failed to generate download URL", http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, presignedURL, http.StatusFound)
}
```

**Helper functions** (same file):

```go
// filenameFromMedia returns the best available filename for download.
// Priority: display_name > name > last segment of URL.
func filenameFromMedia(m *db.Media) string {
    if m.DisplayName.Valid && m.DisplayName.String != "" {
        return m.DisplayName.String
    }
    if m.Name.Valid && m.Name.String != "" {
        return m.Name.String
    }
    // Fallback: extract filename from URL path
    u := string(m.URL)
    if idx := strings.LastIndex(u, "/"); idx >= 0 {
        return u[idx+1:]
    }
    return "download"
}

// extractS3Key strips the public URL prefix and bucket name to recover the S3 object key.
// Stored URL format: {BucketPublicURL}/{Bucket_Media}/{s3Key}
func extractS3Key(storedURL string, c config.Config) string {
    prefix := c.BucketPublicURL() + "/" + c.Bucket_Media + "/"
    if strings.HasPrefix(storedURL, prefix) {
        return storedURL[len(prefix):]
    }
    // Fallback: try endpoint URL (in case public URL differs)
    prefix = c.BucketEndpointURL() + "/" + c.Bucket_Media + "/"
    if strings.HasPrefix(storedURL, prefix) {
        return storedURL[len(prefix):]
    }
    return ""
}

// sanitizeFilename removes characters that are unsafe in Content-Disposition headers.
func sanitizeFilename(name string) string {
    // Replace characters that break Content-Disposition header parsing
    r := strings.NewReplacer(`"`, "'", "\n", "", "\r", "")
    return r.Replace(name)
}
```

### 2. Route Registration

**File:** `internal/router/mux.go`

Add before the existing `/api/v1/media/` catch-all:

```go
mux.Handle("GET /api/v1/media/{id}/download",
    middleware.RequirePermission("media:read")(
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            apiDownloadMedia(w, r, *c)
        }),
    ),
)
```

Order matters: this specific route must be registered before the `/api/v1/media/` handler that dispatches GET/PUT/DELETE by method, or the catch-all will swallow it. Go 1.22+ ServeMux matches more specific patterns first, so `GET /api/v1/media/{id}/download` will match before `/api/v1/media/`.

### 3. Media Response Wrapper

**File:** `internal/router/media_response.go`

```go
// MediaResponse wraps db.Media with computed fields for API responses.
type MediaResponse struct {
    db.Media
    DownloadURL string `json:"download_url"`
}

// toMediaResponse adds the download_url field to a media record.
func toMediaResponse(m db.Media) MediaResponse {
    return MediaResponse{
        Media:       m,
        DownloadURL: "/api/v1/media/" + string(m.MediaID) + "/download",
    }
}

// toMediaListResponse wraps a slice of media records.
func toMediaListResponse(items []db.Media) []MediaResponse {
    resp := make([]MediaResponse, len(items))
    for i, m := range items {
        resp[i] = toMediaResponse(m)
    }
    return resp
}
```

`download_url` is a relative path. The client already knows the API base URL since they're calling it. This avoids scheme/host/port guessing.

### 4. Update Existing API Handlers

**File:** `internal/router/media.go`

Change handlers that return media JSON to use the wrapper:

| Handler | Current | Change |
|---------|---------|--------|
| `apiGetMedia` | `json.NewEncoder(w).Encode(media)` | `json.NewEncoder(w).Encode(toMediaResponse(*media))` |
| `apiListMedia` | `json.NewEncoder(w).Encode(media)` | `json.NewEncoder(w).Encode(toMediaListResponse(*media))` |
| `apiListMediaPaginated` | `json.NewEncoder(w).Encode(media)` | `json.NewEncoder(w).Encode(toMediaListResponse(*media))` |
| `apiCreateMedia` | `json.NewEncoder(w).Encode(row)` | `json.NewEncoder(w).Encode(toMediaResponse(*row))` |

No changes needed to admin panel handlers (they render HTML, not JSON).

### 5. Update MCP Tools

**File:** `mcp/tools_media.go`

MCP tool responses should also include `download_url`. The MCP server has access to media records. Add the same `download_url` field to MCP responses for `get_media`, `list_media`, and `upload_media`.

Since MCP tools construct their own response maps, add the field inline:

```go
"download_url": "/api/v1/media/" + string(m.MediaID) + "/download",
```

### 6. Update SDKs

**TypeScript (`@modulacms/types`):**
```typescript
// Add to Media type
download_url: string;
```

**Go SDK (`sdks/go/types.go`):**
```go
// Add to Media struct
DownloadURL string `json:"download_url"`
```

**Swift SDK (`Types.swift`):**
```swift
// Add to Media struct
let downloadURL: String
// CodingKeys:
case downloadURL = "download_url"
```

### 7. Admin Panel Download Button

**File:** `internal/admin/pages/media_detail.templ`

Add a download link/button on the media detail page that uses the download endpoint:

```html
<a href={ templ.SafeURL("/api/v1/media/" + string(media.MediaID) + "/download") }
   class="btn btn-secondary" download>
   Download
</a>
```

The `download` attribute on the `<a>` tag hints to the browser that this is a download, not navigation. The server's `Content-Disposition: attachment` header from S3 does the actual work.

---

## Edge Cases

| Case | Behavior |
|------|----------|
| No display_name or name set | Falls back to last segment of S3 URL |
| display_name contains quotes | `sanitizeFilename` replaces `"` with `'` |
| display_name contains unicode | Works -- Content-Disposition supports UTF-8 via RFC 6266 |
| S3 credentials not configured | Returns 503 Service Unavailable |
| Media record exists but S3 object deleted | S3 returns 404 after redirect (expected -- same as direct URL) |
| Pre-signed URL expires (15 min) | User re-clicks link, gets fresh pre-signed URL |
| Bucket_Public_URL differs from Bucket_Endpoint | `extractS3Key` tries both prefixes |

## Files Changed

| File | Change |
|------|--------|
| `internal/router/media_download.go` | New: download endpoint handler + helpers |
| `internal/router/media_response.go` | New: MediaResponse wrapper + conversion functions |
| `internal/router/media.go` | Modified: wrap JSON responses with toMediaResponse |
| `internal/router/mux.go` | Modified: register download route |
| `mcp/tools_media.go` | Modified: add download_url to responses |
| `internal/admin/pages/media_detail.templ` | Modified: add download button |
| `sdks/typescript/types/src/media.ts` | Modified: add download_url field |
| `sdks/go/types.go` | Modified: add DownloadURL field |
| `sdks/swift/Sources/ModulaCMS/Types.swift` | Modified: add downloadURL field |

## Not In Scope

- **Setting Content-Disposition at upload time**: Would make direct S3 URLs work but is stale when display_name changes. The pre-signed URL approach always uses the current display_name. Can be added later as a bonus.
- **Inline viewing (Content-Disposition: inline)**: Could add a `?inline=true` query param to the download endpoint that uses `inline` instead of `attachment`. Not needed for initial implementation.
- **Access control on downloads**: Currently uses same `media:read` permission as viewing metadata. If per-file access control is needed later, the download endpoint is the natural place to add it.
