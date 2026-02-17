# API Endpoints

All endpoints are prefixed with `/api/v1`.

## Auth

| Method | Path | Handler |
|--------|------|---------|
| POST | `/api/v1/auth/login` | LoginHandler |
| POST | `/api/v1/auth/logout` | LogoutHandler |
| GET | `/api/v1/auth/me` | MeHandler |
| POST | `/api/v1/auth/register` | RegisterHandler |
| POST | `/api/v1/auth/reset` | ResetHandler |
| GET | `/api/v1/auth/oauth/login` | OauthInitiateHandler |
| GET | `/api/v1/auth/oauth/callback` | OauthCallbackHandler |

## Admin - Content Tree

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/admin/tree/{id}` | AdminTreeHandler |

## Admin - Content Data

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/admincontentdatas` | AdminContentDatasHandler |
| * | `/api/v1/admincontentdatas/{id}` | AdminContentDataHandler |

## Admin - Content Fields

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/admincontentfields` | AdminContentFieldsHandler |
| * | `/api/v1/admincontentfields/{id}` | AdminContentFieldHandler |

## Admin - Datatypes

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/admindatatypes` | AdminDatatypesHandler |
| * | `/api/v1/admindatatypes/{id}` | AdminDatatypeHandler |

## Admin - Fields

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/adminfields` | AdminFieldsHandler |
| * | `/api/v1/adminfields/{id}` | AdminFieldHandler |

## Admin - Datatype Fields

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/admindatatypefields` | AdminDatatypeFieldsHandler |
| * | `/api/v1/admindatatypefields/{id}` | AdminDatatypeFieldHandler |

## Admin - Routes

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/adminroutes` | AdminRoutesHandler |
| * | `/api/v1/adminroutes/{id}` | AdminRouteHandler |

## Admin - Config

| Method | Path | Handler |
|--------|------|---------|
| GET | `/api/v1/admin/config` | ConfigGetHandler |
| PATCH | `/api/v1/admin/config` | ConfigUpdateHandler |
| GET | `/api/v1/admin/config/meta` | ConfigMetaHandler |

## Admin - Plugins

| Method | Path | Handler |
|--------|------|---------|
| GET | `/api/v1/admin/plugins/routes` | pluginRoutesListHandler |
| POST | `/api/v1/admin/plugins/routes/approve` | pluginRoutesApproveHandler |
| POST | `/api/v1/admin/plugins/routes/revoke` | pluginRoutesRevokeHandler |

## Content Data (Public)

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/contentdata` | ContentDatasHandler |
| * | `/api/v1/contentdata/{id}` | ContentDataHandler |

## Content Fields (Public)

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/contentfields` | ContentFieldsHandler |
| * | `/api/v1/contentfields/{id}` | ContentFieldHandler |

## Content Batch

| Method | Path | Handler |
|--------|------|---------|
| POST | `/api/v1/content/batch` | ContentBatchHandler |

## Datatypes (Public)

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/datatype` | DatatypesHandler |
| GET | `/api/v1/datatype/full` | DatatypeFullHandler |
| * | `/api/v1/datatype/{id}` | DatatypeHandler |

## Datatype Fields (Public)

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/datatypefields` | DatatypeFieldsHandler |
| * | `/api/v1/datatypefields/{id}` | DatatypeFieldHandler |

## Fields (Public)

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/fields` | FieldsHandler |
| * | `/api/v1/fields/{id}` | FieldHandler |

## Media

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/media` | MediasHandler |
| GET | `/api/v1/media/health` | MediaHealthHandler |
| DELETE | `/api/v1/media/cleanup` | MediaCleanupHandler |
| * | `/api/v1/media/{id}` | MediaHandler |

## Media Dimensions

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/mediadimensions` | MediaDimensionsHandler |
| * | `/api/v1/mediadimensions/{id}` | MediaDimensionHandler |

## Routes

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/routes` | RoutesHandler |
| * | `/api/v1/routes/{id}` | RouteHandler |

## Roles

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/roles` | RolesHandler |
| * | `/api/v1/roles/{id}` | RoleHandler |

## Permissions

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/permissions` | PermissionsHandler |
| * | `/api/v1/permissions/{id}` | PermissionHandler |

## Sessions

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/sessions` | SessionsHandler |
| * | `/api/v1/sessions/{id}` | SessionHandler |

## Tables

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/tables` | TablesHandler |
| * | `/api/v1/tables/{id}` | TableHandler |

## Tokens

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/tokens` | TokensHandler |
| * | `/api/v1/tokens/{id}` | TokenHandler |

## Users

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/users` | UsersHandler |
| GET | `/api/v1/users/full` | UsersFullHandler |
| GET | `/api/v1/users/full/{id}` | UserFullHandler |
| * | `/api/v1/users/{id}` | UserHandler |

## Users OAuth

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/usersoauth` | UserOauthsHandler |
| * | `/api/v1/usersoauth/{id}` | UserOauthHandler |

## SSH Keys

| Method | Path | Handler |
|--------|------|---------|
| POST | `/api/v1/ssh-keys` | SSHKeyCreateHandler |
| GET | `/api/v1/ssh-keys` | SSHKeyListHandler |
| DELETE | `/api/v1/ssh-keys/{id}` | SSHKeyDeleteHandler |

## Import

| Method | Path | Handler |
|--------|------|---------|
| * | `/api/v1/import` | ImportBulkHandler |
| * | `/api/v1/import/contentful` | ImportContentfulHandler |
| * | `/api/v1/import/sanity` | ImportSanityHandler |
| * | `/api/v1/import/strapi` | ImportStrapiHandler |
| * | `/api/v1/import/wordpress` | ImportWordPressHandler |
| * | `/api/v1/import/clean` | ImportCleanHandler |

## Content Delivery

| Method | Path | Handler |
|--------|------|---------|
| * | `/` | SlugHandler |

## Other

| Method | Path | Handler |
|--------|------|---------|
| * | `/favicon.ico` | (returns 202) |

`*` = method routing handled internally by the handler (typically GET, POST, PUT, DELETE dispatched inside the handler function).
