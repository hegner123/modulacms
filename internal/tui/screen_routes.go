package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// Routes grid: 3 columns
//
//	Col 0 (span 3): Routes list
//	Col 1 (span 6): Details (65%), Info (35%)
//	Col 2 (span 3): Actions (50%), Stats (50%)
var routesGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1, Title: "Routes"},
		}},
		{Span: 6, Cells: []GridCell{
			{Height: 0.65, Title: "Details"},
			{Height: 0.35, Title: "Info"},
		}},
		{Span: 3, Cells: []GridCell{
			{Height: 0.50, Title: "Actions"},
			{Height: 0.50, Title: "Stats"},
		}},
	},
}

// RoutesScreen implements Screen for both the regular routes page (ROUTES) and
// the admin routes page (ADMINROUTES). AdminMode selects which data and
// dialog commands are used.
type RoutesScreen struct {
	GridScreen
	AdminMode   bool
	PageRouteId types.RouteID // active route selection (regular mode only)
	Routes      []db.Routes
	AdminRoutes []db.AdminRoutes
}

// NewRoutesScreen creates a RoutesScreen for either regular or admin mode.
func NewRoutesScreen(adminMode bool, routes []db.Routes, adminRoutes []db.AdminRoutes, pageRouteId types.RouteID) *RoutesScreen {
	return &RoutesScreen{
		GridScreen: GridScreen{
			Grid: routesGrid,
		},
		AdminMode:   adminMode,
		PageRouteId: pageRouteId,
		Routes:      routes,
		AdminRoutes: adminRoutes,
	}
}

func (s *RoutesScreen) PageIndex() PageIndex {
	if s.AdminMode {
		return ADMINROUTES
	}
	return ROUTES
}

func (s *RoutesScreen) cursorMax() int {
	if s.AdminMode {
		return len(s.AdminRoutes) - 1
	}
	return len(s.Routes) - 1
}

func (s *RoutesScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// Select route (only from routes list cell)
		if km.Matches(key, config.ActionSelect) && s.FocusIndex == 0 {
			if s.AdminMode {
				if len(s.AdminRoutes) > 0 && s.Cursor < len(s.AdminRoutes) {
					route := s.AdminRoutes[s.Cursor]
					return s, LogMessageCmd(fmt.Sprintf("Admin route selected: %s (%s)", route.Title, route.AdminRouteID))
				}
			} else {
				if len(s.Routes) > 0 && s.Cursor < len(s.Routes) {
					route := s.Routes[s.Cursor]
					s.PageRouteId = route.RouteID
					return s, LogMessageCmd(fmt.Sprintf("Route selected: %s (%s)", route.Title, route.RouteID))
				}
			}
		}

		// New route
		if km.Matches(key, config.ActionNew) {
			if s.AdminMode {
				return s, ShowRouteFormDialogCmd(FORMDIALOGCREATEADMINROUTE, "New Admin Route")
			}
			return s, ShowRouteFormDialogCmd(FORMDIALOGCREATEROUTE, "New Route")
		}

		// Edit route
		if km.Matches(key, config.ActionEdit) {
			if s.AdminMode {
				if len(s.AdminRoutes) > 0 && s.Cursor < len(s.AdminRoutes) {
					return s, ShowEditAdminRouteDialogCmd(s.AdminRoutes[s.Cursor])
				}
			} else {
				if len(s.Routes) > 0 && s.Cursor < len(s.Routes) {
					return s, ShowEditRouteDialogCmd(s.Routes[s.Cursor])
				}
			}
		}

		// Delete route
		if km.Matches(key, config.ActionDelete) {
			if s.AdminMode {
				if len(s.AdminRoutes) > 0 && s.Cursor < len(s.AdminRoutes) {
					route := s.AdminRoutes[s.Cursor]
					return s, ShowDeleteAdminRouteDialogCmd(route.AdminRouteID, route.Title)
				}
			} else {
				if len(s.Routes) > 0 && s.Cursor < len(s.Routes) {
					route := s.Routes[s.Cursor]
					return s, ShowDeleteRouteDialogCmd(route.RouteID, route.Title)
				}
			}
		}

		// Common keys LAST
		cMax := s.cursorMax()
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, cMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	// Fetch request messages
	case RoutesFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			routes, err := d.ListRoutes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if routes == nil {
				return RoutesFetchResultsMsg{Data: []db.Routes{}}
			}
			return RoutesFetchResultsMsg{Data: *routes}
		}
	case RoutesFetchResultsMsg:
		s.Routes = msg.Data
		s.Cursor = 0
		return s, LoadingStopCmd()
	case AdminRoutesFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			routes, err := d.ListAdminRoutes()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if routes == nil {
				return AdminRoutesFetchResultsMsg{Data: []db.AdminRoutes{}}
			}
			return AdminRoutesFetchResultsMsg{Data: *routes}
		}
	case AdminRoutesFetchResultsMsg:
		s.AdminRoutes = msg.Data
		s.Cursor = 0
		return s, LoadingStopCmd()

	// Data refresh messages (from CMS operations)
	case RoutesSet:
		s.Routes = msg.Routes
		return s, nil
	case AdminRoutesSet:
		s.AdminRoutes = msg.AdminRoutes
		return s, nil
	}

	return s, nil
}

func (s *RoutesScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionSelect), "select"},
		{km.HintString(config.ActionNew), "new"},
		{km.HintString(config.ActionEdit), "edit"},
		{km.HintString(config.ActionDelete), "del"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
	}
}

func (s *RoutesScreen) View(ctx AppContext) string {
	if s.AdminMode {
		return s.viewAdmin(ctx)
	}
	return s.viewRegular(ctx)
}

func (s *RoutesScreen) viewRegular(ctx AppContext) string {
	listLen := len(s.Routes)
	innerH := s.Grid.CellInnerHeight(0, ctx.Height)

	cells := []CellContent{
		{Content: s.renderRoutesList(), TotalLines: listLen, ScrollOffset: ClampScroll(s.Cursor, listLen, innerH)},
		{Content: s.renderRouteDetail()},
		{Content: s.renderRouteInfo()},
		{Content: s.renderRouteActions()},
		{Content: s.renderRouteStats()},
	}
	return s.RenderGrid(ctx, cells)
}

func (s *RoutesScreen) viewAdmin(ctx AppContext) string {
	listLen := len(s.AdminRoutes)
	innerH := s.Grid.CellInnerHeight(0, ctx.Height)

	cells := []CellContent{
		{Content: s.renderAdminRoutesList(), TotalLines: listLen, ScrollOffset: ClampScroll(s.Cursor, listLen, innerH)},
		{Content: s.renderAdminRouteDetail()},
		{Content: s.renderAdminRouteInfo()},
		{Content: s.renderAdminRouteActions()},
		{Content: s.renderAdminRouteStats()},
	}
	return s.RenderGrid(ctx, cells)
}

// ---------------------------------------------------------------------------
// Regular routes render methods
// ---------------------------------------------------------------------------

func (s *RoutesScreen) renderRoutesList() string {
	if len(s.Routes) == 0 {
		return " (no routes)"
	}

	lines := make([]string, 0, len(s.Routes))
	for i, route := range s.Routes {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		active := ""
		if route.RouteID == s.PageRouteId {
			active = " *"
		}
		lines = append(lines, fmt.Sprintf(" %s %s %s%s", cursor, route.Title, route.Slug, active))
	}
	return strings.Join(lines, "\n")
}

func (s *RoutesScreen) renderRouteDetail() string {
	if len(s.Routes) == 0 || s.Cursor >= len(s.Routes) {
		return " No route selected"
	}

	route := s.Routes[s.Cursor]
	accent := lipgloss.NewStyle().Bold(true)

	lines := []string{
		accent.Render(" " + route.Title),
		"",
		fmt.Sprintf(" Slug      %s", route.Slug),
		fmt.Sprintf(" Status    %d", route.Status),
		fmt.Sprintf(" Author    %s", route.AuthorID.String()),
		fmt.Sprintf(" Created   %s", route.DateCreated.String()),
		fmt.Sprintf(" Modified  %s", route.DateModified.String()),
	}

	if route.RouteID == s.PageRouteId {
		lines = append(lines, "", " (active route)")
	}

	return strings.Join(lines, "\n")
}

func (s *RoutesScreen) renderRouteInfo() string {
	if len(s.Routes) == 0 || s.Cursor >= len(s.Routes) {
		return ""
	}

	route := s.Routes[s.Cursor]
	lines := []string{
		fmt.Sprintf(" ID  %s", route.RouteID.String()),
	}

	return strings.Join(lines, "\n")
}

func (s *RoutesScreen) renderRouteActions() string {
	lines := []string{
		" n  New route",
		" e  Edit selected",
		" d  Delete selected",
	}
	return strings.Join(lines, "\n")
}

func (s *RoutesScreen) renderRouteStats() string {
	lines := []string{
		fmt.Sprintf(" Total   %d", len(s.Routes)),
	}

	if !s.PageRouteId.IsZero() {
		lines = append(lines, fmt.Sprintf(" Active  %s", s.PageRouteId))
	}

	return strings.Join(lines, "\n")
}

// ---------------------------------------------------------------------------
// Admin routes render methods
// ---------------------------------------------------------------------------

func (s *RoutesScreen) renderAdminRoutesList() string {
	if len(s.AdminRoutes) == 0 {
		return " (no admin routes)"
	}

	lines := make([]string, 0, len(s.AdminRoutes))
	for i, route := range s.AdminRoutes {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		lines = append(lines, fmt.Sprintf(" %s %s %s", cursor, route.Title, route.Slug))
	}
	return strings.Join(lines, "\n")
}

func (s *RoutesScreen) renderAdminRouteDetail() string {
	if len(s.AdminRoutes) == 0 || s.Cursor >= len(s.AdminRoutes) {
		return " No admin route selected"
	}

	route := s.AdminRoutes[s.Cursor]
	accent := lipgloss.NewStyle().Bold(true)

	lines := []string{
		accent.Render(" " + route.Title),
		"",
		fmt.Sprintf(" Slug      %s", route.Slug),
		fmt.Sprintf(" Status    %d", route.Status),
		fmt.Sprintf(" Author    %s", route.AuthorID.String()),
		fmt.Sprintf(" Created   %s", route.DateCreated.String()),
		fmt.Sprintf(" Modified  %s", route.DateModified.String()),
	}

	return strings.Join(lines, "\n")
}

func (s *RoutesScreen) renderAdminRouteInfo() string {
	if len(s.AdminRoutes) == 0 || s.Cursor >= len(s.AdminRoutes) {
		return ""
	}

	route := s.AdminRoutes[s.Cursor]
	lines := []string{
		fmt.Sprintf(" ID  %s", route.AdminRouteID.String()),
	}
	return strings.Join(lines, "\n")
}

func (s *RoutesScreen) renderAdminRouteActions() string {
	lines := []string{
		" n  New route",
		" e  Edit selected",
		" d  Delete selected",
	}
	return strings.Join(lines, "\n")
}

func (s *RoutesScreen) renderAdminRouteStats() string {
	lines := []string{
		fmt.Sprintf(" Total  %d", len(s.AdminRoutes)),
	}
	return strings.Join(lines, "\n")
}
