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
// with minimal db, log, http, and hooks tables populated with Go-bound stub
// functions. This matches the shape expected by validateVM without requiring
// the full registration functions.
func newTestVMFactory() func() *lua.LState {
	return func() *lua.LState {
		L := lua.NewState(lua.Options{SkipOpenLibs: true})
		ApplySandbox(L, SandboxConfig{})

		// Register minimal db module with Go-bound stubs for all validated functions.
		dbTable := L.NewTable()
		for _, name := range dbFunctionNames {
			dbTable.RawSetString(name, L.NewFunction(func(L *lua.LState) int { return 0 }))
		}
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

		// Register minimal http module with Go-bound stubs for validated functions.
		httpTable := L.NewTable()
		httpTable.RawSetString("handle", L.NewFunction(func(L *lua.LState) int { return 0 }))
		httpTable.RawSetString("use", L.NewFunction(func(L *lua.LState) int { return 0 }))
		L.SetGlobal("http", httpTable)

		// Register minimal hooks module with Go-bound stubs (Phase 3).
		hooksTable := L.NewTable()
		hooksTable.RawSetString("on", L.NewFunction(func(L *lua.LState) int { return 0 }))
		L.SetGlobal("hooks", hooksTable)

		// Freeze modules to match production behavior.
		FreezeModule(L, "db")
		FreezeModule(L, "log")
		FreezeModule(L, "http")
		FreezeModule(L, "hooks")

		return L
	}
}

// newTestPool creates a pool with the test VM factory using VMPoolConfig.
// reserveSize defaults to 0 for backward-compatible tests.
func newTestPool(size int) *VMPool {
	return NewVMPool(VMPoolConfig{
		Size:       size,
		Factory:    newTestVMFactory(),
		InitPath:   "/test/init.lua",
		PluginName: "test_plugin",
	})
}

// newTestPoolWithReserve creates a pool with the specified reserve size.
func newTestPoolWithReserve(size, reserveSize int) *VMPool {
	return NewVMPool(VMPoolConfig{
		Size:        size,
		ReserveSize: reserveSize,
		Factory:     newTestVMFactory(),
		InitPath:    "/test/init.lua",
		PluginName:  "test_plugin",
	})
}

// -- Tests --

func TestNewVMPool_CreatesRequestedSize(t *testing.T) {
	pool := newTestPool(3)
	defer pool.Close()

	if pool.AvailableCount() != 3 {
		t.Errorf("expected pool to have 3 VMs, got %d", pool.AvailableCount())
	}
}

func TestVMPool_GetSetsContext(t *testing.T) {
	pool := newTestPool(1)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

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
	pool := newTestPool(1)
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("first Get failed: %v", err)
	}
	defer pool.Put(L)

	start := time.Now()
	_, err = pool.Get(ctx)
	elapsed := time.Since(start)

	if !errors.Is(err, ErrPoolExhausted) {
		t.Fatalf("expected ErrPoolExhausted, got %v", err)
	}

	if elapsed > 300*time.Millisecond {
		t.Errorf("Get took %v, expected ~100ms (acquisition timeout)", elapsed)
	}
	if elapsed < 50*time.Millisecond {
		t.Errorf("Get returned in %v, suspiciously fast (expected ~100ms)", elapsed)
	}
}

func TestVMPool_GetRespectsContextCancellation(t *testing.T) {
	pool := newTestPool(1)
	defer pool.Close()

	bgCtx := context.Background()

	L, err := pool.Get(bgCtx)
	if err != nil {
		t.Fatalf("first Get failed: %v", err)
	}
	defer pool.Put(L)

	cancelledCtx, cancel := context.WithCancel(bgCtx)
	cancel()

	_, err = pool.Get(cancelledCtx)
	if !errors.Is(err, ErrPoolExhausted) {
		t.Fatalf("expected ErrPoolExhausted with cancelled context, got %v", err)
	}
}

func TestVMPool_ConcurrentGetPut(t *testing.T) {
	poolSize := 4
	pool := newTestPool(poolSize)
	defer pool.Close()

	L, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("setup Get failed: %v", err)
	}
	pool.SnapshotGlobals(L)
	pool.Put(L)

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
					if errors.Is(getErr, ErrPoolExhausted) {
						continue
					}
					errCh <- getErr
					return
				}

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

	if pool.AvailableCount() != poolSize {
		t.Errorf("pool has %d VMs, expected %d", pool.AvailableCount(), poolSize)
	}
}

func TestVMPool_PutReturnsHealthyVM(t *testing.T) {
	pool := newTestPool(1)
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	pool.Put(L)

	L2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("second Get failed: %v", err)
	}
	defer pool.Put(L2)

	if execErr := L2.DoString(`return db.query()`); execErr != nil {
		t.Errorf("VM not functional after Put/Get cycle: %v", execErr)
	}
}

func TestVMPool_PutReplacesUnhealthyVM_CorruptedDB(t *testing.T) {
	pool := newTestPool(1)
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	L.SetGlobal("db", lua.LString("corrupted"))
	pool.Put(L)

	if pool.AvailableCount() != 1 {
		t.Fatalf("pool has %d VMs after unhealthy Put, expected 1", pool.AvailableCount())
	}

	L2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get after replacement failed: %v", err)
	}
	defer pool.Put(L2)

	if !pool.validateVM(L2) {
		t.Error("replacement VM failed validation")
	}
}

func TestVMPool_PutReplacesUnhealthyVM_CorruptedLog(t *testing.T) {
	pool := newTestPool(1)
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	L.SetGlobal("log", lua.LNil)
	pool.Put(L)

	if pool.AvailableCount() != 1 {
		t.Fatalf("pool has %d VMs after unhealthy Put, expected 1", pool.AvailableCount())
	}

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
	pool := newTestPool(1)
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	fakeDB := L.NewTable()
	for _, name := range dbFunctionNames {
		if luaErr := L.DoString(`return function() end`); luaErr != nil {
			t.Fatalf("failed to create Lua function: %v", luaErr)
		}
		luaFn := L.Get(-1)
		L.Pop(1)
		fakeDB.RawSetString(name, luaFn)
	}
	L.SetGlobal("db", fakeDB)

	pool.Put(L)

	if pool.AvailableCount() != 1 {
		t.Fatalf("pool has %d VMs, expected 1", pool.AvailableCount())
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
	pool := newTestPool(1)
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	pool.SnapshotGlobals(L)
	pool.Put(L)

	L, err = pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	L.SetGlobal("leaked_global", lua.LString("should be removed"))
	pool.Put(L)

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
	pool := newTestPool(1)
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	L.SetGlobal("plugin_info", lua.LString("my_plugin"))
	pool.SnapshotGlobals(L)
	pool.Put(L)

	L, err = pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	L.SetGlobal("temp", lua.LString("temporary"))
	pool.Put(L)

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
	pool := newTestPool(1)
	defer pool.Close()

	ctx := context.Background()

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
	pool := newTestPool(3)

	ctx := context.Background()
	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	pool.Close()

	if pool.AvailableCount() != 0 {
		t.Errorf("pool has %d VMs after Close, expected 0", pool.AvailableCount())
	}

	pool.Put(L)

	if pool.AvailableCount() != 0 {
		t.Errorf("pool has %d VMs after late Put, expected 0", pool.AvailableCount())
	}
}

func TestVMPool_Close_GetAfterCloseReturnsExhausted(t *testing.T) {
	pool := newTestPool(1)
	pool.Close()

	_, err := pool.Get(context.Background())
	if !errors.Is(err, ErrPoolExhausted) {
		t.Fatalf("expected ErrPoolExhausted after Close, got %v", err)
	}
}

func TestVMPool_PutClearsStack(t *testing.T) {
	pool := newTestPool(1)
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	L.Push(lua.LString("leftover1"))
	L.Push(lua.LString("leftover2"))
	L.Push(lua.LNumber(42))

	pool.Put(L)

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
	pool := newTestPool(1)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	pool.Put(L)

	// Get the VM from the general channel directly to check its state.
	rawL := <-pool.general
	gotCtx := rawL.Context()
	pool.general <- rawL

	if gotCtx != nil {
		t.Error("expected context to be nil after Put, but it was set")
	}
}

func TestValidateVM_HealthyVM(t *testing.T) {
	pool := newTestPool(1)
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
	pool := newTestPool(1)
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
	pool := newTestPool(1)
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
	for _, funcName := range dbFunctionNames {
		t.Run(funcName, func(t *testing.T) {
			pool := newTestPool(1)
			defer pool.Close()

			L, err := pool.Get(context.Background())
			if err != nil {
				t.Fatalf("Get failed: %v", err)
			}
			defer pool.Put(L)

			fakeDB := L.NewTable()
			for _, name := range dbFunctionNames {
				if name == funcName {
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
	pool := newTestPool(1)
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
	pool := newTestPool(1)
	defer pool.Close()

	L, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

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

func TestValidateVM_MissingHooks(t *testing.T) {
	pool := newTestPool(1)
	defer pool.Close()

	L, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L)

	L.SetGlobal("hooks", lua.LNil)
	if pool.validateVM(L) {
		t.Error("VM with nil hooks should fail validation")
	}
}

func TestVMPool_MultipleSnapshotRestoreCycles(t *testing.T) {
	pool := newTestPool(1)
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	pool.SnapshotGlobals(L)
	pool.Put(L)

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
	if L.GetGlobal("global_a") != lua.LNil {
		t.Error("global_a should still be nil")
	}
}

func TestVMPool_PoolSizeStaysConstantAfterReplacement(t *testing.T) {
	pool := newTestPoolWithReserve(2, 0)
	defer pool.Close()

	ctx := context.Background()

	L1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get L1 failed: %v", err)
	}
	L2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get L2 failed: %v", err)
	}

	L1.SetGlobal("db", lua.LNil)

	pool.Put(L1)
	pool.Put(L2)

	if pool.AvailableCount() != 2 {
		t.Errorf("pool has %d VMs, expected 2 (constant after replacement)", pool.AvailableCount())
	}

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

// -- Phase 3: GetForHook tests --

func TestVMPool_GetForHook_PrefersGeneral(t *testing.T) {
	// Pool with 3 general + 1 reserved = 4 total.
	pool := newTestPoolWithReserve(4, 1)
	defer pool.Close()

	ctx := context.Background()

	// GetForHook should succeed using general pool first.
	L, err := pool.GetForHook(ctx)
	if err != nil {
		t.Fatalf("GetForHook failed: %v", err)
	}
	defer pool.Put(L)

	// Reserved should still have 1 VM.
	if len(pool.reserved) != 1 {
		t.Errorf("reserved channel has %d VMs, expected 1 (GetForHook should prefer general)", len(pool.reserved))
	}
}

func TestVMPool_GetForHook_FallsBackToReserved(t *testing.T) {
	// Pool with 2 general + 1 reserved = 3 total.
	pool := newTestPoolWithReserve(3, 1)
	defer pool.Close()

	ctx := context.Background()

	// Exhaust general pool.
	var generalVMs []*lua.LState
	for range 2 {
		L, err := pool.Get(ctx)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		generalVMs = append(generalVMs, L)
	}

	// GetForHook should now fall back to reserved.
	L, err := pool.GetForHook(ctx)
	if err != nil {
		t.Fatalf("GetForHook failed when general exhausted: %v", err)
	}
	pool.Put(L)

	// Return general VMs.
	for _, vm := range generalVMs {
		pool.Put(vm)
	}
}

func TestVMPool_GetForHook_ExhaustedWhenBothEmpty(t *testing.T) {
	pool := newTestPoolWithReserve(2, 1)
	defer pool.Close()

	ctx := context.Background()

	// Check out all VMs (1 general + 1 reserved via GetForHook).
	L1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer pool.Put(L1)

	L2, err := pool.GetForHook(ctx)
	if err != nil {
		t.Fatalf("GetForHook failed: %v", err)
	}
	defer pool.Put(L2)

	// Both channels are empty now.
	_, err = pool.GetForHook(ctx)
	if !errors.Is(err, ErrPoolExhausted) {
		t.Fatalf("expected ErrPoolExhausted when both channels empty, got %v", err)
	}
}

func TestVMPool_OnReplaceCallback(t *testing.T) {
	var calledWith *lua.LState
	pool := NewVMPool(VMPoolConfig{
		Size:       1,
		Factory:    newTestVMFactory(),
		InitPath:   "/test/init.lua",
		PluginName: "test_plugin",
		OnReplace: func(old *lua.LState) {
			calledWith = old
		},
	})
	defer pool.Close()

	ctx := context.Background()

	L, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	original := L
	L.SetGlobal("db", lua.LNil) // corrupt
	pool.Put(L)

	if calledWith != original {
		t.Error("onReplace should have been called with the unhealthy VM")
	}
}
