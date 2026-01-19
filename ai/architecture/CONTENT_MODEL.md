# CONTENT_MODEL.md

Comprehensive documentation on ModulaCMS's content domain model and database relationships.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md`
**Related Schemas:**
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/6_routes/`
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/7_datatypes/`
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/8_fields/`
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/16_content_data/`
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/17_content_fields/`
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/20_datatypes_fields/`

---

## Overview

ModulaCMS uses a **dynamic schema** system that separates content structure definition from content data. This is fundamentally different from traditional CMSs with fixed schemas.

**Core Concept: Schema vs Data**

**Schema (Structure Definition):**
- `datatypes` - Define content types (Page, Post, Product, etc.)
- `fields` - Define properties (Title, Body, Image, Price, etc.)
- `datatypes_fields` - Link fields to datatypes

**Data (Actual Content):**
- `routes` - Top-level containers (like separate websites)
- `content_data` - Content instances arranged in trees
- `content_fields` - Actual field values

**Why This Matters:**
- Enables dynamic content types without code changes
- Similar to WordPress post types, but more flexible
- Supports arbitrary content structures
- Schema can be modified at runtime
- Each site (route) can have different content structures

---

## The Big Picture

### Architecture Diagram

```
Routes (Sites)
  └─> Content Tree (content_data nodes)
       ├─> Each node is an instance of a Datatype
       │    └─> Datatype defines which Fields it has
       │         └─> Fields define property types (text, image, etc.)
       └─> Content Fields (content_fields) store actual values
```

### Example: Blog Site

**Schema Definition:**
```
Datatype: "Page"
  ├─> Field: "Title" (type: text)
  ├─> Field: "Body" (type: richtext)
  └─> Field: "Favicon" (type: image)

Datatype: "Post"
  ├─> Field: "Title" (type: text)
  ├─> Field: "Excerpt" (type: text)
  ├─> Field: "Body" (type: richtext)
  └─> Field: "Featured Image" (type: image)
```

**Actual Content:**
```
Route: "example.com"
  └─> Content Tree:
       ├─> Home (instance of "Page" datatype)
       │    ├─> Title field: "Welcome"
       │    ├─> Body field: "<p>Welcome to our site...</p>"
       │    └─> Favicon field: "/images/favicon.png"
       ├─> Blog (instance of "Page" datatype)
       │    └─> First Post (instance of "Post" datatype)
       │         ├─> Title field: "Hello World"
       │         └─> Body field: "<p>This is my first post...</p>"
       └─> About (instance of "Page" datatype)
```

---

## Routes: Top-Level Containers

### Purpose

Routes represent **independent content trees**. Think of them as separate websites or sections that don't share content hierarchy.

**Use Cases:**
- Multi-site CMS: Different domains on same installation
- Content segregation: Public site vs internal documentation
- Separate client sites: One CMS installation, multiple clients
- Testing: Production site vs staging site

### Database Schema

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/6_routes/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS routes (
    route_id INTEGER PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,           -- Domain or identifier (e.g., "example.com")
    title TEXT NOT NULL,                 -- Human-readable name
    status INTEGER NOT NULL,             -- Active, inactive, etc.
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
```

**Key Fields:**
- `route_id` - Unique identifier
- `slug` - Domain name or unique identifier (e.g., "example.com", "staging")
- `title` - Human-readable name (e.g., "Main Website", "Client ABC")
- `status` - Whether route is active

### Route Isolation

**Critical:** Content in different routes is completely isolated:
- Each `content_data` row belongs to exactly one route
- Tree structure doesn't span routes
- Query by route_id to get one site's content
- Enables true multi-site architecture

**Example Routes:**
```sql
INSERT INTO routes (slug, title, status, author_id) VALUES
  ('example.com', 'Main Website', 1, 1),
  ('staging.example.com', 'Staging Site', 1, 1),
  ('client-a.com', 'Client A Site', 1, 1);
```

---

## Datatypes: Content Schemas

### Purpose

Datatypes define **what kind of content** can exist. They are the schema templates that content instances follow.

**Analogy:**
- Like "classes" in object-oriented programming
- Like "post types" in WordPress
- Like "content types" in Drupal
- Like "models" in Django

**Difference from Traditional CMSs:**
- Not hardcoded in application
- Stored in database
- Can be created/modified at runtime
- Each datatype can have different fields

### Database Schema

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/7_datatypes/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS datatypes(
    datatype_id INTEGER PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES datatypes ON DELETE SET DEFAULT,
    label TEXT NOT NULL,                 -- "Page", "Post", "Product"
    type TEXT NOT NULL,                  -- "ROOT", "ELEMENT", etc.
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
```

**Key Fields:**
- `datatype_id` - Unique identifier
- `parent_id` - Optional parent datatype (for hierarchical types)
- `label` - Human-readable name ("Page", "Post", "Hero Section")
- `type` - Classification ("ROOT", "ELEMENT", "CONTAINER")

### Datatype Types

**ROOT:**
- Top-level content types
- Can be root nodes in content tree
- Examples: Page, Post, Product, Category

**ELEMENT:**
- Leaf content types
- Typically don't have children
- Examples: Paragraph, Image, Button, Text Block

**CONTAINER:**
- Can contain other elements
- Structural content types
- Examples: Section, Column, Row, Grid

### Datatype Hierarchy

Datatypes can have parent-child relationships via `parent_id`:

```
Datatype: "Page" (ROOT)
  └─> Datatype: "Hero Section" (CONTAINER)
       └─> Datatype: "Hero Title" (ELEMENT)
```

**Purpose:** Organize datatypes into categories, not enforce content structure.

### Go Structure

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/datatype.go:18-28`

```go
type Datatypes struct {
    DatatypeID   int64          `json:"datatype_id"`
    ParentID     sql.NullInt64  `json:"parent_id"`
    Label        string         `json:"label"`
    Type         string         `json:"type"`
    AuthorID     int64          `json:"author_id"`
    DateCreated  sql.NullString `json:"date_created"`
    DateModified sql.NullString `json:"date_modified"`
    History      sql.NullString `json:"history"`
}
```

---

## Fields: Property Definitions

### Purpose

Fields define **individual properties** that datatypes can have. They are the building blocks of content schemas.

**Analogy:**
- Like "fields" in database tables
- Like "properties" in classes
- Like "attributes" in HTML elements
- Like "columns" in spreadsheets

**Examples:**
- Title (text input)
- Body (rich text editor)
- Featured Image (image upload)
- Price (number input)
- Published Date (date picker)
- Author (user reference)

### Database Schema

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/8_fields/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS fields(
    field_id INTEGER PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES datatypes ON DELETE SET DEFAULT,
    label TEXT DEFAULT 'unlabeled' NOT NULL,  -- "Title", "Body", "Price"
    data TEXT NOT NULL,                        -- Additional metadata (JSON)
    type TEXT NOT NULL,                        -- "text", "richtext", "image", "number"
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
```

**Key Fields:**
- `field_id` - Unique identifier
- `parent_id` - Optional parent datatype (legacy, use datatypes_fields instead)
- `label` - Human-readable name ("Title", "Body", "Featured Image")
- `type` - Field type ("text", "richtext", "image", "number", "date")
- `data` - Additional metadata as JSON (validation rules, defaults, etc.)

### Field Types

**Common Field Types:**
- `text` - Single-line text input
- `richtext` - HTML/Markdown editor
- `image` - Image upload (S3 reference)
- `number` - Numeric input
- `date` - Date/time picker
- `boolean` - Checkbox/toggle
- `select` - Dropdown selection
- `reference` - Reference to another content item

**Type Determines:**
- TUI input widget
- Validation rules
- Storage format
- Display rendering

### Field Metadata (data column)

The `data` column stores JSON metadata:

```json
{
  "placeholder": "Enter title here",
  "maxLength": 200,
  "required": true,
  "validation": "^[A-Za-z0-9 ]+$",
  "helpText": "Title appears in browser tab"
}
```

### Go Structure

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/field.go:17-28`

```go
type Fields struct {
    FieldID      int64          `json:"field_id"`
    ParentID     sql.NullInt64  `json:"parent_id"`
    Label        string         `json:"label"`
    Data         string         `json:"data"`          // JSON metadata
    Type         string         `json:"type"`
    AuthorID     int64          `json:"author_id"`
    DateCreated  sql.NullString `json:"date_created"`
    DateModified sql.NullString `json:"date_modified"`
    History      sql.NullString `json:"history"`
}
```

---

## Datatypes_Fields: Junction Table

### Purpose

Links datatypes to their fields. Defines **which fields belong to which datatypes**.

**Why Junction Table?**
- Many-to-many relationship
- One datatype can have many fields
- One field can belong to many datatypes (reusable fields)
- Enables field reuse across datatypes

### Database Schema

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/20_datatypes_fields/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS datatypes_fields (
    id INTEGER PRIMARY KEY,
    datatype_id INTEGER NOT NULL
        CONSTRAINT fk_df_datatype
            REFERENCES datatypes ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        CONSTRAINT fk_df_field
            REFERENCES fields ON DELETE CASCADE
);
```

**Key Fields:**
- `id` - Unique identifier for junction record
- `datatype_id` - References datatypes table
- `field_id` - References fields table

**Cascade Deletion:**
- Deleting a datatype removes all its field associations
- Deleting a field removes all its datatype associations

### Example Relationships

**"Page" Datatype:**
```sql
INSERT INTO datatypes (label, type) VALUES ('Page', 'ROOT');  -- datatype_id = 1

INSERT INTO fields (label, type) VALUES
  ('Title', 'text'),      -- field_id = 1
  ('Body', 'richtext'),   -- field_id = 2
  ('Favicon', 'image');   -- field_id = 3

INSERT INTO datatypes_fields (datatype_id, field_id) VALUES
  (1, 1),  -- Page has Title
  (1, 2),  -- Page has Body
  (1, 3);  -- Page has Favicon
```

**"Post" Datatype (Reusing Title Field):**
```sql
INSERT INTO datatypes (label, type) VALUES ('Post', 'ROOT');  -- datatype_id = 2

INSERT INTO fields (label, type) VALUES
  ('Excerpt', 'text'),         -- field_id = 4
  ('Featured Image', 'image'); -- field_id = 5

INSERT INTO datatypes_fields (datatype_id, field_id) VALUES
  (2, 1),  -- Post has Title (reused from Page!)
  (2, 4),  -- Post has Excerpt
  (2, 2),  -- Post has Body (reused!)
  (2, 5);  -- Post has Featured Image
```

**Result:** "Title" and "Body" fields are shared between Page and Post datatypes.

### Query: Get Fields for Datatype

```sql
SELECT f.*
FROM fields f
JOIN datatypes_fields df ON f.field_id = df.field_id
WHERE df.datatype_id = ?
ORDER BY df.id;  -- Maintains field order
```

---

## Content_Data: Content Instances

### Purpose

Instances of datatypes arranged in a tree structure. These are the **actual content items** in your site.

**Relationship:**
- Each `content_data` row is an **instance** of a datatype
- Like "objects" from "classes"
- Like "posts" from "post types"
- Like "rows" from "table schemas"

**Critical:** content_data uses the sibling-pointer tree structure (see TREE_STRUCTURE.md).

### Database Schema

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/16_content_data/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INTEGER PRIMARY KEY,

    -- Tree structure (sibling pointers)
    parent_id INTEGER
        REFERENCES content_data ON DELETE SET NULL,
    first_child_id INTEGER
        REFERENCES content_data ON DELETE SET NULL,
    next_sibling_id INTEGER
        REFERENCES content_data ON DELETE SET NULL,
    prev_sibling_id INTEGER
        REFERENCES content_data ON DELETE SET NULL,

    -- Content metadata
    route_id INTEGER NOT NULL
        REFERENCES routes ON DELETE CASCADE,
    datatype_id INTEGER NOT NULL
        REFERENCES datatypes ON DELETE SET NULL,
    author_id INTEGER NOT NULL
        REFERENCES users ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
```

**Key Fields:**
- `content_data_id` - Unique identifier for this content item
- Tree pointers: `parent_id`, `first_child_id`, `next_sibling_id`, `prev_sibling_id`
- `route_id` - Which site/route this content belongs to
- `datatype_id` - Which datatype this is an instance of
- `author_id` - Who created this content

**Important:** content_data stores NO actual field values. It only defines:
- What type of content it is (datatype_id)
- Where it sits in the tree (parent/child/sibling pointers)
- Which site it belongs to (route_id)

### Example Content Tree

```sql
-- Create a "Home" page (instance of Page datatype)
INSERT INTO content_data (route_id, datatype_id, parent_id, author_id)
VALUES (1, 1, NULL, 1);  -- content_data_id = 10, root node

-- Create "About" page (sibling of Home)
INSERT INTO content_data (route_id, datatype_id, parent_id, author_id)
VALUES (1, 1, NULL, 1);  -- content_data_id = 11

-- Create "Hero Section" (child of Home)
INSERT INTO content_data (route_id, datatype_id, parent_id, author_id)
VALUES (1, 5, 10, 1);  -- content_data_id = 12, parent is Home
```

**Result:**
```
Route: example.com
  ├─> Home (content_data_id=10, datatype="Page")
  │    └─> Hero Section (content_data_id=12, datatype="Hero Section")
  └─> About (content_data_id=11, datatype="Page")
```

### Go Structure

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/content_data.go:17-32`

```go
type ContentData struct {
    ContentDataID int64          `json:"content_data_id"`
    ParentID      sql.NullInt64  `json:"parent_id"`
    FirstChildID  sql.NullInt64  `json:"first_child_id"`
    NextSiblingID sql.NullInt64  `json:"next_sibling_id"`
    PrevSiblingID sql.NullInt64  `json:"prev_sibling_id"`
    RouteID       int64          `json:"route_id"`
    DatatypeID    int64          `json:"datatype_id"`
    AuthorID      int64          `json:"author_id"`
    DateCreated   sql.NullString `json:"date_created"`
    DateModified  sql.NullString `json:"date_modified"`
    History       sql.NullString `json:"history"`
}
```

---

## Content_Fields: Actual Field Values

### Purpose

Stores the **actual values** for fields on content items.

**Relationship:**
- Each `content_fields` row is a field value for a `content_data` instance
- Links to both `content_data` (which content item) and `fields` (which field)
- Stores the actual text, number, image path, etc.

### Database Schema

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/17_content_fields/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id INTEGER PRIMARY KEY,
    route_id INTEGER
        REFERENCES routes ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id INTEGER NOT NULL
        REFERENCES content_data ON UPDATE CASCADE ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        REFERENCES fields ON UPDATE CASCADE ON DELETE CASCADE,
    field_value TEXT NOT NULL,           -- The actual value!
    author_id INTEGER NOT NULL
        REFERENCES users ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
```

**Key Fields:**
- `content_field_id` - Unique identifier
- `content_data_id` - Which content item this belongs to
- `field_id` - Which field definition this is a value for
- `field_value` - The actual value (stored as TEXT)
- `route_id` - Denormalized for query performance

**Value Storage:**
All values stored as TEXT:
- Text fields: Direct string
- Numbers: String representation ("123.45")
- Images: S3 path ("/bucket/key/image.jpg")
- Dates: ISO 8601 or UNIX timestamp
- JSON: Serialized JSON string

### Example Field Values

```sql
-- Home page (content_data_id = 10) is an instance of Page datatype
-- Page datatype has fields: Title (field_id=1), Body (field_id=2), Favicon (field_id=3)

INSERT INTO content_fields (content_data_id, field_id, field_value, route_id, author_id)
VALUES
  (10, 1, 'Welcome to Our Site', 1, 1),           -- Title
  (10, 2, '<p>This is the home page...</p>', 1, 1),  -- Body
  (10, 3, '/images/favicon.png', 1, 1);           -- Favicon
```

**Result:** Home page now has:
- Title: "Welcome to Our Site"
- Body: "<p>This is the home page...</p>"
- Favicon: "/images/favicon.png"

### Query: Get All Fields for Content Item

```sql
SELECT
    f.field_id,
    f.label as field_label,
    f.type as field_type,
    cf.field_value
FROM content_fields cf
JOIN fields f ON cf.field_id = f.field_id
WHERE cf.content_data_id = ?;
```

**Result:**
```
field_id | field_label | field_type | field_value
---------|-------------|------------|-------------------
1        | Title       | text       | Welcome to Our Site
2        | Body        | richtext   | <p>This is...</p>
3        | Favicon     | image      | /images/favicon.png
```

### Go Structure

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/content_field.go:17-30`

```go
type ContentFields struct {
    ContentFieldID int64          `json:"content_field_id"`
    RouteID        sql.NullInt64  `json:"route_id"`
    ContentDataID  int64          `json:"content_data_id"`
    FieldID        int64          `json:"field_id"`
    FieldValue     string         `json:"field_value"`
    AuthorID       int64          `json:"author_id"`
    DateCreated    sql.NullString `json:"date_created"`
    DateModified   sql.NullString `json:"date_modified"`
    History        sql.NullString `json:"history"`
}
```

---

## Complete Example: Blog Post

Let's walk through creating a complete blog post from schema to data.

### Step 1: Define Schema

**Create Datatype:**
```sql
INSERT INTO datatypes (label, type, author_id)
VALUES ('Blog Post', 'ROOT', 1);
-- Result: datatype_id = 100
```

**Create Fields:**
```sql
INSERT INTO fields (label, type, author_id) VALUES
  ('Post Title', 'text', 1),        -- field_id = 50
  ('Post Body', 'richtext', 1),     -- field_id = 51
  ('Publish Date', 'date', 1),      -- field_id = 52
  ('Featured Image', 'image', 1);   -- field_id = 53
```

**Link Fields to Datatype:**
```sql
INSERT INTO datatypes_fields (datatype_id, field_id) VALUES
  (100, 50),  -- Blog Post has Post Title
  (100, 51),  -- Blog Post has Post Body
  (100, 52),  -- Blog Post has Publish Date
  (100, 53);  -- Blog Post has Featured Image
```

### Step 2: Create Content Instance

**Create Content Data (Instance):**
```sql
INSERT INTO content_data (route_id, datatype_id, parent_id, author_id)
VALUES (1, 100, NULL, 1);
-- Result: content_data_id = 500 (new blog post instance)
```

### Step 3: Populate Field Values

**Add Field Values:**
```sql
INSERT INTO content_fields (content_data_id, field_id, field_value, route_id, author_id)
VALUES
  (500, 50, 'My First Blog Post', 1, 1),
  (500, 51, '<p>This is the content of my blog post...</p>', 1, 1),
  (500, 52, '2026-01-12T10:30:00Z', 1, 1),
  (500, 53, '/media/blog/featured-image.jpg', 1, 1);
```

### Step 4: Query Complete Post

```sql
SELECT
    cd.content_data_id,
    dt.label as datatype_label,
    f.label as field_label,
    f.type as field_type,
    cf.field_value
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
LEFT JOIN content_fields cf ON cd.content_data_id = cf.content_data_id
LEFT JOIN fields f ON cf.field_id = f.field_id
WHERE cd.content_data_id = 500;
```

**Result:**
```
content_data_id | datatype_label | field_label     | field_type | field_value
----------------|----------------|-----------------|------------|----------------------------
500             | Blog Post      | Post Title      | text       | My First Blog Post
500             | Blog Post      | Post Body       | richtext   | <p>This is the content...</p>
500             | Blog Post      | Publish Date    | date       | 2026-01-12T10:30:00Z
500             | Blog Post      | Featured Image  | image      | /media/blog/featured-image.jpg
```

---

## Admin Tables: Parallel Structure

ModulaCMS has a complete parallel set of tables for admin interface content:

**Admin Tables:**
- `admin_routes` - Admin interface sites
- `admin_datatypes` - Admin content types
- `admin_fields` - Admin field definitions
- `admin_datatypes_fields` - Admin junction table
- `admin_content_data` - Admin content instances
- `admin_content_fields` - Admin field values

### Why Separate Admin Tables?

**1. Content Isolation**
- Public site content in regular tables
- Admin interface content in admin_* tables
- No risk of mixing admin UI with public content

**2. Different Schemas**
- Admin interface might have different datatypes
- Different fields needed for admin forms
- Different structure requirements

**3. Security**
- Separate permissions for admin content
- Different access control rules
- Isolation prevents accidental exposure

**4. Flexibility**
- Admin interface can evolve independently
- Public schema changes don't affect admin
- Different optimization strategies

### Example: Admin Dashboard

```sql
-- Create admin route for dashboard
INSERT INTO admin_routes (slug, title, status, author_id)
VALUES ('admin-dashboard', 'Dashboard', 1, 1);

-- Create admin datatype for dashboard widgets
INSERT INTO admin_datatypes (label, type, author_id)
VALUES ('Dashboard Widget', 'ELEMENT', 1);

-- Create widget instance
INSERT INTO admin_content_data (admin_route_id, admin_datatype_id, author_id)
VALUES (1, 1, 1);
```

**Critical:** Admin and public content structures are completely independent.

---

## Schema vs Data: The Key Insight

### Schema Definition (Meta-Layer)

**Tables:**
- `datatypes` - What kinds of content exist
- `fields` - What properties content can have
- `datatypes_fields` - Which properties belong to which kinds

**Modified By:**
- Administrators
- Developers
- Content architects

**Frequency:** Infrequently (setup phase, major changes)

**Example Questions Answered:**
- What content types exist? (Page, Post, Product)
- What fields does a Page have? (Title, Body, Favicon)
- What type is the Title field? (text)

### Data Storage (Content Layer)

**Tables:**
- `routes` - Which sites exist
- `content_data` - Actual content items
- `content_fields` - Actual field values

**Modified By:**
- Content editors
- Authors
- End users

**Frequency:** Constantly (daily content updates)

**Example Questions Answered:**
- What pages exist? (Home, About, Contact)
- What is the title of the Home page? ("Welcome to Our Site")
- What content is under the Blog section? (List of posts)

### The Power of Separation

**Benefits:**
1. **Runtime Schema Changes**
   - Add new content types without code deployment
   - Modify field definitions without migrations
   - Adjust content structure dynamically

2. **Multi-Tenant Support**
   - Same schema definitions across all routes
   - Or different schemas per route
   - Flexible content modeling

3. **Content Reusability**
   - Fields shared across datatypes
   - Consistent field behavior
   - Reduced duplication

4. **Dynamic Forms**
   - TUI generates forms from field definitions
   - No hardcoded form layouts
   - Automatic validation from field metadata

**Trade-offs:**
- More complex queries (multiple joins)
- Requires understanding of meta-model
- Performance considerations for large datasets

---

## Why This Model Needs Performance Optimization

**Critical Understanding:** The dynamic schema model creates significant row proliferation. This is why ModulaCMS uses aggressive performance optimizations (sibling-pointer trees, lazy loading, O(1) operations).

### Row Count Math: The Reality of Dynamic Schemas

**Example: Single Page Instance**

A typical page datatype might have:
- **Core fields:** 23 (slug, title, body, author, dates, status, etc.)
- **Meta fields:** 6 (SEO title, meta description, og:image, etc.)
- **Total fields:** 29

**Database rows required:**
- 1 row in `content_data` (the page instance itself)
- 29 rows in `content_fields` (one per field value)
- **Total: 30 rows per page**

**Schema definition (shared across all pages):**
- 1 row in `datatypes` (Page datatype definition)
- 29 rows in `fields` (field definitions)
- 29 rows in `datatypes_fields` (junction records)
- **Total: 59 rows for schema**

### Scale Comparison

| Site Size | Pages | Content Rows | Total Rows | Query Complexity |
|-----------|-------|--------------|------------|------------------|
| Demo | 10 | 300 | 359 | Negligible |
| Small Site | 100 | 3,000 | 3,059 | Noticeable |
| College | 1,000 | 30,000 | 30,059 | Critical |
| Enterprise | 10,000 | 300,000 | 300,059 | Make-or-break |

**Key Insight:** A modest 1,000-page site generates 30,000+ content field rows. Without optimization, the system becomes unusable.

### The Join Cost Amplification

**Every content query requires multiple joins:**

**Single page load:**
```sql
-- Get page tree node
SELECT * FROM content_data WHERE content_data_id = ?

-- Get all field values (29 rows)
SELECT cf.*, f.label, f.type
FROM content_fields cf
JOIN fields f ON cf.field_id = f.field_id
WHERE cf.content_data_id = ?

-- Get datatype info
SELECT * FROM datatypes WHERE datatype_id = ?
```

**Tree of 100 pages:**
```sql
-- Get all content nodes
SELECT * FROM content_data WHERE route_id = ?  -- 100 rows

-- Get all field values
SELECT cf.*, f.label, f.type
FROM content_fields cf
JOIN fields f ON cf.field_id = f.field_id
WHERE cf.content_data_id IN (...)  -- 2,900 rows

-- Join to datatypes for type info
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
```

**Result:** Loading a content tree of 100 pages requires:
- 100 content_data rows
- 2,900 content_fields rows
- 29 field definitions (cached)
- Multiple joins across tables
- Tree traversal operations

**Without O(1) tree operations**, every parent-child navigation is O(n), compounding the join cost.

### Headless Architecture: Network Amplification Factor

ModulaCMS is a **headless CMS** serving content over HTTP/HTTPS APIs:

**Traditional CMS (WordPress, Drupal):**
```
Database → PHP/Server → Template Render → HTML
(All on same server, ~10ms latency)
```

**Headless CMS (ModulaCMS):**
```
Database → Go Backend → Network → Frontend (Client/Admin) → Render
(Network adds 50-200ms latency per request)
```

**Implications:**
- **Query time is amplified:** 100ms DB query becomes 100ms + 100ms network = 200ms
- **Multiple requests compound:** Loading tree + fields = 2× network penalty
- **Can't hide with SSR:** Frontend is separate, can't optimize away slow queries
- **Users feel the pain:** Every millisecond of backend slowness is visible

### Supporting Two Frontends with Different Patterns

**Client Frontend (Public Site):**
- **Query Pattern:** Read-heavy, specific content items
- **Needs:** Fast individual page loads, search, filtering
- **Users:** Site visitors with high performance expectations
- **Typical:** "Get page by slug" → 1 content_data + 29 content_fields

**Admin Frontend (CMS Interface):**
- **Query Pattern:** Mixed read/write, tree traversal, bulk operations
- **Needs:** Tree navigation, CRUD operations, content organization
- **Users:** Content editors working all day
- **Typical:** "Get entire content tree" → 1,000 content_data + 29,000 content_fields

**Challenge:** Backend must be fast for BOTH patterns:
- Can't optimize tree loading at the expense of single-page queries
- Can't optimize single-page queries at the expense of tree operations
- Must support efficient CRUD operations while maintaining read performance

### Agency Use Case: Unknown Frontend Implementation

**Key Constraint:** Agencies build custom frontends that ModulaCMS has no control over.

**Frontend unknowns:**
- Framework: React, Vue, Svelte, vanilla JS, or something new
- Query strategy: Smart batching vs naive individual requests
- Caching: Redis, in-memory, service worker, or none
- Memory constraints: Desktop browser vs mobile vs embedded
- Developer skill: Expert optimization vs copy-paste code

**Backend requirements:**
- Must be fast by default (can't rely on smart frontend)
- Must handle naive query patterns without falling over
- Must support efficient batching for smart frontends
- Must scale regardless of frontend implementation quality

**This is why:**
- Tree operations must be O(1) - don't know how frontend navigates
- Lazy loading must work - don't know frontend memory constraints
- NodeIndex map is essential - don't know lookup patterns
- Database performance is non-negotiable - it's the foundation

### Field Model Creates Exponential Query Growth

**The problem:** Content_fields table grows exponentially with content.

**Traditional CMS (Fixed Schema):**
```sql
SELECT * FROM pages WHERE id = ?  -- 1 row, all fields in columns
```

**ModulaCMS (Dynamic Schema):**
```sql
-- Step 1: Get page
SELECT * FROM content_data WHERE content_data_id = ?  -- 1 row

-- Step 2: Get field values
SELECT cf.*, f.label, f.type
FROM content_fields cf
JOIN fields f ON cf.field_id = f.field_id
WHERE cf.content_data_id = ?  -- 29 rows

-- Step 3: Assemble in application
-- Map field_id to field values, structure as JSON
```

**Query complexity:**
- Traditional: O(1) - single row lookup
- ModulaCMS: O(f) where f = number of fields per datatype
- For 100 pages: Traditional = 100 queries, ModulaCMS = 100 + (100 × 29) = 3,000 rows

**Why dynamic schema is still worth it:**
- Runtime schema changes without migrations
- Flexible content modeling
- Multi-tenant support with different schemas
- Agency customization without code changes

**But it requires:**
- Aggressive query optimization
- O(1) tree operations (can't afford O(n) on top of O(f))
- Lazy loading (can't load all 3,000 rows upfront)
- Smart caching (schema definitions cached, content queried on-demand)

### Real-World Performance Requirements

**Target performance (API endpoint response times):**
- Get single page: < 50ms
- Get content tree (100 nodes): < 200ms
- Create new content: < 100ms
- Update field value: < 50ms
- Navigate to sibling: < 10ms (no DB query, memory operation)

**Without optimizations:**
- Get single page: 150ms (30 row joins)
- Get content tree: 2,000ms+ (O(n) tree traversal + joins)
- Navigate to sibling: 50ms (query parent's children)

**With optimizations (sibling pointers, lazy load, NodeIndex):**
- Get single page: 30ms (optimized joins, indexed lookups)
- Get content tree: 180ms (O(1) tree ops, lazy load visible nodes only)
- Navigate to sibling: 1ms (pointer access, no DB query)

### The Design Philosophy: Scale-First Architecture

**For small demos:** Optimizations appear excessive.
- 10 pages = 300 rows
- Any algorithm works fine
- Could use simple parent-child arrays
- Could load entire tree upfront

**For production sites:** Optimizations are essential.
- 1,000 pages = 30,000 rows
- O(n) algorithms cause timeouts
- Must use sibling pointers for O(1) ops
- Must lazy load to avoid memory exhaustion

**ModulaCMS philosophy:**
- Design for scale from day one
- Agencies need production-ready systems
- Dynamic schema flexibility requires compensating optimizations
- "Over-engineered" for demos, "appropriately-engineered" for reality

**The trade-off:**
- Higher complexity in tree implementation
- More sophisticated loading algorithms
- Steeper learning curve for developers
- **Benefit:** System that actually works at scale without rewrites

### Summary: Why Performance Optimizations Aren't Optional

1. **Dynamic schema creates 30× row multiplication** (1 page = 30 rows)
2. **Multiple joins compound query cost** (content_data + content_fields + fields + datatypes)
3. **Headless architecture amplifies latency** (network adds 50-200ms per request)
4. **Two frontends with different patterns** (can't optimize for just one)
5. **Unknown frontend implementations** (agencies build unpredictable query patterns)
6. **Scale is the design goal** (1,000+ page sites are the target, not 10-page demos)

**Result:** Sibling-pointer trees, lazy loading, O(1) operations, and NodeIndex maps aren't premature optimization—they're the minimum viable architecture for the dynamic schema model at production scale.

---

## Querying Patterns

### Pattern 1: Get Content with Fields

**Goal:** Retrieve a content item with all its field values.

```sql
SELECT
    cd.content_data_id,
    cd.route_id,
    dt.label as content_type,
    f.label as field_name,
    f.type as field_type,
    cf.field_value
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
LEFT JOIN content_fields cf ON cd.content_data_id = cf.content_data_id
LEFT JOIN fields f ON cf.field_id = f.field_id
WHERE cd.content_data_id = ?;
```

**Performance:** Multiple joins required. Consider using database views or materialized queries for frequently accessed patterns.

### Pattern 2: Get All Content of Type

**Goal:** List all instances of a specific datatype.

```sql
SELECT
    cd.content_data_id,
    cd.date_created,
    GROUP_CONCAT(f.label || ': ' || cf.field_value, ' | ') as fields
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
LEFT JOIN content_fields cf ON cd.content_data_id = cf.content_data_id
LEFT JOIN fields f ON cf.field_id = f.field_id
WHERE dt.label = 'Blog Post'
  AND cd.route_id = ?
GROUP BY cd.content_data_id
ORDER BY cd.date_created DESC;
```

### Pattern 3: Get Schema Definition

**Goal:** Understand what fields a datatype has.

```sql
SELECT
    dt.label as datatype_name,
    f.field_id,
    f.label as field_name,
    f.type as field_type,
    f.data as field_metadata
FROM datatypes dt
JOIN datatypes_fields df ON dt.datatype_id = df.datatype_id
JOIN fields f ON df.field_id = f.field_id
WHERE dt.label = 'Page'
ORDER BY df.id;  -- Maintains field order
```

### Pattern 4: Content Tree with Types

**Goal:** Load content tree with datatype information.

```sql
SELECT
    cd.content_data_id,
    cd.parent_id,
    cd.first_child_id,
    cd.next_sibling_id,
    cd.prev_sibling_id,
    dt.label as datatype_label,
    dt.type as datatype_type
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id = ?
ORDER BY cd.parent_id NULLS FIRST, cd.content_data_id;
```

**Usage:** Feed this to tree loading algorithm (see TREE_STRUCTURE.md).

### Pattern 5: Search Content by Field Value

**Goal:** Find content where a specific field matches a value.

```sql
SELECT DISTINCT
    cd.content_data_id,
    dt.label as content_type,
    cf.field_value
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
JOIN content_fields cf ON cd.content_data_id = cf.content_data_id
JOIN fields f ON cf.field_id = f.field_id
WHERE f.label = 'Title'
  AND cf.field_value LIKE '%search term%'
  AND cd.route_id = ?;
```

---

## Creating New Content: Step-by-Step

### Scenario: Add a Product Catalog

**Step 1: Define Product Datatype**

```go
// Create Product datatype
datatypeID, err := db.CreateDatatype(ctx, CreateDatatypeParams{
    Label:    "Product",
    Type:     "ROOT",
    AuthorID: userID,
})
```

**Step 2: Define Product Fields**

```go
// Create fields
fields := []struct{Label, Type string}{
    {"Product Name", "text"},
    {"Price", "number"},
    {"Description", "richtext"},
    {"SKU", "text"},
    {"Product Image", "image"},
    {"In Stock", "boolean"},
}

fieldIDs := []int64{}
for _, field := range fields {
    fieldID, err := db.CreateField(ctx, CreateFieldParams{
        Label:    field.Label,
        Type:     field.Type,
        AuthorID: userID,
    })
    fieldIDs = append(fieldIDs, fieldID)
}
```

**Step 3: Link Fields to Datatype**

```go
// Associate fields with datatype
for _, fieldID := range fieldIDs {
    err := db.CreateDatatypeField(ctx, CreateDatatypeFieldParams{
        DatatypeID: datatypeID,
        FieldID:    fieldID,
    })
}
```

**Step 4: Create Product Instance**

```go
// Create a product content item
contentDataID, err := db.CreateContentData(ctx, CreateContentDataParams{
    RouteID:     routeID,
    DatatypeID:  datatypeID,
    ParentID:    sql.NullInt64{Valid: false},  // Root node
    AuthorID:    userID,
})
```

**Step 5: Populate Field Values**

```go
// Add field values
values := map[string]string{
    "Product Name":  "Premium Widget",
    "Price":         "49.99",
    "Description":   "<p>High quality widget...</p>",
    "SKU":           "WDG-001",
    "Product Image": "/products/widget-001.jpg",
    "In Stock":      "true",
}

for fieldLabel, fieldValue := range values {
    // Look up field_id by label
    field, err := db.GetFieldByLabel(ctx, fieldLabel)

    // Create content field
    err = db.CreateContentField(ctx, CreateContentFieldParams{
        ContentDataID: contentDataID,
        FieldID:       field.FieldID,
        FieldValue:    fieldValue,
        RouteID:       routeID,
        AuthorID:      userID,
    })
}
```

**Result:** Complete product with all fields populated and queryable.

---

## Modifying Schema at Runtime

### Example: Add New Field to Existing Datatype

**Scenario:** Page datatype needs a "Meta Description" field for SEO.

```go
// 1. Create the new field
fieldID, err := db.CreateField(ctx, CreateFieldParams{
    Label:    "Meta Description",
    Type:     "text",
    Data:     `{"maxLength": 160, "placeholder": "SEO description"}`,
    AuthorID: userID,
})

// 2. Get Page datatype
pageDatatype, err := db.GetDatatypeByLabel(ctx, "Page")

// 3. Associate field with datatype
err = db.CreateDatatypeField(ctx, CreateDatatypeFieldParams{
    DatatypeID: pageDatatype.DatatypeID,
    FieldID:    fieldID,
})

// 4. Existing Page instances automatically have access to new field
// No migration needed!
// Content editors can now add meta descriptions to pages
```

**What Happens:**
- Existing pages don't have this field populated yet
- New pages can have the field populated
- Field appears in TUI forms automatically
- No code changes required
- No database migration for existing content

---

## Performance Considerations

### Indexing Strategy

**Recommended Indexes:**

```sql
-- Content data indexes
CREATE INDEX idx_content_data_route ON content_data(route_id);
CREATE INDEX idx_content_data_datatype ON content_data(datatype_id);
CREATE INDEX idx_content_data_parent ON content_data(parent_id);

-- Content fields indexes
CREATE INDEX idx_content_fields_data ON content_fields(content_data_id);
CREATE INDEX idx_content_fields_field ON content_fields(field_id);
CREATE INDEX idx_content_fields_value ON content_fields(field_value);

-- Junction table index
CREATE INDEX idx_datatypes_fields_dt ON datatypes_fields(datatype_id);
CREATE INDEX idx_datatypes_fields_f ON datatypes_fields(field_id);
```

### Query Optimization

**Problem:** Multiple joins slow down queries.

**Solutions:**

1. **Materialized Views:**
```sql
CREATE VIEW content_with_fields AS
SELECT
    cd.content_data_id,
    cd.route_id,
    dt.label as datatype,
    json_group_array(
        json_object(
            'field_name', f.label,
            'field_type', f.type,
            'field_value', cf.field_value
        )
    ) as fields
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
LEFT JOIN content_fields cf ON cd.content_data_id = cf.content_data_id
LEFT JOIN fields f ON cf.field_id = f.field_id
GROUP BY cd.content_data_id;
```

2. **Denormalization:**
Store frequently accessed field values directly in content_data (trade-off: schema flexibility vs performance).

3. **Caching:**
Cache schema definitions (datatypes, fields, datatypes_fields) in memory since they change infrequently.

4. **Batch Loading:**
Load all content_fields for a route in one query, assemble in application:

```sql
SELECT
    cf.content_data_id,
    f.field_id,
    f.label,
    f.type,
    cf.field_value
FROM content_fields cf
JOIN fields f ON cf.field_id = f.field_id
WHERE cf.route_id = ?;
```

Then group by content_data_id in Go code.

---

## Comparison to Other Systems

### vs WordPress

**WordPress:**
- Hardcoded post types in code
- Custom post types via code registration
- Custom fields via plugins (ACF, CMB2)
- Schema changes require code deployment

**ModulaCMS:**
- Dynamic datatypes in database
- Runtime datatype creation
- Built-in flexible fields
- Schema changes via TUI

### vs Drupal

**Drupal:**
- Content types via admin UI
- Field types extensible via modules
- Schema stored in database
- Complex field API

**ModulaCMS:**
- Similar concept (content types = datatypes)
- Simpler field system
- Direct database access
- No field API complexity

### vs Contentful/Strapi

**Headless CMSs:**
- API-first design
- Content models via admin UI
- Field definitions in database
- GraphQL/REST APIs

**ModulaCMS:**
- Similar flexibility
- SSH TUI instead of web UI
- Designed for agencies building custom frontends
- S3 storage required (not optional)

---

## Common Pitfalls

### Pitfall 1: Missing Field Definitions

**Problem:** Creating content_fields without corresponding datatypes_fields entry.

**Symptom:**
```sql
-- This creates an orphaned field value
INSERT INTO content_fields (content_data_id, field_id, field_value)
VALUES (100, 50, 'Some value');
-- But field_id 50 is not associated with the datatype of content_data_id 100
```

**Fix:** Always verify field belongs to datatype:
```sql
SELECT 1
FROM content_data cd
JOIN datatypes_fields df ON cd.datatype_id = df.datatype_id
WHERE cd.content_data_id = 100
  AND df.field_id = 50;
```

### Pitfall 2: Datatype Changes Breaking Content

**Problem:** Deleting a field that has content_fields values.

**Solution:** CASCADE deletion automatically removes content_fields:
```sql
DELETE FROM fields WHERE field_id = 50;
-- All content_fields with field_id = 50 are automatically deleted
```

**Better:** Mark field as deprecated instead of deleting:
```sql
ALTER TABLE fields ADD COLUMN deprecated BOOLEAN DEFAULT 0;
UPDATE fields SET deprecated = 1 WHERE field_id = 50;
```

### Pitfall 3: Route Confusion

**Problem:** Querying content without filtering by route_id.

**Result:** Mix of content from different sites.

**Fix:** Always include route_id in WHERE clause:
```sql
SELECT * FROM content_data
WHERE route_id = ?  -- REQUIRED
  AND ...;
```

### Pitfall 4: Field Value Type Mismatch

**Problem:** Storing non-text in field_value TEXT column.

**Example:**
```sql
-- Bad: Storing JSON directly
INSERT INTO content_fields (field_value)
VALUES ('{"key": "value"}');  -- Stored as string, needs parsing
```

**Solution:** Accept that field_value is TEXT, parse in application:
```go
var jsonData map[string]interface{}
err := json.Unmarshal([]byte(field.FieldValue), &jsonData)
```

---

## Related Documentation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md` - Content tree implementation
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/DATABASE_LAYER.md` - Database abstraction
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TUI_ARCHITECTURE.md` - How content is managed in TUI

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_TABLES.md` - Creating schema migrations
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_FEATURES.md` - Adding new functionality

**Database:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQL_DIRECTORY.md` - Working with SQL schemas
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md` - Database Go code
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQLC.md` - sqlc reference

**Domain:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/ROUTES_AND_SITES.md` - Multi-site architecture
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/DATATYPES_AND_FIELDS.md` - Schema system
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/CONTENT_TREES.md` - Tree operations

---

## Quick Reference

### Core Tables

**Schema (Meta-Layer):**
- `datatypes` - Content type definitions
- `fields` - Property definitions
- `datatypes_fields` - Links datatypes to fields

**Data (Content Layer):**
- `routes` - Site containers
- `content_data` - Content instances (tree nodes)
- `content_fields` - Field values

**Admin (Parallel):**
- `admin_routes`, `admin_datatypes`, `admin_fields`
- `admin_datatypes_fields`, `admin_content_data`, `admin_content_fields`

### Key Relationships

```
routes (1) ──< (many) content_data
datatypes (1) ──< (many) content_data
datatypes (many) >──< (many) fields  [via datatypes_fields]
content_data (1) ──< (many) content_fields
fields (1) ──< (many) content_fields
```

### Schema Paths
- Routes: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/6_routes/`
- Datatypes: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/7_datatypes/`
- Fields: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/8_fields/`
- Content Data: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/16_content_data/`
- Content Fields: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/17_content_fields/`
- Junction: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/20_datatypes_fields/`

### Go Type Definitions
- `internal/db/datatype.go` - Datatypes struct
- `internal/db/field.go` - Fields struct
- `internal/db/content_data.go` - ContentData struct
- `internal/db/content_field.go` - ContentFields struct

### Common Queries
- Get content with fields: Join content_data → content_fields → fields
- Get datatype schema: Join datatypes → datatypes_fields → fields
- List content by type: Filter content_data by datatype_id
- Search by field value: Join to content_fields, filter on field_value
