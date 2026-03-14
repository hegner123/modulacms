package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/plugin"
)

func init() {
	RegisterFieldInput(FieldInputEntry{
		Key:         "plugin",
		Label:       "Plugin",
		Description: "Plugin-provided field editor",
		NewBubble:   func() FieldBubble { return NewPluginFieldBubble() },
	})
}

// PluginFieldBubble implements FieldBubble for plugin field types.
// Created unconfigured by the type registry. Must be configured via Configure()
// before the content form dialog processes any messages.
//
// In inline mode, the coroutine yields single primitives rendered within the
// field's row. In overlay mode, the bubble shows a read-only display and
// opens a PluginFieldOverlay on enter.
type PluginFieldBubble struct {
	bridge    *plugin.CoroutineBridge
	value     string           // current committed value
	primitive plugin.PluginPrimitive // last yielded primitive for View()
	width     int
	focused   bool
	mode      string // "inline" or "overlay"
	errMsg    string // non-empty if configuration failed
	configured bool
	pluginName string
	ifaceName  string
}

// NewPluginFieldBubble creates an unconfigured plugin field bubble.
// Configuration happens post-construction via Configure().
func NewPluginFieldBubble() *PluginFieldBubble {
	return &PluginFieldBubble{
		mode: "inline",
	}
}

// Configure initializes the bubble with a plugin coroutine bridge.
// Called by ConfigurePluginFields after the content form dialog is created.
func (b *PluginFieldBubble) Configure(mgr *plugin.Manager, pluginName, ifaceName string) {
	b.pluginName = pluginName
	b.ifaceName = ifaceName

	if mgr == nil {
		b.errMsg = "plugin manager not available"
		return
	}

	inst := mgr.GetPlugin(pluginName)
	if inst == nil {
		b.errMsg = fmt.Sprintf("plugin %q not found", pluginName)
		return
	}

	if inst.State != plugin.StateRunning {
		b.errMsg = fmt.Sprintf("plugin %q is %s", pluginName, inst.State)
		return
	}

	// Find the interface definition.
	iface := mgr.PluginInterface(pluginName, ifaceName)
	if iface == nil {
		b.errMsg = fmt.Sprintf("interface %q not found in plugin %q", ifaceName, pluginName)
		return
	}
	b.mode = iface.Mode
	b.configured = true
}

// SetError sets an error message on the bubble (used for missing config).
func (b *PluginFieldBubble) SetError(err error) {
	b.errMsg = err.Error()
}

func (b *PluginFieldBubble) Update(msg tea.Msg) (FieldBubble, tea.Cmd) {
	if b.errMsg != "" || !b.configured {
		return b, nil
	}

	// In overlay mode, the bubble doesn't process keys — the overlay does.
	// Enter opens the overlay.
	if b.mode == "overlay" {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			if keyMsg.String() == "enter" {
				// Open overlay — this would be handled by the content form dialog.
				// For now, return nil (overlay integration deferred).
				return b, nil
			}
		}
		return b, nil
	}

	// Inline mode: forward keys to the bridge if it's running.
	if b.bridge != nil && !b.bridge.Done() {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			key := keyMsg.String()
			event := plugin.BuildKeyEvent(b.bridge.ParentL(), key)
			yv, err := b.bridge.Resume(event)
			if err != nil {
				b.errMsg = err.Error()
				return b, nil
			}
			return b.processYield(yv)
		}
	}

	return b, nil
}

func (b *PluginFieldBubble) processYield(yv plugin.YieldValue) (FieldBubble, tea.Cmd) {
	if yv.IsAction {
		switch yv.Action.Name {
		case "commit":
			if val, ok := yv.Action.Params["value"].(string); ok {
				b.value = val
			}
		case "cancel":
			// Value unchanged.
		}
		return b, nil
	}

	if yv.Primitive != nil {
		b.primitive = yv.Primitive
	}
	return b, nil
}

func (b *PluginFieldBubble) View() string {
	if b.errMsg != "" {
		return fmt.Sprintf(" [plugin error: %s]", b.errMsg)
	}
	if !b.configured {
		return " [plugin field: unconfigured]"
	}

	if b.mode == "overlay" {
		display := b.value
		if display == "" {
			display = "(not set)"
		}
		hint := ""
		if b.focused {
			hint = " [enter to edit]"
		}
		return " " + display + hint
	}

	// Inline mode: render the primitive.
	if b.primitive != nil {
		return plugin.RenderPrimitive(b.primitive, b.width, 3, b.focused)
	}

	display := b.value
	if display == "" {
		display = "(empty)"
	}
	return " " + display
}

func (b *PluginFieldBubble) Value() string    { return b.value }
func (b *PluginFieldBubble) SetValue(v string) { b.value = v }
func (b *PluginFieldBubble) SetWidth(w int)    { b.width = w }

func (b *PluginFieldBubble) Focus() tea.Cmd {
	b.focused = true
	return nil
}

func (b *PluginFieldBubble) Blur()         { b.focused = false }
func (b *PluginFieldBubble) Focused() bool { return b.focused }
