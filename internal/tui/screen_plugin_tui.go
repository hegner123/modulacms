package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/plugin"
	lua "github.com/yuin/gopher-lua"
)

// PluginTUIScreen implements Screen for standalone plugin UI screens.
// It wraps a CoroutineBridge and renders the yielded layout via the
// existing GridScreen infrastructure.
type PluginTUIScreen struct {
	GridScreen
	bridge     *plugin.CoroutineBridge
	layout     *plugin.PluginLayout
	primitive  plugin.PluginPrimitive // for single-primitive yields
	pluginName string
	screenName string
	params     map[string]string
	hints      []plugin.PluginHint
	mgr        *plugin.Manager
	errMsg     string
	width      int
	height     int
}

// NewPluginTUIScreen creates a PluginTUIScreen. The bridge is NOT initialized
// here — it is initialized when the screen receives a PluginScreenInitMsg,
// which is sent by the navigation command after VM checkout and function loading.
func NewPluginTUIScreen(pluginName, screenName string, params map[string]string) *PluginTUIScreen {
	return &PluginTUIScreen{
		GridScreen: GridScreen{
			Grid: Grid{Columns: []GridColumn{
				{Span: 12, Cells: []GridCell{
					{Height: 1.0, Title: fmt.Sprintf("%s/%s", pluginName, screenName)},
				}},
			}},
		},
		pluginName: pluginName,
		screenName: screenName,
		params:     params,
	}
}

func (s *PluginTUIScreen) PageIndex() PageIndex { return PLUGINTUIPAGE }

func (s *PluginTUIScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	// Capture plugin manager reference from context on every update.
	if ctx.PluginManager != nil {
		s.mgr = ctx.PluginManager
	}

	switch msg := msg.(type) {
	case PluginScreenInitMsg:
		return s.handleInit(msg)

	case PluginScreenErrorMsg:
		s.errMsg = msg.Error
		return s, LoadingStopCmd()

	case PluginDataMsg:
		return s.handleDataMsg(msg)

	case PluginDialogResponseMsg:
		return s.handleDialogResponse(msg)

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		if s.bridge != nil && !s.bridge.Done() {
			return s.resumeWithEvent(plugin.BuildResizeEvent(s.bridge.ParentL(), msg.Width, msg.Height))
		}
		return s, nil

	case tea.KeyPressMsg:
		if s.bridge == nil || s.bridge.Done() {
			// Bridge not initialized or dead — handle back/quit only.
			km := ctx.Config.KeyBindings
			key := msg.String()
			_, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
			if handled {
				return s, cmd
			}
			return s, nil
		}
		return s.handleKey(ctx, msg)
	}

	return s, nil
}

func (s *PluginTUIScreen) handleInit(msg PluginScreenInitMsg) (Screen, tea.Cmd) {
	s.bridge = msg.Bridge
	s.width = msg.Width
	s.height = msg.Height

	initEvent := plugin.BuildInitEvent(msg.L, msg.Width, msg.Height, msg.Params)
	yv, err := s.bridge.Start(initEvent)
	if err != nil {
		s.errMsg = fmt.Sprintf("Plugin %q screen %q init error: %s", s.pluginName, s.screenName, err)
		return s, LoadingStopCmd()
	}

	return s.processYield(yv)
}

func (s *PluginTUIScreen) handleKey(ctx AppContext, msg tea.KeyPressMsg) (Screen, tea.Cmd) {
	km := ctx.Config.KeyBindings
	key := msg.String()

	// Focus nav handled by GridScreen.
	if s.HandleFocusNav(key, km) {
		return s, nil
	}

	// Forward key to coroutine.
	return s.resumeWithEvent(plugin.BuildKeyEvent(s.bridge.ParentL(), key))
}

func (s *PluginTUIScreen) handleDataMsg(msg PluginDataMsg) (Screen, tea.Cmd) {
	if s.bridge == nil || s.bridge.Done() {
		return s, nil
	}

	L := s.bridge.ParentL()
	var result lua.LValue
	if msg.Result != nil {
		result = msg.Result
	}
	event := plugin.BuildDataEvent(L, msg.ID, msg.OK, result, msg.Error)
	return s.resumeWithEvent(event)
}

func (s *PluginTUIScreen) handleDialogResponse(msg PluginDialogResponseMsg) (Screen, tea.Cmd) {
	if s.bridge == nil || s.bridge.Done() {
		return s, nil
	}
	event := plugin.BuildDialogEvent(s.bridge.ParentL(), msg.Accepted)
	return s.resumeWithEvent(event)
}

func (s *PluginTUIScreen) resumeWithEvent(event *lua.LTable) (Screen, tea.Cmd) {
	yv, err := s.bridge.Resume(event)
	if err != nil {
		s.errMsg = fmt.Sprintf("Plugin %q error: %s", s.pluginName, err)
		return s, nil
	}
	return s.processYield(yv)
}

func (s *PluginTUIScreen) processYield(yv plugin.YieldValue) (Screen, tea.Cmd) {
	if yv.IsAction {
		return s.handleAction(yv.Action)
	}

	if yv.Layout != nil {
		s.layout = yv.Layout
		s.primitive = nil
		s.hints = yv.Layout.Hints
		s.rebuildGrid()
		return s, nil
	}

	if yv.Primitive != nil {
		s.primitive = yv.Primitive
		s.layout = nil
		return s, nil
	}

	return s, nil
}

func (s *PluginTUIScreen) handleAction(action *plugin.PluginAction) (Screen, tea.Cmd) {
	switch action.Name {
	case "quit":
		return s, HistoryPopCmd()

	case "navigate":
		pageStr, _ := action.Params["page"].(string)
		if pageStr != "" {
			// CMS page navigation — use the standard page map.
			// The plugin can navigate to built-in CMS pages by name.
			return s, func() tea.Msg {
				return NavigateToPluginScreenMsg{
					PluginName: s.pluginName,
					ScreenName: pageStr,
				}
			}
		}
		// Plugin-to-plugin screen navigation.
		targetPlugin, _ := action.Params["plugin"].(string)
		targetScreen, _ := action.Params["screen"].(string)
		if targetPlugin == "" {
			targetPlugin = s.pluginName
		}
		var navParams map[string]string
		if paramsRaw, ok := action.Params["params"]; ok {
			if m, ok := paramsRaw.(map[string]any); ok {
				navParams = make(map[string]string, len(m))
				for k, v := range m {
					navParams[k] = fmt.Sprintf("%v", v)
				}
			}
		}
		return s, func() tea.Msg {
			return NavigateToPluginScreenMsg{
				PluginName: targetPlugin,
				ScreenName: targetScreen,
				Params:     navParams,
			}
		}

	case "confirm":
		title, _ := action.Params["title"].(string)
		message, _ := action.Params["message"].(string)
		return s, func() tea.Msg {
			return ShowPluginConfirmDialogMsg{
				Title:   title,
				Message: message,
			}
		}

	case "toast":
		message, _ := action.Params["message"].(string)
		return s, func() tea.Msg {
			return ActionResultMsg{
				Title:   "Plugin",
				Message: message,
			}
		}

	case "fetch":
		// Async database fetch. The DatabaseAPI is Lua-bound and not
		// safe to call from a goroutine. For now, fetch actions return
		// an error directing plugins to use the db module directly in
		// on_init or via a different pattern. Full async DB support
		// requires a Go-level query API on DatabaseAPI.
		id, _ := action.Params["id"].(string)
		return s, func() tea.Msg {
			return PluginDataMsg{
				ID:    id,
				OK:    false,
				Error: "async fetch requires Go-level DatabaseAPI (use db module in on_init for data loading)",
			}
		}

	case "request":
		id, _ := action.Params["id"].(string)
		return s, PluginRequestCmd(id, action, s.mgr, s.pluginName)
	}

	return s, nil
}

func (s *PluginTUIScreen) rebuildGrid() {
	if s.layout == nil {
		return
	}

	cols := make([]GridColumn, 0, len(s.layout.Columns))
	for _, c := range s.layout.Columns {
		cells := make([]GridCell, 0, len(c.Cells))
		for _, cell := range c.Cells {
			cells = append(cells, GridCell{
				Height: cell.Height,
				Title:  cell.Title,
			})
		}
		cols = append(cols, GridColumn{
			Span:  c.Span,
			Cells: cells,
		})
	}
	s.Grid = Grid{Columns: cols}
	s.CursorMax = s.Grid.CellCount() - 1
	if s.CursorMax < 0 {
		s.CursorMax = 0
	}
	if s.FocusIndex > s.CursorMax {
		s.FocusIndex = 0
	}
}

func (s *PluginTUIScreen) View(ctx AppContext) string {
	if s.errMsg != "" {
		return fmt.Sprintf("\n  Plugin Error: %s\n\n  Press any key to go back.", s.errMsg)
	}

	if s.bridge == nil {
		return "\n  Loading plugin screen..."
	}

	// Single primitive mode (no grid).
	if s.primitive != nil {
		return plugin.RenderPrimitive(s.primitive, ctx.Width, ctx.Height, true)
	}

	if s.layout == nil {
		return "\n  Plugin screen has no content."
	}

	// Build cell contents from layout.
	cells := make([]CellContent, 0)
	for _, col := range s.layout.Columns {
		for _, cell := range col.Cells {
			content := ""
			totalLines := 0
			if cell.Content != nil {
				content = plugin.RenderPrimitive(cell.Content, ctx.Width, ctx.Height, false)
				totalLines = strings.Count(content, "\n") + 1
			}
			cells = append(cells, CellContent{
				Content:    content,
				TotalLines: totalLines,
			})
		}
	}

	return s.RenderGrid(ctx, cells)
}

func (s *PluginTUIScreen) KeyHints(km config.KeyMap) []KeyHint {
	hints := make([]KeyHint, 0, len(s.hints)+3)

	// Plugin-defined hints.
	for _, h := range s.hints {
		hints = append(hints, KeyHint{Key: h.Key, Label: h.Label})
	}

	// Standard hints.
	hints = append(hints,
		KeyHint{km.HintString(config.ActionNextPanel), "panel"},
		KeyHint{km.HintString(config.ActionBack), "back"},
	)

	return hints
}
