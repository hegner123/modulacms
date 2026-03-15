# Composed Endpoints Implementation Plan

Backend endpoints that join related tables into single responses for admin panel convenience.

## Status Key

- [ ] Not started
- [x] Complete

---

## 1. User Management

### `GET /api/v1/users/:id/full` — User profile with all related data

- [ ] Implement

**Joins:** `users` + `roles` + `user_oauth` + `user_ssh_keys` + `sessions` (active count)

**Returns:**
```json
{
  "user_id": "",
  "username": "",
  "email": "",
  "role": {
    "role_id": "",
    "label": "",
    "permissions": {}
  },
  "oauth_providers": [
    {
      "user_oauth_id": "",
      "oauth_provider": "",
      "oauth_provider_user_id": ""
    }
  ],
  "ssh_keys": [
    {
      "ssh_key_id": "",
      "label": "",
      "key_type": "",
      "fingerprint": ""
    }
  ],
  "active_session_count": 0,
  "date_created": "",
  "date_modified": ""
}
```

**Reason:** Admin user detail page needs all of this in one view.

---

### `GET /api/v1/users/full` — User list with role labels

- [ ] Implement

**Joins:** `users` + `roles`

**Returns:** Each user with their role label and permissions inlined.

**Reason:** User management table needs role names, not just role IDs.

---

## 2. Content Authoring (Public)

### `GET /api/v1/contentdata/:id/full` — Complete content entry

- [ ] Implement

**Joins:** `content_data` + `content_fields` + `fields` (definitions) + `datatypes` + `routes` + `content_relations`

**Returns:**
```json
{
  "content_data_id": "",
  "status": "",
  "route": {
    "route_id": "",
    "slug": "",
    "title": ""
  },
  "datatype": {
    "datatype_id": "",
    "label": "",
    "type": ""
  },
  "author": {
    "user_id": "",
    "username": ""
  },
  "fields": [
    {
      "content_field_id": "",
      "field_id": "",
      "field_value": "",
      "definition": {
        "label": "",
        "type": "",
        "validation": {},
        "ui_config": {},
        "data": {}
      }
    }
  ],
  "relations": [
    {
      "content_relation_id": "",
      "target_content_id": "",
      "field_id": "",
      "sort_order": 0
    }
  ],
  "date_created": "",
  "date_modified": ""
}
```

**Reason:** Content editor needs field definitions to render correct input widgets alongside current values.

---

### `GET /api/v1/contentdata/by-route/:route_id` — All content for a route with field values

- [ ] Implement

**Joins:** `content_data` + `content_fields` + `fields` + `datatypes`

**Returns:** List of content entries under a route, each with their field values and datatype info.

**Reason:** Route detail view in admin shows all content entries with summaries.

---

### `GET /api/v1/routes/:id/full` — Route with content tree

- [ ] Implement

**Joins:** `routes` + `content_data` (tree) + `datatypes` + author `users`

**Returns:**
```json
{
  "route_id": "",
  "slug": "",
  "title": "",
  "status": 0,
  "author": {
    "user_id": "",
    "username": ""
  },
  "content_tree": [
    {
      "content_data_id": "",
      "datatype_label": "",
      "status": "",
      "children": []
    }
  ],
  "date_created": "",
  "date_modified": ""
}
```

**Reason:** Route management view needs to show the content hierarchy.

---

## 3. Content Schema (Public)

### `GET /api/v1/datatype/:id/full` — Datatype with all fields

- [ ] Implement

**Joins:** `datatypes` + `datatypes_fields` + `fields` (ordered by sort_order)

**Returns:**
```json
{
  "datatype_id": "",
  "label": "",
  "type": "",
  "parent_id": null,
  "author": {
    "user_id": "",
    "username": ""
  },
  "fields": [
    {
      "field_id": "",
      "label": "",
      "type": "",
      "data": {},
      "validation": {},
      "ui_config": {},
      "sort_order": 0
    }
  ],
  "date_created": "",
  "date_modified": ""
}
```

**Reason:** Most critical composed endpoint. Admin panel needs the full schema definition to render content forms. Currently requires 3 requests (get datatype, list datatype_fields, get each field).

---

### `GET /api/v1/datatype/full` — All datatypes with field counts and parent info

- [ ] Implement

**Joins:** `datatypes` + `datatypes_fields` (count) + `datatypes` (parent label)

**Returns:** Each datatype with field count and parent label (not just parent_id).

**Reason:** Datatype list view needs to show hierarchy and complexity at a glance.

---

## 4. Admin CMS Structure

### `GET /api/v1/admindatatypes/:id/full` — Admin datatype with fields

- [ ] Implement

**Joins:** `admin_datatypes` + `admin_datatypes_fields` + `admin_fields`

Same pattern as public `datatype/:id/full`.

---

### `GET /api/v1/admincontentdatas/:id/full` — Admin content entry with fields

- [ ] Implement

**Joins:** `admin_content_data` + `admin_content_fields` + `admin_fields` + `admin_datatypes`

Same pattern as public `contentdata/:id/full`.

---

**Note:** `GET /api/v1/admin/tree/{slug}` already exists and covers the admin tree-level composite view.

---

## 5. Media

### `GET /api/v1/media/:id/full` — Media with author and usage

- [ ] Implement

**Joins:** `media` + `users` (author) + `content_fields` (where field_value references this media_id)

**Returns:**
```json
{
  "media_id": "",
  "name": "",
  "display_name": "",
  "alt": "",
  "caption": "",
  "url": "",
  "mimetype": "",
  "dimensions": "",
  "srcset": "",
  "author": {
    "user_id": "",
    "username": ""
  },
  "used_by": [
    {
      "content_data_id": "",
      "field_id": "",
      "content_field_id": ""
    }
  ],
  "date_created": "",
  "date_modified": ""
}
```

**Reason:** Media detail view needs to show where an asset is used before allowing deletion.

---

### `GET /api/v1/media/full` — Media list with author names

- [ ] Implement

**Joins:** `media` + `users` (author name)

**Returns:** Media list with author names instead of raw user IDs.

**Reason:** Media library grid needs to show "uploaded by" without N+1 requests.

---

## 6. Audit / Activity

### `GET /api/v1/activity/recent` — Recent changes with user info

- [ ] Implement

**Joins:** `change_events` + `users` (actor name)

**Returns:**
```json
[
  {
    "event_id": "",
    "table_name": "",
    "row_id": "",
    "operation": "",
    "actor": {
      "user_id": "",
      "username": ""
    },
    "old_data": {},
    "new_data": {},
    "timestamp": ""
  }
]
```

**Reason:** Admin dashboard activity feed.

---

## 7. Sessions (Admin)

### `GET /api/v1/users/:id/sessions` — Active sessions for a user

- [ ] Implement

**Joins:** `sessions` (filtered by user_id)

**Returns:** All active sessions with device/IP info.

**Reason:** User management needs "force logout" and session visibility.

---

## Implementation Priority

Implement in this order for maximum incremental value:

| Priority | Endpoint | Unlocks |
|----------|----------|---------|
| 1 | `GET /api/v1/datatype/:id/full` | Form rendering |
| 2 | `GET /api/v1/contentdata/:id/full` | Content editing |
| 3 | `GET /api/v1/users/:id/full` | User management |
| 4 | `GET /api/v1/media/full` | Media library |
| 5 | `GET /api/v1/routes/:id/full` | Route management |
| 6 | `GET /api/v1/activity/recent` | Dashboard |
| 7 | Admin equivalents | Mirror the public patterns |
| 8 | Remaining list endpoints | Polish |

## Implementation Notes

- All composed endpoints are **read-only** (GET). Write operations stay on individual entity endpoints and the existing batch endpoint.
- For SQL: prefer writing new sqlc queries with JOINs over doing multiple queries in Go. Let the database engine optimize the join plan.
- Consider adding `?fields=minimal` query parameter support for list endpoints to allow the SDK to request only IDs + labels when full data isn't needed.
- Auth middleware applies to all of these the same as existing endpoints.
- The existing `POST /api/v1/content/batch` covers the write side well. These composed read endpoints are the missing counterpart.
