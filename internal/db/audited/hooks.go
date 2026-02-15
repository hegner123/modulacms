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
	HookBeforeArchive HookEvent = "before_archive"
	HookAfterArchive  HookEvent = "after_archive"
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
	"before_archive": HookBeforeArchive,
	"after_archive":  HookAfterArchive,
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
// status transition that should fire publish/archive hooks.
//
// Rules (M12):
//   - Only applies to table "content_data". All other tables return nil.
//   - Compares the "status" field (case-sensitive string).
//   - before.Status != "published" AND params contains "status" == "published" -> before_publish
//   - before.Status != "archived" AND params contains "status" == "archived" -> before_archive
//   - If before is nil: skip detection, return nil.
//   - If params has no "status" field: skip detection (partial update not touching status).
//   - Unknown status values: no extra events.
//
// Returns a slice of 0, 1, or 2 extra HookEvents (the "before_" variants).
// The caller is responsible for mapping before_publish -> after_publish, etc.
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
	if newStatus == "archived" && oldStatus != "archived" {
		events = append(events, HookBeforeArchive)
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
	case HookBeforeArchive:
		return HookAfterArchive
	default:
		return event
	}
}
