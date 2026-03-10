package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/types"
)

// Delivery status constants.
const (
	DeliveryStatusPending  = "pending"
	DeliveryStatusSuccess  = "success"
	DeliveryStatusFailed   = "failed"
	DeliveryStatusRetrying = "retrying"
)

// UpdateWebhookDeliveryStatusParams holds parameters for updating a delivery's status.
type UpdateWebhookDeliveryStatusParams struct {
	Status         string                  `json:"status"`
	Attempts       int64                   `json:"attempts"`
	LastStatusCode int64                   `json:"last_status_code"`
	LastError      string                  `json:"last_error"`
	NextRetryAt    string                  `json:"next_retry_at"`
	CompletedAt    string                  `json:"completed_at"`
	DeliveryID     types.WebhookDeliveryID `json:"delivery_id"`
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MapWebhookDelivery converts a sqlc-generated SQLite delivery to the wrapper type.
func (d Database) MapWebhookDelivery(a mdb.WebhookDeliveries) WebhookDelivery {
	var ts types.Timestamp
	// CreatedAt is string in SQLite; scan parses it
	if err := ts.Scan(a.CreatedAt); err != nil {
		ts = types.TimestampNow()
	}
	return WebhookDelivery{
		DeliveryID:     a.DeliveryID,
		WebhookID:      a.WebhookID,
		Event:          a.Event,
		Payload:        a.Payload,
		Status:         a.Status,
		Attempts:       a.Attempts,
		LastStatusCode: a.LastStatusCode.Int64,
		LastError:      a.LastError,
		NextRetryAt:    a.NextRetryAt.String,
		CreatedAt:      ts,
		CompletedAt:    a.CompletedAt.String,
	}
}

// MapCreateWebhookDeliveryParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapCreateWebhookDeliveryParams(a CreateWebhookDeliveryParams) mdb.CreateWebhookDeliveryParams {
	return mdb.CreateWebhookDeliveryParams{
		DeliveryID: types.NewWebhookDeliveryID(),
		WebhookID:  a.WebhookID,
		Event:      a.Event,
		Payload:    a.Payload,
		Status:     a.Status,
		Attempts:   a.Attempts,
		CreatedAt:  a.CreatedAt.String(),
	}
}

// CreateWebhookDelivery inserts a new delivery record (non-audited, system operation).
func (d Database) CreateWebhookDelivery(ctx context.Context, s CreateWebhookDeliveryParams) (*WebhookDelivery, error) {
	queries := mdb.New(d.Connection)
	p := d.MapCreateWebhookDeliveryParams(s)
	row, err := queries.CreateWebhookDelivery(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook delivery: %w", err)
	}
	res := d.MapWebhookDelivery(row)
	return &res, nil
}

// DeleteWebhookDelivery removes a delivery record.
func (d Database) DeleteWebhookDelivery(ctx context.Context, id types.WebhookDeliveryID) error {
	queries := mdb.New(d.Connection)
	return queries.DeleteWebhookDelivery(ctx, mdb.DeleteWebhookDeliveryParams{DeliveryID: id})
}

// UpdateWebhookDeliveryStatus updates a delivery's status fields.
func (d Database) UpdateWebhookDeliveryStatus(ctx context.Context, p UpdateWebhookDeliveryStatusParams) error {
	queries := mdb.New(d.Connection)
	return queries.UpdateWebhookDeliveryStatus(ctx, mdb.UpdateWebhookDeliveryStatusParams{
		Status:         p.Status,
		Attempts:       p.Attempts,
		LastStatusCode: sql.NullInt64{Int64: p.LastStatusCode, Valid: p.LastStatusCode != 0},
		LastError:      p.LastError,
		NextRetryAt:    StringToNullString(p.NextRetryAt),
		CompletedAt:    StringToNullString(p.CompletedAt),
		DeliveryID:     p.DeliveryID,
	})
}

// ListPendingRetries returns deliveries with status='retrying' whose next_retry_at has passed.
func (d Database) ListPendingRetries(now types.Timestamp, limit int64) (*[]WebhookDelivery, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListPendingRetries(d.Context, mdb.ListPendingRetriesParams{
		NextRetryAt: sql.NullString{String: now.String(), Valid: true},
		Limit:       limit,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pending retries: %w", err)
	}
	res := []WebhookDelivery{}
	for _, v := range rows {
		res = append(res, d.MapWebhookDelivery(v))
	}
	return &res, nil
}

// PruneOldDeliveries deletes succeeded/failed deliveries older than the given timestamp.
func (d Database) PruneOldDeliveries(ctx context.Context, before types.Timestamp) error {
	queries := mdb.New(d.Connection)
	return queries.PruneOldDeliveries(ctx, mdb.PruneOldDeliveriesParams{
		CreatedAt: before.String(),
	})
}

///////////////////////////////
//////////////////////////////

///////////////////////////////
//////////////////////////////

// parseNullTime converts a string timestamp to sql.NullTime.
func parseNullTime(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

// MYSQL

// DeleteWebhookDelivery removes a delivery record.
func (d MysqlDatabase) DeleteWebhookDelivery(ctx context.Context, id types.WebhookDeliveryID) error {
	queries := mdbm.New(d.Connection)
	return queries.DeleteWebhookDelivery(ctx, mdbm.DeleteWebhookDeliveryParams{DeliveryID: id})
}

// MapWebhookDelivery converts a sqlc-generated MySQL delivery to the wrapper type.
func (d MysqlDatabase) MapWebhookDelivery(a mdbm.WebhookDeliveries) WebhookDelivery {
	var nextRetry string
	if a.NextRetryAt.Valid {
		nextRetry = a.NextRetryAt.Time.UTC().Format(time.RFC3339)
	}
	var completed string
	if a.CompletedAt.Valid {
		completed = a.CompletedAt.Time.UTC().Format(time.RFC3339)
	}
	return WebhookDelivery{
		DeliveryID:     a.DeliveryID,
		WebhookID:      a.WebhookID,
		Event:          a.Event,
		Payload:        a.Payload,
		Status:         a.Status,
		Attempts:       int64(a.Attempts),
		LastStatusCode: int64(a.LastStatusCode.Int32),
		LastError:      a.LastError,
		NextRetryAt:    nextRetry,
		CreatedAt:      types.NewTimestamp(a.CreatedAt.UTC()),
		CompletedAt:    completed,
	}
}

// MapCreateWebhookDeliveryParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreateWebhookDeliveryParams(a CreateWebhookDeliveryParams) mdbm.CreateWebhookDeliveryParams {
	return mdbm.CreateWebhookDeliveryParams{
		DeliveryID: types.NewWebhookDeliveryID(),
		WebhookID:  a.WebhookID,
		Event:      a.Event,
		Payload:    a.Payload,
		Status:     a.Status,
		Attempts:   int32(a.Attempts),
		CreatedAt:  a.CreatedAt.UTC(),
	}
}

// CreateWebhookDelivery inserts a new delivery record (non-audited, system operation).
func (d MysqlDatabase) CreateWebhookDelivery(ctx context.Context, s CreateWebhookDeliveryParams) (*WebhookDelivery, error) {
	queries := mdbm.New(d.Connection)
	p := d.MapCreateWebhookDeliveryParams(s)
	if err := queries.CreateWebhookDelivery(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to create webhook delivery: %w", err)
	}
	row, err := queries.GetWebhookDelivery(ctx, mdbm.GetWebhookDeliveryParams{DeliveryID: p.DeliveryID})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created webhook delivery: %w", err)
	}
	res := d.MapWebhookDelivery(row)
	return &res, nil
}

// UpdateWebhookDeliveryStatus updates a delivery's status fields.
func (d MysqlDatabase) UpdateWebhookDeliveryStatus(ctx context.Context, p UpdateWebhookDeliveryStatusParams) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdateWebhookDeliveryStatus(ctx, mdbm.UpdateWebhookDeliveryStatusParams{
		Status:         p.Status,
		Attempts:       int32(p.Attempts),
		LastStatusCode: sql.NullInt32{Int32: int32(p.LastStatusCode), Valid: p.LastStatusCode != 0},
		LastError:      p.LastError,
		NextRetryAt:    parseNullTime(p.NextRetryAt),
		CompletedAt:    parseNullTime(p.CompletedAt),
		DeliveryID:     p.DeliveryID,
	})
}

// ListPendingRetries returns deliveries with status='retrying' whose next_retry_at has passed.
func (d MysqlDatabase) ListPendingRetries(now types.Timestamp, limit int64) (*[]WebhookDelivery, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListPendingRetries(d.Context, mdbm.ListPendingRetriesParams{
		NextRetryAt: sql.NullTime{Time: now.UTC(), Valid: true},
		Limit:       int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pending retries: %w", err)
	}
	res := []WebhookDelivery{}
	for _, v := range rows {
		res = append(res, d.MapWebhookDelivery(v))
	}
	return &res, nil
}

// PruneOldDeliveries deletes succeeded/failed deliveries older than the given timestamp.
func (d MysqlDatabase) PruneOldDeliveries(ctx context.Context, before types.Timestamp) error {
	queries := mdbm.New(d.Connection)
	return queries.PruneOldDeliveries(ctx, mdbm.PruneOldDeliveriesParams{
		CreatedAt: before.UTC(),
	})
}

// PSQL

// CreateWebhookDelivery inserts a new delivery record (non-audited, system operation).
func (d PsqlDatabase) CreateWebhookDelivery(ctx context.Context, s CreateWebhookDeliveryParams) (*WebhookDelivery, error) {
	queries := mdbp.New(d.Connection)
	p := d.MapCreateWebhookDeliveryParams(s)
	row, err := queries.CreateWebhookDelivery(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook delivery: %w", err)
	}
	res := d.MapWebhookDelivery(row)
	return &res, nil
}

// DeleteWebhookDelivery removes a delivery record.
func (d PsqlDatabase) DeleteWebhookDelivery(ctx context.Context, id types.WebhookDeliveryID) error {
	queries := mdbp.New(d.Connection)
	return queries.DeleteWebhookDelivery(ctx, mdbp.DeleteWebhookDeliveryParams{DeliveryID: id})
}

// MapWebhookDelivery converts a sqlc-generated PostgreSQL delivery to the wrapper type.
func (d PsqlDatabase) MapWebhookDelivery(a mdbp.WebhookDeliveries) WebhookDelivery {
	var nextRetry string
	if a.NextRetryAt.Valid {
		nextRetry = a.NextRetryAt.Time.UTC().Format(time.RFC3339)
	}
	var completed string
	if a.CompletedAt.Valid {
		completed = a.CompletedAt.Time.UTC().Format(time.RFC3339)
	}
	return WebhookDelivery{
		DeliveryID:     a.DeliveryID,
		WebhookID:      a.WebhookID,
		Event:          a.Event,
		Payload:        a.Payload,
		Status:         a.Status,
		Attempts:       int64(a.Attempts),
		LastStatusCode: int64(a.LastStatusCode.Int32),
		LastError:      a.LastError,
		NextRetryAt:    nextRetry,
		CreatedAt:      types.NewTimestamp(a.CreatedAt.UTC()),
		CompletedAt:    completed,
	}
}

// MapCreateWebhookDeliveryParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreateWebhookDeliveryParams(a CreateWebhookDeliveryParams) mdbp.CreateWebhookDeliveryParams {
	return mdbp.CreateWebhookDeliveryParams{
		DeliveryID: types.NewWebhookDeliveryID(),
		WebhookID:  a.WebhookID,
		Event:      a.Event,
		Payload:    a.Payload,
		Status:     a.Status,
		Attempts:   int32(a.Attempts),
		CreatedAt:  a.CreatedAt.UTC(),
	}
}

// UpdateWebhookDeliveryStatus updates a delivery's status fields.
func (d PsqlDatabase) UpdateWebhookDeliveryStatus(ctx context.Context, p UpdateWebhookDeliveryStatusParams) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdateWebhookDeliveryStatus(ctx, mdbp.UpdateWebhookDeliveryStatusParams{
		Status:         p.Status,
		Attempts:       int32(p.Attempts),
		LastStatusCode: sql.NullInt32{Int32: int32(p.LastStatusCode), Valid: p.LastStatusCode != 0},
		LastError:      p.LastError,
		NextRetryAt:    parseNullTime(p.NextRetryAt),
		CompletedAt:    parseNullTime(p.CompletedAt),
		DeliveryID:     p.DeliveryID,
	})
}

// ListPendingRetries returns deliveries with status='retrying' whose next_retry_at has passed.
func (d PsqlDatabase) ListPendingRetries(now types.Timestamp, limit int64) (*[]WebhookDelivery, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListPendingRetries(d.Context, mdbp.ListPendingRetriesParams{
		NextRetryAt: sql.NullTime{Time: now.UTC(), Valid: true},
		Limit:       int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pending retries: %w", err)
	}
	res := []WebhookDelivery{}
	for _, v := range rows {
		res = append(res, d.MapWebhookDelivery(v))
	}
	return &res, nil
}

// PruneOldDeliveries deletes succeeded/failed deliveries older than the given timestamp.
func (d PsqlDatabase) PruneOldDeliveries(ctx context.Context, before types.Timestamp) error {
	queries := mdbp.New(d.Connection)
	return queries.PruneOldDeliveries(ctx, mdbp.PruneOldDeliveriesParams{
		CreatedAt: before.UTC(),
	})
}
