// export_test.go exposes internal test helpers for use by external test
// packages (package db_test). This is the standard Go idiom for sharing
// test infrastructure across white-box and black-box test files without
// polluting the production API.
//
// Exported names use the ExportedFor prefix to avoid collisions with
// the Test* naming convention that the test framework reserves.
package db

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ExportedIntegrationDB is an exported wrapper around testIntegrationDB for
// use in package db_test.
func ExportedIntegrationDB(t *testing.T) Database {
	return testIntegrationDB(t)
}

// ExportedAuditCtx is an exported wrapper around testAuditCtx.
func ExportedAuditCtx(d Database) audited.AuditContext {
	return testAuditCtx(d)
}

// ExportedAuditCtxWithUser is an exported wrapper around testAuditCtxWithUser.
func ExportedAuditCtxWithUser(d Database, userID types.UserID) audited.AuditContext {
	return testAuditCtxWithUser(d, userID)
}
