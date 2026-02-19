package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerRBACTools(srv *server.MCPServer, client *modulacms.Client) {
	srv.AddTool(
		mcp.NewTool("list_roles",
			mcp.WithDescription("List all roles. Default roles are admin, editor, and viewer."),
		),
		handleListRoles(client),
	)

	srv.AddTool(
		mcp.NewTool("list_permissions",
			mcp.WithDescription("List all permissions. Permissions follow the resource:operation format (e.g. content:read, media:create)."),
		),
		handleListPermissions(client),
	)

	srv.AddTool(
		mcp.NewTool("assign_role_permission",
			mcp.WithDescription("Assign a permission to a role. Creates a role-permission association."),
			mcp.WithString("role_id", mcp.Required(), mcp.Description("Role ID (ULID)")),
			mcp.WithString("permission_id", mcp.Required(), mcp.Description("Permission ID (ULID)")),
		),
		handleAssignRolePermission(client),
	)

	srv.AddTool(
		mcp.NewTool("remove_role_permission",
			mcp.WithDescription("Remove a permission from a role. Requires the role-permission association ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Role-permission association ID (ULID)")),
		),
		handleRemoveRolePermission(client),
	)
}

func handleListRoles(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Roles.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleListPermissions(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Permissions.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAssignRolePermission(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		roleID, err := req.RequireString("role_id")
		if err != nil {
			return mcp.NewToolResultError("role_id is required"), nil
		}
		permID, err := req.RequireString("permission_id")
		if err != nil {
			return mcp.NewToolResultError("permission_id is required"), nil
		}
		params := modulacms.CreateRolePermissionParams{
			RoleID:       modulacms.RoleID(roleID),
			PermissionID: modulacms.PermissionID(permID),
		}
		result, err := client.RolePermissions.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleRemoveRolePermission(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.RolePermissions.Delete(ctx, modulacms.RolePermissionID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
