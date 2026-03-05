package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// MediaScreen implements Screen for the media library page.
type MediaScreen struct {
	Cursor     int
	PanelFocus FocusPanel
	MediaList  []db.Media
}

// NewMediaScreen creates a MediaScreen with the provided media list.
func NewMediaScreen(mediaList []db.Media) *MediaScreen {
	return &MediaScreen{
		Cursor:     0,
		PanelFocus: TreePanel,
		MediaList:  mediaList,
	}
}

func (s *MediaScreen) PageIndex() PageIndex { return MEDIA }

func (s *MediaScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		// Panel navigation
		if km.Matches(key, config.ActionNextPanel) {
			s.PanelFocus = (s.PanelFocus + 1) % 3
			return s, nil
		}
		if km.Matches(key, config.ActionPrevPanel) {
			s.PanelFocus = (s.PanelFocus + 2) % 3
			return s, nil
		}

		// Upload new media (open file picker)
		if km.Matches(key, config.ActionNew) {
			return s, func() tea.Msg {
				return OpenFilePickerMsg{Purpose: FILEPICKER_MEDIA}
			}
		}

		// Delete selected media
		if km.Matches(key, config.ActionDelete) {
			if len(s.MediaList) > 0 && s.Cursor < len(s.MediaList) {
				media := s.MediaList[s.Cursor]
				label := media.MediaID.String()
				if media.DisplayName.Valid && media.DisplayName.String != "" {
					label = media.DisplayName.String
				} else if media.Name.Valid && media.Name.String != "" {
					label = media.Name.String
				}
				return s, ShowDeleteMediaDialogCmd(media.MediaID, label)
			}
		}

		// Common keys (quit, back, cursor)
		cursorMax := len(s.MediaList) - 1
		if cursorMax < 0 {
			cursorMax = 0
		}
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, cursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	// Fetch request messages
	case MediaFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			media, err := d.ListMedia()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			if media == nil {
				return MediaFetchResultsMsg{Data: []db.Media{}}
			}
			return MediaFetchResultsMsg{Data: *media}
		}
	case MediaFetchResultsMsg:
		s.MediaList = msg.Data
		s.Cursor = 0
		if s.Cursor >= len(s.MediaList) {
			s.Cursor = len(s.MediaList) - 1
		}
		if s.Cursor < 0 {
			s.Cursor = 0
		}
		return s, LoadingStopCmd()

	// Data refresh (from CMS operations)
	case MediaListSet:
		s.MediaList = msg.MediaList
		if s.Cursor >= len(s.MediaList) {
			s.Cursor = len(s.MediaList) - 1
		}
		if s.Cursor < 0 {
			s.Cursor = 0
		}
		return s, nil
	}

	return s, nil
}

func (s *MediaScreen) KeyHints(km config.KeyMap) []KeyHint {
	return []KeyHint{
		{km.HintString(config.ActionNew), "upload"},
		{km.HintString(config.ActionDelete), "del"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
		{km.HintString(config.ActionQuit), "quit"},
	}
}

func (s *MediaScreen) View(ctx AppContext) string {
	left := s.renderMediaList()
	center := s.renderMediaDetail()
	right := s.renderMediaInfo()

	layout := layoutForPage(MEDIA)
	leftW := int(float64(ctx.Width) * layout.Ratios[0])
	centerW := int(float64(ctx.Width) * layout.Ratios[1])
	rightW := ctx.Width - leftW - centerW

	if layout.Panels == 1 {
		leftW, rightW = 0, 0
		centerW = ctx.Width
	}

	innerH := PanelInnerHeight(ctx.Height)
	listLen := len(s.MediaList)

	var panels []string
	if leftW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[0], Width: leftW, Height: ctx.Height, Content: left, Focused: s.PanelFocus == TreePanel, TotalLines: listLen, ScrollOffset: ClampScroll(s.Cursor, listLen, innerH)}.Render())
	}
	if centerW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[1], Width: centerW, Height: ctx.Height, Content: center, Focused: s.PanelFocus == ContentPanel}.Render())
	}
	if rightW > 0 {
		panels = append(panels, Panel{Title: layout.Titles[2], Width: rightW, Height: ctx.Height, Content: right, Focused: s.PanelFocus == RoutePanel}.Render())
	}

	return strings.Join(panels, "")
}

func (s *MediaScreen) renderMediaList() string {
	if len(s.MediaList) == 0 {
		return "(no media)"
	}

	lines := make([]string, 0, len(s.MediaList))
	for i, media := range s.MediaList {
		cursor := "   "
		if s.Cursor == i {
			cursor = " ->"
		}
		name := media.MediaID.String()
		if media.DisplayName.Valid && media.DisplayName.String != "" {
			name = media.DisplayName.String
		} else if media.Name.Valid && media.Name.String != "" {
			name = media.Name.String
		}
		mime := ""
		if media.Mimetype.Valid && media.Mimetype.String != "" {
			mime = " [" + media.Mimetype.String + "]"
		}
		lines = append(lines, fmt.Sprintf("%s %s%s", cursor, name, mime))
	}
	return strings.Join(lines, "\n")
}

func (s *MediaScreen) renderMediaDetail() string {
	if len(s.MediaList) == 0 || s.Cursor >= len(s.MediaList) {
		return "No media selected"
	}

	media := s.MediaList[s.Cursor]

	nullStr := func(ns db.NullString) string {
		if ns.Valid {
			return ns.String
		}
		return "(none)"
	}

	lines := []string{
		fmt.Sprintf("Name:        %s", nullStr(media.Name)),
		fmt.Sprintf("Display:     %s", nullStr(media.DisplayName)),
		fmt.Sprintf("Alt:         %s", nullStr(media.Alt)),
		fmt.Sprintf("Caption:     %s", nullStr(media.Caption)),
		fmt.Sprintf("Description: %s", nullStr(media.Description)),
		"",
		fmt.Sprintf("Mimetype:    %s", nullStr(media.Mimetype)),
		fmt.Sprintf("Dimensions:  %s", nullStr(media.Dimensions)),
		fmt.Sprintf("URL:         %s", media.URL),
		"",
		fmt.Sprintf("Created:     %s", media.DateCreated.String()),
		fmt.Sprintf("Modified:    %s", media.DateModified.String()),
	}

	return strings.Join(lines, "\n")
}

func (s *MediaScreen) renderMediaInfo() string {
	lines := []string{
		"Media Library",
		"",
		fmt.Sprintf("  Total: %d", len(s.MediaList)),
	}

	if len(s.MediaList) > 0 && s.Cursor < len(s.MediaList) {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  ID: %s", s.MediaList[s.Cursor].MediaID))
		if s.MediaList[s.Cursor].Class.Valid && s.MediaList[s.Cursor].Class.String != "" {
			lines = append(lines, fmt.Sprintf("  Class: %s", s.MediaList[s.Cursor].Class.String))
		}
	}

	return strings.Join(lines, "\n")
}
