 Plan: ModulaCMS MCP Server

 Context

 ModulaCMS needs an MCP server so AI agents can manage CMS content, schema, media, users, and configuration via the Model Context Protocol. The server will live in mcp/ as a standalone Go
 module, use stdio transport, and authenticate with the CMS API via Bearer token.

 Token Provisioning

 The MCP server authenticates via Authorization: Bearer <token>. To create a token:
 1. Log into ModulaCMS (SSH TUI or API POST /api/v1/auth/login)
 2. Create an API token via POST /api/v1/tokens with token_type: "api_key"
 3. Use the returned token value as MODULACMS_API_KEY

 The token's user must have the admin role for full MCP access. A non-admin token will get 403 errors on RBAC, config, and user management tools.

 Architecture

 - Language: Go (matches project)
 - MCP library: github.com/mark3labs/mcp-go (pin to latest stable release)
 - API client: Import github.com/hegner123/modulacms/sdks/go (published Go SDK)
 - Transport: stdio (standard for Claude Code MCP servers)
 - Config: Environment variables MODULACMS_URL and MODULACMS_API_KEY
 - Module: Separate go.mod at mcp/go.mod (module github.com/hegner123/modulacms/mcp)

 File Structure

 mcp/
   go.mod
   go.sum
   main.go              -- Entry point: env config, SDK client init, server setup, stdio listen
   tools_content.go     -- Content data + content fields + batch + content delivery tools
   tools_schema.go      -- Datatypes + fields + datatype-field links
   tools_media.go       -- Media CRUD + health
   tools_routes.go      -- Route CRUD
   tools_users.go       -- User CRUD + whoami + SSH keys
   tools_rbac.go        -- Roles, permissions, role-permission mappings
   tools_config.go      -- Config get/update/meta
   tools_import.go      -- Import tools (contentful, sanity, strapi, wordpress, clean)
   helpers.go           -- Shared JSON marshaling, error formatting, pointer helpers

 Tools (40 total)

 Content (7 tools) — tools_content.go

 ┌──────────────────┬───────────────────────────┬─────────────────────────────────────────────┐
 │       Tool       │        SDK Method         │                 Key Params                  │
 ├──────────────────┼───────────────────────────┼─────────────────────────────────────────────┤
 │ list_content     │ ContentData.ListPaginated │ limit?, offset?                             │
 ├──────────────────┼───────────────────────────┼─────────────────────────────────────────────┤
 │ get_content      │ ContentData.Get           │ id                                          │
 ├──────────────────┼───────────────────────────┼─────────────────────────────────────────────┤
 │ create_content   │ ContentData.Create        │ parent_id?, route_id?, datatype_id?, status │
 ├──────────────────┼───────────────────────────┼─────────────────────────────────────────────┤
 │ update_content   │ ContentData.Update        │ id, fields to update                        │
 ├──────────────────┼───────────────────────────┼─────────────────────────────────────────────┤
 │ delete_content   │ ContentData.Delete        │ id                                          │
 ├──────────────────┼───────────────────────────┼─────────────────────────────────────────────┤
 │ get_page         │ Content.GetPage           │ slug, format?                               │
 ├──────────────────┼───────────────────────────┼─────────────────────────────────────────────┤
 │ get_content_tree │ AdminTree.Get             │ slug, format?                               │
 └──────────────────┴───────────────────────────┴─────────────────────────────────────────────┘

 Content Fields (4 tools) — tools_content.go

 ┌──────────────────────┬─────────────────────────────┬────────────────────────────────────────┐
 │         Tool         │         SDK Method          │               Key Params               │
 ├──────────────────────┼─────────────────────────────┼────────────────────────────────────────┤
 │ list_content_fields  │ ContentFields.ListPaginated │ limit?, offset?                        │
 ├──────────────────────┼─────────────────────────────┼────────────────────────────────────────┤
 │ create_content_field │ ContentFields.Create        │ content_data_id, field_id, field_value │
 ├──────────────────────┼─────────────────────────────┼────────────────────────────────────────┤
 │ update_content_field │ ContentFields.Update        │ id, field_value                        │
 ├──────────────────────┼─────────────────────────────┼────────────────────────────────────────┤
 │ delete_content_field │ ContentFields.Delete        │ id                                     │
 └──────────────────────┴─────────────────────────────┴────────────────────────────────────────┘

 Note: The API does not support filtering content fields by content_data_id. Use get_content_tree or get_page to see assembled content with its fields. list_content_fields is paginated for
 browsing the full set.

 Content Batch (1 tool) — tools_content.go

 ┌──────────────────────┬─────────────────────┬─────────────────────────────────────────┐
 │         Tool         │     SDK Method      │               Key Params                │
 ├──────────────────────┼─────────────────────┼─────────────────────────────────────────┤
 │ batch_update_content │ ContentBatch.Update │ content_data_id, content_data?, fields? │
 └──────────────────────┴─────────────────────┴─────────────────────────────────────────┘

 Schema (10 tools) — tools_schema.go

 ┌─────────────────┬──────────────────┬──────────────────────────────────────────────────────────────────────────────────────────┐
 │      Tool       │    SDK Method    │                                        Key Params                                        │
 ├─────────────────┼──────────────────┼──────────────────────────────────────────────────────────────────────────────────────────┤
 │ list_datatypes  │ Datatypes.List   │ full? (boolean, uses /full endpoint — returns datatypes with their fields joined)        │
 ├─────────────────┼──────────────────┼──────────────────────────────────────────────────────────────────────────────────────────┤
 │ get_datatype    │ Datatypes.Get    │ id (returns bare datatype without fields; use list_datatypes with full: true for fields) │
 ├─────────────────┼──────────────────┼──────────────────────────────────────────────────────────────────────────────────────────┤
 │ create_datatype │ Datatypes.Create │ label, type, parent_id?                                                                  │
 ├─────────────────┼──────────────────┼──────────────────────────────────────────────────────────────────────────────────────────┤
 │ update_datatype │ Datatypes.Update │ id, label, type                                                                          │
 ├─────────────────┼──────────────────┼──────────────────────────────────────────────────────────────────────────────────────────┤
 │ delete_datatype │ Datatypes.Delete │ id                                                                                       │
 ├─────────────────┼──────────────────┼──────────────────────────────────────────────────────────────────────────────────────────┤
 │ list_fields     │ Fields.List      │ —                                                                                        │
 ├─────────────────┼──────────────────┼──────────────────────────────────────────────────────────────────────────────────────────┤
 │ get_field       │ Fields.Get       │ id                                                                                       │
 ├─────────────────┼──────────────────┼──────────────────────────────────────────────────────────────────────────────────────────┤
 │ create_field    │ Fields.Create    │ label, type, parent_id?, data?, validation?, ui_config?                                  │
 ├─────────────────┼──────────────────┼──────────────────────────────────────────────────────────────────────────────────────────┤
 │ update_field    │ Fields.Update    │ id, label, type, data?, validation?, ui_config?                                          │
 ├─────────────────┼──────────────────┼──────────────────────────────────────────────────────────────────────────────────────────┤
 │ delete_field    │ Fields.Delete    │ id                                                                                       │
 └─────────────────┴──────────────────┴──────────────────────────────────────────────────────────────────────────────────────────┘

 Datatype-Field Links (3 tools) — tools_schema.go

 ┌────────────────────────────┬────────────────────────┬─────────────────────────────────────────────────────┐
 │            Tool            │       SDK Method       │                     Key Params                      │
 ├────────────────────────────┼────────────────────────┼─────────────────────────────────────────────────────┤
 │ list_datatype_fields       │ DatatypeFields.RawList │ datatype_id?, field_id? (API supports both filters) │
 ├────────────────────────────┼────────────────────────┼─────────────────────────────────────────────────────┤
 │ link_field_to_datatype     │ DatatypeFields.Create  │ datatype_id, field_id, sort_order                   │
 ├────────────────────────────┼────────────────────────┼─────────────────────────────────────────────────────┤
 │ unlink_field_from_datatype │ DatatypeFields.Delete  │ id                                                  │
 └────────────────────────────┴────────────────────────┴─────────────────────────────────────────────────────┘

 Media (4 tools) — tools_media.go

 ┌──────────────┬─────────────────────┬─────────────────────┐
 │     Tool     │     SDK Method      │     Key Params      │
 ├──────────────┼─────────────────────┼─────────────────────┤
 │ list_media   │ Media.ListPaginated │ limit?, offset?     │
 ├──────────────┼─────────────────────┼─────────────────────┤
 │ get_media    │ Media.Get           │ id                  │
 ├──────────────┼─────────────────────┼─────────────────────┤
 │ update_media │ Media.Update        │ id, metadata fields │
 ├──────────────┼─────────────────────┼─────────────────────┤
 │ delete_media │ Media.Delete        │ id                  │
 └──────────────┴─────────────────────┴─────────────────────┘

 Note: Media upload (multipart) is omitted — AI agents don't typically upload binary files.

 Routes (5 tools) — tools_routes.go

 ┌──────────────┬───────────────┬─────────────────────┐
 │     Tool     │  SDK Method   │     Key Params      │
 ├──────────────┼───────────────┼─────────────────────┤
 │ list_routes  │ Routes.List   │ —                   │
 ├──────────────┼───────────────┼─────────────────────┤
 │ get_route    │ Routes.Get    │ id                  │
 ├──────────────┼───────────────┼─────────────────────┤
 │ create_route │ Routes.Create │ slug, title, status │
 ├──────────────┼───────────────┼─────────────────────┤
 │ update_route │ Routes.Update │ slug, title, status │
 ├──────────────┼───────────────┼─────────────────────┤
 │ delete_route │ Routes.Delete │ id                  │
 └──────────────┴───────────────┴─────────────────────┘

 Users (5 tools) — tools_users.go

 ┌─────────────┬──────────────┬────────────────────────────────────────────────┐
 │    Tool     │  SDK Method  │                   Key Params                   │
 ├─────────────┼──────────────┼────────────────────────────────────────────────┤
 │ whoami      │ Auth.Me      │ —                                              │
 ├─────────────┼──────────────┼────────────────────────────────────────────────┤
 │ list_users  │ Users.List   │ —                                              │
 ├─────────────┼──────────────┼────────────────────────────────────────────────┤
 │ create_user │ Users.Create │ username, name, email, password, role          │
 ├─────────────┼──────────────┼────────────────────────────────────────────────┤
 │ update_user │ Users.Update │ id, username?, name?, email?, password?, role? │
 ├─────────────┼──────────────┼────────────────────────────────────────────────┤
 │ delete_user │ Users.Delete │ id                                             │
 └─────────────┴──────────────┴────────────────────────────────────────────────┘

 RBAC (4 tools) — tools_rbac.go

 ┌────────────────────────┬────────────────────────┬────────────────────────┐
 │          Tool          │       SDK Method       │       Key Params       │
 ├────────────────────────┼────────────────────────┼────────────────────────┤
 │ list_roles             │ Roles.List             │ —                      │
 ├────────────────────────┼────────────────────────┼────────────────────────┤
 │ list_permissions       │ Permissions.List       │ —                      │
 ├────────────────────────┼────────────────────────┼────────────────────────┤
 │ assign_role_permission │ RolePermissions.Create │ role_id, permission_id │
 ├────────────────────────┼────────────────────────┼────────────────────────┤
 │ remove_role_permission │ RolePermissions.Delete │ id                     │
 └────────────────────────┴────────────────────────┴────────────────────────┘

 Config (3 tools) — tools_config.go

 ┌─────────────────┬───────────────┬───────────────────────────────────────────────────────────────────────────────────┐
 │      Tool       │  SDK Method   │                                    Key Params                                     │
 ├─────────────────┼───────────────┼───────────────────────────────────────────────────────────────────────────────────┤
 │ get_config      │ Config.Get    │ category?                                                                         │
 ├─────────────────┼───────────────┼───────────────────────────────────────────────────────────────────────────────────┤
 │ get_config_meta │ Config.Meta   │ — (returns field metadata: available keys, categories, descriptions, sensitivity) │
 ├─────────────────┼───────────────┼───────────────────────────────────────────────────────────────────────────────────┤
 │ update_config   │ Config.Update │ key-value pairs                                                                   │
 └─────────────────┴───────────────┴───────────────────────────────────────────────────────────────────────────────────┘

 Import (1 tool) — tools_import.go

 ┌────────────────┬────────────────────┬──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │      Tool      │     SDK Method     │                                                                Key Params                                                                │
 ├────────────────┼────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ import_content │ Import.* by format │ format (contentful/sanity/strapi/wordpress/clean), data (JSON). Note: payloads should be kept under ~5MB; large exports should be split. │
 └────────────────┴────────────────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

 Implementation Pattern

 Each tool file follows this structure:

 func registerContentTools(srv *mcp.MCPServer, client *modulacms.Client) {
     srv.AddTool(
         mcp.NewTool("list_content",
             mcp.WithDescription("List content data entries with optional pagination"),
             mcp.WithNumber("limit", mcp.Description("Max items to return"), mcp.DefaultNumber(20)),
             mcp.WithNumber("offset", mcp.Description("Number of items to skip")),
         ),
         handleListContent(client),
     )
     // ... more tools
 }

 func handleListContent(client *modulacms.Client) server.ToolHandlerFunc {
     return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
         limit := int64(optionalNumber(req, "limit", 20))
         offset := int64(optionalNumber(req, "offset", 0))
         result, err := client.ContentData.ListPaginated(ctx, modulacms.PaginationParams{
             Limit: limit, Offset: offset,
         })
         if err != nil {
             return errResult(err), nil
         }
         return jsonResult(result)
     }
 }

 Error Handling

 helpers.go provides errResult(err) which extracts structured info from the SDK's ApiError via errors.As:

 func errResult(err error) *mcp.CallToolResult {
     var apiErr *modulacms.ApiError
     if errors.As(err, &apiErr) {
         detail := map[string]any{
             "status":  apiErr.StatusCode,
             "message": apiErr.Message,
             "body":    apiErr.Body,
         }
         b, _ := json.Marshal(detail)
         return mcp.NewToolResultError(string(b))
     }
     return mcp.NewToolResultError(err.Error())
 }

 This preserves the full API error response (status code, message, and body with validation details) instead of discarding it.

 main.go Structure

 func main() {
     url := os.Getenv("MODULACMS_URL")
     apiKey := os.Getenv("MODULACMS_API_KEY")
     // validate, create SDK client, create MCP server
     // register all tool groups
     // start stdio server
 }

 Omissions (intentional)

 - Admin-namespace CRUD (AdminContentData, AdminDatatypes, etc.) — internal CMS admin, not needed for API agents
 - Media upload — binary multipart not practical for AI agents
 - OAuth endpoints — interactive browser flow
 - Session management — agents use API tokens, not sessions
 - Token management — agents already have a token; creating tokens programmatically is a security concern
 - SSH key management — not relevant for API-based agents
 - Content relations — relation management through content fields; can be added in v2 if needed
 - Plugin management (list/enable/disable/reload, route/hook approval) — admin-only, can be added in v2
 - Tables CRUD — low-level metadata, rarely needed
 - Media dimensions — rarely needed directly
 - Media cleanup — destructive admin operation

 Steps

 1. Create mcp/go.mod with dependencies on mcp-go and the Go SDK
 2. Create mcp/helpers.go — JSON marshaling, error formatting, parameter extraction
 3. Create mcp/tools_content.go — Content, content fields, batch, delivery tools
 4. Create mcp/tools_schema.go — Datatypes, fields, datatype-field link tools
 5. Create mcp/tools_media.go — Media tools
 6. Create mcp/tools_routes.go — Route tools
 7. Create mcp/tools_users.go — User + whoami tools
 8. Create mcp/tools_rbac.go — Roles, permissions, role-permission tools
 9. Create mcp/tools_config.go — Config tools
 10. Create mcp/tools_import.go — Import tool
 11. Create mcp/main.go — Entry point wiring everything together
 12. Run go mod tidy in mcp/ to resolve dependencies
 13. Build and verify: go build -o modulacms-mcp ./mcp/ (or from within mcp/)
 14. Add justfile commands: mcp-build, mcp-install
 15. Add .mcp.json entry for the server

 Verification

 1. cd mcp && go build -o modulacms-mcp . — compiles successfully
 2. MODULACMS_URL=http://localhost:8080 MODULACMS_API_KEY=test ./modulacms-mcp — starts stdio server without crash
 3. Add to .mcp.json and verify tools appear via /mcp in Claude Code
