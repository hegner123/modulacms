package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerUserTools(srv *server.MCPServer, backend UserBackend) {
	srv.AddTool(
		mcp.NewTool("whoami",
			mcp.WithDescription("Get the currently authenticated user's profile. Returns user_id, username, name, email, role, and timestamps. Use this to get your user_id for content authoring."),
		),
		handleWhoami(backend),
	)

	srv.AddTool(
		mcp.NewTool("list_users",
			mcp.WithDescription("List all users."),
		),
		handleListUsers(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_user",
			mcp.WithDescription("Get a single user by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("User ID (ULID)")),
		),
		handleGetUser(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_user",
			mcp.WithDescription("Create a new user. All fields are required."),
			mcp.WithString("username", mcp.Required(), mcp.Description("Username")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Display name")),
			mcp.WithString("email", mcp.Required(), mcp.Description("Email address")),
			mcp.WithString("password", mcp.Required(), mcp.Description("Password")),
			mcp.WithString("role", mcp.Required(), mcp.Description("Role name (e.g. admin, editor, viewer)")),
		),
		handleCreateUser(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_user",
			mcp.WithDescription("Update an existing user. This is a full replacement — provide all fields you want to keep. Omitting a field sets it to empty. Password can be omitted to keep the current password."),
			mcp.WithString("id", mcp.Required(), mcp.Description("User ID (ULID)")),
			mcp.WithString("username", mcp.Description("Username")),
			mcp.WithString("name", mcp.Description("Display name")),
			mcp.WithString("email", mcp.Description("Email address")),
			mcp.WithString("password", mcp.Description("New password (omit to keep current)")),
			mcp.WithString("role", mcp.Description("Role name")),
		),
		handleUpdateUser(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_user",
			mcp.WithDescription("Delete a user by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("User ID (ULID)")),
		),
		handleDeleteUser(backend),
	)

	srv.AddTool(
		mcp.NewTool("list_users_full",
			mcp.WithDescription("List all users with full associated data (roles, permissions, sessions, etc.). Returns raw JSON."),
		),
		handleListUsersFull(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_user_full",
			mcp.WithDescription("Get a single user with full associated data by ID. Returns raw JSON."),
			mcp.WithString("id", mcp.Required(), mcp.Description("User ID (ULID)")),
		),
		handleGetUserFull(backend),
	)
}

func handleWhoami(backend UserBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.Whoami(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleListUsers(backend UserBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListUsers(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetUser(backend UserBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetUser(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateUser(backend UserBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		username, err := req.RequireString("username")
		if err != nil {
			return mcp.NewToolResultError("username is required"), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		email, err := req.RequireString("email")
		if err != nil {
			return mcp.NewToolResultError("email is required"), nil
		}
		password, err := req.RequireString("password")
		if err != nil {
			return mcp.NewToolResultError("password is required"), nil
		}
		role, err := req.RequireString("role")
		if err != nil {
			return mcp.NewToolResultError("role is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"username": username,
			"name":     name,
			"email":    email,
			"password": password,
			"role":     role,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateUser(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateUser(backend UserBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"user_id":  id,
			"username": req.GetString("username", ""),
			"name":     req.GetString("name", ""),
			"email":    req.GetString("email", ""),
			"password": req.GetString("password", ""),
			"role":     req.GetString("role", ""),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateUser(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteUser(backend UserBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteUser(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleListUsersFull(backend UserBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListUsersFull(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetUserFull(backend UserBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetUserFull(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
