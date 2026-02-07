package audited

import (
	"context"
	"database/sql"
)

// DBTX matches sqlc's generated interface — works with both *sql.DB and *sql.Tx.
type DBTX interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// ChangeEventRecorder records audit events within a transaction.
// Each database driver provides its own implementation.
type ChangeEventRecorder interface {
	Record(ctx context.Context, tx DBTX, p ChangeEventParams) error
}

// CreateCommand is the interface for audited create operations.
// Implementations bundle context, audit context, and params into a single struct.
type CreateCommand[T any] interface {
	Context() context.Context
	AuditContext() AuditContext
	Connection() *sql.DB
	Recorder() ChangeEventRecorder
	TableName() string
	Execute(context.Context, DBTX) (T, error)
	GetID(T) string
	Params() any
}

// UpdateCommand is the interface for audited update operations.
type UpdateCommand[T any] interface {
	Context() context.Context
	AuditContext() AuditContext
	Connection() *sql.DB
	Recorder() ChangeEventRecorder
	TableName() string
	GetBefore(context.Context, DBTX) (T, error)
	Execute(context.Context, DBTX) error
	GetID() string
	Params() any
}

// DeleteCommand is the interface for audited delete operations.
type DeleteCommand[T any] interface {
	Context() context.Context
	AuditContext() AuditContext
	Connection() *sql.DB
	Recorder() ChangeEventRecorder
	TableName() string
	GetBefore(context.Context, DBTX) (T, error)
	Execute(context.Context, DBTX) error
	GetID() string
}
