package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

type ChangeEvent struct {
	EventID       types.EventID        `json:"event_id"`
	HlcTimestamp  types.HLC            `json:"hlc_timestamp"`
	WallTimestamp types.Timestamp      `json:"wall_timestamp"`
	NodeID        types.NodeID         `json:"node_id"`
	TableName     string               `json:"table_name"`
	RecordID      string               `json:"record_id"`
	Operation     types.Operation      `json:"operation"`
	Action        types.Action         `json:"action"`
	UserID        types.NullableUserID `json:"user_id"`
	OldValues     types.JSONData       `json:"old_values"`
	NewValues     types.JSONData       `json:"new_values"`
	Metadata      types.JSONData       `json:"metadata"`
	SyncedAt      types.Timestamp      `json:"synced_at"`
	ConsumedAt    types.Timestamp      `json:"consumed_at"`
}

type RecordChangeEventParams struct {
	EventID      types.EventID        `json:"event_id"`
	HlcTimestamp types.HLC            `json:"hlc_timestamp"`
	NodeID       types.NodeID         `json:"node_id"`
	TableName    string               `json:"table_name"`
	RecordID     string               `json:"record_id"`
	Operation    types.Operation      `json:"operation"`
	Action       types.Action         `json:"action"`
	UserID       types.NullableUserID `json:"user_id"`
	OldValues    types.JSONData       `json:"old_values"`
	NewValues    types.JSONData       `json:"new_values"`
	Metadata     types.JSONData       `json:"metadata"`
}

type ListChangeEventsParams struct {
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

type ListChangeEventsByUserParams struct {
	UserID types.NullableUserID `json:"user_id"`
	Limit  int64                `json:"limit"`
	Offset int64                `json:"offset"`
}

type ListChangeEventsByActionParams struct {
	Action types.Action `json:"action"`
	Limit  int64        `json:"limit"`
	Offset int64        `json:"offset"`
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapChangeEvent(a mdb.ChangeEvent) ChangeEvent {
	return ChangeEvent{
		EventID:       a.EventID,
		HlcTimestamp:  a.HlcTimestamp,
		WallTimestamp: a.WallTimestamp,
		NodeID:        a.NodeID,
		TableName:     a.TableName,
		RecordID:      a.RecordID,
		Operation:     a.Operation,
		Action:        a.Action,
		UserID:        a.UserID,
		OldValues:     a.OldValues,
		NewValues:     a.NewValues,
		Metadata:      a.Metadata,
		SyncedAt:      a.SyncedAt,
		ConsumedAt:    a.ConsumedAt,
	}
}

func (d Database) MapRecordChangeEventParams(a RecordChangeEventParams) mdb.RecordChangeEventParams {
	return mdb.RecordChangeEventParams{
		EventID:      a.EventID,
		HlcTimestamp: a.HlcTimestamp,
		NodeID:       a.NodeID,
		TableName:    a.TableName,
		RecordID:     a.RecordID,
		Operation:    a.Operation,
		Action:       a.Action,
		UserID:       a.UserID,
		OldValues:    a.OldValues,
		NewValues:    a.NewValues,
		Metadata:     a.Metadata,
	}
}

// QUERIES

func (d Database) CreateChangeEventsTable() error {
	queries := mdb.New(d.Connection)
	return queries.CreateChangeEventsTable(d.Context)
}

func (d Database) DropChangeEventsTable() error {
	queries := mdb.New(d.Connection)
	return queries.DropChangeEventsTable(d.Context)
}

func (d Database) RecordChangeEvent(params RecordChangeEventParams) (*ChangeEvent, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.RecordChangeEvent(d.Context, d.MapRecordChangeEventParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to record change event: %v", err)
	}
	res := d.MapChangeEvent(row)
	return &res, nil
}

func (d Database) GetChangeEvent(id types.EventID) (*ChangeEvent, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetChangeEvent(d.Context, mdb.GetChangeEventParams{EventID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapChangeEvent(row)
	return &res, nil
}

func (d Database) GetChangeEventsByRecord(tableName string, recordID string) (*[]ChangeEvent, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetChangeEventsByRecord(d.Context, mdb.GetChangeEventsByRecordParams{
		TableName: tableName,
		RecordID:  recordID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get change events: %v", err)
	}
	res := []ChangeEvent{}
	for _, v := range rows {
		res = append(res, d.MapChangeEvent(v))
	}
	return &res, nil
}

func (d Database) GetUnsyncedEvents(limit int64) (*[]ChangeEvent, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetUnsyncedEvents(d.Context, mdb.GetUnsyncedEventsParams{Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("failed to get unsynced events: %v", err)
	}
	res := []ChangeEvent{}
	for _, v := range rows {
		res = append(res, d.MapChangeEvent(v))
	}
	return &res, nil
}

func (d Database) GetUnconsumedEvents(limit int64) (*[]ChangeEvent, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetUnconsumedEvents(d.Context, mdb.GetUnconsumedEventsParams{Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("failed to get unconsumed events: %v", err)
	}
	res := []ChangeEvent{}
	for _, v := range rows {
		res = append(res, d.MapChangeEvent(v))
	}
	return &res, nil
}

func (d Database) MarkEventSynced(id types.EventID) error {
	queries := mdb.New(d.Connection)
	return queries.MarkEventSynced(d.Context, mdb.MarkEventSyncedParams{EventID: id})
}

func (d Database) MarkEventConsumed(id types.EventID) error {
	queries := mdb.New(d.Connection)
	return queries.MarkEventConsumed(d.Context, mdb.MarkEventConsumedParams{EventID: id})
}

func (d Database) ListChangeEvents(params ListChangeEventsParams) (*[]ChangeEvent, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListChangeEvents(d.Context, mdb.ListChangeEventsParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list change events: %v", err)
	}
	res := []ChangeEvent{}
	for _, v := range rows {
		res = append(res, d.MapChangeEvent(v))
	}
	return &res, nil
}

func (d Database) CountChangeEvents() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountChangeEvents(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count change events: %v", err)
	}
	return &c, nil
}

func (d Database) DeleteChangeEvent(id types.EventID) error {
	queries := mdb.New(d.Connection)
	return queries.DeleteChangeEvent(d.Context, mdb.DeleteChangeEventParams{EventID: id})
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapChangeEvent(a mdbm.ChangeEvent) ChangeEvent {
	return ChangeEvent{
		EventID:       a.EventID,
		HlcTimestamp:  a.HlcTimestamp,
		WallTimestamp: a.WallTimestamp,
		NodeID:        a.NodeID,
		TableName:     a.TableName,
		RecordID:      a.RecordID,
		Operation:     a.Operation,
		Action:        a.Action,
		UserID:        a.UserID,
		OldValues:     a.OldValues,
		NewValues:     a.NewValues,
		Metadata:      a.Metadata,
		SyncedAt:      a.SyncedAt,
		ConsumedAt:    a.ConsumedAt,
	}
}

func (d MysqlDatabase) MapRecordChangeEventParams(a RecordChangeEventParams) mdbm.RecordChangeEventParams {
	return mdbm.RecordChangeEventParams{
		EventID:      a.EventID,
		HlcTimestamp: a.HlcTimestamp,
		NodeID:       a.NodeID,
		TableName:    a.TableName,
		RecordID:     a.RecordID,
		Operation:    a.Operation,
		Action:       a.Action,
		UserID:       a.UserID,
		OldValues:    a.OldValues,
		NewValues:    a.NewValues,
		Metadata:     a.Metadata,
	}
}

// QUERIES

func (d MysqlDatabase) CreateChangeEventsTable() error {
	queries := mdbm.New(d.Connection)
	return queries.CreateChangeEventsTable(d.Context)
}

func (d MysqlDatabase) DropChangeEventsTable() error {
	queries := mdbm.New(d.Connection)
	return queries.DropChangeEventsTable(d.Context)
}

func (d MysqlDatabase) RecordChangeEvent(params RecordChangeEventParams) (*ChangeEvent, error) {
	queries := mdbm.New(d.Connection)
	err := queries.RecordChangeEvent(d.Context, d.MapRecordChangeEventParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to record change event: %v", err)
	}
	row, err := queries.GetChangeEvent(d.Context, mdbm.GetChangeEventParams{EventID: params.EventID})
	if err != nil {
		return nil, fmt.Errorf("failed to get recorded change event: %v", err)
	}
	res := d.MapChangeEvent(row)
	return &res, nil
}

func (d MysqlDatabase) GetChangeEvent(id types.EventID) (*ChangeEvent, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetChangeEvent(d.Context, mdbm.GetChangeEventParams{EventID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapChangeEvent(row)
	return &res, nil
}

func (d MysqlDatabase) GetChangeEventsByRecord(tableName string, recordID string) (*[]ChangeEvent, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetChangeEventsByRecord(d.Context, mdbm.GetChangeEventsByRecordParams{
		TableName: tableName,
		RecordID:  recordID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get change events: %v", err)
	}
	res := []ChangeEvent{}
	for _, v := range rows {
		res = append(res, d.MapChangeEvent(v))
	}
	return &res, nil
}

func (d MysqlDatabase) GetUnsyncedEvents(limit int64) (*[]ChangeEvent, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetUnsyncedEvents(d.Context, mdbm.GetUnsyncedEventsParams{Limit: int32(limit)})
	if err != nil {
		return nil, fmt.Errorf("failed to get unsynced events: %v", err)
	}
	res := []ChangeEvent{}
	for _, v := range rows {
		res = append(res, d.MapChangeEvent(v))
	}
	return &res, nil
}

func (d MysqlDatabase) GetUnconsumedEvents(limit int64) (*[]ChangeEvent, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetUnconsumedEvents(d.Context, mdbm.GetUnconsumedEventsParams{Limit: int32(limit)})
	if err != nil {
		return nil, fmt.Errorf("failed to get unconsumed events: %v", err)
	}
	res := []ChangeEvent{}
	for _, v := range rows {
		res = append(res, d.MapChangeEvent(v))
	}
	return &res, nil
}

func (d MysqlDatabase) MarkEventSynced(id types.EventID) error {
	queries := mdbm.New(d.Connection)
	return queries.MarkEventSynced(d.Context, mdbm.MarkEventSyncedParams{EventID: id})
}

func (d MysqlDatabase) MarkEventConsumed(id types.EventID) error {
	queries := mdbm.New(d.Connection)
	return queries.MarkEventConsumed(d.Context, mdbm.MarkEventConsumedParams{EventID: id})
}

func (d MysqlDatabase) ListChangeEvents(params ListChangeEventsParams) (*[]ChangeEvent, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListChangeEvents(d.Context, mdbm.ListChangeEventsParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list change events: %v", err)
	}
	res := []ChangeEvent{}
	for _, v := range rows {
		res = append(res, d.MapChangeEvent(v))
	}
	return &res, nil
}

func (d MysqlDatabase) CountChangeEvents() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountChangeEvents(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count change events: %v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) DeleteChangeEvent(id types.EventID) error {
	queries := mdbm.New(d.Connection)
	return queries.DeleteChangeEvent(d.Context, mdbm.DeleteChangeEventParams{EventID: id})
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapChangeEvent(a mdbp.ChangeEvent) ChangeEvent {
	return ChangeEvent{
		EventID:       a.EventID,
		HlcTimestamp:  a.HlcTimestamp,
		WallTimestamp: a.WallTimestamp,
		NodeID:        a.NodeID,
		TableName:     a.TableName,
		RecordID:      a.RecordID,
		Operation:     a.Operation,
		Action:        a.Action,
		UserID:        a.UserID,
		OldValues:     a.OldValues,
		NewValues:     a.NewValues,
		Metadata:      a.Metadata,
		SyncedAt:      a.SyncedAt,
		ConsumedAt:    a.ConsumedAt,
	}
}

func (d PsqlDatabase) MapRecordChangeEventParams(a RecordChangeEventParams) mdbp.RecordChangeEventParams {
	return mdbp.RecordChangeEventParams{
		EventID:      a.EventID,
		HlcTimestamp: a.HlcTimestamp,
		NodeID:       a.NodeID,
		TableName:    a.TableName,
		RecordID:     a.RecordID,
		Operation:    a.Operation,
		Action:       a.Action,
		UserID:       a.UserID,
		OldValues:    a.OldValues,
		NewValues:    a.NewValues,
		Metadata:     a.Metadata,
	}
}

// QUERIES

func (d PsqlDatabase) CreateChangeEventsTable() error {
	queries := mdbp.New(d.Connection)
	return queries.CreateChangeEventsTable(d.Context)
}

func (d PsqlDatabase) DropChangeEventsTable() error {
	queries := mdbp.New(d.Connection)
	return queries.DropChangeEventsTable(d.Context)
}

func (d PsqlDatabase) RecordChangeEvent(params RecordChangeEventParams) (*ChangeEvent, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.RecordChangeEvent(d.Context, d.MapRecordChangeEventParams(params))
	if err != nil {
		return nil, fmt.Errorf("failed to record change event: %v", err)
	}
	res := d.MapChangeEvent(row)
	return &res, nil
}

func (d PsqlDatabase) GetChangeEvent(id types.EventID) (*ChangeEvent, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetChangeEvent(d.Context, mdbp.GetChangeEventParams{EventID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapChangeEvent(row)
	return &res, nil
}

func (d PsqlDatabase) GetChangeEventsByRecord(tableName string, recordID string) (*[]ChangeEvent, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetChangeEventsByRecord(d.Context, mdbp.GetChangeEventsByRecordParams{
		TableName: tableName,
		RecordID:  recordID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get change events: %v", err)
	}
	res := []ChangeEvent{}
	for _, v := range rows {
		res = append(res, d.MapChangeEvent(v))
	}
	return &res, nil
}

func (d PsqlDatabase) GetUnsyncedEvents(limit int64) (*[]ChangeEvent, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetUnsyncedEvents(d.Context, mdbp.GetUnsyncedEventsParams{Limit: int32(limit)})
	if err != nil {
		return nil, fmt.Errorf("failed to get unsynced events: %v", err)
	}
	res := []ChangeEvent{}
	for _, v := range rows {
		res = append(res, d.MapChangeEvent(v))
	}
	return &res, nil
}

func (d PsqlDatabase) GetUnconsumedEvents(limit int64) (*[]ChangeEvent, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetUnconsumedEvents(d.Context, mdbp.GetUnconsumedEventsParams{Limit: int32(limit)})
	if err != nil {
		return nil, fmt.Errorf("failed to get unconsumed events: %v", err)
	}
	res := []ChangeEvent{}
	for _, v := range rows {
		res = append(res, d.MapChangeEvent(v))
	}
	return &res, nil
}

func (d PsqlDatabase) MarkEventSynced(id types.EventID) error {
	queries := mdbp.New(d.Connection)
	return queries.MarkEventSynced(d.Context, mdbp.MarkEventSyncedParams{EventID: id})
}

func (d PsqlDatabase) MarkEventConsumed(id types.EventID) error {
	queries := mdbp.New(d.Connection)
	return queries.MarkEventConsumed(d.Context, mdbp.MarkEventConsumedParams{EventID: id})
}

func (d PsqlDatabase) ListChangeEvents(params ListChangeEventsParams) (*[]ChangeEvent, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListChangeEvents(d.Context, mdbp.ListChangeEventsParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list change events: %v", err)
	}
	res := []ChangeEvent{}
	for _, v := range rows {
		res = append(res, d.MapChangeEvent(v))
	}
	return &res, nil
}

func (d PsqlDatabase) CountChangeEvents() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountChangeEvents(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count change events: %v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) DeleteChangeEvent(id types.EventID) error {
	queries := mdbp.New(d.Connection)
	return queries.DeleteChangeEvent(d.Context, mdbp.DeleteChangeEventParams{EventID: id})
}
