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
// for a single plugin. It uses buffered channels for lock-free checkout/return
// and validates VM health on every Put() call.
//
// Phase 3 splits the pool into two channels: general (for HTTP requests) and
// reserved (for hooks). GetForHook tries general first (non-blocking), then
// falls back to reserved with the same acquisition timeout as Get().
//
// Phase 4 tri-state lifecycle: open -> draining -> closed.
//
//   - Open: normal operation (Get/Put work normally)
//   - Draining: Get/GetForHook return ErrPoolExhausted; Put still returns VMs to channels
//     so the Drain() method can detect when all checked-out VMs are returned.
//   - Closed: both Get and Put close VMs directly; channels are drained.
//
// Lifecycle:
//  1. NewVMPool creates the pool and fills both channels with factory-produced VMs.
//  2. Get(ctx) checks out a VM from general, setting the caller's context.
//  3. GetForHook(ctx) tries general (non-blocking), then reserved (with timeout).
//  4. Put(L) clears the stack, restores globals, validates health,
//     and returns the VM to whichever channel has capacity (reserved first).
//  5. Drain(timeout) sets draining, waits for checked-out VMs, then closes.
//  6. Close() drains all VMs and prevents further returns.
//
// Thread safety: the channels provide all synchronization for VM checkout/return.
// The initGlobals map is written once (by SnapshotGlobals) and read on every Put();
// the caller must ensure SnapshotGlobals is called before any concurrent Put() calls.
type VMPool struct {
	general     chan *lua.LState // HTTP requests draw from here
	reserved    chan *lua.LState // hooks try general first, then reserved
	factory     func() *lua.LState
	size        int // total VMs (general + reserved)
	reserveSize int // number of VMs in the reserved channel
	initPath    string
	pluginName  string
	closed      atomic.Bool
	draining    atomic.Bool    // Phase 4: true during Drain() -- rejects Get but accepts Put returns
	checkedOut  atomic.Int32   // Phase 4 A2: tracks VMs currently checked out (incremented in Get, decremented in Put)
	initGlobals map[string]bool // snapshot of global names after on_init; nil until SnapshotGlobals called

	// onReplace is called when Put() detects an unhealthy VM and needs to replace it.
	// The callback receives the old (unhealthy) *lua.LState BEFORE it is closed,
	// allowing the caller to clean up associated state (e.g., delete from inst.dbAPIs).
	// Set during construction; nil means no callback. (S5)
	onReplace func(*lua.LState)
}

// VMPoolConfig configures the VM pool at creation time.
type VMPoolConfig struct {
	Size        int                   // total VMs (general + reserved)
	ReserveSize int                   // VMs reserved for hooks (default 1)
	Factory     func() *lua.LState
	InitPath    string
	PluginName  string
	OnReplace   func(*lua.LState) // optional cleanup callback for unhealthy VMs (S5)
}

// NewVMPool creates a pool of `size` pre-initialized Lua VMs using the provided
// factory function. The factory must produce fully sandboxed VMs (ApplySandbox,
// RegisterPluginRequire, RegisterDBAPI, RegisterLogAPI, RegisterHTTPAPI,
// RegisterHooksAPI, and all FreezeModule calls applied).
//
// reserveSize VMs are placed in the reserved channel (for hooks); the remainder
// go to the general channel (for HTTP requests). If reserveSize >= size, all
// VMs go to reserved (general is empty -- Get() will always fail, GetForHook
// is the only way to check out).
//
// The initPath is the absolute path to the plugin's init.lua, used for diagnostic
// logging. The pluginName is used in log messages when unhealthy VMs are detected.
//
// Global snapshot is NOT taken here -- that happens after on_init() in the manager,
// because on_init may define new globals that should be part of the baseline.
func NewVMPool(cfg VMPoolConfig) *VMPool {
	if cfg.Size <= 0 {
		cfg.Size = 4
	}
	if cfg.ReserveSize < 0 {
		cfg.ReserveSize = 0
	}
	if cfg.ReserveSize > cfg.Size {
		cfg.ReserveSize = cfg.Size
	}

	generalSize := cfg.Size - cfg.ReserveSize

	pool := &VMPool{
		general:     make(chan *lua.LState, generalSize),
		reserved:    make(chan *lua.LState, cfg.ReserveSize),
		factory:     cfg.Factory,
		size:        cfg.Size,
		reserveSize: cfg.ReserveSize,
		initPath:    cfg.InitPath,
		pluginName:  cfg.PluginName,
		onReplace:   cfg.OnReplace,
	}

	// Fill general channel first, then reserved.
	for range generalSize {
		L := cfg.Factory()
		pool.general <- L
	}
	for range cfg.ReserveSize {
		L := cfg.Factory()
		pool.reserved <- L
	}

	return pool
}

// Get checks out a VM from the general pool. It applies a short acquisition
// timeout (100ms) for backpressure. On success, the caller's context is set
// on the VM via L.SetContext(ctx) for execution timeout enforcement.
//
// The caller MUST call Put(L) when done, even if the Lua execution fails.
// The caller should also call DatabaseAPI.ResetOpCount() after Get() returns
// to reset the per-checkout operation budget.
//
// Returns ErrPoolExhausted if no VM is available within the acquisition timeout.
func (p *VMPool) Get(ctx context.Context) (*lua.LState, error) {
	// Phase 4: reject checkouts when draining (new requests go to the new pool).
	if p.draining.Load() {
		return nil, ErrPoolExhausted
	}

	// Short acquisition timeout -- fail fast under load.
	acquireCtx, cancel := context.WithTimeout(ctx, acquireTimeout)
	defer cancel()

	select {
	case L := <-p.general:
		p.checkedOut.Add(1) // A2: track checkout
		L.SetContext(ctx)   // execution timeout uses the caller's context
		return L, nil
	case <-acquireCtx.Done():
		return nil, ErrPoolExhausted
	}
}

// GetForHook checks out a VM for hook execution. It tries the general pool
// first (non-blocking) to avoid consuming reserved VMs when general capacity
// is available, then falls back to the reserved pool with the same acquisition
// timeout as Get() (100ms default).
//
// The caller MUST call Put(L) when done.
//
// Returns ErrPoolExhausted if no VM is available in either channel within timeout.
func (p *VMPool) GetForHook(ctx context.Context) (*lua.LState, error) {
	// Phase 4: reject checkouts when draining.
	if p.draining.Load() {
		return nil, ErrPoolExhausted
	}

	// Try general pool first (non-blocking).
	select {
	case L := <-p.general:
		p.checkedOut.Add(1) // A2: track checkout
		L.SetContext(ctx)
		return L, nil
	default:
	}

	// Fall back to reserved with acquisition timeout.
	acquireCtx, cancel := context.WithTimeout(ctx, acquireTimeout)
	defer cancel()

	select {
	case L := <-p.reserved:
		p.checkedOut.Add(1) // A2: track checkout
		L.SetContext(ctx)
		return L, nil
	case <-acquireCtx.Done():
		return nil, ErrPoolExhausted
	}
}

// Put returns a VM to the pool after use. It performs these steps in order:
//  1. Clear the Lua stack (not globals -- globals persist across checkouts).
//  2. Restore global snapshot: remove any globals created after on_init.
//  3. Validate VM health: verify db, log, http, and hooks modules are intact.
//  4. If healthy: return to whichever channel has capacity (reserved first,
//     then general). If pool closed: close the VM directly.
//  5. If unhealthy: invoke onReplace callback, close the VM, create a
//     fresh replacement via factory, and add to the pool.
//
// Put never blocks: if both channels are full (should not happen with correct
// usage), the VM is closed and a warning is logged.
func (p *VMPool) Put(L *lua.LState) {
	// A2: Decrement checkout counter before any other operations.
	// This must happen first so that Drain() sees the return promptly.
	p.checkedOut.Add(-1)

	// Step 1: Clear the Lua stack.
	L.SetTop(0)

	// Clear the context to avoid dangling references from the caller's context.
	// The next Get() call will set a fresh context.
	L.SetContext(nil)

	// Step 2: Restore globals to post-init state.
	p.restoreGlobalSnapshot(L)

	// Step 3: Validate VM health.
	if p.validateVM(L) {
		// Step 4: Healthy VM.
		if p.closed.Load() {
			L.Close()
			return
		}
		p.returnToPool(L)
		return
	}

	// Step 5: Unhealthy VM -- invoke callback, close, and replace.
	if p.onReplace != nil {
		p.onReplace(L)
	}
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
	p.returnToPool(replacement)
}

// returnToPool sends a VM to whichever channel has capacity. Prefers reserved
// first (to keep the hook reserve full), then general.
func (p *VMPool) returnToPool(L *lua.LState) {
	// Try reserved first.
	select {
	case p.reserved <- L:
		return
	default:
	}

	// Then general.
	select {
	case p.general <- L:
		return
	default:
	}

	// Both full -- should not happen with correct usage.
	utility.DefaultLogger.Warn(
		fmt.Sprintf("plugin %q: pool channels full on Put, closing VM", p.pluginName),
		nil,
	)
	L.Close()
}

// Close drains all VMs from both channels and prevents further returns.
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
		case L := <-p.general:
			L.Close()
			drained++
		case L := <-p.reserved:
			L.Close()
			drained++
		default:
			// No more VMs in either channel.
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

// Drain initiates a graceful drain of the pool. It sets the draining flag
// (so Get/GetForHook return ErrPoolExhausted for new requests), then waits
// for all checked-out VMs to be returned via Put(). Once all VMs are back
// (or the timeout expires), it closes all VMs in both channels.
//
// Returns true if all VMs were returned within the timeout, false otherwise.
//
// A2: Uses atomic Int32 checkout counter instead of len(general)+len(reserved)
// to avoid data races between the two channel length reads.
//
// S6: If Drain returns false (timeout), the caller (watcher.reloadPlugin)
// must trip the plugin-level circuit breaker to prevent a reload-drain-timeout
// cycle from masking stuck handlers.
func (p *VMPool) Drain(timeout time.Duration) bool {
	p.draining.Store(true)

	deadline := time.After(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			// Timeout: some VMs still checked out. Close what we can.
			n := p.checkedOut.Load()
			utility.DefaultLogger.Warn(
				fmt.Sprintf("plugin %q: drain timeout with %d VMs still checked out",
					p.pluginName, n),
				nil,
			)
			p.closed.Store(true)
			p.drainChannels()
			return false
		case <-ticker.C:
			n := p.checkedOut.Load()
			if n < 0 {
				// Counter went negative -- indicates a bug (Put without matching Get).
				// Log warning but treat as drained to avoid infinite loop.
				utility.DefaultLogger.Warn(
					fmt.Sprintf("plugin %q: checkout counter is negative (%d), possible double-Put",
						p.pluginName, n),
					nil,
				)
			}
			if n <= 0 {
				// All VMs returned. Close them.
				p.closed.Store(true)
				p.drainChannels()
				return true
			}
		}
	}
}

// drainChannels empties both channel buffers and closes all VMs in them.
func (p *VMPool) drainChannels() {
	for {
		select {
		case L := <-p.general:
			L.Close()
		case L := <-p.reserved:
			L.Close()
		default:
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

// validateVM checks that the db, log, http, and hooks modules are intact
// Go-bound functions. This is the last line of defense: even if a metatable
// attack bypasses the proxy freezing, the VM gets discarded on return.
//
// Checks performed:
//   - db global is a *lua.LTable
//   - db.query, db.query_one, db.count, db.exists, db.insert, db.update,
//     db.delete, db.transaction, db.define_table are all *lua.LFunction with IsG==true
//   - log global is a *lua.LTable
//   - log.info is a *lua.LFunction with IsG==true
//   - http global is a *lua.LTable
//   - http.handle is a *lua.LFunction with IsG==true
//   - hooks global is a *lua.LTable (Phase 3)
//   - hooks.on is a *lua.LFunction with IsG==true (Phase 3)
//
// Design decisions:
//   - Check IsG==true (not pointer equality): IsG is true only for Go-bound functions.
//   - Lightweight: ~15 GetGlobal/GetField calls + type assertions. No Lua execution,
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

	// Validate http module (Phase 2).
	httpVal := L.GetGlobal("http")
	httpTable, ok := httpVal.(*lua.LTable)
	if !ok {
		return false
	}

	if !isGoBoundFunction(L.GetField(httpTable, "handle")) {
		return false
	}

	// Validate hooks module (Phase 3).
	hooksVal := L.GetGlobal("hooks")
	hooksTable, ok := hooksVal.(*lua.LTable)
	if !ok {
		return false
	}

	if !isGoBoundFunction(L.GetField(hooksTable, "on")) {
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

// PoolSize returns the total number of VMs (general + reserved) in the pool.
// Used for diagnostics and testing.
func (p *VMPool) PoolSize() int {
	return p.size
}

// AvailableCount returns the number of VMs currently available across both channels.
// Used for diagnostics and testing. Not safe for capacity decisions (race conditions).
func (p *VMPool) AvailableCount() int {
	return len(p.general) + len(p.reserved)
}
