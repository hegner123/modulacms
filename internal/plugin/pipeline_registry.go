package plugin

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"
)

// PipelineEntry represents a single pipeline step loaded from the DB.
type PipelineEntry struct {
	PipelineID string
	PluginName string
	Handler    string
	Priority   int
	Config     map[string]any
	Enabled    bool
}

// PipelineRegistry is an in-memory cache of pipeline chains loaded from the DB.
// Thread-safe via build-then-swap (same pattern as middleware.PermissionCache).
type PipelineRegistry struct {
	mu     sync.RWMutex
	chains map[string][]PipelineEntry // "table.operation" -> sorted entries
}

// NewPipelineRegistry creates an empty PipelineRegistry.
func NewPipelineRegistry() *PipelineRegistry {
	return &PipelineRegistry{
		chains: make(map[string][]PipelineEntry),
	}
}

// PipelineRow is a minimal interface for pipeline data from the DB.
// Avoids a direct dependency on the db package (which would create an import cycle).
type PipelineRow struct {
	PipelineID string
	PluginID   string
	TableName  string
	Operation  string
	PluginName string
	Handler    string
	Priority   int
	Enabled    bool
	Config     string // JSON string
}

// Build loads pipelines from the given rows and replaces the registry contents.
// Uses build-then-swap: constructs the new map without holding the lock,
// then swaps under write lock (nanosecond hold time).
func (r *PipelineRegistry) Build(rows []PipelineRow) {
	newChains := buildChains(rows)

	r.mu.Lock()
	r.chains = newChains
	r.mu.Unlock()
}

// Reload is an alias for Build, matching the PermissionCache naming convention.
func (r *PipelineRegistry) Reload(rows []PipelineRow) {
	r.Build(rows)
}

// Before returns pipeline entries matching "before_<op>" for the given table.
// Returns nil if no entries match. The caller must not modify the returned slice.
func (r *PipelineRegistry) Before(table, op string) []PipelineEntry {
	key := table + ".before_" + op
	r.mu.RLock()
	chain := r.chains[key]
	r.mu.RUnlock()
	return chain
}

// After returns pipeline entries matching "after_<op>" for the given table.
// Returns nil if no entries match. The caller must not modify the returned slice.
func (r *PipelineRegistry) After(table, op string) []PipelineEntry {
	key := table + ".after_" + op
	r.mu.RLock()
	chain := r.chains[key]
	r.mu.RUnlock()
	return chain
}

// Get returns all pipeline entries for the given table and operation key.
// The operation key is the full value (e.g., "before_create", "after_update").
func (r *PipelineRegistry) Get(table, operation string) []PipelineEntry {
	key := table + "." + operation
	r.mu.RLock()
	chain := r.chains[key]
	r.mu.RUnlock()
	return chain
}

// IsEmpty returns true if the registry has no pipeline entries.
func (r *PipelineRegistry) IsEmpty() bool {
	r.mu.RLock()
	empty := len(r.chains) == 0
	r.mu.RUnlock()
	return empty
}

// DryRunResult represents the pipeline chain for a specific table+operation+phase.
type DryRunResult struct {
	Table     string          `json:"table"`
	Operation string          `json:"operation"`
	Phase     string          `json:"phase"` // "before" or "after"
	Entries   []PipelineEntry `json:"entries"`
}

// ListKeys returns all registered chain keys in sorted order.
func (r *PipelineRegistry) ListKeys() []string {
	r.mu.RLock()
	keys := make([]string, 0, len(r.chains))
	for k := range r.chains {
		keys = append(keys, k)
	}
	r.mu.RUnlock()
	sort.Strings(keys)
	return keys
}

// DryRun returns the before and after pipeline chains for a specific table and operation.
// The op parameter is the base operation ("create", "update", "delete").
func (r *PipelineRegistry) DryRun(table, op string) []DryRunResult {
	var results []DryRunResult

	beforeEntries := r.Before(table, op)
	if len(beforeEntries) > 0 {
		results = append(results, DryRunResult{
			Table:     table,
			Operation: op,
			Phase:     "before",
			Entries:   beforeEntries,
		})
	}

	afterEntries := r.After(table, op)
	if len(afterEntries) > 0 {
		results = append(results, DryRunResult{
			Table:     table,
			Operation: op,
			Phase:     "after",
			Entries:   afterEntries,
		})
	}

	return results
}

// DryRunAll returns all pipeline chains from the registry, parsed into DryRunResults.
// Uses strings.LastIndex for key parsing (handles table names containing dots).
func (r *PipelineRegistry) DryRunAll() []DryRunResult {
	r.mu.RLock()
	chainsCopy := make(map[string][]PipelineEntry, len(r.chains))
	for k, v := range r.chains {
		chainsCopy[k] = v
	}
	r.mu.RUnlock()

	// Sort keys for deterministic output.
	keys := make([]string, 0, len(chainsCopy))
	for k := range chainsCopy {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var results []DryRunResult
	for _, key := range keys {
		dotIdx := strings.LastIndex(key, ".")
		if dotIdx < 0 || dotIdx >= len(key)-1 {
			continue // malformed key
		}

		table := key[:dotIdx]
		operation := key[dotIdx+1:]

		// Validate phase prefix.
		var phase, baseOp string
		if strings.HasPrefix(operation, "before_") {
			phase = "before"
			baseOp = operation[len("before_"):]
		} else if strings.HasPrefix(operation, "after_") {
			phase = "after"
			baseOp = operation[len("after_"):]
		} else {
			continue // not a before/after chain, skip
		}

		results = append(results, DryRunResult{
			Table:     table,
			Operation: baseOp,
			Phase:     phase,
			Entries:   chainsCopy[key],
		})
	}

	return results
}

// buildChains constructs the chain map from pipeline rows.
// Only includes enabled entries. Sorts by priority ascending, then plugin name
// for deterministic ordering when priorities are equal.
func buildChains(rows []PipelineRow) map[string][]PipelineEntry {
	chains := make(map[string][]PipelineEntry)

	for _, row := range rows {
		if !row.Enabled {
			continue
		}

		key := row.TableName + "." + row.Operation

		entry := PipelineEntry{
			PipelineID: row.PipelineID,
			PluginName: row.PluginName,
			Handler:    row.Handler,
			Priority:   row.Priority,
			Enabled:    row.Enabled,
			Config:     parseConfigJSON(row.Config),
		}

		chains[key] = append(chains[key], entry)
	}

	// Sort each chain by priority ascending, then plugin name for stability.
	for key := range chains {
		sort.Slice(chains[key], func(i, j int) bool {
			a, b := chains[key][i], chains[key][j]
			if a.Priority != b.Priority {
				return a.Priority < b.Priority
			}
			return strings.Compare(a.PluginName, b.PluginName) < 0
		})
	}

	return chains
}

// parseConfigJSON parses a JSON config string into a map.
// Returns nil on empty input or any parse error.
func parseConfigJSON(raw string) map[string]any {
	if raw == "" || raw == "{}" {
		return nil
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil
	}
	return result
}
