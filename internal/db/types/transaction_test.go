package types

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// openTestDB creates an in-memory SQLite database for testing transactions.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open error = %v", err)
	}
	t.Cleanup(func() {
		if cerr := db.Close(); cerr != nil {
			t.Errorf("db.Close error = %v", cerr)
		}
	})
	// Create a test table
	_, err = db.Exec("CREATE TABLE test_items (id INTEGER PRIMARY KEY, name TEXT NOT NULL)")
	if err != nil {
		t.Fatalf("CREATE TABLE error = %v", err)
	}
	return db
}

func TestWithTransaction_Commit(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	ctx := context.Background()

	err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO test_items (id, name) VALUES (1, 'alice')")
		return err
	})
	if err != nil {
		t.Fatalf("WithTransaction error = %v", err)
	}

	// Verify row was committed
	var name string
	err = db.QueryRow("SELECT name FROM test_items WHERE id = 1").Scan(&name)
	if err != nil {
		t.Fatalf("SELECT error = %v", err)
	}
	if name != "alice" {
		t.Errorf("name = %q, want %q", name, "alice")
	}
}

func TestWithTransaction_Rollback(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	ctx := context.Background()

	intentionalErr := errors.New("intentional failure")
	err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
		_, execErr := tx.Exec("INSERT INTO test_items (id, name) VALUES (1, 'bob')")
		if execErr != nil {
			return execErr
		}
		return intentionalErr
	})
	if !errors.Is(err, intentionalErr) {
		t.Fatalf("WithTransaction error = %v, want %v", err, intentionalErr)
	}

	// Verify row was NOT committed
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_items").Scan(&count)
	if err != nil {
		t.Fatalf("SELECT COUNT error = %v", err)
	}
	if count != 0 {
		t.Errorf("row count = %d, want 0 (should have rolled back)", count)
	}
}

func TestWithTransaction_CancelledContext(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := WithTransaction(ctx, db, func(tx *sql.Tx) error {
		t.Fatal("fn should not be called with cancelled context")
		return nil
	})
	if err == nil {
		t.Fatal("WithTransaction with cancelled context expected error")
	}
}

func TestWithTransactionResult_Commit(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	ctx := context.Background()

	result, err := WithTransactionResult[int64](ctx, db, func(tx *sql.Tx) (int64, error) {
		res, execErr := tx.Exec("INSERT INTO test_items (id, name) VALUES (1, 'charlie')")
		if execErr != nil {
			return 0, execErr
		}
		return res.LastInsertId()
	})
	if err != nil {
		t.Fatalf("WithTransactionResult error = %v", err)
	}
	if result != 1 {
		t.Errorf("result = %d, want 1", result)
	}

	// Verify committed
	var name string
	err = db.QueryRow("SELECT name FROM test_items WHERE id = 1").Scan(&name)
	if err != nil {
		t.Fatalf("SELECT error = %v", err)
	}
	if name != "charlie" {
		t.Errorf("name = %q, want %q", name, "charlie")
	}
}

func TestWithTransactionResult_Rollback(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	ctx := context.Background()

	intentionalErr := errors.New("intentional failure")
	result, err := WithTransactionResult[string](ctx, db, func(tx *sql.Tx) (string, error) {
		_, execErr := tx.Exec("INSERT INTO test_items (id, name) VALUES (1, 'dave')")
		if execErr != nil {
			return "", execErr
		}
		return "partial", intentionalErr
	})
	if !errors.Is(err, intentionalErr) {
		t.Fatalf("WithTransactionResult error = %v, want %v", err, intentionalErr)
	}
	// Result should still be returned even on error (the "partial" value)
	if result != "partial" {
		t.Errorf("result = %q, want %q", result, "partial")
	}

	// Verify row was NOT committed
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM test_items").Scan(&count)
	if err != nil {
		t.Fatalf("SELECT COUNT error = %v", err)
	}
	if count != 0 {
		t.Errorf("row count = %d, want 0", count)
	}
}

func TestWithTransactionResult_CancelledContext(t *testing.T) {
	t.Parallel()
	db := openTestDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := WithTransactionResult[int](ctx, db, func(tx *sql.Tx) (int, error) {
		t.Fatal("fn should not be called")
		return 0, nil
	})
	if err == nil {
		t.Fatal("WithTransactionResult with cancelled context expected error")
	}
}
