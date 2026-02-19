 Roles & Permissions Dashboard

> **Warning:** The SDK types referenced below may be stale. The authoritative source is the actual TypeScript types in `node_modules/@modulacms/admin-sdk/src/types/users.ts`. Always read the type files directly before implementing.

 Context

 The Role entity was updated — the legacy permissions JSON string column was removed. The SDK now has a proper junction table (RolePermission) linking roles to permissions. The Permission
 entity was also simplified to just { permission_id, label }. The current roles.tsx still references the old permissions field and uses a raw JSON textarea. It needs a full rewrite to use
 the new junction-based model with a proper permission picker UI.

 Data Model (updated SDK)

 - Role: { role_id: RoleID, label: string }
 - Permission: { permission_id: PermissionID, label: string }
 - RolePermission: { id: RolePermissionID, role_id: RoleID, permission_id: PermissionID }
 - SDK methods:
   - sdk.permissions: .list(), .get(id), .create(), .update(id, params), .remove(id) (full CrudResource)
   - sdk.rolePermissions: .list(), .get(id), .create({ role_id, permission_id }), .remove(id), .listByRole(roleId)
 - Note: The SDK exposes `listByRole(roleId)` but we intentionally use `.list()` once at page level and group client-side to avoid N+1 queries. Do not use `listByRole` in this page.

 Backend behavior: Deleting a role or permission CASCADEs through the junction table — the backend cleans up RolePermission records automatically.

 Prerequisites

 Install the Sonner toast component: `npx shadcn@latest add sonner`. Add `<Toaster />` to the root layout (`src/routes/__root.tsx`) if not already present. Import `toast` from `sonner` in mutation hooks for error display.

 Changes

 1. Query keys — src/lib/query-keys.ts

 Add two new key groups:

 permissions: {
   all: ['permissions'] as const,
   list: () => [...queryKeys.permissions.all, 'list'] as const,
 },
 rolePermissions: {
   all: ['rolePermissions'] as const,
   list: () => [...queryKeys.rolePermissions.all, 'list'] as const,
 },

 2. Query hooks — src/queries/users.ts

 Update existing role hooks — remove permissions from the CreateRoleParams/UpdateRoleParams usage (the SDK types already dropped them).

 Add new hooks:

 - usePermissions() — sdk.permissions.list()
 - useCreatePermission() — sdk.permissions.create(), invalidates permissions.all
 - useUpdatePermission() — sdk.permissions.update(), invalidates permissions.all
 - useDeletePermission() — sdk.permissions.remove(), invalidates permissions.all
 - useAllRolePermissions() — sdk.rolePermissions.list(), fetches all junction records at page level
 - useCreateRolePermission() — sdk.rolePermissions.create(), invalidates rolePermissions.all
 - useDeleteRolePermission() — sdk.rolePermissions.remove(), invalidates rolePermissions.all

 All mutation hooks include onError callback: `onError: (err) => toast.error(err.message)` (import `toast` from `sonner`).

 Import new types from SDK: Permission, PermissionID, RolePermission, RolePermissionID, CreatePermissionParams, UpdatePermissionParams, CreateRolePermissionParams. Note: RoleID is already imported in the existing code.

 3. Rewrite roles page — src/routes/_admin/users/roles.tsx

 Layout: Two sections

 Section 1 — Roles (top, primary)
 - DataTable listing all roles
 - Columns: Label, Permissions (render assigned permissions as <Badge> chips), Actions (edit/delete dropdown)
 - "Create Role" button opens dialog
 - Fetch ALL role-permissions once with useAllRolePermissions() and group by role_id client-side for badge rendering

 Section 2 — Permissions Catalog (bottom, secondary)
 - Card with simple list of all defined permissions
 - Each permission row shows label as text, an edit (pencil) icon button, and a delete icon button
 - Clicking edit toggles the row into edit mode: label becomes an Input, edit button becomes save/cancel pair. On save, call useUpdatePermission with { permission_id, label: newLabel }. On cancel, revert to display mode. Only one row can be in edit mode at a time.
 - "Add Permission" inline input + button
 - Loading spinner while permissions are fetching, EmptyState when list is empty
 - Delete button opens ConfirmDialog. If the permission is assigned to roles, the confirmation message states: "This permission is assigned to N role(s). Deleting it will remove it from all of them."
 - This section manages the available permission catalog used by the role picker

 Create/Edit Role Dialog:
 - Label input
 - Permissions section: grid of checkboxes, one per available permission
 - On create:
   1. `const role = await createRole.mutateAsync({ label })` (useCreateRole already invalidates roles.all)
   2. `const results = await Promise.allSettled(checkedPermissionIds.map(pid => sdk.rolePermissions.create({ role_id: role.role_id, permission_id: pid })))`
   3. Invalidate rolePermissions.all
   4. If any results are rejected, keep dialog open and toast which permissions failed. If all fulfilled, close dialog.
 - On edit:
   1. `await updateRole.mutateAsync({ role_id, label })` (useUpdateRole already invalidates roles.all)
   2. Get current junctions: `currentPerms = allRolePermissions.filter(rp => rp.role_id === role_id)`
   3. Compute adds: `toAdd = checkedPermissionIds.filter(pid => !currentPerms.some(rp => rp.permission_id === pid))`
   4. Compute removes: `toRemove = currentPerms.filter(rp => !checkedPermissionIds.has(rp.permission_id))`
   5. `const results = await Promise.allSettled([...toAdd.map(pid => sdk.rolePermissions.create({ role_id, permission_id: pid })), ...toRemove.map(rp => sdk.rolePermissions.remove(rp.id))])`
   6. Invalidate rolePermissions.all
   7. If any results are rejected, keep dialog open, refetch rolePermissions, toast the failures. If all fulfilled, close dialog.
 - Save button disabled while mutations are pending
 - Dialog stays open on any error — only closes on full success

 Delete Role:
 - ConfirmDialog (existing pattern)
 - Backend CASCADEs junction cleanup — no frontend pre-deletion needed

 Error handling

 - All mutations use onError to display a toast notification with the error message
 - Multi-step operations (role create/edit with permissions) use Promise.allSettled to attempt all junction mutations
 - On partial failure: dialog stays open, role-permissions are re-fetched to reflect actual state, toast shows which operations failed
 - On full success: dialog closes, relevant queries are invalidated

 Key implementation details

 - Fetch all permissions and all role-permissions once at the page level to avoid N+1
 - Group role-permissions by role_id client-side for badge rendering in the table
 - Permission picker in dialog uses local Set<PermissionID> state, synced with server on save
 - Reuse existing components: DataTable, ConfirmDialog, EmptyState, Badge, Card, Dialog, Checkbox
 - Loading and empty states for both sections (roles table and permissions catalog)

 Files modified

 ┌───────────────────────────────────┬──────────────────────────────────────────────────────────┐
 │               File                │                          Action                          │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ src/lib/query-keys.ts             │ Add permissions and rolePermissions key groups           │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ src/queries/users.ts              │ Update role hooks, add permission + rolePermission hooks │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ src/routes/_admin/users/roles.tsx │ Full rewrite                                             │
 └───────────────────────────────────┴──────────────────────────────────────────────────────────┘

 Verification

 1. npm run build — TypeScript + Vite build passes
 2. npm run dev — navigate to /users/roles
 3. Test: create a permission, create a role, assign permissions via checkboxes, verify badges in table
 4. Test: edit a role — change label, toggle permissions, save
 5. Test: delete a role, delete a permission (including one assigned to roles — verify CASCADE)
 6. Test: rename a permission, verify label updates in role badges
 7. Test: simulate partial failure (e.g., network interruption mid-save) — verify dialog stays open and shows error
