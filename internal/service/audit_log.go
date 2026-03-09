package service

import (
	"context"

	"github.com/hegner123/modulacms/internal/db"
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
