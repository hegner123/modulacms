# PLUGIN_PACKAGE.md

**Purpose:** Documentation for the ModulaCMS Lua plugin system in `internal/plugin/`

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/plugin/`

**Last Updated:** 2026-01-12

---

## Overview

The plugin package provides an extensibility system for ModulaCMS using embedded Lua scripting via gopher-lua. The plugin system enables:

1. **Output Adapters** - Transform CMS output to match other API formats (e.g., WordPress JSON)
2. **Import Adapters** - Migrate content from other systems into ModulaCMS
3. **Custom Business Logic** - Content transformations, validation, workflow automation without recompiling

**Current Status:** The plugin system is in early development with core structure implemented but API finalization pending. The foundation exists for plugin registration and Lua state management, with planned expansion for full adapter functionality.

---

## Architecture Philosophy

### Why Plugins?

ModulaCMS is designed for agencies that need to:
- Migrate clients from legacy systems (WordPress, Drupal, Contentful)
- Match existing API contracts without frontend rewrites
- Implement client-specific business logic without forking the core
- Extend functionality without recompiling the binary

### Why Lua?

Lua was chosen for several reasons:
1. **Embeddable** - gopher-lua provides pure Go implementation (no C dependencies)
2. **Lightweight** - Small runtime footprint, fast execution
3. **Accessible** - Widely adopted scripting language with low learning curve
4. **Sandboxable** - Can restrict access to dangerous operations
5. **Simple Syntax** - Easy for non-programmers to write simple transformations

### Design Goals

- **Zero Downtime** - Load/unload plugins without restarting the server
- **Isolation** - Plugins cannot corrupt core data or crash the system
- **Performance** - Minimal overhead for plugin execution
- **Flexibility** - Support different plugin types (output, input, logic)
- **Interoperability** - Pass data as bytes, potentially support non-Lua plugins

---

## Current Implementation

### File Structure

```
internal/plugin/
├── plugin.go              # Core plugin system implementation
└── blog/                  # Example blog plugin
    ├── config.json        # Plugin metadata and table definitions
    ├── index.html         # Plugin frontend template
    ├── style.css          # Plugin styles
    └── assets/
        └── js/            # Plugin JavaScript
            ├── index.js
            ├── archive.js
            ├── comments.js
            ├── post.js
            └── taxonomy.js
```

### Core Structures

#### Plugin Struct

**Location:** `internal/plugin/plugin.go:7-10`

```go
type Plugin struct {
	Path string
	Name string
}
```

Represents a plugin with:
- **Path** - Filesystem path to plugin directory or Lua file
- **Name** - Human-readable plugin identifier

#### PluginRegister Struct

**Location:** `internal/plugin/plugin.go:11-14`

```go
type PluginRegister struct {
	State    *lua.LState
	Register []Plugin
}
```

Central registry managing all plugins:
- **State** - Shared Lua VM state (from gopher-lua)
- **Register** - Collection of loaded plugins

### Initialization

#### NewPluginRegister()

**Location:** `internal/plugin/plugin.go:16-23`

```go
func NewPluginRegister() *PluginRegister {
	o := lua.Options{}
	state := lua.NewState(o)
	return &PluginRegister{
		State: state,
	}
}
```

Creates a new plugin registry with fresh Lua state:
1. Initialize empty Lua options
2. Create new Lua VM state with gopher-lua
3. Return registry with initialized state and empty plugin list

**Usage:**
```go
pluginRegistry := plugin.NewPluginRegister()
// State is ready for plugin loading
```

### Registration

#### RegisterPlugin()

**Location:** `internal/plugin/plugin.go:25-30`

```go
func (Pr *PluginRegister) RegisterPlugin(p Plugin) {
	pr := *Pr
	pr.Register = append(pr.Register, p)
}
```

Adds a plugin to the registry:
1. Takes Plugin struct with Path and Name
2. Appends to Register slice
3. Plugin becomes available for execution

**Current Limitation:** Does not actually load or execute the plugin, just registers metadata.

**Usage:**
```go
pluginRegistry.RegisterPlugin(plugin.Plugin{
	Path: "/path/to/plugin.lua",
	Name: "WordPress Output Adapter",
})
```

---

## Blog Plugin Example

The blog plugin (`internal/plugin/blog/`) demonstrates the intended plugin structure.

### Configuration (config.json)

**Location:** `internal/plugin/blog/config.json`

Plugins can declare database requirements:

```json
{
    "id": 1,
    "name": "Blog Plugin",
    "version": "1.1.0",
    "addsTables": true,
    "tables": [
        {
            "id": 0,
            "name": "blogs",
            "columns": [
                {
                    "id": 0,
                    "name": "id",
                    "type": "INTEGER",
                    "nullable": false,
                    "defaultValue": ""
                },
                {
                    "id": 1,
                    "name": "name",
                    "type": "TEXT",
                    "nullable": false,
                    "defaultValue": ""
                }
                // ... more columns
            ],
            "primaryKey": 0,
            "foreignKeys": []
        }
    ]
}
```

**Structure:**
- **id** - Unique plugin identifier
- **name** - Display name
- **version** - Semantic version
- **addsTables** - Boolean indicating if plugin creates tables
- **tables** - Array of table definitions with columns, keys, and relationships

**Purpose:** Allows plugins to extend the database schema for custom functionality without modifying core schema files.

### Frontend Assets

The blog plugin includes frontend components:
- **index.html** - Template for blog views
- **style.css** - Plugin-specific styles
- **assets/js/** - Client-side JavaScript for interactivity

**Note:** These are currently placeholder files for the planned implementation.

---

## Planned Features

### Output Adapters

**Purpose:** Transform ModulaCMS output into other API formats.

**Use Case:** Client has existing Next.js site consuming WordPress JSON API. They want to migrate to ModulaCMS without rewriting frontend.

**Example Lua Implementation:**

```lua
-- wordpress_adapter.lua
function transform_to_wordpress(content)
    -- Transform ModulaCMS content structure
    -- into WordPress API response format
    return {
        id = content.content_data_id,
        title = {
            rendered = content.fields.title
        },
        content = {
            rendered = content.fields.body
        },
        excerpt = {
            rendered = content.fields.excerpt or ""
        },
        date = content.created_at,
        modified = content.updated_at,
        slug = content.slug,
        status = content.status or "publish",
        type = "post",
        link = build_permalink(content),
        author = content.author_id,
        featured_media = content.fields.featured_image or 0,
        -- ... match full WordPress schema
    }
end
```

**How It Works:**
1. HTTP handler receives request for content
2. ModulaCMS loads content with fields
3. Plugin transforms ModulaCMS structure to WordPress format
4. Response sent as WordPress JSON
5. Frontend receives expected format, no changes needed

### Import Adapters

**Purpose:** Migrate content from other systems into ModulaCMS.

**Use Case:** Client migrating from WordPress with 10 years of content.

**Example Lua Implementation:**

```lua
-- wordpress_importer.lua
function import_wordpress_post(wp_post)
    -- Transform WordPress post structure
    -- into ModulaCMS datatypes and fields

    local datatype_id = get_datatype_id("Post")
    local route_id = get_route_id("Main Site")

    local content_data_id = create_content({
        route_id = route_id,
        datatype_id = datatype_id,
        slug = wp_post.post_name,
        created_at = wp_post.post_date,
        updated_at = wp_post.post_modified
    })

    -- Map WordPress fields to ModulaCMS fields
    set_field_value(content_data_id, "Title", wp_post.post_title)
    set_field_value(content_data_id, "Body", wp_post.post_content)
    set_field_value(content_data_id, "Excerpt", wp_post.post_excerpt)

    -- Handle featured image
    if wp_post.featured_image then
        local media_id = import_media(wp_post.featured_image)
        set_field_value(content_data_id, "Featured Image", media_id)
    end

    -- Import taxonomies
    for _, category in ipairs(wp_post.categories) do
        add_taxonomy(content_data_id, "category", category)
    end

    return content_data_id
end

-- Batch import function
function import_wordpress_export(xml_path)
    local posts = parse_wordpress_xml(xml_path)
    local imported = 0

    for _, post in ipairs(posts) do
        local success, content_id = pcall(import_wordpress_post, post)
        if success then
            imported = imported + 1
            print("Imported: " .. post.post_title)
        else
            print("Failed: " .. post.post_title .. " - " .. content_id)
        end
    end

    return imported
end
```

**How It Works:**
1. Parse WordPress export file (XML/SQL)
2. Iterate through posts/pages/custom types
3. Transform each item to ModulaCMS format
4. Create content with mapped fields
5. Handle relationships and taxonomies
6. Report success/failure for each item

### Custom Business Logic

**Purpose:** Implement client-specific rules and transformations.

**Use Cases:**
- Custom field validation
- Content workflow automation
- Dynamic field transformations
- Access control rules
- Content relationships
- Search indexing

**Example Lua Implementation:**

```lua
-- content_validation.lua
function validate_blog_post(content)
    local errors = {}

    -- Title is required
    if not content.fields.title or content.fields.title == "" then
        table.insert(errors, "Title is required")
    end

    -- Title must be under 100 characters
    if content.fields.title and #content.fields.title > 100 then
        table.insert(errors, "Title must be under 100 characters")
    end

    -- Body must be at least 50 characters
    if not content.fields.body or #content.fields.body < 50 then
        table.insert(errors, "Body must be at least 50 characters")
    end

    -- Featured image required for published posts
    if content.status == "published" and not content.fields.featured_image then
        table.insert(errors, "Published posts require a featured image")
    end

    return #errors == 0, errors
end

-- workflow_automation.lua
function on_publish(content)
    -- Send notification when content is published
    notify_author(content.author_id, "Your post '" .. content.fields.title .. "' was published")

    -- Update related content
    update_related_posts_modified_date(content.content_data_id)

    -- Trigger search reindex
    reindex_content(content.content_data_id)

    -- Log audit trail
    log_audit("PUBLISH", content.content_data_id, get_current_user_id())
end

-- field_transformation.lua
function auto_generate_excerpt(content)
    if not content.fields.excerpt or content.fields.excerpt == "" then
        -- Generate excerpt from body (first 150 characters)
        local body = strip_html(content.fields.body)
        content.fields.excerpt = string.sub(body, 1, 150) .. "..."
    end
    return content
end
```

**How It Works:**
1. Register plugin hooks for events (before_save, after_publish, etc.)
2. ModulaCMS calls hook when event occurs
3. Plugin executes custom logic
4. Returns success/failure or modified content
5. ModulaCMS proceeds based on plugin response

---

## Planned Plugin API

The following API functions would be provided to Lua plugins (not yet implemented):

### Content Functions

```lua
-- Content CRUD
create_content(data) -> content_id
get_content(content_id) -> content
update_content(content_id, data) -> success
delete_content(content_id) -> success

-- Field Operations
set_field_value(content_id, field_name, value) -> success
get_field_value(content_id, field_name) -> value
get_all_fields(content_id) -> fields_map

-- Tree Operations
get_children(content_id) -> []content_id
get_parent(content_id) -> content_id
move_content(content_id, new_parent_id, position) -> success
```

### Datatype Functions

```lua
get_datatype_id(name) -> datatype_id
get_datatype(datatype_id) -> datatype
list_datatypes() -> []datatype
get_fields_for_datatype(datatype_id) -> []field
```

### Route Functions

```lua
get_route_id(name) -> route_id
get_route(route_id) -> route
list_routes() -> []route
```

### Database Functions

```lua
-- Raw queries (use carefully)
query(sql, params) -> results
exec(sql, params) -> affected_rows

-- Transaction support
begin_transaction() -> tx
commit_transaction(tx) -> success
rollback_transaction(tx) -> success
```

### Media Functions

```lua
import_media(url_or_path) -> media_id
get_media(media_id) -> media
upload_to_s3(file_data, filename) -> url
```

### Utility Functions

```lua
-- Logging
log(level, message)  -- levels: debug, info, warn, error

-- HTTP requests
http_get(url, headers) -> response
http_post(url, body, headers) -> response

-- JSON handling
json_encode(table) -> string
json_decode(string) -> table

-- String utilities
strip_html(html) -> text
markdown_to_html(markdown) -> html
slugify(text) -> slug
```

---

## Plugin Development Workflow

### Creating a New Plugin

**Step 1: Create Plugin Directory**

```bash
mkdir -p internal/plugin/my_plugin
cd internal/plugin/my_plugin
```

**Step 2: Create config.json (if adding tables)**

```json
{
    "id": 100,
    "name": "My Plugin",
    "version": "1.0.0",
    "addsTables": true,
    "tables": [
        {
            "id": 0,
            "name": "my_plugin_data",
            "columns": [
                {
                    "id": 0,
                    "name": "id",
                    "type": "INTEGER",
                    "nullable": false,
                    "defaultValue": ""
                }
            ],
            "primaryKey": 0,
            "foreignKeys": []
        }
    ]
}
```

**Step 3: Create Lua Script**

```bash
touch internal/plugin/my_plugin/main.lua
```

```lua
-- main.lua
-- Plugin initialization
function init()
    log("info", "My Plugin initialized")
    return true
end

-- Register hooks
register_hook("before_save", function(content)
    -- Your logic here
    return content
end)

-- Export functions
return {
    init = init,
    version = "1.0.0"
}
```

**Step 4: Register Plugin (in application code)**

```go
import "modulacms/internal/plugin"

func initPlugins() {
    registry := plugin.NewPluginRegister()

    registry.RegisterPlugin(plugin.Plugin{
        Path: "/path/to/internal/plugin/my_plugin/main.lua",
        Name: "My Plugin",
    })

    // TODO: Load and execute plugin
}
```

**Step 5: Test Plugin**

Currently, plugin testing requires manual integration since the execution framework is not yet complete.

---

## Security Considerations

### Sandboxing

**Planned Restrictions:**
- Disable dangerous Lua standard libraries (io, os, package)
- Limit file system access
- Restrict network access to approved domains
- Set execution timeouts to prevent infinite loops
- Memory limits per plugin

**Implementation Strategy:**
```go
func createSecureLuaState() *lua.LState {
    state := lua.NewState()

    // Remove dangerous libraries
    state.SetGlobal("os", lua.LNil)
    state.SetGlobal("io", lua.LNil)
    state.SetGlobal("package", lua.LNil)
    state.SetGlobal("dofile", lua.LNil)
    state.SetGlobal("loadfile", lua.LNil)

    return state
}
```

### Input Validation

All data passed to plugins must be validated:
- Sanitize content before passing to Lua
- Validate plugin output before using in queries
- Prevent SQL injection via parameter binding
- Escape HTML output to prevent XSS

### Code Review

Plugin code should be:
- Reviewed before deployment
- Tested in staging environment
- Monitored for performance issues
- Audited for security vulnerabilities

### Permission System (Planned)

Plugins should declare required permissions:
```json
{
    "permissions": [
        "read_content",
        "write_content",
        "execute_queries",
        "http_requests"
    ]
}
```

Administrators approve permissions before plugin activation.

---

## Integration Points

### HTTP Handlers

Plugins could modify HTTP responses:

```go
func contentHandler(w http.ResponseWriter, r *http.Request) {
    // Load content from database
    content := loadContent(contentID)

    // Check for output adapter plugin
    if plugin := findOutputAdapter(r.Header.Get("X-Output-Format")); plugin != nil {
        content = plugin.Transform(content)
    }

    json.NewEncoder(w).Encode(content)
}
```

### Content Lifecycle Hooks

Plugins could hook into content events:

```go
func saveContent(content *model.ContentData) error {
    // Before save hooks
    for _, plugin := range registry.GetPlugins("before_save") {
        content = plugin.Execute(content)
    }

    // Save to database
    err := db.SaveContent(content)
    if err != nil {
        return err
    }

    // After save hooks
    for _, plugin := range registry.GetPlugins("after_save") {
        plugin.Execute(content)
    }

    return nil
}
```

### Database Migrations

Plugins with `addsTables: true` could auto-generate migrations:

```go
func loadPlugin(pluginPath string) error {
    config := loadPluginConfig(pluginPath + "/config.json")

    if config.AddsTables {
        for _, table := range config.Tables {
            // Generate CREATE TABLE statement
            migration := generateMigration(table)

            // Apply to all database drivers
            applyMigration(migration)
        }
    }

    return nil
}
```

---

## Performance Considerations

### Lua VM Pooling

For high-traffic scenarios, maintain a pool of Lua states:

```go
type PluginPool struct {
    states chan *lua.LState
    size   int
}

func NewPluginPool(size int) *PluginPool {
    pool := &PluginPool{
        states: make(chan *lua.LState, size),
        size:   size,
    }

    // Pre-create states
    for i := 0; i < size; i++ {
        state := createSecureLuaState()
        pool.states <- state
    }

    return pool
}

func (p *PluginPool) Execute(script string) (lua.LValue, error) {
    // Get state from pool
    state := <-p.states
    defer func() { p.states <- state }()

    // Execute script
    err := state.DoString(script)
    return state.Get(-1), err
}
```

### Caching

Cache plugin outputs when possible:
- Cache transformed content for output adapters
- Cache validation results
- Invalidate on content changes

### Monitoring

Track plugin performance:
- Execution time per plugin
- Memory usage
- Error rates
- Success/failure counts

Log slow plugins for optimization.

---

## Future Enhancements

### Plugin Marketplace

- Central repository for community plugins
- Version management
- Dependency resolution
- Automatic updates

### Hot Reload

- Reload plugins without restarting server
- Update plugin code on-the-fly
- Rollback on errors

### Multi-Language Support

The docs mention "passing []bytes back and forth" and "any programming language or micro-service". Future versions could support:
- Python plugins via subprocess
- JavaScript plugins via goja
- WebAssembly plugins
- HTTP-based plugin services

### Visual Plugin Editor

TUI or web UI for:
- Writing simple Lua scripts
- Testing transformations
- Browsing plugin marketplace
- Managing plugin configuration

---

## Troubleshooting

### Plugin Not Loading

**Symptom:** RegisterPlugin() called but plugin not available

**Cause:** Current implementation only registers metadata, doesn't load Lua

**Solution:** Wait for API completion, or implement custom loading:
```go
registry := plugin.NewPluginRegister()
registry.RegisterPlugin(plugin.Plugin{Path: "path/to/plugin.lua", Name: "My Plugin"})

// Manually load until built-in support exists
err := registry.State.DoFile("path/to/plugin.lua")
if err != nil {
    log.Fatal(err)
}
```

### Lua Syntax Errors

**Symptom:** Plugin fails to execute

**Cause:** Syntax errors in Lua script

**Solution:** Test Lua scripts separately with `lua` command or online REPL before integrating

### Performance Issues

**Symptom:** Slow response times when plugins enabled

**Cause:** Plugin executing expensive operations

**Solution:**
- Profile plugin execution time
- Move expensive operations to background jobs
- Cache plugin results
- Optimize Lua code

---

## Related Documentation

- **[ADDING_FEATURES.md](../workflows/ADDING_FEATURES.md)** - Adding plugin hooks to core
- **[DATABASE_LAYER.md](../architecture/DATABASE_LAYER.md)** - Database access from plugins
- **[PLUGIN_ARCHITECTURE.md](../architecture/PLUGIN_ARCHITECTURE.md)** - High-level plugin design (Phase 5)
- **[MODEL_PACKAGE.md](MODEL_PACKAGE.md)** - Data structures plugins work with

---

## Quick Reference

### Key Files
- `internal/plugin/plugin.go` - Core implementation
- `internal/plugin/blog/` - Example plugin
- `docs/plugins.md` - Plugin operation notes

### Current Status
- ✅ Basic structure implemented
- ✅ gopher-lua integration
- ✅ Plugin registration
- ⏳ API design in progress
- ⏳ Execution framework pending
- ⏳ Security sandbox pending

### Dependencies
- **gopher-lua** (github.com/yuin/gopher-lua v1.1.1) - Pure Go Lua VM

### Key Concepts
- **Output Adapter** - Transform CMS output to match external APIs
- **Import Adapter** - Migrate data from other systems
- **Custom Logic** - Client-specific rules and workflows
- **Sandbox** - Restricted execution environment for safety

### Next Steps
1. Finalize plugin API design
2. Implement plugin loading and execution
3. Add security sandbox
4. Create example plugins (WordPress adapter, validation)
5. Document plugin development workflow
6. Build plugin marketplace infrastructure

---

**Last Updated:** 2026-01-12
