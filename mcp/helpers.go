package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

// errResult formats an error into an MCP error result, preserving API error details.
func errResult(err error) *mcp.CallToolResult {
	var apiErr *modulacms.ApiError
	if errors.As(err, &apiErr) {
		detail := map[string]any{
			"status":  apiErr.StatusCode,
			"message": apiErr.Message,
			"body":    apiErr.Body,
		}
		b, jsonErr := json.Marshal(detail)
		if jsonErr != nil {
			return mcp.NewToolResultError(err.Error())
		}
		return mcp.NewToolResultError(string(b))
	}
	return mcp.NewToolResultError(err.Error())
}

// jsonResult marshals a value to pretty-printed JSON and returns it as a text result.
func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return mcp.NewToolResultText(string(b)), nil
}

// optionalIDPtr extracts an optional string parameter and converts it to a pointer of the given ID type.
// Returns nil if the parameter is missing or empty.
func optionalIDPtr[ID ~string](req mcp.CallToolRequest, key string) *ID {
	s := req.GetString(key, "")
	if s == "" {
		return nil
	}
	id := ID(s)
	return &id
}

// optionalStrPtr extracts an optional string parameter as a pointer.
// Returns nil if the parameter was not provided in the request arguments.
func optionalStrPtr(req mcp.CallToolRequest, key string) *string {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return nil
	}
	return &s
}

// optionalFloat64Ptr extracts an optional float64 parameter as a pointer.
// Returns nil if the parameter was not provided in the request arguments.
func optionalFloat64Ptr(req mcp.CallToolRequest, key string) *float64 {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok {
		return nil
	}
	f, ok := v.(float64)
	if !ok {
		return nil
	}
	return &f
}
