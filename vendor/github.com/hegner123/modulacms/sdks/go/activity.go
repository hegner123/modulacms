package modula

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ActivityItem represents a single recent activity event from the audit log,
// including the actor who performed the action. These events are returned by
// [ActivityResource.ListRecent] and are suitable for building activity feeds
// and admin dashboards.
type ActivityItem struct {
	// EventID is the unique identifier for this change event.
	EventID string `json:"event_id"`
	// TableName is the database table that was modified (e.g. "content_data", "media").
	TableName string `json:"table_name"`
	// RecordID is the primary key of the affected record.
	RecordID string `json:"record_id"`
	// Operation is the type of database operation (e.g. "INSERT", "UPDATE", "DELETE").
	Operation string `json:"operation"`
	// Action is a human-readable action label (e.g. "create", "update", "delete").
	Action string `json:"action"`
	// Actor contains the user who performed the action, or nil for system events.
	Actor *ActivityActor `json:"actor,omitempty"`
	// Timestamp is the wall-clock time when the event occurred (ISO 8601).
	Timestamp string `json:"timestamp"`
}

// ActivityActor identifies the user who performed an activity event.
type ActivityActor struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// ActivityResource provides access to the audit activity feed.
// Requires audit:read permission.
// It is accessed via [Client].Activity.
type ActivityResource struct {
	http *httpClient
}

// ListRecent returns the most recent activity events, up to the given limit.
// Pass 0 for limit to use the server default (25). Maximum is 100.
func (a *ActivityResource) ListRecent(ctx context.Context, limit int) ([]ActivityItem, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	var result []ActivityItem
	if err := a.http.get(ctx, "/api/v1/activity/recent", params, &result); err != nil {
		return nil, fmt.Errorf("list recent activity: %w", err)
	}
	return result, nil
}
