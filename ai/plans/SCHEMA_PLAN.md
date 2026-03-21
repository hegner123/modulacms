# Schema-from-Filesystem Implementation Plan

Define content schemas via a `schema/` directory using TypeScript files that compile to JSON. Modula reads the JSON and syncs datatypes/fields into the database.

**Prior art:** A working prototype exists at `/Users/home/Documents/Code/Go_dev/cms/modulacms.com/modula/schema/` with 81 schema files covering 6 root types. This plan incorporates its proven conventions and extends them with Go-side sync, validation support, and a publishable package.

## Architecture

The entire schema compiler is embedded in the Modula binary. No Node, Bun, or external JS toolchain required.

```
schema/*.ts                     User-authored TypeScript files
  |
  v  (esbuild Go API: strip types, bundle imports, output CommonJS)
  |
  v  (goja: evaluate JS, extract default export)
  |
  v  Go structs ([]FlatDatatype)
  |
  ├──> schema/.build/           Optional JSON output (schema.json + per-root splits)
  |
  v  (diff + sync)
internal/schema/                Parse, diff, sync engine
  |
  v  (reuses existing definitions system)
internal/definitions/           SchemaDefinition -> Install/Reinstall
  |
  v
internal/service/schema.go     Database CRUD via SchemaService
```

### Embedded Compiler: esbuild + goja

The binary embeds two Go libraries for TypeScript compilation:

- **esbuild** (Go API) — transforms `.ts` files to JavaScript. Strips TypeScript syntax, resolves imports, bundles each schema file into a self-contained CommonJS module. esbuild is already a project dependency (admin panel block editor bundling).
- **goja** (pure Go JS runtime) — evaluates the bundled JavaScript and extracts the `export default` value. The schema files are pure data (object literals + `f.*` helper calls), so goja's ES5.1+ coverage is sufficient after esbuild transforms modern syntax.

For each `.ts` file in `schema/`:
1. esbuild bundles it (resolving `field-types.ts` imports, stripping `@modulacms/schema` type imports)
2. goja evaluates the bundle and reads the default export
3. The result is marshaled into `FlatDatatype` Go structs

This means `modula schema build` does not shell out to any external process.

### Future: typescript-go (AST-based extraction)

Microsoft's [typescript-go](https://github.com/microsoft/typescript-go) is the official TypeScript compiler ported to Go. When its API stabilizes, it could replace both esbuild and goja:

```
schema/*.ts → typescript-go (parse AST + type check) → walk AST → Go structs
```

Benefits: full TypeScript type checking inside the binary, no JS evaluation at all (pure static AST extraction), and one dependency instead of two. The schema files are simple enough (object literals + function calls) that AST-based extraction is feasible.

The user-facing experience is identical — same schema files, same CLI commands, same output. The swap is internal to `internal/schema/compiler.go`.

Not pursued initially because typescript-go is new and its Go API may shift. esbuild + goja work today and are battle-tested.

## Phase 1: `@modulacms/schema` TypeScript Package (optional)

**Location:** `sdks/typescript/schema/`
**npm name:** `@modulacms/schema`

Optional package for IDE type checking. Not required to use the schema system — the Modula binary compiles `.ts` files natively. Users who want TypeScript editor support (red squiggles on invalid fields, autocomplete on `SchemaNode` properties) install this package. Users who don't can write schema files without it — the embedded compiler handles everything.

### Type Definitions

The package ships structural interfaces, validation types, and a generic field helper. It does NOT define `FieldType` — that is generated into the user's `schema/` directory from their database (see Phase 3).

**`src/types.ts`** — validation and config types

```typescript
// Datatype types matching Modula's type column
// Reserved types start with _; user-defined types are open strings
export type DatatypeType =
  | "_root" | "_global" | "_reference" | "_nested_root" | "_collection"
  | string; // user-defined: "section", "layout", "content", "navigation", etc.

// Validation rules matching internal/db/types/types_field_config.go
export type RuleOp =
  | "required" | "contains" | "starts_with" | "ends_with"
  | "equals" | "length" | "count" | "range" | "item_count" | "one_of";

export type Cmp = "eq" | "neq" | "gt" | "gte" | "lt" | "lte";
export type CharClass = "uppercase" | "lowercase" | "digits" | "symbols" | "spaces";

export type ValidationRule = {
  op: RuleOp;
  value?: string;
  values?: string[];
  class?: CharClass;
  cmp?: Cmp;
  n?: number;
  negate?: boolean;
  message?: string;
};

export type RuleEntry = {
  rule?: ValidationRule;
  group?: RuleGroup;
};

export type RuleGroup = {
  all_of?: RuleEntry[];
  any_of?: RuleEntry[];
};

export type ValidationConfig = {
  rules?: RuleEntry[];
};

export type UIConfig = {
  widget?: string;
  placeholder?: string;
  help_text?: string;
  hidden?: boolean;
};

export type RichTextConfig = {
  toolbar?: string[];
};

export type RelationConfig = {
  target_datatype: string; // name, not ID (resolved at sync time)
  cardinality: "one" | "many";
  max_depth?: number;
};
```

**`src/schema.ts`** — interfaces and generic field helper

```typescript
import type { DatatypeType, ValidationConfig, UIConfig } from "./types.js";

export interface Field {
  name: string;
  label: string;
  type: string; // matches field_types.type column (name + label pairs in DB)
  data?: Record<string, unknown>;
  validation?: ValidationConfig;
  ui?: UIConfig;
  translatable?: boolean;
  roles?: string[];
}

// SchemaNode is what each schema file exports.
// Parent-child relationships are defined by directory nesting.
export interface SchemaNode {
  name: string;
  label: string;
  type: DatatypeType;
  fields: Field[];
}

// Generic field helper for custom field types
type FieldOpts = Partial<Pick<Field, "data" | "validation" | "ui" | "translatable" | "roles">>;

export function field(type: string, name: string, label: string, opts?: FieldOpts): Field {
  return { name, label, type, ...opts };
}
```

**Build:** tsup, ESM+CJS dual. Ships `dist/` only.

### Generated Schema Boilerplate (`modula schema init`)

`modula schema init` reads the `field_types` table from the connected database and generates a `schema/field-types.ts` file with:
- A `FieldType` union from all registered field types
- An `f` helper object with one method per field type (providing autocomplete and type safety)

**Generated `schema/field-types.ts`:**

The generator reads every row from `field_types` (columns: `type`, `label`) and produces:
- A `FieldType` string union from all `type` values
- An `f.*` helper per row, using `type` as the method name and `label` as the default label

```typescript
// Auto-generated by: modula schema init
// Regenerate after adding field types: modula schema generate-types
import { type Field } from "@modulacms/schema";

// Generated from field_types table (type + label columns)
export type FieldType =
  | "_id" | "_title" | "boolean" | "date" | "datetime"
  | "email" | "json" | "media" | "number" | "richtext"
  | "select" | "slug" | "text" | "textarea" | "url" | "plugin";

type FieldOpts = Partial<Pick<Field, "data" | "validation" | "ui" | "translatable" | "roles">>;

// One helper per field_types row: f.{type}(name, label, opts?)
export const f = {
  _id:      (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "_id", ...opts }),
  _title:   (name = "title", label = "Title", opts?: FieldOpts): Field =>
              ({ name, label, type: "_title", ...opts }),
  boolean:  (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "boolean", ...opts }),
  date:     (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "date", ...opts }),
  datetime: (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "datetime", ...opts }),
  email:    (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "email", ...opts }),
  json:     (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "json", ...opts }),
  media:    (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "media", ...opts }),
  number:   (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "number", ...opts }),
  richtext: (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "richtext", ...opts }),
  select:   (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "select", ...opts }),
  slug:     (name = "slug", label = "Slug", opts?: FieldOpts): Field =>
              ({ name, label, type: "slug", ...opts }),
  text:     (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "text", ...opts }),
  textarea: (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "textarea", ...opts }),
  url:      (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "url", ...opts }),
  plugin:   (name: string, label: string, opts?: FieldOpts): Field =>
              ({ name, label, type: "plugin", ...opts }),
};
```

Every helper has the same signature: `f.{type}(name, label, opts?)`. The generator produces them mechanically from the two columns in `field_types`. When a user registers a custom field type (e.g., `type: "color_picker"`, `label: "Color Picker"`), running `modula schema generate-types` adds `f.color_picker()` to the file.

Users can also edit this file directly — it's their code, not a locked artifact.

### User Experience

```typescript
// schema/page/_root.ts
import { type SchemaNode } from "@modulacms/schema";
import { f } from "../field-types.js";

const schema: SchemaNode = {
  name: "page",
  label: "Page",
  type: "_root",
  fields: [
    f.title(),
    f.textarea("description", "Description"),
    f.media("featured_image", "Featured Image"),
    f.text("meta_title", "Meta Title"),
    f.textarea("meta_description", "Meta Description"),
    f.media("og_image", "OG Image"),
  ],
};

export default schema;
```

```typescript
// schema/page/hero.ts (leaf child of page)
import { type SchemaNode } from "@modulacms/schema";
import { f } from "../field-types.js";

const schema: SchemaNode = {
  name: "hero",
  label: "Hero",
  type: "content",
  fields: [
    f.text("header", "Header"),
    f.textarea("subtitle", "Subtitle"),
    f.media("image", "Image"),
    f.url("link_url", "Link URL"),
  ],
};

export default schema;
```

```typescript
// schema/page/row/_fields.ts (intermediate container with children)
import { type SchemaNode } from "@modulacms/schema";
import { f } from "../../field-types.js";

const schema: SchemaNode = {
  name: "row",
  label: "Row",
  type: "layout",
  fields: [
    f.boolean("full_width", "Full Width"),
  ],
};

export default schema;
```

```typescript
// schema/menu/_global.ts (global content type)
import { type SchemaNode } from "@modulacms/schema";
import { f } from "../field-types.js";

const schema: SchemaNode = {
  name: "menu",
  label: "Menu",
  type: "_global",
  fields: [
    f.title(),
    f.text("position", "Position"),
  ],
};

export default schema;
```

```typescript
// schema/menu-reference.ts (reference type, leaf file at root)
import { type SchemaNode } from "@modulacms/schema";
import { f } from "../field-types.js";

const schema: SchemaNode = {
  name: "menu_reference",
  label: "Menu Reference",
  type: "_reference",
  fields: [
    f.id("menu_id", "Menu"),
  ],
};

export default schema;
```

```typescript
// Field with validation
f.text("email", "Email", {
  validation: { rules: [{ rule: { op: "required" } }] },
  ui: { placeholder: "user@example.com" },
})
```

## Phase 2: JSON Intermediate Format

The compiler outputs two kinds of files to `schema/.build/`:
- `schema.json`: flat array of ALL datatypes (full schema)
- `{root_name}.json`: per-root files containing only that root and its descendants

Both are flat arrays (no envelope object). Go reads the full `schema.json`.

```json
[
  {
    "name": "page",
    "label": "Page",
    "type": "_root",
    "parent": null,
    "fields": [
      {
        "name": "title",
        "label": "Title",
        "type": "_title"
      },
      {
        "name": "description",
        "label": "Description",
        "type": "textarea"
      }
    ]
  },
  {
    "name": "hero",
    "label": "Hero",
    "type": "content",
    "parent": "page",
    "fields": [
      {
        "name": "header",
        "label": "Header",
        "type": "text"
      }
    ]
  }
]
```

The `parent` field is a string name reference (not an ID), resolved at sync time. Parent-child relationships are inferred from directory structure by the compiler.

Fields include `data`, `validation`, `ui`, `translatable`, and `roles` only when set (omitted when absent to keep output clean).

Per-root splits enable selective imports: a frontend that only renders `page` content can import `page.json` without loading the full schema.

## Phase 3: Embedded Compiler

**Location:** `internal/schema/compiler.go`

The compiler is embedded in the Modula binary. `modula schema build` walks the `schema/` directory and compiles `.ts` files to JSON using esbuild + goja — no external toolchain required.

```go
// Compile walks the schema directory, evaluates each .ts file, and returns
// the flat list of datatypes with parent relationships inferred from directory nesting.
func Compile(schemaDir string) ([]FlatDatatype, error)
```

Per-file compilation:
1. **esbuild** bundles the `.ts` file (resolves imports like `field-types.ts`, strips type-only imports from `@modulacms/schema`, outputs CommonJS)
2. **goja** evaluates the bundle and reads the `module.exports.default` value
3. The result is marshaled into a `FlatDatatype` Go struct

The compiler skips `field-types.ts`, `build.ts`, `types.ts`, and non-`.ts` files.

### Directory Structure Encoding

```
schema/
  page/                    # directory = "page" has children
    _root.ts               # page's own fields + type (_root)
    hero.ts                # leaf child of page
    card.ts                # leaf child of page
    row/                   # "row" has children
      _fields.ts           # row's own fields + type
      columns.ts           # leaf child of row
    grid/
      _fields.ts           # grid's own fields
      area.ts              # leaf child of grid
  post/
    _root.ts               # another root datatype
    post-content.ts        # leaf child of post
  menu/
    _global.ts             # global datatype (menus, headers, footers)
    menu-link.ts           # leaf child
    menu-list/             # nested container
      _fields.ts
      menu-list-link.ts
  menu-reference.ts        # _reference type with _id field (leaf, no directory needed)
```

### File Convention

| File | Meaning |
|------|---------|
| `_root.ts` | Root datatype definition (`type: "_root"`) |
| `_global.ts` | Global datatype definition (`type: "_global"`, for site-wide content like menus, headers, footers) |
| `_fields.ts` | Intermediate container definition (has children, any type) |
| `*.ts` | Leaf datatype (no children) |

`_root.ts`, `_global.ts`, and `_fields.ts` are all directory self-descriptors. The distinction between `_root.ts` and `_global.ts` is semantic: `_global.ts` maps to `type: "_global"` and signals that this content tree is site-wide. The compiler enforces that `_global.ts` files set `type: "_global"`.

Reference types (`_reference`) are just regular leaf or container datatypes with `type: "_reference"`. They have no special filesystem convention because they have no semantic requirements beyond being a named type with an `_id` field. Organize them however makes sense for the project.

Rules:
- A directory MUST contain exactly one of `_root.ts`, `_global.ts`, or `_fields.ts` for its own definition
- Files in a directory are children of that directory's datatype
- Files at the root of `schema/` with no directory are leaf root datatypes
- Leaf promotion: `hero.ts` -> `hero/_fields.ts` + add children (git tracks the rename)
- Datatype names must be globally unique (compiler validates this)
- `types.ts`, `build.ts`, `field-types.ts`, and non-`.ts` files are skipped by the compiler

### Compiler Validation (build time)

The compiler enforces structural rules that don't require database access:

**Singleton field types:** A datatype cannot have more than one field of type `_id` or `_title`. These are system fields with special meaning in Modula's content model. The compiler rejects schemas that violate this.

**Type consistency:** `_global.ts` files must set `type: "_global"`. `_root.ts` files must set `type: "_root"`. The compiler rejects mismatches.

**Name uniqueness:** No two datatypes can share the same `name` across the entire schema. The compiler collects all names and reports collisions with file paths for both sides.

**Reference targets:** If a field uses `RelationConfig`, the `target_datatype` must name a datatype that exists in the schema. The compiler validates all references resolve.

### Sync Validation (Go side, requires database)

The Go sync engine validates rules that depend on database state:

**Field type existence:** Every field's `type` value must match a `type` column in the `field_types` table (which stores only `type` and `label`, no validation logic). The sync reports unresolvable field types with the datatype and field name for context.

**Workflow for custom field types:**
1. Register the custom field type in Modula (via API, TUI, or admin panel) — just a name and label
2. Run `modula schema generate-types` to regenerate `schema/field-types.ts`
3. Use the new `f.your_custom_type()` helper in schema files (with full LSP autocomplete)
4. Run `modula schema build && modula schema sync`

The `field_types` table is the source of truth for what field types exist. `field-types.ts` is a generated projection of that table into TypeScript.

### Compiler Walk Algorithm (proven in prototype)

Two-pass per directory:
1. **First pass:** find `_root.ts`, `_global.ts`, or `_fields.ts`, register it as this directory's datatype
2. **Second pass:** process sibling `.ts` files as leaf children, recurse into subdirectories

Descendant calculation uses index-based tracking to handle duplicate names across different roots (e.g., `code_block` under both `marketing_page` and `documentation`).

## Phase 4: Go-Side JSON Reader and Sync Engine

**New package:** `internal/schema/`

| File | Purpose |
|------|---------|
| `compiler.go` | Embedded TS compiler: esbuild transform + goja evaluation |
| `schema.go` | Go types (`FlatDatatype`, `Field`), `Read()` for JSON files |
| `convert.go` | Convert `[]FlatDatatype` to `definitions.SchemaDefinition` |
| `diff.go` | Diff current DB state against compiled schema |
| `sync.go` | Orchestrate: compile, diff, apply changes |
| `watcher.go` | File-polling watcher (follows `plugin/watcher.go` pattern) |
| `lockfile.go` | Lock file for name-to-ID mappings |
| `generate.go` | Generate `field-types.ts` from `field_types` table |

### Types (`schema.go`)

```go
package schema

type FlatDatatype struct {
    Name     string  `json:"name"`
    Label    string  `json:"label"`
    Type     string  `json:"type"`
    Parent   *string `json:"parent"`
    Fields   []Field `json:"fields"`
}

type Field struct {
    Name         string          `json:"name"`
    Label        string          `json:"label"`
    Type         string          `json:"type"`
    Data         json.RawMessage `json:"data,omitempty"`
    Validation   json.RawMessage `json:"validation,omitempty"`
    UI           json.RawMessage `json:"ui,omitempty"`
    Translatable bool            `json:"translatable,omitempty"`
    Roles        []string        `json:"roles,omitempty"`
}

// Read parses schema/.build/schema.json (flat array format).
func Read(path string) ([]FlatDatatype, error)
```

### Diff Engine (`diff.go`)

```go
type DiffAction int
const (
    ActionCreate DiffAction = iota
    ActionUpdate
    ActionDelete
    ActionNoop
)

type DatatypeDiff struct {
    Action  DiffAction
    Name    string
    Current *db.Datatypes
    Desired *FlatDatatype
    Fields  []FieldDiff
}

type FieldDiff struct {
    Action  DiffAction
    Name    string
    Current *db.Fields
    Desired *Field
}

type SyncPlan struct {
    Creates []DatatypeDiff
    Updates []DatatypeDiff
    Deletes []DatatypeDiff
    Noops   []DatatypeDiff
}

func Diff(ctx context.Context, schema []FlatDatatype, driver db.DbDriver) (*SyncPlan, error)
```

Matches datatypes and fields by `name`. This is the stable identifier across filesystem and database.

### Sync Orchestrator (`sync.go`)

```go
type SyncResult struct {
    DatatypesCreated int
    DatatypesUpdated int
    DatatypesDeleted int
    FieldsCreated    int
    FieldsUpdated    int
    FieldsDeleted    int
    Errors           []string
}

type SyncOptions struct {
    DryRun       bool
    AllowDeletes bool  // default false for safety
    AuthorID     types.UserID
}

func Sync(ctx context.Context, opts SyncOptions, schemaPath string, driver db.DbDriver) (*SyncResult, error)
```

Sync strategy:
1. Read `schema/.build/schema.json`
2. List all datatypes and fields from DB
3. Match by name, compute diffs
4. Create new datatypes (roots first, then children iteratively, same as `definitions.Install()`)
5. Update changed datatypes (label, type, parent changes)
6. Create/update/delete fields within each datatype
7. Delete removed datatypes only if `AllowDeletes` is true and no content exists
8. All mutations use audited context with source "schema-sync"

### Lock File (`lockfile.go`)

Stored at `schema/.build/schema.lock.json`. Tracks name-to-ID mappings:

```json
{
  "version": 1,
  "synced_at": "2026-03-20T10:00:00Z",
  "datatypes": {
    "page": "01JEXAMPLE000000000000001",
    "row": "01JEXAMPLE000000000000002"
  },
  "fields": {
    "page.title": "01JEXAMPLE000000000000010",
    "page.description": "01JEXAMPLE000000000000011"
  }
}
```

Committed to git for team consistency. Prevents orphaning, detects renames vs create+delete.

### Watcher (`watcher.go`)

Follows `internal/plugin/watcher.go` pattern: SHA256 polling, debounce, context cancellation.

Watches `.ts` files in `schema/` directly (not the JSON output). On change, recompiles and syncs.

## Phase 5: CLI Commands

**New cobra command:** `cmd/schema.go`

All commands are native — no external toolchain required.

```
modula schema init             # Scaffold schema/ directory with field-types.ts from DB
modula schema generate-types   # Regenerate field-types.ts from current field_types table
modula schema build            # Compile schema/*.ts to schema/.build/ JSON (embedded compiler)
modula schema sync             # Compile + diff + apply changes to DB
modula schema diff             # Compile + diff, show what would change (dry run)
modula schema watch            # Watch schema/*.ts, recompile + sync on change
modula schema validate         # Compile and validate without writing JSON or syncing
modula schema export           # Export current DB schema to schema/ directory as .ts files
```

Flags:
- `--allow-deletes` on `sync`: permit deleting removed datatypes/fields
- `--schema-dir` on all: override schema directory (default `./schema`)
- `--dry-run` on `sync`: same as `diff`

## Phase 6: Config Integration

```go
Schema_Enabled   bool   `json:"schema_enabled"`    // default false
Schema_Directory string  `json:"schema_directory"`  // default "schema"
Schema_Auto_Sync bool   `json:"schema_auto_sync"`  // watch and sync on server start
```

When `schema_auto_sync` is true and `schema/.build/schema.json` exists, the serve command starts the `SchemaWatcher` alongside existing server goroutines.

## Phase 7: Justfile and CI

### Justfile

All commands go through the binary — no external dependencies.

```just
schema action:
    #!/usr/bin/env bash
    set -euo pipefail
    case "{{action}}" in
        init)           ./modula-x86 schema init ;;
        generate-types) ./modula-x86 schema generate-types ;;
        build)          ./modula-x86 schema build ;;
        sync)           ./modula-x86 schema sync ;;
        diff)           ./modula-x86 schema diff ;;
        watch)          ./modula-x86 schema watch ;;
        validate)       ./modula-x86 schema validate ;;
        export)         ./modula-x86 schema export ;;
        *)              echo "Unknown action: {{action}}"; exit 1 ;;
    esac
```

### GitHub Actions

No workflow changes needed. The existing CI already covers the schema system:

- **`go.yml`** runs `go test -v ./...` which picks up `internal/schema/*_test.go` (compiler, diff, sync, generate tests). Since the compiler is embedded Go code using esbuild + goja, no Node/Bun setup step is needed.
- **`sdks.yml`** triggers on `sdks/typescript/**` changes, which covers the optional `@modulacms/schema` package if added later. Already runs pnpm install, build, typecheck, test.

## Phase 8: Integration

### Dual Content Schema

Filesystem schema applies to **public content only**. Admin tables remain internally managed. An `--admin` flag can be added later if needed.

### Conflict Resolution

1. Filesystem is source of truth during `modula schema sync`
2. Database changes not in conflict are preserved (content created under schema datatypes stays)
3. Deletes are opt-in (`--allow-deletes`)
4. Datatypes with content instances cannot be deleted even with `--allow-deletes`
5. Lock file tracks name-to-ID mappings to prevent orphaning

### API/TUI/Admin Panel

Changes made via API/TUI/admin panel are NOT written back to the filesystem. `modula schema export` is the reverse path for bootstrapping schema files from an existing database.

## Dependency Graph

```
Phase 1 (TS package, optional)    independent, can be done anytime
Phase 2 (JSON format)             defines the contract
Phase 3 (Embedded compiler)       depends on 2
Phase 4 (Go sync engine)          depends on 2
                |        |
                +--------+
                |
Phase 5 (CLI)                     depends on 3 + 4
Phase 6 (Config)                  depends on 5
Phase 7 (Justfile)                depends on 5 + 6
Phase 8 (Integration)             depends on all
```

**Parallel work:**
- Phase 3 (compiler) and Phase 4 (sync engine) can be built in parallel after Phase 2
- Phase 1 (TS package) is independent — it's optional IDE tooling, not on the critical path

## Testing

| Area | Test Type | Location |
|------|-----------|----------|
| Embedded compiler | Go unit (compile fixture `.ts` files, verify output) | `internal/schema/compiler_test.go` |
| JSON read | Go unit (parse fixture JSON) | `internal/schema/schema_test.go` |
| Convert | Go unit (convert to SchemaDefinition) | `internal/schema/convert_test.go` |
| Diff | Go unit (diff against mock DB state) | `internal/schema/diff_test.go` |
| Sync | Go integration (SQLite in testdb/) | `internal/schema/sync_test.go` |
| Generate types | Go unit (generate field-types.ts from mock data) | `internal/schema/generate_test.go` |
| CLI commands | Go integration (`exec.Command`) | `cmd/schema_test.go` |
| `@modulacms/schema` (optional) | Vitest unit | `sdks/typescript/schema/test/` |

## Risks

1. **Name collisions:** Two datatypes with the same name in different directories. Compiler validates unique names globally.
2. **goja limitations:** goja supports ES5.1 with some ES6. esbuild must downlevel all modern syntax before goja evaluates it. Schema files that use features esbuild can't transform (unlikely) would fail at compile time with a clear error.
3. **Content orphaning:** Sync refuses to delete datatypes with content. `--force-delete` flag for override.
4. **Ordering:** Fields are ordered by their position in the array. Directories are processed in sorted order for deterministic output.

## Key Files to Reference

| File | Why |
|------|-----|
| `internal/definitions/definition.go` | `SchemaDefinition`, `DatatypeDef`, `FieldDef` types the new package must produce |
| `internal/definitions/install.go` | `Install()`/`Reinstall()` 3-phase sync pattern to follow |
| `internal/db/types/types_field_config.go` | All validation rule types the TS package must mirror |
| `internal/plugin/watcher.go` | File-polling watcher pattern to follow |
| `sdks/typescript/types/src/enums.ts` | TS type conventions to follow |
| Prior art: `~/Documents/Code/Go_dev/cms/modulacms.com/modula/schema/` | Working prototype with 81 schema files, proven compiler and conventions |
