package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerValidationTools(srv *server.MCPServer, backend ValidationBackend) {
	// --- Public validations ---

	srv.AddTool(
		mcp.NewTool("list_validations",
			mcp.WithDescription("List all validation rules."),
		),
		handleListValidations(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_validation",
			mcp.WithDescription("Get a validation rule by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Validation ID (ULID)")),
		),
		handleGetValidation(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_validation",
			mcp.WithDescription("Create a validation rule."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Validation name")),
			mcp.WithString("description", mcp.Description("Validation description")),
			mcp.WithObject("config", mcp.Required(), mcp.Description("Validation rules configuration (JSON object)")),
		),
		handleCreateValidation(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_validation",
			mcp.WithDescription("Update a validation rule."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Validation ID (ULID)")),
			mcp.WithString("name", mcp.Description("Validation name")),
			mcp.WithString("description", mcp.Description("Validation description")),
			mcp.WithObject("config", mcp.Description("Validation rules configuration (JSON object)")),
		),
		handleUpdateValidation(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_validation",
			mcp.WithDescription("Delete a validation rule."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Validation ID (ULID)")),
		),
		handleDeleteValidation(backend),
	)

	srv.AddTool(
		mcp.NewTool("search_validations",
			mcp.WithDescription("Search validation rules by name."),
			mcp.WithString("query", mcp.Required(), mcp.Description("Name substring to search for")),
		),
		handleSearchValidations(backend),
	)

	// --- Admin validations ---

	srv.AddTool(
		mcp.NewTool("admin_list_validations",
			mcp.WithDescription("List all admin validation rules."),
		),
		handleAdminListValidations(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_get_validation",
			mcp.WithDescription("Get an admin validation rule by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin validation ID (ULID)")),
		),
		handleAdminGetValidation(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_create_validation",
			mcp.WithDescription("Create an admin validation rule."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Validation name")),
			mcp.WithString("description", mcp.Description("Validation description")),
			mcp.WithObject("config", mcp.Required(), mcp.Description("Validation rules configuration (JSON object)")),
		),
		handleAdminCreateValidation(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_update_validation",
			mcp.WithDescription("Update an admin validation rule."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin validation ID (ULID)")),
			mcp.WithString("name", mcp.Description("Validation name")),
			mcp.WithString("description", mcp.Description("Validation description")),
			mcp.WithObject("config", mcp.Description("Validation rules configuration (JSON object)")),
		),
		handleAdminUpdateValidation(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_delete_validation",
			mcp.WithDescription("Delete an admin validation rule."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin validation ID (ULID)")),
		),
		handleAdminDeleteValidation(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_search_validations",
			mcp.WithDescription("Search admin validation rules by name."),
			mcp.WithString("query", mcp.Required(), mcp.Description("Name substring to search for")),
		),
		handleAdminSearchValidations(backend),
	)
}

// --- Public handlers ---

func handleListValidations(backend ValidationBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListValidations(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetValidation(backend ValidationBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetValidation(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateValidation(backend ValidationBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		configStr, err := marshalObjectParam(req, "config")
		if err != nil {
			return mcp.NewToolResultError("config is required and must be a JSON object"), nil
		}
		params, err := marshalParams(map[string]any{
			"name":        name,
			"description": optionalStrPtr(req, "description"),
			"config":      configStr,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateValidation(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateValidation(backend ValidationBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		m := map[string]any{
			"validation_id": id,
		}
		if v := optionalStrPtr(req, "name"); v != nil {
			m["name"] = *v
		}
		if v := optionalStrPtr(req, "description"); v != nil {
			m["description"] = *v
		}
		if configStr, cErr := marshalObjectParam(req, "config"); cErr == nil {
			m["config"] = configStr
		}
		params, err := marshalParams(m)
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateValidation(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteValidation(backend ValidationBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		if err = backend.DeleteValidation(ctx, id); err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleSearchValidations(backend ValidationBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError("query is required"), nil
		}
		data, err := backend.SearchValidations(ctx, query)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

// --- Admin handlers ---

func handleAdminListValidations(backend ValidationBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.AdminListValidations(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminGetValidation(backend ValidationBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.AdminGetValidation(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminCreateValidation(backend ValidationBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		configStr, err := marshalObjectParam(req, "config")
		if err != nil {
			return mcp.NewToolResultError("config is required and must be a JSON object"), nil
		}
		params, err := marshalParams(map[string]any{
			"name":        name,
			"description": optionalStrPtr(req, "description"),
			"config":      configStr,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.AdminCreateValidation(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminUpdateValidation(backend ValidationBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		m := map[string]any{
			"admin_validation_id": id,
		}
		if v := optionalStrPtr(req, "name"); v != nil {
			m["name"] = *v
		}
		if v := optionalStrPtr(req, "description"); v != nil {
			m["description"] = *v
		}
		if configStr, cErr := marshalObjectParam(req, "config"); cErr == nil {
			m["config"] = configStr
		}
		params, err := marshalParams(m)
		if err != nil {
			return nil, err
		}
		data, err := backend.AdminUpdateValidation(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminDeleteValidation(backend ValidationBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		if err = backend.AdminDeleteValidation(ctx, id); err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleAdminSearchValidations(backend ValidationBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError("query is required"), nil
		}
		data, err := backend.AdminSearchValidations(ctx, query)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

// marshalObjectParam extracts a JSON object parameter from the request and
// marshals it to a string. MCP delivers object params as map[string]any,
// so we re-serialize to produce the JSON string the backends expect.
func marshalObjectParam(req mcp.CallToolRequest, key string) (string, error) {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok {
		return "", fmt.Errorf("parameter %q is required", key)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshal %q: %w", key, err)
	}
	return string(b), nil
}
