# router

HTTP routing package using standard library net/http.ServeMux for ModulaCMS REST API.

## Overview

Package router provides HTTP handlers for the ModulaCMS REST API. Uses standard library ServeMux for routing. All handlers follow RESTful conventions with JSON request and response bodies. Handlers delegate to db package for persistence and use middleware for authentication, CORS, and audit logging.

## Mux Configuration

### NewModulacmsMux

Creates and configures HTTP router with all endpoint registrations. Returns configured ServeMux ready for use by HTTP server.

Parameters: config.Config containing CORS origins, rate limits, OAuth settings.

Returns: Configured http.ServeMux with all routes registered.

Rate limiting applied to auth endpoints at 10 requests per minute per IP. CORS middleware wraps auth and OAuth endpoints. Most endpoints use pattern matching with method prefixes like GET, POST, PUT, DELETE.

## Authentication Handlers

### LoginHandler

Handles password-based authentication via POST to /api/v1/auth/login. Validates email and password credentials, creates database session, sets secure HTTP-only cookie.

Request body: JSON with email and password fields. Returns JSON with user_id, email, username, created_at on success. Returns 401 Unauthorized for invalid credentials. Does not reveal whether user exists to prevent enumeration attacks.

### LogoutHandler

Handles session termination via POST to /api/v1/auth/logout. Clears session cookie by setting MaxAge to negative one. Returns success message even if no active session exists.

### MeHandler

Returns currently authenticated user information via GET to /api/v1/auth/me. Validates session cookie and returns user details without sensitive data.

Returns JSON with user_id, email, username, name, role. Returns 401 Unauthorized if session invalid or missing.

### RegisterHandler

Handles user registration via POST to /api/v1/auth/register. Delegates to ApiCreateUser for creation logic.

### ResetPasswordHandler

Handles password reset via POST to /api/v1/auth/reset. Delegates to ApiUpdateUser for update logic.

## OAuth Handlers

### OauthInitiateHandler

Starts OAuth flow with PKCE and state parameter for CSRF protection. Accessible via GET /api/v1/auth/oauth/login.

Generates cryptographically secure state parameter, creates PKCE code verifier, stores verifier associated with state, builds OAuth2 config from application config, generates authorization URL with state and PKCE challenge.

Redirects user to OAuth provider authorization endpoint. Returns 500 Internal Server Error if OAuth not configured or state generation fails.

### OauthCallbackHandler

Handles OAuth provider redirect with state validation and PKCE code exchange. Accessible via GET /api/v1/auth/oauth/callback.

Retrieves authorization code from query parameter, validates state parameter for CSRF protection, retrieves PKCE verifier associated with state, exchanges code for access token using verifier, fetches user info from provider, provisions user in database, creates session and sets cookie.

Redirects to success URL on completion. Returns 400 Bad Request if code or state missing or invalid. Returns 500 Internal Server Error if token exchange or user provisioning fails.

## User Handlers

### UsersHandler

Collection endpoint at /api/v1/users supporting GET to list all users and POST to create new user.

### UserHandler

Individual resource endpoint at /api/v1/users/ supporting GET, PUT, DELETE for specific user identified by query parameter q containing UserID.

### ApiGetUser

Retrieves single user by ID from query parameter q. Returns JSON representation of user record. Returns 400 Bad Request if ID invalid, 500 Internal Server Error if database error.

### ApiListUsers

Returns JSON array of all users. Returns 500 Internal Server Error if database error.

### ApiCreateUser

Creates new user from JSON request body. Expects CreateUserParams struct. Returns created user with 201 Created status. Returns 400 Bad Request if JSON invalid, 500 Internal Server Error if creation fails.

### ApiUpdateUser

Updates existing user from JSON request body. Expects UpdateUserParams struct with user_id field. Returns updated user with 200 OK status. Returns 400 Bad Request if JSON invalid, 500 Internal Server Error if update fails.

### ApiDeleteUser

Deletes user identified by query parameter q. Returns 200 OK with empty body on success. Returns 400 Bad Request if ID invalid, 500 Internal Server Error if deletion fails.

## Role Handlers

### RolesHandler

Collection endpoint at /api/v1/roles supporting GET to list roles and POST to create role.

### RoleHandler

Individual resource endpoint at /api/v1/roles/ supporting GET, PUT, DELETE for specific role identified by query parameter q containing RoleID.

Handler functions follow same pattern as user handlers: apiGetRole, apiListRoles, apiCreateRole, apiUpdateRole, apiDeleteRole.

## Session Handlers

### SessionsHandler

Collection endpoint at /api/v1/sessions. GET and POST return 405 Method Not Allowed as sessions created through login endpoint.

### SessionHandler

Individual resource endpoint at /api/v1/sessions/ supporting PUT to update session and DELETE to remove session. Identified by query parameter q containing SessionID.

## Token Handlers

### TokensHandler

Collection endpoint at /api/v1/tokens supporting GET to list tokens and POST to create token. GET returns 405 Method Not Allowed.

### TokenHandler

Individual resource endpoint at /api/v1/tokens/ supporting GET, PUT, DELETE for specific token identified by query parameter q containing token ID string.

Token IDs validated as strings not ULID types. Missing token ID returns 400 Bad Request with error message.

## Table Handlers

### TablesHandler

Collection endpoint at /api/v1/tables supporting GET to list all tables. POST returns 405 Method Not Allowed.

### TableHandler

Individual resource endpoint at /api/v1/tables/ supporting GET, PUT, DELETE for specific table identified by query parameter q containing table ID string.

## SSH Key Handlers

### AddSSHKeyHandler

Handles POST /api/v1/ssh-keys to add SSH public key to authenticated user account. Requires authentication context.

Parses request body expecting public_key and optional label fields. Validates SSH key format and extracts key type and fingerprint. Checks for duplicate keys by fingerprint. Creates user_ssh_keys record in database.

Returns created SSH key record with 201 Created. Returns 401 Unauthorized if not authenticated, 400 Bad Request if key invalid, 409 Conflict if key already exists.

### ListSSHKeysHandler

Handles GET /api/v1/ssh-keys to return all SSH keys for authenticated user. Returns summary without full public key for security.

Response includes ssh_key_id, key_type, fingerprint, label, date_created, last_used. Returns 401 Unauthorized if not authenticated.

### DeleteSSHKeyHandler

Handles DELETE /api/v1/ssh-keys/:id to remove SSH key. Verifies key belongs to authenticated user before deletion.

Extracts key ID from URL path. Fetches key to verify ownership. Returns 403 Forbidden if key belongs to different user, 404 Not Found if key does not exist.

## Content Batch Handler

### ContentBatchHandler

Handles POST /api/v1/content/batch for atomic content updates. Applies optional content_data update plus map of field value upserts in single request.

Request body expects BatchContentUpdateRequest with content_data_id, optional content_data params, optional fields map. At least one of content_data or fields must be present.

Returns BatchContentUpdateResponse with content_data_id, content_data_updated boolean, fields_updated count, fields_created count, fields_failed count, errors array. Always returns 200 OK even on partial failures.

For field updates, fetches existing content fields, builds map by field_id, performs upsert for each field in request. Creates new content field if not exists, updates existing field otherwise. Derives author_id from authenticated user in context.

## Pagination Helpers

### ParsePaginationParams

Extracts limit and offset from query parameters. Returns PaginationParams struct with defaults.

Default limit fifty, default offset zero. Maximum limit one thousand. Negative values ignored. Parses query parameters limit and offset as int64.

### HasPaginationParams

Returns true if request contains limit or offset query parameters. Used to conditionally enable pagination in collection handlers.

## Admin Content Handlers

### AdminContentDatasHandler

Collection endpoint at /api/v1/admincontentdatas supporting GET with optional pagination and POST to create.

Returns full list or paginated response based on query parameters. Uses HasPaginationParams to detect pagination request.

### AdminContentDataHandler

Individual resource endpoint at /api/v1/admincontentdatas/ supporting POST, PUT, DELETE for specific item identified by query parameter q containing AdminContentID.

PUT handler fetches old data, marshals for audit trail, updates with new params, sets date_modified to current timestamp.

### AdminContentFieldsHandler

Collection endpoint at /api/v1/admincontentfields supporting GET with optional pagination and POST to create.

### AdminContentFieldHandler

Individual resource endpoint at /api/v1/admincontentfields/ supporting GET, PUT, DELETE for specific item identified by query parameter q containing AdminContentFieldID.

### AdminDatatypesHandler

Collection endpoint at /api/v1/admindatatypes supporting GET with optional pagination and POST to create.

### AdminDatatypeHandler

Individual resource endpoint at /api/v1/admindatatypes/ supporting GET, PUT, DELETE for specific item identified by query parameter q containing AdminDatatypeID.

### AdminFieldsHandler

Collection endpoint at /api/v1/adminfields supporting GET with optional pagination and POST to create.

Create and update handlers set Validation and UIConfig to EmptyJSON if not provided in request.

### AdminFieldHandler

Individual resource endpoint at /api/v1/adminfields/ supporting GET, PUT, DELETE for specific item identified by query parameter q containing AdminFieldID.

### AdminDatatypeFieldsHandler

Collection endpoint at /api/v1/admindatatypefields supporting GET with optional pagination, POST to create, DELETE to remove.

GET supports query filters: admin_datatype_id to filter by datatype, admin_field_id to filter by field, or no filter for all records.

### AdminDatatypeFieldHandler

Individual resource endpoint at /api/v1/admindatatypefields/ supporting PUT to update. Identified by composite key in request body.

### AdminRoutesHandler

Collection endpoint at /api/v1/adminroutes supporting GET with optional pagination or ordering, POST to create.

GET with query parameter ordered=true returns routes sorted by Order field on root admin content node. Routes without Order value appear last. Ordering logic finds Order admin field by label, fetches admin content data for each route, finds root node with null parent, searches content fields for Order field value, parses as integer for sorting.

### AdminRouteHandler

Individual resource endpoint at /api/v1/adminroutes/ supporting GET by slug, PUT, DELETE by ID.

GET uses query parameter q containing Slug not ID. DELETE uses query parameter q containing AdminRouteID.

## Content Data Handlers

### ContentDatasHandler

Collection endpoint at /api/v1/contentdata supporting GET with optional pagination and POST to create.

### ContentDataHandler

Individual resource endpoint at /api/v1/contentdata/ supporting GET, POST, PUT, DELETE for specific item identified by query parameter q containing ContentID.

### ContentFieldsHandler

Collection endpoint at /api/v1/contentfields supporting GET with optional pagination and POST to create.

### ContentFieldHandler

Individual resource endpoint at /api/v1/contentfields/ supporting GET, PUT, DELETE for specific item identified by query parameter q containing ContentFieldID.

## Datatype Handlers

### DatatypesHandler

Collection endpoint at /api/v1/datatype supporting GET with optional pagination, POST to create, DELETE to remove.

Create handler generates new DatatypeID if not provided, sets date_created and date_modified to current UTC timestamp if not provided.

### DatatypeHandler

Individual resource endpoint at /api/v1/datatype/ supporting GET and PUT for specific item identified by query parameter q containing DatatypeID.

### DatatypeFieldsHandler

Collection endpoint at /api/v1/datatypefields supporting GET with optional pagination, POST to create, DELETE to remove.

GET supports query filters: datatype_id to filter by datatype, field_id to filter by field, or no filter for all records.

Create handler generates new DatatypeFieldID if ID field empty in request.

### DatatypeFieldHandler

Individual resource endpoint at /api/v1/datatypefields/ supporting PUT to update. Identified by composite key in request body.

## Field Handlers

### FieldsHandler

Collection endpoint at /api/v1/fields supporting GET with optional pagination and POST to create.

Create handler generates new FieldID if not provided, sets date_created and date_modified to current UTC timestamp if not provided, sets Validation and UIConfig to EmptyJSON if not provided.

### FieldHandler

Individual resource endpoint at /api/v1/fields/ supporting GET, PUT, DELETE for specific item identified by query parameter q containing FieldID.

Update handler sets Validation and UIConfig to EmptyJSON if empty strings in request.

## Media Handlers

### MediasHandler

Collection endpoint at /api/v1/media supporting GET with optional pagination and POST to upload media file.

GET checks HasPaginationParams to route to full list or paginated response.

### MediaHandler

Individual resource endpoint at /api/v1/media/ supporting GET, PUT, DELETE for specific item identified by query parameter q containing MediaID.

### apiCreateMedia

Handles multipart file upload with maximum size limit from media package. Parses multipart form, extracts file field, delegates to media.ProcessMediaUpload for validation and pipeline execution.

Pipeline function calls media.HandleMediaUpload to process uploaded file. Returns created Media record with 201 Created on success.

Returns 400 Bad Request if form invalid or file too large. Returns 409 Conflict if duplicate media by hash. Returns 400 Bad Request if invalid MIME type. Returns 500 Internal Server Error for other errors.

Uses typed errors from media package: DuplicateMediaError, InvalidMediaTypeError, FileTooLargeError.

### apiListMediaPaginated

Returns paginated media list with total count. Response uses PaginatedResponse generic struct with Data array, Total count, Limit, Offset fields.

## Media Dimension Handlers

### MediaDimensionsHandler

Collection endpoint at /api/v1/mediadimensions supporting GET to list and POST to create.

### MediaDimensionHandler

Individual resource endpoint at /api/v1/mediadimensions/ supporting GET, PUT, DELETE for specific item identified by query parameter q containing media dimension ID string.

Media dimension IDs validated as strings not ULID types.

## Route Handlers

### RoutesHandler

Collection endpoint at /api/v1/routes supporting GET with optional pagination and POST to create.

Create handler generates new RouteID if not provided, sets date_created and date_modified to current UTC timestamp if not provided.

### RouteHandler

Individual resource endpoint at /api/v1/routes/ supporting GET, PUT, DELETE for specific item identified by query parameter q containing RouteID.

## User OAuth Handlers

### UserOauthsHandler

Collection endpoint at /api/v1/usersoauth supporting POST to create OAuth connection. GET returns 405 Method Not Allowed.

### UserOauthHandler

Individual resource endpoint at /api/v1/usersoauth/ supporting PUT and DELETE for specific item identified by query parameter q containing UserOauthID.

GET returns 405 Method Not Allowed. OAuth connections retrieved through user endpoint.

## Slug Content Handler

### SlugHandler

Catch-all handler registered at / for dynamic content routing by slug. Only supports GET method.

Fetches route by URL path, lists content data by route, builds datatypes and fields, constructs content tree using model.BuildTree, transforms output based on configured or query parameter format.

Supports format query parameter to override config.Output_Format. Valid formats: contentful, sanity, strapi, wordpress, clean, raw. Invalid format returns 400 Bad Request with error message listing valid options.

Uses transform.TransformConfig to create transformer for specified format. Calls TransformAndWrite to serialize and send response.

## Admin Tree Handler

### AdminTreeHandler

Handles GET /api/v1/admin/tree/:slug for admin content trees by slug. Similar to SlugHandler but operates on admin tables.

Extracts slug from URL path after prefix. Fetches admin route by slug. Lists admin content data by route. Filters to entries with valid admin_datatype_id and fetches corresponding datatypes. Lists admin content fields by route. Filters to entries with valid admin_field_id and fetches corresponding fields.

Calls model.BuildAdminTree with filtered parallel arrays of content data, datatypes, content fields, fields. Returns 404 Not Found if slug does not match admin route.

Supports format query parameter to override config.Output_Format. Same format validation as SlugHandler.

## Import Handlers

### ImportContentfulHandler

Handles POST /api/v1/import/contentful to import Contentful format data. Delegates to apiImportContent with FormatContentful.

### ImportSanityHandler

Handles POST /api/v1/import/sanity to import Sanity format data. Delegates to apiImportContent with FormatSanity.

### ImportStrapiHandler

Handles POST /api/v1/import/strapi to import Strapi format data. Delegates to apiImportContent with FormatStrapi.

### ImportWordPressHandler

Handles POST /api/v1/import/wordpress to import WordPress format data. Delegates to apiImportContent with FormatWordPress.

### ImportCleanHandler

Handles POST /api/v1/import/clean to import Clean ModulaCMS format data. Delegates to apiImportContent with FormatClean.

### ImportBulkHandler

Handles POST /api/v1/import with format query parameter. Validates format parameter against valid options. Returns 400 Bad Request if format missing or invalid.

### apiImportContent

Core import logic shared by all import handlers. Reads request body, creates transformer for specified format, parses CMS format to ModulaCMS Root structure, imports to database via importRootToDatabase.

Accepts optional route_id query parameter as NullableRouteID. Returns ImportResult JSON with success boolean, counts of created entities, errors array, message string.

Returns 400 Bad Request if body read fails or parse fails. Returns 500 Internal Server Error if database import fails.

### ImportResult

Response structure for import operations. Fields: Success boolean, DatatypesCreated int, FieldsCreated int, ContentCreated int, Errors string array, Message string.

### importRootToDatabase

Imports parsed Root structure to database. Creates importContext to track state during recursive import. Calls importNode on root node with null parent. Sets success to true if no errors, constructs summary message.

Returns early with errors if Root.Node is null. Logs import completion with entity counts and error count.

### importNode

Recursively imports single node and children into database. Returns ContentID of created content_data row or empty if creation failed.

Finds or creates datatype using findOrCreateDatatype. Creates content_data with null sibling pointers. Creates fields and content_fields for node. Recurses into children with current content_data as parent. Patches sibling pointers after all children created.

Increments result.ContentCreated on success. Appends errors to result.Errors on failure.

### findOrCreateDatatype

Returns existing or newly-created DatatypeID for node type. De-duplicates by cache key combining label and type string. Cache key format: label pipe type.

Creates datatype with null parent_id. Caches created ID. Increments result.DatatypesCreated on creation. Returns empty DatatypeID on error.

### createFieldAndContentField

Creates field definition, content_field linking to content_data, datatype_field linking to datatype. Single field creation creates three database records.

Creates field with parent_id set to datatype, type from field info, empty validation and ui_config. Creates content_field with route_id, content_data_id, field_id, field_value from parsed data. Creates datatype_field with datatype_id and field_id.

Increments result.FieldsCreated on success. Appends errors to result.Errors on any failure.

### patchSiblingPointers

Sets first_child_id on parent and links children with prev_sibling_id and next_sibling_id pointers. Implements doubly-linked list structure for tree children.

For children array c0, c1, c2: parent.FirstChildID equals c0.ID, c0.PrevSiblingID null and c0.NextSiblingID equals c1.ID, c1.PrevSiblingID equals c0.ID and c1.NextSiblingID equals c2.ID, c2.PrevSiblingID equals c1.ID and c2.NextSiblingID null.

Fetches full child row before update to preserve other fields. Updates content_data with sibling pointers. Appends errors to result.Errors on fetch or update failure.

## Utilities

### respond

Helper function to marshal data to JSON and write to response writer. Returns error if marshal or write fails. Used by OAuth handlers.

### generateSessionToken

Creates cryptographically secure random session token. Generates 32 random bytes, encodes as URL-safe base64 string. Returns error if random read fails.

### writeJSON

Helper function to set Content-Type header to application/json, write 200 OK status, encode value as JSON to response. Used by ContentBatchHandler.

### authcontext

String type used as context key for authenticated user. Value "authenticated" used to store and retrieve user from request context. Used by SSH key handlers.

## Request Structures

### BatchContentUpdateRequest

Request body for POST /api/v1/content/batch. Fields: ContentDataID required, ContentData optional UpdateContentDataParams, Fields optional map from FieldID to string value.

At least one of ContentData or Fields must be present. Validation returns 400 Bad Request if both null.

## Response Structures

### BatchContentUpdateResponse

Response body for POST /api/v1/content/batch. Fields: ContentDataID, ContentDataUpdated boolean, ContentDataError optional string, FieldsUpdated int, FieldsCreated int, FieldsFailed int, Errors optional string array.

Caller checks fields_failed greater than zero or content_data_error not empty for partial failures. HTTP status always 200 OK.

## Import Context

### importContext

Tracks state during recursive import operation. Fields: ctx context.Context, ac audited.AuditContext, driver DbDriver, routeID NullableRouteID, authorID NullableUserID, datatypeCache map from string to DatatypeID, result pointer to ImportResult.

Cache key format: label pipe type. Prevents duplicate datatype creation for same label and type combination.
