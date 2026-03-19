package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/service"
)

var importGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Format"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.5, Title: "Import"},
			{Height: 0.5, Title: "Results"},
		}},
	},
}

// ImportFileSelectedMsg is sent when the file picker returns a path.
type ImportFileSelectedMsg struct {
	Path string
}

// ImportRequestMsg triggers the actual import operation.
type ImportRequestMsg struct {
	Format config.OutputFormat
	Path   string
}

// ImportCompleteMsg carries the result of an import operation.
type ImportCompleteMsg struct {
	Result *service.ImportResult
	Err    error
}

// importFormat describes a supported import format.
type importFormat struct {
	Label  string
	Format config.OutputFormat
	Desc   string
}

var importFormats = []importFormat{
	{"Contentful", config.FormatContentful, "Import from Contentful CMS export JSON"},
	{"Sanity", config.FormatSanity, "Import from Sanity dataset export"},
	{"Strapi", config.FormatStrapi, "Import from Strapi content export"},
	{"WordPress", config.FormatWordPress, "Import from WordPress REST API JSON"},
	{"Clean JSON", config.FormatClean, "Import ModulaCMS native JSON format"},
}

// ImportScreen implements Screen for the content import wizard.
type ImportScreen struct {
	GridScreen
	Formats    []importFormat
	FilePath   string                // selected file path
	Importing  bool                  // true while import is running
	Result     *service.ImportResult // nil until import completes
	ResultErr  error                 // non-nil if import failed
}

func NewImportScreen() *ImportScreen {
	return &ImportScreen{
		GridScreen: GridScreen{
			Grid:      importGrid,
			CursorMax: len(importFormats) - 1,
		},
		Formats: importFormats,
	}
}

func (s *ImportScreen) PageIndex() PageIndex { return IMPORTPAGE }

func (s *ImportScreen) selectedFormat() *importFormat {
	if s.Cursor >= len(s.Formats) {
		return nil
	}
	return &s.Formats[s.Cursor]
}

func (s *ImportScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// Enter on format list → open file picker
		if km.Matches(key, config.ActionSelect) && !s.Importing {
			s.FilePath = ""
			s.Result = nil
			s.ResultErr = nil
			return s, func() tea.Msg {
				return OpenFilePickerMsg{Purpose: FILEPICKER_IMPORT}
			}
		}

		cursorMax := len(s.Formats) - 1
		if cursorMax < 0 {
			cursorMax = 0
		}
		s.CursorMax = cursorMax
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	// File selected from picker
	case ImportFileSelectedMsg:
		s.FilePath = msg.Path
		f := s.selectedFormat()
		if f == nil {
			return s, nil
		}
		s.Importing = true
		s.Result = nil
		s.ResultErr = nil
		return s, func() tea.Msg {
			return ImportRequestMsg{Format: f.Format, Path: msg.Path}
		}

	// Import request — execute via service
	case ImportRequestMsg:
		return s, RunImportCmd(ctx, msg.Format, msg.Path)

	// Import complete
	case ImportCompleteMsg:
		s.Importing = false
		s.Result = msg.Result
		s.ResultErr = msg.Err
		return s, LoadingStopCmd()
	}

	return s, nil
}

func (s *ImportScreen) KeyHints(km config.KeyMap) []KeyHint {
	hints := []KeyHint{
		{km.HintString(config.ActionSelect), "select file"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
	return hints
}

func (s *ImportScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderFormatList()},
		{Content: s.renderImportPanel()},
		{Content: s.renderResults()},
	}
	return s.RenderGrid(ctx, cells)
}

func (s *ImportScreen) renderFormatList() string {
	lines := make([]string, 0, len(s.Formats))
	for i, f := range s.Formats {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		lines = append(lines, fmt.Sprintf("%s %s", cursor, f.Label))
	}
	return strings.Join(lines, "\n")
}

func (s *ImportScreen) renderImportPanel() string {
	accent := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)
	faint := lipgloss.NewStyle().Faint(true)

	f := s.selectedFormat()
	if f == nil {
		return " Select a format"
	}

	lines := []string{
		accent.Render(fmt.Sprintf(" %s", f.Label)),
		faint.Render(fmt.Sprintf(" %s", f.Desc)),
		"",
	}

	if s.FilePath == "" {
		lines = append(lines, " Press enter to select a JSON file.")
	} else {
		lines = append(lines, fmt.Sprintf(" File: %s", s.FilePath))
	}

	if s.Importing {
		lines = append(lines, "", " Importing...")
	}

	return strings.Join(lines, "\n")
}

func (s *ImportScreen) renderResults() string {
	if s.Result == nil && s.ResultErr == nil {
		return lipgloss.NewStyle().Faint(true).Render(" No import results yet")
	}

	if s.ResultErr != nil {
		return fmt.Sprintf(" Import failed:\n %s", s.ResultErr)
	}

	r := s.Result
	status := "Success"
	if !r.Success {
		status = "Partial"
	}

	lines := []string{
		fmt.Sprintf(" Status:     %s", status),
		fmt.Sprintf(" Datatypes:  %d created", r.DatatypesCreated),
		fmt.Sprintf(" Fields:     %d created", r.FieldsCreated),
		fmt.Sprintf(" Content:    %d created", r.ContentCreated),
	}

	if r.Message != "" {
		lines = append(lines, "", fmt.Sprintf(" %s", r.Message))
	}

	if len(r.Errors) > 0 {
		lines = append(lines, "", fmt.Sprintf(" Errors (%d):", len(r.Errors)))
		for _, e := range r.Errors {
			lines = append(lines, fmt.Sprintf("   - %s", e))
		}
	}

	return strings.Join(lines, "\n")
}
