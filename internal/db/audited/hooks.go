package audited

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// HookEvent identifies the type of audited mutation hook.
type HookEvent string

const (
	HookBeforeCreate  HookEvent = "before_create"
	HookAfterCreate   HookEvent = "after_create"
	HookBeforeUpdate  HookEvent = "before_update"
	HookAfterUpdate   HookEvent = "after_update"
	HookBeforeDelete  HookEvent = "before_delete"
	HookAfterDelete   HookEvent = "after_delete"
	HookBeforePublish HookEvent = "before_publish"
	HookAfterPublish  HookEvent = "after_publish"
	HookBeforeRead    HookEvent = "before_read"
	HookAfterRead     HookEvent = "after_read"
)

// ValidHookEvents is the complete set of valid hook event strings for
// validation in hooks_api.go. Keyed by the string form of HookEvent.
var ValidHookEvents = map[string]HookEvent{
	"before_create":  HookBeforeCreate,
	"after_create":   HookAfterCreate,
	"before_update":  HookBeforeUpdate,
	"after_update":   HookAfterUpdate,
	"before_delete":  HookBeforeDelete,
	"after_delete":   HookAfterDelete,
	"before_publish": HookBeforePublish,
	"after_publish":  HookAfterPublish,
	"before_read":    HookBeforeRead,
	"after_read":     HookAfterRead,
}

// IsReadHookEvent returns true if the event is a read lifecycle hook
// (before_read or after_read). Read hooks have different execution semantics
// than mutation hooks: they run outside any database transaction, so db.*,
// core.*, and request.* calls are permitted.
func IsReadHookEvent(event HookEvent) bool {
	return event == HookBeforeRead || event == HookAfterRead
}

// ReadHookResponse is the structured response returned by a before_read hook
// when it wants to abort delivery and return a custom HTTP response.
// The Lua table shape is { status, headers, json } — same as plugin route
// handler responses.
type ReadHookResponse struct {
	Status  int               // HTTP status code (e.g., 402)
	Headers map[string]string // HTTP headers to set
	Body    map[string]any    // JSON body to marshal and write
}

// ReadHookRunner is the interface for read lifecycle hook dispatch. Separate
// from HookRunner to avoid breaking the 147+ files that implement or reference
// the existing HookRunner interface.
//
// RunBeforeReadHooks executes all matching before_read hooks synchronously.
// Returns a non-nil ReadHookResponse if a hook wants to abort delivery, plus
// a state map that captures any _-prefixed keys set by hooks on the data table.
//
// RunAfterReadHooks executes all matching after_read hooks synchronously.
// Returns collected headers to append to the HTTP response.
type ReadHookRunner interface {
	HasReadHooks(table string) bool
	RunBeforeReadHooks(ctx context.Context, table string, data map[string]any) (*ReadHookResponse, map[string]any, error)
	RunAfterReadHooks(ctx context.Context, table string, data map[string]any, state map[string]any) (map[string]string, error)
}

// HookRunner is the interface between the audited command layer and the plugin
// hook engine. The audited functions call HasHooks for a zero-allocation fast-path
// check, then RunBeforeHooks/RunAfterHooks only when hooks actually exist.
//
// RunBeforeHooks accepts entity as `any` (the raw Go struct); the implementation
// calls StructToMap internally only when hooks match.
// RunAfterHooks is fire-and-forget; errors are logged, not returned.
type HookRunner interface {
	HasHooks(event HookEvent, table string) bool
	RunBeforeHooks(ctx context.Context, event HookEvent, table string, entity any) error
	RunAfterHooks(ctx context.Context, event HookEvent, table string, entity any)
}

// HookError is a sentinel error type returned by RunBeforeHooks when a plugin
// hook aborts a mutation. The original Lua error message is stored in an unexported
// field to prevent accidental inclusion in HTTP responses. Handlers can type-assert
// *HookError to return 422 instead of 500.
type HookError struct {
	PluginName      string
	Event           HookEvent
	Table           string
	originalMessage string // unexported -- prevents accidental inclusion in HTTP responses
}

// Error returns a sanitized message safe for HTTP clients.
func (e *HookError) Error() string {
	return fmt.Sprintf("operation blocked by plugin %q", e.PluginName)
}

// LogMessage returns the original Lua error message for structured logging only.
func (e *HookError) LogMessage() string {
	return e.originalMessage
}

// NewHookError creates a HookError with the given fields.
func NewHookError(pluginName string, event HookEvent, table string, originalMessage string) *HookError {
	return &HookError{
		PluginName:      pluginName,
		Event:           event,
		Table:           table,
		originalMessage: originalMessage,
	}
}

// StructToMap converts a Go struct to map[string]any via a two-step JSON roundtrip:
// (1) marshal to JSON bytes via json.Marshal, (2) decode back into map[string]any
// using json.NewDecoder with UseNumber() to preserve numeric precision as json.Number
// (string-backed) rather than float64.
//
// This approach is necessary because json.Unmarshal into map[string]any coerces all
// numbers to float64, losing precision for large integers. The UseNumber() decoder
// preserves them as json.Number which GoValueToLua handles via the json.Number case.
//
// Returns nil, nil for nil input (defensive: should not happen for Update/Delete
// which always have a before entity, but prevents panics).
func StructToMap(v any) (map[string]any, error) {
	if v == nil {
		return nil, nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("structToMap marshal: %w", err)
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()

	var result map[string]any
	if err := dec.Decode(&result); err != nil {
		return nil, fmt.Errorf("structToMap decode: %w", err)
	}

	return result, nil
}

// DetectStatusTransition checks whether an Update to content_data involves a
// status transition that should fire publish hooks.
//
// Rules (M12):
//   - Only applies to table "content_data". All other tables return nil.
//   - Compares the "status" field (case-sensitive string).
//   - before.Status != "published" AND params contains "status" == "published" -> before_publish
//   - If before is nil: skip detection, return nil.
//   - If params has no "status" field: skip detection (partial update not touching status).
//   - Unknown status values: no extra events.
//
// Returns a slice of 0 or 1 extra HookEvents (the "before_" variant).
// The caller is responsible for mapping before_publish -> after_publish.
func DetectStatusTransition(table string, before map[string]any, params map[string]any) []HookEvent {
	if table != "content_data" {
		return nil
	}
	if before == nil || params == nil {
		return nil
	}

	newStatusRaw, hasStatus := params["status"]
	if !hasStatus {
		return nil
	}
	newStatus, ok := newStatusRaw.(string)
	if !ok {
		return nil
	}

	oldStatusRaw := before["status"]
	oldStatus, _ := oldStatusRaw.(string)

	var events []HookEvent

	if newStatus == "published" && oldStatus != "published" {
		events = append(events, HookBeforePublish)
	}

	return events
}

// BeforeToAfterEvent maps a before-hook event to its corresponding after-hook event.
// Returns the input unchanged if no mapping exists.
func BeforeToAfterEvent(event HookEvent) HookEvent {
	switch event {
	case HookBeforeCreate:
		return HookAfterCreate
	case HookBeforeUpdate:
		return HookAfterUpdate
	case HookBeforeDelete:
		return HookAfterDelete
	case HookBeforePublish:
		return HookAfterPublish
	default:
		return event
	}
}
