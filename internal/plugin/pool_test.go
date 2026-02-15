package plugin

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// -- Test helpers --

// dbFunctionNames lists all db module functions checked by validateVM.
var dbFunctionNames = []string{
	"query", "query_one", "count", "exists",
	"insert", "update", "delete", "transaction", "define_table",
}

// newTestVMFactory returns a factory function that creates sandboxed LStates
// with minimal db and log tables populated with Go-bound stub functions.
// This matches the shape expected by validateVM without requiring the full
// db_api.go or log_api.go registration.
func newTestVMFactory() func() *lua.LState {
	return func() *lua.LState {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		ApplySandbox(L, SandboxConfig{})

		// Register minimal db module with Go-bound stubs for all validated functions.
		dbTable := L.NewTable()
		for _, name := range dbFunctionNames {
			dbTable.RawSetString(name, L.NewFunction(func(L *lua.LState) int { return 0 }))
		}
		// Also register ulid and timestamp which are part of the db API but not
		// checked by validateVM -- included for completeness.
		dbTable.RawSetString("ulid", L.NewFunction(func(L *lua.LState) int {
			L.Push(lua.LString("test-ulid"))
			return 1
		}))
		dbTable.RawSetString("timestamp", L.NewFunction(func(L *lua.LState) int {
			L.Push(lua.LString("2026-01-01T00:00:00Z"))
			return 1
		}))
		L.SetGlobal("db", dbTable)

		// Register minimal log module with Go-bound stubs.
		logTable := L.NewTable()
		logTable.RawSetString("info", L.NewFunction(func(L *lua.LState) int { return 0 }))
		logTable.RawSetString("warn", L.NewFunction(func(L *lua.LState) int { return 0 }))
		logTable.RawSetString("error", L.NewFunction(func(L *lua.LState) int { return 0 }))
		logTable.RawSetString("debug", L.NewFunction(func(L *lua.LState) int { return 0 }))
		L.SetGlobal("log", logTable)

		// Freeze modules to match production behavior.
		FreezeModule(L, "db")
		FreezeModule(L, "log")

		return L
	}
}

// -- Tests --

func TestNewVMPool_CreatesRequestedSize(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(3, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	if len(pool.states) != 3 {
		t.Errorf("expected pool to have 3 VMs, got %d", len(pool.states))
	}
}

func TestVMPool_GetSetsContext(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

	// Verify that L.Context() returns the caller's context (not a derived one).
	// The key property is that the returned context has the same deadline.
	gotCtx := L.Context()
	if gotCtx == nil {
		t.Fatal("expected context to be set on LState after Get, got nil")
	}

	gotDeadline, gotOk := gotCtx.Deadline()
	wantDeadline, wantOk := ctx.Deadline()
	if gotOk != wantOk {
		t.Fatalf("context deadline presence mismatch: got %v, want %v", gotOk, wantOk)
	}
	if gotOk && !gotDeadline.Equal(wantDeadline) {
		t.Errorf("context deadline mismatch: got %v, want %v", gotDeadline, wantDeadline)
	}
}

func TestVMPool_GetReturnsErrPoolExhausted(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx := context.Background()

	// Check out the only VM.
	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("first Get failed: %v", err)
	}
	defer pool.Put(L)

	// Second Get should fail within ~100ms, not hang.
	start := time.Now()
	_, err = pool.Get(ctx)
	elapsed := time.Since(start)

	if !errors.Is(err, ErrPoolExhausted) {
		t.Fatalf("expected ErrPoolExhausted, got %v", err)
	}

	// The acquisition timeout is 100ms. Allow some slack for scheduling.
	if elapsed > 300*time.Millisecond {
		t.Errorf("Get took %v, expected ~100ms (acquisition timeout)", elapsed)
	}
	if elapsed < 50*time.Millisecond {
		t.Errorf("Get returned in %v, suspiciously fast (expected ~100ms)", elapsed)
	}
}

func TestVMPool_GetRespectsContextCancellation(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	bgCtx := context.Background()

	// Check out the only VM.
	L, err := pool.Get(bgCtx)
	if err != nil {
		t.Fatalf("first Get failed: %v", err)
	}
	defer pool.Put(L)

	// Create a context that is already cancelled.
	cancelledCtx, cancel := context.WithCancel(bgCtx)
	cancel()

	_, err = pool.Get(cancelledCtx)
	if !errors.Is(err, ErrPoolExhausted) {
		t.Fatalf("expected ErrPoolExhausted with cancelled context, got %v", err)
	}
}

func TestVMPool_ConcurrentGetPut(t *testing.T) {
	factory := newTestVMFactory()
	poolSize := 4
	pool := NewVMPool(poolSize, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	// Take a snapshot so Put() exercises restoreGlobalSnapshot.
	L, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("setup Get failed: %v", err)
	}
	pool.SnapshotGlobals(L)
	pool.Put(L)

	// Launch many goroutines that all compete for VMs.
	const numGoroutines = 20
	const opsPerGoroutine = 10

	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines*opsPerGoroutine)

	for range numGoroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range opsPerGoroutine {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				vm, getErr := pool.Get(ctx)
				cancel()
				if getErr != nil {
					// Pool exhaustion is expected under heavy contention.
					if errors.Is(getErr, ErrPoolExhausted) {
						continue
					}
					errCh <- getErr
					return
				}

				// Simulate some work: create a global that should be cleaned up.
				vm.SetGlobal("temp_var", lua.LString("temporary"))

				pool.Put(vm)
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for e := range errCh {
		t.Errorf("unexpected error during concurrent access: %v", e)
	}

	// After all goroutines are done, the pool should still have all VMs.
	if len(pool.states) != poolSize {
		t.Errorf("pool has %d VMs, expected %d", len(pool.states), poolSize)
	}
}

func TestVMPool_PutReturnsHealthyVM(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx := context.Background()

	// Get, use, and put back a healthy VM.
	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	pool.Put(L)

	// Should be able to get it again.
	L2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("second Get failed: %v", err)
	}
	defer pool.Put(L2)

	// The returned VM should still be functional.
	if execErr := L2.DoString(`return db.query()`); execErr != nil {
		t.Errorf("VM not functional after Put/Get cycle: %v", execErr)
	}
}

func TestVMPool_PutReplacesUnhealthyVM_CorruptedDB(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Corrupt the db module by replacing the global with a string.
	// We need to bypass the proxy protection to simulate corruption.
	// Since the proxy has __newindex protection, we use L.SetGlobal directly
	// which sets at the Go level, bypassing Lua metamethods.
	L.SetGlobal("db", lua.LString("corrupted"))

	// Put should detect corruption and replace the VM.
	pool.Put(L)

	// The pool should still have 1 VM (the replacement).
	if len(pool.states) != 1 {
		t.Fatalf("pool has %d VMs after unhealthy Put, expected 1", len(pool.states))
	}

	// Get the replacement and verify it is healthy.
	L2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get after replacement failed: %v", err)
	}
	defer pool.Put(L2)

	// The replacement should be a fresh VM with intact db module.
	if !pool.validateVM(L2) {
		t.Error("replacement VM failed validation")
	}
}

func TestVMPool_PutReplacesUnhealthyVM_CorruptedLog(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Corrupt the log module by replacing the global with nil.
	L.SetGlobal("log", lua.LNil)

	pool.Put(L)

	// Pool should still have 1 VM.
	if len(pool.states) != 1 {
		t.Fatalf("pool has %d VMs after unhealthy Put, expected 1", len(pool.states))
	}

	// The replacement should be healthy.
	L2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get after replacement failed: %v", err)
	}
	defer pool.Put(L2)

	if !pool.validateVM(L2) {
		t.Error("replacement VM failed validation")
	}
}

func TestVMPool_PutReplacesUnhealthyVM_DBFunctionReplacedWithLuaFunc(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Replace the db global with a table that has a Lua function (IsG == false)
	// instead of a Go-bound function. This simulates an attack where the proxy
	// was bypassed and db.query replaced with a pure Lua function.
	fakeDB := L.NewTable()
	for _, name := range dbFunctionNames {
		// Create a pure Lua function (not Go-bound) by compiling Lua code.
		if luaErr := L.DoString(`return function() end`); luaErr != nil {
			t.Fatalf("failed to create Lua function: %v", luaErr)
		}
		luaFn := L.Get(-1)
		L.Pop(1)
		fakeDB.RawSetString(name, luaFn)
	}
	L.SetGlobal("db", fakeDB)

	pool.Put(L)

	// Pool should have a replacement.
	if len(pool.states) != 1 {
		t.Fatalf("pool has %d VMs, expected 1", len(pool.states))
	}

	L2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get after replacement failed: %v", err)
	}
	defer pool.Put(L2)

	if !pool.validateVM(L2) {
		t.Error("replacement VM should be healthy")
	}
}

func TestVMPool_SnapshotGlobals_RemovesNewGlobals(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx := context.Background()

	// Take snapshot of the initial state.
	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	pool.SnapshotGlobals(L)
	pool.Put(L)

	// Get the VM, add a new global, and put it back.
	L, err = pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	L.SetGlobal("leaked_global", lua.LString("should be removed"))
	pool.Put(L)

	// Get the VM again and verify the leaked global was removed.
	L, err = pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

	val := L.GetGlobal("leaked_global")
	if val != lua.LNil {
		t.Errorf("leaked global should be nil after Put, got %s (%s)", val.String(), val.Type())
	}
}

func TestVMPool_SnapshotGlobals_PreservesInitTimeGlobals(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx := context.Background()

	// Before snapshot, add a "plugin_info" global that simulates on_init setup.
	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	L.SetGlobal("plugin_info", lua.LString("my_plugin"))
	pool.SnapshotGlobals(L)
	pool.Put(L)

	// Get the VM, do some work, put it back.
	L, err = pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	L.SetGlobal("temp", lua.LString("temporary"))
	pool.Put(L)

	// Verify plugin_info was preserved but temp was removed.
	L, err = pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

	piVal := L.GetGlobal("plugin_info")
	if piVal == lua.LNil {
		t.Error("plugin_info should be preserved after Put, got nil")
	}
	if piVal.String() != "my_plugin" {
		t.Errorf("plugin_info = %q, want %q", piVal.String(), "my_plugin")
	}

	tempVal := L.GetGlobal("temp")
	if tempVal != lua.LNil {
		t.Errorf("temp should be nil after Put, got %s", tempVal.String())
	}
}

func TestVMPool_SnapshotGlobals_NoSnapshotIsNoOp(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx := context.Background()

	// Without calling SnapshotGlobals, Put should not remove any globals.
	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	L.SetGlobal("custom_global", lua.LString("kept"))
	pool.Put(L)

	L, err = pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

	val := L.GetGlobal("custom_global")
	if val == lua.LNil {
		t.Error("custom_global should be preserved when no snapshot is taken")
	}
}

func TestVMPool_Close_DrainsSafely(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(3, factory, "/test/init.lua", "test_plugin")

	// Check out one VM before closing.
	ctx := context.Background()
	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Close the pool -- should drain the 2 VMs still in the channel.
	pool.Close()

	// The channel should be empty after draining.
	if len(pool.states) != 0 {
		t.Errorf("pool has %d VMs after Close, expected 0", len(pool.states))
	}

	// Return the checked-out VM after Close.
	// Put should detect the closed flag and close the VM directly.
	pool.Put(L)

	// Pool should still be empty (VM was closed, not returned to channel).
	if len(pool.states) != 0 {
		t.Errorf("pool has %d VMs after late Put, expected 0", len(pool.states))
	}
}

func TestVMPool_Close_GetAfterCloseReturnsExhausted(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	pool.Close()

	_, err := pool.Get(context.Background())
	if !errors.Is(err, ErrPoolExhausted) {
		t.Fatalf("expected ErrPoolExhausted after Close, got %v", err)
	}
}

func TestVMPool_PutClearsStack(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Push some values onto the stack.
	L.Push(lua.LString("leftover1"))
	L.Push(lua.LString("leftover2"))
	L.Push(lua.LNumber(42))

	pool.Put(L)

	// Get the VM back and verify stack is clean.
	L, err = pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

	if L.GetTop() != 0 {
		t.Errorf("stack should be empty after Put, got %d items", L.GetTop())
	}
}

func TestVMPool_PutClearsContext(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	pool.Put(L)

	// Get the VM from the channel directly to check its state before Get sets a new context.
	rawL := <-pool.states
	gotCtx := rawL.Context()
	pool.states <- rawL // put it back

	// Context should be nil after Put (cleared to avoid dangling references).
	if gotCtx != nil {
		t.Error("expected context to be nil after Put, but it was set")
	}
}

func TestValidateVM_HealthyVM(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	L, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

	if !pool.validateVM(L) {
		t.Error("healthy VM should pass validation")
	}
}

func TestValidateVM_MissingDB(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	L, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

	L.SetGlobal("db", lua.LNil)
	if pool.validateVM(L) {
		t.Error("VM with nil db should fail validation")
	}
}

func TestValidateVM_DBNotTable(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	L, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

	L.SetGlobal("db", lua.LString("not a table"))
	if pool.validateVM(L) {
		t.Error("VM with string db should fail validation")
	}
}

func TestValidateVM_MissingDBFunction(t *testing.T) {
	// Test each db function individually to ensure all are checked.
	for _, funcName := range dbFunctionNames {
		t.Run(funcName, func(t *testing.T) {
			factory := newTestVMFactory()
			pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
			defer pool.Close()

			L, err := pool.Get(context.Background())
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}
			defer pool.Put(L)

			// Replace the db global with a fresh table missing one function.
			// We build a table with all functions except the one under test.
			fakeDB := L.NewTable()
			for _, name := range dbFunctionNames {
				if name == funcName {
					// Set to a string instead of a function.
					fakeDB.RawSetString(name, lua.LString("not a function"))
				} else {
					fakeDB.RawSetString(name, L.NewFunction(func(L *lua.LState) int { return 0 }))
				}
			}
			L.SetGlobal("db", fakeDB)

			if pool.validateVM(L) {
				t.Errorf("VM with corrupted db.%s should fail validation", funcName)
			}
		})
	}
}

func TestValidateVM_MissingLog(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	L, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

	L.SetGlobal("log", lua.LNil)
	if pool.validateVM(L) {
		t.Error("VM with nil log should fail validation")
	}
}

func TestValidateVM_LogInfoNotGoFunction(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	L, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

	// Replace log with a table where info is a pure Lua function.
	if luaErr := L.DoString(`return function() end`); luaErr != nil {
		t.Fatalf("failed to create Lua function: %v", luaErr)
	}
	luaFn := L.Get(-1)
	L.Pop(1)

	fakeLog := L.NewTable()
	fakeLog.RawSetString("info", luaFn)
	L.SetGlobal("log", fakeLog)

	if pool.validateVM(L) {
		t.Error("VM with Lua (not Go-bound) log.info should fail validation")
	}
}

func TestVMPool_MultipleSnapshotRestoreCycles(t *testing.T) {
	// Verify that snapshot/restore works correctly over multiple Get/Put cycles
	// with different globals being created and cleaned up each time.
	factory := newTestVMFactory()
	pool := NewVMPool(1, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx := context.Background()

	// Take initial snapshot.
	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	pool.SnapshotGlobals(L)
	pool.Put(L)

	// Cycle 1: set global_a, verify it is cleaned up.
	L, err = pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	L.SetGlobal("global_a", lua.LString("a"))
	pool.Put(L)

	L, err = pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if L.GetGlobal("global_a") != lua.LNil {
		t.Error("global_a should be nil after cycle 1 Put")
	}
	pool.Put(L)

	// Cycle 2: set global_b, verify it is cleaned up.
	L, err = pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	L.SetGlobal("global_b", lua.LNumber(42))
	pool.Put(L)

	L, err = pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)
	if L.GetGlobal("global_b") != lua.LNil {
		t.Error("global_b should be nil after cycle 2 Put")
	}
	// Also verify global_a is still nil.
	if L.GetGlobal("global_a") != lua.LNil {
		t.Error("global_a should still be nil")
	}
}

func TestVMPool_PoolSizeStaysConstantAfterReplacement(t *testing.T) {
	factory := newTestVMFactory()
	pool := NewVMPool(2, factory, "/test/init.lua", "test_plugin")
	defer pool.Close()

	ctx := context.Background()

	// Get both VMs.
	L1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get L1 failed: %v", err)
	}
	L2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get L2 failed: %v", err)
	}

	// Corrupt L1, keep L2 healthy.
	L1.SetGlobal("db", lua.LNil)

	// Return both.
	pool.Put(L1) // should replace
	pool.Put(L2) // should return normally

	// Pool should have exactly 2 VMs.
	if len(pool.states) != 2 {
		t.Errorf("pool has %d VMs, expected 2 (constant after replacement)", len(pool.states))
	}

	// Both should be healthy.
	for i := range 2 {
		vm, getErr := pool.Get(ctx)
		if getErr != nil {
			t.Fatalf("Get %d after replacement failed: %v", i, getErr)
		}
		if !pool.validateVM(vm) {
			t.Errorf("VM %d should be healthy after replacement cycle", i)
		}
		pool.Put(vm)
	}
}
