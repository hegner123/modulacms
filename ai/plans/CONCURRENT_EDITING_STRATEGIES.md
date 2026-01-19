# Concurrent Editing Strategies for ModulaCMS

## Executive Summary

**Recommended approach: Git-Like Merge + Tree Snapshots + Heartbeat Awareness**

Simple, proven approach inspired by WordPress ACF and Git:
- **Full tree snapshots**: Save entire content tree to root's `history` column on each save
- **Git-like merge**: Field-level conflict detection, auto-merge non-conflicting changes
- **Safety over cleverness**: Tree structure conflicts = duplicate nodes (user resolves manually)
- **Real-time heartbeat**: Show "User X editing" with second-level updates
- **DB → S3 aging**: Keep last 10 versions in DB (instant), older than 1 week → S3
- **S3 failure = warning**: Don't block saves, show warning banner until S3 recovers
- **Atomic**: Single record update, no distributed state

**Result**: Simple to implement, handles 90% of conflicts automatically, proven pattern (Git), works like WordPress ACF (which powers millions of sites), and atomic updates prevent data loss.

## Problem Statement

ModulaCMS needs a concurrent editing strategy that:
- Allows multiple users to work simultaneously (better than WordPress one-at-a-time)
- Doesn't require real-time websocket infrastructure (simpler than Notion)
- Fits the atomic nature of our tree-based content structure
- Is feasible for client implementations
- Prevents data loss from conflicting edits
- Provides full audit trail and undo/restore capability

## Current Architecture Advantages

Our tree structure with sibling pointers provides several atomic operations:
- Each content node is independent
- Sibling pointers enable O(1) reordering
- Parent-child relationships are explicit
- Field data is separate from tree structure

This atomicity gives us natural boundaries for locking/conflict resolution.

## Strategy Options (Ordered from Simplest to Most Complex)

### 1. Optimistic Locking with Version Numbers

**Concept**: Track version number on each content_data record. On save, check version matches.

**Implementation**:
```sql
-- Add to content_data table
ALTER TABLE content_data ADD COLUMN version INTEGER DEFAULT 1;

-- On update, increment and check
UPDATE content_data
SET field_value = ?, version = version + 1
WHERE content_data_id = ? AND version = ?;

-- If rows affected = 0, version mismatch = conflict
```

**Flow**:
1. User A and B load content (both see version 5)
2. User A saves → version becomes 6
3. User B tries to save with version 5 → fails
4. User B gets conflict message: "Content changed, please refresh"

**Pros**:
- Simple to implement (one column, one WHERE clause)
- No locks held during editing
- Works across HTTP requests
- Database-agnostic

**Cons**:
- User B loses their work if they don't merge manually
- No visibility into who's editing
- Frustrating for slow editors

**Best for**: Low-traffic sites, occasional edits, technical users

---

### 2. Last-Write-Wins with Conflict Detection + Undo

**Concept**: Allow overwrites but track change history for recovery.

**Implementation**:
```sql
-- Add revision table
CREATE TABLE content_revisions (
    revision_id INTEGER PRIMARY KEY,
    content_data_id INTEGER,
    field_value TEXT,
    modified_by INTEGER,
    modified_at TEXT,
    version INTEGER,
    FOREIGN KEY (content_data_id) REFERENCES content_data(content_data_id)
);

-- Before any update, save current state to revisions
INSERT INTO content_revisions
SELECT NULL, content_data_id, field_value, modified_by, modified_at, version
FROM content_data WHERE content_data_id = ?;

-- Then update normally
UPDATE content_data SET field_value = ?, modified_by = ?, modified_at = ?
WHERE content_data_id = ?;
```

**Flow**:
1. Every save creates a revision
2. If save happens within X minutes of another user's save, show warning
3. Provide UI to view/restore previous revisions

**Pros**:
- Never lose data - everything is recoverable
- Simple mental model (just save)
- No blocking, no conflicts
- Easy to audit changes

**Cons**:
- Storage grows with edits
- Still possible to overwrite
- Requires UI for revision browsing

**Best for**: Content sites where audit trail matters, teams that communicate

---

### 3. Advisory Locks with Active Editor Display

**Concept**: Show who's currently editing but don't enforce locks.

**Implementation**:
```sql
CREATE TABLE edit_sessions (
    session_id INTEGER PRIMARY KEY,
    content_data_id INTEGER,
    user_id INTEGER,
    started_at TEXT,
    last_activity TEXT,
    FOREIGN KEY (content_data_id) REFERENCES content_data(content_data_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- When user opens content for editing
INSERT INTO edit_sessions (content_data_id, user_id, started_at, last_activity)
VALUES (?, ?, datetime('now'), datetime('now'));

-- Heartbeat every 30 seconds
UPDATE edit_sessions
SET last_activity = datetime('now')
WHERE session_id = ?;

-- Clean up stale sessions (>5 minutes inactive)
DELETE FROM edit_sessions
WHERE last_activity < datetime('now', '-5 minutes');

-- Before saving, check for active editors
SELECT user_id FROM edit_sessions
WHERE content_data_id = ?
AND last_activity > datetime('now', '-1 minute');
```

**Flow**:
1. User A opens content → edit_sessions entry created
2. User B opens same content → sees "User A is editing (2 minutes ago)"
3. User B can choose to:
   - Wait for User A to finish
   - Edit anyway with warning
   - Message User A (if chat integration exists)
4. On save, combine with optimistic locking (strategy #1)

**Pros**:
- Social awareness prevents most conflicts
- Non-blocking (warnings, not errors)
- Works well with version checking
- Heartbeat detects crashes/disconnects

**Cons**:
- Requires periodic heartbeat requests
- UI complexity (showing active editors)
- Users can ignore warnings

**Best for**: Small teams, collaborative environments, CMS with active users

---

### 4. Field-Level Locking

**Concept**: Lock individual fields, not entire content nodes.

**Implementation**:
```sql
CREATE TABLE field_locks (
    lock_id INTEGER PRIMARY KEY,
    content_field_id INTEGER,
    user_id INTEGER,
    locked_at TEXT,
    expires_at TEXT,
    FOREIGN KEY (content_field_id) REFERENCES content_field(content_field_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

-- When user focuses on a field
INSERT INTO field_locks (content_field_id, user_id, locked_at, expires_at)
VALUES (?, ?, datetime('now'), datetime('now', '+2 minutes'));

-- Extend lock on continued activity
UPDATE field_locks
SET expires_at = datetime('now', '+2 minutes')
WHERE lock_id = ?;

-- Check lock before save
SELECT user_id FROM field_locks
WHERE content_field_id = ?
AND expires_at > datetime('now')
AND user_id != ?;
```

**Flow**:
1. User A edits "title" field → locked
2. User B can edit "body" field simultaneously → no conflict
3. User B tries to edit "title" → sees "User A is editing this field"
4. Locks auto-expire after 2 minutes of inactivity

**Pros**:
- Granular control reduces conflicts
- Multiple users can edit same content
- Natural fit for form-based editing
- Auto-expiring locks prevent deadlocks

**Cons**:
- More complex to implement
- Requires frontend field-tracking
- Overhead for many fields

**Best for**: Content with many independent fields, large editing teams

---

### 5. Operational Transformation (Lite)

**Concept**: Track operations, not final state. Merge conflicting operations.

**Implementation**:
```sql
CREATE TABLE operations_log (
    op_id INTEGER PRIMARY KEY,
    content_data_id INTEGER,
    user_id INTEGER,
    operation TEXT, -- 'insert', 'delete', 'replace'
    field_name TEXT,
    position INTEGER,
    old_value TEXT,
    new_value TEXT,
    timestamp TEXT,
    base_version INTEGER,
    FOREIGN KEY (content_data_id) REFERENCES content_data(content_data_id)
);
```

**Flow** (simplified):
1. User A and B both load version 5
2. User A: changes title "Hello" → "Hello World" (operation: append " World" at position 5)
3. User B: changes title "Hello" → "Hello!" (operation: append "!" at position 5)
4. User A saves first → version 6
5. User B saves → server detects base_version=5, current=6
6. Server applies User B's operation to version 6: "Hello World!"
7. Result: "Hello World!" (merged)

**Pros**:
- True concurrent editing
- Automatic merge of compatible changes
- No data loss
- Best user experience

**Cons**:
- Complex to implement correctly
- Only works well for text fields
- Conflicts still possible (need resolution UI)
- Hard to reason about edge cases

**Best for**: Rich text editors, documentation, collaborative writing

---

## Recommended Approach: Hybrid Strategy

Combine **#2 (Revisions)** + **#3 (Advisory Locks)** + **#1 (Optimistic Locking)**

### Existing Infrastructure to Leverage

**ModulaCMS already has revision infrastructure!**

All content tables have `history TEXT` columns:
- `content_data.history`
- `content_fields.history`
- `admin_content_data.history`
- `admin_content_fields.history`
- `routes.history`
- `datatypes.history`
- And more...

Each table has corresponding `*HistoryEntry` structs defined:
- `ContentDataHistoryEntry`
- `ContentFieldsHistoryEntry`
- etc.

The `internal/db/history.go` file provides:
- `MapHistory[T Historied](entry T)` - Appends new history entry to JSON array
- `PopHistory[T Historied](entry T)` - Retrieves/removes last entry (incomplete)
- `Historied` interface with `MapHistoryEntry()` and `GetHistory()` methods

### Tree Snapshot Architecture

**Key principle: Full tree snapshot at root level, Git-like merge on conflicts**

This is simple and atomic:

1. **Any save** → Full content tree saved to root `content_data.history`
2. **Version mismatch** → Server runs Git-like 3-way merge:
   - Base: Version user started editing from
   - Theirs: Current version on server
   - Yours: User's changes
   - Auto-merge: Different fields/nodes modified
   - Conflict: Same field on same node modified by both
3. **Tree structure conflict** → Create duplicate nodes (safety over cleverness)
4. **Root delete** → Full tree saved to `datatype.history` (recovery point)
5. **DB aging** → After 1 week, move old snapshots to S3, keep last 10 in DB

**Benefits:**
- **Atomic**: Single UPDATE statement, no distributed state
- **Simple restore**: Just apply old snapshot, no reconstruction
- **Proven pattern**: Git merge users already understand
- **No orphans**: Full tree always consistent
- **ACF does this**: Works for millions of WordPress sites

**What's needed:**
1. Tree serialization/deserialization
2. 3-way merge algorithm (field-level diff)
3. Conflict detection and resolution UI
4. Heartbeat for "who's editing" awareness
5. S3 aging with failure warnings

### Implementation Plan

**Phase 1: Tree Snapshot Foundation (Week 1)**
- Add `version INTEGER DEFAULT 1` to root `content_data` records
- Implement tree serialization: `SerializeContentTree(rootID) → JSON`
- Implement tree deserialization: `DeserializeContentTree(JSON) → Tree`
- Add `TreeHistoryEntry` struct:
  ```go
  type TreeHistoryEntry struct {
      Hash      string      `json:"hash"`
      Version   int64       `json:"version"`
      Timestamp string      `json:"timestamp"`
      AuthorID  int64       `json:"author_id"`
      Tree      ContentTree `json:"tree"`
  }
  ```
- Modify save operation to append full tree to root.history
- Trim to last 10 entries in DB

**Phase 2: Git-Like Merge (Week 2)**
- Implement 3-way diff: `DiffTrees(base, theirs, yours) → Changes`
- Field-level change detection:
  ```go
  type FieldChange struct {
      NodeID   int64
      FieldName string
      OldValue string
      NewValue string
  }
  ```
- Conflict detection: same field on same node changed by both
- Auto-merge: apply both changesets if no conflicts
- Tree structure conflicts: duplicate nodes (node_123 and node_123_copy)
- API endpoint: `POST /api/v1/content/:root_id/save` with merge logic

**Phase 3: Heartbeat & Awareness (Week 3)**
- Create `edit_sessions` table:
  ```sql
  CREATE TABLE edit_sessions (
      session_id INTEGER PRIMARY KEY,
      content_root_id INTEGER NOT NULL,
      user_id INTEGER NOT NULL,
      last_heartbeat TEXT NOT NULL,
      FOREIGN KEY (content_root_id) REFERENCES content_data(content_data_id),
      FOREIGN KEY (user_id) REFERENCES users(user_id)
  );
  CREATE INDEX idx_sessions_root ON edit_sessions(content_root_id);
  CREATE INDEX idx_sessions_heartbeat ON edit_sessions(last_heartbeat);
  ```
- Heartbeat endpoint: `PUT /api/v1/edit-session/:root_id` (every 5 seconds)
- Active editors endpoint: `GET /api/v1/content/:root_id/editors`
- Cleanup stale sessions (>30 seconds inactive)
- UI: Show "User X, User Y editing" banner

**Phase 4: S3 Aging & Cleanup (Week 4)**
- Background job runs daily
- For each root with history:
  - Parse history array
  - Find entries older than 7 days
  - Upload to S3: `history/{route_id}/{root_id}/v{version}_{timestamp}.json`
  - On S3 confirmation: remove from history array
  - On S3 failure: keep in DB, add warning flag
- S3 failure detection:
  - Set `s3_sync_status` flag on content_data
  - API includes warning in JSON response if flag set
  - UI shows persistent warning banner
  - TUI shows warning in bottom toolbar

**Phase 5: Conflict Resolution UI (Week 5)**
- When save returns conflict:
  ```json
  {
    "status": "conflict",
    "conflicts": [
      {
        "node_id": 123,
        "field": "title",
        "base_value": "Hello",
        "their_value": "Hello World",
        "your_value": "Hello There"
      }
    ]
  }
  ```
- Client shows conflict resolution UI:
  - Side-by-side diff view
  - Buttons: "Keep mine", "Keep theirs", "Edit manually"
- Resubmit with conflict resolutions:
  ```json
  {
    "resolve_conflicts": {
      "node_123.title": "your_value"  // or "their_value" or custom
    }
  }
  ```

**Phase 6: History Browsing & Restore (Week 6)**
- API endpoints:
  - `GET /api/v1/content/:root_id/history` - List versions (last 10 from DB)
  - `GET /api/v1/content/:root_id/history/:version` - Get specific snapshot
  - `GET /api/v1/content/:root_id/history/s3/:key` - Fetch from S3 if archived
  - `POST /api/v1/content/:root_id/restore/:version` - Restore to version
- UI: Timeline view with:
  - Version number, timestamp, author
  - Preview diff from previous version
  - "Restore to this version" button
- Restore creates new version with old tree as content

### Why This Works

1. **Atomic**: Single UPDATE statement, no partial failures or orphaned data
2. **Proven Pattern**: Git merge algorithm users already understand
3. **ACF Precedent**: WordPress ACF stores full page trees in one column, works for millions
4. **90% Auto-Merge**: Field-level conflicts are rare in practice
5. **Safety First**: Tree structure conflicts = duplicate nodes (user decides)
6. **Simple Restore**: Apply old snapshot, no reconstruction or dependency resolution
7. **DB Fast, S3 Cheap**: Last 10 versions instant, old versions archived cheaply
8. **Failure Tolerant**: S3 down? Warning banner, keep working
9. **Real-time Awareness**: Second-level heartbeat shows who's editing

### API Endpoints Needed

```
# Editing Sessions
POST   /api/v1/edit-session/:root_id         # Start editing session
PUT    /api/v1/edit-session/:root_id         # Heartbeat (every 5sec)
DELETE /api/v1/edit-session/:root_id         # End session
GET    /api/v1/content/:root_id/editors      # Who's editing?

# Save with Merge
POST   /api/v1/content/:root_id/save         # Save tree (returns conflict or success)
POST   /api/v1/content/:root_id/save/resolve # Save with conflict resolutions

# History
GET    /api/v1/content/:root_id/history      # List versions (last 10 from DB)
GET    /api/v1/content/:root_id/history/:ver # Get specific snapshot
GET    /api/v1/content/:root_id/history/s3/:key # Fetch from S3 (if archived)
POST   /api/v1/content/:root_id/restore/:ver # Restore to version

# Datatype Recovery (on root delete)
GET    /api/v1/datatype/:id/deleted-trees    # List deleted trees for this datatype
POST   /api/v1/datatype/:id/recover/:hash    # Recover deleted tree
```

### Database Schema Additions

```sql
-- Add version tracking to root content_data records only
ALTER TABLE content_data ADD COLUMN version INTEGER DEFAULT 1;
ALTER TABLE content_data ADD COLUMN s3_sync_status TEXT DEFAULT NULL; -- 'pending', 'failed', NULL

-- History column already exists!
-- Will store array of TreeHistoryEntry JSON

-- Edit sessions for heartbeat awareness
CREATE TABLE edit_sessions (
    session_id INTEGER PRIMARY KEY AUTOINCREMENT,
    content_root_id INTEGER NOT NULL,  -- Root content_data_id
    user_id INTEGER NOT NULL,
    last_heartbeat TEXT NOT NULL,
    FOREIGN KEY (content_root_id) REFERENCES content_data(content_data_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    UNIQUE(content_root_id, user_id)  -- One session per user per tree
);
CREATE INDEX idx_sessions_root ON edit_sessions(content_root_id);
CREATE INDEX idx_sessions_heartbeat ON edit_sessions(last_heartbeat);

-- No other tables needed!
-- Datatype already has history column for deleted trees
```

### History Column Format

The root `content_data.history` stores full tree snapshots:

```json
[
  {
    "hash": "sha256:a3f5b8c9d2e1f4...",
    "version": 1,
    "timestamp": "2024-01-15T14:20:00Z",
    "author_id": 1,
    "tree": {
      "root": {
        "content_data_id": 123,
        "route_id": 5,
        "datatype_id": 2,
        "parent_id": null,
        "first_child_id": 456
      },
      "nodes": [
        {
          "content_data_id": 456,
          "parent_id": 123,
          "datatype_id": 3,
          "fields": [
            {"field_id": 1, "field_value": "Hello World"},
            {"field_id": 2, "field_value": "Page body content..."}
          ]
        },
        {
          "content_data_id": 457,
          "parent_id": 123,
          "datatype_id": 3,
          "fields": [
            {"field_id": 1, "field_value": "Second Section"},
            {"field_id": 2, "field_value": "More content..."}
          ]
        }
      ]
    }
  },
  {
    "hash": "sha256:d7e2a1b9c4f8...",
    "version": 2,
    "timestamp": "2024-01-16T09:15:00Z",
    "author_id": 2,
    "s3_key": "history/route_5/root_123/v2_20240116T091500.json"
  }
]
```

**Key fields:**
- `hash`: SHA256 of tree JSON (for integrity)
- `version`: Incremental counter
- `timestamp`: When this version was saved
- `author_id`: Who saved it
- `tree`: Full tree inline (if recent)
- `s3_key`: S3 location (if archived after 7 days)

**Last 10 versions kept in DB, older moved to S3.**

### Core Implementation Examples

**Step 1: Tree Serialization**

```go
// in internal/db/content_tree.go

type ContentTree struct {
    Root  ContentData         `json:"root"`
    Nodes []ContentDataWithFields `json:"nodes"`
}

type ContentDataWithFields struct {
    ContentData
    Fields []ContentFields `json:"fields"`
}

// Serialize entire content tree from root
func SerializeContentTree(dbc *DbDriver, rootID int64) (*ContentTree, error) {
    tree := &ContentTree{}

    // Get root node
    root, err := dbc.GetContentData(rootID)
    if err != nil {
        return nil, err
    }
    tree.Root = *root

    // Recursively get all descendants
    nodes, err := dbc.GetTreeDescendants(rootID)
    if err != nil {
        return nil, err
    }

    // For each node, get its fields
    for _, node := range nodes {
        fields, _ := dbc.GetContentFieldsByContentDataID(node.ContentDataID)
        tree.Nodes = append(tree.Nodes, ContentDataWithFields{
            ContentData: node,
            Fields:      fields,
        })
    }

    return tree, nil
}

// Deserialize and apply tree to database
func DeserializeContentTree(dbc *DbDriver, tree *ContentTree) error {
    // Start transaction
    tx, _ := dbc.DB.Begin()
    defer tx.Rollback()

    // Create/update root
    _, err := tx.Exec(`
        INSERT OR REPLACE INTO content_data
        (content_data_id, route_id, datatype_id, parent_id, first_child_id, ...)
        VALUES (?, ?, ?, ?, ?, ...)
    `, tree.Root.ContentDataID, tree.Root.RouteID, ...)

    // Create/update all nodes and their fields
    for _, node := range tree.Nodes {
        // Insert node
        tx.Exec(`INSERT OR REPLACE INTO content_data ...`)

        // Insert fields
        for _, field := range node.Fields {
            tx.Exec(`INSERT OR REPLACE INTO content_fields ...`)
        }
    }

    return tx.Commit()
}
```

**Step 2: Save with History**

```go
// in internal/db/content_data.go

type TreeHistoryEntry struct {
    Hash      string       `json:"hash"`
    Version   int64        `json:"version"`
    Timestamp string       `json:"timestamp"`
    AuthorID  int64        `json:"author_id"`
    Tree      *ContentTree `json:"tree,omitempty"`   // Inline if recent
    S3Key     string       `json:"s3_key,omitempty"` // If archived
}

func (dbc *DbDriver) SaveContentTree(rootID int64, newTree *ContentTree, userID int64, baseVersion int64) error {
    // 1. Get current root
    root, err := dbc.GetContentData(rootID)
    if err != nil {
        return err
    }

    // 2. Version mismatch = need to merge
    if root.Version != baseVersion {
        return dbc.MergeAndSave(root, newTree, userID, baseVersion)
    }

    // 3. No conflict - serialize current tree before overwriting
    currentTree, _ := SerializeContentTree(dbc, rootID)

    // 4. Parse existing history
    var history []TreeHistoryEntry
    if root.History.Valid {
        json.Unmarshal([]byte(root.History.String), &history)
    }

    // 5. Add current state to history
    hash := sha256.Sum256([]byte(mustJSON(currentTree)))
    history = append(history, TreeHistoryEntry{
        Hash:      fmt.Sprintf("sha256:%x", hash),
        Version:   root.Version,
        Timestamp: time.Now().Format(time.RFC3339),
        AuthorID:  userID,
        Tree:      currentTree,
    })

    // 6. Trim to last 10
    if len(history) > 10 {
        history = history[len(history)-10:]
    }

    // 7. Apply new tree
    if err := DeserializeContentTree(dbc, newTree); err != nil {
        return err
    }

    // 8. Update root with new history and version
    historyJSON, _ := json.Marshal(history)
    _, err = dbc.DB.Exec(`
        UPDATE content_data
        SET history = ?, version = ?, date_modified = ?
        WHERE content_data_id = ? AND version = ?
    `, historyJSON, root.Version+1, time.Now(), rootID, root.Version)

    return err
}
```

**Step 3: Git-Like 3-Way Merge**

```go
// in internal/db/merge.go

type FieldChange struct {
    NodeID    int64  `json:"node_id"`
    FieldID   int64  `json:"field_id"`
    OldValue  string `json:"old_value"`
    NewValue  string `json:"new_value"`
}

type MergeConflict struct {
    NodeID      int64  `json:"node_id"`
    FieldID     int64  `json:"field_id"`
    BaseValue   string `json:"base_value"`
    TheirValue  string `json:"their_value"`
    YourValue   string `json:"your_value"`
}

// 3-way diff: compare base → theirs and base → yours
func DiffTrees(base, theirs, yours *ContentTree) ([]FieldChange, []FieldChange, []MergeConflict) {
    theirChanges := []FieldChange{}
    yourChanges := []FieldChange{}
    conflicts := []MergeConflict{}

    // Build maps for quick lookup
    baseFields := buildFieldMap(base)
    theirFields := buildFieldMap(theirs)
    yourFields := buildFieldMap(yours)

    // Find changes
    for key, baseField := range baseFields {
        theirField, theirExists := theirFields[key]
        yourField, yourExists := yourFields[key]

        theirChanged := theirExists && theirField.FieldValue != baseField.FieldValue
        yourChanged := yourExists && yourField.FieldValue != baseField.FieldValue

        if theirChanged && yourChanged {
            // Both changed same field = CONFLICT
            conflicts = append(conflicts, MergeConflict{
                NodeID:     baseField.ContentDataID,
                FieldID:    baseField.FieldID,
                BaseValue:  baseField.FieldValue,
                TheirValue: theirField.FieldValue,
                YourValue:  yourField.FieldValue,
            })
        } else if theirChanged {
            theirChanges = append(theirChanges, FieldChange{
                NodeID:   baseField.ContentDataID,
                FieldID:  baseField.FieldID,
                OldValue: baseField.FieldValue,
                NewValue: theirField.FieldValue,
            })
        } else if yourChanged {
            yourChanges = append(yourChanges, FieldChange{
                NodeID:   baseField.ContentDataID,
                FieldID:  baseField.FieldID,
                OldValue: baseField.FieldValue,
                NewValue: yourField.FieldValue,
            })
        }
    }

    return theirChanges, yourChanges, conflicts
}

// MergeAndSave handles version conflict with 3-way merge
func (dbc *DbDriver) MergeAndSave(root *ContentData, yourTree *ContentTree, userID int64, baseVersion int64) error {
    // 1. Get base tree (version user started editing from)
    baseTree, err := dbc.GetHistoryVersion(root.ContentDataID, baseVersion)
    if err != nil {
        return fmt.Errorf("cannot find base version %d", baseVersion)
    }

    // 2. Get current tree (theirs - what's on server now)
    theirTree, _ := SerializeContentTree(dbc, root.ContentDataID)

    // 3. Run 3-way diff
    theirChanges, yourChanges, conflicts := DiffTrees(baseTree, theirTree, yourTree)

    // 4. If conflicts, return to user for resolution
    if len(conflicts) > 0 {
        return &MergeConflictError{Conflicts: conflicts}
    }

    // 5. No conflicts - auto-merge
    merged := baseTree.Clone()
    ApplyChanges(merged, theirChanges)
    ApplyChanges(merged, yourChanges)

    // 6. Save merged tree
    return dbc.SaveContentTree(root.ContentDataID, merged, userID, root.Version)
}

    // 2. Get next version number
    nextVersion := current.Version + 1

    // 3. Generate operation ID
    operationID := NewOperationID()

    // 4. Save current state to history before updating
    historyJSON, err := MapHistoryWithS3(
        current,
        operationID,
        "update",
        userID,
        nextVersion,
        nil, // not a restore
        dbc.Config,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to save history: %w", err)
    }
    params.History = sql.NullString{String: historyJSON, Valid: true}
    params.Version = nextVersion

    // 5. Perform the update (with new history and version)
    switch dbc.Driver {
    case "sqlite":
        return dbc.UpdateContentDataSQLite(params)
    case "mysql":
        return dbc.UpdateContentDataMySQL(params)
    case "postgres":
        return dbc.UpdateContentDataPostgres(params)
    default:
        return nil, fmt.Errorf("unsupported driver: %s", dbc.Driver)
    }
}
```

**Step 5: Move Operation (Linked Operations)**

```go
// Move content_data from one parent to another
func (dbc *DbDriver) MoveContentData(
    nodeID int64,
    oldParentID int64,
    newParentID int64,
    userID int64,
) error {
    // Generate single operation ID for linked operations
    operationID := NewOperationID()

    // Get current states
    node, _ := dbc.GetContentData(nodeID)
    oldParent, _ := dbc.GetContentData(oldParentID)
    newParent, _ := dbc.GetContentData(newParentID)

    // Create history entry for old parent (delete_child)
    oldParentHistory, _ := MapHistoryWithS3(
        oldParent,
        operationID,
        "delete_child",
        userID,
        oldParent.Version + 1,
        nil,
        dbc.Config,
    )

    // Create history entry for new parent (add_child)
    newParentHistory, _ := MapHistoryWithS3(
        newParent,
        operationID,
        "add_child",
        userID,
        newParent.Version + 1,
        nil,
        dbc.Config,
    )

    // Create history entry for node itself (move)
    nodeHistory, _ := MapHistoryWithS3(
        node,
        operationID,
        "move",
        userID,
        node.Version + 1,
        nil,
        dbc.Config,
    )

    // Update all three records with their histories
    // ... execute updates ...

    // Record in operations index for fast discovery
    RecordLinkedOperation(operationID, []int64{oldParentID, newParentID, nodeID})

    return nil
}
```

**Step 6: Restore with Overwrite-to-History**

```go
// in internal/db/history.go

// GetHistoryEntryByHash finds a specific history entry by its hash
func GetHistoryEntryByHash(entry Historied, hash string) (*HistoryEntry, error) {
    data := []byte(entry.GetHistory())
    var entries []HistoryEntry
    if err := json.Unmarshal(data, &entries); err != nil {
        return nil, err
    }

    for _, e := range entries {
        if e.Hash == hash {
            return &e, nil
        }
    }

    return nil, fmt.Errorf("history entry with hash %s not found", hash)
}

// FetchSnapshot retrieves snapshot data from inline or S3
func FetchSnapshot(snapshot HistorySnapshot, config *config.Config) ([]byte, error) {
    if snapshot.Type == "inline" {
        return json.Marshal(snapshot.Data)
    }

    if snapshot.Type == "s3" {
        return DownloadSnapshotFromS3(snapshot.Bucket, snapshot.Key)
    }

    return nil, fmt.Errorf("unknown snapshot type: %s", snapshot.Type)
}

// RestoreFromHistory restores entry to a previous version
// CRITICAL: Saves current state to history before overwriting
func RestoreFromHistory[T Historied](
    entry T,
    restoreToHash string,
    userID int64,
    config *config.Config,
) error {
    // 1. Find the history entry to restore
    historyEntry, err := GetHistoryEntryByHash(entry, restoreToHash)
    if err != nil {
        return err
    }

    // 2. Fetch snapshot data (from inline or S3)
    snapshotBytes, err := FetchSnapshot(historyEntry.Snapshot, config)
    if err != nil {
        return fmt.Errorf("failed to fetch snapshot: %w", err)
    }

    // 3. Parse snapshot into correct type
    var snapshotData T
    if err := json.Unmarshal(snapshotBytes, &snapshotData); err != nil {
        return err
    }

    // 4. BEFORE applying restore, save current state to history
    nextVersion := entry.GetVersion() + 1
    operationID := NewOperationID()

    currentStateHistory, err := MapHistoryWithS3(
        entry,
        operationID,
        "restore",
        userID,
        nextVersion,
        &restoreToHash, // Link back to what we're restoring
        config,
    )
    if err != nil {
        return fmt.Errorf("failed to save current state before restore: %w", err)
    }

    // 5. Apply the restored values to current entry
    // This is type-specific - example for ContentData:
    // entry.RouteID = snapshotData.RouteID
    // entry.ParentID = snapshotData.ParentID
    // ... etc

    // 6. Update history with the new entry (current state before restore)
    entry.UpdateHistory([]byte(currentStateHistory))

    // 7. Increment version
    entry.SetVersion(nextVersion)

    return nil
}

// CheckLinkedOperations warns if restoring half of a linked operation
func CheckLinkedOperations(operationID string) ([]LinkedOperation, error) {
    // Query history_operations table for this operation_id
    // Return all affected records

    // If found multiple records, warn user:
    // "This restore is part of a move operation. Restoring only this part
    //  will leave the node orphaned. Restore all affected records?"

    return nil, nil
}
```

**Step 7: Operations Index for Discovery**

```sql
-- Fast lookup for linked operations
CREATE TABLE history_operations (
    operation_id TEXT PRIMARY KEY,
    timestamp TEXT NOT NULL,
    route_id INTEGER NOT NULL,
    action TEXT NOT NULL,  -- "move", "delete_subtree", etc.
    affected_records TEXT NOT NULL,  -- JSON: {"content_data": [123, 456], "content_fields": [789]}
    FOREIGN KEY (route_id) REFERENCES routes(route_id)
);

CREATE INDEX idx_history_ops_timestamp ON history_operations(timestamp);
CREATE INDEX idx_history_ops_route ON history_operations(route_id);
```

```go
// Record a linked operation for discovery
func RecordLinkedOperation(
    operationID string,
    action string,
    routeID int64,
    affectedRecords map[string][]int64,
) error {
    recordsJSON, _ := json.Marshal(affectedRecords)

    return db.Exec(`
        INSERT INTO history_operations
        (operation_id, timestamp, route_id, action, affected_records)
        VALUES (?, ?, ?, ?, ?)
    `, operationID, time.Now().Format(time.RFC3339), routeID, action, recordsJSON)
}

// Find all records affected by an operation
func GetLinkedOperations(operationID string) (map[string][]int64, error) {
    var recordsJSON string
    err := db.QueryRow(`
        SELECT affected_records FROM history_operations WHERE operation_id = ?
    `, operationID).Scan(&recordsJSON)

    if err != nil {
        return nil, err
    }

    var records map[string][]int64
    json.Unmarshal([]byte(recordsJSON), &records)
    return records, nil
}
```

### Configuration Options

Allow clients to configure behavior:

```json
{
  "concurrent_editing": {
    "strategy": "advisory_locks",
    "edit_session_timeout": 300,
    "heartbeat_interval": 30,
    "conflict_resolution": "prompt_user",
    "show_active_editors": true
  },
  "history": {
    "enabled": true,
    "storage_threshold_bytes": 51200,
    "s3_bucket": "modulacms-history",
    "s3_prefix": "snapshots",
    "max_entries_per_record": 50,
    "retention": {
      "inline_days": 90,
      "s3_days": 365,
      "archive_after_days": 180
    },
    "cleanup": {
      "enabled": true,
      "schedule": "0 2 * * *",
      "min_retention_days": 1
    }
  }
}
```

**Configuration field explanations:**

- `storage_threshold_bytes`: Snapshots larger than this go to S3 (default: 50KB)
  - Suggested values: 10KB (heavy traffic), 50KB (standard), 100KB (lightweight)
- `s3_bucket`: S3-compatible bucket for large snapshots
- `s3_prefix`: Prefix/folder for snapshot objects (e.g., "snapshots")
- `max_entries_per_record`: Limit history array length (0 = unlimited)
- `retention.inline_days`: Keep inline snapshots this many days (0 = forever)
- `retention.s3_days`: Keep S3 snapshots this many days (0 = forever)
- `retention.archive_after_days`: Move old inline snapshots to archive table
- `cleanup.schedule`: Cron expression for cleanup job
- `cleanup.min_retention_days`: Safety floor (never delete if newer than this)

## Storage and Performance Considerations

### Distributed Storage Benefits

By storing history where changes occur, we eliminate redundancy:

**Traditional approach (bad):**
- Update single field → Save entire tree to history
- 100 field updates = 100 copies of entire tree
- Database bloat, slow queries

**Distributed approach (good):**
- Update single field → Save only that field's history
- 100 field updates = 100 field snapshots (tiny)
- Tree structure changes saved at node level
- Large deletions offloaded to S3

**Storage comparison example:**
```
Scenario: 500-node tree, user updates 10 fields

Traditional:
- 10 saves × 500 nodes × 2KB = 10MB in history column

Distributed + S3:
- 10 field saves × 200 bytes = 2KB inline
- Tree structure unchanged = 0 bytes
- Total: 2KB vs 10MB (5000x reduction)
```

### S3 Storage Strategy

**When snapshots exceed threshold (default 50KB):**

1. **Upload to S3:**
   - Path: `snapshots/{table}/{record_id}_v{version}_{timestamp}.json`
   - Example: `snapshots/content_data/123_v45_20260117T103045.json`
   - Content-Type: `application/json`
   - Metadata: hash, version, operation_id

2. **Store reference in history:**
   ```json
   {
     "snapshot": {
       "type": "s3",
       "bucket": "modulacms-history",
       "key": "snapshots/content_data/123_v45_20260117T103045.json",
       "size": 125840
     }
   }
   ```

3. **Cleanup on schedule:**
   - Cron job deletes S3 objects older than retention policy
   - Orphan detection: don't delete if still referenced in history
   - Configurable: 1 day to forever

**S3 access patterns:**
- **Write:** On large tree operations (delete, move subtree)
- **Read:** On restore from history (rare)
- **Delete:** Scheduled cleanup job (nightly)

Cost-effective: mostly writes, rare reads, eventual deletes.

### History Column Growth Management

Even with S3, inline history arrays grow. Strategies:

1. **Max entries limit**: Trim to last N entries (default: 50)
   - Oldest entries pruned first
   - Keeps most recent changes accessible
   - Configurable per use case

2. **Archive old history**: Move to separate table
   ```sql
   CREATE TABLE content_history_archive (
       archive_id INTEGER PRIMARY KEY,
       table_name TEXT NOT NULL,
       record_id INTEGER NOT NULL,
       archived_history TEXT,
       archived_at TEXT,
       s3_keys TEXT  -- JSON array of S3 keys to delete on cleanup
   );
   ```

3. **Cleanup job logic:**
   ```
   For each content record:
     1. Parse history array
     2. Find entries older than retention.inline_days
     3. Move to archive table
     4. Update record with trimmed history
     5. Schedule S3 cleanup if snapshot.type = "s3"
   ```

### Query Performance

**History column impact:**

| Operation | Impact | Mitigation |
|-----------|--------|------------|
| SELECT (normal queries) | None | History column not in WHERE/ORDER BY |
| UPDATE (append history) | Minimal | JSON append is fast, trimmed to max entries |
| Restore (read history) | Low | Indexed by hash, S3 fetch only if needed |
| Tree traversal | None | History not traversed, only current state |

**Optimization techniques:**

1. **Lazy loading:** Don't deserialize history unless requested
2. **Hash indexing:** O(1) lookup by hash for restores
3. **S3 prefetch:** If UI shows history, prefetch S3 snapshots in background
4. **Operations index:** Fast discovery of linked operations via separate table

### Database Size Impact

**Example: 10,000 content nodes, 50 edits each:**

Without distributed storage:
- 10,000 nodes × 50 versions × 2KB = 1GB

With distributed storage + S3:
- Field edits (90%): 450,000 × 200 bytes = 90MB inline
- Tree ops (10%): 50,000 × 50 bytes (S3 ref) = 2.5MB inline
- S3 storage: Tree ops × 10KB avg = 500MB (cheap, compressed)
- Total inline: ~95MB vs 1GB (10.5x reduction)

Plus: S3 is cheaper than database storage, can be compressed, and cleaned up aggressively.

## Alternative: Start Simple, Add Complexity Later

If timeline is tight, start with **completing just the existing history system**:
- Implement `MapHistoryEntry()` methods
- Call `MapHistory()` in Update operations
- Provide "undo" via history browsing
- Add version column + advisory locks later when needed

This gives 80% of the benefit with 20% of the complexity, and leverages infrastructure you've already built.

## Tree Structure Considerations

Our sibling-pointer tree adds unique constraints:

1. **Reordering conflicts**: If User A and User B both reorder siblings
   - Solution: Version the parent's `first_child_id`
   - On conflict, show tree diff UI

2. **Moving nodes**: If User A moves node X under Parent A, User B moves it under Parent B
   - Solution: Lock node during move operation
   - Or: Last move wins, with notification

3. **Deleting nodes**: If User A edits node, User B deletes it
   - Solution: Soft delete flag + revision check
   - Editing a deleted node restores it (with warning)

4. **Creating siblings**: Two users add child to same parent
   - Solution: Last write wins for `first_child_id`
   - Both nodes exist, order determined by timestamps

## Testing Strategy

Simulate concurrent scenarios:
```go
func TestConcurrentEdit(t *testing.T) {
    // Two goroutines edit same content
    // Verify both saves succeed or conflict detected
}

func TestAdvisoryLock(t *testing.T) {
    // User A starts session
    // User B checks active editors
    // Verify User B sees User A
}

func TestRevisionCreation(t *testing.T) {
    // Update content
    // Verify revision created
    // Verify old value preserved
}
```

## Metrics to Track

Monitor concurrent editing in production:
- Concurrent edit session frequency
- Conflict rate (version mismatches)
- Revision restoration frequency
- Average session duration
- Stale session cleanup rate

This data informs whether to add more sophisticated strategies later.

## Recommendation

**Start with Distributed History + S3 + Advisory Locks + Optimistic Locking**

This approach maximizes the value of ModulaCMS's existing infrastructure:

### What Makes This Ideal for ModulaCMS

1. **Leverages existing `history` columns** - Already in schema, just need to activate
2. **Fits the atomic tree structure** - Each node manages its own history
3. **Scales with storage, not connections** - S3 handles large snapshots
4. **No redundant data** - Only changed fields create history entries
5. **Works with HTTP** - No websockets or real-time infrastructure
6. **Granular restoration** - Restore field, node, or entire subtree
7. **Audit trail built-in** - Hash integrity + operation linking

### Implementation Phases

**MVP (Week 1-2):**
- Enhanced history structures (hash, operation_id, metadata)
- S3 integration with threshold
- Basic update operations with history capture
- Version column for optimistic locking

**Phase 2 (Week 3-4):**
- Linked operations (move = delete + add)
- Operations index table for discovery
- Restore with overwrite-to-history
- Orphan detection warnings

**Phase 3 (Week 5-6):**
- Advisory locks (edit sessions)
- History timeline API
- Cleanup jobs and retention policies
- UI for browsing/restoring history

### Key Design Decisions Summary

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| **Storage location** | Distributed (where change occurs) | Eliminates redundancy, enables granular restore |
| **Large snapshots** | S3 (threshold: 50KB default) | Cheaper, scalable, doesn't bloat database |
| **Operation IDs** | ULID | Time-sortable, no coordination, discover paired ops |
| **Hashing** | SHA256 of content fields | Integrity check, deduplication, restoration lookup |
| **Restore strategy** | Save current before overwrite | Creates bidirectional chain, can undo restores |
| **Linked operations** | Same operation_id | Move/delete tracked across affected records |
| **Version control** | Integer counter per record | Optimistic locking, conflict detection |
| **Retention** | Configurable (1 day to forever) | Balance storage cost vs audit needs |

### Why This Beats Alternatives

**vs WordPress (exclusive locks):**
- ✅ Multiple users can edit different parts simultaneously
- ✅ No blocking, just conflict detection
- ✅ Full undo/restore capability

**vs Notion (real-time collaboration):**
- ✅ No websocket infrastructure required
- ✅ Simpler client implementation
- ✅ Works over HTTP/REST
- ✅ Less server resources

**vs Traditional revision systems:**
- ✅ No redundant tree snapshots
- ✅ S3 offloads large storage
- ✅ Distributed = better performance
- ✅ Linked operations track complex changes

This is the sweet spot for ModulaCMS: sophisticated enough to handle real-world concurrent editing, simple enough for clients to implement, and leverages the atomic tree structure you've already built.
