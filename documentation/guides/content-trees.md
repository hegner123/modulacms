# Content Trees

Content in ModulaCMS is organized as trees. Each route has one content tree, and every piece of content is a node in that tree. A blog post page might have a root "Page" node with child nodes for a "Hero Section" and a "Cards" container, where the cards container holds individual "Card" nodes. The tree structure defines both hierarchy and order.

This guide covers how content trees work, how to manipulate them through the API, and how they are delivered to your frontend.

## Concepts

**Content tree** -- A hierarchy of content data nodes belonging to a single route. Every route has exactly one tree, rooted at a content data node whose datatype has `type = "_root"`.

**Sibling pointers** -- Content trees use a doubly-linked sibling list for ordering. Each node stores four pointers: `parent_id` (its parent node), `first_child_id` (its leftmost child), `next_sibling_id` (the next sibling in order), and `prev_sibling_id` (the previous sibling in order). This design enables O(1) navigation and reordering without renumbering.

**Root node** -- The single top-level node in a content tree. Its `parent_id` is null, and its datatype must have `type = "_root"`. Every route needs exactly one root node.

## Tree Structure

Here is a concrete example. An "About" page with a hero section and two cards:

```
Page (root)
+-- Hero Section
|   +-- Image
|   +-- Text
+-- Cards
    +-- Featured Card
    +-- Featured Card
```

In the database, this is represented as:

| Node | parent_id | first_child_id | next_sibling_id | prev_sibling_id |
|------|-----------|----------------|-----------------|-----------------|
| Page | null | Hero Section | null | null |
| Hero Section | Page | Image | Cards | null |
| Cards | Page | Featured Card 1 | null | Hero Section |
| Image | Hero Section | null | Text | null |
| Text | Hero Section | null | null | Image |
| Featured Card 1 | Cards | null | Featured Card 2 | null |
| Featured Card 2 | Cards | null | null | Featured Card 1 |

To traverse the children of any node, follow the `first_child_id` pointer and then walk the `next_sibling_id` chain. To go backwards, follow `prev_sibling_id`. To go up, follow `parent_id`.

### Why Sibling Pointers?

Traditional approaches use a `sort_order` integer column. Inserting a node in the middle requires renumbering all subsequent siblings. With sibling pointers:

- **Insert**: Update at most three nodes (the new node, its prev sibling, its next sibling).
- **Move**: Update the old neighbors and the new neighbors. No renumbering.
- **Reorder**: Walk the chain and update pointers. No renumbering.
- **Delete**: Update at most two neighbor nodes and possibly the parent's `first_child_id`.

The tradeoff is that getting all children in order requires following the linked list rather than a simple `ORDER BY sort_order`. The system handles this during tree assembly.

## Creating Content in a Tree

### Creating the Root Node

When you create a route, create a root content data node for it:

```bash
curl -X POST http://localhost:8080/api/v1/contentdata \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "route_id": "01JNRW9P2DKTZ6Q4M8W3B5J7CL",
    "datatype_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
    "status": "draft"
  }'
```

The root node has no `parent_id`. Its datatype must be a `_root`-typed datatype.

### Adding Child Nodes

Add a child node by setting `parent_id` to the parent's content data ID. The system links new nodes into the sibling chain:

```bash
curl -X POST http://localhost:8080/api/v1/contentdata \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "route_id": "01JNRW9P2DKTZ6Q4M8W3B5J7CL",
    "parent_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "datatype_id": "01JNRW6A7BMXY4K9P2Q5TH3JCR",
    "status": "draft"
  }'
```

## Moving Nodes

Move a node to a different parent (or to a different position under the same parent) with `POST /api/v1/contentdata/move`:

```bash
curl -X POST http://localhost:8080/api/v1/contentdata/move \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "node_id": "01JNRWCN5GPRZ8S6P0Y5D7L9EN",
    "new_parent_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "position": 0
  }'
```

| Field | Type | Description |
|-------|------|-------------|
| `node_id` | string | The content data ID of the node to move |
| `new_parent_id` | string or null | The new parent's content data ID. Null moves to root level. |
| `position` | integer | Zero-based position among the new parent's children. 0 = first child. |

The move operation:

1. Unlinks the node from its current sibling chain (updating its old neighbors' pointers).
2. Checks for cycles -- you cannot move a node under its own descendant.
3. Links the node into the new parent's sibling chain at the specified position.

Response:

```json
{
  "node_id": "01JNRWCN5GPRZ8S6P0Y5D7L9EN",
  "old_parent_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
  "new_parent_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM"
}
```

## Reordering Siblings

Reorder all children under a parent in a single operation with `POST /api/v1/contentdata/reorder`:

```bash
curl -X POST http://localhost:8080/api/v1/contentdata/reorder \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "ordered_ids": [
      "01JNRWDN6HQSZ9T7Q1Z6E8M0FP",
      "01JNRWCN5GPRZ8S6P0Y5D7L9EN"
    ]
  }'
```

| Field | Type | Description |
|-------|------|-------------|
| `parent_id` | string or null | The parent node whose children are being reordered |
| `ordered_ids` | array of strings | Content data IDs in the desired order. Must include all siblings under this parent. |

The system validates that every ID in `ordered_ids` belongs to the specified parent, rejects duplicates, and atomically rewrites the sibling chain to match the new order. The parent's `first_child_id` is updated to point to the first item in the list.

Response:

```json
{
  "updated": 2,
  "parent_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM"
}
```

## Bulk Tree Operations

For complex tree changes that involve creating new nodes, updating pointers, and deleting nodes in a single request, use `POST /api/v1/content/tree`:

```bash
curl -X POST http://localhost:8080/api/v1/content/tree \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "creates": [
      {
        "client_id": "temp-1",
        "datatype_id": "01JNRW6A7BMXY4K9P2Q5TH3JCR",
        "parent_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM"
      }
    ],
    "updates": [
      {
        "content_data_id": "01JNRWCN5GPRZ8S6P0Y5D7L9EN",
        "next_sibling_id": "temp-1"
      }
    ],
    "deletes": ["01JNRWDN6HQSZ9T7Q1Z6E8M0FP"]
  }'
```

The `content_id` field identifies the context node (used to resolve the route and author). The `creates` array uses `client_id` as a temporary identifier -- the server generates real ULIDs and returns the mapping. Updates and deletes can reference these temporary IDs, and the server remaps them to the server-generated IDs.

Response:

```json
{
  "created": 1,
  "updated": 1,
  "deleted": 1,
  "id_map": {
    "temp-1": "01JNRWFQ8KRUZ0V8R2A7F9N1GQ"
  }
}
```

Operations execute in order: creates first (with a two-phase approach for pointer resolution), then deletes, then updates.

## Content Delivery

When a frontend client requests content via `GET /api/v1/content/{slug}`, the system:

1. Resolves the slug to a route.
2. Fetches all content data nodes for that route.
3. Fetches all datatype definitions for those nodes.
4. Fetches all content field values for that route.
5. Fetches all field definitions for those content fields.
6. Assembles the tree: the `_root`-typed node becomes the root, children are ordered by following the `first_child_id` to `next_sibling_id` chain.
7. Fields are attached to their owning node.
8. Returns the assembled tree as JSON.

Example response (simplified):

```json
{
  "root": {
    "datatype": {
      "info": {"label": "Page", "type": "_root"},
      "content": {"content_data_id": "...", "status": "published"}
    },
    "fields": [
      {"info": {"label": "title", "type": "text"}, "content": {"field_value": "About Us"}},
      {"info": {"label": "meta_description", "type": "textarea"}, "content": {"field_value": "Learn about our company"}}
    ],
    "nodes": [
      {
        "datatype": {"info": {"label": "Hero Section", "type": "section"}},
        "fields": [
          {"info": {"label": "heading", "type": "text"}, "content": {"field_value": "Welcome"}},
          {"info": {"label": "background_image", "type": "media"}, "content": {"field_value": "hero.jpg"}}
        ],
        "nodes": []
      },
      {
        "datatype": {"info": {"label": "Cards", "type": "container"}},
        "fields": [],
        "nodes": [
          {
            "datatype": {"info": {"label": "Featured Card", "type": "card"}},
            "fields": [
              {"info": {"label": "title", "type": "text"}, "content": {"field_value": "Service 1"}}
            ],
            "nodes": []
          }
        ]
      }
    ]
  }
}
```

In the delivered JSON, children appear as a `nodes` array (ordered by the sibling chain) rather than as linked-list pointers. This is the representation your frontend consumes.

### Output Formats

The content delivery endpoint supports multiple output formats that restructure the JSON to match other CMS conventions. Set the default format in `config.json` with `output_format`, or override per-request with the `format` query parameter:

```bash
curl "http://localhost:8080/api/v1/content/about?format=clean"
```

Available formats: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `raw`. The default is `raw` (the native tree structure shown above).

## API Reference

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/content/{slug}` | Public | Deliver content tree for a slug |
| POST | `/api/v1/contentdata` | `content:create` | Create a content data node |
| GET | `/api/v1/contentdata/` | `content:read` | Get a single content data node (`?q=ID`) |
| PUT | `/api/v1/contentdata/` | `content:update` | Update a content data node |
| DELETE | `/api/v1/contentdata/` | `content:delete` | Delete a content data node (`?q=ID`) |
| POST | `/api/v1/contentdata/move` | `content:update` | Move a node to a new parent/position |
| POST | `/api/v1/contentdata/reorder` | `content:update` | Reorder sibling nodes |
| POST | `/api/v1/content/tree` | `content:update` | Bulk tree operations (create, update, delete) |
| POST | `/api/v1/content/batch` | `content:update` | Batch content updates |
| POST | `/api/v1/admin/content/heal` | `content:update` | Repair malformed tree pointers (admin) |

## Notes

- **Cycle detection.** The move operation walks the parent chain from the destination to the root, rejecting moves that would create cycles.
- **Orphan detection.** During tree assembly, nodes whose `parent_id` references a nonexistent node are tracked as orphans. The tree builder retries linking them (in case of out-of-order processing) with a retry limit to detect circular references.
- **Content field values include route_id.** The `route_id` is denormalized on content field records for query performance. Always include `route_id` when creating content fields.
- **Deleting a node** does not automatically delete its children. Update child nodes to reassign them or delete them separately. Deleting the root node of a route's tree leaves the route without content.
- **Tree composition.** When a content field of type `relation` holds a reference to another content data node, the delivery endpoint can compose the referenced node's subtree inline. The composition depth limit is configurable via `composition_max_depth` in `config.json` (default: 10).
