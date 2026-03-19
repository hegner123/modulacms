package mcp

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/service"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// errResult formats an error into an MCP error result, handling both
// service errors (direct mode) and SDK errors (remote mode).
func errResult(err error) *mcp.CallToolResult {
	if service.IsNotFound(err) {
		return mcp.NewToolResultError(formatMCPError(404, err.Error()))
	}
	if service.IsValidation(err) {
		return mcp.NewToolResultError(formatMCPError(422, err.Error()))
	}
	if service.IsConflict(err) {
		return mcp.NewToolResultError(formatMCPError(409, err.Error()))
	}
	if service.IsForbidden(err) {
		return mcp.NewToolResultError(formatMCPError(403, err.Error()))
	}
	if service.IsUnauthorized(err) {
		return mcp.NewToolResultError(formatMCPError(401, err.Error()))
	}

	var apiErr *modula.ApiError
	if errors.As(err, &apiErr) {
		return mcp.NewToolResultError(formatMCPError(apiErr.StatusCode, apiErr.Message))
	}

	return mcp.NewToolResultError(formatMCPError(500, err.Error()))
}

// formatMCPError creates a consistent JSON error string for MCP tool results.
func formatMCPError(status int, message string) string {
	b, jsonErr := json.Marshal(map[string]any{"status": status, "message": message})
	if jsonErr != nil {
		return message
	}
	return string(b)
}

// rawJSONResult wraps pre-marshaled JSON data as an MCP text result.
func rawJSONResult(data json.RawMessage) *mcp.CallToolResult {
	return mcp.NewToolResultText(string(data))
}

// marshalParams marshals a map of parameters to json.RawMessage for backend calls.
func marshalParams(m map[string]any) (json.RawMessage, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}
	return json.RawMessage(b), nil
}

// jsonResult marshals a value to pretty-printed JSON and returns it as a text result.
func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return mcp.NewToolResultText(string(b)), nil
}

// mcpMediaResponse wraps db.Media with a download_url field for MCP responses.
type mcpMediaResponse struct {
	db.Media
	DownloadURL string `json:"download_url"`
}

func toMCPMediaResponse(m db.Media) mcpMediaResponse {
	return mcpMediaResponse{
		Media:       m,
		DownloadURL: "/api/v1/media/" + string(m.MediaID) + "/download",
	}
}

func toMCPMediaList(items []db.Media) []mcpMediaResponse {
	resp := make([]mcpMediaResponse, len(items))
	for i, m := range items {
		resp[i] = toMCPMediaResponse(m)
	}
	return resp
}

// mcpAdminMediaResponse wraps db.AdminMedia with a download_url field for MCP responses.
type mcpAdminMediaResponse struct {
	db.AdminMedia
	DownloadURL string `json:"download_url"`
}

func toMCPAdminMediaResponse(m db.AdminMedia) mcpAdminMediaResponse {
	return mcpAdminMediaResponse{
		AdminMedia:  m,
		DownloadURL: "/api/v1/adminmedia/" + string(m.AdminMediaID) + "/download",
	}
}

func toMCPAdminMediaList(items []db.AdminMedia) []mcpAdminMediaResponse {
	resp := make([]mcpAdminMediaResponse, len(items))
	for i, m := range items {
		resp[i] = toMCPAdminMediaResponse(m)
	}
	return resp
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
