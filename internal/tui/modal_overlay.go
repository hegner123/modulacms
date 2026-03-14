package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ModalOverlay is the interface satisfied by all dialog types
// (DialogModel, FormDialogModel, ContentFormDialogModel, etc.).
// When non-nil on Model.ActiveOverlay, it captures all key input
// and renders on top of the base UI.
type ModalOverlay interface {
	OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd)
	OverlayView(width, height int) string
}

// OverlayTicker is an optional interface that overlays can implement
// to receive non-key messages (cursor blink, timer ticks, etc.).
// Without this, text input cursors in form dialogs don't animate
// and typed text may not render until focus changes.
type OverlayTicker interface {
	OverlayTick(msg tea.Msg) (ModalOverlay, tea.Cmd)
}

// OverlaySetMsg replaces the per-type Set messages.
type OverlaySetMsg struct {
	Overlay ModalOverlay
}

// OverlayClearMsg replaces the per-type ActiveSet(false) messages.
type OverlayClearMsg struct{}

// OverlaySetCmd creates a command to set the active overlay.
func OverlaySetCmd(overlay ModalOverlay) tea.Cmd {
	return func() tea.Msg { return OverlaySetMsg{Overlay: overlay} }
}

// OverlayClearCmd creates a command to clear the active overlay.
func OverlayClearCmd() tea.Cmd {
	return func() tea.Msg { return OverlayClearMsg{} }
}

// RenderOverlay positions an overlay centered over base content.
func RenderOverlay(base string, overlay ModalOverlay, width, height int) string {
	content := overlay.OverlayView(width, height)
	w := lipgloss.Width(content)
	h := lipgloss.Height(content)
	x := (width - w) / 2
	y := (height - h) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	return Composite(base, Overlay{
		Content: content,
		X:       x,
		Y:       y,
		Width:   w,
		Height:  h,
	})
}
