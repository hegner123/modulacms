//go:build !debug

package tui

// Stringify returns an empty string in non-debug builds.
// Build with -tags debug for full model debug output.
func (m Model) Stringify() string { return "" }
