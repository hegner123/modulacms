package tui

import (
	"fmt"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// Media grid: 2 columns
//
//	Col 0 (span 3): Media tree with inline search
//	Col 1 (span 9): Summary (top), Metadata (bottom)
var mediaGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Media"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.40, Title: "Summary"},
			{Height: 0.60, Title: "Metadata"},
		}},
	},
}

// MediaScreen implements Screen for the media library page.
type MediaScreen struct {
	GridScreen
	MediaList    []db.Media       // full list from DB
	FilteredList []db.Media       // displayed subset (== MediaList when no filter)
	MediaTree    []*MediaTreeNode // tree built from FilteredList
	FlatList     []*MediaTreeNode // flattened for cursor navigation
	Searching    bool             // true when search input is active
	SearchInput  textinput.Model
	SearchQuery  string // persisted query (set on enter)
}

// NewMediaScreen creates a MediaScreen with the provided media list.
func NewMediaScreen(mediaList []db.Media) *MediaScreen {
	if mediaList == nil {
		mediaList = []db.Media{}
	}

	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 128

	s := &MediaScreen{
		GridScreen: GridScreen{
			Grid:       mediaGrid,
			FocusIndex: 0,
		},
		MediaList:    mediaList,
		FilteredList: mediaList,
		SearchInput:  ti,
	}
	s.rebuildTree()
	return s
}

func (s *MediaScreen) PageIndex() PageIndex { return MEDIA }

func (s *MediaScreen) rebuildTree() {
	s.MediaTree = BuildMediaTree(s.FilteredList)
	s.FlatList = FlattenMediaTree(s.MediaTree)
	s.CursorMax = len(s.FlatList) - 1
	if s.CursorMax < 0 {
		s.CursorMax = 0
	}
	if s.Cursor > s.CursorMax {
		s.Cursor = s.CursorMax
	}
}

// selectedMedia returns the db.Media for the currently selected file node, or nil.
func (s *MediaScreen) selectedMedia() *db.Media {
	if len(s.FlatList) == 0 || s.Cursor >= len(s.FlatList) {
		return nil
	}
	node := s.FlatList[s.Cursor]
	if node.Kind == MediaNodeFile {
		return node.Media
	}
	return nil
}

func (s *MediaScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		km := ctx.Config.KeyBindings
		key := msg.String()

		// Search mode: all input goes to textinput except enter/esc
		if s.Searching {
			switch key {
			case "enter":
				s.Searching = false
				s.SearchQuery = s.SearchInput.Value()
				s.SearchInput.Blur()
				return s, nil
			case "esc":
				s.Searching = false
				s.SearchQuery = ""
				s.SearchInput.SetValue("")
				s.SearchInput.Blur()
				s.FilteredList = s.MediaList
				s.Cursor = 0
				s.rebuildTree()
				return s, nil
			default:
				var cmd tea.Cmd
				s.SearchInput, cmd = s.SearchInput.Update(msg)
				// Live filter on each keystroke
				s.FilteredList = FilterMediaList(s.MediaList, s.SearchInput.Value())
				s.Cursor = 0
				s.rebuildTree()
				return s, cmd
			}
		}

		// Browse mode
		if s.HandleFocusNav(key, km) {
			return s, nil
		}

		// Search activation
		if km.Matches(key, config.ActionSearch) {
			s.Searching = true
			s.SearchInput.Focus()
			return s, textinput.Blink
		}

		// Expand/collapse or select on enter (only when tree panel focused)
		if km.Matches(key, config.ActionSelect) && s.FocusIndex == 0 {
			if len(s.FlatList) > 0 && s.Cursor < len(s.FlatList) {
				node := s.FlatList[s.Cursor]
				if node.Kind == MediaNodeFolder {
					node.Expand = !node.Expand
					s.FlatList = FlattenMediaTree(s.MediaTree)
					s.CursorMax = len(s.FlatList) - 1
					if s.CursorMax < 0 {
						s.CursorMax = 0
					}
					if s.Cursor > s.CursorMax {
						s.Cursor = s.CursorMax
					}
					return s, nil
				}
				// File node: no-op (details shown automatically)
			}
			return s, nil
		}

		// Upload new media (open file picker — requires local filesystem)
		if km.Matches(key, config.ActionNew) {
			if ctx.IsSSH {
				dialog := NewDialog(
					"Upload Not Available",
					"Media upload requires a local file picker and is not available over SSH connections.\n\nTo upload media, run the CMS binary locally and use the 'connect' command to access this server.",
					false,
					DIALOGGENERIC,
				)
				return s, tea.Batch(
					OverlaySetCmd(&dialog),
					FocusSetCmd(DIALOGFOCUS),
				)
			}
			return s, func() tea.Msg {
				return OpenFilePickerMsg{Purpose: FILEPICKER_MEDIA}
			}
		}

		// Delete selected media
		if km.Matches(key, config.ActionDelete) {
			media := s.selectedMedia()
			if media != nil {
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
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
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
		s.FilteredList = FilterMediaList(s.MediaList, s.SearchQuery)
		s.Cursor = 0
		s.rebuildTree()
		return s, LoadingStopCmd()

	// Data refresh (from CMS operations)
	case MediaListSet:
		s.MediaList = msg.MediaList
		s.FilteredList = FilterMediaList(s.MediaList, s.SearchQuery)
		s.rebuildTree()
		return s, nil
	}

	return s, nil
}

func (s *MediaScreen) KeyHints(km config.KeyMap) []KeyHint {
	if s.Searching {
		return []KeyHint{
			{"type", "filter"},
			{"enter", "accept"},
			{"esc", "clear"},
		}
	}
	return []KeyHint{
		{km.HintString(config.ActionSearch), "search"},
		{km.HintString(config.ActionNew), "upload"},
		{km.HintString(config.ActionDelete), "del"},
		{"enter", "expand"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
	}
}
