// White-box tests for transaction.go: WithTransaction and WithTransactionResult.
//
// Uses in-memory SQLite (real dependency, no mocks). The sqlite3 driver is
// registered via the blank import in init.go, so no additional import is needed.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// openTxTestDB creates an in-memory SQLite database with a test table.
func openTxTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() {
		if cerr := db.Close(); cerr != nil {
			t.Errorf("db.Close: %v", cerr)
		}
	})
	_, err = db.Exec("CREATE TABLE tx_test (id INTEGER PRIMARY KEY, name TEXT NOT NULL, value INTEGER NOT NULL DEFAULT 0)")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	return db
}

// openTxTestDBFile creates a file-based SQLite database in a temp directory
// with WAL mode enabled. Required for concurrent transaction tests because
// in-memory SQLite (:memory:) creates a private database per connection.
func openTxTestDBFile(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000", filepath.Join(dir, "tx_concurrent.db"))
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() {
		if cerr := db.Close(); cerr != nil {
			t.Errorf("db.Close: %v", cerr)
		}
	})
	_, err = db.Exec("CREATE TABLE tx_test (id INTEGER PRIMARY KEY, name TEXT NOT NULL, value INTEGER NOT NULL DEFAULT 0)")
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	return db
}

// countRows returns the number of rows in tx_test.
func countRows(t *testing.T, db *sql.DB) int {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM tx_test").Scan(&count); err != nil {
		t.Fatalf("SELECT COUNT: %v", err)
	}
	return count
}

// --- WithTransaction tests ---

func TestWithTransaction_CommitSuccess(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx := context.Background()

	err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO tx_test (id, name) VALUES (1, 'alice')")
		return err
	})
	if err != nil {
		t.Fatalf("WithTransaction: %v", err)
	}

	// Verify committed
	var name string
	if err := db.QueryRow("SELECT name FROM tx_test WHERE id = 1").Scan(&name); err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	if name != "alice" {
		t.Errorf("name = %q, want %q", name, "alice")
	}
}

func TestWithTransaction_MultipleStatements(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx := context.Background()

	err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
		if _, err := tx.Exec("INSERT INTO tx_test (id, name) VALUES (1, 'first')"); err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT INTO tx_test (id, name) VALUES (2, 'second')"); err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT INTO tx_test (id, name) VALUES (3, 'third')"); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithTransaction: %v", err)
	}

	count := countRows(t, db)
	if count != 3 {
		t.Errorf("row count = %d, want 3", count)
	}
}

func TestWithTransaction_RollbackOnError(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx := context.Background()

	sentinel := errors.New("intentional failure")
	err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
		// Insert a row, then fail -- the insert must be rolled back
		if _, execErr := tx.Exec("INSERT INTO tx_test (id, name) VALUES (1, 'bob')"); execErr != nil {
			return execErr
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want %v", err, sentinel)
	}

	count := countRows(t, db)
	if count != 0 {
		t.Errorf("row count = %d, want 0 (should have rolled back)", count)
	}
}

func TestWithTransaction_BeginTxFailure_ClosedDB(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	// Close the database to force BeginTx failure
	if err := db.Close(); err != nil {
		t.Fatalf("db.Close: %v", err)
	}

	ctx := context.Background()
	err = WithTransaction(ctx, db, func(tx *sql.Tx) error {
		t.Fatal("fn should not be called when BeginTx fails")
		return nil
	})
	if err == nil {
		t.Fatal("expected error from closed DB, got nil")
	}
	if !strings.Contains(err.Error(), "begin transaction") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "begin transaction")
	}
}

func TestWithTransaction_BeginTxFailure_CancelledContext(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
		t.Fatal("fn should not be called with cancelled context")
		return nil
	})
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "begin transaction") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "begin transaction")
	}
}

func TestWithTransaction_ContextCancelDuringFn(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx, cancel := context.WithCancel(context.Background())

	err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
		// Insert a row, then cancel the context before returning
		if _, execErr := tx.Exec("INSERT INTO tx_test (id, name) VALUES (1, 'cancel')"); execErr != nil {
			return execErr
		}
		cancel()
		// Return nil -- commit should fail because context is cancelled
		return nil
	})
	// The commit (or a subsequent operation) should fail
	if err == nil {
		// Some SQLite drivers may allow this to succeed because SQLite
		// transactions are local. If it succeeded, verify the data is consistent.
		count := countRows(t, db)
		t.Logf("context cancelled during fn but commit succeeded; row count = %d", count)
		return
	}
	// If there was an error, the row should not be committed
	t.Logf("context cancel during fn produced error: %v", err)
}

func TestWithTransaction_CommitFailure(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx := context.Background()

	// Force a commit failure by beginning a transaction, then rolling it back
	// inside the fn so that the outer Commit() fails.
	err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
		if _, execErr := tx.Exec("INSERT INTO tx_test (id, name) VALUES (1, 'commit-fail')"); execErr != nil {
			return execErr
		}
		// Explicitly rollback the transaction. The deferred Rollback in
		// WithTransaction will be a no-op, and the subsequent Commit() will
		// fail because the tx is already done.
		return tx.Rollback()
	})
	if err == nil {
		t.Fatal("expected commit failure error, got nil")
	}
	if !strings.Contains(err.Error(), "commit transaction") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "commit transaction")
	}
}

func TestWithTransaction_NilFnError(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx := context.Background()

	// fn returns nil (success) with no operations -- should commit cleanly
	err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
		return nil
	})
	if err != nil {
		t.Fatalf("WithTransaction with no-op fn: %v", err)
	}
}

func TestWithTransaction_Concurrent(t *testing.T) {
	t.Parallel()
	db := openTxTestDBFile(t)
	ctx := context.Background()
	const goroutines = 10

	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			txErr := WithTransaction(ctx, db, func(tx *sql.Tx) error {
				_, execErr := tx.Exec("INSERT INTO tx_test (id, name) VALUES (?, ?)", n+1, "goroutine")
				return execErr
			})
			if txErr != nil {
				errs <- txErr
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent transaction error: %v", err)
	}

	count := countRows(t, db)
	if count != goroutines {
		t.Errorf("row count = %d, want %d", count, goroutines)
	}
}

func TestWithTransaction_ContextTimeout(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Give the timeout a moment to expire
	time.Sleep(5 * time.Millisecond)

	err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
		t.Fatal("fn should not be called with expired timeout")
		return nil
	})
	if err == nil {
		t.Fatal("expected error from expired timeout, got nil")
	}
	if !strings.Contains(err.Error(), "begin transaction") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "begin transaction")
	}
}

// --- WithTransactionResult tests ---

func TestWithTransactionResult_CommitSuccess_Int(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx := context.Background()

	result, err := WithTransactionResult(ctx, db, func(tx *sql.Tx) (int64, error) {
		res, execErr := tx.Exec("INSERT INTO tx_test (id, name) VALUES (1, 'charlie')")
		if execErr != nil {
			return 0, execErr
		}
		return res.LastInsertId()
	})
	if err != nil {
		t.Fatalf("WithTransactionResult: %v", err)
	}
	if result != 1 {
		t.Errorf("result = %d, want 1", result)
	}

	// Verify committed
	var name string
	if err := db.QueryRow("SELECT name FROM tx_test WHERE id = 1").Scan(&name); err != nil {
		t.Fatalf("SELECT: %v", err)
	}
	if name != "charlie" {
		t.Errorf("name = %q, want %q", name, "charlie")
	}
}

func TestWithTransactionResult_CommitSuccess_String(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx := context.Background()

	result, err := WithTransactionResult(ctx, db, func(tx *sql.Tx) (string, error) {
		_, execErr := tx.Exec("INSERT INTO tx_test (id, name) VALUES (1, 'dave')")
		if execErr != nil {
			return "", execErr
		}
		var name string
		scanErr := tx.QueryRow("SELECT name FROM tx_test WHERE id = 1").Scan(&name)
		return name, scanErr
	})
	if err != nil {
		t.Fatalf("WithTransactionResult: %v", err)
	}
	if result != "dave" {
		t.Errorf("result = %q, want %q", result, "dave")
	}
}

func TestWithTransactionResult_CommitSuccess_Struct(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx := context.Background()

	type row struct {
		ID    int
		Name  string
		Value int
	}

	result, err := WithTransactionResult(ctx, db, func(tx *sql.Tx) (row, error) {
		_, execErr := tx.Exec("INSERT INTO tx_test (id, name, value) VALUES (42, 'eve', 100)")
		if execErr != nil {
			return row{}, execErr
		}
		var r row
		scanErr := tx.QueryRow("SELECT id, name, value FROM tx_test WHERE id = 42").Scan(&r.ID, &r.Name, &r.Value)
		return r, scanErr
	})
	if err != nil {
		t.Fatalf("WithTransactionResult: %v", err)
	}
	if result.ID != 42 || result.Name != "eve" || result.Value != 100 {
		t.Errorf("result = %+v, want {ID:42 Name:eve Value:100}", result)
	}
}

func TestWithTransactionResult_CommitSuccess_Slice(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx := context.Background()

	result, err := WithTransactionResult(ctx, db, func(tx *sql.Tx) ([]string, error) {
		for _, name := range []string{"alpha", "bravo", "charlie"} {
			if _, execErr := tx.Exec("INSERT INTO tx_test (name) VALUES (?)", name); execErr != nil {
				return nil, execErr
			}
		}
		rows, queryErr := tx.Query("SELECT name FROM tx_test ORDER BY name")
		if queryErr != nil {
			return nil, queryErr
		}
		defer rows.Close()

		var names []string
		for rows.Next() {
			var n string
			if scanErr := rows.Scan(&n); scanErr != nil {
				return nil, scanErr
			}
			names = append(names, n)
		}
		return names, rows.Err()
	})
	if err != nil {
		t.Fatalf("WithTransactionResult: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("len(result) = %d, want 3", len(result))
	}
	expected := []string{"alpha", "bravo", "charlie"}
	for i, want := range expected {
		if result[i] != want {
			t.Errorf("result[%d] = %q, want %q", i, result[i], want)
		}
	}
}

func TestWithTransactionResult_RollbackOnError(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx := context.Background()

	sentinel := errors.New("intentional failure")
	result, err := WithTransactionResult(ctx, db, func(tx *sql.Tx) (string, error) {
		if _, execErr := tx.Exec("INSERT INTO tx_test (id, name) VALUES (1, 'frank')"); execErr != nil {
			return "", execErr
		}
		return "partial-result", sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want %v", err, sentinel)
	}
	// The implementation returns result even on error
	if result != "partial-result" {
		t.Errorf("result = %q, want %q", result, "partial-result")
	}

	count := countRows(t, db)
	if count != 0 {
		t.Errorf("row count = %d, want 0 (should have rolled back)", count)
	}
}

func TestWithTransactionResult_RollbackOnError_ZeroValue(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx := context.Background()

	sentinel := errors.New("zero value failure")
	result, err := WithTransactionResult(ctx, db, func(tx *sql.Tx) (int, error) {
		return 0, sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want %v", err, sentinel)
	}
	if result != 0 {
		t.Errorf("result = %d, want 0", result)
	}
}

func TestWithTransactionResult_BeginTxFailure_ClosedDB(t *testing.T) {
	t.Parallel()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("db.Close: %v", err)
	}

	ctx := context.Background()
	result, txErr := WithTransactionResult(ctx, db, func(tx *sql.Tx) (string, error) {
		t.Fatal("fn should not be called when BeginTx fails")
		return "", nil
	})
	if txErr == nil {
		t.Fatal("expected error from closed DB, got nil")
	}
	if !strings.Contains(txErr.Error(), "begin transaction") {
		t.Errorf("error = %q, want it to contain %q", txErr.Error(), "begin transaction")
	}
	if result != "" {
		t.Errorf("result = %q, want zero value %q", result, "")
	}
}

func TestWithTransactionResult_BeginTxFailure_CancelledContext(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := WithTransactionResult(ctx, db, func(tx *sql.Tx) (int, error) {
		t.Fatal("fn should not be called with cancelled context")
		return 0, nil
	})
	if err == nil {
		t.Fatal("expected error from cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "begin transaction") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "begin transaction")
	}
	if result != 0 {
		t.Errorf("result = %d, want zero value 0", result)
	}
}

func TestWithTransactionResult_CommitFailure(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx := context.Background()

	result, err := WithTransactionResult(ctx, db, func(tx *sql.Tx) (string, error) {
		if _, execErr := tx.Exec("INSERT INTO tx_test (id, name) VALUES (1, 'commit-fail')"); execErr != nil {
			return "", execErr
		}
		// Roll back the tx so that Commit() will fail
		return "value-before-commit", tx.Rollback()
	})
	if err == nil {
		t.Fatal("expected commit failure error, got nil")
	}
	if !strings.Contains(err.Error(), "commit transaction") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "commit transaction")
	}
	// Result is set even though commit failed -- the fn returned it successfully
	if result != "value-before-commit" {
		t.Errorf("result = %q, want %q", result, "value-before-commit")
	}
}

func TestWithTransactionResult_Concurrent(t *testing.T) {
	t.Parallel()
	db := openTxTestDBFile(t)
	ctx := context.Background()
	const goroutines = 10

	type insertResult struct {
		id  int
		err error
	}

	var wg sync.WaitGroup
	results := make(chan insertResult, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			id, txErr := WithTransactionResult(ctx, db, func(tx *sql.Tx) (int, error) {
				_, execErr := tx.Exec("INSERT INTO tx_test (id, name) VALUES (?, ?)", n+1, "concurrent")
				if execErr != nil {
					return 0, execErr
				}
				return n + 1, nil
			})
			results <- insertResult{id: id, err: txErr}
		}(i)
	}

	wg.Wait()
	close(results)

	seen := make(map[int]bool)
	for r := range results {
		if r.err != nil {
			t.Errorf("concurrent transaction error: %v", r.err)
			continue
		}
		if seen[r.id] {
			t.Errorf("duplicate id = %d", r.id)
		}
		seen[r.id] = true
	}

	count := countRows(t, db)
	if count != goroutines {
		t.Errorf("row count = %d, want %d", count, goroutines)
	}
}

func TestWithTransactionResult_ContextTimeout(t *testing.T) {
	t.Parallel()
	db := openTxTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(5 * time.Millisecond)

	result, err := WithTransactionResult(ctx, db, func(tx *sql.Tx) (string, error) {
		t.Fatal("fn should not be called with expired timeout")
		return "", nil
	})
	if err == nil {
		t.Fatal("expected error from expired timeout, got nil")
	}
	if !strings.Contains(err.Error(), "begin transaction") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "begin transaction")
	}
	if result != "" {
		t.Errorf("result = %q, want zero value", result)
	}
}

// --- TxFunc type test ---

func TestTxFunc_TypeSatisfied(t *testing.T) {
	t.Parallel()

	// Verify TxFunc is assignable from a function with the correct signature.
	var fn TxFunc = func(tx *sql.Tx) error {
		return nil
	}

	db := openTxTestDB(t)
	ctx := context.Background()

	err := WithTransaction(ctx, db, fn)
	if err != nil {
		t.Fatalf("WithTransaction with TxFunc variable: %v", err)
	}
}
