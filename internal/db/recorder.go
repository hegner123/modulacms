package db

import (
	"context"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
)

// Package-level change event recorder instances for each database driver.
var (
	SQLiteRecorder audited.ChangeEventRecorder = sqliteRecorder{}
	MysqlRecorder  audited.ChangeEventRecorder = mysqlRecorder{}
	PsqlRecorder   audited.ChangeEventRecorder = psqlRecorder{}
)

// sqliteRecorder records change events using the SQLite sqlc-generated code.
type sqliteRecorder struct{}

func (sqliteRecorder) Record(ctx context.Context, tx audited.DBTX, p audited.ChangeEventParams) error {
	queries := mdb.New(tx)
	_, err := queries.RecordChangeEvent(ctx, mdb.RecordChangeEventParams{
		EventID:      p.EventID,
		HlcTimestamp: p.HlcTimestamp,
		NodeID:       p.NodeID,
		TableName:    p.TableName,
		RecordID:     p.RecordID,
		Operation:    p.Operation,
		Action:       p.Action,
		UserID:       p.UserID,
		OldValues:    p.OldValues,
		NewValues:    p.NewValues,
		Metadata:     p.Metadata,
		RequestId:    p.RequestID,
		Ip:           p.IP,
	})
	return err
}

// mysqlRecorder records change events using the MySQL sqlc-generated code.
type mysqlRecorder struct{}

func (mysqlRecorder) Record(ctx context.Context, tx audited.DBTX, p audited.ChangeEventParams) error {
	queries := mdbm.New(tx)
	return queries.RecordChangeEvent(ctx, mdbm.RecordChangeEventParams{
		EventID:      p.EventID,
		HlcTimestamp: p.HlcTimestamp,
		NodeID:       p.NodeID,
		TableName:    p.TableName,
		RecordID:     p.RecordID,
		Operation:    p.Operation,
		Action:       p.Action,
		UserID:       p.UserID,
		OldValues:    p.OldValues,
		NewValues:    p.NewValues,
		Metadata:     p.Metadata,
		RequestId:    p.RequestID,
		Ip:           p.IP,
	})
}

// psqlRecorder records change events using the PostgreSQL sqlc-generated code.
type psqlRecorder struct{}

func (psqlRecorder) Record(ctx context.Context, tx audited.DBTX, p audited.ChangeEventParams) error {
	queries := mdbp.New(tx)
	_, err := queries.RecordChangeEvent(ctx, mdbp.RecordChangeEventParams{
		EventID:      p.EventID,
		HlcTimestamp: p.HlcTimestamp,
		NodeID:       p.NodeID,
		TableName:    p.TableName,
		RecordID:     p.RecordID,
		Operation:    p.Operation,
		Action:       p.Action,
		UserID:       p.UserID,
		OldValues:    p.OldValues,
		NewValues:    p.NewValues,
		Metadata:     p.Metadata,
		RequestId:    p.RequestID,
		Ip:           p.IP,
	})
	return err
}
