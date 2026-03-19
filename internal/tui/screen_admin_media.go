package tui

import (
	"fmt"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// Admin media grid: 2 columns
//
//	Col 0 (span 3): Media tree with inline search
//	Col 1 (span 9): Summary (top), Metadata (bottom)
var adminMediaGrid = Grid{
	Columns: []GridColumn{
		{Span: 3, Cells: []GridCell{
			{Height: 1.0, Title: "Admin Media"},
		}},
		{Span: 9, Cells: []GridCell{
			{Height: 0.40, Title: "Summary"},
			{Height: 0.60, Title: "Metadata"},
		}},
	},
}

// AdminMediaScreen implements Screen for the admin media library page.
type AdminMediaScreen struct {
	GridScreen
	MediaList    []db.AdminMedia       // full list from DB
	FolderList   []db.AdminMediaFolder // all folders from DB
	FilteredList []db.AdminMedia       // displayed subset (== MediaList when no filter)
	MediaTree    []*AdminMediaTreeNode // tree built from FilteredList + folders
	FlatList     []*AdminMediaTreeNode // flattened for cursor navigation
	Searching    bool                  // true when search input is active
	SearchInput  textinput.Model
	SearchQuery  string // persisted query (set on enter)
}

// NewAdminMediaScreen creates an AdminMediaScreen with the provided media list.
func NewAdminMediaScreen(mediaList []db.AdminMedia, folderList []db.AdminMediaFolder) *AdminMediaScreen {
	if mediaList == nil {
		mediaList = []db.AdminMedia{}
	}
	if folderList == nil {
		folderList = []db.AdminMediaFolder{}
	}

	ti := textinput.New()
	ti.Placeholder = "filter..."
	ti.CharLimit = 128

	s := &AdminMediaScreen{
		GridScreen: GridScreen{
			Grid:       adminMediaGrid,
			FocusIndex: 0,
		},
		MediaList:    mediaList,
		FolderList:   folderList,
		FilteredList: mediaList,
		SearchInput:  ti,
	}
	s.rebuildTree()
	return s
}

func (s *AdminMediaScreen) PageIndex() PageIndex { return ADMINMEDIA }

func (s *AdminMediaScreen) rebuildTree() {
	if s.SearchQuery != "" {
		filteredFolders, filteredItems := FilterAdminMediaTree(s.FolderList, s.FilteredList, s.SearchQuery)
		s.MediaTree = BuildAdminMediaTree(filteredFolders, filteredItems)
	} else {
		s.MediaTree = BuildAdminMediaTree(s.FolderList, s.FilteredList)
	}
	s.FlatList = FlattenAdminMediaTree(s.MediaTree)
	s.CursorMax = len(s.FlatList) - 1
	if s.CursorMax < 0 {
		s.CursorMax = 0
	}
	if s.Cursor > s.CursorMax {
		s.Cursor = s.CursorMax
	}
}

// selectedMedia returns the db.AdminMedia for the currently selected file node, or nil.
func (s *AdminMediaScreen) selectedMedia() *db.AdminMedia {
	if len(s.FlatList) == 0 || s.Cursor >= len(s.FlatList) {
		return nil
	}
	node := s.FlatList[s.Cursor]
	if node.Kind == MediaNodeFile {
		return node.Media
	}
	return nil
}

// selectedNode returns the currently selected tree node, or nil.
func (s *AdminMediaScreen) selectedNode() *AdminMediaTreeNode {
	if len(s.FlatList) == 0 || s.Cursor >= len(s.FlatList) {
		return nil
	}
	return s.FlatList[s.Cursor]
}

// selectedFolderID returns the folder context for the current cursor position.
func (s *AdminMediaScreen) selectedFolderID() types.NullableAdminMediaFolderID {
	node := s.selectedNode()
	if node == nil {
		return types.NullableAdminMediaFolderID{}
	}
	if node.Kind == MediaNodeFolder && !node.FolderID.IsZero() {
		return types.NullableAdminMediaFolderID{ID: node.FolderID, Valid: true}
	}
	if node.Kind == MediaNodeFile && node.Media != nil {
		return node.Media.FolderID
	}
	return types.NullableAdminMediaFolderID{}
}

func (s *AdminMediaScreen) Update(ctx AppContext, msg tea.Msg) (Screen, tea.Cmd) {
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
				s.FilteredList = FilterAdminMediaList(s.MediaList, s.SearchQuery)
				s.Cursor = 0
				s.rebuildTree()
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
				s.FilteredList = FilterAdminMediaList(s.MediaList, s.SearchInput.Value())
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
					s.FlatList = FlattenAdminMediaTree(s.MediaTree)
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

		// Create new folder (n key)
		if km.Matches(key, config.ActionNew) && s.FocusIndex == 0 {
			node := s.selectedNode()
			if node != nil && node.Kind == MediaNodeFolder && !node.FolderID.IsZero() {
				// Create subfolder under selected folder
				return s, func() tea.Msg {
					return ShowCreateAdminMediaFolderDialogMsg{
						ParentID: types.NullableAdminMediaFolderID{ID: node.FolderID, Valid: true},
					}
				}
			}
			// Create root folder
			return s, func() tea.Msg {
				return ShowCreateAdminMediaFolderDialogMsg{
					ParentID: types.NullableAdminMediaFolderID{},
				}
			}
		}

		// Upload new media (open file picker) -- uses ActionNew when not on tree panel
		if km.Matches(key, config.ActionNew) && s.FocusIndex != 0 {
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
				return OpenFilePickerMsg{Purpose: FILEPICKER_ADMINMEDIA}
			}
		}

		// Rename folder (e key on a folder node)
		if km.Matches(key, config.ActionEdit) && s.FocusIndex == 0 {
			node := s.selectedNode()
			if node != nil && node.Kind == MediaNodeFolder && !node.FolderID.IsZero() {
				return s, func() tea.Msg {
					return ShowRenameAdminMediaFolderDialogMsg{
						FolderID:    node.FolderID,
						CurrentName: node.Label,
					}
				}
			}
		}

		// Delete selected item (d key)
		if km.Matches(key, config.ActionDelete) {
			node := s.selectedNode()
			if node == nil {
				return s, nil
			}
			if node.Kind == MediaNodeFolder && !node.FolderID.IsZero() {
				return s, func() tea.Msg {
					return ShowDeleteAdminMediaFolderDialogMsg{
						FolderID: node.FolderID,
						Name:     node.Label,
					}
				}
			}
			if node.Kind == MediaNodeFile && node.Media != nil {
				media := node.Media
				label := media.AdminMediaID.String()
				if media.DisplayName.Valid && media.DisplayName.String != "" {
					label = media.DisplayName.String
				} else if media.Name.Valid && media.Name.String != "" {
					label = media.Name.String
				}
				return s, ShowDeleteAdminMediaDialogCmd(media.AdminMediaID, label)
			}
		}

		// Move media to folder (m key on a file node)
		if km.Matches(key, config.ActionMove) && s.FocusIndex == 0 {
			media := s.selectedMedia()
			if media != nil {
				label := media.AdminMediaID.String()
				if media.DisplayName.Valid && media.DisplayName.String != "" {
					label = media.DisplayName.String
				} else if media.Name.Valid && media.Name.String != "" {
					label = media.Name.String
				}
				return s, func() tea.Msg {
					return ShowMoveAdminMediaToFolderDialogMsg{
						AdminMediaID: media.AdminMediaID,
						Label:        label,
					}
				}
			}
		}

		// Common keys (quit, back, cursor)
		newCursor, cmd, handled := HandleCommonKeys(key, km, s.Cursor, s.CursorMax)
		if handled {
			s.Cursor = newCursor
			return s, cmd
		}

	// Fetch request messages
	case AdminMediaFetchMsg:
		d := ctx.DB
		if d == nil {
			return s, func() tea.Msg { return FetchErrMsg{Error: fmt.Errorf("database not connected")} }
		}
		return s, func() tea.Msg {
			media, err := d.ListAdminMedia()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			folders, err := d.ListAdminMediaFolders()
			if err != nil {
				return FetchErrMsg{Error: err}
			}
			var mediaData []db.AdminMedia
			if media != nil {
				mediaData = *media
			}
			var folderData []db.AdminMediaFolder
			if folders != nil {
				folderData = *folders
			}
			return AdminMediaFetchResultsMsg{Data: mediaData, Folders: folderData}
		}
	case AdminMediaFetchResultsMsg:
		s.MediaList = msg.Data
		s.FolderList = msg.Folders
		s.FilteredList = FilterAdminMediaList(s.MediaList, s.SearchQuery)
		s.Cursor = 0
		s.rebuildTree()
		return s, LoadingStopCmd()

	// Data refresh (from CMS operations)
	case AdminMediaListSet:
		s.MediaList = msg.MediaList
		s.FolderList = msg.FolderList
		s.FilteredList = FilterAdminMediaList(s.MediaList, s.SearchQuery)
		s.rebuildTree()
		return s, nil
	}

	return s, nil
}

func (s *AdminMediaScreen) KeyHints(km config.KeyMap) []KeyHint {
	if s.Searching {
		return []KeyHint{
			{"type", "filter"},
			{"enter", "accept"},
			{"esc", "clear"},
		}
	}
	node := s.selectedNode()
	if node != nil && node.Kind == MediaNodeFolder {
		return []KeyHint{
			{km.HintString(config.ActionSearch), "search"},
			{km.HintString(config.ActionNew), "new folder"},
			{km.HintString(config.ActionEdit), "rename"},
			{km.HintString(config.ActionDelete), "del folder"},
			{"enter", "expand"},
			{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
			{km.HintString(config.ActionNextPanel), "panel"},
			{km.HintString(config.ActionBack), "back"},
		}
	}
	return []KeyHint{
		{km.HintString(config.ActionSearch), "search"},
		{km.HintString(config.ActionNew), "new folder"},
		{km.HintString(config.ActionMove), "move"},
		{km.HintString(config.ActionDelete), "del"},
		{km.HintString(config.ActionUp) + "/" + km.HintString(config.ActionDown), "nav"},
		{km.HintString(config.ActionNextPanel), "panel"},
		{km.HintString(config.ActionBack), "back"},
	}
}
