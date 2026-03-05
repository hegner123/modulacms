package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
)

// HomeDashboardData holds all data displayed on the home dashboard.
type HomeDashboardData struct {
	ContentCount  int64
	DatatypeCount int64
	MediaCount    int64
	UserCount     int64
	RouteCount    int64
	Plugins       []db.Plugin
	Backups       []db.Backup
	RecentEvents  []db.ChangeEvent
}

// HomeDashboardDataMsg delivers fetched dashboard data to the home screen.
type HomeDashboardDataMsg struct {
	Data HomeDashboardData
	Err  error
}

// HomeDashboardFetchCmd fetches all dashboard data from the database.
func HomeDashboardFetchCmd(driver db.DbDriver) tea.Cmd {
	return func() tea.Msg {
		var data HomeDashboardData

		if c, err := driver.CountContentData(); err == nil && c != nil {
			data.ContentCount = *c
		}
		if c, err := driver.CountDatatypes(); err == nil && c != nil {
			data.DatatypeCount = *c
		}
		if c, err := driver.CountMedia(); err == nil && c != nil {
			data.MediaCount = *c
		}
		if c, err := driver.CountUsers(); err == nil && c != nil {
			data.UserCount = *c
		}
		if c, err := driver.CountRoutes(); err == nil && c != nil {
			data.RouteCount = *c
		}
		if plugins, err := driver.ListPlugins(); err == nil && plugins != nil {
			data.Plugins = *plugins
		}
		if backups, err := driver.ListBackups(db.ListBackupsParams{Limit: 5, Offset: 0}); err == nil && backups != nil {
			data.Backups = *backups
		}
		if events, err := driver.ListChangeEvents(db.ListChangeEventsParams{Limit: 10, Offset: 0}); err == nil && events != nil {
			data.RecentEvents = *events
		}

		return HomeDashboardDataMsg{Data: data}
	}
}

// HomeScreen implements Screen for the home dashboard.
type HomeScreen struct {
	GridScreen
	Menu     []Page
	Username string
	Data     HomeDashboardData
	Loaded   bool
}

// Home dashboard grid: 3 columns
//
//	Col 0 (span 2): Nav, Site Config
//	Col 1 (span 5): Plugins, Connections
//	Col 2 (span 5): Recent Activity, Backups
var homeGrid = Grid{
	Columns: []GridColumn{
		{Span: 2, Cells: []GridCell{
			{Height: 0.70, Title: "Nav"},
			{Height: 0.30, Title: "Site"},
		}},
		{Span: 5, Cells: []GridCell{
			{Height: 0.65, Title: "Plugins"},
			{Height: 0.35, Title: "Connections"},
		}},
		{Span: 5, Cells: []GridCell{
			{Height: 0.65, Title: "Activity"},
			{Height: 0.35, Title: "Backups"},
		}},
	},
}

// NewHomeScreen creates a HomeScreen with the given menu and username.
func NewHomeScreen(menu []Page, username string) *HomeScreen {
	return &HomeScreen{
		GridScreen: GridScreen{
			Grid:      homeGrid,
			CursorMax: len(menu) - 1,
		},
		Menu:     menu,
		Username: username,
	}
}

func (s *HomeScreen) PageIndex() PageIndex { return HOMEPAGE }

func (s *HomeScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case PageMenuSet:
		s.Menu = msg.PageMenu
		s.CursorMax = len(s.Menu) - 1
		if s.Cursor > s.CursorMax && s.CursorMax >= 0 {
			s.Cursor = s.CursorMax
		}
		return s, nil

	case HomeDashboardDataMsg:
		if msg.Err == nil {
			s.Data = msg.Data
			s.Loaded = true
		}
		return s, nil

	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		if km.Matches(key, config.ActionTitlePrev) {
			return s, TitleFontPreviousCmd()
		}
		if km.Matches(key, config.ActionTitleNext) {
			return s, TitleFontNextCmd()
		}

		if km.Matches(key, config.ActionSelect) {
			if s.FocusIndex == 0 && s.Cursor < len(s.Menu) {
				target := s.Menu[s.Cursor]
				if ctx.AdminMode {
					target.Index = AdminPageIndex(target.Index)
				}
				return s, NavigateToPageCmd(target)
			}
		}

		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}
	}
	return s, nil
}

func (s *HomeScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionSelect), "select"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionAdminToggle), "admin/client"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *HomeScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderNav()},
		{Content: s.renderSiteConfig(ctx)},
		{Content: s.renderPlugins()},
		{Content: s.renderConnections(ctx)},
		{Content: s.renderActivity()},
		{Content: s.renderBackups()},
	}
	return s.RenderGrid(ctx, cells)
}

func (s *HomeScreen) renderNav() string {
	if len(s.Menu) == 0 {
		return "(no items)"
	}
	lines := make([]string, 0, len(s.Menu))
	for i, item := range s.Menu {
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, item.Label))
	}
	return strings.Join(lines, "\n")
}

func (s *HomeScreen) renderSiteConfig(ctx AppContext) string {
	if ctx.Config == nil {
		return "(no config)"
	}
	c := ctx.Config
	lines := []string{
		fmt.Sprintf(" HTTP %s", c.Port),
		fmt.Sprintf(" SSL  %s", c.SSL_Port),
		fmt.Sprintf(" SSH  %s", c.SSH_Port),
	}
	if len(c.Cors_Origins) > 0 {
		lines = append(lines, fmt.Sprintf(" CORS %s", c.Cors_Origins[0]))
	}
	return strings.Join(lines, "\n")
}

func (s *HomeScreen) renderPlugins() string {
	if !s.Loaded {
		return " Loading..."
	}
	if len(s.Data.Plugins) == 0 {
		return " No plugins installed"
	}

	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	faint := lipgloss.NewStyle().Faint(true)

	lines := make([]string, 0, len(s.Data.Plugins))
	for _, p := range s.Data.Plugins {
		status := faint.Render(string(p.Status))
		if p.Status == "active" {
			status = accent.Render(string(p.Status))
		}
		lines = append(lines, fmt.Sprintf(" %s  %s", p.Name, status))
	}
	return strings.Join(lines, "\n")
}

func (s *HomeScreen) renderConnections(ctx AppContext) string {
	if ctx.Config == nil {
		return " (no config)"
	}
	c := ctx.Config

	okStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	failStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn)
	offStyle := lipgloss.NewStyle().Faint(true)

	dbStatus := okStyle.Render("OK")
	dbLabel := string(c.Db_Driver)

	s3Status := offStyle.Render("OFF")
	s3Label := "not configured"
	if c.Bucket_Endpoint != "" {
		s3Status = okStyle.Render("OK")
		s3Label = c.Bucket_Endpoint
	}

	remoteStatus := offStyle.Render("OFF")
	remoteLabel := "n/a"
	if ctx.IsRemote {
		remoteLabel = ctx.Config.Remote_URL
		remoteStatus = okStyle.Render("OK")
		if ctx.SSHFingerprint == "" {
			remoteStatus = failStyle.Render("FAIL")
		}
	}

	lines := []string{
		fmt.Sprintf(" DB:     %-12s %s", dbLabel, dbStatus),
		fmt.Sprintf(" S3:     %-12s %s", s3Label, s3Status),
		fmt.Sprintf(" Remote: %-12s %s", remoteLabel, remoteStatus),
	}
	return strings.Join(lines, "\n")
}

func (s *HomeScreen) renderActivity() string {
	if !s.Loaded {
		return " Loading..."
	}
	if len(s.Data.RecentEvents) == 0 {
		return " No recent activity"
	}

	faint := lipgloss.NewStyle().Faint(true)

	lines := make([]string, 0, len(s.Data.RecentEvents))
	for _, e := range s.Data.RecentEvents {
		ts := ""
		if e.WallTimestamp.Valid {
			ts = e.WallTimestamp.Time.Format("15:04")
		}
		lines = append(lines, fmt.Sprintf(" %s %s %s", faint.Render(ts), e.Operation, e.TableName))
	}
	return strings.Join(lines, "\n")
}

func (s *HomeScreen) renderBackups() string {
	if !s.Loaded {
		return " Loading..."
	}
	if len(s.Data.Backups) == 0 {
		return " No backups"
	}

	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	warn := lipgloss.NewStyle().Foreground(config.DefaultStyle.Warn)

	lines := make([]string, 0, len(s.Data.Backups))
	for _, b := range s.Data.Backups {
		status := accent.Render(string(b.Status))
		if b.Status != "completed" {
			status = warn.Render(string(b.Status))
		}
		ts := ""
		if b.StartedAt.Valid {
			ts = b.StartedAt.Time.Format("2006-01-02")
		}
		lines = append(lines, fmt.Sprintf(" %s  %s  %s", ts, b.BackupType, status))
	}
	return strings.Join(lines, "\n")
}

// renderVersion returns the version string for the title header.
func (s *HomeScreen) renderVersion() string {
	return "v" + utility.Version
}
