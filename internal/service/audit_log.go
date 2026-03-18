package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// AuditLogService provides read-only access to change events for the audit log.
type AuditLogService struct {
	driver db.DbDriver
}

// NewAuditLogService creates an AuditLogService.
func NewAuditLogService(driver db.DbDriver) *AuditLogService {
	return &AuditLogService{driver: driver}
}

// ListChangeEvents returns a paginated list of change events.
func (s *AuditLogService) ListChangeEvents(ctx context.Context, params db.ListChangeEventsParams) (*[]db.ChangeEvent, error) {
	return s.driver.ListChangeEvents(params)
}

// CountChangeEvents returns the total number of change events.
func (s *AuditLogService) CountChangeEvents(ctx context.Context) (*int64, error) {
	return s.driver.CountChangeEvents()
}

// GetChangeEvent retrieves a single change event by ID.
func (s *AuditLogService) GetChangeEvent(ctx context.Context, id types.EventID) (*db.ChangeEvent, error) {
	if err := id.Validate(); err != nil {
		return nil, fmt.Errorf("invalid event ID: %w", err)
	}
	return s.driver.GetChangeEvent(id)
}

// GetChangeEventsByRecord retrieves all change events for a specific record.
func (s *AuditLogService) GetChangeEventsByRecord(ctx context.Context, tableName, recordID string) (*[]db.ChangeEvent, error) {
	return s.driver.GetChangeEventsByRecord(tableName, recordID)
}

// GetRecentActivity returns recent change events with actor info for dashboards.
func (s *AuditLogService) GetRecentActivity(ctx context.Context, limit int64) ([]db.ActivityEventView, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	events, err := s.driver.ListChangeEvents(db.ListChangeEventsParams{
		Limit:  limit,
		Offset: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("list change events: %w", err)
	}
	return db.AssembleRecentActivity(s.driver, *events), nil
}
