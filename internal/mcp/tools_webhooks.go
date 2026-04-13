package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerWebhookTools(srv *server.MCPServer, backend WebhookBackend) {
	srv.AddTool(
		mcp.NewTool("list_webhooks",
			mcp.WithDescription("List all registered webhooks."),
		),
		handleListWebhooks(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_webhook",
			mcp.WithDescription("Get a webhook by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("webhook ID (ULID)")),
		),
		handleGetWebhook(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_webhook",
			mcp.WithDescription("Create a new webhook."),
			mcp.WithString("url", mcp.Required(), mcp.Description("Webhook endpoint URL")),
			mcp.WithObject("events", mcp.Required(), mcp.Description("Array of event names to subscribe to (e.g. [\"content.published\", \"content.updated\"])")),
			mcp.WithString("name", mcp.Description("Webhook name")),
			mcp.WithString("secret", mcp.Description("Signing secret (auto-generated if omitted)")),
			mcp.WithBoolean("active", mcp.Description("Whether the webhook is active (default true)")),
		),
		handleCreateWebhook(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_webhook",
			mcp.WithDescription("Update a webhook."),
			mcp.WithString("id", mcp.Required(), mcp.Description("webhook ID (ULID)")),
			mcp.WithString("url", mcp.Description("Webhook endpoint URL")),
			mcp.WithObject("events", mcp.Description("Array of event names to subscribe to")),
			mcp.WithString("name", mcp.Description("Webhook name")),
			mcp.WithString("secret", mcp.Description("Signing secret")),
			mcp.WithBoolean("active", mcp.Description("Whether the webhook is active")),
		),
		handleUpdateWebhook(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_webhook",
			mcp.WithDescription("Delete a webhook by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("webhook ID (ULID)")),
		),
		handleDeleteWebhook(backend),
	)

	srv.AddTool(
		mcp.NewTool("test_webhook",
			mcp.WithDescription("Send a test event to a webhook."),
			mcp.WithString("id", mcp.Required(), mcp.Description("webhook ID (ULID)")),
		),
		handleTestWebhook(backend),
	)

	srv.AddTool(
		mcp.NewTool("list_webhook_deliveries",
			mcp.WithDescription("List delivery history for a webhook."),
			mcp.WithString("webhook_id", mcp.Required(), mcp.Description("webhook ID (ULID)")),
		),
		handleListWebhookDeliveries(backend),
	)

	srv.AddTool(
		mcp.NewTool("retry_webhook_delivery",
			mcp.WithDescription("Retry a failed webhook delivery."),
			mcp.WithString("delivery_id", mcp.Required(), mcp.Description("delivery ID (ULID)")),
		),
		handleRetryWebhookDelivery(backend),
	)
}

func handleListWebhooks(backend WebhookBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListWebhooks(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetWebhook(backend WebhookBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetWebhook(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateWebhook(backend WebhookBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		url, err := req.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError("url is required"), nil
		}

		events, err := extractStringArray(req, "events")
		if err != nil {
			return mcp.NewToolResultError("events is required and must be an array of strings"), nil
		}
		if len(events) == 0 {
			return mcp.NewToolResultError("events must contain at least one event name"), nil
		}

		active := req.GetBool("active", true)

		params, err := marshalParams(map[string]any{
			"name":      req.GetString("name", ""),
			"url":       url,
			"secret":    req.GetString("secret", ""),
			"events":    events,
			"is_active": active,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateWebhook(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateWebhook(backend WebhookBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}

		m := map[string]any{
			"webhook_id": id,
		}

		if v := optionalStrPtr(req, "name"); v != nil {
			m["name"] = *v
		}
		if v := optionalStrPtr(req, "url"); v != nil {
			m["url"] = *v
		}
		if v := optionalStrPtr(req, "secret"); v != nil {
			m["secret"] = *v
		}

		args := req.GetArguments()
		if _, ok := args["events"]; ok {
			events, evErr := extractStringArray(req, "events")
			if evErr != nil {
				return mcp.NewToolResultError("events must be an array of strings"), nil
			}
			m["events"] = events
		}
		if _, ok := args["active"]; ok {
			m["is_active"] = req.GetBool("active", true)
		}

		params, err := marshalParams(m)
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateWebhook(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteWebhook(backend WebhookBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteWebhook(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleTestWebhook(backend WebhookBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.TestWebhook(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleListWebhookDeliveries(backend WebhookBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		webhookID, err := req.RequireString("webhook_id")
		if err != nil {
			return mcp.NewToolResultError("webhook_id is required"), nil
		}
		data, err := backend.ListWebhookDeliveries(ctx, webhookID)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleRetryWebhookDelivery(backend WebhookBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		deliveryID, err := req.RequireString("delivery_id")
		if err != nil {
			return mcp.NewToolResultError("delivery_id is required"), nil
		}
		err = backend.RetryWebhookDelivery(ctx, deliveryID)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("retry queued"), nil
	}
}

// extractStringArray extracts an array of strings from a request argument.
// Returns an error if the key is missing or the value is not an array.
func extractStringArray(req mcp.CallToolRequest, key string) ([]string, error) {
	args := req.GetArguments()
	raw, ok := args[key]
	if !ok {
		return nil, fmt.Errorf("%s is required", key)
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an array", key)
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result, nil
}
