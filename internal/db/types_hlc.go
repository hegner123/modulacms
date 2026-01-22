// internal/db/types_hlc.go
package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// HLC represents a Hybrid Logical Clock timestamp
// Format: (wall_time_ms << 16) | logical_counter
// - Upper 48 bits: milliseconds since Unix epoch
// - Lower 16 bits: logical counter (for ordering within same millisecond)
type HLC int64

var (
	hlcMu      sync.Mutex
	hlcLast    HLC
	hlcCounter uint16
)

// HLCNow returns a new HLC timestamp greater than any previously issued
func HLCNow() HLC {
	hlcMu.Lock()
	defer hlcMu.Unlock()

	wallMs := time.Now().UnixMilli()
	physical := HLC(wallMs << 16)

	if physical > hlcLast {
		hlcLast = physical
		hlcCounter = 0
	} else {
		hlcCounter++
		if hlcCounter == 0 {
			// Counter overflow, wait for wall clock to advance
			time.Sleep(time.Millisecond)
			hlcMu.Unlock()
			result := HLCNow()
			hlcMu.Lock()
			return result
		}
	}

	hlcLast = physical | HLC(hlcCounter)
	return hlcLast
}

// HLCUpdate merges a received HLC with local time (for receiving events from other nodes)
func HLCUpdate(received HLC) HLC {
	hlcMu.Lock()
	defer hlcMu.Unlock()

	wallMs := time.Now().UnixMilli()
	physical := HLC(wallMs << 16)

	maxPhysical := physical
	if received > maxPhysical {
		maxPhysical = received & ^HLC(0xFFFF) // Extract physical part
	}
	if hlcLast > maxPhysical {
		maxPhysical = hlcLast & ^HLC(0xFFFF)
	}

	if maxPhysical == hlcLast&^HLC(0xFFFF) {
		hlcCounter++
	} else {
		hlcCounter = 0
	}

	hlcLast = maxPhysical | HLC(hlcCounter)
	return hlcLast
}

// Physical extracts the wall time from the HLC
func (h HLC) Physical() time.Time {
	ms := int64(h >> 16)
	return time.UnixMilli(ms)
}

// Logical extracts the counter from the HLC
func (h HLC) Logical() uint16 {
	return uint16(h & 0xFFFF)
}

// String formats the HLC as "HLC(wallms:counter)"
func (h HLC) String() string {
	return fmt.Sprintf("HLC(%d:%d)", h>>16, h&0xFFFF)
}

// Value implements driver.Valuer for database storage
func (h HLC) Value() (driver.Value, error) {
	return int64(h), nil
}

// Scan implements sql.Scanner for database retrieval
func (h *HLC) Scan(value any) error {
	if value == nil {
		*h = 0
		return nil
	}
	switch v := value.(type) {
	case int64:
		*h = HLC(v)
	case int:
		*h = HLC(v)
	default:
		return fmt.Errorf("HLC: cannot scan %T", value)
	}
	return nil
}

// MarshalJSON marshals the HLC as an int64
func (h HLC) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(h))
}

// UnmarshalJSON unmarshals the HLC from an int64
func (h *HLC) UnmarshalJSON(data []byte) error {
	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("HLC: %w", err)
	}
	*h = HLC(v)
	return nil
}

// Before returns true if h happened before other
func (h HLC) Before(other HLC) bool {
	return h < other
}

// After returns true if h happened after other
func (h HLC) After(other HLC) bool {
	return h > other
}
