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
			mcp.WithString("id", mcp.Required(), mcp.Description("user ID (ULID)")),
		),
		handleGetUser(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_user",
			mcp.WithDescription("Create a new user. All fields are required."),
			mcp.WithString("username", mcp.Required(), mcp.Description("Username")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Display name")),
			mcp.WithString("email", mcp.Required(), mcp.Description("email address")),
			mcp.WithString("password", mcp.Required(), mcp.Description("Password")),
			mcp.WithString("role", mcp.Required(), mcp.Description("Role name (e.g. admin, editor, viewer)")),
		),
		handleCreateUser(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_user",
			mcp.WithDescription("Update an existing user."),
			mcp.WithString("id", mcp.Required(), mcp.Description("user ID (ULID)")),
			mcp.WithString("username", mcp.Description("Username")),
			mcp.WithString("name", mcp.Description("Display name")),
			mcp.WithString("email", mcp.Description("email address")),
			mcp.WithString("password", mcp.Description("New password (omit to keep current)")),
			mcp.WithString("role", mcp.Description("Role name")),
		),
		handleUpdateUser(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_user",
			mcp.WithDescription("Delete a user by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("user ID (ULID)")),
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
			mcp.WithString("id", mcp.Required(), mcp.Description("user ID (ULID)")),
		),
		handleGetUserFull(backend),
	)

	srv.AddTool(
		mcp.NewTool("reassign_and_delete_user",
			mcp.WithDescription("Reassign all content from one user to another, then delete the original user."),
			mcp.WithString("user_id", mcp.Required(), mcp.Description("ID of the user to delete (ULID)")),
			mcp.WithString("reassign_to", mcp.Required(), mcp.Description("ID of the user to receive reassigned content (ULID)")),
		),
		handleReassignAndDeleteUser(backend),
	)

	srv.AddTool(
		mcp.NewTool("list_user_sessions",
			mcp.WithDescription("List sessions for the authenticated user."),
		),
		handleListUserSessions(backend),
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
			"username": optionalStrPtr(req, "username"),
			"name":     optionalStrPtr(req, "name"),
			"email":    optionalStrPtr(req, "email"),
			"password": optionalStrPtr(req, "password"),
			"role":     optionalStrPtr(req, "role"),
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

func handleReassignAndDeleteUser(backend UserBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID, err := req.RequireString("user_id")
		if err != nil {
			return mcp.NewToolResultError("user_id is required"), nil
		}
		reassignTo, err := req.RequireString("reassign_to")
		if err != nil {
			return mcp.NewToolResultError("reassign_to is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"user_id":     userID,
			"reassign_to": reassignTo,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.ReassignAndDeleteUser(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleListUserSessions(backend UserBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListUserSessions(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
