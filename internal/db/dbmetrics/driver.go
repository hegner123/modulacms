package dbmetrics

import (
	"context"
	"database/sql/driver"
	"time"
)

// metricsDriver wraps a database/sql/driver.Driver to record query metrics
// on every connection it produces.
type metricsDriver struct {
	inner      driver.Driver
	driverName string // "sqlite", "mysql", "postgres"
}

func (d *metricsDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.inner.Open(name)
	if err != nil {
		return nil, err
	}
	return &metricsConn{inner: conn, driverName: d.driverName}, nil
}

// metricsConn wraps a driver.Conn to intercept query execution.
// Implements the fast-path interfaces (ExecerContext, QueryerContext) so
// database/sql calls these directly instead of falling back to Prepare.
type metricsConn struct {
	inner      driver.Conn
	driverName string
}

func (c *metricsConn) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.inner.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &metricsStmt{inner: stmt, query: query, driverName: c.driverName}, nil
}

func (c *metricsConn) Close() error {
	return c.inner.Close()
}

func (c *metricsConn) Begin() (driver.Tx, error) {
	tx, err := c.inner.Begin() //nolint:staticcheck // required by driver.Conn interface
	if err != nil {
		return nil, err
	}
	return &metricsTx{inner: tx}, nil
}

// BeginTx implements driver.ConnBeginTx for context-aware transaction starts.
func (c *metricsConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if cbt, ok := c.inner.(driver.ConnBeginTx); ok {
		tx, err := cbt.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &metricsTx{inner: tx}, nil
	}
	return c.Begin()
}

// ExecContext implements driver.ExecerContext — the fast path for Exec calls.
func (c *metricsConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if ec, ok := c.inner.(driver.ExecerContext); ok {
		start := time.Now()
		result, err := ec.ExecContext(ctx, query, args)
		RecordQueryMetrics(query, c.driverName, time.Since(start), err)
		return result, err
	}
	// Fallback: database/sql will use Prepare → Stmt.Exec instead.
	// Return driver.ErrSkip to signal this.
	return nil, driver.ErrSkip
}

// QueryContext implements driver.QueryerContext — the fast path for Query calls.
func (c *metricsConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if qc, ok := c.inner.(driver.QueryerContext); ok {
		start := time.Now()
		rows, err := qc.QueryContext(ctx, query, args)
		RecordQueryMetrics(query, c.driverName, time.Since(start), err)
		return rows, err
	}
	return nil, driver.ErrSkip
}

// PrepareContext implements driver.ConnPrepareContext.
func (c *metricsConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if pc, ok := c.inner.(driver.ConnPrepareContext); ok {
		stmt, err := pc.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		return &metricsStmt{inner: stmt, query: query, driverName: c.driverName}, nil
	}
	return c.Prepare(query)
}

// Ping implements driver.Pinger if the underlying connection supports it.
func (c *metricsConn) Ping(ctx context.Context) error {
	if p, ok := c.inner.(driver.Pinger); ok {
		return p.Ping(ctx)
	}
	return nil
}

// ResetSession implements driver.SessionResetter if the underlying connection supports it.
func (c *metricsConn) ResetSession(ctx context.Context) error {
	if rs, ok := c.inner.(driver.SessionResetter); ok {
		return rs.ResetSession(ctx)
	}
	return nil
}

// IsValid implements driver.Validator if the underlying connection supports it.
func (c *metricsConn) IsValid() bool {
	if v, ok := c.inner.(driver.Validator); ok {
		return v.IsValid()
	}
	return true
}

// metricsTx wraps driver.Tx. No query interception needed — queries within
// transactions flow through the same metricsConn methods.
type metricsTx struct {
	inner driver.Tx
}

func (t *metricsTx) Commit() error   { return t.inner.Commit() }
func (t *metricsTx) Rollback() error { return t.inner.Rollback() }

// metricsStmt wraps driver.Stmt to record metrics using the query string
// captured at Prepare time.
type metricsStmt struct {
	inner      driver.Stmt
	query      string
	driverName string
}

func (s *metricsStmt) Close() error                               { return s.inner.Close() }
func (s *metricsStmt) NumInput() int                              { return s.inner.NumInput() }

func (s *metricsStmt) Exec(args []driver.Value) (driver.Result, error) {
	start := time.Now()
	result, err := s.inner.Exec(args) //nolint:staticcheck // required by driver.Stmt interface
	RecordQueryMetrics(s.query, s.driverName, time.Since(start), err)
	return result, err
}

func (s *metricsStmt) Query(args []driver.Value) (driver.Rows, error) {
	start := time.Now()
	rows, err := s.inner.Query(args) //nolint:staticcheck // required by driver.Stmt interface
	RecordQueryMetrics(s.query, s.driverName, time.Since(start), err)
	return rows, err
}

// ExecContext implements driver.StmtExecContext.
func (s *metricsStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	if sec, ok := s.inner.(driver.StmtExecContext); ok {
		start := time.Now()
		result, err := sec.ExecContext(ctx, args)
		RecordQueryMetrics(s.query, s.driverName, time.Since(start), err)
		return result, err
	}
	// Convert NamedValue to Value for fallback.
	vals := make([]driver.Value, len(args))
	for i, nv := range args {
		vals[i] = nv.Value
	}
	return s.Exec(vals)
}

// QueryContext implements driver.StmtQueryContext.
func (s *metricsStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	if sqc, ok := s.inner.(driver.StmtQueryContext); ok {
		start := time.Now()
		rows, err := sqc.QueryContext(ctx, args)
		RecordQueryMetrics(s.query, s.driverName, time.Since(start), err)
		return rows, err
	}
	vals := make([]driver.Value, len(args))
	for i, nv := range args {
		vals[i] = nv.Value
	}
	return s.Query(vals)
}
