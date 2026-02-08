package plugin

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hegner123/modulacms/internal/utility"
	lua "github.com/yuin/gopher-lua"
)

// ErrPoolExhausted is returned by Get() when no VM is available within the
// acquisition timeout (100ms). The caller should treat this as backpressure:
// Phase 1 treats it as a fatal plugin load error; Phase 2 HTTP bridge
// translates it to HTTP 503 Service Unavailable with a Retry-After header.
var ErrPoolExhausted = errors.New("VM pool exhausted")

// acquireTimeout is the maximum time Get() waits for an available VM.
// This is deliberately short to provide backpressure under load rather than
// tying up goroutines for the full execution timeout (5s default).
const acquireTimeout = 100 * time.Millisecond

// VMPool is a goroutine-safe pool of pre-initialized lua.LState instances
// for a single plugin. It uses a buffered channel for lock-free checkout/return
// and validates VM health on every Put() call.
//
// Lifecycle:
//  1. NewVMPool creates the pool and fills it with factory-produced VMs.
//  2. Get(ctx) checks out a VM, setting the caller's context for timeout enforcement.
//  3. Put(L) clears the stack, restores globals to post-init state, validates health,
//     and either returns the VM to the pool or replaces it if corrupted.
//  4. Close() drains all VMs and prevents further returns.
//
// Thread safety: the channel provides all synchronization for VM checkout/return.
// The initGlobals map is written once (by SnapshotGlobals) and read on every Put();
// the caller must ensure SnapshotGlobals is called before any concurrent Put() calls.
type VMPool struct {
	states      chan *lua.LState
	factory     func() *lua.LState
	size        int
	initPath    string
	pluginName  string
	closed      atomic.Bool
	initGlobals map[string]bool // snapshot of global names after on_init; nil until SnapshotGlobals called
}

// NewVMPool creates a pool of `size` pre-initialized Lua VMs using the provided
// factory function. The factory must produce fully sandboxed VMs (ApplySandbox,
// RegisterPluginRequire, RegisterDBAPI, RegisterLogAPI, FreezeModule all applied).
//
// The initPath is the absolute path to the plugin's init.lua, used for diagnostic
// logging. The pluginName is used in log messages when unhealthy VMs are detected.
//
// Global snapshot is NOT taken here -- that happens after on_init() in the manager,
// because on_init may define new globals that should be part of the baseline.
func NewVMPool(size int, factory func() *lua.LState, initPath string, pluginName string) *VMPool {
	pool := &VMPool{
		states:     make(chan *lua.LState, size),
		factory:    factory,
		size:       size,
		initPath:   initPath,
		pluginName: pluginName,
	}

	for range size {
		L := factory()
		pool.states <- L
	}

	return pool
}

// Get checks out a VM from the pool. It applies a short acquisition timeout
// (100ms) for backpressure. On success, the caller's context is set on the VM
// via L.SetContext(ctx) for execution timeout enforcement.
//
// The caller MUST call Put(L) when done, even if the Lua execution fails.
// The caller should also call DatabaseAPI.ResetOpCount() after Get() returns
// to reset the per-checkout operation budget.
//
// Returns ErrPoolExhausted if no VM is available within the acquisition timeout.
func (p *VMPool) Get(ctx context.Context) (*lua.LState, error) {
	// Short acquisition timeout -- fail fast under load.
	acquireCtx, cancel := context.WithTimeout(ctx, acquireTimeout)
	defer cancel()

	select {
	case L := <-p.states:
		L.SetContext(ctx) // execution timeout uses the caller's context
		return L, nil
	case <-acquireCtx.Done():
		return nil, ErrPoolExhausted
	}
}

// Put returns a VM to the pool after use. It performs these steps in order:
//  1. Clear the Lua stack (not globals -- globals persist across checkouts).
//  2. Restore global snapshot: remove any globals created after on_init.
//  3. Validate VM health: verify db and log modules are intact Go-bound functions.
//  4. If healthy and pool not closed: return to channel.
//  5. If healthy and pool closed: close the VM directly.
//  6. If unhealthy: close the VM, log a warning, create a fresh replacement via factory.
//
// Put never blocks: if the channel is full (should not happen with correct usage),
// the VM is closed and a warning is logged.
func (p *VMPool) Put(L *lua.LState) {
	// Step 1: Clear the Lua stack.
	L.SetTop(0)

	// Clear the context to avoid dangling references from the caller's context.
	// The next Get() call will set a fresh context.
	L.SetContext(nil)

	// Step 2: Restore globals to post-init state.
	p.restoreGlobalSnapshot(L)

	// Step 3: Validate VM health.
	if p.validateVM(L) {
		// Step 4/5: Healthy VM.
		if p.closed.Load() {
			L.Close()
			return
		}
		select {
		case p.states <- L:
			// Returned to pool successfully.
		default:
			// Channel full -- should not happen with correct usage.
			// Close the VM to avoid leaking it.
			utility.DefaultLogger.Warn(
				fmt.Sprintf("plugin %q: pool channel full on Put, closing VM", p.pluginName),
				nil,
			)
			L.Close()
		}
		return
	}

	// Step 6: Unhealthy VM -- close and replace.
	L.Close()
	utility.DefaultLogger.Warn(
		fmt.Sprintf("plugin %q: unhealthy VM detected on Put, replacing with fresh VM", p.pluginName),
		nil,
	)

	if p.closed.Load() {
		// Pool is closing, do not create replacement.
		return
	}

	replacement := p.factory()
	select {
	case p.states <- replacement:
		// Replacement added to pool.
	default:
		// Channel full -- should not happen.
		replacement.Close()
	}
}

// Close drains all VMs from the pool and prevents further returns.
// VMs that are currently checked out will be closed when they are returned
// via Put() (which checks the closed flag).
//
// Close logs a diagnostic if the pool was not fully drained (indicates VMs
// still checked out at shutdown time -- they will be closed when returned).
func (p *VMPool) Close() {
	p.closed.Store(true)

	drained := 0
	for {
		select {
		case L := <-p.states:
			L.Close()
			drained++
		default:
			// No more VMs in the channel.
			if drained < p.size {
				utility.DefaultLogger.Warn(
					fmt.Sprintf("plugin %q: pool closed with %d/%d VMs drained (remaining are checked out)",
						p.pluginName, drained, p.size),
					nil,
				)
			}
			return
		}
	}
}

// SnapshotGlobals records all current global names on the given VM.
// This must be called once after on_init() has run on the first VM.
// The snapshot is shared across all VMs in the pool (they all loaded the same
// init.lua, so their baseline global sets are identical).
//
// After this call, Put() will remove any global not in the snapshot.
// This catches the common case of plugin code forgetting `local` without
// requiring expensive deep-clone state restoration.
func (p *VMPool) SnapshotGlobals(L *lua.LState) {
	p.initGlobals = make(map[string]bool)
	globals := L.Get(lua.GlobalsIndex).(*lua.LTable)
	globals.ForEach(func(key, _ lua.LValue) {
		if s, ok := key.(lua.LString); ok {
			p.initGlobals[string(s)] = true
		}
	})
}

// restoreGlobalSnapshot removes any globals that were created after the
// snapshot was taken (i.e., not present in initGlobals). If no snapshot
// has been taken yet (initGlobals is nil), this is a no-op.
//
// Limitation: this removes NEW globals but does not restore MUTATED globals.
// Full deep-clone snapshotting is too expensive. The snapshot catches the
// common case (forgetting `local`); the validateVM health check catches the
// critical case (corrupted db/log modules).
func (p *VMPool) restoreGlobalSnapshot(L *lua.LState) {
	if p.initGlobals == nil {
		return
	}

	var toRemove []string
	globals := L.Get(lua.GlobalsIndex).(*lua.LTable)
	globals.ForEach(func(key, _ lua.LValue) {
		if s, ok := key.(lua.LString); ok {
			if !p.initGlobals[string(s)] {
				toRemove = append(toRemove, string(s))
			}
		}
	})
	for _, name := range toRemove {
		L.SetGlobal(name, lua.LNil)
	}
}

// validateVM checks that the db and log modules are intact Go-bound functions.
// This is the last line of defense: even if a metatable attack bypasses the
// proxy freezing, the VM gets discarded on return.
//
// Checks performed:
//   - db global is a *lua.LTable
//   - db.query, db.query_one, db.count, db.exists, db.insert, db.update,
//     db.delete, db.transaction, db.define_table are all *lua.LFunction with IsG==true
//   - log global is a *lua.LTable
//   - log.info is a *lua.LFunction with IsG==true
//
// Design decisions:
//   - Check IsG==true (not pointer equality): IsG is true only for Go-bound functions.
//     Catches replacement with pure Lua functions or non-function values without
//     storing original pointers.
//   - Lightweight: ~11 GetGlobal/GetField calls + type assertions. No Lua execution,
//     no allocations. Nanosecond-scale overhead per Put().
func (p *VMPool) validateVM(L *lua.LState) bool {
	// Validate db module.
	dbVal := L.GetGlobal("db")
	dbTable, ok := dbVal.(*lua.LTable)
	if !ok {
		return false
	}

	dbFuncs := []string{
		"query", "query_one", "count", "exists",
		"insert", "update", "delete", "transaction", "define_table",
	}
	for _, fname := range dbFuncs {
		if !isGoBoundFunction(L.GetField(dbTable, fname)) {
			return false
		}
	}

	// Validate log module.
	logVal := L.GetGlobal("log")
	logTable, ok := logVal.(*lua.LTable)
	if !ok {
		return false
	}

	if !isGoBoundFunction(L.GetField(logTable, "info")) {
		return false
	}

	return true
}

// isGoBoundFunction checks that a Lua value is an *lua.LFunction with IsG==true,
// meaning it is a Go-bound function (not a pure Lua function).
func isGoBoundFunction(v lua.LValue) bool {
	fn, ok := v.(*lua.LFunction)
	if !ok {
		return false
	}
	return fn.IsG
}
