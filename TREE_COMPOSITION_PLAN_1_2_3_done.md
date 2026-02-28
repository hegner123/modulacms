# Tree Composition, Reserved Types, and Tree Refactor Plan

## Overview

Three interconnected changes:

1. **Reserved Types Library** — Formal system for engine-interpreted datatype types (`_root`, `_reference`, `_nested_root`, future reserved types), with validation preventing user creation of reserved names
2. **Tree Package Refactor** — Extract shared tree-building core from `internal/tree` so `model`, `admin`, and `cli` all use the same algorithm
3. **Tree Composition Layer** — Resolve `_reference` nodes by fetching and attaching external content trees via the existing tree-building pipeline, producing a unified tree for rendering

## Current State

- `_root` migration from `ROOT` is complete across all files
- Datatype `Type` field is a plain `string` — no validation, no typed enum
- `internal/model` builds trees for API output (JSON-friendly types, slice-based children)
- `internal/tree` builds trees for TUI (raw DB types, sibling pointers, expand/collapse)
- Admin panel does O(n^2) brute-force tree rendering in templ templates, no tree package used
- No validation on datatype type at creation or update — any string accepted

## Design Decisions

1. **Reference field identification:** New `FieldType` enum value `content_tree_ref`. The engine finds reference fields by field type, not label convention.

2. **Multiple references per node:** A single `_reference` node can have an array of `content_tree_ref` field values. Each value resolves to a subtree, and all resolved subtrees are attached as children of the reference node in field order.

3. **Reference target scope:** Any `content_data_id`. The referenced ID IS the root of the subtree — the engine fetches that row and all its descendants via parent_id, regardless of the node's original type. The node's type is set to `_nested_root` before tree assembly. `BuildTree` identifies tree roots by checking for `_root` or `_nested_root`.

4. **Broken references:** Graceful degradation. When a reference cannot be resolved, the reference node's child becomes a synthetic `_system_log` node with structured fields describing the failure. Client sites handle `_system_log` nodes however they choose.

5. **Composition reuses existing tree build.** No separate composition pipeline. For each referenced content ID, the engine runs the same full `BuildTree` / `BuildAdminTree` path (fetch content data + datatypes + content fields + field definitions, assemble tree). This gives full datatype, field, content_data, and content_field merging with zero new data-loading code.

6. **Concurrent composition.** Reference resolution uses `errgroup` with bounded concurrency. When a tree has multiple `_reference` nodes, all are resolved concurrently. Nested references within subtrees spawn additional goroutines recursively, bounded by `MaxDepth` and `MaxConcurrency`. Each request gets its own cache and visited set.

7. **Two core build functions, not generics.** `core.BuildTree` takes `db.ContentData`/`db.Datatypes`/`db.ContentFields`/`db.Fields`. `core.BuildAdminTree` takes `db.AdminContentData`/`db.AdminDatatypes`/`db.AdminContentFields`/`db.AdminFields`. Copy-paste with admin prefixes, same as the existing `model.BuildTree`/`model.BuildAdminTree` pattern.

8. **Greenfield deployment.** No existing deployments, no legacy behavior to preserve. No backward compatibility concerns for SDKs or API consumers.

## Dependencies Between Phases

```
Phase 1: Reserved Types Library (no dependencies)
Phase 2: Tree Package Refactor (no dependencies, can parallel with Phase 1)
Phase 3: Tree Composition Layer (depends on Phase 1 + Phase 2)
Phase 4: Admin Panel Integration (depends on Phase 2)
```

Phases 1 and 2 can be developed in parallel.

---

## Phase 1: Reserved Types Library

**Package:** `internal/db/types/reserved.go`

### 1.1 Define the DatatypeType

Follow the same pattern as `FieldType` — a named string type with const values, `Validate()`, `Scan()`, `Value()`, `MarshalJSON()`, `UnmarshalJSON()`.

```go
// DatatypeType represents the type classification of a datatype.
// Values prefixed with underscore are engine-reserved and trigger
// built-in behavior. All other values are user-defined pass-through.
type DatatypeType string

const (
    DatatypeTypeRoot       DatatypeType = "_root"
    DatatypeTypeReference  DatatypeType = "_reference"
    DatatypeTypeNestedRoot DatatypeType = "_nested_root"
    DatatypeTypeSystemLog  DatatypeType = "_system_log"
)
```

### 1.2 Reserved Type Registry

```go
// reservedTypes maps each reserved type to a description of its engine behavior.
var reservedTypes = map[DatatypeType]string{
    DatatypeTypeRoot:       "Tree entry point, one per route",
    DatatypeTypeReference:  "Triggers tree composition — resolves content_tree_ref field values, attaches referenced trees as children",
    DatatypeTypeNestedRoot: "Root of a composed subtree, assigned by the engine during tree composition",
    DatatypeTypeSystemLog:  "Synthetic node injected when a reference cannot be resolved",
}

// IsReserved returns true if the type is engine-reserved.
func (t DatatypeType) IsReserved() bool {
    _, ok := reservedTypes[t]
    return ok
}

// IsReservedPrefix returns true if the string starts with underscore,
// which is the reserved namespace. Used to reject user-created types
// that start with underscore even if not currently in the registry.
func IsReservedPrefix(t string) bool {
    return len(t) > 0 && t[0] == '_'
}

// IsRootType returns true if the type identifies a tree root node.
// Used by the tree-building algorithm to find root nodes.
func (t DatatypeType) IsRootType() bool {
    return t == DatatypeTypeRoot || t == DatatypeTypeNestedRoot
}

// ReservedTypes returns a copy of the registry for documentation/UI purposes.
func ReservedTypes() map[DatatypeType]string { ... }
```

### 1.3 Add FieldTypeContentTreeRef

**File:** `internal/db/types/types_enums.go`

Add to the `FieldType` const block:

```go
FieldTypeContentTreeRef FieldType = "content_tree_ref"
```

Update the `Validate()` switch case and error message to include it.

This field type:
- Stores a content data ULID as its value
- Can appear multiple times on a single `_reference` datatype node
- Each value triggers a tree fetch during composition

### 1.4 Validation Behavior

The `DatatypeType` is **not** a closed enum like `FieldType`. It must accept arbitrary user strings (`page`, `hero`, `blog_post`) while reserving the `_` prefix for engine types. Validation rules:

- Empty string: error
- Starts with `_` but not in registry: error ("reserved prefix, unknown type")
- Starts with `_` and in registry: valid reserved type
- Does not start with `_`: valid user type, no engine behavior

### 1.5 Add Validation to Creation and Update Paths

**Files to modify (creation):**

- `internal/router/datatypes.go` -> `apiCreateDatatype()`: reject types starting with `_`, also add basic non-empty validation for label and type
- `internal/admin/handlers/datatypes.go` -> `DatatypeCreateHandler()`: same validation

**Files to modify (update):**

- `internal/router/datatypes.go` -> `apiUpdateDatatype()`: reject type changes to `_`-prefixed values
- `internal/admin/handlers/datatypes.go` -> `DatatypeUpdateHandler()`: same validation

**Files to modify (CLI):**

- `internal/cli/form_dialog.go` -> datatype creation/update forms: validate on submit — if value starts with `_`, show error "Type names starting with '_' are reserved for system use. Choose a different name." Do not modify the placeholder. Reserved types are created by the engine, not by users in forms.

**Validation rule:** User-created datatypes CANNOT use the `_` prefix. Reserved types are only created by the engine (bootstrap, import definitions). This is a creation-and-update-time guard, not a schema constraint.

### 1.6 Migration Scope

**Type field migration (plain `string` -> `DatatypeType`):**

Defer. Phase 1 delivers the registry and validation functions. The `Type` field stays `string` across DB structs. Validation is called explicitly at API boundaries. The full type migration can follow as a dedicated refactor once the composition layer is proven.

### 1.7 Deliverables

- [ ] `internal/db/types/reserved.go` — `DatatypeType`, const block, registry, `IsReserved()`, `IsReservedPrefix()`, `IsRootType()`
- [ ] `internal/db/types/reserved_test.go` — validation tests
- [ ] `FieldTypeContentTreeRef` added to `types_enums.go` with updated `Validate()`
- [ ] Validation added to `apiCreateDatatype()` and `apiUpdateDatatype()` in `internal/router/datatypes.go`
- [ ] Validation added to `DatatypeCreateHandler()` and `DatatypeUpdateHandler()` in `internal/admin/handlers/datatypes.go`
- [ ] CLI form dialogs updated to reject `_`-prefixed type input with error message
- [ ] Existing `"_root"` string comparisons across codebase replaced with `types.DatatypeTypeRoot` constant

---

## Phase 2: Tree Package Refactor

### 2.1 Problem

Two independent tree implementations:

| | `internal/model` | `internal/tree` |
|---|---|---|
| Children | `[]*Node` slice | Sibling pointers (`FirstChild`/`NextSibling`/`PrevSibling`) |
| Node data | Mapped JSON types | Raw DB types |
| Construction | `BuildNodes()` single-pass | `LoadFromRows()` 4-phase with retry |
| Orphan handling | Log + error | Retry loop + circular detection |
| Reorder | `reorderChildren()` | `reorderByPointers()` |
| Consumers | Router API, transform layer | TUI |
| Needed by | Router API, transform, **admin panel**, **composition** | TUI |

The composition layer (Phase 3) needs a shared tree-building path. The admin panel needs proper tree building. Duplicating logic across packages is not viable.

### 2.2 Architecture

Extract a shared core and let each consumer wrap it:

```
internal/tree/core/          <-- NEW: shared algorithm + Node type
internal/tree/               <-- TUI layer (wraps core, adds expand/collapse/cursor)
internal/model/              <-- API layer (wraps core, maps to JSON-friendly types)
internal/admin/handlers/     <-- Uses core directly for server-rendered tree
```

### 2.3 Core Package Design (`internal/tree/core/`)

**`core.go` — Types:**

```go
package core

// Node is the shared tree node used by all consumers.
// It holds raw DB types and uses sibling pointers for O(1) manipulation.
type Node struct {
    ContentData   *db.ContentData    // Pointer: nil for synthetic nodes (e.g., _system_log without a real DB row)
    Datatype      db.Datatypes       // Value: always populated (zero value has empty strings, not nil)
    ContentFields []db.ContentFields // Field values for this node
    Fields        []db.Fields        // Field definitions for this node's content fields
    Parent        *Node
    FirstChild    *Node
    NextSibling   *Node
    PrevSibling   *Node
}

// Root is the top-level tree container with an index for O(1) lookups.
type Root struct {
    Node      *Node
    NodeIndex map[types.ContentID]*Node
}

// LoadStats tracks tree construction diagnostics.
type LoadStats struct {
    NodesCount      int
    OrphansResolved int
    RetryAttempts   int
    CircularRefs    []types.ContentID
    FinalOrphans    []types.ContentID
}
```

**`build.go` — Construction:**

The core build algorithm combines the best of both current implementations:

- 4-phase construction from `internal/tree` (create, assign hierarchy, resolve orphans, reorder)
- Circular reference detection from `internal/tree`
- Sibling pointer reordering from both (they're equivalent algorithms)
- Root identification via `types.DatatypeType.IsRootType()` — matches both `_root` and `_nested_root`

**Root identification algorithm:** A node is the tree root if its type satisfies `IsRootType()` (`_root` or `_nested_root`). If multiple nodes satisfy `IsRootType()`, the first one encountered becomes the root; subsequent ones are treated as regular nodes (logged as warning in LoadStats). If no node has a root type, fall back to the first node with no ParentID. If still no root found, return error.

Two build functions per type family (no generics):

```go
// BuildFromRows constructs a tree from TUI/admin query rows.
// Used by the TUI and admin panel where data comes from GetContentTreeByRoute.
func BuildFromRows(rows []db.GetContentTreeByRouteRow) (*Root, *LoadStats, error)

// BuildTree constructs a tree from the full set of content data, datatypes,
// content fields, and field definitions. Used by the API layer where data
// is fetched as parallel slices. Composition uses this for referenced subtrees.
// Contract: cd[i] pairs with dt[i], cf[i] pairs with df[i].
func BuildTree(cd []db.ContentData, dt []db.Datatypes, cf []db.ContentFields, df []db.Fields) (*Root, *LoadStats, error)

// BuildAdminTree constructs a tree from admin-prefixed DB types.
// Same algorithm as BuildTree, different input types. Copy-paste with admin prefixes,
// matching the existing model.BuildTree / model.BuildAdminTree pattern.
func BuildAdminTree(cd []db.AdminContentData, dt []db.AdminDatatypes, cf []db.AdminContentFields, df []db.AdminFields) (*Root, *LoadStats, error)
```

All three produce the same `*Root` with fully populated nodes. `BuildTree` and `BuildAdminTree` are the full pipelines that composition calls for each referenced tree.

**`traverse.go` — Shared traversal:**

```go
func CountVisible(root *Node, isExpanded func(*Node) bool) int
func FlattenVisible(root *Node, isExpanded func(*Node) bool) []*Node
func NodeAtIndex(root *Node, index int, isExpanded func(*Node) bool) *Node
func FindByContentID(root *Root, id types.ContentID) *Node
func IsDescendantOf(node, ancestor *Node) bool
```

Visibility traversal takes a callback `isExpanded func(*Node) bool` rather than reading an `Expand` field. The TUI layer provides its expand state. The admin/API layer provides `func(*Node) bool { return true }` (always expanded).

**`mutate.go` — Shared mutation:**

```go
func InsertNode(root *Root, parent, firstChild, prevSibling, nextSibling, node *Node)
func DeleteNode(root *Root, node *Node) bool
```

### 2.4 TUI Layer Changes (`internal/tree/`)

`internal/tree/` becomes a thin wrapper around `core`. The existing `tree.Node` and `tree.Root` types are preserved — the TUI layer keeps its current field names and API. No renames in `internal/cli/`.

The TUI layer:
- Wraps `core.BuildFromRows()` to produce its own `Root`/`Node` types
- Stores expand/collapse/indent state per-node
- Delegates traversal to `core` functions with its expand callback
- Delegates mutation to `core.InsertNode()` / `core.DeleteNode()`

### 2.5 Model Layer Changes (`internal/model/`)

`internal/model/` keeps its JSON-friendly types (`Root`, `Node`, `Datatype`, `Field`) for the transform/serialization layer. Its `BuildTree` / `BuildAdminTree` functions now delegate to the corresponding `core` function, then convert the `core.Root` to `model.Root` for JSON output.

The `model.Root` gains a `CoreRoot` field so composition can operate on the core tree after model build:

```go
type Root struct {
    Node     *Node       `json:"node"`
    CoreRoot *core.Root  `json:"-"` // Retained for composition pass, not serialized
}

// BuildTree builds a model tree for API responses.
// Delegates to core.BuildTree for assembly, then converts to model types.
// Retains the core.Root for optional composition pass.
func BuildTree(log Logger, cd []db.ContentData, dt []db.Datatypes,
               cf []db.ContentFields, df []db.Fields) (Root, error) {
    coreRoot, stats, err := core.BuildTree(cd, dt, cf, df)
    if err != nil {
        log.Warn("BuildTree: issues during assembly", err)
    }
    root := fromCoreRoot(coreRoot)
    root.CoreRoot = coreRoot
    return root, err
}

// BuildAdminTree builds a model tree from admin DB types.
// Same pattern: delegates to core.BuildAdminTree, converts, retains core.Root.
func BuildAdminTree(log Logger, cd []db.AdminContentData, dt []db.AdminDatatypes,
                    cf []db.AdminContentFields, df []db.AdminFields) (Root, error) {
    coreRoot, stats, err := core.BuildAdminTree(cd, dt, cf, df)
    if err != nil {
        log.Warn("BuildAdminTree: issues during assembly", err)
    }
    root := fromCoreRoot(coreRoot)
    root.CoreRoot = coreRoot
    return root, err
}

// fromCoreRoot walks the core tree and creates model.Node for each core.Node,
// mapping db types to JSON types via the existing Map*JSON functions.
func fromCoreRoot(coreRoot *core.Root) Root { ... }

// RebuildFromCore re-converts the core tree to model types after composition.
// Called after ComposeTrees mutates the core tree.
func (r *Root) RebuildFromCore() {
    if r.CoreRoot != nil {
        rebuilt := fromCoreRoot(r.CoreRoot)
        r.Node = rebuilt.Node
    }
}
```

### 2.6 Admin Panel Changes

The content handler (`internal/admin/handlers/content.go`) replaces its raw `[]db.GetContentTreeByRouteRow` pass-through with:

```go
tree, stats, err := core.BuildFromRows(rows)
// Pass tree to templ template, walk via FirstChild/NextSibling
```

The templ template walks the tree via pointers instead of O(n^2) scanning.

### 2.7 Deliverables

- [ ] `internal/tree/core/core.go` — shared `Node`, `Root`, `LoadStats`
- [ ] `internal/tree/core/build.go` — `BuildFromRows()`, `BuildTree()`, and `BuildAdminTree()` algorithms
- [ ] `internal/tree/core/traverse.go` — `CountVisible`, `FlattenVisible`, `NodeAtIndex`, `FindByContentID`, `IsDescendantOf`
- [ ] `internal/tree/core/mutate.go` — `InsertNode`, `DeleteNode`
- [ ] `internal/tree/core/*_test.go` — tests (migrate from existing `tree_test.go`)
- [ ] `internal/tree/tree.go` refactored to wrap `core` (preserving existing TUI field names and API)
- [ ] `internal/model/build.go` refactored to delegate to `core.BuildTree()` and `core.BuildAdminTree()`
- [ ] `internal/model/model.go` updated: `Root.CoreRoot` field, `RebuildFromCore()` method
- [ ] `internal/admin/handlers/content.go` uses `core.BuildFromRows()`
- [ ] `internal/admin/pages/content_list.templ` walks tree via pointers
- [ ] Existing tests updated and passing

---

## Phase 3: Tree Composition Layer

**Package:** `internal/tree/core/compose.go` (lives in the shared core)

### 3.1 Key Principle: Reuse the Existing Tree Build

Composition does NOT introduce a new data-fetching or tree-assembly pipeline. For each referenced content ID, it runs the same `core.BuildTree` assembly used by the slug handler:

1. The referenced content_data_id IS the root of the new subtree
2. Fetch that content_data row + all content_data whose parent_id chains down from it (descendants)
3. Fetch datatypes and content fields + field definitions for those content_data rows
4. Set the referenced node's datatype type to `_nested_root`
5. Call `core.BuildTree(cd, dt, cf, df)` — full datatype, field, content_data, content_field merging

No route lookup. No walking up to find a parent root. The referenced content_data_id is the root, period. Everything below it is its tree.

### 3.2 Concurrent Resolution Algorithm

Composition uses `errgroup` with bounded concurrency for fan-out/fan-in:

```
ComposeTrees(ctx, root, fetcher, opts)
  1. Walk tree looking for nodes where DatatypeType == "_reference"
  2. Collect all _reference nodes
  3. If none found or depth >= MaxDepth, return
  4. Create errgroup with MaxConcurrency limit
  5. For each _reference node, launch a goroutine:
     a. Collect all content_tree_ref field values (array of ULIDs)
     b. For each ULID in order:
        i.   Check visited set (sync.Mutex protected) — if seen, attach _system_log, continue
        ii.  Add to visited set
        iii. Call fetcher.FetchAndBuildTree(ctx, contentID)
        iv.  If fetch fails or not found — attach _system_log, continue
        v.   Recursively ComposeTrees on the subtree (depth + 1)
        vi.  Attach subtree root as child of the _reference node
     c. Subtree order matches field order (sequential within each goroutine)
  6. errgroup.Wait() — all _reference nodes resolve concurrently
  7. Return
```

Concurrency model:
- Multiple `_reference` nodes in the same tree resolve in **parallel** (one goroutine per `_reference` node)
- Multiple field values within a single `_reference` node resolve **sequentially** (preserves ordering)
- Nested references spawn additional goroutines recursively, bounded by `errgroup.SetLimit()`
- `context.Context` propagates cancellation from the HTTP request
- Graceful degradation: a failing goroutine attaches `_system_log` and returns nil (does not kill the errgroup). Other subtrees still resolve.

Example fan-out:

```
Page tree builds -> finds 3 _reference nodes (nav, body, footer)
  |-- goroutine: fetch nav tree -> builds -> no refs -> done
  |-- goroutine: fetch body tree -> builds -> finds 1 _reference (sidebar)
  |   |-- goroutine: fetch sidebar tree -> builds -> done
  |-- goroutine: fetch footer tree -> builds -> no refs -> done
All collapse back, subtrees attach to their _reference parents
```

### 3.3 Interface Design

```go
// TreeFetcher abstracts the full tree-building pipeline for composition.
// Each call fetches all data for a content tree and builds it using the
// standard core.BuildTree path.
type TreeFetcher interface {
    // FetchAndBuildTree fetches the content data descendants for the given
    // content data ID, sets the root type to _nested_root, and returns
    // the built tree via core.BuildTree.
    FetchAndBuildTree(ctx context.Context, id types.ContentID) (*Root, error)
}

// ComposeOptions controls composition behavior.
type ComposeOptions struct {
    MaxDepth       int // From config.json "composition_max_depth" (default 10)
    MaxConcurrency int // Max goroutines for parallel reference resolution (default 10)
}

// composeState tracks shared state across concurrent recursive calls.
type composeState struct {
    mu      sync.Mutex
    visited map[types.ContentID]bool
}

// ComposeTrees resolves all _reference nodes in the tree by fetching
// and building the referenced content trees via the standard BuildTree
// pipeline. Operates recursively up to MaxDepth levels. Uses errgroup
// for concurrent resolution of sibling references. Broken references
// produce _system_log nodes instead of errors.
func ComposeTrees(ctx context.Context, root *Root, fetcher TreeFetcher, opts ComposeOptions) error
```

`ComposeTrees` mutates the tree in place. Reference nodes gain children. The `_reference` node stays as a container — the `_nested_root` subtrees become its children:

```
_reference (label: "navigation")      <-- stays, acts as container
  +-- _nested_root (resolved nav)     <-- child: first content_tree_ref value
  |   +-- Home (link)
  |   +-- About (link)
  +-- _nested_root (resolved sec nav) <-- child: second content_tree_ref value
      +-- Blog (link)
```

### 3.4 TreeFetcher Implementation

The concrete `TreeFetcher` treats the referenced content_data_id as the root and fetches its descendants:

```go
type dbTreeFetcher struct {
    driver db.DbDriver
    log    Logger
}

func (f *dbTreeFetcher) FetchAndBuildTree(ctx context.Context, id types.ContentID) (*Root, error) {
    // 1. Fetch the referenced content_data row + all its descendants.
    //    id IS the root. The query fetches everything with parent_id
    //    chaining down from it.
    cd, err := f.driver.GetContentDataDescendants(id)
    if err != nil {
        return nil, fmt.Errorf("fetch tree for content %s: %w", id, err)
    }
    if len(cd) == 0 {
        return nil, fmt.Errorf("no content data found for %s", id)
    }

    // 2. Fetch datatypes for all content_data rows (parallel slice: dt[i] is for cd[i])
    dt, err := f.fetchDatatypesForContentData(ctx, cd)
    if err != nil {
        return nil, err
    }

    // 3. Fetch content fields + field definitions for all content_data rows
    cf, df, err := f.fetchFieldsForContentData(ctx, cd)
    if err != nil {
        return nil, err
    }

    // 4. Set the root's datatype type to _nested_root.
    //    Match by content_data_id, not index position.
    for i, c := range cd {
        if c.ContentDataID == id {
            dt[i].Type = string(types.DatatypeTypeNestedRoot)
            break
        }
    }

    // 5. Build tree using the standard pipeline
    root, _, err := BuildTree(cd, dt, cf, df)
    return root, err
}
```

**Helper method contracts:**

```go
// fetchDatatypesForContentData returns a parallel slice where dt[i]
// is the datatype for cd[i]. Iterates cd, calls driver.GetDatatype
// for each unique DatatypeID, caches results for duplicate DatatypeIDs.
// Contract: dt[i] corresponds to cd[i] by index.
// FK integrity at the DB level guarantees every cd[i].DatatypeID exists.
func (f *dbTreeFetcher) fetchDatatypesForContentData(ctx context.Context, cd []db.ContentData) ([]db.Datatypes, error)

// fetchFieldsForContentData returns parallel slices where cf[i]
// pairs with df[i]. For each content_data row, calls
// driver.ListContentFieldsByContentData to get field values,
// then driver.GetField for each field value's FieldID to get
// the field definition.
// FK integrity at the DB level guarantees every field reference exists.
func (f *dbTreeFetcher) fetchFieldsForContentData(ctx context.Context, cd []db.ContentData) ([]db.ContentFields, []db.Fields, error)
```

The referenced content_data_id is the root regardless of what its original type was. It gets `_nested_root`, `IsRootType()` returns true, `BuildTree` treats it as root and assembles everything below it.

### 3.5 Reference Node Contract

A `_reference` datatype node:
- Has `DatatypeType == "_reference"` (identified by `types.DatatypeTypeReference`)
- Has one or more fields of type `content_tree_ref`
- Each `content_tree_ref` field value is a content data ULID
- Fields are resolved in order — subtrees are attached as children in the same order
- If a field value is empty, it is skipped (no child attached, no error)
- If a referenced content ID cannot be found, a `_system_log` child is attached instead
- The `_reference` node itself stays as a container — subtrees are children, not replacements

### 3.6 System Log Node Construction

When a reference fails to resolve, the `_system_log` node should be informational — tell the developer what went wrong and how to fix it. The engine doesn't hand-hold, but it explains itself.

Each `_system_log` node carries structured data via paired `ContentFields` + `Fields` slices:

| Field Label | Value | Purpose |
|---|---|---|
| `error` | Short error type identifier | Machine-parseable |
| `message` | Human-readable explanation with troubleshooting hints | Developer-facing |
| `reference_id` | The content_data_id that was referenced | So you can look it up |
| `parent_label` | The `_reference` node's datatype label | Context for where in the tree |

**Error messages by failure type:**

- **not_found:** `"Reference 'navigation' points to content_data_id 01ABC... which does not exist. Check that the content has not been deleted, or update the reference field to point to a valid content_data_id."`
- **circular_reference:** `"Reference 'navigation' points to content_data_id 01ABC... which creates a circular reference (this content is already in the current composition chain). Each reference must point to content outside the current composition chain."`
- **max_depth:** `"Reference 'navigation' points to content_data_id 01ABC... but the maximum composition depth (${maxDepth}) has been reached. This limit is configurable via composition_max_depth in config.json. This usually means references are nested too deeply. Check for unintended reference chains."`
- **build_failed:** `"Reference 'navigation' resolved content_data_id 01ABC... but the subtree failed to build: ${err}. The referenced content may have data integrity issues."`

```go
// newSystemLogNode creates a synthetic _system_log node with structured fields
// describing a composition failure. Uses real DB struct fields.
func newSystemLogNode(parentLabel string, refID types.ContentID, errType, message string) *Node {
    nodeID := types.NewContentID()
    dtID := types.NewDatatypeID()

    // Generate field IDs for the 4 structured fields
    errorFieldID := types.NewFieldID()
    messageFieldID := types.NewFieldID()
    refFieldID := types.NewFieldID()
    parentFieldID := types.NewFieldID()

    now := types.TimestampNow()

    return &Node{
        ContentData: &db.ContentData{
            ContentDataID: nodeID,
            DatatypeID:    types.NullableDatatypeID{ID: dtID, Valid: true},
            Status:        types.ContentStatusPublished,
            DateCreated:   now,
            DateModified:  now,
        },
        Datatype: db.Datatypes{
            DatatypeID:   dtID,
            Label:        "_system_log",
            Type:         string(types.DatatypeTypeSystemLog),
            DateCreated:  now,
            DateModified: now,
        },
        // ContentFields: field values (what's stored in the DB content_fields table)
        ContentFields: []db.ContentFields{
            {
                ContentFieldID: types.NewContentFieldID(),
                ContentDataID:  types.NullableContentID{ID: nodeID, Valid: true},
                FieldID:        types.NullableFieldID{ID: errorFieldID, Valid: true},
                FieldValue:     errType,
                DateCreated:    now,
                DateModified:   now,
            },
            {
                ContentFieldID: types.NewContentFieldID(),
                ContentDataID:  types.NullableContentID{ID: nodeID, Valid: true},
                FieldID:        types.NullableFieldID{ID: messageFieldID, Valid: true},
                FieldValue:     message,
                DateCreated:    now,
                DateModified:   now,
            },
            {
                ContentFieldID: types.NewContentFieldID(),
                ContentDataID:  types.NullableContentID{ID: nodeID, Valid: true},
                FieldID:        types.NullableFieldID{ID: refFieldID, Valid: true},
                FieldValue:     refID.String(),
                DateCreated:    now,
                DateModified:   now,
            },
            {
                ContentFieldID: types.NewContentFieldID(),
                ContentDataID:  types.NullableContentID{ID: nodeID, Valid: true},
                FieldID:        types.NullableFieldID{ID: parentFieldID, Valid: true},
                FieldValue:     parentLabel,
                DateCreated:    now,
                DateModified:   now,
            },
        },
        // Fields: field definitions (what's stored in the DB fields table)
        Fields: []db.Fields{
            {
                FieldID:      errorFieldID,
                Label:        "error",
                Type:         types.FieldTypeText,
                DateCreated:  now,
                DateModified: now,
            },
            {
                FieldID:      messageFieldID,
                Label:        "message",
                Type:         types.FieldTypeText,
                DateCreated:  now,
                DateModified: now,
            },
            {
                FieldID:      refFieldID,
                Label:        "reference_id",
                Type:         types.FieldTypeText,
                DateCreated:  now,
                DateModified: now,
            },
            {
                FieldID:      parentFieldID,
                Label:        "parent_label",
                Type:         types.FieldTypeText,
                DateCreated:  now,
                DateModified: now,
            },
        },
    }
}
```

Client sites check for `_system_log` type nodes and handle them however they choose — hide in production, show a developer-visible error banner in staging, log to an error tracker, etc. The structured fields make it straightforward to build conditional rendering logic on the client side.

### 3.7 Cycle Detection

Maintain a `composeState` with a `sync.Mutex`-protected `map[types.ContentID]bool` visited set shared across all goroutines within a single `ComposeTrees` call chain. Before building a referenced tree, check if its content data ID is in the visited set. If so, inject a `_system_log` child with `errType: "circular_reference"`.

### 3.8 Depth Limit

Track current depth as an integer passed through recursion. If `depth >= opts.MaxDepth`, inject `_system_log` children for remaining references with `errType: "max_depth"`. Depth is per-branch — a 3-deep nav chain doesn't count against a 1-deep footer.

Default `MaxDepth` is 10. Configurable via `composition_max_depth` in `config.json`. The slug handler reads it from `config.Config.CompositionMaxDepth()` and passes it into `ComposeOptions`.

### 3.9 New DB Query

**Required:** `GetContentDataDescendants` — fetches a content_data row and all its descendants (everything with parent_id chaining down from it).

```sql
-- name: GetContentDataDescendants :many
WITH RECURSIVE tree(content_data_id) AS (
    SELECT content_data_id FROM content_data WHERE content_data_id = ?
    UNION ALL
    SELECT cd.content_data_id FROM content_data cd
    JOIN tree t ON cd.parent_id = t.content_data_id
)
SELECT cd.* FROM content_data cd
JOIN tree t ON cd.content_data_id = t.content_data_id;
```

The referenced ID is the root. The CTE walks down from it. All three database backends support recursive CTEs (SQLite 3.8.3+, MySQL 8.0+, PostgreSQL 8.4+). After adding to the appropriate `queries*.sql` files, run `just sqlc`, add to `DbDriver` interface, implement on all three wrapper structs.

**Index requirement:** `parent_id` on `content_data` should already have an index (it's a FK). Verify this exists across all three backends.

### 3.10 Caching

Per-request cache shared across goroutines within a single composition call. Uses `sync.Mutex` since multiple goroutines may read/write concurrently:

```go
type CachedFetcher struct {
    inner TreeFetcher
    mu    sync.Mutex
    cache map[types.ContentID]*Root
}

func (c *CachedFetcher) FetchAndBuildTree(ctx context.Context, id types.ContentID) (*Root, error) {
    c.mu.Lock()
    if root, ok := c.cache[id]; ok {
        c.mu.Unlock()
        return root, nil
    }
    c.mu.Unlock()

    root, err := c.inner.FetchAndBuildTree(ctx, id)
    if err != nil {
        return nil, err
    }

    c.mu.Lock()
    c.cache[id] = root
    c.mu.Unlock()
    return root, nil
}
```

A page with two references to the same nav tree builds it once. The cache is created per HTTP request — no cross-request sharing, no stale data.

### 3.11 Config Integration

**File:** `internal/config/config.go`

Add to `Config` struct:

```go
// Tree composition depth limit
Composition_Max_Depth int `json:"composition_max_depth"` // default 10
```

Add accessor method (same pattern as `MaxUploadSize()`):

```go
// CompositionMaxDepth returns the configured maximum composition depth.
// Falls back to 10 if no positive value is configured.
func (c Config) CompositionMaxDepth() int {
    if c.Composition_Max_Depth <= 0 {
        return 10
    }
    return c.Composition_Max_Depth
}
```

### 3.12 Integration Points

**Router API (`internal/router/slugs.go`):**

```go
// Build primary tree (existing code)
root, err := model.BuildTree(log, cd, dt, cf, df)

// Compose referenced subtrees (new code)
fetcher := &CachedFetcher{
    inner: &dbTreeFetcher{driver: d, log: log},
    cache: make(map[types.ContentID]*Root),
}
err = ComposeTrees(r.Context(), root.CoreRoot, fetcher, ComposeOptions{
    MaxDepth:       cfg.CompositionMaxDepth(),
    MaxConcurrency: 10,
})
// Rebuild model tree from composed core tree
root.RebuildFromCore()
// root is now composed with all referenced subtrees attached
```

**Admin panel:** Shows individual trees for editing, not composed ones.

**TUI:** Edits individual trees, no composition.

### 3.13 Deliverables

- [ ] `internal/tree/core/compose.go` — `ComposeTrees()`, `TreeFetcher` interface, `ComposeOptions`, `composeState`, `newSystemLogNode()`
- [ ] `internal/tree/core/compose_test.go` — unit tests with mock fetcher: single reference, multiple references on one node, nested references, cycle detection, depth limit, broken reference producing `_system_log`, empty field value skipped, concurrent resolution of sibling references
- [ ] `internal/tree/core/cached_fetcher.go` — `CachedFetcher` wrapper with `sync.Mutex`
- [ ] `Composition_Max_Depth` field added to `internal/config/config.go` Config struct with `CompositionMaxDepth()` accessor (default 10)
- [ ] New SQL query `GetContentDataDescendants` (recursive CTE) in all three backends + `just sqlc`
- [ ] `DbDriver` interface updated with new method, implemented on all three wrapper structs
- [ ] `dbTreeFetcher` implementation with `fetchDatatypesForContentData` and `fetchFieldsForContentData` helpers
- [ ] `internal/router/slugs.go` updated to call `ComposeTrees()` after building primary tree, then `RebuildFromCore()`
- [ ] Integration test: page with nav + body + footer references, all resolved correctly
- [ ] Integration test: broken reference produces `_system_log` node with correct message
- [ ] Integration test: circular reference detected and handled gracefully
- [ ] Integration test: concurrent resolution — multiple references resolve in parallel

---

## Phase 4: Admin Panel Integration

Depends on Phase 2 (tree refactor). Independent of Phase 3 (composition).

### 4.1 Changes

- `internal/admin/handlers/content.go` calls `core.BuildFromRows()` instead of passing raw rows
- `internal/admin/pages/content_list.templ` renders tree by walking `core.Node` pointers
- `ContentTreeNodes` templ component receives `*core.Root` instead of `[]db.GetContentTreeByRouteRow`
- Sibling ordering is now correct (was random before)
- Orphan nodes are handled (were silently dropped before)
- Circular references are detected (would have caused infinite recursion in templ before)

### 4.2 Deliverables

- [ ] `internal/admin/handlers/content.go` updated
- [ ] `internal/admin/pages/content_list.templ` refactored to pointer-based traversal
- [ ] `just admin-generate` run to regenerate templ Go code
- [ ] Visual verification via shakespeare (screenshot before/after)

---

## Implementation Order

```
Week 1:  Phase 1 (reserved types) + Phase 2 (tree refactor) in parallel
Week 2:  Phase 4 (admin integration) + Phase 3 (composition) in parallel
Week 3:  Integration testing, edge cases, documentation
```

All phases can be developed on feature branches and merged independently, with Phase 3 merging last since it depends on both Phase 1 and Phase 2.

## Composition Example

A page with navigation, body, and footer:

```
Page (_root, route: /about)
|
+-- Menu (_reference, label: "navigation")
|   |  content_tree_ref field value: "01ABC..."  -> nav tree
|   |  content_tree_ref field value: "01DEF..."  -> secondary nav
|   |
|   +-- Nav Tree Root (_nested_root, resolved from 01ABC...)
|   |   +-- Home (link)
|   |   +-- About (link)
|   |   +-- Contact (link)
|   |
|   +-- Secondary Nav (_nested_root, resolved from 01DEF...)
|       +-- Blog (link)
|       +-- Docs (link)
|
+-- Body (page_section)
|   +-- Hero (hero_banner)
|   +-- Content (richtext_block)
|
+-- Footer (_reference, label: "footer")
    |  content_tree_ref field value: "01GHI..."  -> footer tree
    |  content_tree_ref field value: "01XXX..."  -> deleted/missing
    |
    +-- Footer Tree Root (_nested_root, resolved from 01GHI...)
    |   +-- Links (link_group)
    |   +-- Copyright (text_block)
    |
    +-- _system_log (error: "not_found", reference_id: "01XXX...")
```

## Concurrent Resolution Visualization

```
Request arrives: GET /about
  |
  v
model.BuildTree(cd, dt, cf, df)        <- synchronous, builds primary tree
  |
  v
ComposeTrees(ctx, root.CoreRoot, fetcher, opts)
  |
  +-- find _reference nodes: [nav, footer]
  |
  +-- errgroup.Go (nav):                <- goroutine 1
  |   +-- fetch nav tree (01ABC)
  |   +-- fetch sec nav tree (01DEF)
  |   +-- ComposeTrees(subtree)         <- recursive: finds no refs, returns
  |   +-- attach both as children
  |
  +-- errgroup.Go (footer):             <- goroutine 2 (concurrent with nav)
  |   +-- fetch footer tree (01GHI)
  |   +-- fetch 01XXX -> not found
  |   +-- attach footer subtree + _system_log
  |
  +-- errgroup.Wait()                   <- blocks until both complete
  |
  v
root.RebuildFromCore()                  <- re-convert core tree to model types
  |
  v
Serialize to JSON, send response
```
