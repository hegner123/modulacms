package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

var mediaDimensionsGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Dimensions"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.6, Title: "Details"},
			{Height: 0.4, Title: "Info"},
		}},
	},
}

// MediaDimensionsScreen implements Screen for media dimension presets.
type MediaDimensionsScreen struct {
	GridScreen
	DimensionsList []db.MediaDimensions
}

func NewMediaDimensionsScreen(dims []db.MediaDimensions) *MediaDimensionsScreen {
	cursorMax := len(dims) - 1
	if cursorMax < 0 {
		cursorMax = 0
	}
	return &MediaDimensionsScreen{
		GridScreen: GridScreen{
			Grid:      mediaDimensionsGrid,
			CursorMax: cursorMax,
		},
		DimensionsList: dims,
	}
}

func (s *MediaDimensionsScreen) PageIndex() PageIndex { return MEDIADIMENSIONSPAGE }

func (s *MediaDimensionsScreen) selectedDimension() *db.MediaDimensions {
	if len(s.DimensionsList) == 0 || s.Cursor >= len(s.DimensionsList) {
		return nil
	}
	return &s.DimensionsList[s.Cursor]
}

func (s *MediaDimensionsScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		if km.Matches(key, config.ActionNew) {
			return s, ShowCreateMediaDimensionDialogCmd()
		}

		if km.Matches(key, config.ActionEdit) {
			if dim := s.selectedDimension(); dim != nil {
				return s, ShowEditMediaDimensionDialogCmd(*dim)
			}
		}

		if km.Matches(key, config.ActionDelete) {
			if dim := s.selectedDimension(); dim != nil {
				label := dim.MdID[:8]
				if dim.Label.Valid {
					label = dim.Label.String
				}
				return s, ShowDeleteMediaDimensionDialogCmd(dim.MdID, label)
			}
		}

		cursorMax := len(s.DimensionsList) - 1
		if cursorMax < 0 {
			cursorMax = 0
		}
		s.CursorMax = cursorMax
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	case MediaDimensionsFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			list, err := d.ListMediaDimensions()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			data := make([]db.MediaDimensions, 0)
			if list != nil {
				data = *list
			}
			return MediaDimensionsFetchResultsMsg{Data: data}
		}

	case MediaDimensionsFetchResultsMsg:
		s.DimensionsList = msg.Data
		s.Cursor = 0
		s.CursorMax = len(s.DimensionsList) - 1
		if s.CursorMax < 0 {
			s.CursorMax = 0
		}
		return s, LoadingStopCmd()

	case MediaDimensionCreatedMsg, MediaDimensionUpdatedMsg, MediaDimensionDeletedMsg:
		s.Cursor = 0
		return s, MediaDimensionsFetchCmd()
	}

	return s, nil
}

func (s *MediaDimensionsScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionNew), "new"},
		{km.HintString(config.ActionEdit), "edit"},
		{km.HintString(config.ActionDelete), "del"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
	}
}

func (s *MediaDimensionsScreen) View(ctx AppContext) string {
	cells := []CellContent{
		{Content: s.renderList(), TotalLines: len(s.DimensionsList), ScrollOffset: ClampScroll(s.Cursor, len(s.DimensionsList), ctx.Height)},
		{Content: s.renderDetail()},
		{Content: s.renderInfo()},
	}
	return s.RenderGrid(ctx, cells)
}

func (s *MediaDimensionsScreen) renderList() string {
	if len(s.DimensionsList) == 0 {
		return "(no dimensions)"
	}
	lines := make([]string, 0, len(s.DimensionsList))
	for i, dim := range s.DimensionsList {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		label := dim.MdID[:8]
		if dim.Label.Valid {
			label = dim.Label.String
		}
		size := ""
		if dim.Width.Valid && dim.Height.Valid {
			size = fmt.Sprintf(" %dx%d", dim.Width.Int64, dim.Height.Int64)
		}
		lines = append(lines, fmt.Sprintf("%s %s%s", cursor, label, size))
	}
	return strings.Join(lines, "\n")
}

func (s *MediaDimensionsScreen) renderDetail() string {
	if len(s.DimensionsList) == 0 || s.Cursor >= len(s.DimensionsList) {
		return " No dimension selected"
	}
	dim := s.DimensionsList[s.Cursor]
	label := "(none)"
	if dim.Label.Valid {
		label = dim.Label.String
	}
	width := "(none)"
	if dim.Width.Valid {
		width = fmt.Sprintf("%d px", dim.Width.Int64)
	}
	height := "(none)"
	if dim.Height.Valid {
		height = fmt.Sprintf("%d px", dim.Height.Int64)
	}
	aspect := "(none)"
	if dim.AspectRatio.Valid {
		aspect = dim.AspectRatio.String
	}
	lines := []string{
		fmt.Sprintf(" ID      %s", dim.MdID),
		fmt.Sprintf(" Label   %s", label),
		fmt.Sprintf(" Width   %s", width),
		fmt.Sprintf(" Height  %s", height),
		fmt.Sprintf(" Ratio   %s", aspect),
	}
	return strings.Join(lines, "\n")
}

func (s *MediaDimensionsScreen) renderInfo() string {
	lines := []string{
		" Media Dimensions",
		"",
		fmt.Sprintf("   Presets: %d", len(s.DimensionsList)),
		"",
		" Dimension presets define image",
		" sizes for automatic resizing",
		" on media upload.",
	}
	return strings.Join(lines, "\n")
}
