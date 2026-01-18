# Import API - CMS Content Migration

**Created:** 2026-01-17
**Status:** 🔄 STUB - Database insertion not implemented
**Location:** `internal/router/import.go`

---

## Overview

The Import API provides endpoints to accept content in various CMS formats (Contentful, Sanity, Strapi, WordPress) and import it into ModulaCMS's database structure.

This enables **migration from other CMSs** by accepting their JSON export format and automatically creating:
- Datatypes
- Fields
- Content Data
- Content Fields

---

## Endpoints

### 1. Format-Specific Import

Each CMS format has its own endpoint:

```bash
POST /api/v1/import/contentful
POST /api/v1/import/sanity
POST /api/v1/import/strapi
POST /api/v1/import/wordpress
POST /api/v1/import/clean
```

**Request:**
```bash
curl -X POST http://localhost:3000/api/v1/import/contentful \
  -H "Content-Type: application/json" \
  -d @contentful-export.json
```

**Response:**
```json
{
  "success": true,
  "datatypes_created": 3,
  "fields_created": 15,
  "content_created": 10,
  "message": "Import stubbed - database insertion not implemented"
}
```

---

### 2. Bulk Import with Format Parameter

Use a single endpoint with format specified in query parameter:

```bash
POST /api/v1/import?format={format}
```

**Example:**
```bash
curl -X POST "http://localhost:3000/api/v1/import?format=contentful" \
  -H "Content-Type: application/json" \
  -d @export.json
```

**Valid formats:**
- `contentful`
- `sanity`
- `strapi`
- `wordpress`
- `clean`

---

## Request Format

### Contentful Example

```json
{
  "sys": {
    "id": "42",
    "type": "Entry",
    "contentType": {
      "sys": { "id": "blogPost" }
    },
    "createdAt": "2026-01-15T10:30:00Z",
    "updatedAt": "2026-01-16T14:22:00Z"
  },
  "fields": {
    "title": "Hello World",
    "slug": "hello-world",
    "published": true,
    "content": "This is the post content..."
  }
}
```

---

### Clean ModulaCMS Example

```json
{
  "id": 42,
  "type": "Blog Post",
  "title": "Hello World",
  "slug": "hello-world",
  "published": true,
  "_meta": {
    "authorId": 1,
    "routeId": 1,
    "dateCreated": "2026-01-15T10:30:00Z",
    "dateModified": "2026-01-16T14:22:00Z"
  }
}
```

---

## Response Format

### Success Response

```json
{
  "success": true,
  "datatypes_created": 3,
  "fields_created": 15,
  "content_created": 10,
  "message": "Content imported successfully"
}
```

### Error Response

```json
{
  "success": false,
  "datatypes_created": 0,
  "fields_created": 0,
  "content_created": 0,
  "errors": [
    "Failed to parse datatype 'blogPost'",
    "Missing required field 'title'"
  ],
  "message": "Import failed"
}
```

---

## How It Works

### 1. Parse Input

The handler reads the request body and uses the appropriate transformer to parse it:

```go
// Create transformer for the format
transformer, _ := transformCfg.GetTransformer()

// Parse CMS format to ModulaCMS
root, err := transformer.Parse(body)
```

### 2. Extract Database Structures

From the parsed `model.Root`, extract:

- **Datatypes**: From `root.Node.Datatype.Info`
- **Fields**: From `root.Node.Fields`
- **Content Data**: From `root.Node.Datatype.Content`
- **Content Fields**: From `root.Node.Fields` linked to content
- **Child Nodes**: Recursively from `root.Node.Nodes`

### 3. Insert to Database (STUB)

**Current Status:** Stubbed - returns counts but doesn't insert

**To Implement:**
```go
// 1. Create or get Datatype
datatypeID := createOrGetDatatype(root.Node.Datatype.Info)

// 2. Create Fields for this datatype
for _, field := range root.Node.Fields {
    fieldID := createField(field.Info, datatypeID)
}

// 3. Create ContentData
contentID := createContentData(root.Node.Datatype.Content, datatypeID)

// 4. Create ContentFields linking content to fields
for _, field := range root.Node.Fields {
    createContentField(contentID, field)
}

// 5. Recursively process child nodes
for _, child := range root.Node.Nodes {
    importRootToDatabase(driver, model.Root{Node: child})
}
```

---

## Current Implementation Status

### ✅ Complete

- [x] Endpoint routing
- [x] Format-specific handlers
- [x] Bulk import with format parameter
- [x] Input parsing via transform package
- [x] Error handling
- [x] Response structure
- [x] Stub counting logic

### 🔄 To Implement

- [ ] Database insertion for Datatypes
- [ ] Database insertion for Fields
- [ ] Database insertion for ContentData
- [ ] Database insertion for ContentFields
- [ ] Recursive child node processing
- [ ] Relationship handling (parent/child, siblings)
- [ ] Duplicate detection (update vs insert)
- [ ] Transaction support (all-or-nothing import)
- [ ] Batch import optimization
- [ ] Import validation
- [ ] Import rollback on error

---

## Implementation Plan

### Phase 1: Single Entry Import

Implement database insertion for a single root entry:

```go
func importRootToDatabase(driver db.DbDriver, root model.Root) (*ImportResult, error) {
    // Start transaction
    tx, _ := driver.BeginTx()

    // 1. Create/Get Datatype
    datatypeID, err := getOrCreateDatatype(tx, root.Node.Datatype.Info)

    // 2. Create Fields
    fieldMap := make(map[int64]int64) // old ID -> new ID
    for _, field := range root.Node.Fields {
        newFieldID, err := createField(tx, field.Info, datatypeID)
        fieldMap[field.Info.FieldID] = newFieldID
    }

    // 3. Create ContentData
    contentID, err := createContentData(tx, root.Node.Datatype.Content, datatypeID)

    // 4. Create ContentFields
    for _, field := range root.Node.Fields {
        newFieldID := fieldMap[field.Info.FieldID]
        err := createContentField(tx, contentID, newFieldID, field.Content.FieldValue)
    }

    // Commit transaction
    tx.Commit()

    return &ImportResult{Success: true, ContentCreated: 1}, nil
}
```

### Phase 2: Recursive Tree Import

Handle child nodes and tree relationships:

```go
// After creating parent content
for _, child := range root.Node.Nodes {
    childResult, err := importRootToDatabase(driver, model.Root{Node: child})
    // Link child to parent via parent_id, first_child_id, etc.
}
```

### Phase 3: Duplicate Handling

Check if content already exists and update instead of insert:

```go
// Check if content exists by external ID
existingContent, err := driver.GetContentDataByExternalID(externalID)
if err == nil {
    // Update existing
    driver.UpdateContentData(existingContent.ContentDataID, updateParams)
} else {
    // Create new
    driver.CreateContentData(createParams)
}
```

### Phase 4: Batch Optimization

Accept arrays of entries for bulk import:

```json
POST /api/v1/import/contentful
{
  "items": [
    { "sys": {...}, "fields": {...} },
    { "sys": {...}, "fields": {...} },
    { "sys": {...}, "fields": {...} }
  ]
}
```

---

## Usage Examples

### Import Single Contentful Entry

```bash
curl -X POST http://localhost:3000/api/v1/import/contentful \
  -H "Content-Type: application/json" \
  -d '{
    "sys": {
      "id": "42",
      "type": "Entry",
      "contentType": { "sys": { "id": "blogPost" } }
    },
    "fields": {
      "title": "Hello World",
      "published": true
    }
  }'
```

**Response:**
```json
{
  "success": true,
  "datatypes_created": 1,
  "fields_created": 2,
  "content_created": 1,
  "message": "Import stubbed - database insertion not implemented"
}
```

---

### Import Clean Format

```bash
curl -X POST http://localhost:3000/api/v1/import/clean \
  -H "Content-Type: application/json" \
  -d '{
    "id": 42,
    "type": "Blog Post",
    "title": "Hello World",
    "published": true
  }'
```

---

### Bulk Import with Format Parameter

```bash
curl -X POST "http://localhost:3000/api/v1/import?format=sanity" \
  -H "Content-Type: application/json" \
  -d @sanity-export.json
```

---

## Error Handling

### Invalid Format

```bash
curl -X POST "http://localhost:3000/api/v1/import?format=invalid"
```

**Response:** `400 Bad Request`
```
Invalid format. Valid options: contentful, sanity, strapi, wordpress, clean
```

---

### Invalid JSON

```bash
curl -X POST http://localhost:3000/api/v1/import/contentful \
  -d 'not valid json'
```

**Response:** `400 Bad Request`
```
Failed to parse input: invalid character 'o' in literal null
```

---

### Unsupported Parser

```bash
curl -X POST http://localhost:3000/api/v1/import/sanity \
  -d '{"_id": "42"}'
```

**Response:** `400 Bad Request`
```
Failed to parse input: Sanity inbound parsing not yet implemented - use format='clean' or 'raw' for POST/PUT requests
```

---

## Integration with Transform Package

The import API uses the transform package's `Parse()` method:

```go
// From internal/router/import.go
transformer, _ := transformCfg.GetTransformer()
root, err := transformer.Parse(body)
```

**Fully Supported (Bidirectional):**
- ✅ Contentful
- ✅ Clean
- ✅ Raw

**Stub Only (Parse Not Implemented):**
- ⚠️ Sanity
- ⚠️ Strapi
- ⚠️ WordPress

---

## Migration Workflow Example

### Step 1: Export from Contentful

```bash
# Use Contentful CLI to export
contentful space export --space-id=abc123 > contentful-export.json
```

### Step 2: Import to ModulaCMS

```bash
# Import to ModulaCMS
curl -X POST http://localhost:3000/api/v1/import/contentful \
  -H "Content-Type: application/json" \
  -d @contentful-export.json
```

### Step 3: Verify Import

```bash
# Check imported content
curl http://localhost:3000/api/v1/contentdata
```

### Step 4: Test Output

```bash
# View in Contentful format (same as original)
curl http://localhost:3000/blog/my-post?format=contentful

# Or view in clean format
curl http://localhost:3000/blog/my-post?format=clean
```

---

## Testing

### Manual Test

```bash
# Start ModulaCMS
./modulacms-x86 --cli

# In another terminal, test import
curl -X POST http://localhost:3000/api/v1/import/clean \
  -H "Content-Type: application/json" \
  -d '{
    "type": "Test",
    "title": "Test Import"
  }'
```

### Expected Response (Stub)

```json
{
  "success": true,
  "datatypes_created": 1,
  "fields_created": 1,
  "content_created": 1,
  "message": "Import stubbed - database insertion not implemented"
}
```

---

## Future Enhancements

### 1. Import Progress Tracking

Long-running imports should provide progress:

```json
{
  "import_id": "uuid-123",
  "status": "in_progress",
  "progress": 45,
  "total": 100,
  "message": "Importing content... 45/100"
}
```

### 2. Import History

Track all imports with ability to view/rollback:

```bash
GET /api/v1/imports
GET /api/v1/imports/{import_id}
DELETE /api/v1/imports/{import_id}  # Rollback
```

### 3. Import Mapping Configuration

Allow custom field mappings:

```json
{
  "format": "contentful",
  "mapping": {
    "sys.id": "content_data_id",
    "fields.title": "title",
    "fields.slug": "slug"
  },
  "data": { ... }
}
```

### 4. Import Validation

Validate before importing:

```bash
POST /api/v1/import/validate?format=contentful
```

Returns validation errors without creating database entries.

---

## Summary

**Current Status:** Import API endpoints are implemented and routes are registered. Parsing works via transform package. Database insertion is stubbed and needs implementation.

**Next Steps:**
1. Implement `importRootToDatabase()` function
2. Add transaction support
3. Handle duplicate detection
4. Add recursive tree processing
5. Add batch import support
6. Add import validation

**Files:**
- `internal/router/import.go` - Import handlers (stubbed)
- `internal/router/mux.go` - Route registration (complete)
- `internal/transform/*.go` - Parsing logic (complete for Contentful/Clean)

---

**Last Updated:** 2026-01-17
