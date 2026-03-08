package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modula "github.com/hegner123/modulacms/sdks/go"
)

func registerRBACTools(srv *server.MCPServer, client *modula.Client) {
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

	// --- Role CRUD ---

	srv.AddTool(
		mcp.NewTool("get_role",
			mcp.WithDescription("Get a single role by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Role ID (ULID)")),
		),
		handleGetRole(client),
	)

	srv.AddTool(
		mcp.NewTool("create_role",
			mcp.WithDescription("Create a new role."),
			mcp.WithString("label", mcp.Required(), mcp.Description("Role label (e.g. 'moderator')")),
		),
		handleCreateRole(client),
	)

	srv.AddTool(
		mcp.NewTool("update_role",
			mcp.WithDescription("Update a role's label."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Role ID (ULID)")),
			mcp.WithString("label", mcp.Required(), mcp.Description("New role label")),
		),
		handleUpdateRole(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_role",
			mcp.WithDescription("Delete a role by ID. System-protected roles cannot be deleted."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Role ID (ULID)")),
		),
		handleDeleteRole(client),
	)

	// --- Permission CRUD ---

	srv.AddTool(
		mcp.NewTool("get_permission",
			mcp.WithDescription("Get a single permission by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Permission ID (ULID)")),
		),
		handleGetPermission(client),
	)

	srv.AddTool(
		mcp.NewTool("create_permission",
			mcp.WithDescription("Create a new permission. Label must follow resource:operation format (e.g. 'content:read')."),
			mcp.WithString("label", mcp.Required(), mcp.Description("Permission label (resource:operation format)")),
		),
		handleCreatePermission(client),
	)

	srv.AddTool(
		mcp.NewTool("update_permission",
			mcp.WithDescription("Update a permission's label."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Permission ID (ULID)")),
			mcp.WithString("label", mcp.Required(), mcp.Description("New permission label (resource:operation format)")),
		),
		handleUpdatePermission(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_permission",
			mcp.WithDescription("Delete a permission by ID. System-protected permissions cannot be deleted."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Permission ID (ULID)")),
		),
		handleDeletePermission(client),
	)

	// --- Role-Permission queries ---

	srv.AddTool(
		mcp.NewTool("list_role_permissions",
			mcp.WithDescription("List all role-permission associations."),
		),
		handleListRolePermissions(client),
	)

	srv.AddTool(
		mcp.NewTool("get_role_permission",
			mcp.WithDescription("Get a single role-permission association by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Role-permission ID (ULID)")),
		),
		handleGetRolePermission(client),
	)

	srv.AddTool(
		mcp.NewTool("list_role_permissions_by_role",
			mcp.WithDescription("List all permissions assigned to a specific role."),
			mcp.WithString("role_id", mcp.Required(), mcp.Description("Role ID (ULID)")),
		),
		handleListRolePermissionsByRole(client),
	)
}

func handleListRoles(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Roles.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleListPermissions(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Permissions.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAssignRolePermission(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		roleID, err := req.RequireString("role_id")
		if err != nil {
			return mcp.NewToolResultError("role_id is required"), nil
		}
		permID, err := req.RequireString("permission_id")
		if err != nil {
			return mcp.NewToolResultError("permission_id is required"), nil
		}
		params := modula.CreateRolePermissionParams{
			RoleID:       modula.RoleID(roleID),
			PermissionID: modula.PermissionID(permID),
		}
		result, err := client.RolePermissions.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleRemoveRolePermission(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.RolePermissions.Delete(ctx, modula.RolePermissionID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleGetRole(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.Roles.Get(ctx, modula.RoleID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateRole(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params := modula.CreateRoleParams{Label: label}
		result, err := client.Roles.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateRole(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params := modula.UpdateRoleParams{
			RoleID: modula.RoleID(id),
			Label:  label,
		}
		result, err := client.Roles.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteRole(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.Roles.Delete(ctx, modula.RoleID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleGetPermission(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.Permissions.Get(ctx, modula.PermissionID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreatePermission(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params := modula.CreatePermissionParams{Label: label}
		result, err := client.Permissions.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdatePermission(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params := modula.UpdatePermissionParams{
			PermissionID: modula.PermissionID(id),
			Label:        label,
		}
		result, err := client.Permissions.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeletePermission(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.Permissions.Delete(ctx, modula.PermissionID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleListRolePermissions(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.RolePermissions.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetRolePermission(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.RolePermissions.Get(ctx, modula.RolePermissionID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleListRolePermissionsByRole(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		roleID, err := req.RequireString("role_id")
		if err != nil {
			return mcp.NewToolResultError("role_id is required"), nil
		}
		result, err := client.RolePermissions.ListByRole(ctx, modula.RoleID(roleID))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}
