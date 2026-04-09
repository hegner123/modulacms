package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerImportTools(srv *server.MCPServer, backend ImportBackend) {
	srv.AddTool(
		mcp.NewTool("import_content",
			mcp.WithDescription("import content from another CMS format. Payloads should be kept under ~5MB. The format determines how the data is parsed and mapped to ModulaCMS entities."),
			mcp.WithString("format", mcp.Required(), mcp.Description("Source CMS format"), mcp.Enum("contentful", "sanity", "strapi", "wordpress", "clean")),
			mcp.WithObject("data", mcp.Required(), mcp.Description("import data as a JSON object (structure depends on format)")),
		),
		handleImportContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("import_bulk",
			mcp.WithDescription("Bulk import data. Accepts a raw JSON object that is posted directly to the bulk import endpoint."),
			mcp.WithString("format", mcp.Required(), mcp.Description("import format"), mcp.Enum("contentful", "sanity", "strapi", "wordpress", "clean")),
			mcp.WithObject("data", mcp.Required(), mcp.Description("import data as a JSON object")),
		),
		handleImportBulk(backend),
	)
}

func handleImportContent(backend ImportBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		format, err := req.RequireString("format")
		if err != nil {
			return mcp.NewToolResultError("format is required"), nil
		}
		args := req.GetArguments()
		data, ok := args["data"]
		if !ok {
			return mcp.NewToolResultError("data is required"), nil
		}
		result, err := backend.ImportContent(ctx, format, data)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(result), nil
	}
}

func handleImportBulk(backend ImportBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		format, err := req.RequireString("format")
		if err != nil {
			return mcp.NewToolResultError("format is required"), nil
		}
		args := req.GetArguments()
		data, ok := args["data"]
		if !ok {
			return mcp.NewToolResultError("data is required"), nil
		}
		result, err := backend.ImportBulk(ctx, format, data)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(result), nil
	}
}
