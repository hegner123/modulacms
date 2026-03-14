package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/plugin"
)

// PluginFieldOverlay implements ModalOverlay for overlay-mode field interfaces.
// It wraps a CoroutineBridge that renders full grid layouts in a modal overlay.
// The coroutine yields grid layouts; commit/cancel actions close the overlay.
type PluginFieldOverlay struct {
	bridge    *plugin.CoroutineBridge
	layout    *plugin.PluginLayout
	title     string
	value     string // committed value (set on commit action)
	done      bool
	committed bool
	errMsg    string
}

// NewPluginFieldOverlay creates an overlay for an overlay-mode plugin field interface.
func NewPluginFieldOverlay(bridge *plugin.CoroutineBridge, title string) *PluginFieldOverlay {
	return &PluginFieldOverlay{
		bridge: bridge,
		title:  title,
	}
}

func (o *PluginFieldOverlay) OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd) {
	if o.bridge == nil || o.bridge.Done() || o.done {
		// Overlay is finished — clear it.
		return nil, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	}

	key := msg.String()
	event := plugin.BuildKeyEvent(o.bridge.ParentL(), key)
	yv, err := o.bridge.Resume(event)
	if err != nil {
		o.errMsg = err.Error()
		o.done = true
		return nil, tea.Batch(
			OverlayClearCmd(),
			FocusSetCmd(PAGEFOCUS),
		)
	}

	return o.processYield(yv)
}

func (o *PluginFieldOverlay) processYield(yv plugin.YieldValue) (ModalOverlay, tea.Cmd) {
	if yv.IsAction {
		switch yv.Action.Name {
		case "commit":
			if val, ok := yv.Action.Params["value"].(string); ok {
				o.value = val
			}
			o.done = true
			o.committed = true
			return nil, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
			)
		case "cancel":
			o.done = true
			return nil, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
			)
		case "quit":
			o.done = true
			return nil, tea.Batch(
				OverlayClearCmd(),
				FocusSetCmd(PAGEFOCUS),
			)
		}
		return o, nil
	}

	if yv.Layout != nil {
		o.layout = yv.Layout
	}
	return o, nil
}

func (o *PluginFieldOverlay) OverlayView(width, height int) string {
	if o.errMsg != "" {
		return fmt.Sprintf("\n  Plugin Error: %s\n\n  Press any key to close.", o.errMsg)
	}

	if o.layout == nil {
		return "\n  Loading plugin interface..."
	}

	// Build cell contents from layout and render grid.
	grid := Grid{Columns: make([]GridColumn, 0, len(o.layout.Columns))}
	var cells []CellContent
	for _, col := range o.layout.Columns {
		gridCells := make([]GridCell, 0, len(col.Cells))
		for _, cell := range col.Cells {
			content := ""
			totalLines := 0
			if cell.Content != nil {
				content = plugin.RenderPrimitive(cell.Content, width, height, false)
				totalLines = strings.Count(content, "\n") + 1
			}
			cells = append(cells, CellContent{
				Content:    content,
				TotalLines: totalLines,
			})
			gridCells = append(gridCells, GridCell{
				Height: cell.Height,
				Title:  cell.Title,
			})
		}
		grid.Columns = append(grid.Columns, GridColumn{
			Span:  col.Span,
			Cells: gridCells,
		})
	}

	return grid.Render(cells, width, height, 0)
}

// Value returns the committed value (only meaningful after Done() is true).
func (o *PluginFieldOverlay) Value() string { return o.value }

// Done returns true if the overlay has been closed (commit or cancel).
func (o *PluginFieldOverlay) Done() bool { return o.done }

// Committed returns true if the overlay closed via a commit action.
func (o *PluginFieldOverlay) Committed() bool { return o.committed }
