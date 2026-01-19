# DATATYPES_AND_FIELDS.md

Comprehensive documentation on ModulaCMS's dynamic content schema system using datatypes and fields.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/DATATYPES_AND_FIELDS.md`
**Related Schemas:**
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/7_datatypes/`
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/8_fields/`
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/20_datatypes_fields/`

---

## Overview

ModulaCMS uses a **dynamic schema system** where content structure is defined in the database rather than hardcoded in the application. This system revolves around two core concepts:

**Datatypes** = Content type definitions (like "Page", "Post", "Product")
**Fields** = Property definitions (like "Title", "Body", "Price")

**Why This Matters:**
- Define new content types without code changes
- Modify schemas at runtime via TUI
- Share field definitions across datatypes
- Agencies can customize content models per client
- True headless CMS with flexible content modeling

**Key Insight:** Datatypes are like **classes** in OOP or **post types** in WordPress, while fields are like **properties** or **attributes**. The junction table `datatypes_fields` connects them in a many-to-many relationship.

---

## What Are Datatypes?

### Definition

A **datatype** is a content schema template that defines:
1. What kind of content it represents (Page, Post, Product, etc.)
2. Which fields it contains (via junction table)
3. Its classification (ROOT, ELEMENT, CONTAINER)
4. Optional hierarchy (parent datatypes)

**Analogy:**
- Like "post types" in WordPress (Page, Post, Custom Post Type)
- Like "content types" in Drupal
- Like "models" in Django
- Like "classes" in object-oriented programming

**Difference from Traditional CMSs:**
- Not hardcoded in application code
- Stored in database as data
- Can be created/modified at runtime
- No code deployment needed for new types

### Database Schema

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/7_datatypes/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS datatypes(
    datatype_id INTEGER PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES datatypes ON DELETE SET DEFAULT,
    label TEXT NOT NULL,                 -- "Page", "Post", "Product"
    type TEXT NOT NULL,                  -- "ROOT", "ELEMENT", "CONTAINER"
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
```

**Key Fields:**
- `datatype_id` - Unique identifier for this datatype
- `parent_id` - Optional parent datatype (for organizing datatypes hierarchically)
- `label` - Human-readable name displayed in TUI ("Page", "Blog Post", "Product")
- `type` - Classification: ROOT, ELEMENT, or CONTAINER
- `author_id` - User who created this datatype
- `history` - JSON audit trail of changes

### Datatype Classifications (type field)

**ROOT:**
- Top-level content types that can be root nodes in content trees
- Typically user-facing content types
- Examples: Page, Post, Product, Category, Event

**ELEMENT:**
- Leaf content types that typically don't have children
- Often components or building blocks
- Examples: Paragraph, Image, Button, Text Block, Video Embed

**CONTAINER:**
- Structural content types that contain other elements
- Used for layout and organization
- Examples: Section, Column, Row, Grid, Accordion

**Purpose of Classification:**
- Helps TUI present appropriate options when creating content
- Guides content structure decisions
- Documents intent (not enforced by database)

### Go Structure

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/datatype.go`

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

## What Are Fields?

### Definition

A **field** is a property definition that describes:
1. What data it stores (Title, Body, Price, etc.)
2. The input type (text, richtext, image, number, date)
3. Metadata and validation rules (max length, required, etc.)
4. Which datatypes use it (via junction table)

**Analogy:**
- Like "custom fields" in WordPress (ACF, CMB2)
- Like "columns" in database tables
- Like "properties" in classes
- Like "attributes" in HTML elements

**Key Concept:** Fields are **reusable** across multiple datatypes. A "Title" field can be used by both Page and Post datatypes.

### Database Schema

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/8_fields/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS fields(
    field_id INTEGER PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES datatypes ON DELETE SET DEFAULT,
    label TEXT DEFAULT 'unlabeled' NOT NULL,  -- "Title", "Body", "Price"
    data TEXT NOT NULL,                        -- JSON metadata
    type TEXT NOT NULL,                        -- "text", "richtext", "image"
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);
```

**Key Fields:**
- `field_id` - Unique identifier for this field
- `parent_id` - Legacy field (use datatypes_fields junction instead)
- `label` - Human-readable name ("Title", "Body", "Featured Image")
- `type` - Field type determining input widget and storage format
- `data` - JSON metadata with validation rules, defaults, help text
- `author_id` - User who created this field
- `history` - JSON audit trail

### Field Types

ModulaCMS supports various field types that determine the input widget and validation:

**Text Types:**
- `text` - Single-line text input (titles, names, slugs)
- `richtext` - HTML/Markdown editor (body content, descriptions)
- `textarea` - Multi-line plain text

**Numeric Types:**
- `number` - Integer or decimal numbers (prices, quantities, ratings)
- `integer` - Whole numbers only

**Media Types:**
- `image` - Image upload to S3 (featured images, logos, icons)
- `file` - Generic file upload (PDFs, documents)
- `video` - Video upload or embed URL

**Date/Time Types:**
- `date` - Date picker (publish dates, event dates)
- `datetime` - Date and time picker
- `time` - Time picker only

**Selection Types:**
- `boolean` - Checkbox/toggle (published, featured, active)
- `select` - Dropdown selection (category, status)
- `radio` - Radio button group
- `checkbox` - Multiple checkboxes

**Reference Types:**
- `reference` - Reference to another content item
- `user` - Reference to a user
- `media` - Reference to media library item

**Structured Types:**
- `json` - Structured JSON data
- `array` - List of values
- `object` - Key-value pairs

### Field Metadata (data column)

The `data` column stores JSON configuration for the field:

```json
{
  "placeholder": "Enter page title here",
  "maxLength": 200,
  "minLength": 5,
  "required": true,
  "default": "",
  "validation": "^[A-Za-z0-9\\s-]+$",
  "helpText": "Title appears in browser tab and search results",
  "widget": "input",
  "options": ["Draft", "Published", "Archived"],
  "multiple": false
}
```

**Common Metadata Fields:**
- `placeholder` - Placeholder text for input
- `maxLength` / `minLength` - Length constraints
- `required` - Whether field is mandatory
- `default` - Default value when creating new content
- `validation` - Regex pattern for validation
- `helpText` - Help text shown to user
- `widget` - Override default widget for type
- `options` - Available options for select/radio/checkbox
- `multiple` - Allow multiple selections

### Go Structure

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/field.go`

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

## The Junction Table: datatypes_fields

### Purpose

The `datatypes_fields` table creates a **many-to-many relationship** between datatypes and fields:
- One datatype can have many fields
- One field can belong to many datatypes (reusability!)
- Controls which fields appear in content editor for each datatype
- Maintains field ordering per datatype

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
- `id` - Unique identifier (order determines field sequence)
- `datatype_id` - Which datatype this association belongs to
- `field_id` - Which field is included

**Cascade Deletion:**
- Delete datatype → Removes all field associations
- Delete field → Removes all datatype associations
- Does NOT delete the field/datatype itself

### Field Ordering

The `id` (primary key) determines field order in the content editor:

```sql
-- Fields will appear in order: Title, Body, Featured Image
INSERT INTO datatypes_fields (datatype_id, field_id) VALUES
  (1, 10),  -- id=1: Title appears first
  (1, 20),  -- id=2: Body appears second
  (1, 30);  -- id=3: Featured Image appears third
```

**Query to Get Fields in Order:**

```sql
SELECT f.field_id, f.label, f.type, f.data
FROM fields f
JOIN datatypes_fields df ON f.field_id = df.field_id
WHERE df.datatype_id = ?
ORDER BY df.id;  -- Important: order by junction table id
```

---

## Creating a Complete Content Schema

Let's walk through creating a complete "Blog Post" datatype with fields.

### Step 1: Create the Datatype

**SQL:**

```sql
-- name: CreateDatatype :one
INSERT INTO datatypes (label, type, author_id)
VALUES ('Blog Post', 'ROOT', 1)
RETURNING *;
```

**Go Usage:**

```go
datatype, err := db.CreateDatatype(ctx, db.CreateDatatypeParams{
    Label:    "Blog Post",
    Type:     "ROOT",
    AuthorID: userID,
})

if err != nil {
    return fmt.Errorf("failed to create datatype: %w", err)
}

fmt.Printf("Created datatype_id: %d\n", datatype.DatatypeID)
```

**Result:** `datatype_id = 100` (example)

### Step 2: Create Fields

**SQL:**

```sql
-- Create Title field
INSERT INTO fields (label, type, data, author_id)
VALUES (
    'Post Title',
    'text',
    '{"required": true, "maxLength": 200, "placeholder": "Enter post title"}',
    1
);

-- Create Body field
INSERT INTO fields (label, type, data, author_id)
VALUES (
    'Post Body',
    'richtext',
    '{"required": true, "helpText": "Main post content"}',
    1
);

-- Create Excerpt field
INSERT INTO fields (label, type, data, author_id)
VALUES (
    'Excerpt',
    'textarea',
    '{"maxLength": 300, "placeholder": "Brief summary of post"}',
    1
);

-- Create Featured Image field
INSERT INTO fields (label, type, data, author_id)
VALUES (
    'Featured Image',
    'image',
    '{"required": false, "helpText": "Image shown in post listing"}',
    1
);

-- Create Publish Date field
INSERT INTO fields (label, type, data, author_id)
VALUES (
    'Publish Date',
    'datetime',
    '{"required": true, "default": "now"}',
    1
);

-- Create Published Status field
INSERT INTO fields (label, type, data, author_id)
VALUES (
    'Published',
    'boolean',
    '{"default": false, "helpText": "Make post visible on site"}',
    1
);
```

**Go Usage:**

```go
type FieldDefinition struct {
    Label string
    Type  string
    Data  string
}

fieldDefs := []FieldDefinition{
    {
        Label: "Post Title",
        Type:  "text",
        Data:  `{"required": true, "maxLength": 200}`,
    },
    {
        Label: "Post Body",
        Type:  "richtext",
        Data:  `{"required": true}`,
    },
    {
        Label: "Excerpt",
        Type:  "textarea",
        Data:  `{"maxLength": 300}`,
    },
    {
        Label: "Featured Image",
        Type:  "image",
        Data:  `{"required": false}`,
    },
    {
        Label: "Publish Date",
        Type:  "datetime",
        Data:  `{"required": true, "default": "now"}`,
    },
    {
        Label: "Published",
        Type:  "boolean",
        Data:  `{"default": false}`,
    },
}

fieldIDs := []int64{}
for _, fieldDef := range fieldDefs {
    field, err := db.CreateField(ctx, db.CreateFieldParams{
        Label:    fieldDef.Label,
        Type:     fieldDef.Type,
        Data:     fieldDef.Data,
        AuthorID: userID,
    })

    if err != nil {
        return fmt.Errorf("failed to create field %s: %w", fieldDef.Label, err)
    }

    fieldIDs = append(fieldIDs, field.FieldID)
    fmt.Printf("Created field: %s (id=%d)\n", field.Label, field.FieldID)
}
```

**Result:**
```
Created field: Post Title (id=50)
Created field: Post Body (id=51)
Created field: Excerpt (id=52)
Created field: Featured Image (id=53)
Created field: Publish Date (id=54)
Created field: Published (id=55)
```

### Step 3: Link Fields to Datatype

**SQL:**

```sql
-- Link fields to Blog Post datatype in desired order
INSERT INTO datatypes_fields (datatype_id, field_id) VALUES
  (100, 50),  -- Post Title
  (100, 51),  -- Post Body
  (100, 52),  -- Excerpt
  (100, 53),  -- Featured Image
  (100, 54),  -- Publish Date
  (100, 55);  -- Published
```

**Go Usage:**

```go
for _, fieldID := range fieldIDs {
    err := db.CreateDatatypeField(ctx, db.CreateDatatypeFieldParams{
        DatatypeID: datatype.DatatypeID,
        FieldID:    fieldID,
    })

    if err != nil {
        return fmt.Errorf("failed to link field %d: %w", fieldID, err)
    }
}

fmt.Println("All fields linked to Blog Post datatype")
```

### Step 4: Verify Schema

**Query to see complete schema:**

```sql
SELECT
    dt.label as datatype,
    dt.type as datatype_type,
    f.label as field_name,
    f.type as field_type,
    f.data as field_config,
    df.id as field_order
FROM datatypes dt
JOIN datatypes_fields df ON dt.datatype_id = df.datatype_id
JOIN fields f ON df.field_id = f.field_id
WHERE dt.datatype_id = 100
ORDER BY df.id;
```

**Result:**
```
datatype   | datatype_type | field_name      | field_type | field_order
-----------|---------------|-----------------|------------|------------
Blog Post  | ROOT          | Post Title      | text       | 1
Blog Post  | ROOT          | Post Body       | richtext   | 2
Blog Post  | ROOT          | Excerpt         | textarea   | 3
Blog Post  | ROOT          | Featured Image  | image      | 4
Blog Post  | ROOT          | Publish Date    | datetime   | 5
Blog Post  | ROOT          | Published       | boolean    | 6
```

---

## Reusing Fields Across Datatypes

### Why Reuse Fields?

**Benefits:**
- Consistency across content types (same "Title" field everywhere)
- Reduced duplication in database
- Shared validation rules
- Easier maintenance (change field once, affects all datatypes)

### Example: Sharing Title Field

```sql
-- Create Title field once
INSERT INTO fields (label, type, data, author_id)
VALUES ('Title', 'text', '{"required": true, "maxLength": 200}', 1);
-- Result: field_id = 1

-- Use in Page datatype
INSERT INTO datatypes_fields (datatype_id, field_id) VALUES (1, 1);

-- Use in Post datatype
INSERT INTO datatypes_fields (datatype_id, field_id) VALUES (2, 1);

-- Use in Product datatype
INSERT INTO datatypes_fields (datatype_id, field_id) VALUES (3, 1);
```

**Result:** Pages, Posts, and Products all use the same Title field definition with consistent validation.

### When to Reuse vs Create New

**Reuse when:**
- Same semantic meaning (all "Title" fields are titles)
- Same validation rules
- Same input type
- Want consistency across types

**Create new when:**
- Different validation needed (Page Title: max 100, Post Title: max 200)
- Different semantic meaning (Product Name vs Page Title)
- Different metadata requirements
- Want independence between types

---

## Working with Datatypes and Fields

### Querying All Datatypes

```sql
-- name: ListDatatypes :many
SELECT *
FROM datatypes
ORDER BY label;
```

**Go Usage:**

```go
datatypes, err := db.ListDatatypes(ctx)
if err != nil {
    return fmt.Errorf("failed to list datatypes: %w", err)
}

for _, dt := range datatypes {
    fmt.Printf("Datatype: %s (type=%s, id=%d)\n",
        dt.Label, dt.Type, dt.DatatypeID)
}
```

### Querying Fields for a Datatype

```sql
-- name: GetFieldsForDatatype :many
SELECT
    f.field_id,
    f.label,
    f.type,
    f.data
FROM fields f
JOIN datatypes_fields df ON f.field_id = df.field_id
WHERE df.datatype_id = ?
ORDER BY df.id;
```

**Go Usage:**

```go
fields, err := db.GetFieldsForDatatype(ctx, datatypeID)
if err != nil {
    return fmt.Errorf("failed to get fields: %w", err)
}

for _, field := range fields {
    fmt.Printf("Field: %s (type=%s)\n", field.Label, field.Type)

    // Parse metadata
    var metadata map[string]interface{}
    json.Unmarshal([]byte(field.Data), &metadata)

    if required, ok := metadata["required"].(bool); ok && required {
        fmt.Println("  - Required field")
    }
}
```

### Querying Which Datatypes Use a Field

```sql
-- name: GetDatatypesForField :many
SELECT
    dt.datatype_id,
    dt.label,
    dt.type
FROM datatypes dt
JOIN datatypes_fields df ON dt.datatype_id = df.datatype_id
WHERE df.field_id = ?
ORDER BY dt.label;
```

**Usage:** Understand impact of changing a field (which datatypes will be affected).

### Finding Content Using a Datatype

```sql
-- name: GetContentByDatatype :many
SELECT
    cd.content_data_id,
    cd.route_id,
    cd.date_created
FROM content_data cd
WHERE cd.datatype_id = ?
ORDER BY cd.date_created DESC;
```

**Usage:** List all instances of a specific content type (all blog posts, all products, etc.).

---

## Modifying Schemas at Runtime

### Adding a Field to Existing Datatype

**Scenario:** Add "SEO Meta Description" field to existing "Page" datatype.

```go
// 1. Create new field
seoField, err := db.CreateField(ctx, db.CreateFieldParams{
    Label: "SEO Meta Description",
    Type:  "textarea",
    Data:  `{"maxLength": 160, "helpText": "Description for search engines"}`,
    AuthorID: userID,
})

// 2. Get Page datatype ID
pageDatatype, err := db.GetDatatypeByLabel(ctx, "Page")

// 3. Link field to datatype
err = db.CreateDatatypeField(ctx, db.CreateDatatypeFieldParams{
    DatatypeID: pageDatatype.DatatypeID,
    FieldID:    seoField.FieldID,
})

fmt.Println("SEO field added to Page datatype")
```

**Result:**
- Existing Page content instances don't automatically have this field populated
- New Pages can have this field populated
- Field appears in content editor for Pages immediately
- No code deployment needed
- No database migration for existing content_data rows

### Removing a Field from Datatype

**Warning:** This deletes field values for all content instances!

```sql
-- Remove field from datatype (deletes junction record)
DELETE FROM datatypes_fields
WHERE datatype_id = ? AND field_id = ?;

-- If field is not used by any other datatype, optionally delete it
-- This CASCADE deletes all content_fields values!
DELETE FROM fields
WHERE field_id = ?;
```

**Safer Alternative:** Mark field as deprecated instead:

```sql
-- Add deprecated flag to fields table (migration)
ALTER TABLE fields ADD COLUMN deprecated BOOLEAN DEFAULT 0;

-- Mark field as deprecated
UPDATE fields SET deprecated = 1 WHERE field_id = ?;

-- Query non-deprecated fields only
SELECT f.*
FROM fields f
JOIN datatypes_fields df ON f.field_id = df.field_id
WHERE df.datatype_id = ?
  AND (f.deprecated IS NULL OR f.deprecated = 0)
ORDER BY df.id;
```

### Reordering Fields

**Problem:** Fields appear in wrong order in content editor.

**Solution:** Delete and re-insert junction records in desired order.

```sql
-- Delete existing associations
DELETE FROM datatypes_fields WHERE datatype_id = 100;

-- Re-insert in new order (id auto-increments, establishing new order)
INSERT INTO datatypes_fields (datatype_id, field_id) VALUES
  (100, 52),  -- Excerpt now first
  (100, 50),  -- Title second
  (100, 51),  -- Body third
  (100, 53),  -- Featured Image fourth
  (100, 54),  -- Publish Date fifth
  (100, 55);  -- Published sixth
```

---

## Dynamic Schema Benefits

### 1. No Code Deployments for Schema Changes

**Traditional CMS:**
```go
// Hardcoded schema in code
type Page struct {
    Title string
    Body  string
    Image string
}

// Adding field requires:
// 1. Code change
// 2. Database migration
// 3. Deploy new version
```

**ModulaCMS:**
```sql
-- Add field via database only
INSERT INTO fields (label, type) VALUES ('Author Bio', 'richtext');
INSERT INTO datatypes_fields (datatype_id, field_id) VALUES (1, 99);

-- Field immediately available in TUI
-- No code changes needed
-- No deployment needed
```

### 2. Per-Client Customization

**Agency Use Case:** Different clients need different content models.

```sql
-- Client A: Simple blog with basic fields
-- datatype_id=1: Blog Post (Title, Body, Date)

-- Client B: Complex blog with SEO and social
-- datatype_id=2: Blog Post (Title, Body, Date, Meta Description,
--                           OG Image, Twitter Card, Author, Categories)

-- Same codebase, different schemas per route
```

### 3. Rapid Prototyping

**Workflow:**
1. SSH into server via TUI
2. Create new datatype: "Product Review"
3. Add fields: Title, Rating (1-5), Pros, Cons, Verdict
4. Start creating content immediately
5. Iterate on schema based on feedback
6. No development cycle needed

### 4. Content Migration Support

**Scenario:** Migrate from WordPress to ModulaCMS.

```go
// Create WordPress-compatible datatypes dynamically
func MigrateWordPressSchema(wpPosts []WordPressPost) error {
    // Analyze WordPress fields
    fieldNames := extractFieldNames(wpPosts)

    // Create datatype
    dt, _ := db.CreateDatatype(ctx, CreateDatatypeParams{
        Label: "WP Post",
        Type:  "ROOT",
    })

    // Create fields dynamically
    for _, fieldName := range fieldNames {
        field, _ := db.CreateField(ctx, CreateFieldParams{
            Label: fieldName,
            Type:  inferFieldType(fieldName),
        })

        db.CreateDatatypeField(ctx, CreateDatatypeFieldParams{
            DatatypeID: dt.DatatypeID,
            FieldID:    field.FieldID,
        })
    }

    return nil
}
```

---

## Comparison to Other Systems

### vs WordPress

**WordPress:**
- Hardcoded post types in PHP: `register_post_type()`
- Custom fields via plugins (ACF, CMB2, Pods)
- Schema changes require code + plugin updates
- Metadata stored in wp_postmeta (EAV pattern)

**ModulaCMS:**
- Datatypes in database
- Built-in flexible field system
- Schema changes via database/TUI
- Field values in content_fields table
- Similar flexibility, cleaner architecture

### vs Drupal

**Drupal:**
- Content types via admin UI (stored in config)
- Field API with field types
- Export configuration to YAML files
- Complex field API with formatters, widgets

**ModulaCMS:**
- Similar concept (datatypes = content types)
- Simpler field system
- Direct database access
- No configuration export/import (database is source of truth)

### vs Contentful/Strapi

**Headless CMSs:**
- Content models via admin UI
- Field definitions with validation
- API-first design
- GraphQL/REST APIs

**ModulaCMS:**
- Similar flexibility
- TUI instead of web admin (SSH-based)
- REST APIs
- Designed for agencies hosting multiple clients

---

## Common Patterns

### Pattern 1: Standard Page Datatype

```sql
-- Create Page datatype with common fields
INSERT INTO datatypes (label, type) VALUES ('Page', 'ROOT');
INSERT INTO fields (label, type, data) VALUES
  ('Title', 'text', '{"required": true, "maxLength": 200}'),
  ('Slug', 'text', '{"required": true, "pattern": "^[a-z0-9-]+$"}'),
  ('Body', 'richtext', '{"required": true}'),
  ('Featured Image', 'image', '{}'),
  ('Meta Description', 'textarea', '{"maxLength": 160}'),
  ('Published', 'boolean', '{"default": true}');

-- Link all fields to Page
INSERT INTO datatypes_fields (datatype_id, field_id)
SELECT 1, field_id FROM fields WHERE label IN (
  'Title', 'Slug', 'Body', 'Featured Image', 'Meta Description', 'Published'
);
```

### Pattern 2: E-commerce Product Datatype

```sql
-- Create Product datatype
INSERT INTO datatypes (label, type) VALUES ('Product', 'ROOT');

-- Create product-specific fields
INSERT INTO fields (label, type, data) VALUES
  ('Product Name', 'text', '{"required": true}'),
  ('SKU', 'text', '{"required": true, "pattern": "^[A-Z0-9-]+$"}'),
  ('Price', 'number', '{"required": true, "min": 0, "step": 0.01}'),
  ('Sale Price', 'number', '{"min": 0, "step": 0.01}'),
  ('Description', 'richtext', '{"required": true}'),
  ('Product Images', 'image', '{"multiple": true, "max": 10}'),
  ('In Stock', 'boolean', '{"default": true}'),
  ('Stock Quantity', 'integer', '{"min": 0}'),
  ('Category', 'select', '{"options": ["Electronics", "Clothing", "Books"]}');
```

### Pattern 3: Hierarchical Datatypes

```sql
-- Create parent datatype
INSERT INTO datatypes (label, type) VALUES ('Content Block', 'CONTAINER');
-- parent_id = 1

-- Create child datatypes
INSERT INTO datatypes (parent_id, label, type) VALUES
  (1, 'Text Block', 'ELEMENT'),
  (1, 'Image Block', 'ELEMENT'),
  (1, 'Video Block', 'ELEMENT');

-- Organize datatypes logically in TUI
```

---

## Common Pitfalls and Debugging

### Pitfall 1: Missing Junction Record

**Problem:** Field exists, datatype exists, but field doesn't appear for content.

**Symptom:**
```sql
-- Field exists
SELECT * FROM fields WHERE field_id = 50;  -- ✓ Returns row

-- Datatype exists
SELECT * FROM datatypes WHERE datatype_id = 1;  -- ✓ Returns row

-- But no junction record!
SELECT * FROM datatypes_fields
WHERE datatype_id = 1 AND field_id = 50;  -- ✗ Empty
```

**Fix:**
```sql
INSERT INTO datatypes_fields (datatype_id, field_id) VALUES (1, 50);
```

### Pitfall 2: Orphaned Field Values

**Problem:** content_fields rows reference deleted fields.

**Symptom:**
```sql
-- Field value exists
SELECT * FROM content_fields WHERE field_id = 99;

-- But field is deleted
SELECT * FROM fields WHERE field_id = 99;  -- Empty!
```

**Prevention:** CASCADE deletion should handle this:
```sql
REFERENCES fields ON DELETE CASCADE
```

**Manual Fix:**
```sql
DELETE FROM content_fields WHERE field_id NOT IN (SELECT field_id FROM fields);
```

### Pitfall 3: Field Type Mismatch

**Problem:** Storing wrong type of data in field_value.

**Example:**
```sql
-- Number field
INSERT INTO fields (label, type) VALUES ('Price', 'number');

-- But storing non-numeric value
INSERT INTO content_fields (field_id, field_value) VALUES (10, 'expensive');
```

**Solution:** Validate in application before inserting:

```go
func ValidateFieldValue(field db.Fields, value string) error {
    switch field.Type {
    case "number", "integer":
        if _, err := strconv.ParseFloat(value, 64); err != nil {
            return fmt.Errorf("invalid number: %s", value)
        }
    case "boolean":
        if value != "true" && value != "false" {
            return fmt.Errorf("invalid boolean: %s", value)
        }
    case "datetime", "date":
        if _, err := time.Parse(time.RFC3339, value); err != nil {
            return fmt.Errorf("invalid date: %s", value)
        }
    }
    return nil
}
```

### Pitfall 4: Field Order Confusion

**Problem:** Fields appear in unexpected order in content editor.

**Debug:**
```sql
-- Check current order
SELECT df.id, f.label
FROM fields f
JOIN datatypes_fields df ON f.field_id = df.field_id
WHERE df.datatype_id = 1
ORDER BY df.id;
```

**Fix:** See "Reordering Fields" section above.

### Pitfall 5: Deleting Used Datatype

**Problem:** Attempting to delete datatype that has content instances.

**Symptom:**
```sql
DELETE FROM datatypes WHERE datatype_id = 1;
-- Error: foreign key constraint (content_data references datatypes)
```

**Check first:**
```sql
-- Count content instances
SELECT COUNT(*) FROM content_data WHERE datatype_id = 1;
```

**Options:**
1. Delete all content first (dangerous!)
2. Change content to different datatype (migration)
3. Don't delete, mark as deprecated instead

### Debugging Tips

**1. Verify Complete Schema:**

```sql
-- Full schema overview
SELECT
    dt.label as datatype,
    COUNT(DISTINCT df.field_id) as field_count,
    COUNT(DISTINCT cd.content_data_id) as content_count
FROM datatypes dt
LEFT JOIN datatypes_fields df ON dt.datatype_id = df.datatype_id
LEFT JOIN content_data cd ON dt.datatype_id = cd.datatype_id
GROUP BY dt.datatype_id, dt.label
ORDER BY dt.label;
```

**2. Find Unused Fields:**

```sql
-- Fields not associated with any datatype
SELECT f.field_id, f.label
FROM fields f
LEFT JOIN datatypes_fields df ON f.field_id = df.field_id
WHERE df.id IS NULL;
```

**3. Find Unused Datatypes:**

```sql
-- Datatypes with no content instances
SELECT dt.datatype_id, dt.label
FROM datatypes dt
LEFT JOIN content_data cd ON dt.datatype_id = cd.datatype_id
WHERE cd.content_data_id IS NULL;
```

**4. Validate Field Metadata JSON:**

```go
func ValidateFieldMetadata(field db.Fields) error {
    var metadata map[string]interface{}
    if err := json.Unmarshal([]byte(field.Data), &metadata); err != nil {
        return fmt.Errorf("invalid JSON in field %d: %w", field.FieldID, err)
    }

    // Validate metadata structure
    if required, ok := metadata["required"]; ok {
        if _, isBool := required.(bool); !isBool {
            return fmt.Errorf("'required' must be boolean")
        }
    }

    return nil
}
```

---

## Related Documentation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` - Complete domain model context
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/DATABASE_LAYER.md` - Database abstraction
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md` - How content instances form trees

**Domain:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/ROUTES_AND_SITES.md` - Routes that contain content
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/CONTENT_TREES.md` - Working with content trees

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_FEATURES.md` - Feature development workflow
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_TABLES.md` - Schema migrations

**Database:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQL_DIRECTORY.md` - SQL organization
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQLC.md` - Type-safe queries
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md` - Database operations

---

## Quick Reference

### Core Tables

**Schema Definition:**
- `datatypes` - Content type definitions (Page, Post, Product)
- `fields` - Property definitions (Title, Body, Price)
- `datatypes_fields` - Many-to-many junction table

**Content Storage:**
- `content_data` - Instances of datatypes (tree nodes)
- `content_fields` - Field values for content instances

### Relationships

```
datatypes (1) ──< (many) content_data (instances)
datatypes (many) >──< (many) fields [via datatypes_fields]
content_data (1) ──< (many) content_fields (values)
fields (1) ──< (many) content_fields (values)
```

### Key Concepts

- **Datatype** = Content schema template (like a class or post type)
- **Field** = Property definition (like an attribute or column)
- **Junction** = Links datatypes to their fields (many-to-many)
- **Reusability** = Fields can be shared across datatypes
- **Dynamic Schema** = Create/modify content types without code changes
- **Field Order** = Determined by datatypes_fields.id sequence

### Common Field Types

- `text` - Single-line text
- `richtext` - HTML editor
- `number` - Numeric values
- `boolean` - True/false
- `image` - Image upload
- `datetime` - Date picker
- `select` - Dropdown

### Schema Paths

- Datatypes: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/7_datatypes/`
- Fields: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/8_fields/`
- Junction: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/20_datatypes_fields/`

### Go Type Definitions

- Datatypes: `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/datatype.go`
- Fields: `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/field.go`

### Essential Queries

```sql
-- Get fields for datatype
SELECT f.* FROM fields f
JOIN datatypes_fields df ON f.field_id = df.field_id
WHERE df.datatype_id = ?
ORDER BY df.id;

-- Get datatypes using field
SELECT dt.* FROM datatypes dt
JOIN datatypes_fields df ON dt.datatype_id = df.datatype_id
WHERE df.field_id = ?;

-- Get content instances of datatype
SELECT * FROM content_data WHERE datatype_id = ?;
```

### Key Operations

```go
// Create datatype
dt, err := db.CreateDatatype(ctx, CreateDatatypeParams{
    Label: "Blog Post",
    Type: "ROOT",
})

// Create field
field, err := db.CreateField(ctx, CreateFieldParams{
    Label: "Title",
    Type: "text",
    Data: `{"required": true}`,
})

// Link field to datatype
err := db.CreateDatatypeField(ctx, CreateDatatypeFieldParams{
    DatatypeID: dt.DatatypeID,
    FieldID: field.FieldID,
})
```
