# Tree Structure

ModulaCMS stores content as trees rather than flat records. Each route has one content tree, and every content data node occupies a position in that tree. This design supports component-based page composition: a page is not a blob of text with a title, but a hierarchy of typed sections, containers, cards, and other building blocks.

This document covers the data model behind content trees. For API usage and code examples, see the [Content Trees](../guides/content-trees.md) guide.

## Why Trees

Traditional headless CMS platforms model content as flat collections of typed entries (a "blog post" has a title, body, and image). This works for simple use cases but forces frontend developers to either (a) build page structure outside the CMS, or (b) use nested JSON fields that the CMS cannot validate or query.

ModulaCMS makes structure a first-class concept. A page is a tree of typed content nodes, each with its own datatype schema and field values. The frontend receives the assembled tree and renders each node according to its type. The CMS validates structure, enforces schemas, and delivers fully composed pages without the frontend needing to assemble anything.

## The Four Pointers

Each content data node stores four pointer fields that define its position in the tree:

| Pointer | Type | Description |
|---------|------|-------------|
| `parent_id` | `*ContentID` | The node's parent. Null for root nodes. |
| `first_child_id` | `*string` | The leftmost child of this node. Null if the node has no children. |
| `next_sibling_id` | `*string` | The next sibling in display order. Null for the last sibling. |
| `prev_sibling_id` | `*string` | The previous sibling in display order. Null for the first sibling. |

Children of a node form a doubly-linked list through `next_sibling_id` and `prev_sibling_id`. The parent holds a pointer to the head of this list via `first_child_id`.

### Example

Consider this page structure:

```
Page (root)
+-- Hero
|   +-- Heading
|   +-- Image
+-- Cards Container
    +-- Card A
    +-- Card B
```

The pointer values:

| Node | parent_id | first_child_id | next_sibling_id | prev_sibling_id |
|------|-----------|----------------|-----------------|-----------------|
| Page | null | Hero | null | null |
| Hero | Page | Heading | Cards Container | null |
| Cards Container | Page | Card A | null | Hero |
| Heading | Hero | null | Image | null |
| Image | Hero | null | null | Heading |
| Card A | Cards Container | null | Card B | null |
| Card B | Cards Container | null | null | Card A |

## O(1) Operations

The sibling-pointer design gives constant-time insert, move, and delete operations. Compare this to the common alternative of a `sort_order` integer column, where inserting in the middle requires renumbering every subsequent sibling.

### Insert

Inserting a new node as the Nth child of a parent updates at most three nodes:

1. The new node (set its `parent_id`, `prev_sibling_id`, `next_sibling_id`).
2. The node previously at position N (update its `prev_sibling_id` to point to the new node).
3. The node previously at position N-1, or the parent's `first_child_id` if inserting at position 0.

No other nodes in the tree are touched.

### Move

Moving a node to a different parent (or a different position under the same parent):

1. Unlink the node from its current position: update the old previous sibling's `next_sibling_id`, the old next sibling's `prev_sibling_id`, and possibly the old parent's `first_child_id`.
2. Link the node into the new position: update the new neighbors and possibly the new parent's `first_child_id`.

This touches at most five nodes regardless of tree size.

### Delete

Deleting a node updates at most two neighbors (the previous and next siblings) and possibly the parent's `first_child_id`. Deleting a node does not automatically delete its children -- you must reassign or delete them separately.

### Reorder

Reordering all children of a parent rewrites the sibling chain for those children. This is O(k) where k is the number of children being reordered, not O(n) where n is the total tree size.

## Tree Assembly Algorithm

When content is delivered to a frontend client, the server assembles flat database rows into a nested tree:

1. **Fetch all content data nodes** for the route.
2. **Identify the root node** -- the node whose datatype has `type = "_root"` and whose `parent_id` is null.
3. **Build a lookup map** from content data ID to node.
4. **For each node**, follow `first_child_id` to get the first child, then walk the `next_sibling_id` chain to collect all children in order. Attach them as a `nodes` array on the parent.
5. **Attach field values** to each node by matching `content_data_id`.
6. **Return the root** with its recursively assembled children.

The delivered JSON uses nested `nodes` arrays rather than pointer fields:

```json
{
  "root": {
    "datatype": {"info": {"label": "Page", "type": "_root"}},
    "fields": [...],
    "nodes": [
      {
        "datatype": {"info": {"label": "Hero", "type": "section"}},
        "fields": [...],
        "nodes": [...]
      }
    ]
  }
}
```

## Root Node Invariant

Every content tree has exactly one root node. The root node's datatype must have `type = "_root"`, and its `parent_id` must be null. A route without a root node has no content. Creating a route's content begins by creating a root node.

The system enforces that only one `_root`-typed node exists per route. Attempting to create a second root node under the same route fails.

## Orphan Detection

During tree assembly, a node whose `parent_id` references a nonexistent node is an orphan. The tree builder handles orphans with a retry strategy:

1. First pass: attempt to attach each node to its parent.
2. Nodes that failed (parent not yet processed) are queued for retry.
3. Retry up to a fixed limit to handle out-of-order processing.
4. Nodes that still cannot be attached after all retries are flagged as orphans.

Orphans can result from bugs in tree operations, incomplete deletes, or data corruption. The admin heal endpoint (`POST /api/v1/admin/content/heal`) can repair malformed tree pointers.

## Cycle Prevention

The move operation prevents cycles by walking the parent chain from the destination node up to the root before allowing the move. If the node being moved appears anywhere in the destination's ancestor chain, the move is rejected.

This check is O(d) where d is the depth of the tree at the destination point -- typically small for CMS content trees.

## Composition via Relations

Content nodes can reference other content nodes through `_id`-type fields. At delivery time, the referenced node's subtree can be composed inline. This enables shared content: a "Featured Author" component referenced by multiple blog posts is defined once and composed into each post's delivery response.

Referenced datatypes are resolved from their own published snapshots, giving both predictability (the reference is a specific published version) and shared-content flexibility (update the reference once, and all consumers see the change on next publish).

The composition depth limit prevents infinite recursion. Configure it via `composition_max_depth` in `modula.config.json` (default: 10).

## Tradeoffs

The sibling-pointer model has one significant tradeoff: getting all children of a node in order requires following the linked list rather than a simple `ORDER BY sort_order` query. The system handles this during tree assembly, so the tradeoff is internal to the delivery pipeline and invisible to API consumers.

The benefit is that structural mutations (insert, move, delete, reorder) are constant-time operations that touch a bounded number of rows, regardless of how many siblings exist. For CMS content where pages are frequently restructured, this is the right tradeoff.
