# Media Package Fix Plan

**Created:** 2026-01-18
**Status:** Ready for Implementation
**Priority:** High - Launch Blocker Issues Present
**Estimated Effort:** 2-3 days

---

## Executive Summary

The media package has **23 identified issues** including 6 critical bugs that affect core functionality. The most severe issues are:
- Temp directory cleanup never executes (disk space leak)
- File path construction bugs (incorrect S3 paths)
- Files created in wrong directory (cleanup fails)
- No transaction handling (data inconsistency)
- Error swallowing in S3 upload (silent failures)

This plan addresses all issues in 4 phases: Critical Fixes, High Priority, Code Quality, and Testing.

---

## Phase 1: Critical Fixes (Launch Blockers)

**Priority:** P0 - Must fix before any production use
**Estimated Time:** 4-6 hours

### Issue #3: Temp Directory Cleanup Never Executes

**File:** `internal/router/mediaUpload.go:67`

**Current Code:**
```go
defer exec.Command("rm", "-r", tmp)
```

**Problem:** `exec.Command()` creates a command object but doesn't execute it. Temp directories accumulate forever.

**Fix:**
```go
defer os.RemoveAll(tmp)
```

**Testing:**
1. Upload image
2. Check filesystem - temp directory should be deleted
3. Upload 100 images, verify no temp directories remain

**Files Modified:** `internal/router/mediaUpload.go`

---

### Issue #2: Temp Directory in CWD

**File:** `internal/router/mediaUpload.go:61`

**Current Code:**
```go
tmp, err := os.MkdirTemp(".", "temp")
```

**Problem:** Creates temp directories in application working directory, causing pollution and permission issues.

**Fix:**
```go
tmp, err := os.MkdirTemp("", "modulacms-media")
```

**Behavior:** Uses system temp directory (typically `/tmp` on Unix, `%TEMP%` on Windows)

**Testing:**
1. Upload image
2. Verify temp directory created in system temp location (not CWD)
3. Verify cleanup works across platforms

**Files Modified:** `internal/router/mediaUpload.go`

---

### Issue #1: Path Construction Bug

**File:** `internal/media/media_upload.go:48`

**Current Code:**
```go
for _, f := range *optimized {
    file, err := os.Open(f)
    if err != nil {
        return err
    }
    newPath := fmt.Sprintf("%s/%d/%d/%s", bucketDir, year, month, f)
    uploadPath := fmt.Sprintf("https://%s%s", c.Bucket_Endpoint, newPath)
    // ...
}
```

**Problem:** Variable `f` contains full path from `OptimizeUpload()`, not just filename. Results in paths like:
`media/2026/01//tmp/modulacms-media123456/photo-1920x1080.jpg`

**Fix:**
```go
for _, f := range *optimized {
    file, err := os.Open(f)
    if err != nil {
        return err
    }

    // Extract just the filename
    filename := filepath.Base(f)
    newPath := fmt.Sprintf("%s/%d/%d/%s", bucketDir, year, month, filename)
    uploadPath := fmt.Sprintf("https://%s%s", c.Bucket_Endpoint, newPath)

    prep, err := bucket.UploadPrep(newPath, c.Bucket_Media, file)
    if err != nil {
        return err
    }

    _, err = bucket.ObjectUpload(s3Session, prep)
    if err != nil {
        return err
    }
    srcset = append(srcset, uploadPath)
}
```

**Import Required:** Add `"path/filepath"` to imports

**Testing:**
1. Upload image
2. Verify S3 paths are `media/2026/01/photo-1920x1080.jpg` (not full temp path)
3. Verify srcset URLs in database are correct

**Files Modified:** `internal/media/media_upload.go`

---

### Issue #5: Files Created in Wrong Directory

**File:** `internal/media/media_optomize.go:102`

**Current Code:**
```go
for i, im := range images {
    widthString := strconv.FormatInt(dimensions[i].Width.Int64, 10)
    heightString := strconv.FormatInt(dimensions[i].Height.Int64, 10)
    size := widthString + "x" + heightString
    filename := fmt.Sprintf("%s-%v%s", baseName, size, ext)
    files = append(files, filename)
    f, err := os.Create(filename)  // ❌ Creates in CWD, not dstPath
    // ...
}
```

**Problem:** Creates optimized files in current working directory instead of using `dstPath` parameter. Makes cleanup impossible.

**Fix:**
```go
for i, im := range images {
    widthString := strconv.FormatInt(dimensions[i].Width.Int64, 10)
    heightString := strconv.FormatInt(dimensions[i].Height.Int64, 10)
    size := widthString + "x" + heightString
    filename := fmt.Sprintf("%s-%v%s", baseName, size, ext)

    // Create file in destination path
    fullPath := filepath.Join(dstPath, filename)
    files = append(files, fullPath)  // Store full path for return

    f, err := os.Create(fullPath)
    if err != nil {
        return nil, fmt.Errorf("error creating file %s: %w", fullPath, err)
    }
    defer f.Close()

    // Encoding logic...
}
```

**Impact:** This change affects `media_upload.go` which now receives full paths in the `optimized` array.

**Testing:**
1. Upload image
2. Verify optimized files created in temp directory
3. Verify files can be opened for S3 upload
4. Verify cleanup removes all optimized files

**Files Modified:** `internal/media/media_optomize.go`

---

### Issue #9: ObjectUpload Swallows Errors

**File:** `internal/bucket/object_storage.go:58-64`

**Current Code:**
```go
func ObjectUpload(s3 *s3.S3, payload *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
    upload, err := s3.PutObject(payload)
    if err != nil {
        utility.DefaultLogger.Error("failed to upload ", err)
    }
    return upload, nil  // ❌ Returns nil error even if upload failed
}
```

**Problem:** Logs error but returns `nil`, causing caller to think upload succeeded.

**Fix:**
```go
func ObjectUpload(s3 *s3.S3, payload *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
    upload, err := s3.PutObject(payload)
    if err != nil {
        return nil, fmt.Errorf("failed to upload to S3: %w", err)
    }
    return upload, nil
}
```

**Import Required:** Add `"fmt"` to imports

**Testing:**
1. Upload with valid credentials - should succeed
2. Upload with invalid credentials - should return error
3. Upload with network failure - should return error
4. Verify error propagates to HTTP handler

**Files Modified:** `internal/bucket/object_storage.go`

---

### Issue #4: Hard Dependency on db-sqlite

**File:** `internal/bucket/object_storage.go:11`

**Current Code:**
```go
mdb "github.com/hegner123/modulacms/internal/db-sqlite"
```

**Problem:** Violates database abstraction layer. Import appears unused (only referenced in empty `ParseMetaData` stub).

**Fix:**
```go
// Remove the import entirely
```

**Verification:** Search codebase for usage of `mdb` in bucket package.

**Testing:**
1. Run `make build` with MySQL driver
2. Run `make build` with PostgreSQL driver
3. Verify no compilation errors

**Files Modified:** `internal/bucket/object_storage.go`

---

### Issue #6: AuthorID Type Mismatch

**File:** `internal/router/mediaUpload.go:53-58`

**Current Code:**
```go
forms := db.CreateMediaFormParams{
    Name:     header.Filename,
    AuthorID: "1",  // ❌ String literal, should be int64
}
params := db.MapCreateMediaParams(forms)
```

**Problem:** Hardcoded string value for AuthorID. Should get from authenticated session.

**Fix (Temporary - Session Integration Later):**
```go
forms := db.CreateMediaFormParams{
    Name:     header.Filename,
    AuthorID: "1",  // TODO: Get from authenticated session
}
params := db.MapCreateMediaParams(forms)
```

**Fix (Proper - Requires Session Package):**
```go
// Get author from session
session, err := middleware.GetSession(r, c)
if err != nil {
    utility.DefaultLogger.Error("get session", err)
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}

forms := db.CreateMediaFormParams{
    Name:     header.Filename,
    AuthorID: fmt.Sprintf("%d", session.UserID),
}
params := db.MapCreateMediaParams(forms)
```

**Decision:** Add TODO comment for now, fix properly when session integration is implemented project-wide.

**Testing:**
1. Verify media records created with author_id = 1
2. After session integration: verify author_id matches logged-in user

**Files Modified:** `internal/router/mediaUpload.go`

---

## Phase 2: High Priority (Pre-Production)

**Priority:** P1 - Must fix before production deployment
**Estimated Time:** 6-8 hours

### Issue #7: No Transaction Handling

**Problem:** Database INSERT and S3 uploads are not atomic. If S3 upload fails after DB insert, database has orphaned records pointing to non-existent S3 files.

**Current Flow:**
```
1. Create media record in DB (with empty srcset)
2. Optimize images
3. Upload to S3
4. Update media record with srcset
   ↳ If this fails, S3 has files but DB doesn't reference them
```

**Better Flow:**
```
1. Validate filename uniqueness
2. Create temp directory
3. Optimize images
4. Upload ALL to S3 (track uploaded files)
5. If ALL uploads succeed:
   → Create media record in DB with srcset
6. If ANY upload fails:
   → Delete uploaded S3 files
   → Return error
```

**Implementation:**

**New File:** `internal/media/media_upload.go` (refactor)

```go
func HandleMediaUpload(srcFile string, dstPath string, c config.Config) error {
    d := db.ConfigDB(c)
    bucketDir := c.Bucket_Media
    now := time.Now()
    year := now.Year()
    month := now.Month()

    filename := filepath.Base(srcFile)
    baseName := strings.TrimSuffix(filename, filepath.Ext(filename))

    // Step 1: Optimize images
    optimized, err := OptimizeUpload(srcFile, dstPath, c)
    if err != nil {
        return fmt.Errorf("optimization failed: %w", err)
    }

    // Step 2: Upload ALL to S3 (track successes for rollback)
    s3Creds := bucket.S3Credentials{
        AccessKey: c.Bucket_Access_Key,
        SecretKey: c.Bucket_Secret_Key,
        URL:       c.Bucket_Endpoint,
    }

    s3Session, err := s3Creds.GetBucket()
    if err != nil {
        return fmt.Errorf("S3 session failed: %w", err)
    }

    srcset := []string{}
    uploadedKeys := []string{}  // Track for rollback

    for _, fullPath := range *optimized {
        file, err := os.Open(fullPath)
        if err != nil {
            // Rollback previous uploads
            rollbackS3Uploads(s3Session, c.Bucket_Media, uploadedKeys)
            return fmt.Errorf("failed to open optimized file: %w", err)
        }
        defer file.Close()

        filename := filepath.Base(fullPath)
        s3Key := fmt.Sprintf("%s/%d/%d/%s", bucketDir, year, month, filename)
        uploadPath := fmt.Sprintf("https://%s/%s", c.Bucket_Endpoint, s3Key)

        prep, err := bucket.UploadPrep(s3Key, c.Bucket_Media, file)
        if err != nil {
            rollbackS3Uploads(s3Session, c.Bucket_Media, uploadedKeys)
            return fmt.Errorf("upload prep failed: %w", err)
        }

        _, err = bucket.ObjectUpload(s3Session, prep)
        if err != nil {
            rollbackS3Uploads(s3Session, c.Bucket_Media, uploadedKeys)
            return fmt.Errorf("S3 upload failed: %w", err)
        }

        uploadedKeys = append(uploadedKeys, s3Key)
        srcset = append(srcset, uploadPath)
    }

    // Step 3: All uploads succeeded - update database
    srcsetJSON, err := json.Marshal(srcset)
    if err != nil {
        rollbackS3Uploads(s3Session, c.Bucket_Media, uploadedKeys)
        return fmt.Errorf("failed to marshal srcset: %w", err)
    }

    // Get existing media record and update srcset
    rowPtr, err := d.GetMediaByName(baseName)
    if err != nil {
        rollbackS3Uploads(s3Session, c.Bucket_Media, uploadedKeys)
        return fmt.Errorf("failed to get media record: %w", err)
    }

    row := *rowPtr
    params := MapMediaParams(row)
    params.Srcset = db.StringToNullString(string(srcsetJSON))

    _, err = d.UpdateMedia(params)
    if err != nil {
        rollbackS3Uploads(s3Session, c.Bucket_Media, uploadedKeys)
        return fmt.Errorf("database update failed: %w", err)
    }

    return nil
}

// Rollback helper - delete uploaded files from S3
func rollbackS3Uploads(s3Session *s3.S3, bucketName string, keys []string) {
    for _, key := range keys {
        _, err := s3Session.DeleteObject(&s3.DeleteObjectInput{
            Bucket: aws.String(bucketName),
            Key:    aws.String(key),
        })
        if err != nil {
            utility.DefaultLogger.Error(fmt.Sprintf("rollback failed for key %s", key), err)
        }
    }
}
```

**New Import Required:**
```go
"github.com/aws/aws-sdk-go/service/s3"
```

**Testing:**
1. Upload with valid S3 - all files uploaded, DB updated
2. Simulate S3 failure midway - verify rollback deletes uploaded files
3. Simulate DB update failure - verify rollback deletes S3 files
4. Check database - no orphaned records

**Files Modified:** `internal/media/media_upload.go`

---

### Issue #8: Partial Upload Failure Handling

**Status:** ✅ Solved by Issue #7 transaction handling

---

### Issue #10: No Cleanup of Optimized Files

**Status:** ✅ Solved by Issue #2 and #3 (temp directory cleanup)

---

### Issue #22: No Rollback on Optimization Failure

**Problem:** If `OptimizeUpload()` fails partway through generating dimensions, partial files remain.

**Fix:** Wrap optimization in cleanup logic:

```go
func OptimizeUpload(fSrc string, dstPath string, c config.Config) (*[]string, error) {
    // ... existing decode logic ...

    files := []string{}
    var optimizationErr error

    // Generate all optimized images
    for i, im := range images {
        widthString := strconv.FormatInt(dimensions[i].Width.Int64, 10)
        heightString := strconv.FormatInt(dimensions[i].Height.Int64, 10)
        size := widthString + "x" + heightString
        filename := fmt.Sprintf("%s-%v%s", baseName, size, ext)
        fullPath := filepath.Join(dstPath, filename)

        f, err := os.Create(fullPath)
        if err != nil {
            optimizationErr = fmt.Errorf("error creating file %s: %w", fullPath, err)
            break
        }

        // Encode image
        switch ext {
        case ".png":
            err = png.Encode(f, im)
        case ".jpg", ".jpeg":
            err = jpeg.Encode(f, im, nil)
        case ".gif":
            err = gif.Encode(f, im, nil)
        default:
            err = fmt.Errorf("unsupported encoding for extension: %s", ext)
        }

        f.Close()

        if err != nil {
            optimizationErr = fmt.Errorf("error encoding image %s: %w", filename, err)
            break
        }

        files = append(files, fullPath)
    }

    // If any optimization failed, clean up partial files
    if optimizationErr != nil {
        for _, file := range files {
            os.Remove(file)
        }
        return nil, optimizationErr
    }

    return &files, nil
}
```

**Testing:**
1. Simulate encoding failure (invalid format)
2. Verify partial files are deleted
3. Verify error is returned

**Files Modified:** `internal/media/media_optomize.go`

---

## Phase 3: Security Fixes

**Priority:** P1 - Security vulnerabilities
**Estimated Time:** 3-4 hours

### Issue #11: No Image Dimension Validation

**Problem:** No check on source image dimensions before loading into memory. Attacker can upload massive images (100000x100000 pixels) causing memory exhaustion.

**Fix:**

Add validation constants:
```go
const (
    MaxImageWidth  = 10000  // 10k pixels
    MaxImageHeight = 10000  // 10k pixels
    MaxImagePixels = 50000000  // 50 megapixels
)
```

Add validation in `OptimizeUpload()`:
```go
func OptimizeUpload(fSrc string, dstPath string, c config.Config) (*[]string, error) {
    // ... existing file open logic ...

    // Decode the image
    var dImg image.Image
    switch ext {
    case ".png":
        dImg, err = png.Decode(file)
    case ".jpg", ".jpeg":
        dImg, err = jpeg.Decode(file)
    case ".webp":
        dImg, err = webp.Decode(file)
    case ".gif":
        dImg, err = gif.Decode(file)
    default:
        return nil, fmt.Errorf("unsupported file extension: %s", ext)
    }
    if err != nil {
        return nil, fmt.Errorf("error decoding image: %w", err)
    }
    if dImg == nil {
        return nil, fmt.Errorf("decoded image is nil")
    }

    // Validate dimensions
    bounds := dImg.Bounds()
    width := bounds.Dx()
    height := bounds.Dy()
    pixels := width * height

    if width > MaxImageWidth {
        return nil, fmt.Errorf("image width %d exceeds maximum %d", width, MaxImageWidth)
    }
    if height > MaxImageHeight {
        return nil, fmt.Errorf("image height %d exceeds maximum %d", height, MaxImageHeight)
    }
    if pixels > MaxImagePixels {
        return nil, fmt.Errorf("image size %d pixels exceeds maximum %d", pixels, MaxImagePixels)
    }

    // Continue with existing logic...
}
```

**Testing:**
1. Upload normal image (1920x1080) - should succeed
2. Upload oversized image (15000x15000) - should reject
3. Verify error message returned to user

**Files Modified:** `internal/media/media_optomize.go`

---

### Issue #12: No File Size Validation

**Problem:** Only enforces multipart limit (10MB), but doesn't explicitly validate file size.

**Fix:**

Add file size check in handler:
```go
func apiCreateMediaUpload(w http.ResponseWriter, r *http.Request, c config.Config) {
    const MaxFileSize = 10 << 20 // 10 MB

    // Parse the multipart form
    err := r.ParseMultipartForm(MaxFileSize)
    if err != nil {
        utility.DefaultLogger.Error("parse form", err)
        http.Error(w, "File too large or invalid multipart form", http.StatusBadRequest)
        return
    }

    // Retrieve the file
    file, header, err := r.FormFile("file")
    if err != nil {
        utility.DefaultLogger.Error("parse file", err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Explicit size validation
    if header.Size > MaxFileSize {
        utility.DefaultLogger.Error("file too large", fmt.Errorf("size: %d", header.Size))
        http.Error(w, fmt.Sprintf("File size %d exceeds maximum %d", header.Size, MaxFileSize), http.StatusBadRequest)
        return
    }

    // Continue with existing logic...
}
```

**Testing:**
1. Upload 5MB file - should succeed
2. Upload 11MB file - should reject
3. Verify appropriate error message

**Files Modified:** `internal/router/mediaUpload.go`

---

### Issue #23: No MIME Type Validation

**Problem:** Relies on file extension, doesn't verify actual file content. Possible file type confusion attacks.

**Fix:**

Add MIME type validation:
```go
func apiCreateMediaUpload(w http.ResponseWriter, r *http.Request, c config.Config) {
    // ... existing file parsing ...

    // Read first 512 bytes to detect content type
    buffer := make([]byte, 512)
    _, err = file.Read(buffer)
    if err != nil && err != io.EOF {
        utility.DefaultLogger.Error("read file header", err)
        http.Error(w, "Failed to read file", http.StatusInternalServerError)
        return
    }

    // Reset file pointer
    _, err = file.Seek(0, 0)
    if err != nil {
        utility.DefaultLogger.Error("seek file", err)
        http.Error(w, "Failed to process file", http.StatusInternalServerError)
        return
    }

    // Validate MIME type
    contentType := http.DetectContentType(buffer)
    validTypes := map[string]bool{
        "image/png":  true,
        "image/jpeg": true,
        "image/gif":  true,
        "image/webp": true,
    }

    if !validTypes[contentType] {
        utility.DefaultLogger.Error("invalid content type", fmt.Errorf("type: %s", contentType))
        http.Error(w, fmt.Sprintf("Invalid file type: %s. Only images allowed.", contentType), http.StatusBadRequest)
        return
    }

    // Continue with existing logic...
}
```

**Testing:**
1. Upload valid PNG - should succeed
2. Upload .jpg with PNG content - should succeed (validates actual content)
3. Upload .jpg with PDF content - should reject
4. Upload text file renamed to .jpg - should reject

**Files Modified:** `internal/router/mediaUpload.go`

---

### Issue #13: Public-Read ACL for All Files

**Problem:** All uploads are public by default. May not be desired for private content.

**Fix (Simple - Make Configurable):**

Add to `config.Config`:
```go
type Config struct {
    // ... existing fields ...
    Bucket_Default_ACL string `json:"bucket_default_acl"` // "public-read", "private", etc.
}
```

Update `UploadPrep()`:
```go
func UploadPrep(uploadPath string, bucketName string, data *os.File, acl string) (*s3.PutObjectInput, error) {
    return &s3.PutObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(uploadPath),
        Body:   data,
        ACL:    aws.String(acl),
    }, nil
}
```

Update caller in `media_upload.go`:
```go
acl := c.Bucket_Default_ACL
if acl == "" {
    acl = "public-read"  // Default for backwards compatibility
}

prep, err := bucket.UploadPrep(s3Key, c.Bucket_Media, file, acl)
```

**Testing:**
1. Upload with default config - should be public-read
2. Upload with ACL set to "private" - should be private
3. Verify S3 object permissions match config

**Files Modified:**
- `internal/config/structs.go`
- `internal/bucket/object_storage.go`
- `internal/media/media_upload.go`

---

## Phase 4: Code Quality

**Priority:** P2 - Technical debt and maintainability
**Estimated Time:** 2-3 hours

### Issue #14: Typo in Filename

**Current:** `media_optomize.go`
**Should be:** `media_optimize.go`

**Fix:**
```bash
cd internal/media
git mv media_optomize.go media_optimize.go
```

Update any imports or references (should be none - internal package).

**Testing:** Run `make build` to verify no broken references

**Files Modified:** Rename `internal/media/media_optomize.go` → `internal/media/media_optimize.go`

---

### Issue #15: Inconsistent Variable Naming

**Fix throughout media package:**

| Current | Should Be | Reason |
|---------|-----------|--------|
| `dimensionsPTR` | `dimensions` | Go convention - pointer is implicit |
| `dImg` | `decodedImg` | Clarity over brevity |
| `fSrc` | `srcFile` | Consistent with parameter names |
| `dx` | `dim` | More descriptive |
| `im` | `img` | Standard abbreviation |

**Example refactor in `OptimizeUpload()`:**
```go
func OptimizeUpload(srcFile string, dstPath string, c config.Config) (*[]string, error) {
    d := db.ConfigDB(c)

    file, err := os.Open(srcFile)
    if err != nil {
        return nil, fmt.Errorf("couldn't find tmp file: %w", err)
    }
    defer file.Close()

    dimensions, err := d.ListMediaDimensions()
    if err != nil {
        return nil, fmt.Errorf("failed to list media dimensions: %w", err)
    }
    if dimensions == nil {
        return nil, fmt.Errorf("dimensions list is nil")
    }

    // ... decode logic ...

    var decodedImg image.Image
    switch ext {
    case ".png":
        decodedImg, err = png.Decode(file)
    // ...
    }

    // ... rest of function with consistent naming ...
}
```

**Files Modified:** `internal/media/media_optimize.go`

---

### Issue #16: Excessive Debug Logging

**File:** `internal/media/media_optimize.go:46-48`

**Current Code:**
```go
utility.DefaultLogger.Debug("last", last)
utility.DefaultLogger.Debug("trimmedprefix", trimmedPrefix)
utility.DefaultLogger.Debug("baseName", baseName)
```

**Fix:** Remove debug logging or make conditional:
```go
// Remove entirely - not useful in production
```

**Files Modified:** `internal/media/media_optimize.go`

---

### Issue #17: Incorrect Debug Log

**File:** `internal/media/media_upload.go:68`

**Current Code:**
```go
utility.DefaultLogger.Debug("SQL Filter Condition:", srcFile)
```

**Fix:** Remove or log correct value:
```go
utility.DefaultLogger.Debug(fmt.Sprintf("Fetching media record for: %s", baseName))
```

**Files Modified:** `internal/media/media_upload.go`

---

### Issue #18: Magic Numbers

**Fix:** Define constants at package level:

**New File:** `internal/media/constants.go`
```go
package media

const (
    // File size limits
    MaxUploadSize = 10 << 20 // 10 MB

    // Image dimension limits
    MaxImageWidth  = 10000    // 10k pixels
    MaxImageHeight = 10000    // 10k pixels
    MaxImagePixels = 50000000 // 50 megapixels

    // S3 configuration
    DefaultS3Region = "us-southeast-1"

    // Temp directory prefix
    TempDirPrefix = "modulacms-media"
)
```

Update usage throughout package to reference constants.

**Files Created:** `internal/media/constants.go`
**Files Modified:** `internal/media/*.go`, `internal/router/mediaUpload.go`

---

### Issue #21: media_assert.go Stub

**File:** `internal/media/media_assert.go`

**Current Code:**
```go
package media

func AssertMedia(pName string) {
}

// function to return list of files in path
// function to take list of files and compare to database entries
```

**Decision:** Either implement or delete.

**Recommendation:** Delete for now. If assertion functionality is needed later, implement properly with:
1. List S3 bucket contents
2. Compare to database records
3. Report orphaned files (in S3 but not DB)
4. Report missing files (in DB but not S3)

**Fix:**
```bash
rm internal/media/media_assert.go
```

**Files Deleted:** `internal/media/media_assert.go`

---

## Phase 5: Missing Features & Enhancements

**Priority:** P3 - Nice to have, not launch blockers
**Estimated Time:** 4-6 hours

### Issue #19: WebP Asymmetry

**Problem:** Can decode WebP but cannot encode. WebP uploads get converted to JPEG/PNG.

**Fix Option 1 (Simple):** Document the limitation

Add to handler:
```go
// Note: WebP uploads will be converted to JPEG format
if contentType == "image/webp" {
    utility.DefaultLogger.Info("WebP upload will be converted to JPEG")
}
```

**Fix Option 2 (Proper):** Add WebP encoding support

Requires external library (no WebP encoder in Go stdlib):
```go
import "github.com/chai2010/webp"
```

Add encoding case:
```go
switch ext {
case ".png":
    err = png.Encode(f, img)
case ".jpg", ".jpeg":
    err = jpeg.Encode(f, img, nil)
case ".gif":
    err = gif.Encode(f, img, nil)
case ".webp":
    err = webp.Encode(f, img, &webp.Options{Lossless: false, Quality: 90})
default:
    err = fmt.Errorf("unsupported encoding for extension: %s", ext)
}
```

**Recommendation:** Option 1 for now (document limitation), Option 2 post-launch if needed.

**Files Modified:** `internal/router/mediaUpload.go` (documentation comment)

---

### Issue #20: No Tests

**Create test file:** `internal/media/media_test.go`

**Test Cases:**

```go
package media

import (
    "os"
    "testing"
    "path/filepath"

    "github.com/hegner123/modulacms/internal/config"
)

func TestOptimizeUpload_ValidImage(t *testing.T) {
    // Test: Upload valid PNG, verify all dimensions generated
}

func TestOptimizeUpload_InvalidFormat(t *testing.T) {
    // Test: Upload non-image file, verify error returned
}

func TestOptimizeUpload_OversizedImage(t *testing.T) {
    // Test: Upload image exceeding size limits, verify rejection
}

func TestOptimizeUpload_FilesInCorrectDirectory(t *testing.T) {
    // Test: Verify optimized files created in dstPath, not CWD
}

func TestHandleMediaUpload_S3Failure_Rollback(t *testing.T) {
    // Test: Simulate S3 failure, verify rollback deletes uploaded files
}

func TestHandleMediaUpload_DBFailure_Rollback(t *testing.T) {
    // Test: Simulate DB update failure, verify S3 files deleted
}

func TestHandleMediaUpload_Success(t *testing.T) {
    // Test: Full upload flow succeeds, verify DB and S3 state
}
```

**Mock Requirements:**
- Mock S3 client (use interface)
- Mock database (use interface)
- Test image files (small PNG, JPEG)

**Estimated Effort:** 4-6 hours for comprehensive test coverage

**Files Created:** `internal/media/media_test.go`

---

## Implementation Order

### Day 1: Critical Fixes (Phase 1)
1. ✅ Fix temp directory cleanup (#3) - 30 min
2. ✅ Fix temp directory location (#2) - 15 min
3. ✅ Fix path construction bug (#1) - 30 min
4. ✅ Fix file creation directory (#5) - 30 min
5. ✅ Fix ObjectUpload error handling (#9) - 15 min
6. ✅ Remove db-sqlite import (#4) - 10 min
7. ✅ Add AuthorID TODO (#6) - 10 min
8. **Test all fixes** - 2 hours

**Total Day 1:** ~4-5 hours

---

### Day 2: High Priority & Security (Phases 2 & 3)
1. ✅ Implement transaction handling (#7) - 2 hours
2. ✅ Add rollback on optimization failure (#22) - 1 hour
3. ✅ Add image dimension validation (#11) - 45 min
4. ✅ Add file size validation (#12) - 30 min
5. ✅ Add MIME type validation (#23) - 45 min
6. ✅ Make ACL configurable (#13) - 30 min
7. **Test all fixes** - 2 hours

**Total Day 2:** ~7-8 hours

---

### Day 3: Code Quality & Testing (Phases 4 & 5)
1. ✅ Rename file (#14) - 5 min
2. ✅ Fix variable naming (#15) - 30 min
3. ✅ Clean up debug logs (#16, #17) - 15 min
4. ✅ Extract magic numbers (#18) - 30 min
5. ✅ Delete media_assert.go (#21) - 5 min
6. ✅ Document WebP limitation (#19) - 15 min
7. ✅ Write comprehensive tests (#20) - 4 hours
8. **Integration testing** - 2 hours

**Total Day 3:** ~7-8 hours

---

## Testing Strategy

### Unit Tests
- `TestOptimizeUpload_*` - Image optimization logic
- `TestHandleMediaUpload_*` - Upload orchestration
- `TestMapMediaParams` - Parameter mapping
- `TestRollbackS3Uploads` - Rollback logic

### Integration Tests
- End-to-end upload with real temp files
- Mock S3 client for upload testing
- Mock database for transaction testing
- Verify file cleanup in all scenarios

### Manual Testing
1. Upload valid images (PNG, JPEG, GIF)
2. Upload invalid files (should reject)
3. Upload oversized images (should reject)
4. Simulate S3 failures (verify rollback)
5. Simulate DB failures (verify rollback)
6. Upload 100 images sequentially (check for leaks)
7. Test on SQLite, MySQL, PostgreSQL

---

## Success Criteria

### Phase 1 Complete When:
- ✅ Temp directories cleaned up automatically
- ✅ Temp directories created in system temp location
- ✅ S3 paths are correct (no temp path fragments)
- ✅ Files created in correct destination directory
- ✅ S3 upload errors propagate correctly
- ✅ No db-sqlite import in bucket package
- ✅ All Phase 1 tests passing

### Phase 2 Complete When:
- ✅ Transaction handling prevents orphaned records
- ✅ S3 failures trigger rollback (delete uploaded files)
- ✅ DB failures trigger rollback (delete uploaded files)
- ✅ Optimization failures clean up partial files
- ✅ All Phase 2 tests passing

### Phase 3 Complete When:
- ✅ Oversized images rejected before processing
- ✅ File size validation enforced
- ✅ MIME type validated (not just extension)
- ✅ ACL configurable via config.json
- ✅ All Phase 3 tests passing

### Phase 4 Complete When:
- ✅ File renamed to media_optimize.go
- ✅ Variable names consistent throughout
- ✅ No excessive debug logging
- ✅ Constants defined for magic numbers
- ✅ media_assert.go deleted
- ✅ WebP limitation documented

### Phase 5 Complete When:
- ✅ Comprehensive test suite passing
- ✅ Integration tests validate full workflow
- ✅ Manual testing completed across all databases

---

## Files Modified Summary

**Created:**
- `internal/media/constants.go`
- `internal/media/media_test.go`

**Modified:**
- `internal/media/media_upload.go`
- `internal/media/media_optomize.go` → `internal/media/media_optimize.go` (renamed)
- `internal/media/media_create.go`
- `internal/bucket/object_storage.go`
- `internal/bucket/structs.go`
- `internal/router/mediaUpload.go`
- `internal/config/structs.go`

**Deleted:**
- `internal/media/media_assert.go`

**Total:** 2 created, 8 modified, 1 deleted

---

## Risk Assessment

### Low Risk Changes
- Temp directory cleanup fix (#2, #3)
- Remove unused import (#4)
- Rename file (#14)
- Clean up logging (#16, #17)
- Delete stub file (#21)

### Medium Risk Changes
- Path construction fix (#1, #5)
- Error handling fix (#9)
- Variable naming (#15)
- Constants extraction (#18)

### High Risk Changes
- Transaction handling (#7)
- Rollback logic (#22)
- Validation changes (#11, #12, #23)
- ACL configuration (#13)

**Mitigation:** Comprehensive testing at each phase, especially for high-risk changes.

---

## Rollback Plan

If critical issues discovered after deployment:

1. **Phase 1 Issues:** Revert to previous version, apply hotfix individually
2. **Phase 2 Issues:** Disable transaction handling, fall back to old flow (accept orphaned records risk)
3. **Phase 3 Issues:** Disable validation temporarily (security risk - only as last resort)
4. **Phase 4 Issues:** Revert cosmetic changes only (low impact)

**Critical Rollback:** Keep previous version tagged in git for quick revert.

---

## Post-Implementation

### Documentation Updates
- Update `ai/domain/MEDIA_SYSTEM.md` with:
  - New transaction handling flow
  - Validation limits (10MB, 10k pixels, 50MP)
  - Rollback behavior
  - ACL configuration options
  - WebP limitation notes

### Team Memory
- Store implementation decisions
- Document rollback patterns for future features
- Record validation limits rationale

### Monitoring
- Log upload failures with reason
- Track rollback frequency
- Monitor S3 storage usage
- Alert on orphaned file detection

---

## Related Issues

This plan addresses media package issues. Related work:
- **AuthorID Session Integration:** Affects all packages (media, CMS content creation, etc.)
- **Generic Error Handling:** Pattern established here can be applied elsewhere
- **Transaction Patterns:** Can be reused for other multi-step operations

---

## Questions for User

1. **ACL Configuration:** Should we make ACL per-upload or global config? (Recommendation: global with per-upload override capability)
2. **Image Size Limits:** Are 10MB file / 10k pixels / 50MP reasonable? (Adjust based on use case)
3. **WebP Encoding:** Do we need WebP output support for launch? (Recommendation: no, document limitation)
4. **Test Coverage:** Should we write integration tests with real S3 or mock? (Recommendation: mock for CI/CD)

---

**Last Updated:** 2026-01-18
**Status:** Ready for Implementation
**Next Step:** Begin Phase 1 - Critical Fixes
