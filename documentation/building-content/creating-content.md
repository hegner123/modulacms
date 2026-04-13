# Creating Content

Create content nodes, organize them into hierarchical trees, and control their order and position.

## Content is hierarchical

ModulaCMS organizes content as trees. A page isn't a flat collection of fields -- it's a hierarchy of typed content nodes, each with its own datatype and field values. Nest pages under pages, sections under pages, cards under sections. Your frontend receives a ready-to-render tree.

```
Page (root)
+-- Hero Section
|   +-- Heading
|   +-- Background Image
+-- Cards Container
    +-- Card A
    +-- Card B
```

Every tree starts with a single root node. You add child nodes under it, and children under those, to any depth.

## Two kinds of trees

ModulaCMS supports two kinds of content trees:

| Tree Type | Root Datatype | Accessed Via | Use Case |
|-----------|---------------|-------------|----------|
| **Route-based** | `_root` | Content delivery by slug (`/api/v1/content/{slug}`) | Page content tied to a URL |
| **Global** | `_global` | `/api/v1/globals` endpoint | Site-wide content (menus, footers, settings) not tied to a route |

Each route has exactly one content tree. Multiple independent global trees can coexist.

## Create the root node

Start a content tree by creating its root node. The root node has no parent and uses a datatype with `type = "_root"` (for route-based content) or `type = "_global"` (for global content).

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

Response:

```json
{
  "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
  "route_id": "01JNRW9P2DKTZ6Q4M8W3B5J7CL",
  "datatype_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
  "status": "draft"
}
```

## Add child nodes

Add a child by setting `parent_id` to the parent's content data ID. ModulaCMS positions new children at the end of the parent's child list automatically.

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

## Set field values

Once you've created a content node, populate its fields by creating content field records:

```bash
curl -X POST http://localhost:8080/api/v1/contentfields \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "field_id": "01JNRW7K8CNQZ5P3R9W6TJ4MAS",
    "field_value": "My First Blog Post",
    "route_id": "01JNRW9P2DKTZ6Q4M8W3B5J7CL"
  }'
```

> **Good to know**: Always include `route_id` when creating content fields. ModulaCMS uses it for query performance.

## Move nodes

Move a content node to a different parent or a different position under the same parent:

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

| Field | Description |
|-------|-------------|
| `node_id` | The content node to move |
| `new_parent_id` | The new parent's content data ID |
| `position` | Zero-based position among the new parent's children (0 = first child) |

ModulaCMS prevents circular moves -- you can't move a node under its own descendant.

Response:

```json
{
  "node_id": "01JNRWCN5GPRZ8S6P0Y5D7L9EN",
  "old_parent_id": "01JNRW5V6QNPZ3R8W4T2YH9B0D",
  "new_parent_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM"
}
```

## Reorder siblings

Reorder all children under a parent in a single operation by specifying the desired sequence:

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

The `ordered_ids` array must include all children of the specified parent. ModulaCMS validates that every ID belongs to the parent and rejects duplicates.

Response:

```json
{
  "updated": 2,
  "parent_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM"
}
```

> **Good to know**: Move and reorder operations update only the affected nodes and their immediate neighbors, making them fast even for large trees.

## Bulk tree operations

For complex changes that involve creating, updating, and deleting nodes in a single request, use the bulk tree endpoint:

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

Use `client_id` as a temporary identifier in the `creates` array. The server generates real IDs and returns the mapping so you can reference newly created nodes in updates.

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

## Delete nodes

Delete a content node with a DELETE request:

```bash
curl -X DELETE "http://localhost:8080/api/v1/contentdata/?q=01JNRWCN5GPRZ8S6P0Y5D7L9EN" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

> **Good to know**: Deleting a node does not automatically delete its children. Reassign or delete child nodes separately before removing a parent.

## Content delivery

When you [publish content](/docs/building-content/publishing), ModulaCMS assembles the tree into a snapshot. Frontend clients requesting content by slug receive this snapshot as nested JSON:

```json
{
  "root": {
    "datatype": {"info": {"label": "Page", "type": "_root"}},
    "fields": [
      {"info": {"label": "title", "type": "text"}, "content": {"field_value": "About Us"}}
    ],
    "nodes": [
      {
        "datatype": {"info": {"label": "Hero Section", "type": "section"}},
        "fields": [
          {"info": {"label": "heading", "type": "text"}, "content": {"field_value": "Welcome"}}
        ],
        "nodes": []
      }
    ]
  }
}
```

Children appear as an ordered `nodes` array. Your frontend consumes this structure directly -- no assembly required.

### Output formats

The content delivery endpoint supports multiple output formats that restructure the JSON to match other CMS conventions. Set the default in `modula.config.json` with `output_format`, or override per request:

```bash
curl "http://localhost:8080/api/v1/content/about?format=clean"
```

Available formats: `contentful`, `sanity`, `strapi`, `wordpress`, `clean`, `raw`. The default is `raw` (the native tree structure).

### Preview mode

Preview unpublished changes through the delivery endpoint by adding `?preview=true`. This builds the tree from live data instead of the published snapshot, so editors can see draft content through the client delivery path.

## Heal malformed trees

If tree structure becomes inconsistent (from bugs, interrupted operations, or data issues), repair it with the heal endpoint:

```bash
curl -X POST http://localhost:8080/api/v1/admin/content/heal \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

## API reference

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/content/{slug}` | Public | Deliver content tree for a slug |
| GET | `/api/v1/globals` | Public | Deliver all published global trees |
| POST | `/api/v1/contentdata` | `content:create` | Create a content node |
| GET | `/api/v1/contentdata/` | `content:read` | Get a content node (`?q=ID`) |
| PUT | `/api/v1/contentdata/` | `content:update` | Update a content node |
| DELETE | `/api/v1/contentdata/` | `content:delete` | Delete a content node (`?q=ID`) |
| POST | `/api/v1/contentdata/move` | `content:update` | Move a node to a new parent/position |
| POST | `/api/v1/contentdata/reorder` | `content:update` | Reorder sibling nodes |
| POST | `/api/v1/content/tree` | `content:update` | Bulk tree operations (create, update, delete) |
| POST | `/api/v1/contentfields` | `content:create` | Create a content field value |
| GET | `/api/v1/contentfields/` | `content:read` | Get a content field (`?q=ID`) |
| PUT | `/api/v1/contentfields/` | `content:update` | Update a content field value |
| DELETE | `/api/v1/contentfields/` | `content:delete` | Delete a content field value (`?q=ID`) |
| POST | `/api/v1/admin/content/heal` | `content:update` | Repair malformed tree structure |

## Next steps

Your content is organized and populated. [Publish it to make it live](/docs/building-content/publishing).
