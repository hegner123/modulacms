package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerRBACTools(srv *server.MCPServer, backend RBACBackend) {
	srv.AddTool(
		mcp.NewTool("list_roles",
			mcp.WithDescription("List all roles. Default roles are admin, editor, and viewer."),
		),
		handleListRoles(backend),
	)

	srv.AddTool(
		mcp.NewTool("list_permissions",
			mcp.WithDescription("List all permissions. Permissions follow the resource:operation format (e.g. content:read, media:create)."),
		),
		handleListPermissions(backend),
	)

	srv.AddTool(
		mcp.NewTool("assign_role_permission",
			mcp.WithDescription("Assign a permission to a role. Creates a role-permission association."),
			mcp.WithString("role_id", mcp.Required(), mcp.Description("Role ID (ULID)")),
			mcp.WithString("permission_id", mcp.Required(), mcp.Description("Permission ID (ULID)")),
		),
		handleAssignRolePermission(backend),
	)

	srv.AddTool(
		mcp.NewTool("remove_role_permission",
			mcp.WithDescription("Remove a permission from a role. Requires the role-permission association ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Role-permission association ID (ULID)")),
		),
		handleRemoveRolePermission(backend),
	)

	// --- Role CRUD ---

	srv.AddTool(
		mcp.NewTool("get_role",
			mcp.WithDescription("Get a single role by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Role ID (ULID)")),
		),
		handleGetRole(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_role",
			mcp.WithDescription("Create a new role."),
			mcp.WithString("label", mcp.Required(), mcp.Description("Role label (e.g. 'moderator')")),
		),
		handleCreateRole(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_role",
			mcp.WithDescription("Update a role's label."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Role ID (ULID)")),
			mcp.WithString("label", mcp.Required(), mcp.Description("New role label")),
		),
		handleUpdateRole(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_role",
			mcp.WithDescription("Delete a role by ID. System-protected roles cannot be deleted."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Role ID (ULID)")),
		),
		handleDeleteRole(backend),
	)

	// --- Permission CRUD ---

	srv.AddTool(
		mcp.NewTool("get_permission",
			mcp.WithDescription("Get a single permission by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Permission ID (ULID)")),
		),
		handleGetPermission(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_permission",
			mcp.WithDescription("Create a new permission. Label must follow resource:operation format (e.g. 'content:read')."),
			mcp.WithString("label", mcp.Required(), mcp.Description("Permission label (resource:operation format)")),
		),
		handleCreatePermission(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_permission",
			mcp.WithDescription("Update a permission's label."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Permission ID (ULID)")),
			mcp.WithString("label", mcp.Required(), mcp.Description("New permission label (resource:operation format)")),
		),
		handleUpdatePermission(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_permission",
			mcp.WithDescription("Delete a permission by ID. System-protected permissions cannot be deleted."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Permission ID (ULID)")),
		),
		handleDeletePermission(backend),
	)

	// --- Role-Permission queries ---

	srv.AddTool(
		mcp.NewTool("list_role_permissions",
			mcp.WithDescription("List all role-permission associations."),
		),
		handleListRolePermissions(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_role_permission",
			mcp.WithDescription("Get a single role-permission association by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Role-permission ID (ULID)")),
		),
		handleGetRolePermission(backend),
	)

	srv.AddTool(
		mcp.NewTool("list_role_permissions_by_role",
			mcp.WithDescription("List all permissions assigned to a specific role."),
			mcp.WithString("role_id", mcp.Required(), mcp.Description("Role ID (ULID)")),
		),
		handleListRolePermissionsByRole(backend),
	)
}

func handleListRoles(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListRoles(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleListPermissions(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListPermissions(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAssignRolePermission(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		roleID, err := req.RequireString("role_id")
		if err != nil {
			return mcp.NewToolResultError("role_id is required"), nil
		}
		permID, err := req.RequireString("permission_id")
		if err != nil {
			return mcp.NewToolResultError("permission_id is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"role_id":       roleID,
			"permission_id": permID,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.AssignRolePermission(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleRemoveRolePermission(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.RemoveRolePermission(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleGetRole(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetRole(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateRole(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"label": label,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateRole(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateRole(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"role_id": id,
			"label":   label,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateRole(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteRole(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteRole(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleGetPermission(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetPermission(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreatePermission(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"label": label,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreatePermission(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdatePermission(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"permission_id": id,
			"label":         label,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdatePermission(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeletePermission(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeletePermission(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleListRolePermissions(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListRolePermissions(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetRolePermission(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetRolePermission(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleListRolePermissionsByRole(backend RBACBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		roleID, err := req.RequireString("role_id")
		if err != nil {
			return mcp.NewToolResultError("role_id is required"), nil
		}
		data, err := backend.ListRolePermissionsByRole(ctx, roleID)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
