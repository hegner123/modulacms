package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerSSHKeyTools(srv *server.MCPServer, backend SSHKeyBackend) {
	srv.AddTool(
		mcp.NewTool("list_ssh_keys",
			mcp.WithDescription("List all SSH keys for the authenticated user."),
		),
		handleListSSHKeys(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_ssh_key",
			mcp.WithDescription("Add a new SSH key for the authenticated user."),
			mcp.WithString("public_key", mcp.Required(), mcp.Description("SSH public key (e.g. 'ssh-ed25519 AAAA...')")),
			mcp.WithString("label", mcp.Required(), mcp.Description("Label for the key (e.g. 'work laptop')")),
		),
		handleCreateSSHKey(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_ssh_key",
			mcp.WithDescription("Delete an SSH key by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("SSH key ID (ULID)")),
		),
		handleDeleteSSHKey(backend),
	)
}

func handleListSSHKeys(backend SSHKeyBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListSSHKeys(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateSSHKey(backend SSHKeyBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		publicKey, err := req.RequireString("public_key")
		if err != nil {
			return mcp.NewToolResultError("public_key is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"public_key": publicKey,
			"label":      label,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateSSHKey(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteSSHKey(backend SSHKeyBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteSSHKey(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
