package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerLocaleTools(srv *server.MCPServer, backend LocaleBackend) {
	srv.AddTool(
		mcp.NewTool("list_locales",
			mcp.WithDescription("List all enabled locales (public)."),
		),
		handleListLocales(backend),
	)

	srv.AddTool(
		mcp.NewTool("list_admin_locales",
			mcp.WithDescription("List all locales including disabled ones (admin)."),
		),
		handleListAdminLocales(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_locale",
			mcp.WithDescription("Get a locale by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Locale ID (ULID)")),
		),
		handleGetLocale(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_locale",
			mcp.WithDescription("Create a new locale."),
			mcp.WithString("code", mcp.Required(), mcp.Description("BCP 47 locale code (e.g. 'en', 'fr-CA')")),
			mcp.WithString("label", mcp.Required(), mcp.Description("Human-readable label (e.g. 'English', 'French (Canada)')")),
			mcp.WithBoolean("enabled", mcp.Description("Whether the locale is enabled (default false)")),
			mcp.WithBoolean("is_default", mcp.Description("Whether this is the default locale (default false)")),
			mcp.WithString("fallback_code", mcp.Description("Fallback locale code (e.g. 'en' for 'en-US')")),
			mcp.WithNumber("sort_order", mcp.Description("Display sort order (default 0)")),
		),
		handleCreateLocale(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_locale",
			mcp.WithDescription("Update a locale."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Locale ID (ULID)")),
			mcp.WithString("code", mcp.Description("BCP 47 locale code")),
			mcp.WithString("label", mcp.Description("Human-readable label")),
			mcp.WithBoolean("enabled", mcp.Description("Whether the locale is enabled")),
			mcp.WithBoolean("is_default", mcp.Description("Whether this is the default locale")),
			mcp.WithString("fallback_code", mcp.Description("Fallback locale code")),
			mcp.WithNumber("sort_order", mcp.Description("Display sort order")),
		),
		handleUpdateLocale(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_locale",
			mcp.WithDescription("Delete a locale by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Locale ID (ULID)")),
		),
		handleDeleteLocale(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_translation",
			mcp.WithDescription("Create a translation for a content item. Copies translatable fields from the default locale into the target locale as a starting point."),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
			mcp.WithString("locale", mcp.Required(), mcp.Description("Target locale code (e.g. 'fr')")),
		),
		handleCreateTranslation(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_create_translation",
			mcp.WithDescription("Create a translation for an admin content item. Copies translatable fields from the default locale into the target locale as a starting point."),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Admin content data ID (ULID)")),
			mcp.WithString("locale", mcp.Required(), mcp.Description("Target locale code (e.g. 'fr')")),
		),
		handleAdminCreateTranslation(backend),
	)
}

func handleListLocales(backend LocaleBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListLocales(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleListAdminLocales(backend LocaleBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListAdminLocales(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetLocale(backend LocaleBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetLocale(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateLocale(backend LocaleBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		code, err := req.RequireString("code")
		if err != nil {
			return mcp.NewToolResultError("code is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		enabled := req.GetBool("enabled", false)
		isDefault := req.GetBool("is_default", false)
		fallbackCode := req.GetString("fallback_code", "")
		sortOrder := int64(req.GetFloat("sort_order", 0))

		params, err := marshalParams(map[string]any{
			"code":          code,
			"label":         label,
			"is_enabled":    enabled,
			"is_default":    isDefault,
			"fallback_code": fallbackCode,
			"sort_order":    sortOrder,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateLocale(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateLocale(backend LocaleBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"locale_id":     id,
			"code":          optionalStrPtr(req, "code"),
			"label":         optionalStrPtr(req, "label"),
			"is_enabled":    optionalStrPtr(req, "enabled"),
			"is_default":    optionalStrPtr(req, "is_default"),
			"fallback_code": optionalStrPtr(req, "fallback_code"),
			"sort_order":    optionalFloat64Ptr(req, "sort_order"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateLocale(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteLocale(backend LocaleBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteLocale(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleCreateTranslation(backend LocaleBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		locale, err := req.RequireString("locale")
		if err != nil {
			return mcp.NewToolResultError("locale is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"locale": locale,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateTranslation(ctx, contentID, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminCreateTranslation(backend LocaleBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		locale, err := req.RequireString("locale")
		if err != nil {
			return mcp.NewToolResultError("locale is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"locale": locale,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.AdminCreateTranslation(ctx, contentID, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
