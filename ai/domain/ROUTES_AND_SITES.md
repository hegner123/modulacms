# ROUTES_AND_SITES.md

Comprehensive documentation on ModulaCMS's multi-site architecture using routes as content tree roots.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/ROUTES_AND_SITES.md`
**Related Schemas:**
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/6_routes/`
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/5_admin_routes/`
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/16_content_data/`

---

## Overview

Routes are **top-level containers for content trees** in ModulaCMS. Each route represents an independent content hierarchy, enabling true multi-site architecture where multiple websites or content sections can coexist in a single CMS installation.

**Core Concept:**
```
Route = Content Tree Root + Domain Identifier
```

**Why This Matters:**
- Enables multi-site CMS (one installation, many sites)
- Provides content isolation between sites
- Allows different clients to share infrastructure
- Separates client-facing content from admin content
- Essential for agency use cases

---

## What Are Routes?

### Definition

A **route** is a named content tree root that:
1. Contains an independent content hierarchy
2. Has a unique slug (typically a domain name)
3. Isolates content from other routes
4. Serves as the entry point for content queries

**Analogy:**
- Like separate WordPress installations on the same server
- Like separate databases, but within one database
- Like separate folders on a filesystem (but for content)

### Routes vs Admin Routes

ModulaCMS has **two parallel route systems:**

**Client Routes (routes table):**
- Public-facing content for client websites
- Main content management focus
- Each route represents a client site or section
- Table: `routes`

**Admin Routes (admin_routes table):**
- Content for the admin interface itself
- Dashboard widgets, admin panels, settings pages
- Separate from client content
- Table: `admin_routes`

**Why Separate?**
- Security: Admin content never mixes with public content
- Flexibility: Admin interface can evolve independently
- Isolation: Client operations can't break admin interface

---

## Database Schema

### Routes Table

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/6_routes/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS routes (
    route_id INTEGER PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,           -- Domain or identifier
    title TEXT NOT NULL,                 -- Human-readable name
    status INTEGER NOT NULL,             -- Active/inactive
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT                         -- Audit trail
);
```

**Key Fields:**
- `route_id` - Unique identifier for this route
- `slug` - Unique slug, typically a domain name (e.g., "example.com", "staging.example.com")
- `title` - Human-readable name for the route (e.g., "Main Website", "Client ABC")
- `status` - Whether this route is active (1) or inactive (0)
- `author_id` - User who created this route
- `history` - JSON audit trail of changes

**Unique Constraint:** The `slug` must be unique across all routes.

### Admin Routes Table

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/5_admin_routes/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS admin_routes (
    admin_route_id INTEGER PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    status INTEGER NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
```

**Structure:** Identical to routes table, but separate for admin content.

### Go Structure

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/route.go:16-25`

```go
type Routes struct {
    RouteID      int64          `json:"route_id"`
    Slug         string         `json:"slug"`
    Title        string         `json:"title"`
    Status       int64          `json:"status"`
    AuthorID     int64          `json:"author_id"`
    DateCreated  sql.NullString `json:"date_created"`
    DateModified sql.NullString `json:"date_modified"`
    History      sql.NullString `json:"history"`
}
```

---

## Multi-Site Architecture

### How Routes Enable Multi-Site

Each route functions as an **independent website** within ModulaCMS:

```
ModulaCMS Installation
├─> Route: "example.com" (route_id=1)
│    └─> Content Tree: Home, About, Blog, Contact
│
├─> Route: "staging.example.com" (route_id=2)
│    └─> Content Tree: Test Home, Test About
│
├─> Route: "clienta.com" (route_id=3)
│    └─> Content Tree: Client A's content
│
└─> Route: "clientb.com" (route_id=4)
     └─> Content Tree: Client B's content
```

**Key Principle:** Content in one route **never** interacts with content in another route.

### Use Cases

**1. Multiple Client Sites**
- Agency hosts multiple client sites on one CMS installation
- Each client has their own route
- Clients can't see or access each other's content
- Shared infrastructure, isolated data

**2. Staging vs Production**
- Production site: "example.com" (route_id=1)
- Staging site: "staging.example.com" (route_id=2)
- Test changes on staging without affecting production
- Copy content between routes when ready

**3. Multi-Language Sites**
- English site: "example.com" (route_id=1)
- Spanish site: "es.example.com" (route_id=2)
- Separate content trees for each language
- Or: Use single route with language field (implementation choice)

**4. Internal vs External Content**
- Public website: "company.com" (route_id=1)
- Internal docs: "internal.company.com" (route_id=2)
- Different access controls per route
- Complete content separation

**5. Multi-Brand Architecture**
- Brand A site: "branda.com" (route_id=1)
- Brand B site: "brandb.com" (route_id=2)
- Same CMS, different brands
- Separate content, shared infrastructure

---

## Content Isolation by Route

### How Isolation Works

Every `content_data` row **must** have a `route_id`:

**Schema:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/16_content_data/schema.sql:16-18`

```sql
route_id INTEGER NOT NULL
    REFERENCES routes
    ON DELETE CASCADE,
```

**Critical:** `ON DELETE CASCADE` means deleting a route deletes all its content.

### Query Pattern: Always Filter by Route

**Rule:** Every content query MUST include `WHERE route_id = ?`

**Example: Get Content Tree**

```sql
-- name: GetContentTreeByRoute :many
SELECT cd.content_data_id,
       cd.parent_id,
       cd.first_child_id,
       cd.next_sibling_id,
       cd.prev_sibling_id,
       cd.datatype_id,
       cd.route_id,
       dt.label as datatype_label,
       dt.type as datatype_type
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id = ?
ORDER BY cd.parent_id NULLS FIRST, cd.content_data_id;
```

**Why Always Filter?**
- Without `route_id` filter, query returns content from ALL routes
- Results in mixed content from different sites
- Breaks content isolation
- Confuses tree structure (orphaned nodes from wrong route)

### Content Fields Also Denormalize route_id

**Schema:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/17_content_fields/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id INTEGER PRIMARY KEY,
    route_id INTEGER REFERENCES routes ON DELETE SET NULL,  -- Denormalized
    content_data_id INTEGER NOT NULL REFERENCES content_data ON DELETE CASCADE,
    field_id INTEGER NOT NULL REFERENCES fields ON DELETE CASCADE,
    field_value TEXT NOT NULL,
    ...
);
```

**Why Denormalize route_id?**
- Performance: Filter content_fields by route without joining to content_data
- Query optimization: Direct index on route_id
- Common pattern: "Get all field values for a route"

**Example: Get All Field Values for Route**

```sql
-- name: GetContentFieldsByRoute :many
SELECT cf.content_data_id, cf.field_id, cf.field_value
FROM content_data cd
JOIN content_fields cf ON cd.content_data_id = cf.content_data_id
WHERE cd.route_id = ?
ORDER BY cf.content_data_id, cf.field_id;
```

---

## Route Operations

### Creating a Route

**Step 1: Create Route Record**

```go
// Create route
route := db.CreateRoute(ctx, db.CreateRouteParams{
    Slug:   "example.com",
    Title:  "Example Website",
    Status: 1,  // Active
    AuthorID: userID,
})

fmt.Printf("Created route_id: %d\n", route.RouteID)
```

**Generated SQL:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/6_routes/queries.sql:42-60`

```sql
-- name: CreateRoute :one
INSERT INTO routes (
    slug,
    title,
    status,
    author_id,
    history,
    date_created,
    date_modified
) VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;
```

**Step 2: Create Root Content Node**

Every route needs a root node in the content tree:

```go
// Create root content node for this route
rootNode := db.CreateContentData(ctx, db.CreateContentDataParams{
    RouteID:    route.RouteID,
    ParentID:   sql.NullInt64{Valid: false},  // NULL parent = root
    DatatypeID: rootDatatypeID,
    AuthorID:   userID,
})
```

**Result:** Route now has an empty content tree ready for content.

### Listing Routes

**Query:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/6_routes/queries.sql:37-40`

```sql
-- name: ListRoute :many
SELECT *
FROM routes
ORDER BY slug;
```

**Go Usage:**

```go
routes, err := db.ListRoutes(ctx)
if err != nil {
    return fmt.Errorf("failed to list routes: %w", err)
}

for _, route := range *routes {
    fmt.Printf("Route: %s (%s)\n", route.Title, route.Slug)
}
```

**Output:**
```
Route: Client A (clienta.com)
Route: Client B (clientb.com)
Route: Main Website (example.com)
Route: Staging Site (staging.example.com)
```

### Getting Route by Slug

**Query:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/6_routes/queries.sql:31-35`

```sql
-- name: GetRouteIDBySlug :one
SELECT route_id
FROM routes
WHERE slug = ?
LIMIT 1;
```

**Go Usage:**

```go
routeID, err := db.GetRouteID(ctx, "example.com")
if err != nil {
    return fmt.Errorf("route not found: %w", err)
}

// Now load content tree for this route
rows, err := db.GetContentTreeByRoute(ctx, *routeID)
```

**Common Pattern:** Convert domain name → route_id → load content

### Updating a Route

**Query:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/6_routes/queries.sql:62-72`

```sql
-- name: UpdateRoute :exec
UPDATE routes
SET slug = ?,
    title = ?,
    status = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
WHERE slug = ?
RETURNING *;
```

**Go Usage:**

```go
result, err := db.UpdateRoute(ctx, db.UpdateRouteParams{
    Slug:   "newdomain.com",
    Title:  "Updated Title",
    Status: 1,
    // ... other fields
    Slug_2: "example.com",  // WHERE slug = this value
})
```

**Important:** `Slug_2` is the current slug (WHERE clause), `Slug` is the new slug.

### Deleting a Route

**Query:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/6_routes/queries.sql:74-76`

```sql
-- name: DeleteRoute :exec
DELETE FROM routes
WHERE route_id = ?;
```

**Critical:** Due to `ON DELETE CASCADE`, deleting a route also deletes:
- All content_data rows for this route
- All content_fields rows for this route
- Entire content tree for this route

**Use with caution!**

```go
err := db.DeleteRoute(ctx, routeID)
if err != nil {
    return fmt.Errorf("failed to delete route: %w", err)
}

// All content for this route is now deleted
```

---

## Working with Routes in Practice

### Pattern 1: Loading Content Tree for a Route

**Step-by-step:**

```go
// 1. Get route by domain
routeID, err := db.GetRouteID(ctx, "example.com")
if err != nil {
    return fmt.Errorf("route not found: %w", err)
}

// 2. Load entire content tree for route
rows, err := db.GetContentTreeByRoute(ctx, *routeID)
if err != nil {
    return fmt.Errorf("failed to load tree: %w", err)
}

// 3. Build tree structure using three-phase algorithm (see TREE_STRUCTURE.md)
treeRoot := model.NewTreeRoot()
stats, err := treeRoot.LoadFromRows(rows)
if err != nil {
    return fmt.Errorf("failed to build tree: %w", err)
}

// 4. Tree is ready for navigation
rootNode := treeRoot.Root
```

**Query Used:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/22_joins/queries.sql:13-28`

```sql
-- name: GetContentTreeByRoute :many
SELECT cd.content_data_id,
       cd.parent_id,
       cd.first_child_id,
       cd.next_sibling_id,
       cd.prev_sibling_id,
       cd.datatype_id,
       cd.route_id,
       cd.author_id,
       cd.date_created,
       cd.date_modified,
       dt.label as datatype_label,
       dt.type as datatype_type
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id = ?
ORDER BY cd.parent_id NULLS FIRST, cd.content_data_id;
```

### Pattern 2: Creating Content in a Route

**Step-by-step:**

```go
// 1. Get route ID
routeID, err := db.GetRouteID(ctx, "example.com")
if err != nil {
    return fmt.Errorf("route not found: %w", err)
}

// 2. Create content node
newPage := db.CreateContentData(ctx, db.CreateContentDataParams{
    RouteID:    *routeID,
    ParentID:   sql.NullInt64{Int64: parentNodeID, Valid: true},
    DatatypeID: pageDatatypeID,
    AuthorID:   userID,
})

// 3. Add field values
db.CreateContentField(ctx, db.CreateContentFieldParams{
    ContentDataID: newPage.ContentDataID,
    RouteID:       sql.NullInt64{Int64: *routeID, Valid: true},  // Denormalized
    FieldID:       titleFieldID,
    FieldValue:    "New Page Title",
    AuthorID:      userID,
})
```

**Important:** Always set `route_id` on both content_data and content_fields.

### Pattern 3: Shallow Tree Loading (Lazy Loading)

**Query:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/22_joins/queries.sql:2-11`

```sql
-- name: GetShallowTreeByRouteId :many
SELECT cd.*, dt.label as datatype_label, dt.type as datatype_type
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id = ?
AND (cd.parent_id IS NULL OR cd.parent_id IN (
    SELECT content_data_id FROM content_data
    WHERE cd.parent_id IS NULL AND cd.route_id = ?
))
ORDER BY cd.parent_id NULLS FIRST, cd.content_data_id;
```

**Purpose:** Load only root node and first-level children for performance.

**Go Usage:**

```go
// Load shallow tree (root + immediate children only)
rows, err := db.GetShallowTreeByRouteId(ctx, routeID, routeID)

// Build partial tree
partialTree := model.NewTreeRoot()
partialTree.LoadFromRows(rows)

// Load children on-demand as user expands nodes
```

**When to Use:**
- Large content trees (1,000+ nodes)
- TUI initial load
- Mobile clients with memory constraints
- API responses where full tree isn't needed

### Pattern 4: Counting Content by Route

```go
// Count total content items in a route
countQuery := `
    SELECT COUNT(*)
    FROM content_data
    WHERE route_id = ?
`
var count int64
err := db.QueryRow(ctx, countQuery, routeID).Scan(&count)

fmt.Printf("Route has %d content items\n", count)
```

### Pattern 5: Searching Content Within a Route

```sql
-- Search for content with specific field value in a route
SELECT DISTINCT cd.content_data_id, cd.datatype_id
FROM content_data cd
JOIN content_fields cf ON cd.content_data_id = cf.content_data_id
JOIN fields f ON cf.field_id = f.field_id
WHERE cd.route_id = ?
  AND f.label = 'Title'
  AND cf.field_value LIKE '%search term%'
ORDER BY cd.date_created DESC;
```

**Critical:** Always include `cd.route_id = ?` to limit search to one route.

---

## Route Status

### Status Values

Routes use an integer status field:

**Common Status Values:**
- `0` - Inactive/Disabled
- `1` - Active/Published
- `2` - Archived
- `3` - Maintenance Mode

**Implementation varies by project needs.**

### Checking Route Status

```go
route, err := db.GetRoute(ctx, routeID)
if err != nil {
    return fmt.Errorf("route not found: %w", err)
}

if route.Status != 1 {
    return fmt.Errorf("route is not active")
}

// Proceed with loading content
```

### Filtering by Status

```sql
-- List only active routes
SELECT *
FROM routes
WHERE status = 1
ORDER BY slug;
```

---

## Route History and Audit Trail

### History Field

The `history` field stores a JSON audit trail:

```json
{
  "changes": [
    {
      "timestamp": "2026-01-12T10:30:00Z",
      "user_id": 1,
      "action": "created",
      "details": {
        "slug": "example.com",
        "title": "Example Website"
      }
    },
    {
      "timestamp": "2026-01-13T14:20:00Z",
      "user_id": 1,
      "action": "updated",
      "details": {
        "old_title": "Example Website",
        "new_title": "Updated Website"
      }
    }
  ]
}
```

### Recording Changes

```go
type RouteHistory struct {
    Changes []RouteHistoryEntry `json:"changes"`
}

type RouteHistoryEntry struct {
    Timestamp string                 `json:"timestamp"`
    UserID    int64                  `json:"user_id"`
    Action    string                 `json:"action"`
    Details   map[string]interface{} `json:"details"`
}

// Add history entry
history := RouteHistory{}
if route.History.Valid {
    json.Unmarshal([]byte(route.History.String), &history)
}

history.Changes = append(history.Changes, RouteHistoryEntry{
    Timestamp: time.Now().Format(time.RFC3339),
    UserID:    userID,
    Action:    "updated",
    Details: map[string]interface{}{
        "field": "status",
        "old":   0,
        "new":   1,
    },
})

historyJSON, _ := json.Marshal(history)
route.History = sql.NullString{String: string(historyJSON), Valid: true}
```

---

## Common Pitfalls and Debugging

### Pitfall 1: Forgetting to Filter by route_id

**Problem:** Query returns content from all routes mixed together.

```go
// BAD: Missing route_id filter
rows, err := db.Query(`
    SELECT * FROM content_data
    ORDER BY content_data_id
`)

// Results in mixed content from all routes
```

**Fix:** Always include route_id in WHERE clause.

```go
// GOOD: Filter by route_id
rows, err := db.Query(`
    SELECT * FROM content_data
    WHERE route_id = ?
    ORDER BY content_data_id
`, routeID)
```

### Pitfall 2: Using Wrong route_id

**Problem:** Loading content for route A, but accidentally using route_id from route B.

**Symptom:** Empty tree or wrong content appears.

**Debug:**

```go
// Verify route_id matches expected slug
route, err := db.GetRoute(ctx, routeID)
if err != nil {
    log.Printf("Route not found: %v", err)
}

log.Printf("Loading content for route: %s (%d)", route.Slug, route.RouteID)
```

### Pitfall 3: Deleting Route Without Backup

**Problem:** `DELETE FROM routes WHERE route_id = ?` deletes ALL content due to CASCADE.

**Solution:** Always backup before deletion.

```go
// 1. Backup route content first
backup, err := CreateRouteBackup(ctx, routeID)
if err != nil {
    return fmt.Errorf("backup failed: %w", err)
}

// 2. Confirm deletion
fmt.Printf("WARNING: This will delete all content for route %d. Continue? (yes/no): ", routeID)
var response string
fmt.Scanln(&response)

if response != "yes" {
    return fmt.Errorf("deletion cancelled")
}

// 3. Now safe to delete
err = db.DeleteRoute(ctx, routeID)
```

### Pitfall 4: Duplicate Slugs

**Problem:** Attempting to create route with existing slug.

**Error:** `UNIQUE constraint failed: routes.slug`

**Prevention:**

```go
// Check if slug exists before creating
existingID, err := db.GetRouteID(ctx, slug)
if err == nil {
    return fmt.Errorf("route with slug '%s' already exists (route_id: %d)", slug, *existingID)
}

// Slug is available, create route
route := db.CreateRoute(ctx, params)
```

### Pitfall 5: NULL route_id in content_data

**Problem:** Some databases allow NULL even with NOT NULL constraint during migrations.

**Symptom:** Content appears orphaned, not associated with any route.

**Fix:**

```sql
-- Find orphaned content
SELECT * FROM content_data
WHERE route_id IS NULL;

-- Assign to route or delete
UPDATE content_data
SET route_id = ?
WHERE route_id IS NULL;
```

### Debugging Tips

**1. Verify Route Exists:**

```go
route, err := db.GetRoute(ctx, routeID)
if err != nil {
    log.Printf("Route %d not found", routeID)
    return
}
log.Printf("Route: %s (%s)", route.Title, route.Slug)
```

**2. Count Content Per Route:**

```sql
SELECT r.route_id, r.slug, COUNT(cd.content_data_id) as content_count
FROM routes r
LEFT JOIN content_data cd ON r.route_id = cd.route_id
GROUP BY r.route_id, r.slug
ORDER BY content_count DESC;
```

**3. Check for Content Isolation:**

```sql
-- Verify no content_data rows exist without valid route_id
SELECT cd.*
FROM content_data cd
LEFT JOIN routes r ON cd.route_id = r.route_id
WHERE r.route_id IS NULL;
```

**4. Audit Route Changes:**

```go
// Parse and display history
if route.History.Valid {
    var history RouteHistory
    json.Unmarshal([]byte(route.History.String), &history)

    for _, change := range history.Changes {
        log.Printf("%s: %s by user %d",
            change.Timestamp,
            change.Action,
            change.UserID)
    }
}
```

---

## Performance Considerations

### Index Recommendations

**Essential Indexes:**

```sql
-- Route slug lookup (already UNIQUE, creates index)
CREATE UNIQUE INDEX IF NOT EXISTS idx_routes_slug ON routes(slug);

-- Content by route (critical for isolation)
CREATE INDEX IF NOT EXISTS idx_content_data_route ON content_data(route_id);

-- Content fields by route (denormalized for performance)
CREATE INDEX IF NOT EXISTS idx_content_fields_route ON content_fields(route_id);
```

### Query Optimization

**Prefer route_id over slug:**

```go
// SLOW: Join to routes table on every query
rows, err := db.Query(`
    SELECT cd.*
    FROM content_data cd
    JOIN routes r ON cd.route_id = r.route_id
    WHERE r.slug = ?
`, "example.com")

// FAST: Look up route_id once, use it directly
routeID, err := db.GetRouteID(ctx, "example.com")
rows, err := db.Query(`
    SELECT cd.*
    FROM content_data cd
    WHERE cd.route_id = ?
`, *routeID)
```

**Cache route_id:** If loading content repeatedly for same domain, cache the route_id in memory.

### Lazy Loading by Route

**Problem:** Loading entire content tree for large route is slow.

**Solution:** Load incrementally using shallow queries.

```go
// 1. Load root + first level only
shallowRows := db.GetShallowTreeByRouteId(ctx, routeID, routeID)

// 2. Load children on-demand as user navigates
childRows := db.Query(`
    SELECT * FROM content_data
    WHERE route_id = ? AND parent_id = ?
`, routeID, parentNodeID)
```

---

## Comparison to Other Systems

### vs WordPress Multi-Site

**WordPress Multi-Site:**
- Separate tables per site (wp_1_posts, wp_2_posts)
- Complex database structure
- Harder to query across sites
- Network admin for super users

**ModulaCMS Routes:**
- Single tables with route_id column
- Simple query pattern: `WHERE route_id = ?`
- Easy to query across routes if needed
- Flexible per-route permissions

### vs Drupal Multi-Site

**Drupal Multi-Site:**
- Separate databases per site (via settings.php)
- Each site is isolated at database level
- Shared codebase, separate data

**ModulaCMS Routes:**
- Single database, isolated via route_id
- Shared schema, shared infrastructure
- Content isolation via foreign keys
- Easier backup and migration

### vs Contentful Spaces

**Contentful Spaces:**
- Separate content spaces within organization
- Each space has own content models
- API key per space

**ModulaCMS Routes:**
- Similar concept to spaces
- Each route has own content tree
- Can share datatypes/fields across routes
- Or use different schemas per route

---

## Advanced Use Cases

### Case 1: Copying Content Between Routes

**Scenario:** Copy staging content to production route.

```go
func CopyContentBetweenRoutes(ctx context.Context, db DbDriver, fromRouteID, toRouteID int64) error {
    // 1. Load source tree
    sourceRows, err := db.GetContentTreeByRoute(ctx, fromRouteID)
    if err != nil {
        return err
    }

    // 2. Load field values
    sourceFields, err := db.GetContentFieldsByRoute(ctx, fromRouteID)
    if err != nil {
        return err
    }

    // 3. Create new content in target route
    idMapping := make(map[int64]int64)  // old_id -> new_id

    for _, row := range sourceRows {
        newNode := db.CreateContentData(ctx, CreateContentDataParams{
            RouteID:    toRouteID,  // Different route!
            ParentID:   remapParentID(row.ParentID, idMapping),
            DatatypeID: row.DatatypeID,
            AuthorID:   row.AuthorID,
        })

        idMapping[row.ContentDataID] = newNode.ContentDataID
    }

    // 4. Copy field values
    for _, field := range sourceFields {
        newContentID := idMapping[field.ContentDataID]
        db.CreateContentField(ctx, CreateContentFieldParams{
            ContentDataID: newContentID,
            RouteID:       sql.NullInt64{Int64: toRouteID, Valid: true},
            FieldID:       field.FieldID,
            FieldValue:    field.FieldValue,
            AuthorID:      field.AuthorID,
        })
    }

    return nil
}
```

### Case 2: Route-Specific Permissions

**Scenario:** User can only edit content in specific routes.

```go
type UserRoutePermissions struct {
    UserID   int64
    RouteIDs []int64  // Routes user has access to
}

func CanUserAccessRoute(userID, routeID int64) bool {
    perms := GetUserPermissions(userID)

    for _, allowedRouteID := range perms.RouteIDs {
        if allowedRouteID == routeID {
            return true
        }
    }

    return false
}

// Before allowing content edit
if !CanUserAccessRoute(userID, routeID) {
    return fmt.Errorf("user does not have permission for this route")
}
```

### Case 3: Cross-Route References

**Scenario:** Content in route A references content in route B.

**Challenge:** Foreign keys don't allow cross-route references in content_data tree.

**Solution:** Use content_fields to store cross-route references.

```go
// Create field type "route_reference"
// Store as JSON in field_value
referenceData := map[string]interface{}{
    "target_route_id":         targetRouteID,
    "target_content_data_id":  targetContentID,
}

referenceJSON, _ := json.Marshal(referenceData)

db.CreateContentField(ctx, CreateContentFieldParams{
    ContentDataID: sourceContentID,
    RouteID:       sourceRouteID,
    FieldID:       referenceFieldID,
    FieldValue:    string(referenceJSON),
    AuthorID:      userID,
})
```

---

## Related Documentation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md` - Content tree implementation
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` - Domain model (routes in context)
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/DATABASE_LAYER.md` - DbDriver interface

**Domain:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/DATATYPES_AND_FIELDS.md` - Content schemas
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/CONTENT_TREES.md` - Tree operations (within routes)

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_FEATURES.md` - Feature development
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_TABLES.md` - Schema changes

**Database:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQL_DIRECTORY.md` - SQL organization
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md` - Database operations
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQLC.md` - Query generation

---

## Quick Reference

### Key Concepts
- **Route** = Top-level container for an independent content tree
- **route_id** = Unique identifier for each route
- **slug** = Human-readable identifier, typically domain name
- **Isolation** = Content in one route never interacts with another route

### Essential Tables
- `routes` - Client site routes (public content)
- `admin_routes` - Admin interface routes (admin content)
- `content_data` - Content nodes, each belongs to one route via `route_id`
- `content_fields` - Field values, denormalize `route_id` for performance

### Critical Query Pattern
```sql
-- ALWAYS filter by route_id
WHERE route_id = ?
```

### Key Foreign Keys
```sql
-- Content belongs to route
content_data.route_id REFERENCES routes.route_id ON DELETE CASCADE

-- Delete route = delete all content (CASCADE)
```

### Common Operations
```go
// Get route by domain
routeID, err := db.GetRouteID(ctx, "example.com")

// Load content tree for route
rows, err := db.GetContentTreeByRoute(ctx, *routeID)

// Create content in route
newContent := db.CreateContentData(ctx, CreateContentDataParams{
    RouteID:    routeID,
    ParentID:   parentID,
    DatatypeID: datatypeID,
    AuthorID:   userID,
})
```

### Schema Paths
- Routes schema: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/6_routes/`
- Admin routes schema: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/5_admin_routes/`
- Content data schema: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/16_content_data/`

### Go Types
- Route struct: `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/route.go:16-25`
- Route operations: `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/route.go`

### Multi-Site Pattern
```
One CMS Installation
  ├─> Route 1 (example.com)    → Independent content tree
  ├─> Route 2 (staging.com)    → Independent content tree
  └─> Route 3 (clienta.com)    → Independent content tree
```

### Performance Tips
- Cache route_id after slug lookup
- Use route_id directly in queries (not slug)
- Create indexes on route_id columns
- Use lazy loading for large trees
- Filter by route_id early in query plan
