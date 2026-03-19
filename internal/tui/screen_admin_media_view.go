package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// View renders the AdminMediaScreen using the grid layout system.
func (s *AdminMediaScreen) View(ctx AppContext) string {
	innerH := PanelInnerHeight(ctx.Height)
	treeTotal := len(s.FlatList)

	cells := []CellContent{
		{
			Content:      s.renderAdminMediaTree(),
			TotalLines:   treeTotal,
			ScrollOffset: ClampScroll(s.Cursor, treeTotal, innerH),
		},
		{Content: s.renderAdminMediaSummary()},
		{Content: s.renderAdminMediaMetadata()},
	}
	return s.RenderGrid(ctx, cells)
}

// renderAdminMediaTree renders the folder/file tree with cursor and search input.
func (s *AdminMediaScreen) renderAdminMediaTree() string {
	if len(s.FlatList) == 0 && !s.Searching && s.SearchQuery == "" {
		return " (no admin media)"
	}

	faint := lipgloss.NewStyle().Faint(true)
	folderStyle := lipgloss.NewStyle().Bold(true)
	accentStyle := lipgloss.NewStyle().Foreground(config.DefaultStyle.Accent)

	var lines []string

	// Filter badge when not searching but query is active
	if s.SearchQuery != "" && !s.Searching {
		lines = append(lines, accentStyle.Render(fmt.Sprintf(" filter: %s", s.SearchQuery)))
		lines = append(lines, "")
	}

	if len(s.FlatList) == 0 {
		lines = append(lines, faint.Render(" (no matches)"))
	}

	for i, node := range s.FlatList {
		indent := strings.Repeat("  ", node.Depth)
		cursor := "  "
		if s.Cursor == i {
			cursor = "->"
		}

		var label string
		if node.Kind == MediaNodeFolder {
			arrow := ">"
			if node.Expand {
				arrow = "v"
			}
			label = folderStyle.Render(arrow + " " + node.Label + "/")
		} else {
			name := node.Label
			if node.Media != nil {
				if node.Media.DisplayName.Valid && node.Media.DisplayName.String != "" {
					name = node.Media.DisplayName.String
				} else if node.Media.Name.Valid && node.Media.Name.String != "" {
					name = node.Media.Name.String
				}
			}
			label = name
		}

		lines = append(lines, fmt.Sprintf(" %s %s%s", cursor, indent, label))
	}

	// Search input line at the bottom
	if s.Searching {
		lines = append(lines, "")
		lines = append(lines, " / "+s.SearchInput.View())
	}

	return strings.Join(lines, "\n")
}

// renderAdminMediaSummary renders identity fields for the selected admin media item.
func (s *AdminMediaScreen) renderAdminMediaSummary() string {
	media := s.selectedMedia()
	if media == nil {
		return " No admin media selected"
	}

	accent := lipgloss.NewStyle().Bold(true)

	name := media.AdminMediaID.String()
	if media.DisplayName.Valid && media.DisplayName.String != "" {
		name = media.DisplayName.String
	} else if media.Name.Valid && media.Name.String != "" {
		name = media.Name.String
	}

	lines := []string{
		accent.Render(" " + name),
		"",
		fmt.Sprintf(" Name      %s", adminMediaNullStr(media.Name)),
		fmt.Sprintf(" Display   %s", adminMediaNullStr(media.DisplayName)),
		fmt.Sprintf(" Mimetype  %s", adminMediaNullStr(media.Mimetype)),
		fmt.Sprintf(" Size      %s", adminMediaNullStr(media.Dimensions)),
		"",
		fmt.Sprintf(" URL       %s", string(media.URL)),
	}

	if media.Srcset.Valid && media.Srcset.String != "" {
		lines = append(lines, fmt.Sprintf(" Srcset    %s", media.Srcset.String))
	}

	// Show folder info if media is in a folder
	if media.FolderID.Valid && !media.FolderID.ID.IsZero() {
		folderName := s.folderNameByID(media.FolderID.ID)
		lines = append(lines, fmt.Sprintf(" Folder    %s", folderName))
	}

	return strings.Join(lines, "\n")
}

// renderAdminMediaMetadata renders editorial/metadata fields for the selected admin media item.
func (s *AdminMediaScreen) renderAdminMediaMetadata() string {
	media := s.selectedMedia()
	if media == nil {
		return " No admin media selected"
	}

	faint := lipgloss.NewStyle().Faint(true)

	lines := []string{
		fmt.Sprintf(" Alt       %s", adminMediaNullStr(media.Alt)),
		fmt.Sprintf(" Caption   %s", adminMediaNullStr(media.Caption)),
		fmt.Sprintf(" Desc      %s", adminMediaNullStr(media.Description)),
		"",
		fmt.Sprintf(" Focal X   %s", adminMediaNullFloat(media.FocalX)),
		fmt.Sprintf(" Focal Y   %s", adminMediaNullFloat(media.FocalY)),
		fmt.Sprintf(" Class     %s", adminMediaNullStr(media.Class)),
		fmt.Sprintf(" Folder    %s", adminMediaNullAdminFolderID(media.FolderID)),
		"",
		fmt.Sprintf(" Author    %s", mediaNullUserID(media.AuthorID)),
		fmt.Sprintf(" Created   %s", media.DateCreated.String()),
		fmt.Sprintf(" Modified  %s", media.DateModified.String()),
		"",
		faint.Render(fmt.Sprintf(" ID  %s", media.AdminMediaID)),
	}

	return strings.Join(lines, "\n")
}

func adminMediaNullStr(ns db.NullString) string {
	if ns.Valid {
		if ns.String == "" {
			return "(empty)"
		}
		return ns.String
	}
	return "(none)"
}

func adminMediaNullFloat(nf types.NullableFloat64) string {
	if nf.Valid {
		return fmt.Sprintf("%.2f", nf.Float64)
	}
	return "(none)"
}

func adminMediaNullAdminFolderID(nid types.NullableAdminMediaFolderID) string {
	if nid.Valid {
		return string(nid.ID)
	}
	return "null"
}

// folderNameByID looks up a folder name from the screen's FolderList.
func (s *AdminMediaScreen) folderNameByID(id types.AdminMediaFolderID) string {
	for _, f := range s.FolderList {
		if f.AdminFolderID == id {
			return f.Name
		}
	}
	return string(id)
}
