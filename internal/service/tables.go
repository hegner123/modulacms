package service

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
)

// TableService manages CMS metadata table CRUD.
type TableService struct {
	driver db.DbDriver
}

// NewTableService creates a TableService.
func NewTableService(driver db.DbDriver) *TableService {
	return &TableService{driver: driver}
}

// GetTable retrieves a table by ID. Returns NotFoundError if not found.
func (s *TableService) GetTable(ctx context.Context, tableID string) (*db.Tables, error) {
	table, err := s.driver.GetTable(tableID)
	if err != nil {
		return nil, &NotFoundError{Resource: "table", ID: tableID}
	}
	return table, nil
}

// ListTables returns all registered tables.
func (s *TableService) ListTables(ctx context.Context) (*[]db.Tables, error) {
	return s.driver.ListTables()
}

// UpdateTable updates a table and returns the freshly-fetched result.
func (s *TableService) UpdateTable(ctx context.Context, ac audited.AuditContext, params db.UpdateTableParams) (*db.Tables, error) {
	_, err := s.driver.UpdateTable(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("update table %q: %w", params.ID, err)
	}

	table, err := s.driver.GetTable(params.ID)
	if err != nil {
		return nil, &NotFoundError{Resource: "table", ID: params.ID}
	}
	return table, nil
}

// DeleteTable removes a table by ID.
func (s *TableService) DeleteTable(ctx context.Context, ac audited.AuditContext, tableID string) error {
	return s.driver.DeleteTable(ctx, ac, tableID)
}
