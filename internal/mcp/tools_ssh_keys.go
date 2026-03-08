package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modula "github.com/hegner123/modulacms/sdks/go"
)

func registerSSHKeyTools(srv *server.MCPServer, client *modula.Client) {
	srv.AddTool(
		mcp.NewTool("list_ssh_keys",
			mcp.WithDescription("List all SSH keys for the authenticated user."),
		),
		handleListSSHKeys(client),
	)

	srv.AddTool(
		mcp.NewTool("create_ssh_key",
			mcp.WithDescription("Add a new SSH key for the authenticated user."),
			mcp.WithString("public_key", mcp.Required(), mcp.Description("SSH public key (e.g. 'ssh-ed25519 AAAA...')")),
			mcp.WithString("label", mcp.Required(), mcp.Description("Label for the key (e.g. 'work laptop')")),
		),
		handleCreateSSHKey(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_ssh_key",
			mcp.WithDescription("Delete an SSH key by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("SSH key ID (ULID)")),
		),
		handleDeleteSSHKey(client),
	)
}

func handleListSSHKeys(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.SSHKeys.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateSSHKey(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		publicKey, err := req.RequireString("public_key")
		if err != nil {
			return mcp.NewToolResultError("public_key is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params := modula.CreateSSHKeyParams{
			PublicKey: publicKey,
			Label:     label,
		}
		result, err := client.SSHKeys.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteSSHKey(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.SSHKeys.Delete(ctx, modula.UserSshKeyID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
