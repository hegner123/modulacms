package webhooks

import "time"

// Event name constants for webhook notifications.
const (
	EventContentPublished   = "content.published"
	EventContentUnpublished = "content.unpublished"
	EventContentUpdated     = "content.updated"
	EventContentScheduled   = "content.scheduled"
	EventContentDeleted     = "content.deleted"
	EventLocalePublished    = "locale.published"
	EventVersionCreated     = "version.created"

	// Admin tree mirrors.
	EventAdminContentPublished   = "admin.content.published"
	EventAdminContentUnpublished = "admin.content.unpublished"
	EventAdminContentUpdated     = "admin.content.updated"
	EventAdminContentDeleted     = "admin.content.deleted"

	// System events.
	EventUpdateAvailable = "update.available"
)

// Payload is the top-level envelope sent to webhook endpoints.
type Payload struct {
	ID         string         `json:"id"`
	Event      string         `json:"event"`
	OccurredAt time.Time      `json:"occurred_at"`
	Data       map[string]any `json:"data"`
}
