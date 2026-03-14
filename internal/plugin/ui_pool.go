package plugin

import (
	"context"
	"sync/atomic"

	"github.com/hegner123/modulacms/internal/utility"
	lua "github.com/yuin/gopher-lua"
)

// UIVMPool is a separate VM pool for long-held UI coroutines.
// Unlike the standard VMPool (100ms acquire timeout), the UI pool has
// no acquisition timeout — VMs are checked out when screens/fields open
// and held indefinitely until the UI closes.
//
// UI pool is bounded by MaxUIVMs (default 4) to cap concurrent UI sessions
// per plugin. Pool exhaustion shows "Plugin busy" instead of blocking.
type UIVMPool struct {
	vms        chan *lua.LState
	factory    func() *lua.LState
	size       int
	pluginName string
	closed     atomic.Bool
	draining   atomic.Bool
	checkedOut atomic.Int32
}

// UIVMPoolConfig configures the UI VM pool at creation time.
type UIVMPoolConfig struct {
	Size       int // max concurrent UI VMs (default 4)
	Factory    func() *lua.LState
	PluginName string
}

// NewUIVMPool creates a pool of pre-initialized VMs for plugin UI coroutines.
func NewUIVMPool(cfg UIVMPoolConfig) *UIVMPool {
	if cfg.Size <= 0 {
		cfg.Size = 4
	}

	pool := &UIVMPool{
		vms:        make(chan *lua.LState, cfg.Size),
		factory:    cfg.Factory,
		size:       cfg.Size,
		pluginName: cfg.PluginName,
	}

	for range cfg.Size {
		pool.vms <- cfg.Factory()
	}

	return pool
}

// Get checks out a UI VM. Unlike the standard pool, this does NOT have
// an acquisition timeout — if the pool is full, it returns ErrPoolExhausted
// immediately (non-blocking). The caller should show "Plugin busy".
func (p *UIVMPool) Get(ctx context.Context) (*lua.LState, error) {
	if p.draining.Load() || p.closed.Load() {
		return nil, ErrPoolExhausted
	}

	select {
	case L := <-p.vms:
		p.checkedOut.Add(1)
		L.SetContext(ctx)
		return L, nil
	default:
		return nil, ErrPoolExhausted
	}
}

// Put returns a UI VM to the pool.
func (p *UIVMPool) Put(L *lua.LState) {
	p.checkedOut.Add(-1)
	L.SetTop(0)
	L.SetContext(nil)

	if p.closed.Load() {
		L.Close()
		return
	}

	select {
	case p.vms <- L:
	default:
		// Pool is full (should not happen with correct usage).
		utility.DefaultLogger.Warn("UI VM pool full, closing VM", nil)
		L.Close()
	}
}

// CloseVM closes a VM without returning it to the pool.
// Used when a coroutine is done and the VM should be discarded
// (e.g., after hot reload, the old VM is closed instead of returned).
func (p *UIVMPool) CloseVM(L *lua.LState) {
	p.checkedOut.Add(-1)
	L.Close()
}

// Close shuts down the pool, closing all VMs.
func (p *UIVMPool) Close() {
	if !p.closed.CompareAndSwap(false, true) {
		return
	}
	close(p.vms)
	for L := range p.vms {
		L.Close()
	}
}

// Drain rejects new checkouts and waits for checked-out VMs to return.
func (p *UIVMPool) Drain() {
	p.draining.Store(true)
	// Active UI coroutines continue on old VMs until they finish.
	// The caller closes old VMs via CloseVM when the coroutine exits.
}

// ScreenDefinition describes a plugin screen from the manifest.
type ScreenDefinition struct {
	Name   string
	Label  string
	Icon   string
	Hidden bool
}

// InterfaceDefinition describes a plugin field interface from the manifest.
type InterfaceDefinition struct {
	Name  string
	Label string
	Mode  string // "inline" or "overlay"
}

// ErrPoolExhausted is reused from pool.go.
