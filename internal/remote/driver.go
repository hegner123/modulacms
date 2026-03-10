// Package remote implements db.DbDriver over the Modula Go SDK.
//
// Audited mutation methods accept audited.AuditContext to satisfy the DbDriver
// interface but discard it -- the remote server creates audit records from the
// authenticated request context (Bearer token identity + request metadata).
package remote

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	config "github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	modula "github.com/hegner123/modulacms/sdks/go"
)

// RemoteStatus represents the connection state of the remote driver.
type RemoteStatus int32

const (
	// StatusUnknown is the initial state before any SDK call completes.
	StatusUnknown RemoteStatus = iota
	// StatusConnected indicates the last SDK call succeeded or returned an HTTP error (server is reachable).
	StatusConnected
	// StatusDisconnected indicates the last SDK call failed with a network error (timeout, connection refused, etc.).
	StatusDisconnected
)

// String returns a human-readable label for the status.
func (s RemoteStatus) String() string {
	switch s {
	case StatusConnected:
		return "connected"
	case StatusDisconnected:
		return "disconnected"
	default:
		return "unknown"
	}
}

// RemoteDriver implements db.DbDriver by delegating to the Modula Go SDK.
// Read-only methods (List*, Get*, Count*) are backed by real SDK calls.
// DDL, raw SQL, and infrastructure methods return ErrNotSupported or ErrRemoteMode.
type RemoteDriver struct {
	client *modula.Client
	url    string
	status atomic.Int32 // stores RemoteStatus
}

// compile-time check: RemoteDriver implements db.DbDriver.
var _ db.DbDriver = (*RemoteDriver)(nil)

// NewDriver creates a RemoteDriver that talks to a CMS server at url.
// It validates the client config and performs a health check before returning.
func NewDriver(url, apiKey string) (*RemoteDriver, error) {
	client, err := modula.NewClient(modula.ClientConfig{
		BaseURL:    url,
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("invalid client config for %s: %w", url, err)
	}
	if _, err := client.Health.Check(context.Background()); err != nil {
		return nil, fmt.Errorf("cannot reach %s: %w", url, err)
	}
	return &RemoteDriver{client: client, url: url}, nil
}

// Status returns the current connection status (thread-safe).
func (r *RemoteDriver) Status() RemoteStatus {
	return RemoteStatus(r.status.Load())
}

// RemoteStatus returns a human-readable connection label for the TUI status bar.
// The TUI calls this via consumer-defined interface without importing this package.
func (r *RemoteDriver) RemoteConnectionStatus() string {
	return RemoteStatus(r.status.Load()).String()
}

// trackStatus updates the connection status based on the SDK call result.
// nil or HTTP-level errors (4xx/5xx) -> Connected (server reachable).
// Network errors (timeout, connection refused) -> Disconnected.
func (r *RemoteDriver) trackStatus(err error) {
	if err == nil {
		r.status.Store(int32(StatusConnected))
		return
	}
	// If we got an ApiError, the server responded — it's reachable
	var apiErr *modula.ApiError
	if isApiError(err, &apiErr) {
		r.status.Store(int32(StatusConnected))
		return
	}
	// Network-level failure
	r.status.Store(int32(StatusDisconnected))
}

// isApiError unwraps through fmt.Errorf wrappers to find an ApiError.
func isApiError(err error, target **modula.ApiError) bool {
	return err != nil && findApiError(err, target)
}

func findApiError(err error, target **modula.ApiError) bool {
	for e := err; e != nil; {
		if ae, ok := e.(*modula.ApiError); ok {
			*target = ae
			return true
		}
		u, ok := e.(interface{ Unwrap() error })
		if !ok {
			return false
		}
		e = u.Unwrap()
	}
	return false
}

// doRead wraps a read operation with single-retry on transient failures
// and updates the connection status based on the result.
func doRead[T any](r *RemoteDriver, fn func() (T, error)) (T, error) {
	result, err := retryRead(fn)
	r.trackStatus(err)
	return result, err
}

// doWrite runs a write operation and updates the connection status.
// No retry for writes (not idempotent).
func doWrite[T any](r *RemoteDriver, fn func() (T, error)) (T, error) {
	result, err := fn()
	r.trackStatus(err)
	return result, err
}

// doWriteErr runs a write operation that returns only an error.
func doWriteErr(r *RemoteDriver, fn func() error) error {
	err := fn()
	r.trackStatus(err)
	return err
}

// ---------------------------------------------------------------------------
// Connection
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CreateAllTables() error {
	return ErrNotSupported{Method: "CreateAllTables"}
}

func (r *RemoteDriver) CreateBootstrapData(adminHash string) error {
	return ErrNotSupported{Method: "CreateBootstrapData"}
}

func (r *RemoteDriver) DropAllTables() error {
	return ErrNotSupported{Method: "DropAllTables"}
}

func (r *RemoteDriver) DumpSql(_ config.Config) error {
	return ErrNotSupported{Method: "DumpSql"}
}

func (r *RemoteDriver) ExecuteQuery(_ string, _ db.DBTable) (*sql.Rows, error) {
	return nil, ErrNotSupported{Method: "ExecuteQuery"}
}

func (r *RemoteDriver) GetConnection() (*sql.DB, context.Context, error) {
	return nil, nil, ErrRemoteMode
}

func (r *RemoteDriver) GetForeignKeys(_ []string) *sql.Rows {
	return nil
}

func (r *RemoteDriver) Ping() error {
	_, err := r.client.Health.Check(context.Background())
	if err != nil {
		return fmt.Errorf("remote ping failed: %w", err)
	}
	return nil
}

func (r *RemoteDriver) Query(_ *sql.DB, _ string) (sql.Result, error) {
	return nil, ErrNotSupported{Method: "Query"}
}

func (r *RemoteDriver) ScanForeignKeyQueryRows(_ *sql.Rows) []db.SqliteForeignKeyQueryRow {
	return nil
}

func (r *RemoteDriver) SelectColumnFromTable(_ string, _ string) {
	return
}

func (r *RemoteDriver) SortTables() error {
	return ErrNotSupported{Method: "SortTables"}
}

func (r *RemoteDriver) ValidateBootstrapData() error {
	return ErrNotSupported{Method: "ValidateBootstrapData"}
}

// ---------------------------------------------------------------------------
// AdminContentData
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountAdminContentData() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.AdminContentData.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountAdminContentData: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateAdminContentData(ctx context.Context, _ audited.AuditContext, params db.CreateAdminContentDataParams) (*db.AdminContentData, error) {
	return doWrite(r, func() (*db.AdminContentData, error) {
		sdkParams := adminContentDataCreateFromDb(params)
		result, err := r.client.AdminContentData.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateAdminContentData: %w", err)
		}
		row := adminContentDataToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateAdminContentDataTable() error {
	return ErrNotSupported{Method: "CreateAdminContentDataTable"}
}

func (r *RemoteDriver) DeleteAdminContentData(ctx context.Context, _ audited.AuditContext, id types.AdminContentID) error {
	return doWriteErr(r, func() error {
		if err := r.client.AdminContentData.Delete(ctx, modula.AdminContentID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteAdminContentData: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetAdminContentData(id types.AdminContentID) (*db.AdminContentData, error) {
	return doRead(r, func() (*db.AdminContentData, error) {
		ctx := context.Background()
		sdkID := modula.AdminContentID(string(id))
		item, err := r.client.AdminContentData.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetAdminContentData: %w", err)
		}
		result := adminContentDataToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentData() (*[]db.AdminContentData, error) {
	return doRead(r, func() (*[]db.AdminContentData, error) {
		ctx := context.Background()
		items, err := r.client.AdminContentData.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentData: %w", err)
		}
		result := make([]db.AdminContentData, len(items))
		for i := range items {
			result[i] = adminContentDataToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentDataByRoute(routeID types.NullableAdminRouteID) (*[]db.AdminContentData, error) {
	return doRead(r, func() (*[]db.AdminContentData, error) {
		ctx := context.Background()
		params := url.Values{}
		if routeID.Valid {
			params.Set("admin_route_id", string(routeID.ID))
		}
		raw, err := r.client.AdminContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentDataByRoute: %w", err)
		}
		var sdkItems []modula.AdminContentData
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentDataByRoute: decode: %w", err)
		}
		result := make([]db.AdminContentData, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminContentDataToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentDataPaginated(p db.PaginationParams) (*[]db.AdminContentData, error) {
	return doRead(r, func() (*[]db.AdminContentData, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.AdminContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentDataPaginated: %w", err)
		}
		var sdkItems []modula.AdminContentData
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentDataPaginated: decode: %w", err)
		}
		result := make([]db.AdminContentData, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminContentDataToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentDataTopLevelPaginated(p db.PaginationParams) (*[]db.AdminContentDataTopLevel, error) {
	return doRead(r, func() (*[]db.AdminContentDataTopLevel, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		params.Set("top_level", "true")
		raw, err := r.client.AdminContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentDataTopLevelPaginated: %w", err)
		}
		var sdkItems []modula.AdminContentData
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentDataTopLevelPaginated: decode: %w", err)
		}
		result := make([]db.AdminContentDataTopLevel, len(sdkItems))
		for i := range sdkItems {
			result[i] = db.AdminContentDataTopLevel{
				AdminContentData: adminContentDataToDb(&sdkItems[i]),
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentDataByRoutePaginated(p db.ListAdminContentDataByRoutePaginatedParams) (*[]db.AdminContentData, error) {
	return doRead(r, func() (*[]db.AdminContentData, error) {
		ctx := context.Background()
		params := url.Values{}
		if p.AdminRouteID.Valid {
			params.Set("admin_route_id", string(p.AdminRouteID.ID))
		}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.AdminContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentDataByRoutePaginated: %w", err)
		}
		var sdkItems []modula.AdminContentData
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentDataByRoutePaginated: decode: %w", err)
		}
		result := make([]db.AdminContentData, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminContentDataToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentDataWithDatatypeByRoute(_ types.NullableAdminRouteID) (*[]db.AdminContentDataWithDatatypeRow, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentDataWithDatatypeByRoute"}
}

func (r *RemoteDriver) CountAdminContentDataTopLevel() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("count", "true")
		params.Set("top_level", "true")
		raw, err := r.client.AdminContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: CountAdminContentDataTopLevel: %w", err)
		}
		var countResp struct {
			Count int64 `json:"count"`
		}
		if err := json.Unmarshal(raw, &countResp); err != nil {
			return nil, fmt.Errorf("remote: CountAdminContentDataTopLevel: decode: %w", err)
		}
		return &countResp.Count, nil
	})
}

func (r *RemoteDriver) UpdateAdminContentData(ctx context.Context, _ audited.AuditContext, params db.UpdateAdminContentDataParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := adminContentDataUpdateFromDb(params)
		if _, err := r.client.AdminContentData.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateAdminContentData: %w", err)
		}
		id := string(params.AdminContentDataID)
		return &id, nil
	})
}

func (r *RemoteDriver) UpdateAdminContentDataPublishMeta(_ context.Context, _ db.UpdateAdminContentDataPublishMetaParams) error {
	return ErrNotSupported{Method: "UpdateAdminContentDataPublishMeta"}
}

func (r *RemoteDriver) UpdateAdminContentDataWithRevision(_ context.Context, _ db.UpdateAdminContentDataWithRevisionParams) error {
	return ErrNotSupported{Method: "UpdateAdminContentDataWithRevision"}
}

func (r *RemoteDriver) UpdateAdminContentDataSchedule(_ context.Context, _ db.UpdateAdminContentDataScheduleParams) error {
	return ErrNotSupported{Method: "UpdateAdminContentDataSchedule"}
}

func (r *RemoteDriver) ClearAdminContentDataSchedule(_ context.Context, _ db.ClearAdminContentDataScheduleParams) error {
	return ErrNotSupported{Method: "ClearAdminContentDataSchedule"}
}

func (r *RemoteDriver) ListAdminContentDataDueForPublish(_ types.Timestamp) (*[]db.AdminContentData, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentDataDueForPublish"}
}

func (r *RemoteDriver) GetAdminContentDataDescendants(_ context.Context, _ types.AdminContentID) (*[]db.AdminContentData, error) {
	return nil, ErrNotSupported{Method: "GetAdminContentDataDescendants"}
}

func (r *RemoteDriver) ListAdminContentDataByRootID(_ types.NullableAdminContentID) (*[]db.AdminContentData, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentDataByRootID"}
}

func (r *RemoteDriver) ListAdminContentDataWithDatatypeByRootID(_ types.NullableAdminContentID) (*[]db.AdminContentDataWithDatatypeRow, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentDataWithDatatypeByRootID"}
}

// ---------------------------------------------------------------------------
// AdminContentFields
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountAdminContentFields() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.AdminContentFields.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountAdminContentFields: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateAdminContentField(ctx context.Context, _ audited.AuditContext, params db.CreateAdminContentFieldParams) (*db.AdminContentFields, error) {
	return doWrite(r, func() (*db.AdminContentFields, error) {
		sdkParams := modula.CreateAdminContentFieldParams{
			AdminRouteID:       adminRouteIDPtr(params.AdminRouteID),
			AdminContentDataID: adminContentIDPtr(params.AdminContentDataID),
			AdminFieldID:       adminFieldIDPtr(params.AdminFieldID),
			AdminFieldValue:    params.AdminFieldValue,
			AuthorID:           userIDToSdkPtr(params.AuthorID),
		}
		result, err := r.client.AdminContentFields.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateAdminContentField: %w", err)
		}
		row := adminContentFieldToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateAdminContentFieldTable() error {
	return ErrNotSupported{Method: "CreateAdminContentFieldTable"}
}

func (r *RemoteDriver) DeleteAdminContentField(ctx context.Context, _ audited.AuditContext, id types.AdminContentFieldID) error {
	return doWriteErr(r, func() error {
		if err := r.client.AdminContentFields.Delete(ctx, modula.AdminContentFieldID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteAdminContentField: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetAdminContentField(id types.AdminContentFieldID) (*db.AdminContentFields, error) {
	return doRead(r, func() (*db.AdminContentFields, error) {
		ctx := context.Background()
		sdkID := modula.AdminContentFieldID(string(id))
		item, err := r.client.AdminContentFields.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetAdminContentField: %w", err)
		}
		result := adminContentFieldToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentFields() (*[]db.AdminContentFields, error) {
	return doRead(r, func() (*[]db.AdminContentFields, error) {
		ctx := context.Background()
		items, err := r.client.AdminContentFields.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentFields: %w", err)
		}
		result := make([]db.AdminContentFields, len(items))
		for i := range items {
			result[i] = adminContentFieldToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentFieldsByRoute(routeID types.NullableAdminRouteID) (*[]db.AdminContentFields, error) {
	return doRead(r, func() (*[]db.AdminContentFields, error) {
		ctx := context.Background()
		params := url.Values{}
		if routeID.Valid {
			params.Set("admin_route_id", string(routeID.ID))
		}
		raw, err := r.client.AdminContentFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentFieldsByRoute: %w", err)
		}
		var sdkItems []modula.AdminContentField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentFieldsByRoute: decode: %w", err)
		}
		result := make([]db.AdminContentFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminContentFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentFieldsPaginated(p db.PaginationParams) (*[]db.AdminContentFields, error) {
	return doRead(r, func() (*[]db.AdminContentFields, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.AdminContentFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentFieldsPaginated: %w", err)
		}
		var sdkItems []modula.AdminContentField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentFieldsPaginated: decode: %w", err)
		}
		result := make([]db.AdminContentFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminContentFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentFieldsByRoutePaginated(p db.ListAdminContentFieldsByRoutePaginatedParams) (*[]db.AdminContentFields, error) {
	return doRead(r, func() (*[]db.AdminContentFields, error) {
		ctx := context.Background()
		params := url.Values{}
		if p.AdminRouteID.Valid {
			params.Set("admin_route_id", string(p.AdminRouteID.ID))
		}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.AdminContentFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentFieldsByRoutePaginated: %w", err)
		}
		var sdkItems []modula.AdminContentField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentFieldsByRoutePaginated: decode: %w", err)
		}
		result := make([]db.AdminContentFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminContentFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentFieldsWithFieldByRoute(_ types.NullableAdminRouteID) (*[]db.AdminContentFieldsWithFieldRow, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentFieldsWithFieldByRoute"}
}

func (r *RemoteDriver) ListAdminContentFieldsByContentData(_ types.NullableAdminContentID) (*[]db.AdminContentFields, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentFieldsByContentData"}
}

func (r *RemoteDriver) ListAdminContentFieldsWithFieldByContentData(_ types.NullableAdminContentID) (*[]db.AdminContentFieldsWithFieldRow, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentFieldsWithFieldByContentData"}
}

func (r *RemoteDriver) ListAdminContentFieldsByContentDataIDs(_ context.Context, _ []types.AdminContentID, _ string) (*[]db.AdminContentFields, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentFieldsByContentDataIDs"}
}

func (r *RemoteDriver) ListAdminContentFieldsByContentDataAndLocale(contentDataID types.NullableAdminContentID, locale string) (*[]db.AdminContentFields, error) {
	return doRead(r, func() (*[]db.AdminContentFields, error) {
		ctx := context.Background()
		params := url.Values{}
		if contentDataID.Valid {
			params.Set("admin_content_data_id", string(contentDataID.ID))
		}
		params.Set("locale", locale)
		raw, err := r.client.AdminContentFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentFieldsByContentDataAndLocale: %w", err)
		}
		var sdkItems []modula.AdminContentField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentFieldsByContentDataAndLocale: decode: %w", err)
		}
		result := make([]db.AdminContentFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminContentFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentFieldsByRouteAndLocale(routeID types.NullableAdminRouteID, locale string) (*[]db.AdminContentFields, error) {
	return doRead(r, func() (*[]db.AdminContentFields, error) {
		ctx := context.Background()
		params := url.Values{}
		if routeID.Valid {
			params.Set("admin_route_id", string(routeID.ID))
		}
		params.Set("locale", locale)
		raw, err := r.client.AdminContentFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentFieldsByRouteAndLocale: %w", err)
		}
		var sdkItems []modula.AdminContentField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentFieldsByRouteAndLocale: decode: %w", err)
		}
		result := make([]db.AdminContentFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminContentFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateAdminContentField(ctx context.Context, _ audited.AuditContext, params db.UpdateAdminContentFieldParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := modula.UpdateAdminContentFieldParams{
			AdminContentFieldID: modula.AdminContentFieldID(string(params.AdminContentFieldID)),
			AdminRouteID:        adminRouteIDPtr(params.AdminRouteID),
			AdminContentDataID:  adminContentIDPtr(params.AdminContentDataID),
			AdminFieldID:        adminFieldIDPtr(params.AdminFieldID),
			AdminFieldValue:     params.AdminFieldValue,
			AuthorID:            userIDToSdkPtr(params.AuthorID),
		}
		if _, err := r.client.AdminContentFields.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateAdminContentField: %w", err)
		}
		id := string(params.AdminContentFieldID)
		return &id, nil
	})
}

func (r *RemoteDriver) ListAdminContentFieldsByRootID(_ types.NullableAdminContentID) (*[]db.AdminContentFields, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentFieldsByRootID"}
}

func (r *RemoteDriver) ListAdminContentFieldsByRootIDAndLocale(_ types.NullableAdminContentID, _ string) (*[]db.AdminContentFields, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentFieldsByRootIDAndLocale"}
}

// ---------------------------------------------------------------------------
// AdminContentRelations
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountAdminContentRelations() (*int64, error) {
	return nil, ErrNotSupported{Method: "CountAdminContentRelations"}
}

func (r *RemoteDriver) CreateAdminContentRelation(_ context.Context, _ audited.AuditContext, _ db.CreateAdminContentRelationParams) (*db.AdminContentRelations, error) {
	return nil, ErrNotSupported{Method: "CreateAdminContentRelation"}
}

func (r *RemoteDriver) CreateAdminContentRelationTable() error {
	return ErrNotSupported{Method: "CreateAdminContentRelationTable"}
}

func (r *RemoteDriver) DeleteAdminContentRelation(_ context.Context, _ audited.AuditContext, _ types.AdminContentRelationID) error {
	return ErrNotSupported{Method: "DeleteAdminContentRelation"}
}

func (r *RemoteDriver) DropAdminContentRelationTable() error {
	return ErrNotSupported{Method: "DropAdminContentRelationTable"}
}

func (r *RemoteDriver) GetAdminContentRelation(_ types.AdminContentRelationID) (*db.AdminContentRelations, error) {
	return nil, ErrNotSupported{Method: "GetAdminContentRelation"}
}

func (r *RemoteDriver) ListAdminContentRelations() (*[]db.AdminContentRelations, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentRelations"}
}

func (r *RemoteDriver) ListAdminContentRelationsBySource(_ types.AdminContentID) (*[]db.AdminContentRelations, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentRelationsBySource"}
}

func (r *RemoteDriver) ListAdminContentRelationsByTarget(_ types.AdminContentID) (*[]db.AdminContentRelations, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentRelationsByTarget"}
}

func (r *RemoteDriver) ListAdminContentRelationsBySourceAndField(_ types.AdminContentID, _ types.AdminFieldID) (*[]db.AdminContentRelations, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentRelationsBySourceAndField"}
}

func (r *RemoteDriver) UpdateAdminContentRelationSortOrder(_ context.Context, _ audited.AuditContext, _ db.UpdateAdminContentRelationSortOrderParams) error {
	return ErrNotSupported{Method: "UpdateAdminContentRelationSortOrder"}
}

// ---------------------------------------------------------------------------
// AdminContentVersions
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountAdminContentVersions() (*int64, error) {
	return nil, ErrNotSupported{Method: "CountAdminContentVersions"}
}

func (r *RemoteDriver) CountAdminContentVersionsByContent(_ types.AdminContentID) (*int64, error) {
	return nil, ErrNotSupported{Method: "CountAdminContentVersionsByContent"}
}

func (r *RemoteDriver) CreateAdminContentVersion(_ context.Context, _ audited.AuditContext, _ db.CreateAdminContentVersionParams) (*db.AdminContentVersion, error) {
	return nil, ErrNotSupported{Method: "CreateAdminContentVersion"}
}

func (r *RemoteDriver) CreateAdminContentVersionTable() error {
	return ErrNotSupported{Method: "CreateAdminContentVersionTable"}
}

func (r *RemoteDriver) DropAdminContentVersionTable() error {
	return ErrNotSupported{Method: "DropAdminContentVersionTable"}
}

func (r *RemoteDriver) DeleteAdminContentVersion(ctx context.Context, _ audited.AuditContext, id types.AdminContentVersionID) error {
	return doWriteErr(r, func() error {
		if err := r.client.AdminPublishing.DeleteVersion(ctx, string(id)); err != nil {
			return fmt.Errorf("remote: DeleteAdminContentVersion: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetAdminContentVersion(id types.AdminContentVersionID) (*db.AdminContentVersion, error) {
	return doRead(r, func() (*db.AdminContentVersion, error) {
		ctx := context.Background()
		item, err := r.client.AdminPublishing.GetAdminVersion(ctx, string(id))
		if err != nil {
			return nil, fmt.Errorf("remote: GetAdminContentVersion: %w", err)
		}
		row := adminContentVersionToDb(item)
		return &row, nil
	})
}

func (r *RemoteDriver) GetAdminPublishedSnapshot(_ types.AdminContentID, _ string) (*db.AdminContentVersion, error) {
	return nil, ErrNotSupported{Method: "GetAdminPublishedSnapshot"}
}

func (r *RemoteDriver) ListAdminContentVersionsByContent(contentID types.AdminContentID) (*[]db.AdminContentVersion, error) {
	return doRead(r, func() (*[]db.AdminContentVersion, error) {
		ctx := context.Background()
		items, err := r.client.AdminPublishing.ListAdminVersions(ctx, string(contentID))
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminContentVersionsByContent: %w", err)
		}
		result := make([]db.AdminContentVersion, len(items))
		for i := range items {
			result[i] = adminContentVersionToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminContentVersionsByContentLocale(_ types.AdminContentID, _ string) (*[]db.AdminContentVersion, error) {
	return nil, ErrNotSupported{Method: "ListAdminContentVersionsByContentLocale"}
}

func (r *RemoteDriver) ClearAdminPublishedFlag(_ types.AdminContentID, _ string) error {
	return ErrNotSupported{Method: "ClearAdminPublishedFlag"}
}

func (r *RemoteDriver) GetAdminMaxVersionNumber(_ types.AdminContentID, _ string) (int64, error) {
	return 0, ErrNotSupported{Method: "GetAdminMaxVersionNumber"}
}

func (r *RemoteDriver) PruneAdminOldVersions(_ types.AdminContentID, _ string, _ int64) error {
	return ErrNotSupported{Method: "PruneAdminOldVersions"}
}

// ---------------------------------------------------------------------------
// AdminDatatypes
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountAdminDatatypes() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.AdminDatatypes.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountAdminDatatypes: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateAdminDatatype(ctx context.Context, _ audited.AuditContext, params db.CreateAdminDatatypeParams) (*db.AdminDatatypes, error) {
	return doWrite(r, func() (*db.AdminDatatypes, error) {
		sdkParams := adminDatatypeCreateFromDb(params)
		result, err := r.client.AdminDatatypes.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateAdminDatatype: %w", err)
		}
		row := adminDatatypeToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateAdminDatatypeTable() error {
	return ErrNotSupported{Method: "CreateAdminDatatypeTable"}
}

func (r *RemoteDriver) DeleteAdminDatatype(ctx context.Context, _ audited.AuditContext, id types.AdminDatatypeID) error {
	return doWriteErr(r, func() error {
		if err := r.client.AdminDatatypes.Delete(ctx, modula.AdminDatatypeID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteAdminDatatype: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetAdminDatatypeById(id types.AdminDatatypeID) (*db.AdminDatatypes, error) {
	return doRead(r, func() (*db.AdminDatatypes, error) {
		ctx := context.Background()
		sdkID := modula.AdminDatatypeID(string(id))
		item, err := r.client.AdminDatatypes.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetAdminDatatypeById: %w", err)
		}
		result := adminDatatypeToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminDatatypes() (*[]db.AdminDatatypes, error) {
	return doRead(r, func() (*[]db.AdminDatatypes, error) {
		ctx := context.Background()
		items, err := r.client.AdminDatatypes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminDatatypes: %w", err)
		}
		result := make([]db.AdminDatatypes, len(items))
		for i := range items {
			result[i] = adminDatatypeToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminDatatypesPaginated(p db.PaginationParams) (*[]db.AdminDatatypes, error) {
	return doRead(r, func() (*[]db.AdminDatatypes, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.AdminDatatypes.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminDatatypesPaginated: %w", err)
		}
		var sdkItems []modula.AdminDatatype
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminDatatypesPaginated: decode: %w", err)
		}
		result := make([]db.AdminDatatypes, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminDatatypeToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminDatatypeChildrenPaginated(p db.ListAdminDatatypeChildrenPaginatedParams) (*[]db.AdminDatatypes, error) {
	return doRead(r, func() (*[]db.AdminDatatypes, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("parent_id", string(p.ParentID))
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.AdminDatatypes.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminDatatypeChildrenPaginated: %w", err)
		}
		var sdkItems []modula.AdminDatatype
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminDatatypeChildrenPaginated: decode: %w", err)
		}
		result := make([]db.AdminDatatypes, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminDatatypeToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateAdminDatatype(ctx context.Context, _ audited.AuditContext, params db.UpdateAdminDatatypeParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := adminDatatypeUpdateFromDb(params)
		if _, err := r.client.AdminDatatypes.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateAdminDatatype: %w", err)
		}
		id := string(params.AdminDatatypeID)
		return &id, nil
	})
}

// ---------------------------------------------------------------------------
// AdminFields
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountAdminFields() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.AdminFields.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountAdminFields: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateAdminField(ctx context.Context, _ audited.AuditContext, params db.CreateAdminFieldParams) (*db.AdminFields, error) {
	return doWrite(r, func() (*db.AdminFields, error) {
		sdkParams := adminFieldCreateFromDb(params)
		result, err := r.client.AdminFields.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateAdminField: %w", err)
		}
		row := adminFieldToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateAdminFieldTable() error {
	return ErrNotSupported{Method: "CreateAdminFieldTable"}
}

func (r *RemoteDriver) DeleteAdminField(ctx context.Context, _ audited.AuditContext, id types.AdminFieldID) error {
	return doWriteErr(r, func() error {
		if err := r.client.AdminFields.Delete(ctx, modula.AdminFieldID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteAdminField: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetAdminField(id types.AdminFieldID) (*db.AdminFields, error) {
	return doRead(r, func() (*db.AdminFields, error) {
		ctx := context.Background()
		sdkID := modula.AdminFieldID(string(id))
		item, err := r.client.AdminFields.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetAdminField: %w", err)
		}
		result := adminFieldToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminFields() (*[]db.AdminFields, error) {
	return doRead(r, func() (*[]db.AdminFields, error) {
		ctx := context.Background()
		items, err := r.client.AdminFields.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminFields: %w", err)
		}
		result := make([]db.AdminFields, len(items))
		for i := range items {
			result[i] = adminFieldToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminFieldsPaginated(p db.PaginationParams) (*[]db.AdminFields, error) {
	return doRead(r, func() (*[]db.AdminFields, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.AdminFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminFieldsPaginated: %w", err)
		}
		var sdkItems []modula.AdminField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminFieldsPaginated: decode: %w", err)
		}
		result := make([]db.AdminFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminFieldsByParentIDPaginated(p db.ListAdminFieldsByParentIDPaginatedParams) (*[]db.AdminFields, error) {
	return doRead(r, func() (*[]db.AdminFields, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("parent_id", string(p.ParentID))
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.AdminFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminFieldsByParentIDPaginated: %w", err)
		}
		var sdkItems []modula.AdminField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminFieldsByParentIDPaginated: decode: %w", err)
		}
		result := make([]db.AdminFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminFieldsByDatatypeID(parentID types.NullableAdminDatatypeID) (*[]db.AdminFields, error) {
	return doRead(r, func() (*[]db.AdminFields, error) {
		ctx := context.Background()
		params := url.Values{}
		if parentID.Valid {
			params.Set("parent_id", string(parentID.ID))
		}
		raw, err := r.client.AdminFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminFieldsByDatatypeID: %w", err)
		}
		var sdkItems []modula.AdminField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminFieldsByDatatypeID: decode: %w", err)
		}
		result := make([]db.AdminFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateAdminField(ctx context.Context, _ audited.AuditContext, params db.UpdateAdminFieldParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := adminFieldUpdateFromDb(params)
		if _, err := r.client.AdminFields.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateAdminField: %w", err)
		}
		id := string(params.AdminFieldID)
		return &id, nil
	})
}

func (r *RemoteDriver) UpdateAdminFieldSortOrder(_ context.Context, _ audited.AuditContext, _ db.UpdateAdminFieldSortOrderParams) error {
	return ErrNotSupported{Method: "UpdateAdminFieldSortOrder"}
}

func (r *RemoteDriver) GetMaxAdminSortOrderByParentID(parentID types.NullableAdminDatatypeID) (int64, error) {
	return doRead(r, func() (int64, error) {
		if !parentID.Valid {
			return 0, nil
		}
		// Use the admin fields list and compute max sort order client-side.
		ctx := context.Background()
		params := url.Values{}
		params.Set("parent_id", string(parentID.ID))
		raw, err := r.client.AdminFields.RawList(ctx, params)
		if err != nil {
			return 0, fmt.Errorf("remote: GetMaxAdminSortOrderByParentID: %w", err)
		}
		var sdkItems []modula.AdminField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return 0, fmt.Errorf("remote: GetMaxAdminSortOrderByParentID: decode: %w", err)
		}
		var maxOrder int64
		for _, f := range sdkItems {
			if f.SortOrder > maxOrder {
				maxOrder = f.SortOrder
			}
		}
		return maxOrder, nil
	})
}

// ---------------------------------------------------------------------------
// AdminFieldTypes
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountAdminFieldTypes() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.AdminFieldTypes.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountAdminFieldTypes: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateAdminFieldType(ctx context.Context, _ audited.AuditContext, params db.CreateAdminFieldTypeParams) (*db.AdminFieldTypes, error) {
	return doWrite(r, func() (*db.AdminFieldTypes, error) {
		sdkParams := adminFieldTypeCreateFromDb(params)
		result, err := r.client.AdminFieldTypes.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateAdminFieldType: %w", err)
		}
		row := adminFieldTypeToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateAdminFieldTypeTable() error {
	return ErrNotSupported{Method: "CreateAdminFieldTypeTable"}
}

func (r *RemoteDriver) DeleteAdminFieldType(ctx context.Context, _ audited.AuditContext, id types.AdminFieldTypeID) error {
	return doWriteErr(r, func() error {
		if err := r.client.AdminFieldTypes.Delete(ctx, modula.AdminFieldTypeID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteAdminFieldType: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetAdminFieldType(id types.AdminFieldTypeID) (*db.AdminFieldTypes, error) {
	return doRead(r, func() (*db.AdminFieldTypes, error) {
		ctx := context.Background()
		sdkID := modula.AdminFieldTypeID(string(id))
		item, err := r.client.AdminFieldTypes.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetAdminFieldType: %w", err)
		}
		result := adminFieldTypeToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetAdminFieldTypeByType(t string) (*db.AdminFieldTypes, error) {
	return doRead(r, func() (*db.AdminFieldTypes, error) {
		ctx := context.Background()
		items, err := r.client.AdminFieldTypes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetAdminFieldTypeByType: %w", err)
		}
		for i := range items {
			if items[i].Type == t {
				result := adminFieldTypeToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetAdminFieldTypeByType: not found: %s", t)
	})
}

func (r *RemoteDriver) ListAdminFieldTypes() (*[]db.AdminFieldTypes, error) {
	return doRead(r, func() (*[]db.AdminFieldTypes, error) {
		ctx := context.Background()
		items, err := r.client.AdminFieldTypes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminFieldTypes: %w", err)
		}
		result := make([]db.AdminFieldTypes, len(items))
		for i := range items {
			result[i] = adminFieldTypeToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateAdminFieldType(ctx context.Context, _ audited.AuditContext, params db.UpdateAdminFieldTypeParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := adminFieldTypeUpdateFromDb(params)
		if _, err := r.client.AdminFieldTypes.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateAdminFieldType: %w", err)
		}
		id := string(params.AdminFieldTypeID)
		return &id, nil
	})
}

// ---------------------------------------------------------------------------
// AdminRoutes
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountAdminRoutes() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.AdminRoutes.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountAdminRoutes: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateAdminRoute(ctx context.Context, _ audited.AuditContext, params db.CreateAdminRouteParams) (*db.AdminRoutes, error) {
	return doWrite(r, func() (*db.AdminRoutes, error) {
		sdkParams := adminRouteCreateFromDb(params)
		result, err := r.client.AdminRoutes.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateAdminRoute: %w", err)
		}
		row := adminRouteToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateAdminRouteTable() error {
	return ErrNotSupported{Method: "CreateAdminRouteTable"}
}

func (r *RemoteDriver) DeleteAdminRoute(ctx context.Context, _ audited.AuditContext, id types.AdminRouteID) error {
	return doWriteErr(r, func() error {
		if err := r.client.AdminRoutes.Delete(ctx, modula.AdminRouteID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteAdminRoute: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetAdminRoute(slug types.Slug) (*db.AdminRoutes, error) {
	return doRead(r, func() (*db.AdminRoutes, error) {
		ctx := context.Background()
		// Admin routes are keyed by slug; list and filter client-side.
		items, err := r.client.AdminRoutes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetAdminRoute: %w", err)
		}
		for i := range items {
			if string(items[i].Slug) == string(slug) {
				result := adminRouteToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetAdminRoute: not found: %s", string(slug))
	})
}

func (r *RemoteDriver) ListAdminRoutes() (*[]db.AdminRoutes, error) {
	return doRead(r, func() (*[]db.AdminRoutes, error) {
		ctx := context.Background()
		items, err := r.client.AdminRoutes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminRoutes: %w", err)
		}
		result := make([]db.AdminRoutes, len(items))
		for i := range items {
			result[i] = adminRouteToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListAdminRoutesPaginated(p db.PaginationParams) (*[]db.AdminRoutes, error) {
	return doRead(r, func() (*[]db.AdminRoutes, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.AdminRoutes.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListAdminRoutesPaginated: %w", err)
		}
		var sdkItems []modula.AdminRoute
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListAdminRoutesPaginated: decode: %w", err)
		}
		result := make([]db.AdminRoutes, len(sdkItems))
		for i := range sdkItems {
			result[i] = adminRouteToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateAdminRoute(ctx context.Context, _ audited.AuditContext, params db.UpdateAdminRouteParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := adminRouteUpdateFromDb(params)
		if _, err := r.client.AdminRoutes.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateAdminRoute: %w", err)
		}
		id := string(params.Slug)
		return &id, nil
	})
}

// ---------------------------------------------------------------------------
// Backups
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountBackups() (*int64, error) {
	return nil, ErrNotSupported{Method: "CountBackups"}
}

func (r *RemoteDriver) CreateBackup(_ db.CreateBackupParams) (*db.Backup, error) {
	return nil, ErrNotSupported{Method: "CreateBackup"}
}

func (r *RemoteDriver) CreateBackupTables() error {
	return ErrNotSupported{Method: "CreateBackupTables"}
}

func (r *RemoteDriver) DeleteBackup(_ types.BackupID) error {
	return ErrNotSupported{Method: "DeleteBackup"}
}

func (r *RemoteDriver) DropBackupTables() error {
	return ErrNotSupported{Method: "DropBackupTables"}
}

func (r *RemoteDriver) GetBackup(_ types.BackupID) (*db.Backup, error) {
	return nil, ErrNotSupported{Method: "GetBackup"}
}

func (r *RemoteDriver) GetLatestBackup(_ types.NodeID) (*db.Backup, error) {
	return nil, ErrNotSupported{Method: "GetLatestBackup"}
}

func (r *RemoteDriver) ListBackups(_ db.ListBackupsParams) (*[]db.Backup, error) {
	return nil, ErrNotSupported{Method: "ListBackups"}
}

func (r *RemoteDriver) UpdateBackupStatus(_ db.UpdateBackupStatusParams) error {
	return ErrNotSupported{Method: "UpdateBackupStatus"}
}

// ---------------------------------------------------------------------------
// BackupSets
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountBackupSets() (*int64, error) {
	return nil, ErrNotSupported{Method: "CountBackupSets"}
}

func (r *RemoteDriver) CreateBackupSet(_ db.CreateBackupSetParams) (*db.BackupSet, error) {
	return nil, ErrNotSupported{Method: "CreateBackupSet"}
}

func (r *RemoteDriver) GetBackupSet(_ types.BackupSetID) (*db.BackupSet, error) {
	return nil, ErrNotSupported{Method: "GetBackupSet"}
}

func (r *RemoteDriver) GetPendingBackupSets() (*[]db.BackupSet, error) {
	return nil, ErrNotSupported{Method: "GetPendingBackupSets"}
}

// ---------------------------------------------------------------------------
// BackupVerifications
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountVerifications() (*int64, error) {
	return nil, ErrNotSupported{Method: "CountVerifications"}
}

func (r *RemoteDriver) CreateVerification(_ db.CreateVerificationParams) (*db.BackupVerification, error) {
	return nil, ErrNotSupported{Method: "CreateVerification"}
}

func (r *RemoteDriver) GetLatestVerification(_ types.BackupID) (*db.BackupVerification, error) {
	return nil, ErrNotSupported{Method: "GetLatestVerification"}
}

func (r *RemoteDriver) GetVerification(_ types.VerificationID) (*db.BackupVerification, error) {
	return nil, ErrNotSupported{Method: "GetVerification"}
}

// ---------------------------------------------------------------------------
// ChangeEvents
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountChangeEvents() (*int64, error) {
	return nil, ErrNotSupported{Method: "CountChangeEvents"}
}

func (r *RemoteDriver) CreateChangeEventsTable() error {
	return ErrNotSupported{Method: "CreateChangeEventsTable"}
}

func (r *RemoteDriver) DeleteChangeEvent(_ types.EventID) error {
	return ErrNotSupported{Method: "DeleteChangeEvent"}
}

func (r *RemoteDriver) DropChangeEventsTable() error {
	return ErrNotSupported{Method: "DropChangeEventsTable"}
}

func (r *RemoteDriver) GetChangeEvent(_ types.EventID) (*db.ChangeEvent, error) {
	return nil, ErrNotSupported{Method: "GetChangeEvent"}
}

func (r *RemoteDriver) GetChangeEventsByRecord(_ string, _ string) (*[]db.ChangeEvent, error) {
	return nil, ErrNotSupported{Method: "GetChangeEventsByRecord"}
}

func (r *RemoteDriver) GetUnconsumedEvents(_ int64) (*[]db.ChangeEvent, error) {
	return nil, ErrNotSupported{Method: "GetUnconsumedEvents"}
}

func (r *RemoteDriver) GetUnsyncedEvents(_ int64) (*[]db.ChangeEvent, error) {
	return nil, ErrNotSupported{Method: "GetUnsyncedEvents"}
}

func (r *RemoteDriver) ListChangeEvents(_ db.ListChangeEventsParams) (*[]db.ChangeEvent, error) {
	return nil, ErrNotSupported{Method: "ListChangeEvents"}
}

func (r *RemoteDriver) MarkEventConsumed(_ types.EventID) error {
	return ErrNotSupported{Method: "MarkEventConsumed"}
}

func (r *RemoteDriver) MarkEventSynced(_ types.EventID) error {
	return ErrNotSupported{Method: "MarkEventSynced"}
}

func (r *RemoteDriver) RecordChangeEvent(_ db.RecordChangeEventParams) (*db.ChangeEvent, error) {
	return nil, ErrNotSupported{Method: "RecordChangeEvent"}
}

// ---------------------------------------------------------------------------
// ContentData
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountContentData() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.ContentData.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountContentData: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateContentData(ctx context.Context, _ audited.AuditContext, params db.CreateContentDataParams) (*db.ContentData, error) {
	return doWrite(r, func() (*db.ContentData, error) {
		sdkParams := contentDataCreateFromDb(params)
		result, err := r.client.ContentData.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateContentData: %w", err)
		}
		row := contentDataToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateContentDataTable() error {
	return ErrNotSupported{Method: "CreateContentDataTable"}
}

func (r *RemoteDriver) DeleteContentData(ctx context.Context, _ audited.AuditContext, id types.ContentID) error {
	return doWriteErr(r, func() error {
		if err := r.client.ContentData.Delete(ctx, modula.ContentID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteContentData: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetContentData(id types.ContentID) (*db.ContentData, error) {
	return doRead(r, func() (*db.ContentData, error) {
		ctx := context.Background()
		sdkID := modula.ContentID(string(id))
		item, err := r.client.ContentData.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetContentData: %w", err)
		}
		result := contentDataToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentData() (*[]db.ContentData, error) {
	return doRead(r, func() (*[]db.ContentData, error) {
		ctx := context.Background()
		items, err := r.client.ContentData.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentData: %w", err)
		}
		result := make([]db.ContentData, len(items))
		for i := range items {
			result[i] = contentDataToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentDataByRoute(routeID types.NullableRouteID) (*[]db.ContentData, error) {
	return doRead(r, func() (*[]db.ContentData, error) {
		ctx := context.Background()
		params := url.Values{}
		if routeID.Valid {
			params.Set("route_id", string(routeID.ID))
		}
		raw, err := r.client.ContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentDataByRoute: %w", err)
		}
		var sdkItems []modula.ContentData
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentDataByRoute: decode: %w", err)
		}
		result := make([]db.ContentData, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentDataToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentDataByDatatypeID(datatypeID types.DatatypeID) (*[]db.ContentData, error) {
	return doRead(r, func() (*[]db.ContentData, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("datatype_id", string(datatypeID))
		raw, err := r.client.ContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentDataByDatatypeID: %w", err)
		}
		var sdkItems []modula.ContentData
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentDataByDatatypeID: decode: %w", err)
		}
		result := make([]db.ContentData, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentDataToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentDataGlobal() (*[]db.ContentData, error) {
	return nil, ErrNotSupported{Method: "ListContentDataGlobal"}
}

func (r *RemoteDriver) ReassignContentDataAuthor(_ context.Context, _, _ types.UserID) error {
	return ErrNotSupported{Method: "ReassignContentDataAuthor"}
}

func (r *RemoteDriver) CountContentDataByAuthor(_ context.Context, _ types.UserID) (int64, error) {
	return 0, ErrNotSupported{Method: "CountContentDataByAuthor"}
}

func (r *RemoteDriver) ReassignDatatypeAuthor(_ context.Context, _, _ types.UserID) error {
	return ErrNotSupported{Method: "ReassignDatatypeAuthor"}
}

func (r *RemoteDriver) CountDatatypesByAuthor(_ context.Context, _ types.UserID) (int64, error) {
	return 0, ErrNotSupported{Method: "CountDatatypesByAuthor"}
}

func (r *RemoteDriver) ReassignAdminContentDataAuthor(_ context.Context, _, _ types.UserID) error {
	return ErrNotSupported{Method: "ReassignAdminContentDataAuthor"}
}

func (r *RemoteDriver) CountAdminContentDataByAuthor(_ context.Context, _ types.UserID) (int64, error) {
	return 0, ErrNotSupported{Method: "CountAdminContentDataByAuthor"}
}

func (r *RemoteDriver) ListContentDataPaginated(p db.PaginationParams) (*[]db.ContentData, error) {
	return doRead(r, func() (*[]db.ContentData, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.ContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentDataPaginated: %w", err)
		}
		var sdkItems []modula.ContentData
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentDataPaginated: decode: %w", err)
		}
		result := make([]db.ContentData, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentDataToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentDataTopLevelPaginated(p db.PaginationParams) (*[]db.ContentDataTopLevel, error) {
	return doRead(r, func() (*[]db.ContentDataTopLevel, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		params.Set("top_level", "true")
		raw, err := r.client.ContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentDataTopLevelPaginated: %w", err)
		}
		var sdkItems []modula.ContentData
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentDataTopLevelPaginated: decode: %w", err)
		}
		result := make([]db.ContentDataTopLevel, len(sdkItems))
		for i := range sdkItems {
			result[i] = db.ContentDataTopLevel{
				ContentData: contentDataToDb(&sdkItems[i]),
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentDataTopLevelPaginatedByStatus(p db.PaginationParams, status types.ContentStatus) (*[]db.ContentDataTopLevel, error) {
	return doRead(r, func() (*[]db.ContentDataTopLevel, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		params.Set("top_level", "true")
		params.Set("status", string(status))
		raw, err := r.client.ContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentDataTopLevelPaginatedByStatus: %w", err)
		}
		var sdkItems []modula.ContentData
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentDataTopLevelPaginatedByStatus: decode: %w", err)
		}
		result := make([]db.ContentDataTopLevel, len(sdkItems))
		for i := range sdkItems {
			result[i] = db.ContentDataTopLevel{
				ContentData: contentDataToDb(&sdkItems[i]),
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentDataByRoutePaginated(p db.ListContentDataByRoutePaginatedParams) (*[]db.ContentData, error) {
	return doRead(r, func() (*[]db.ContentData, error) {
		ctx := context.Background()
		params := url.Values{}
		if p.RouteID.Valid {
			params.Set("route_id", string(p.RouteID.ID))
		}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.ContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentDataByRoutePaginated: %w", err)
		}
		var sdkItems []modula.ContentData
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentDataByRoutePaginated: decode: %w", err)
		}
		result := make([]db.ContentData, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentDataToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) CountContentDataTopLevel() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("count", "true")
		params.Set("top_level", "true")
		raw, err := r.client.ContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: CountContentDataTopLevel: %w", err)
		}
		var countResp struct {
			Count int64 `json:"count"`
		}
		if err := json.Unmarshal(raw, &countResp); err != nil {
			return nil, fmt.Errorf("remote: CountContentDataTopLevel: decode: %w", err)
		}
		return &countResp.Count, nil
	})
}

func (r *RemoteDriver) CountContentDataTopLevelByStatus(status types.ContentStatus) (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("count", "true")
		params.Set("top_level", "true")
		params.Set("status", string(status))
		raw, err := r.client.ContentData.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: CountContentDataTopLevelByStatus: %w", err)
		}
		var countResp struct {
			Count int64 `json:"count"`
		}
		if err := json.Unmarshal(raw, &countResp); err != nil {
			return nil, fmt.Errorf("remote: CountContentDataTopLevelByStatus: decode: %w", err)
		}
		return &countResp.Count, nil
	})
}

func (r *RemoteDriver) GetContentDataDescendants(_ context.Context, _ types.ContentID) (*[]db.ContentData, error) {
	return nil, ErrNotSupported{Method: "GetContentDataDescendants"}
}

func (r *RemoteDriver) ListRootContentSummary() (*[]db.RootContentSummary, error) {
	return nil, ErrNotSupported{Method: "ListRootContentSummary"}
}

func (r *RemoteDriver) UpdateContentData(ctx context.Context, _ audited.AuditContext, params db.UpdateContentDataParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := contentDataUpdateFromDb(params)
		if _, err := r.client.ContentData.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateContentData: %w", err)
		}
		id := string(params.ContentDataID)
		return &id, nil
	})
}

func (r *RemoteDriver) UpdateContentDataPublishMeta(_ context.Context, _ db.UpdateContentDataPublishMetaParams) error {
	return ErrNotSupported{Method: "UpdateContentDataPublishMeta"}
}

func (r *RemoteDriver) UpdateContentDataWithRevision(_ context.Context, _ db.UpdateContentDataWithRevisionParams) error {
	return ErrNotSupported{Method: "UpdateContentDataWithRevision"}
}

func (r *RemoteDriver) UpdateContentDataSchedule(_ context.Context, _ db.UpdateContentDataScheduleParams) error {
	return ErrNotSupported{Method: "UpdateContentDataSchedule"}
}

func (r *RemoteDriver) ClearContentDataSchedule(_ context.Context, _ db.ClearContentDataScheduleParams) error {
	return ErrNotSupported{Method: "ClearContentDataSchedule"}
}

func (r *RemoteDriver) ListContentDataDueForPublish(_ types.Timestamp) (*[]db.ContentData, error) {
	return nil, ErrNotSupported{Method: "ListContentDataDueForPublish"}
}

func (r *RemoteDriver) ListContentDataByRootID(_ types.NullableContentID) (*[]db.ContentData, error) {
	return nil, ErrNotSupported{Method: "ListContentDataByRootID"}
}

// ---------------------------------------------------------------------------
// ContentFields
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountContentFields() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.ContentFields.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountContentFields: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateContentField(ctx context.Context, _ audited.AuditContext, params db.CreateContentFieldParams) (*db.ContentFields, error) {
	return doWrite(r, func() (*db.ContentFields, error) {
		sdkParams := contentFieldCreateFromDb(params)
		result, err := r.client.ContentFields.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateContentField: %w", err)
		}
		row := contentFieldToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateContentFieldTable() error {
	return ErrNotSupported{Method: "CreateContentFieldTable"}
}

func (r *RemoteDriver) DeleteContentField(ctx context.Context, _ audited.AuditContext, id types.ContentFieldID) error {
	return doWriteErr(r, func() error {
		if err := r.client.ContentFields.Delete(ctx, modula.ContentFieldID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteContentField: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetContentField(id types.ContentFieldID) (*db.ContentFields, error) {
	return doRead(r, func() (*db.ContentFields, error) {
		ctx := context.Background()
		sdkID := modula.ContentFieldID(string(id))
		item, err := r.client.ContentFields.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetContentField: %w", err)
		}
		result := contentFieldToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetContentFieldsByRoute(_ types.NullableRouteID) (*[]db.GetContentFieldsByRouteRow, error) {
	return nil, ErrNotSupported{Method: "GetContentFieldsByRoute"}
}

func (r *RemoteDriver) ListContentFields() (*[]db.ContentFields, error) {
	return doRead(r, func() (*[]db.ContentFields, error) {
		ctx := context.Background()
		items, err := r.client.ContentFields.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentFields: %w", err)
		}
		result := make([]db.ContentFields, len(items))
		for i := range items {
			result[i] = contentFieldToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentFieldsByRoute(routeID types.NullableRouteID) (*[]db.ContentFields, error) {
	return doRead(r, func() (*[]db.ContentFields, error) {
		ctx := context.Background()
		params := url.Values{}
		if routeID.Valid {
			params.Set("route_id", string(routeID.ID))
		}
		raw, err := r.client.ContentFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsByRoute: %w", err)
		}
		var sdkItems []modula.ContentField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsByRoute: decode: %w", err)
		}
		result := make([]db.ContentFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentFieldsByContentData(contentDataID types.NullableContentID) (*[]db.ContentFields, error) {
	return doRead(r, func() (*[]db.ContentFields, error) {
		ctx := context.Background()
		params := url.Values{}
		if contentDataID.Valid {
			params.Set("content_data_id", string(contentDataID.ID))
		}
		raw, err := r.client.ContentFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsByContentData: %w", err)
		}
		var sdkItems []modula.ContentField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsByContentData: decode: %w", err)
		}
		result := make([]db.ContentFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentFieldsPaginated(p db.PaginationParams) (*[]db.ContentFields, error) {
	return doRead(r, func() (*[]db.ContentFields, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.ContentFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsPaginated: %w", err)
		}
		var sdkItems []modula.ContentField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsPaginated: decode: %w", err)
		}
		result := make([]db.ContentFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentFieldsByRoutePaginated(p db.ListContentFieldsByRoutePaginatedParams) (*[]db.ContentFields, error) {
	return doRead(r, func() (*[]db.ContentFields, error) {
		ctx := context.Background()
		params := url.Values{}
		if p.RouteID.Valid {
			params.Set("route_id", string(p.RouteID.ID))
		}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.ContentFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsByRoutePaginated: %w", err)
		}
		var sdkItems []modula.ContentField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsByRoutePaginated: decode: %w", err)
		}
		result := make([]db.ContentFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentFieldsByContentDataPaginated(p db.ListContentFieldsByContentDataPaginatedParams) (*[]db.ContentFields, error) {
	return doRead(r, func() (*[]db.ContentFields, error) {
		ctx := context.Background()
		params := url.Values{}
		if p.ContentDataID.Valid {
			params.Set("content_data_id", string(p.ContentDataID.ID))
		}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.ContentFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsByContentDataPaginated: %w", err)
		}
		var sdkItems []modula.ContentField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsByContentDataPaginated: decode: %w", err)
		}
		result := make([]db.ContentFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentFieldsWithFieldByContentData(contentDataID types.NullableContentID) (*[]db.ContentFieldWithFieldRow, error) {
	return doRead(r, func() (*[]db.ContentFieldWithFieldRow, error) {
		ctx := context.Background()
		// Fetch content fields filtered by content data ID.
		cfParams := url.Values{}
		if contentDataID.Valid {
			cfParams.Set("content_data_id", string(contentDataID.ID))
		}
		cfRaw, err := r.client.ContentFields.RawList(ctx, cfParams)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsWithFieldByContentData: list content fields: %w", err)
		}
		var sdkCFs []modula.ContentField
		if err := json.Unmarshal(cfRaw, &sdkCFs); err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsWithFieldByContentData: decode content fields: %w", err)
		}
		// Collect unique field IDs to fetch field definitions.
		fieldMap := make(map[string]*modula.Field)
		for _, cf := range sdkCFs {
			if cf.FieldID != nil {
				fieldMap[string(*cf.FieldID)] = nil
			}
		}
		// Fetch each field definition.
		for fid := range fieldMap {
			f, fErr := r.client.Fields.Get(ctx, modula.FieldID(fid))
			if fErr != nil {
				continue // field may have been deleted; leave nil
			}
			fieldMap[fid] = f
		}
		// Join content fields with field definitions.
		result := make([]db.ContentFieldWithFieldRow, len(sdkCFs))
		for i, cf := range sdkCFs {
			dbCF := contentFieldToDb(&cf)
			row := db.ContentFieldWithFieldRow{
				ContentFieldID: dbCF.ContentFieldID,
				RouteID:        dbCF.RouteID,
				ContentDataID:  dbCF.ContentDataID,
				FieldID:        dbCF.FieldID,
				FieldValue:     dbCF.FieldValue,
				AuthorID:       dbCF.AuthorID,
				DateCreated:    dbCF.DateCreated,
				DateModified:   dbCF.DateModified,
			}
			if cf.FieldID != nil {
				if f, ok := fieldMap[string(*cf.FieldID)]; ok && f != nil {
					row.FFieldID = types.FieldID(string(f.FieldID))
					row.FLabel = f.Label
					row.FType = types.FieldType(string(f.Type))
				}
			}
			result[i] = row
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentFieldsByContentDataAndLocale(contentDataID types.NullableContentID, locale string) (*[]db.ContentFields, error) {
	return doRead(r, func() (*[]db.ContentFields, error) {
		ctx := context.Background()
		params := url.Values{}
		if contentDataID.Valid {
			params.Set("content_data_id", string(contentDataID.ID))
		}
		params.Set("locale", locale)
		raw, err := r.client.ContentFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsByContentDataAndLocale: %w", err)
		}
		var sdkItems []modula.ContentField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsByContentDataAndLocale: decode: %w", err)
		}
		result := make([]db.ContentFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentFieldsByContentDataIDs(_ context.Context, _ []types.ContentID, _ string) (*[]db.ContentFields, error) {
	return nil, ErrNotSupported{Method: "ListContentFieldsByContentDataIDs"}
}

func (r *RemoteDriver) ListContentFieldsByRouteAndLocale(routeID types.NullableRouteID, locale string) (*[]db.ContentFields, error) {
	return doRead(r, func() (*[]db.ContentFields, error) {
		ctx := context.Background()
		params := url.Values{}
		if routeID.Valid {
			params.Set("route_id", string(routeID.ID))
		}
		params.Set("locale", locale)
		raw, err := r.client.ContentFields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsByRouteAndLocale: %w", err)
		}
		var sdkItems []modula.ContentField
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentFieldsByRouteAndLocale: decode: %w", err)
		}
		result := make([]db.ContentFields, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentFieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateContentField(ctx context.Context, _ audited.AuditContext, params db.UpdateContentFieldParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := contentFieldUpdateFromDb(params)
		if _, err := r.client.ContentFields.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateContentField: %w", err)
		}
		id := string(params.ContentFieldID)
		return &id, nil
	})
}

func (r *RemoteDriver) ListContentFieldsByRootID(_ types.NullableContentID) (*[]db.ContentFields, error) {
	return nil, ErrNotSupported{Method: "ListContentFieldsByRootID"}
}

func (r *RemoteDriver) ListContentFieldsByRootIDAndLocale(_ types.NullableContentID, _ string) (*[]db.ContentFields, error) {
	return nil, ErrNotSupported{Method: "ListContentFieldsByRootIDAndLocale"}
}

// ---------------------------------------------------------------------------
// ContentRelations
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountContentRelations() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.ContentRelations.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountContentRelations: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateContentRelation(ctx context.Context, _ audited.AuditContext, params db.CreateContentRelationParams) (*db.ContentRelations, error) {
	return doWrite(r, func() (*db.ContentRelations, error) {
		sdkParams := modula.CreateContentRelationParams{
			SourceContentID: modula.ContentID(string(params.SourceContentID)),
			TargetContentID: modula.ContentID(string(params.TargetContentID)),
			FieldID:         modula.FieldID(string(params.FieldID)),
			SortOrder:       params.SortOrder,
		}
		result, err := r.client.ContentRelations.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateContentRelation: %w", err)
		}
		row := contentRelationToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateContentRelationTable() error {
	return ErrNotSupported{Method: "CreateContentRelationTable"}
}

func (r *RemoteDriver) DeleteContentRelation(ctx context.Context, _ audited.AuditContext, id types.ContentRelationID) error {
	return doWriteErr(r, func() error {
		if err := r.client.ContentRelations.Delete(ctx, modula.ContentRelationID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteContentRelation: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) DropContentRelationTable() error {
	return ErrNotSupported{Method: "DropContentRelationTable"}
}

func (r *RemoteDriver) GetContentRelation(id types.ContentRelationID) (*db.ContentRelations, error) {
	return doRead(r, func() (*db.ContentRelations, error) {
		ctx := context.Background()
		sdkID := modula.ContentRelationID(string(id))
		item, err := r.client.ContentRelations.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetContentRelation: %w", err)
		}
		result := contentRelationToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentRelations() (*[]db.ContentRelations, error) {
	return doRead(r, func() (*[]db.ContentRelations, error) {
		ctx := context.Background()
		items, err := r.client.ContentRelations.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentRelations: %w", err)
		}
		result := make([]db.ContentRelations, len(items))
		for i := range items {
			result[i] = contentRelationToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentRelationsBySource(sourceID types.ContentID) (*[]db.ContentRelations, error) {
	return doRead(r, func() (*[]db.ContentRelations, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("source_content_id", string(sourceID))
		raw, err := r.client.ContentRelations.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentRelationsBySource: %w", err)
		}
		var sdkItems []modula.ContentRelation
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentRelationsBySource: decode: %w", err)
		}
		result := make([]db.ContentRelations, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentRelationToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentRelationsByTarget(targetID types.ContentID) (*[]db.ContentRelations, error) {
	return doRead(r, func() (*[]db.ContentRelations, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("target_content_id", string(targetID))
		raw, err := r.client.ContentRelations.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentRelationsByTarget: %w", err)
		}
		var sdkItems []modula.ContentRelation
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentRelationsByTarget: decode: %w", err)
		}
		result := make([]db.ContentRelations, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentRelationToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentRelationsBySourceAndField(sourceID types.ContentID, fieldID types.FieldID) (*[]db.ContentRelations, error) {
	return doRead(r, func() (*[]db.ContentRelations, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("source_content_id", string(sourceID))
		params.Set("field_id", string(fieldID))
		raw, err := r.client.ContentRelations.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentRelationsBySourceAndField: %w", err)
		}
		var sdkItems []modula.ContentRelation
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListContentRelationsBySourceAndField: decode: %w", err)
		}
		result := make([]db.ContentRelations, len(sdkItems))
		for i := range sdkItems {
			result[i] = contentRelationToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateContentRelationSortOrder(ctx context.Context, _ audited.AuditContext, params db.UpdateContentRelationSortOrderParams) error {
	return doWriteErr(r, func() error {
		sdkParams := modula.UpdateContentRelationParams{
			ContentRelationID: modula.ContentRelationID(string(params.ContentRelationID)),
			SortOrder:         params.SortOrder,
		}
		if _, err := r.client.ContentRelations.Update(ctx, sdkParams); err != nil {
			return fmt.Errorf("remote: UpdateContentRelationSortOrder: %w", err)
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// ContentVersions
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountContentVersions() (*int64, error) {
	return nil, ErrNotSupported{Method: "CountContentVersions"}
}

func (r *RemoteDriver) CountContentVersionsByContent(_ types.ContentID) (*int64, error) {
	return nil, ErrNotSupported{Method: "CountContentVersionsByContent"}
}

func (r *RemoteDriver) CreateContentVersion(_ context.Context, _ audited.AuditContext, _ db.CreateContentVersionParams) (*db.ContentVersion, error) {
	return nil, ErrNotSupported{Method: "CreateContentVersion"}
}

func (r *RemoteDriver) CreateContentVersionTable() error {
	return ErrNotSupported{Method: "CreateContentVersionTable"}
}

func (r *RemoteDriver) DropContentVersionTable() error {
	return ErrNotSupported{Method: "DropContentVersionTable"}
}

func (r *RemoteDriver) DeleteContentVersion(ctx context.Context, _ audited.AuditContext, id types.ContentVersionID) error {
	return doWriteErr(r, func() error {
		if err := r.client.Publishing.DeleteVersion(ctx, string(id)); err != nil {
			return fmt.Errorf("remote: DeleteContentVersion: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetContentVersion(id types.ContentVersionID) (*db.ContentVersion, error) {
	return doRead(r, func() (*db.ContentVersion, error) {
		ctx := context.Background()
		item, err := r.client.Publishing.GetVersion(ctx, string(id))
		if err != nil {
			return nil, fmt.Errorf("remote: GetContentVersion: %w", err)
		}
		row := contentVersionToDb(item)
		return &row, nil
	})
}

func (r *RemoteDriver) GetPublishedSnapshot(_ types.ContentID, _ string) (*db.ContentVersion, error) {
	return nil, ErrNotSupported{Method: "GetPublishedSnapshot"}
}

func (r *RemoteDriver) ListContentVersionsByContent(contentID types.ContentID) (*[]db.ContentVersion, error) {
	return doRead(r, func() (*[]db.ContentVersion, error) {
		ctx := context.Background()
		sdkID := modula.ContentID(string(contentID))
		items, err := r.client.ContentVersions.ListByContent(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: ListContentVersionsByContent: %w", err)
		}
		result := make([]db.ContentVersion, len(items))
		for i := range items {
			result[i] = contentVersionToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListContentVersionsByContentLocale(_ types.ContentID, _ string) (*[]db.ContentVersion, error) {
	return nil, ErrNotSupported{Method: "ListContentVersionsByContentLocale"}
}

func (r *RemoteDriver) ClearPublishedFlag(_ types.ContentID, _ string) error {
	return ErrNotSupported{Method: "ClearPublishedFlag"}
}

func (r *RemoteDriver) GetMaxVersionNumber(_ types.ContentID, _ string) (int64, error) {
	return 0, ErrNotSupported{Method: "GetMaxVersionNumber"}
}

func (r *RemoteDriver) PruneOldVersions(_ types.ContentID, _ string, _ int64) error {
	return ErrNotSupported{Method: "PruneOldVersions"}
}

// ---------------------------------------------------------------------------
// Datatypes
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountDatatypes() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.Datatypes.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountDatatypes: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateDatatype(ctx context.Context, _ audited.AuditContext, params db.CreateDatatypeParams) (*db.Datatypes, error) {
	return doWrite(r, func() (*db.Datatypes, error) {
		sdkParams := datatypeCreateFromDb(params)
		result, err := r.client.Datatypes.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateDatatype: %w", err)
		}
		row := datatypeToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateDatatypeTable() error {
	return ErrNotSupported{Method: "CreateDatatypeTable"}
}

func (r *RemoteDriver) DeleteDatatype(ctx context.Context, _ audited.AuditContext, id types.DatatypeID) error {
	return doWriteErr(r, func() error {
		if err := r.client.Datatypes.Delete(ctx, modula.DatatypeID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteDatatype: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetDatatype(id types.DatatypeID) (*db.Datatypes, error) {
	return doRead(r, func() (*db.Datatypes, error) {
		ctx := context.Background()
		sdkID := modula.DatatypeID(string(id))
		item, err := r.client.Datatypes.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetDatatype: %w", err)
		}
		result := datatypeToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetDatatypeByType(t string) (*db.Datatypes, error) {
	return doRead(r, func() (*db.Datatypes, error) {
		ctx := context.Background()
		items, err := r.client.Datatypes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetDatatypeByType: %w", err)
		}
		for i := range items {
			if items[i].Type == t {
				result := datatypeToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetDatatypeByType: not found: %s", t)
	})
}

func (r *RemoteDriver) GetDatatypeByName(name string) (*db.Datatypes, error) {
	return doRead(r, func() (*db.Datatypes, error) {
		ctx := context.Background()
		items, err := r.client.Datatypes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetDatatypeByName: %w", err)
		}
		for i := range items {
			if items[i].Name == name {
				result := datatypeToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetDatatypeByName: not found: %s", name)
	})
}

func (r *RemoteDriver) ListDatatypes() (*[]db.Datatypes, error) {
	return doRead(r, func() (*[]db.Datatypes, error) {
		ctx := context.Background()
		items, err := r.client.Datatypes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListDatatypes: %w", err)
		}
		result := make([]db.Datatypes, len(items))
		for i := range items {
			result[i] = datatypeToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListDatatypesRoot() (*[]db.Datatypes, error) {
	return doRead(r, func() (*[]db.Datatypes, error) {
		ctx := context.Background()
		items, err := r.client.Datatypes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListDatatypesRoot: %w", err)
		}
		var result []db.Datatypes
		for i := range items {
			if items[i].ParentID == nil {
				result = append(result, datatypeToDb(&items[i]))
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListDatatypesGlobal() (*[]db.Datatypes, error) {
	return nil, ErrNotSupported{Method: "ListDatatypesGlobal"}
}

func (r *RemoteDriver) ListDatatypeChildren(parentID types.DatatypeID) (*[]db.Datatypes, error) {
	return doRead(r, func() (*[]db.Datatypes, error) {
		ctx := context.Background()
		items, err := r.client.Datatypes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListDatatypeChildren: %w", err)
		}
		var result []db.Datatypes
		for i := range items {
			if items[i].ParentID != nil && string(*items[i].ParentID) == string(parentID) {
				result = append(result, datatypeToDb(&items[i]))
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListDatatypesPaginated(p db.PaginationParams) (*[]db.Datatypes, error) {
	return doRead(r, func() (*[]db.Datatypes, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.Datatypes.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListDatatypesPaginated: %w", err)
		}
		var sdkItems []modula.Datatype
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListDatatypesPaginated: decode: %w", err)
		}
		result := make([]db.Datatypes, len(sdkItems))
		for i := range sdkItems {
			result[i] = datatypeToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListDatatypeChildrenPaginated(p db.ListDatatypeChildrenPaginatedParams) (*[]db.Datatypes, error) {
	return doRead(r, func() (*[]db.Datatypes, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("parent_id", string(p.ParentID))
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.Datatypes.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListDatatypeChildrenPaginated: %w", err)
		}
		var sdkItems []modula.Datatype
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListDatatypeChildrenPaginated: decode: %w", err)
		}
		result := make([]db.Datatypes, len(sdkItems))
		for i := range sdkItems {
			result[i] = datatypeToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateDatatype(ctx context.Context, _ audited.AuditContext, params db.UpdateDatatypeParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := datatypeUpdateFromDb(params)
		if _, err := r.client.Datatypes.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateDatatype: %w", err)
		}
		id := string(params.DatatypeID)
		return &id, nil
	})
}

// ---------------------------------------------------------------------------
// Fields
// ---------------------------------------------------------------------------

func (r *RemoteDriver) ListFieldsWithSortOrderByDatatypeID(parentID types.NullableDatatypeID) (*[]db.FieldWithSortOrderRow, error) {
	return doRead(r, func() (*[]db.FieldWithSortOrderRow, error) {
		ctx := context.Background()
		params := url.Values{}
		if parentID.Valid {
			params.Set("parent_id", string(parentID.ID))
		}
		raw, err := r.client.Fields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListFieldsWithSortOrderByDatatypeID: %w", err)
		}
		var sdkItems []modula.Field
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListFieldsWithSortOrderByDatatypeID: decode: %w", err)
		}
		result := make([]db.FieldWithSortOrderRow, len(sdkItems))
		for i, f := range sdkItems {
			result[i] = db.FieldWithSortOrderRow{
				SortOrder:  f.SortOrder,
				FieldID:    types.FieldID(string(f.FieldID)),
				Label:      f.Label,
				Type:       types.FieldType(string(f.Type)),
				Data:       f.Data,
				Validation: f.Validation,
				UIConfig:   f.UIConfig,
				Roles:      rolesToNullableString(f.Roles),
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) CountFields() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.Fields.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountFields: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateField(ctx context.Context, _ audited.AuditContext, params db.CreateFieldParams) (*db.Fields, error) {
	return doWrite(r, func() (*db.Fields, error) {
		sdkParams := fieldCreateFromDb(params)
		result, err := r.client.Fields.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateField: %w", err)
		}
		row := fieldToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateFieldTable() error {
	return ErrNotSupported{Method: "CreateFieldTable"}
}

func (r *RemoteDriver) DeleteField(ctx context.Context, _ audited.AuditContext, id types.FieldID) error {
	return doWriteErr(r, func() error {
		if err := r.client.Fields.Delete(ctx, modula.FieldID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteField: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetField(id types.FieldID) (*db.Fields, error) {
	return doRead(r, func() (*db.Fields, error) {
		ctx := context.Background()
		sdkID := modula.FieldID(string(id))
		item, err := r.client.Fields.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetField: %w", err)
		}
		result := fieldToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetFieldsByIDs(_ context.Context, _ []types.FieldID) ([]db.Fields, error) {
	return nil, ErrNotSupported{Method: "GetFieldsByIDs"}
}

func (r *RemoteDriver) GetFieldDefinitionsByRoute(_ types.NullableRouteID) (*[]db.GetFieldDefinitionsByRouteRow, error) {
	return nil, ErrNotSupported{Method: "GetFieldDefinitionsByRoute"}
}

func (r *RemoteDriver) ListFields() (*[]db.Fields, error) {
	return doRead(r, func() (*[]db.Fields, error) {
		ctx := context.Background()
		items, err := r.client.Fields.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListFields: %w", err)
		}
		result := make([]db.Fields, len(items))
		for i := range items {
			result[i] = fieldToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListFieldsByDatatypeID(parentID types.NullableDatatypeID) (*[]db.Fields, error) {
	return doRead(r, func() (*[]db.Fields, error) {
		ctx := context.Background()
		params := url.Values{}
		if parentID.Valid {
			params.Set("parent_id", string(parentID.ID))
		}
		raw, err := r.client.Fields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListFieldsByDatatypeID: %w", err)
		}
		var sdkItems []modula.Field
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListFieldsByDatatypeID: decode: %w", err)
		}
		result := make([]db.Fields, len(sdkItems))
		for i := range sdkItems {
			result[i] = fieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListFieldsPaginated(p db.PaginationParams) (*[]db.Fields, error) {
	return doRead(r, func() (*[]db.Fields, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.Fields.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListFieldsPaginated: %w", err)
		}
		var sdkItems []modula.Field
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListFieldsPaginated: decode: %w", err)
		}
		result := make([]db.Fields, len(sdkItems))
		for i := range sdkItems {
			result[i] = fieldToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateField(ctx context.Context, _ audited.AuditContext, params db.UpdateFieldParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := fieldUpdateFromDb(params)
		if _, err := r.client.Fields.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateField: %w", err)
		}
		id := string(params.FieldID)
		return &id, nil
	})
}

func (r *RemoteDriver) UpdateFieldSortOrder(ctx context.Context, _ audited.AuditContext, params db.UpdateFieldSortOrderParams) error {
	return doWriteErr(r, func() error {
		if err := r.client.FieldsExtra.UpdateSortOrder(ctx, modula.FieldID(string(params.FieldID)), params.SortOrder); err != nil {
			return fmt.Errorf("remote: UpdateFieldSortOrder: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetMaxSortOrderByParentID(parentID types.NullableDatatypeID) (int64, error) {
	return doRead(r, func() (int64, error) {
		if !parentID.Valid {
			return 0, nil
		}
		ctx := context.Background()
		sdkDtID := modula.DatatypeID(string(parentID.ID))
		maxOrder, err := r.client.FieldsExtra.MaxSortOrder(ctx, sdkDtID)
		if err != nil {
			return 0, fmt.Errorf("remote: GetMaxSortOrderByParentID: %w", err)
		}
		return maxOrder, nil
	})
}

// ---------------------------------------------------------------------------
// Datatype Sort Order
// ---------------------------------------------------------------------------

func (r *RemoteDriver) UpdateDatatypeSortOrder(ctx context.Context, _ audited.AuditContext, params db.UpdateDatatypeSortOrderParams) error {
	return doWriteErr(r, func() error {
		if err := r.client.DatatypesExtra.UpdateSortOrder(ctx, modula.DatatypeID(string(params.DatatypeID)), params.SortOrder); err != nil {
			return fmt.Errorf("remote: UpdateDatatypeSortOrder: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetMaxDatatypeSortOrder(parentID types.NullableDatatypeID) (int64, error) {
	return doRead(r, func() (int64, error) {
		ctx := context.Background()
		var sdkParent *modula.DatatypeID
		if parentID.Valid {
			id := modula.DatatypeID(string(parentID.ID))
			sdkParent = &id
		}
		maxOrder, err := r.client.DatatypesExtra.MaxSortOrder(ctx, sdkParent)
		if err != nil {
			return 0, fmt.Errorf("remote: GetMaxDatatypeSortOrder: %w", err)
		}
		return maxOrder, nil
	})
}

func (r *RemoteDriver) UpdateAdminDatatypeSortOrder(ctx context.Context, _ audited.AuditContext, params db.UpdateAdminDatatypeSortOrderParams) error {
	return doWriteErr(r, func() error {
		if err := r.client.AdminDatatypesExtra.UpdateSortOrder(ctx, modula.AdminDatatypeID(string(params.AdminDatatypeID)), params.SortOrder); err != nil {
			return fmt.Errorf("remote: UpdateAdminDatatypeSortOrder: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetMaxAdminDatatypeSortOrder(parentID types.NullableAdminDatatypeID) (int64, error) {
	return doRead(r, func() (int64, error) {
		ctx := context.Background()
		var sdkParent *modula.AdminDatatypeID
		if parentID.Valid {
			id := modula.AdminDatatypeID(string(parentID.ID))
			sdkParent = &id
		}
		maxOrder, err := r.client.AdminDatatypesExtra.MaxSortOrder(ctx, sdkParent)
		if err != nil {
			return 0, fmt.Errorf("remote: GetMaxAdminDatatypeSortOrder: %w", err)
		}
		return maxOrder, nil
	})
}

// ---------------------------------------------------------------------------
// FieldTypes
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountFieldTypes() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.FieldTypes.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountFieldTypes: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateFieldType(ctx context.Context, _ audited.AuditContext, params db.CreateFieldTypeParams) (*db.FieldTypes, error) {
	return doWrite(r, func() (*db.FieldTypes, error) {
		sdkParams := fieldTypeCreateFromDb(params)
		result, err := r.client.FieldTypes.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateFieldType: %w", err)
		}
		row := fieldTypeToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateFieldTypeTable() error {
	return ErrNotSupported{Method: "CreateFieldTypeTable"}
}

func (r *RemoteDriver) DeleteFieldType(ctx context.Context, _ audited.AuditContext, id types.FieldTypeID) error {
	return doWriteErr(r, func() error {
		if err := r.client.FieldTypes.Delete(ctx, modula.FieldTypeID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteFieldType: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetFieldType(id types.FieldTypeID) (*db.FieldTypes, error) {
	return doRead(r, func() (*db.FieldTypes, error) {
		ctx := context.Background()
		sdkID := modula.FieldTypeID(string(id))
		item, err := r.client.FieldTypes.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetFieldType: %w", err)
		}
		result := fieldTypeToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetFieldTypeByType(t string) (*db.FieldTypes, error) {
	return doRead(r, func() (*db.FieldTypes, error) {
		ctx := context.Background()
		items, err := r.client.FieldTypes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetFieldTypeByType: %w", err)
		}
		for i := range items {
			if items[i].Type == t {
				result := fieldTypeToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetFieldTypeByType: not found: %s", t)
	})
}

func (r *RemoteDriver) ListFieldTypes() (*[]db.FieldTypes, error) {
	return doRead(r, func() (*[]db.FieldTypes, error) {
		ctx := context.Background()
		items, err := r.client.FieldTypes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListFieldTypes: %w", err)
		}
		result := make([]db.FieldTypes, len(items))
		for i := range items {
			result[i] = fieldTypeToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateFieldType(ctx context.Context, _ audited.AuditContext, params db.UpdateFieldTypeParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := fieldTypeUpdateFromDb(params)
		if _, err := r.client.FieldTypes.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateFieldType: %w", err)
		}
		id := string(params.FieldTypeID)
		return &id, nil
	})
}

// ---------------------------------------------------------------------------
// Media
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountMedia() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.Media.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountMedia: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateMedia(_ context.Context, _ audited.AuditContext, _ db.CreateMediaParams) (*db.Media, error) {
	return nil, ErrNotSupported{Method: "CreateMedia"}
}

func (r *RemoteDriver) CreateMediaTable() error {
	return ErrNotSupported{Method: "CreateMediaTable"}
}

func (r *RemoteDriver) DeleteMedia(ctx context.Context, _ audited.AuditContext, id types.MediaID) error {
	return doWriteErr(r, func() error {
		if err := r.client.Media.Delete(ctx, modula.MediaID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteMedia: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetMedia(id types.MediaID) (*db.Media, error) {
	return doRead(r, func() (*db.Media, error) {
		ctx := context.Background()
		sdkID := modula.MediaID(string(id))
		item, err := r.client.Media.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetMedia: %w", err)
		}
		result := mediaToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetMediaByName(name string) (*db.Media, error) {
	return doRead(r, func() (*db.Media, error) {
		ctx := context.Background()
		items, err := r.client.Media.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetMediaByName: %w", err)
		}
		for i := range items {
			if items[i].Name != nil && *items[i].Name == name {
				result := mediaToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetMediaByName: not found: %s", name)
	})
}

func (r *RemoteDriver) GetMediaByURL(u types.URL) (*db.Media, error) {
	return doRead(r, func() (*db.Media, error) {
		ctx := context.Background()
		items, err := r.client.Media.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetMediaByURL: %w", err)
		}
		for i := range items {
			if string(items[i].URL) == string(u) {
				result := mediaToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetMediaByURL: not found: %s", string(u))
	})
}

func (r *RemoteDriver) ListMedia() (*[]db.Media, error) {
	return doRead(r, func() (*[]db.Media, error) {
		ctx := context.Background()
		items, err := r.client.Media.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListMedia: %w", err)
		}
		result := make([]db.Media, len(items))
		for i := range items {
			result[i] = mediaToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListMediaPaginated(p db.PaginationParams) (*[]db.Media, error) {
	return doRead(r, func() (*[]db.Media, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.Media.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListMediaPaginated: %w", err)
		}
		var sdkItems []modula.Media
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListMediaPaginated: decode: %w", err)
		}
		result := make([]db.Media, len(sdkItems))
		for i := range sdkItems {
			result[i] = mediaToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateMedia(ctx context.Context, _ audited.AuditContext, params db.UpdateMediaParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := modula.UpdateMediaParams{
			MediaID:     modula.MediaID(string(params.MediaID)),
			Name:        dbStrPtr(params.Name),
			DisplayName: dbStrPtr(params.DisplayName),
			Alt:         dbStrPtr(params.Alt),
			Caption:     dbStrPtr(params.Caption),
			Description: dbStrPtr(params.Description),
			Class:       dbStrPtr(params.Class),
			FocalX:      float64Ptr(params.FocalX),
			FocalY:      float64Ptr(params.FocalY),
		}
		if _, err := r.client.Media.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateMedia: %w", err)
		}
		id := string(params.MediaID)
		return &id, nil
	})
}

// ---------------------------------------------------------------------------
// MediaDimensions
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountMediaDimensions() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.MediaDimensions.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountMediaDimensions: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateMediaDimension(ctx context.Context, _ audited.AuditContext, params db.CreateMediaDimensionParams) (*db.MediaDimensions, error) {
	return doWrite(r, func() (*db.MediaDimensions, error) {
		sdkParams := modula.CreateMediaDimensionParams{
			Label:       dbStrPtr(params.Label),
			Width:       int64Ptr(params.Width),
			Height:      int64Ptr(params.Height),
			AspectRatio: dbStrPtr(params.AspectRatio),
		}
		result, err := r.client.MediaDimensions.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateMediaDimension: %w", err)
		}
		row := mediaDimensionToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateMediaDimensionTable() error {
	return ErrNotSupported{Method: "CreateMediaDimensionTable"}
}

func (r *RemoteDriver) DeleteMediaDimension(ctx context.Context, _ audited.AuditContext, id string) error {
	return doWriteErr(r, func() error {
		if err := r.client.MediaDimensions.Delete(ctx, modula.MediaDimensionID(id)); err != nil {
			return fmt.Errorf("remote: DeleteMediaDimension: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetMediaDimension(id string) (*db.MediaDimensions, error) {
	return doRead(r, func() (*db.MediaDimensions, error) {
		ctx := context.Background()
		sdkID := modula.MediaDimensionID(id)
		item, err := r.client.MediaDimensions.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetMediaDimension: %w", err)
		}
		result := mediaDimensionToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) ListMediaDimensions() (*[]db.MediaDimensions, error) {
	return doRead(r, func() (*[]db.MediaDimensions, error) {
		ctx := context.Background()
		items, err := r.client.MediaDimensions.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListMediaDimensions: %w", err)
		}
		result := make([]db.MediaDimensions, len(items))
		for i := range items {
			result[i] = mediaDimensionToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateMediaDimension(ctx context.Context, _ audited.AuditContext, params db.UpdateMediaDimensionParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := modula.UpdateMediaDimensionParams{
			MdID:        modula.MediaDimensionID(params.MdID),
			Label:       dbStrPtr(params.Label),
			Width:       int64Ptr(params.Width),
			Height:      int64Ptr(params.Height),
			AspectRatio: dbStrPtr(params.AspectRatio),
		}
		if _, err := r.client.MediaDimensions.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateMediaDimension: %w", err)
		}
		return &params.MdID, nil
	})
}

// ---------------------------------------------------------------------------
// Permissions
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountPermissions() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.Permissions.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountPermissions: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreatePermission(ctx context.Context, _ audited.AuditContext, params db.CreatePermissionParams) (*db.Permissions, error) {
	return doWrite(r, func() (*db.Permissions, error) {
		sdkParams := modula.CreatePermissionParams{
			Label: params.Label,
		}
		result, err := r.client.Permissions.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreatePermission: %w", err)
		}
		row := permissionToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreatePermissionTable() error {
	return ErrNotSupported{Method: "CreatePermissionTable"}
}

func (r *RemoteDriver) DeletePermission(ctx context.Context, _ audited.AuditContext, id types.PermissionID) error {
	return doWriteErr(r, func() error {
		if err := r.client.Permissions.Delete(ctx, modula.PermissionID(string(id))); err != nil {
			return fmt.Errorf("remote: DeletePermission: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetPermission(id types.PermissionID) (*db.Permissions, error) {
	return doRead(r, func() (*db.Permissions, error) {
		ctx := context.Background()
		sdkID := modula.PermissionID(string(id))
		item, err := r.client.Permissions.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetPermission: %w", err)
		}
		result := permissionToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetPermissionByLabel(label string) (*db.Permissions, error) {
	return doRead(r, func() (*db.Permissions, error) {
		ctx := context.Background()
		items, err := r.client.Permissions.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetPermissionByLabel: %w", err)
		}
		for i := range items {
			if items[i].Label == label {
				result := permissionToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetPermissionByLabel: not found: %s", label)
	})
}

func (r *RemoteDriver) ListPermissions() (*[]db.Permissions, error) {
	return doRead(r, func() (*[]db.Permissions, error) {
		ctx := context.Background()
		items, err := r.client.Permissions.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListPermissions: %w", err)
		}
		result := make([]db.Permissions, len(items))
		for i := range items {
			result[i] = permissionToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdatePermission(ctx context.Context, _ audited.AuditContext, params db.UpdatePermissionParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := modula.UpdatePermissionParams{
			PermissionID: modula.PermissionID(string(params.PermissionID)),
			Label:        params.Label,
		}
		if _, err := r.client.Permissions.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdatePermission: %w", err)
		}
		id := string(params.PermissionID)
		return &id, nil
	})
}

// ---------------------------------------------------------------------------
// Roles
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountRoles() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.Roles.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountRoles: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateRole(ctx context.Context, _ audited.AuditContext, params db.CreateRoleParams) (*db.Roles, error) {
	return doWrite(r, func() (*db.Roles, error) {
		sdkParams := modula.CreateRoleParams{
			Label: params.Label,
		}
		result, err := r.client.Roles.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateRole: %w", err)
		}
		row := roleToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateRoleTable() error {
	return ErrNotSupported{Method: "CreateRoleTable"}
}

func (r *RemoteDriver) DeleteRole(ctx context.Context, _ audited.AuditContext, id types.RoleID) error {
	return doWriteErr(r, func() error {
		if err := r.client.Roles.Delete(ctx, modula.RoleID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteRole: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetRole(id types.RoleID) (*db.Roles, error) {
	return doRead(r, func() (*db.Roles, error) {
		ctx := context.Background()
		sdkID := modula.RoleID(string(id))
		item, err := r.client.Roles.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetRole: %w", err)
		}
		result := roleToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetRoleByLabel(label string) (*db.Roles, error) {
	return doRead(r, func() (*db.Roles, error) {
		ctx := context.Background()
		items, err := r.client.Roles.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetRoleByLabel: %w", err)
		}
		for i := range items {
			if items[i].Label == label {
				result := roleToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetRoleByLabel: not found: %s", label)
	})
}

func (r *RemoteDriver) ListRoles() (*[]db.Roles, error) {
	return doRead(r, func() (*[]db.Roles, error) {
		ctx := context.Background()
		items, err := r.client.Roles.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListRoles: %w", err)
		}
		result := make([]db.Roles, len(items))
		for i := range items {
			result[i] = roleToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateRole(ctx context.Context, _ audited.AuditContext, params db.UpdateRoleParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := modula.UpdateRoleParams{
			RoleID: modula.RoleID(string(params.RoleID)),
			Label:  params.Label,
		}
		if _, err := r.client.Roles.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateRole: %w", err)
		}
		id := string(params.RoleID)
		return &id, nil
	})
}

// ---------------------------------------------------------------------------
// RolePermissions
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountRolePermissions() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		items, err := r.client.RolePermissions.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountRolePermissions: %w", err)
		}
		count := int64(len(items))
		return &count, nil
	})
}

func (r *RemoteDriver) CreateRolePermission(ctx context.Context, _ audited.AuditContext, params db.CreateRolePermissionParams) (*db.RolePermissions, error) {
	return doWrite(r, func() (*db.RolePermissions, error) {
		sdkParams := modula.CreateRolePermissionParams{
			RoleID:       modula.RoleID(string(params.RoleID)),
			PermissionID: modula.PermissionID(string(params.PermissionID)),
		}
		result, err := r.client.RolePermissions.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateRolePermission: %w", err)
		}
		row := rolePermissionToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateRolePermissionsTable() error {
	return ErrNotSupported{Method: "CreateRolePermissionsTable"}
}

func (r *RemoteDriver) DeleteRolePermission(ctx context.Context, _ audited.AuditContext, id types.RolePermissionID) error {
	return doWriteErr(r, func() error {
		if err := r.client.RolePermissions.Delete(ctx, modula.RolePermissionID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteRolePermission: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) DeleteRolePermissionsByRoleID(ctx context.Context, _ audited.AuditContext, roleID types.RoleID) error {
	return doWriteErr(r, func() error {
		sdkRoleID := modula.RoleID(string(roleID))
		rps, err := r.client.RolePermissions.ListByRole(ctx, sdkRoleID)
		if err != nil {
			return fmt.Errorf("remote: DeleteRolePermissionsByRoleID: list: %w", err)
		}
		for _, rp := range rps {
			if err := r.client.RolePermissions.Delete(ctx, rp.ID); err != nil {
				return fmt.Errorf("remote: DeleteRolePermissionsByRoleID: delete %s: %w", string(rp.ID), err)
			}
		}
		return nil
	})
}

func (r *RemoteDriver) GetRolePermission(id types.RolePermissionID) (*db.RolePermissions, error) {
	return doRead(r, func() (*db.RolePermissions, error) {
		ctx := context.Background()
		sdkID := modula.RolePermissionID(string(id))
		item, err := r.client.RolePermissions.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetRolePermission: %w", err)
		}
		result := rolePermissionToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) ListRolePermissions() (*[]db.RolePermissions, error) {
	return doRead(r, func() (*[]db.RolePermissions, error) {
		ctx := context.Background()
		items, err := r.client.RolePermissions.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListRolePermissions: %w", err)
		}
		result := make([]db.RolePermissions, len(items))
		for i := range items {
			result[i] = rolePermissionToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListRolePermissionsByRoleID(roleID types.RoleID) (*[]db.RolePermissions, error) {
	return doRead(r, func() (*[]db.RolePermissions, error) {
		ctx := context.Background()
		sdkRoleID := modula.RoleID(string(roleID))
		items, err := r.client.RolePermissions.ListByRole(ctx, sdkRoleID)
		if err != nil {
			return nil, fmt.Errorf("remote: ListRolePermissionsByRoleID: %w", err)
		}
		result := make([]db.RolePermissions, len(items))
		for i := range items {
			result[i] = rolePermissionToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListRolePermissionsByPermissionID(permID types.PermissionID) (*[]db.RolePermissions, error) {
	return doRead(r, func() (*[]db.RolePermissions, error) {
		ctx := context.Background()
		all, err := r.client.RolePermissions.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListRolePermissionsByPermissionID: %w", err)
		}
		target := modula.PermissionID(string(permID))
		var result []db.RolePermissions
		for i := range all {
			if all[i].PermissionID == target {
				result = append(result, rolePermissionToDb(&all[i]))
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListPermissionLabelsByRoleID(roleID types.RoleID) (*[]string, error) {
	return doRead(r, func() (*[]string, error) {
		ctx := context.Background()
		sdkRoleID := modula.RoleID(string(roleID))
		rps, err := r.client.RolePermissions.ListByRole(ctx, sdkRoleID)
		if err != nil {
			return nil, fmt.Errorf("remote: ListPermissionLabelsByRoleID: %w", err)
		}
		// Fetch all permissions to build ID->label map.
		perms, err := r.client.Permissions.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListPermissionLabelsByRoleID: list permissions: %w", err)
		}
		permMap := make(map[string]string, len(perms))
		for _, p := range perms {
			permMap[string(p.PermissionID)] = p.Label
		}
		labels := make([]string, 0, len(rps))
		for _, rp := range rps {
			if label, ok := permMap[string(rp.PermissionID)]; ok {
				labels = append(labels, label)
			}
		}
		return &labels, nil
	})
}

// ---------------------------------------------------------------------------
// Routes
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountRoutes() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.Routes.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountRoutes: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateRoute(ctx context.Context, _ audited.AuditContext, params db.CreateRouteParams) (*db.Routes, error) {
	return doWrite(r, func() (*db.Routes, error) {
		sdkParams := routeCreateFromDb(params)
		result, err := r.client.Routes.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateRoute: %w", err)
		}
		row := routeToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateRouteTable() error {
	return ErrNotSupported{Method: "CreateRouteTable"}
}

func (r *RemoteDriver) DeleteRoute(ctx context.Context, _ audited.AuditContext, id types.RouteID) error {
	return doWriteErr(r, func() error {
		if err := r.client.Routes.Delete(ctx, modula.RouteID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteRoute: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetContentTreeByRoute(routeID types.NullableRouteID) (*[]db.GetContentTreeByRouteRow, error) {
	return doRead(r, func() (*[]db.GetContentTreeByRouteRow, error) {
		if !routeID.Valid {
			return nil, fmt.Errorf("remote: GetContentTreeByRoute: route_id is required")
		}
		ctx := context.Background()
		sdkRouteID := modula.RouteID(string(routeID.ID))
		nodes, err := r.client.ContentTree.GetByRoute(ctx, sdkRouteID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetContentTreeByRoute: %w", err)
		}
		result := make([]db.GetContentTreeByRouteRow, len(nodes))
		for i, n := range nodes {
			result[i] = db.GetContentTreeByRouteRow{
				ContentDataID: types.ContentID(string(n.ContentID)),
				ParentID:      nullContentIDFromString(n.ParentID),
				FirstChildID:  nullContentIDFromString(n.FirstChildID),
				NextSiblingID: nullContentIDFromString(n.NextSiblingID),
				PrevSiblingID: nullContentIDFromString(n.PrevSiblingID),
				DatatypeID: func() types.NullableDatatypeID {
					if n.DatatypeID == nil {
						return types.NullableDatatypeID{}
					}
					return types.NullableDatatypeID{ID: types.DatatypeID(*n.DatatypeID), Valid: true}
				}(),
				RouteID: func() types.NullableRouteID {
					if n.RouteID == nil {
						return types.NullableRouteID{}
					}
					return types.NullableRouteID{ID: types.RouteID(*n.RouteID), Valid: true}
				}(),
				Status:        types.ContentStatus(n.Status),
				DatatypeLabel: n.Title,
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) GetContentTreeByRootID(_ types.NullableContentID) (*[]db.GetContentTreeByRouteRow, error) {
	return nil, ErrNotSupported{Method: "GetContentTreeByRootID"}
}

func (r *RemoteDriver) GetRoute(id types.RouteID) (*db.Routes, error) {
	return doRead(r, func() (*db.Routes, error) {
		ctx := context.Background()
		sdkID := modula.RouteID(string(id))
		item, err := r.client.Routes.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetRoute: %w", err)
		}
		result := routeToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetRouteID(slug string) (*types.RouteID, error) {
	return doRead(r, func() (*types.RouteID, error) {
		ctx := context.Background()
		items, err := r.client.Routes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetRouteID: %w", err)
		}
		for _, item := range items {
			if string(item.Slug) == slug {
				id := types.RouteID(string(item.RouteID))
				return &id, nil
			}
		}
		return nil, fmt.Errorf("remote: GetRouteID: not found: %s", slug)
	})
}

func (r *RemoteDriver) GetRouteTreeByRouteID(routeID types.NullableRouteID) (*[]db.GetRouteTreeByRouteIDRow, error) {
	return doRead(r, func() (*[]db.GetRouteTreeByRouteIDRow, error) {
		if !routeID.Valid {
			return nil, fmt.Errorf("remote: GetRouteTreeByRouteID: route_id is required")
		}
		ctx := context.Background()
		sdkRouteID := modula.RouteID(string(routeID.ID))
		nodes, err := r.client.ContentTree.GetByRoute(ctx, sdkRouteID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetRouteTreeByRouteID: %w", err)
		}
		result := make([]db.GetRouteTreeByRouteIDRow, len(nodes))
		for i, n := range nodes {
			result[i] = db.GetRouteTreeByRouteIDRow{
				ContentDataID: types.ContentID(string(n.ContentID)),
				ParentID:      nullContentIDFromString(n.ParentID),
				FirstChildID:  nullContentIDFromString(n.FirstChildID),
				NextSiblingID: nullContentIDFromString(n.NextSiblingID),
				PrevSiblingID: nullContentIDFromString(n.PrevSiblingID),
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListRoutes() (*[]db.Routes, error) {
	return doRead(r, func() (*[]db.Routes, error) {
		ctx := context.Background()
		items, err := r.client.Routes.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListRoutes: %w", err)
		}
		result := make([]db.Routes, len(items))
		for i := range items {
			result[i] = routeToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListRoutesByDatatype(datatypeID types.DatatypeID) (*[]db.Routes, error) {
	return doRead(r, func() (*[]db.Routes, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("datatype_id", string(datatypeID))
		raw, err := r.client.Routes.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListRoutesByDatatype: %w", err)
		}
		var sdkItems []modula.Route
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListRoutesByDatatype: decode: %w", err)
		}
		result := make([]db.Routes, len(sdkItems))
		for i := range sdkItems {
			result[i] = routeToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListRoutesPaginated(p db.PaginationParams) (*[]db.Routes, error) {
	return doRead(r, func() (*[]db.Routes, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.Routes.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListRoutesPaginated: %w", err)
		}
		var sdkItems []modula.Route
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListRoutesPaginated: decode: %w", err)
		}
		result := make([]db.Routes, len(sdkItems))
		for i := range sdkItems {
			result[i] = routeToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateRoute(ctx context.Context, _ audited.AuditContext, params db.UpdateRouteParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := routeUpdateFromDb(params)
		if _, err := r.client.Routes.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateRoute: %w", err)
		}
		id := string(params.Slug)
		return &id, nil
	})
}

// ---------------------------------------------------------------------------
// Sessions
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountSessions() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		items, err := r.client.Sessions.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountSessions: %w", err)
		}
		count := int64(len(items))
		return &count, nil
	})
}

func (r *RemoteDriver) CreateSession(_ context.Context, _ audited.AuditContext, _ db.CreateSessionParams) (*db.Sessions, error) {
	return nil, ErrNotSupported{Method: "CreateSession"}
}

func (r *RemoteDriver) CreateSessionTable() error {
	return ErrNotSupported{Method: "CreateSessionTable"}
}

func (r *RemoteDriver) DeleteSession(ctx context.Context, _ audited.AuditContext, id types.SessionID) error {
	return doWriteErr(r, func() error {
		if err := r.client.Sessions.Remove(ctx, modula.SessionID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteSession: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetSession(id types.SessionID) (*db.Sessions, error) {
	return doRead(r, func() (*db.Sessions, error) {
		ctx := context.Background()
		sdkID := modula.SessionID(string(id))
		item, err := r.client.Sessions.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetSession: %w", err)
		}
		result := sessionToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetSessionByUserId(userID types.NullableUserID) (*db.Sessions, error) {
	return doRead(r, func() (*db.Sessions, error) {
		ctx := context.Background()
		items, err := r.client.Sessions.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetSessionByUserId: %w", err)
		}
		target := userIDPtr(userID)
		for i := range items {
			if target == nil && items[i].UserID == nil {
				result := sessionToDb(&items[i])
				return &result, nil
			}
			if target != nil && items[i].UserID != nil && *items[i].UserID == *target {
				result := sessionToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetSessionByUserId: session not found for user")
	})
}

func (r *RemoteDriver) ListSessions() (*[]db.Sessions, error) {
	return doRead(r, func() (*[]db.Sessions, error) {
		ctx := context.Background()
		items, err := r.client.Sessions.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListSessions: %w", err)
		}
		result := make([]db.Sessions, len(items))
		for i := range items {
			result[i] = sessionToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateSession(ctx context.Context, _ audited.AuditContext, params db.UpdateSessionParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		var lastAccess *string
		if params.LastAccess.Valid {
			s := params.LastAccess.String()
			lastAccess = &s
		}
		sdkParams := modula.UpdateSessionParams{
			SessionID:   modula.SessionID(string(params.SessionID)),
			UserID:      userIDPtr(params.UserID),
			ExpiresAt:   dbTimestampToSdk(params.ExpiresAt),
			LastAccess:  lastAccess,
			IpAddress:   dbStrPtr(params.IpAddress),
			UserAgent:   dbStrPtr(params.UserAgent),
			SessionData: dbStrPtr(params.SessionData),
		}
		if _, err := r.client.Sessions.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateSession: %w", err)
		}
		id := string(params.SessionID)
		return &id, nil
	})
}

// ---------------------------------------------------------------------------
// Tables
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountTables() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.Tables.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountTables: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateTable(ctx context.Context, _ audited.AuditContext, params db.CreateTableParams) (*db.Tables, error) {
	return doWrite(r, func() (*db.Tables, error) {
		sdkParams := modula.CreateTableParams{
			Label: params.Label,
		}
		result, err := r.client.Tables.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateTable: %w", err)
		}
		row := tableToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateTableTable() error {
	return ErrNotSupported{Method: "CreateTableTable"}
}

func (r *RemoteDriver) DeleteTable(ctx context.Context, _ audited.AuditContext, id string) error {
	return doWriteErr(r, func() error {
		if err := r.client.Tables.Delete(ctx, modula.TableID(id)); err != nil {
			return fmt.Errorf("remote: DeleteTable: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetTable(id string) (*db.Tables, error) {
	return doRead(r, func() (*db.Tables, error) {
		ctx := context.Background()
		sdkID := modula.TableID(id)
		item, err := r.client.Tables.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetTable: %w", err)
		}
		result := tableToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) ListTables() (*[]db.Tables, error) {
	return doRead(r, func() (*[]db.Tables, error) {
		ctx := context.Background()
		items, err := r.client.Tables.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListTables: %w", err)
		}
		result := make([]db.Tables, len(items))
		for i := range items {
			result[i] = tableToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateTable(ctx context.Context, _ audited.AuditContext, params db.UpdateTableParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := modula.UpdateTableParams{
			ID:    modula.TableID(params.ID),
			Label: params.Label,
		}
		if _, err := r.client.Tables.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateTable: %w", err)
		}
		return &params.ID, nil
	})
}

// ---------------------------------------------------------------------------
// Tokens
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountTokens() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.Tokens.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountTokens: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateToken(ctx context.Context, _ audited.AuditContext, params db.CreateTokenParams) (*db.Tokens, error) {
	return doWrite(r, func() (*db.Tokens, error) {
		sdkParams := tokenCreateFromDb(params)
		result, err := r.client.Tokens.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateToken: %w", err)
		}
		row := tokenToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateTokenTable() error {
	return ErrNotSupported{Method: "CreateTokenTable"}
}

func (r *RemoteDriver) DeleteToken(ctx context.Context, _ audited.AuditContext, id string) error {
	return doWriteErr(r, func() error {
		if err := r.client.Tokens.Delete(ctx, modula.TokenID(id)); err != nil {
			return fmt.Errorf("remote: DeleteToken: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetToken(id string) (*db.Tokens, error) {
	return doRead(r, func() (*db.Tokens, error) {
		ctx := context.Background()
		sdkID := modula.TokenID(id)
		item, err := r.client.Tokens.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetToken: %w", err)
		}
		result := tokenToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetTokenByTokenValue(tokenValue string) (*db.Tokens, error) {
	return doRead(r, func() (*db.Tokens, error) {
		ctx := context.Background()
		items, err := r.client.Tokens.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetTokenByTokenValue: %w", err)
		}
		for i := range items {
			if items[i].Token == tokenValue {
				result := tokenToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetTokenByTokenValue: not found")
	})
}

func (r *RemoteDriver) GetTokenByUserId(userID types.NullableUserID) (*[]db.Tokens, error) {
	return doRead(r, func() (*[]db.Tokens, error) {
		ctx := context.Background()
		items, err := r.client.Tokens.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetTokenByUserId: %w", err)
		}
		var result []db.Tokens
		for i := range items {
			if userID.Valid && items[i].UserID != nil && string(*items[i].UserID) == string(userID.ID) {
				result = append(result, tokenToDb(&items[i]))
			} else if !userID.Valid && items[i].UserID == nil {
				result = append(result, tokenToDb(&items[i]))
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListTokens() (*[]db.Tokens, error) {
	return doRead(r, func() (*[]db.Tokens, error) {
		ctx := context.Background()
		items, err := r.client.Tokens.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListTokens: %w", err)
		}
		result := make([]db.Tokens, len(items))
		for i := range items {
			result[i] = tokenToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateToken(ctx context.Context, _ audited.AuditContext, params db.UpdateTokenParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := modula.UpdateTokenParams{
			ID:        modula.TokenID(params.ID),
			Token:     params.Token,
			IssuedAt:  params.IssuedAt.String(),
			ExpiresAt: dbTimestampToSdk(params.ExpiresAt),
			Revoked:   params.Revoked,
		}
		if _, err := r.client.Tokens.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateToken: %w", err)
		}
		return &params.ID, nil
	})
}

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountUsers() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.Users.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountUsers: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateUser(ctx context.Context, _ audited.AuditContext, params db.CreateUserParams) (*db.Users, error) {
	return doWrite(r, func() (*db.Users, error) {
		sdkParams := userCreateFromDb(params)
		result, err := r.client.Users.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateUser: %w", err)
		}
		row := userToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateUserTable() error {
	return ErrNotSupported{Method: "CreateUserTable"}
}

func (r *RemoteDriver) DeleteUser(ctx context.Context, _ audited.AuditContext, id types.UserID) error {
	return doWriteErr(r, func() error {
		if err := r.client.Users.Delete(ctx, modula.UserID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteUser: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetUser(id types.UserID) (*db.Users, error) {
	return doRead(r, func() (*db.Users, error) {
		ctx := context.Background()
		sdkID := modula.UserID(string(id))
		item, err := r.client.Users.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetUser: %w", err)
		}
		result := userToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetUserByEmail(email types.Email) (*db.Users, error) {
	return doRead(r, func() (*db.Users, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("email", string(email))
		raw, err := r.client.Users.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: GetUserByEmail: %w", err)
		}
		var user modula.User
		if err := json.Unmarshal(raw, &user); err != nil {
			return nil, fmt.Errorf("remote: GetUserByEmail: decode: %w", err)
		}
		result := userToDb(&user)
		return &result, nil
	})
}

func (r *RemoteDriver) GetUserBySSHFingerprint(_ string) (*db.Users, error) {
	return nil, ErrNotSupported{Method: "GetUserBySSHFingerprint"}
}

func (r *RemoteDriver) ListUsers() (*[]db.Users, error) {
	return doRead(r, func() (*[]db.Users, error) {
		ctx := context.Background()
		items, err := r.client.Users.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListUsers: %w", err)
		}
		result := make([]db.Users, len(items))
		for i := range items {
			result[i] = userToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListUsersWithRoleLabel() (*[]db.UserWithRoleLabelRow, error) {
	return doRead(r, func() (*[]db.UserWithRoleLabelRow, error) {
		ctx := context.Background()
		users, err := r.client.Users.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListUsersWithRoleLabel: list users: %w", err)
		}
		roles, err := r.client.Roles.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListUsersWithRoleLabel: list roles: %w", err)
		}
		// Build role ID -> label map. The user's Role field is the role ID.
		roleMap := make(map[string]string, len(roles))
		for _, role := range roles {
			roleMap[string(role.RoleID)] = role.Label
		}
		result := make([]db.UserWithRoleLabelRow, len(users))
		for i, u := range users {
			result[i] = db.UserWithRoleLabelRow{
				UserID:       types.UserID(string(u.UserID)),
				Username:     u.Username,
				Name:         u.Name,
				Email:        types.Email(string(u.Email)),
				Role:         u.Role,
				RoleLabel:    roleMap[u.Role],
				DateCreated:  sdkTimestampToDb(u.DateCreated),
				DateModified: sdkTimestampToDb(u.DateModified),
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateUser(ctx context.Context, _ audited.AuditContext, params db.UpdateUserParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := userUpdateFromDb(params)
		if _, err := r.client.Users.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateUser: %w", err)
		}
		id := string(params.UserID)
		return &id, nil
	})
}

// ---------------------------------------------------------------------------
// UserOauths
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountUserOauths() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.UsersOauth.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountUserOauths: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateUserOauth(ctx context.Context, _ audited.AuditContext, params db.CreateUserOauthParams) (*db.UserOauth, error) {
	return doWrite(r, func() (*db.UserOauth, error) {
		sdkParams := modula.CreateUserOauthParams{
			UserID:              userIDPtr(params.UserID),
			OauthProvider:       params.OauthProvider,
			OauthProviderUserID: params.OauthProviderUserID,
			AccessToken:         params.AccessToken,
			RefreshToken:        params.RefreshToken,
			TokenExpiresAt:      params.TokenExpiresAt.String(),
			DateCreated:         dbTimestampToSdk(params.DateCreated),
		}
		result, err := r.client.UsersOauth.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateUserOauth: %w", err)
		}
		row := userOauthToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateUserOauthTable() error {
	return ErrNotSupported{Method: "CreateUserOauthTable"}
}

func (r *RemoteDriver) DeleteUserOauth(ctx context.Context, _ audited.AuditContext, id types.UserOauthID) error {
	return doWriteErr(r, func() error {
		if err := r.client.UsersOauth.Delete(ctx, modula.UserOauthID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteUserOauth: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetUserOauth(id types.UserOauthID) (*db.UserOauth, error) {
	return doRead(r, func() (*db.UserOauth, error) {
		ctx := context.Background()
		sdkID := modula.UserOauthID(string(id))
		item, err := r.client.UsersOauth.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetUserOauth: %w", err)
		}
		result := userOauthToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetUserOauthByProviderID(provider string, providerUserID string) (*db.UserOauth, error) {
	return doRead(r, func() (*db.UserOauth, error) {
		ctx := context.Background()
		items, err := r.client.UsersOauth.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetUserOauthByProviderID: %w", err)
		}
		for i := range items {
			if items[i].OauthProvider == provider && items[i].OauthProviderUserID == providerUserID {
				result := userOauthToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetUserOauthByProviderID: not found: %s/%s", provider, providerUserID)
	})
}

func (r *RemoteDriver) GetUserOauthByUserId(userID types.NullableUserID) (*db.UserOauth, error) {
	return doRead(r, func() (*db.UserOauth, error) {
		ctx := context.Background()
		items, err := r.client.UsersOauth.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetUserOauthByUserId: %w", err)
		}
		for i := range items {
			if userID.Valid && items[i].UserID != nil && string(*items[i].UserID) == string(userID.ID) {
				result := userOauthToDb(&items[i])
				return &result, nil
			}
			if !userID.Valid && items[i].UserID == nil {
				result := userOauthToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetUserOauthByUserId: not found")
	})
}

func (r *RemoteDriver) ListUserOauths() (*[]db.UserOauth, error) {
	return doRead(r, func() (*[]db.UserOauth, error) {
		ctx := context.Background()
		items, err := r.client.UsersOauth.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListUserOauths: %w", err)
		}
		result := make([]db.UserOauth, len(items))
		for i := range items {
			result[i] = userOauthToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateUserOauth(ctx context.Context, _ audited.AuditContext, params db.UpdateUserOauthParams) (*string, error) {
	return doWrite(r, func() (*string, error) {
		sdkParams := modula.UpdateUserOauthParams{
			UserOauthID:    modula.UserOauthID(string(params.UserOauthID)),
			AccessToken:    params.AccessToken,
			RefreshToken:   params.RefreshToken,
			TokenExpiresAt: params.TokenExpiresAt.String(),
		}
		if _, err := r.client.UsersOauth.Update(ctx, sdkParams); err != nil {
			return nil, fmt.Errorf("remote: UpdateUserOauth: %w", err)
		}
		id := string(params.UserOauthID)
		return &id, nil
	})
}

// ---------------------------------------------------------------------------
// UserSshKeys
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountUserSshKeys() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		items, err := r.client.SSHKeys.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountUserSshKeys: %w", err)
		}
		count := int64(len(items))
		return &count, nil
	})
}

func (r *RemoteDriver) CreateUserSshKey(ctx context.Context, _ audited.AuditContext, params db.CreateUserSshKeyParams) (*db.UserSshKeys, error) {
	return doWrite(r, func() (*db.UserSshKeys, error) {
		sdkParams := userSshKeyCreateFromDb(params)
		result, err := r.client.SSHKeys.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateUserSshKey: %w", err)
		}
		row := userSshKeyToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateUserSshKeyTable() error {
	return ErrNotSupported{Method: "CreateUserSshKeyTable"}
}

func (r *RemoteDriver) DeleteUserSshKey(ctx context.Context, _ audited.AuditContext, id string) error {
	return doWriteErr(r, func() error {
		if err := r.client.SSHKeys.Delete(ctx, modula.UserSshKeyID(id)); err != nil {
			return fmt.Errorf("remote: DeleteUserSshKey: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetUserSshKey(id string) (*db.UserSshKeys, error) {
	return doRead(r, func() (*db.UserSshKeys, error) {
		ctx := context.Background()
		// SSH keys list endpoint returns SshKeyListItem; filter by ID.
		items, err := r.client.SSHKeys.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetUserSshKey: %w", err)
		}
		for _, item := range items {
			if string(item.SshKeyID) == id {
				result := db.UserSshKeys{
					SshKeyID:    string(item.SshKeyID),
					KeyType:     item.KeyType,
					Fingerprint: item.Fingerprint,
					Label:       item.Label,
					DateCreated: sdkTimestampToDb(item.DateCreated),
					LastUsed:    item.LastUsed,
				}
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetUserSshKey: not found: %s", id)
	})
}

func (r *RemoteDriver) GetUserSshKeyByFingerprint(fingerprint string) (*db.UserSshKeys, error) {
	return doRead(r, func() (*db.UserSshKeys, error) {
		ctx := context.Background()
		item, err := r.client.SSHKeys.GetByFingerprint(ctx, fingerprint)
		if err != nil {
			return nil, fmt.Errorf("remote: GetUserSshKeyByFingerprint: %w", err)
		}
		result := db.UserSshKeys{
			SshKeyID:    string(item.SshKeyID),
			KeyType:     item.KeyType,
			Fingerprint: item.Fingerprint,
			Label:       item.Label,
			DateCreated: sdkTimestampToDb(item.DateCreated),
			LastUsed:    item.LastUsed,
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListUserSshKeys(_ types.NullableUserID) (*[]db.UserSshKeys, error) {
	return doRead(r, func() (*[]db.UserSshKeys, error) {
		ctx := context.Background()
		items, err := r.client.SSHKeys.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListUserSshKeys: %w", err)
		}
		result := make([]db.UserSshKeys, len(items))
		for i, item := range items {
			result[i] = db.UserSshKeys{
				SshKeyID:    string(item.SshKeyID),
				KeyType:     item.KeyType,
				Fingerprint: item.Fingerprint,
				Label:       item.Label,
				DateCreated: sdkTimestampToDb(item.DateCreated),
				LastUsed:    item.LastUsed,
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateUserSshKeyLabel(_ context.Context, _ audited.AuditContext, _ string, _ string) error {
	return ErrNotSupported{Method: "UpdateUserSshKeyLabel"}
}

func (r *RemoteDriver) UpdateUserSshKeyLastUsed(_ string, _ string) error {
	return ErrNotSupported{Method: "UpdateUserSshKeyLastUsed"}
}

// ---------------------------------------------------------------------------
// Plugins
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountPlugins() (*int64, error) {
	return nil, ErrNotSupported{Method: "CountPlugins"}
}

func (r *RemoteDriver) CreatePlugin(_ context.Context, _ audited.AuditContext, _ db.CreatePluginParams) (*db.Plugin, error) {
	return nil, ErrNotSupported{Method: "CreatePlugin"}
}

func (r *RemoteDriver) CreatePluginTable() error {
	return ErrNotSupported{Method: "CreatePluginTable"}
}

func (r *RemoteDriver) DeletePlugin(_ context.Context, _ audited.AuditContext, _ types.PluginID) error {
	return ErrNotSupported{Method: "DeletePlugin"}
}

func (r *RemoteDriver) GetPlugin(_ types.PluginID) (*db.Plugin, error) {
	// The SDK plugin resource uses name-based lookup, not ID.
	// Return not supported since the TUI typically uses GetPluginByName.
	return nil, ErrNotSupported{Method: "GetPlugin"}
}

func (r *RemoteDriver) GetPluginByName(name string) (*db.Plugin, error) {
	return doRead(r, func() (*db.Plugin, error) {
		ctx := context.Background()
		info, err := r.client.Plugins.Get(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("remote: GetPluginByName: %w", err)
		}
		result := db.Plugin{
			Name:        info.Name,
			Version:     info.Version,
			Description: info.Description,
			Author:      info.Author,
			Status:      types.PluginStatus(info.State),
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListPlugins() (*[]db.Plugin, error) {
	return doRead(r, func() (*[]db.Plugin, error) {
		ctx := context.Background()
		items, err := r.client.Plugins.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListPlugins: %w", err)
		}
		result := make([]db.Plugin, len(items))
		for i, item := range items {
			result[i] = db.Plugin{
				Name:        item.Name,
				Version:     item.Version,
				Description: item.Description,
				Status:      types.PluginStatus(item.State),
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListPluginsByStatus(status types.PluginStatus) (*[]db.Plugin, error) {
	return doRead(r, func() (*[]db.Plugin, error) {
		ctx := context.Background()
		all, err := r.client.Plugins.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListPluginsByStatus: %w", err)
		}
		var result []db.Plugin
		for _, item := range all {
			if types.PluginStatus(item.State) == status {
				result = append(result, db.Plugin{
					Name:        item.Name,
					Version:     item.Version,
					Description: item.Description,
					Status:      types.PluginStatus(item.State),
				})
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdatePlugin(_ context.Context, _ audited.AuditContext, _ db.UpdatePluginParams) error {
	return ErrNotSupported{Method: "UpdatePlugin"}
}

func (r *RemoteDriver) UpdatePluginStatus(_ context.Context, _ audited.AuditContext, _ types.PluginID, _ types.PluginStatus) error {
	return ErrNotSupported{Method: "UpdatePluginStatus"}
}

// ---------------------------------------------------------------------------
// Locales
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountLocales() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.Locales.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountLocales: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateLocale(ctx context.Context, _ audited.AuditContext, params db.CreateLocaleParams) (*db.Locale, error) {
	return doWrite(r, func() (*db.Locale, error) {
		sdkParams := modula.CreateLocaleRequest{
			Code:         params.Code,
			Label:        params.Label,
			IsDefault:    params.IsDefault,
			IsEnabled:    params.IsEnabled,
			FallbackCode: params.FallbackCode,
			SortOrder:    params.SortOrder,
		}
		result, err := r.client.Locales.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateLocale: %w", err)
		}
		row := localeToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateLocaleTable() error {
	return ErrNotSupported{Method: "CreateLocaleTable"}
}

func (r *RemoteDriver) DeleteLocale(ctx context.Context, _ audited.AuditContext, id types.LocaleID) error {
	return doWriteErr(r, func() error {
		if err := r.client.Locales.Delete(ctx, modula.LocaleID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteLocale: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetLocale(id types.LocaleID) (*db.Locale, error) {
	return doRead(r, func() (*db.Locale, error) {
		ctx := context.Background()
		sdkID := modula.LocaleID(string(id))
		item, err := r.client.Locales.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetLocale: %w", err)
		}
		result := localeToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) GetLocaleByCode(code string) (*db.Locale, error) {
	return doRead(r, func() (*db.Locale, error) {
		ctx := context.Background()
		items, err := r.client.Locales.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetLocaleByCode: %w", err)
		}
		for i := range items {
			if items[i].Code == code {
				result := localeToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetLocaleByCode: not found: %s", code)
	})
}

func (r *RemoteDriver) GetDefaultLocale() (*db.Locale, error) {
	return doRead(r, func() (*db.Locale, error) {
		ctx := context.Background()
		items, err := r.client.Locales.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: GetDefaultLocale: %w", err)
		}
		for i := range items {
			if items[i].IsDefault {
				result := localeToDb(&items[i])
				return &result, nil
			}
		}
		return nil, fmt.Errorf("remote: GetDefaultLocale: no default locale configured")
	})
}

func (r *RemoteDriver) ListLocales() (*[]db.Locale, error) {
	return doRead(r, func() (*[]db.Locale, error) {
		ctx := context.Background()
		items, err := r.client.Locales.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListLocales: %w", err)
		}
		result := make([]db.Locale, len(items))
		for i := range items {
			result[i] = localeToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListEnabledLocales() (*[]db.Locale, error) {
	return doRead(r, func() (*[]db.Locale, error) {
		ctx := context.Background()
		items, err := r.client.Locales.ListEnabled(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListEnabledLocales: %w", err)
		}
		result := make([]db.Locale, len(items))
		for i := range items {
			result[i] = localeToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListLocalesPaginated(p db.PaginationParams) (*[]db.Locale, error) {
	return doRead(r, func() (*[]db.Locale, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.Locales.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListLocalesPaginated: %w", err)
		}
		var sdkItems []modula.Locale
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListLocalesPaginated: decode: %w", err)
		}
		result := make([]db.Locale, len(sdkItems))
		for i := range sdkItems {
			result[i] = localeToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateLocale(ctx context.Context, _ audited.AuditContext, params db.UpdateLocaleParams) error {
	return doWriteErr(r, func() error {
		sdkParams := modula.UpdateLocaleRequest{
			LocaleID:     modula.LocaleID(string(params.LocaleID)),
			Code:         params.Code,
			Label:        params.Label,
			IsDefault:    params.IsDefault,
			IsEnabled:    params.IsEnabled,
			FallbackCode: params.FallbackCode,
			SortOrder:    params.SortOrder,
		}
		if _, err := r.client.Locales.Update(ctx, sdkParams); err != nil {
			return fmt.Errorf("remote: UpdateLocale: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) ClearDefaultLocale(_ context.Context) error {
	return ErrNotSupported{Method: "ClearDefaultLocale"}
}

// ---------------------------------------------------------------------------
// Pipelines
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountPipelines() (*int64, error) {
	return nil, ErrNotSupported{Method: "CountPipelines"}
}

func (r *RemoteDriver) CreatePipeline(_ context.Context, _ audited.AuditContext, _ db.CreatePipelineParams) (*db.Pipeline, error) {
	return nil, ErrNotSupported{Method: "CreatePipeline"}
}

func (r *RemoteDriver) CreatePipelineTable() error {
	return ErrNotSupported{Method: "CreatePipelineTable"}
}

func (r *RemoteDriver) DeletePipeline(_ context.Context, _ audited.AuditContext, _ types.PipelineID) error {
	return ErrNotSupported{Method: "DeletePipeline"}
}

func (r *RemoteDriver) DeletePipelinesByPluginID(_ context.Context, _ audited.AuditContext, _ types.PluginID) error {
	return ErrNotSupported{Method: "DeletePipelinesByPluginID"}
}

func (r *RemoteDriver) GetPipeline(_ types.PipelineID) (*db.Pipeline, error) {
	return nil, ErrNotSupported{Method: "GetPipeline"}
}

func (r *RemoteDriver) ListPipelines() (*[]db.Pipeline, error) {
	return nil, ErrNotSupported{Method: "ListPipelines"}
}

func (r *RemoteDriver) ListPipelinesByTable(_ string) (*[]db.Pipeline, error) {
	return nil, ErrNotSupported{Method: "ListPipelinesByTable"}
}

func (r *RemoteDriver) ListPipelinesByPluginID(_ types.PluginID) (*[]db.Pipeline, error) {
	return nil, ErrNotSupported{Method: "ListPipelinesByPluginID"}
}

func (r *RemoteDriver) ListPipelinesByTableOperation(_ string, _ string) (*[]db.Pipeline, error) {
	return nil, ErrNotSupported{Method: "ListPipelinesByTableOperation"}
}

func (r *RemoteDriver) ListEnabledPipelines() (*[]db.Pipeline, error) {
	return nil, ErrNotSupported{Method: "ListEnabledPipelines"}
}

func (r *RemoteDriver) UpdatePipeline(_ context.Context, _ audited.AuditContext, _ db.UpdatePipelineParams) error {
	return ErrNotSupported{Method: "UpdatePipeline"}
}

func (r *RemoteDriver) UpdatePipelineEnabled(_ context.Context, _ audited.AuditContext, _ types.PipelineID, _ bool) error {
	return ErrNotSupported{Method: "UpdatePipelineEnabled"}
}

// ---------------------------------------------------------------------------
// Webhooks
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountWebhooks() (*int64, error) {
	return doRead(r, func() (*int64, error) {
		ctx := context.Background()
		count, err := r.client.Webhooks.Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: CountWebhooks: %w", err)
		}
		return &count, nil
	})
}

func (r *RemoteDriver) CreateWebhook(ctx context.Context, _ audited.AuditContext, params db.CreateWebhookParams) (*db.Webhook, error) {
	return doWrite(r, func() (*db.Webhook, error) {
		sdkParams := modula.CreateWebhookRequest{
			Name:     params.Name,
			URL:      params.URL,
			Secret:   params.Secret,
			Events:   params.Events,
			IsActive: params.IsActive,
			Headers:  params.Headers,
		}
		result, err := r.client.Webhooks.Create(ctx, sdkParams)
		if err != nil {
			return nil, fmt.Errorf("remote: CreateWebhook: %w", err)
		}
		row := webhookToDb(result)
		return &row, nil
	})
}

func (r *RemoteDriver) CreateWebhookTable() error {
	return ErrNotSupported{Method: "CreateWebhookTable"}
}

func (r *RemoteDriver) DeleteWebhook(ctx context.Context, _ audited.AuditContext, id types.WebhookID) error {
	return doWriteErr(r, func() error {
		if err := r.client.Webhooks.Delete(ctx, modula.WebhookID(string(id))); err != nil {
			return fmt.Errorf("remote: DeleteWebhook: %w", err)
		}
		return nil
	})
}

func (r *RemoteDriver) GetWebhook(id types.WebhookID) (*db.Webhook, error) {
	return doRead(r, func() (*db.Webhook, error) {
		ctx := context.Background()
		sdkID := modula.WebhookID(string(id))
		item, err := r.client.Webhooks.Get(ctx, sdkID)
		if err != nil {
			return nil, fmt.Errorf("remote: GetWebhook: %w", err)
		}
		result := webhookToDb(item)
		return &result, nil
	})
}

func (r *RemoteDriver) ListWebhooks() (*[]db.Webhook, error) {
	return doRead(r, func() (*[]db.Webhook, error) {
		ctx := context.Background()
		items, err := r.client.Webhooks.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListWebhooks: %w", err)
		}
		result := make([]db.Webhook, len(items))
		for i := range items {
			result[i] = webhookToDb(&items[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListActiveWebhooks() (*[]db.Webhook, error) {
	return doRead(r, func() (*[]db.Webhook, error) {
		ctx := context.Background()
		items, err := r.client.Webhooks.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("remote: ListActiveWebhooks: %w", err)
		}
		var result []db.Webhook
		for i := range items {
			if items[i].IsActive {
				result = append(result, webhookToDb(&items[i]))
			}
		}
		return &result, nil
	})
}

func (r *RemoteDriver) ListWebhooksPaginated(p db.PaginationParams) (*[]db.Webhook, error) {
	return doRead(r, func() (*[]db.Webhook, error) {
		ctx := context.Background()
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", p.Limit))
		params.Set("offset", fmt.Sprintf("%d", p.Offset))
		raw, err := r.client.Webhooks.RawList(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("remote: ListWebhooksPaginated: %w", err)
		}
		var sdkItems []modula.Webhook
		if err := json.Unmarshal(raw, &sdkItems); err != nil {
			return nil, fmt.Errorf("remote: ListWebhooksPaginated: decode: %w", err)
		}
		result := make([]db.Webhook, len(sdkItems))
		for i := range sdkItems {
			result[i] = webhookToDb(&sdkItems[i])
		}
		return &result, nil
	})
}

func (r *RemoteDriver) UpdateWebhook(ctx context.Context, _ audited.AuditContext, params db.UpdateWebhookParams) error {
	return doWriteErr(r, func() error {
		sdkParams := modula.UpdateWebhookRequest{
			WebhookID: modula.WebhookID(string(params.WebhookID)),
			Name:      params.Name,
			URL:       params.URL,
			Secret:    params.Secret,
			Events:    params.Events,
			IsActive:  params.IsActive,
			Headers:   params.Headers,
		}
		if _, err := r.client.Webhooks.Update(ctx, sdkParams); err != nil {
			return fmt.Errorf("remote: UpdateWebhook: %w", err)
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// Webhook Deliveries
// ---------------------------------------------------------------------------

func (r *RemoteDriver) CountWebhookDeliveries() (*int64, error) {
	return nil, ErrNotSupported{Method: "CountWebhookDeliveries"}
}

func (r *RemoteDriver) CreateWebhookDelivery(_ context.Context, _ db.CreateWebhookDeliveryParams) (*db.WebhookDelivery, error) {
	return nil, ErrNotSupported{Method: "CreateWebhookDelivery"}
}

func (r *RemoteDriver) CreateWebhookDeliveryTable() error {
	return ErrNotSupported{Method: "CreateWebhookDeliveryTable"}
}

func (r *RemoteDriver) DeleteWebhookDelivery(_ context.Context, _ types.WebhookDeliveryID) error {
	return ErrNotSupported{Method: "DeleteWebhookDelivery"}
}

func (r *RemoteDriver) GetWebhookDelivery(_ types.WebhookDeliveryID) (*db.WebhookDelivery, error) {
	return nil, ErrNotSupported{Method: "GetWebhookDelivery"}
}

func (r *RemoteDriver) ListWebhookDeliveries() (*[]db.WebhookDelivery, error) {
	return nil, ErrNotSupported{Method: "ListWebhookDeliveries"}
}

func (r *RemoteDriver) ListWebhookDeliveriesByWebhook(_ types.WebhookID) (*[]db.WebhookDelivery, error) {
	return nil, ErrNotSupported{Method: "ListWebhookDeliveriesByWebhook"}
}

func (r *RemoteDriver) ListPendingRetries(_ types.Timestamp, _ int64) (*[]db.WebhookDelivery, error) {
	return nil, ErrNotSupported{Method: "ListPendingRetries"}
}

func (r *RemoteDriver) UpdateWebhookDeliveryStatus(_ context.Context, _ db.UpdateWebhookDeliveryStatusParams) error {
	return ErrNotSupported{Method: "UpdateWebhookDeliveryStatus"}
}

func (r *RemoteDriver) PruneOldDeliveries(_ context.Context, _ types.Timestamp) error {
	return ErrNotSupported{Method: "PruneOldDeliveries"}
}
