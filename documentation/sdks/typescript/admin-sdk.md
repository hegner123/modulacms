# Admin SDK

Use `@modulacms/admin-sdk` for full CRUD access to every ModulaCMS API resource from admin panels, automation scripts, CI/CD pipelines, and data migration tools.

## Creating a Client

```typescript
import { createAdminClient } from '@modulacms/admin-sdk'

const client = createAdminClient({
  baseUrl: 'https://cms.example.com',
  apiKey: process.env.CMS_API_KEY,
})
```

See [Getting Started](getting-started.md) for `ClientConfig` options.

## CrudResource Pattern

Most resources on the admin client share the same generic interface:

```typescript
type CrudResource<Entity, CreateParams, UpdateParams, Id = string> = {
  list:          (opts?: RequestOptions) => Promise<Entity[]>
  get:           (id: Id, opts?: RequestOptions) => Promise<Entity>
  create:        (params: CreateParams, opts?: RequestOptions) => Promise<Entity>
  update:        (params: UpdateParams, opts?: RequestOptions) => Promise<Entity>
  remove:        (id: Id, opts?: RequestOptions) => Promise<void>
  listPaginated: (params: PaginationParams, opts?: RequestOptions) => Promise<PaginatedResponse<Entity>>
  count:         (opts?: RequestOptions) => Promise<number>
}
```

Every CRUD resource provides these seven methods. `listPaginated` returns a `PaginatedResponse<Entity>` envelope with `data`, `total`, `limit`, and `offset` fields.

### URL Patterns

The underlying HTTP calls follow a consistent pattern:

| Method | HTTP | URL |
|--------|------|-----|
| `list` | `GET` | `/api/v1/{resource}` |
| `get` | `GET` | `/api/v1/{resource}/?q={id}` |
| `create` | `POST` | `/api/v1/{resource}` |
| `update` | `PUT` | `/api/v1/{resource}/` |
| `remove` | `DELETE` | `/api/v1/{resource}/?q={id}` |
| `listPaginated` | `GET` | `/api/v1/{resource}?limit={n}&offset={n}` |

### RequestOptions

Every method accepts an optional `RequestOptions` parameter:

```typescript
type RequestOptions = {
  signal?: AbortSignal
}
```

The client merges your abort signal with its default timeout signal. Either one aborting cancels the request.

## Authentication

```typescript
// Login
const response = await client.auth.login({
  email: 'admin@example.com' as Email,
  password: 'secret',
})
// response: { user_id, email, username, created_at }

// Get current user
const me = await client.auth.me()
// me: { user_id, email, username, name, role }

// Logout
await client.auth.logout()

// Register
const user = await client.auth.register({
  username: 'newuser',
  name: 'New User',
  email: 'new@example.com' as Email,
  password: 'strongpassword',
  role: 'editor',
  date_created: new Date().toISOString(),
  date_modified: new Date().toISOString(),
})
```

## Content Data

```typescript
// Standard CRUD
const nodes = await client.contentData.list()
const node = await client.contentData.get(contentId)

// Create a node with fields in one request
const result = await client.contentData.createWithFields({
  datatype_id: datatypeId,
  parent_id: parentId,
  route_id: routeId,
  fields: { title: 'Hello World', body: '<p>Content</p>' },
})

// Reorder siblings
await client.contentData.reorder({
  parent_id: parentId,
  ordered_ids: [id1, id2, id3],
})

// Move to a new parent
await client.contentData.move({
  node_id: nodeId,
  new_parent_id: newParentId,
  position: 0,
})

// Batch update content data + field values
await client.contentData.batch({
  content_data_id: contentId,
  fields: { [fieldId]: 'updated value' },
})

// Recursive delete
const deleteResult = await client.contentData.deleteRecursive(rootId)
// deleteResult: { deleted_root, total_deleted, deleted_ids }
```

## Content Tree Save

Apply creates, deletes, and tree structure updates atomically in a single HTTP request. This is the preferred method for persisting structural changes from a block editor or tree manipulation UI.

```typescript
const result = await client.contentTree.save({
  content_id: rootContentId,
  creates: [
    {
      client_id: 'temp-1',
      datatype_id: datatypeId,
      parent_id: rootContentId,
      first_child_id: null,
      next_sibling_id: null,
      prev_sibling_id: null,
    },
  ],
  updates: [
    {
      content_data_id: existingNodeId,
      parent_id: rootContentId,
      first_child_id: 'temp-1', // references new node by client ID
      next_sibling_id: null,
      prev_sibling_id: null,
    },
  ],
  deletes: [oldNodeId],
})

// result.id_map maps client IDs to server-generated ULIDs
const serverId = result.id_map?.['temp-1']
```

## Datatypes and Fields

```typescript
// List datatypes
const datatypes = await client.datatypes.list()

// Get a fully composed datatype with field definitions
const full = await client.datatypes.getFull(datatypeId)
// full: { datatype_id, name, label, type, fields: Field[] }

// Cascade delete -- removes the datatype and all content nodes using it
const result = await client.datatypes.deleteCascade(datatypeId)
// result: { deleted_datatype_id, content_deleted, errors }

// Sort order management
await client.datatypes.updateSortOrder(datatypeId, 5)
const maxOrder = await client.datatypes.maxSortOrder(parentId)
```

## Media

```typescript
// Standard CRUD
const media = await client.media.list()
const asset = await client.media.get(mediaId)

// Upload a file
const uploaded = await client.mediaUpload.upload(file, {
  path: 'products/shoes', // optional S3 key prefix
})

// Health check -- find orphaned S3 objects
const health = await client.media.health()
// health: { total_objects, tracked_keys, orphaned_keys, orphan_count }

// Clean up orphaned files
const cleanup = await client.media.cleanup()

// Find content fields referencing a media asset
const refs = await client.media.getReferences(mediaId)

// Delete with reference cleanup
await client.media.deleteWithCleanup(mediaId)
```

## Admin Media

Admin media items are stored in a separate bucket and power the admin panel UI. The API mirrors the public media resources.

```typescript
// List and get admin media
const adminMedia = await client.adminMedia.list()
const adminAsset = await client.adminMedia.get(adminMediaId)

// Upload a file to admin media
const uploaded = await client.adminMediaUpload.upload(file)

// Update admin media metadata
await client.adminMedia.update({ admin_media_id: adminMediaId, alt: 'Logo' })

// Delete admin media
await client.adminMedia.remove(adminMediaId)
```

## Admin Media Folders

```typescript
// Get the full admin media folder tree
const tree = await client.adminMediaFolders.tree()

// List media in an admin folder
const folderMedia = await client.adminMediaFolders.listMedia(folderId, {
  limit: 20,
  offset: 0,
})

// Move admin media items to a folder (or null for root)
await client.adminMediaFolders.moveMedia({
  media_ids: [id1, id2],
  folder_id: targetFolderId,
})
```

## Users

```typescript
// List users with role labels
const users = await client.users.listFull()

// Get a fully composed user profile
const fullUser = await client.users.getFull(userId)
// Includes: oauth, ssh_keys, sessions, tokens (all safe views)

// Reassign content and delete user
const result = await client.users.reassignDelete({
  user_id: targetUserId,
  reassign_to: newOwnerId,
})
```

## Roles and Permissions

```typescript
// Manage roles
const roles = await client.roles.list()
const role = await client.roles.create({ label: 'moderator' })

// Manage permissions
const perms = await client.permissions.list()

// Role-permission associations
const assocs = await client.rolePermissions.list()
const byRole = await client.rolePermissions.listByRole(roleId)
await client.rolePermissions.create({
  role_id: roleId,
  permission_id: permId,
})
await client.rolePermissions.remove(assocId)
```

## Publishing and Versioning

```typescript
// Publish content (creates a version snapshot)
await client.publishing.publish({
  content_data_id: contentId,
  locale: 'en',
})

// Unpublish
await client.publishing.unpublish({ content_data_id: contentId })

// Schedule future publication
await client.publishing.schedule({
  content_data_id: contentId,
  publish_at: '2026-04-01T00:00:00Z',
})

// List version history
const versions = await client.publishing.listVersions(contentId)

// Restore to a previous version
await client.publishing.restore({
  content_data_id: contentId,
  content_version_id: versionId,
})

// Admin content has a separate publishing resource
await client.adminPublishing.publish({
  admin_content_data_id: adminContentId,
})
```

## Content Delivery

The admin SDK also includes content delivery for fetching rendered content trees:

```typescript
import type { Slug, ContentFormat } from '@modulacms/types'

const tree = await client.contentDelivery.getPage(
  'about' as Slug,
  'clean' as ContentFormat,
  'en',
)
```

## Admin Tree

Fetch the full admin content tree for an admin route:

```typescript
const tree = await client.adminTree.get('settings' as Slug)
// tree: { route: AdminRoute, tree: ContentTreeNode[] }

// With format conversion
const formatted = await client.adminTree.get('settings' as Slug, 'contentful')
```

## Plugins

```typescript
// List installed plugins
const plugins = await client.plugins.list()

// Get detailed plugin info
const info = await client.plugins.get('my-plugin')

// Lifecycle management
await client.plugins.reload('my-plugin')
await client.plugins.enable('my-plugin')
await client.plugins.disable('my-plugin')

// Cleanup orphaned plugin tables
const dryRun = await client.plugins.cleanupDryRun()
if (dryRun.count > 0) {
  await client.plugins.cleanupDrop({
    confirm: true,
    tables: dryRun.orphaned_tables,
  })
}

// Plugin route approval
const routes = await client.pluginRoutes.list()
await client.pluginRoutes.approve([{ plugin: 'my-plugin', method: 'GET', path: '/custom' }])

// Plugin hook approval
const hooks = await client.pluginHooks.list()
await client.pluginHooks.approve([{ plugin: 'my-plugin', event: 'after_create', table: 'content_data' }])
```

## Deploy

```typescript
// Health check
const health = await client.deploy.health()
// health: { status, version, node_id }

// Export CMS data
const payload = await client.deploy.export()

// Export specific tables
const partial = await client.deploy.export(['content_data', 'content_fields'])

// Import (dry run)
const dryResult = await client.deploy.importPayload(payload, true)

// Import (apply)
const result = await client.deploy.importPayload(payload, false)
```

## Webhooks

```typescript
// CRUD
const webhooks = await client.webhooks.list()
const wh = await client.webhooks.create({
  name: 'Deploy Notification',
  url: 'https://hooks.example.com/deploy',
  events: ['content.published', 'content.updated'],
  is_active: true,
})

// Test delivery
const testResult = await client.webhooks.test(webhookId)

// Delivery history
const deliveries = await client.webhooks.listDeliveries(webhookId)

// Retry a failed delivery
await client.webhooks.retryDelivery(deliveryId)
```

## Locales

```typescript
// CRUD
const locales = await client.locales.list()
const locale = await client.locales.create({
  code: 'fr',
  label: 'French',
  is_default: false,
  is_enabled: true,
  sort_order: 1,
})

// Create translations for a content node
const result = await client.locales.createTranslation(contentDataId, {
  locale: 'fr',
})
// result: { locale, fields_created }
```

## Content Query

```typescript
const result = await client.query.query('blog-post', {
  sort: '-published_at',
  limit: 10,
  filters: { category: 'news' },
})
```

See [Read-Only SDK](read-only-sdk.md) for `QueryParams` details.

## Configuration

```typescript
// Get current config (sensitive fields redacted)
const config = await client.config.get()

// Update config fields
const result = await client.config.update({
  port: 8080,
  cors_allowed_origins: '*',
})
if (result.restart_required?.length) {
  console.log('Restart required for:', result.restart_required)
}

// Get field metadata for building config UIs
const meta = await client.config.meta()
// meta: { fields: ConfigFieldMeta[], categories: string[] }
```

## Import

Import content from external CMS platforms:

```typescript
// Platform-specific imports
await client.import.contentful(exportData)
await client.import.sanity(exportData)
await client.import.strapi(exportData)
await client.import.wordpress(exportData)
await client.import.clean(exportData)

// Dynamic format
await client.import.bulk('contentful', exportData)
```

## Content Heal

Scan and repair structural inconsistencies in the content tree:

```typescript
// Dry run -- preview repairs without changes
const report = await client.contentHeal.heal(true)

// Apply repairs
const result = await client.contentHeal.heal(false)
// result: { dry_run, content_data_scanned, content_data_repairs, ... }
```

## Sessions and SSH Keys

```typescript
// Sessions (created via login, managed here)
await client.sessions.update(sessionParams)
await client.sessions.remove(sessionId)

// SSH keys
const keys = await client.sshKeys.list()
const key = await client.sshKeys.create({
  public_key: 'ssh-ed25519 AAAA...',
  label: 'My Laptop',
})
await client.sshKeys.remove(keyId)
```
