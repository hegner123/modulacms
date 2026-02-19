package main

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerImportTools(srv *server.MCPServer, client *modulacms.Client) {
	srv.AddTool(
		mcp.NewTool("import_content",
			mcp.WithDescription("Import content from another CMS format. Payloads should be kept under ~5MB. The format determines how the data is parsed and mapped to ModulaCMS entities."),
			mcp.WithString("format", mcp.Required(), mcp.Description("Source CMS format"), mcp.Enum("contentful", "sanity", "strapi", "wordpress", "clean")),
			mcp.WithObject("data", mcp.Required(), mcp.Description("Import data as a JSON object (structure depends on format)")),
		),
		handleImportContent(client),
	)
}

func handleImportContent(client *modulacms.Client) server.ToolHandlerFunc {
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

		var result json.RawMessage
		switch format {
		case "contentful":
			result, err = client.Import.Contentful(ctx, data)
		case "sanity":
			result, err = client.Import.Sanity(ctx, data)
		case "strapi":
			result, err = client.Import.Strapi(ctx, data)
		case "wordpress":
			result, err = client.Import.WordPress(ctx, data)
		case "clean":
			result, err = client.Import.Clean(ctx, data)
		default:
			return mcp.NewToolResultError("invalid format: must be one of contentful, sanity, strapi, wordpress, clean"), nil
		}

		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText(string(result)), nil
	}
}
