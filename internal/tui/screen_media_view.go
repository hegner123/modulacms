package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// View renders the MediaScreen using the grid layout system.
func (s *MediaScreen) View(ctx AppContext) string {
	innerH := PanelInnerHeight(ctx.Height)
	treeTotal := len(s.FlatList)

	cells := []CellContent{
		{
			Content:      s.renderMediaTree(),
			TotalLines:   treeTotal,
			ScrollOffset: ClampScroll(s.Cursor, treeTotal, innerH),
		},
		{Content: s.renderMediaSummary()},
		{Content: s.renderMediaMetadata()},
	}
	return s.RenderGrid(ctx, cells)
}

// renderMediaTree renders the folder/file tree with cursor and search input.
func (s *MediaScreen) renderMediaTree() string {
	if len(s.FlatList) == 0 && !s.Searching && s.SearchQuery == "" {
		return " (no media)"
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
			arrow := "▸"
			if node.Expand {
				arrow = "▾"
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

// renderMediaSummary renders identity fields for the selected media item.
func (s *MediaScreen) renderMediaSummary() string {
	media := s.selectedMedia()
	if media == nil {
		return " No media selected"
	}

	accent := lipgloss.NewStyle().Bold(true)

	name := media.MediaID.String()
	if media.DisplayName.Valid && media.DisplayName.String != "" {
		name = media.DisplayName.String
	} else if media.Name.Valid && media.Name.String != "" {
		name = media.Name.String
	}

	lines := []string{
		accent.Render(" " + name),
		"",
		fmt.Sprintf(" Name      %s", mediaNullStr(media.Name)),
		fmt.Sprintf(" Display   %s", mediaNullStr(media.DisplayName)),
		fmt.Sprintf(" Mimetype  %s", mediaNullStr(media.Mimetype)),
		fmt.Sprintf(" Size      %s", mediaNullStr(media.Dimensions)),
		"",
		fmt.Sprintf(" URL       %s", string(media.URL)),
	}

	if media.Srcset.Valid && media.Srcset.String != "" {
		lines = append(lines, fmt.Sprintf(" Srcset    %s", media.Srcset.String))
	}

	return strings.Join(lines, "\n")
}

// renderMediaMetadata renders editorial/metadata fields for the selected media item.
func (s *MediaScreen) renderMediaMetadata() string {
	media := s.selectedMedia()
	if media == nil {
		return " No media selected"
	}

	faint := lipgloss.NewStyle().Faint(true)

	lines := []string{
		fmt.Sprintf(" Alt       %s", mediaNullStr(media.Alt)),
		fmt.Sprintf(" Caption   %s", mediaNullStr(media.Caption)),
		fmt.Sprintf(" Desc      %s", mediaNullStr(media.Description)),
		"",
		fmt.Sprintf(" Focal X   %s", mediaNullFloat(media.FocalX)),
		fmt.Sprintf(" Focal Y   %s", mediaNullFloat(media.FocalY)),
		fmt.Sprintf(" Class     %s", mediaNullStr(media.Class)),
		"",
		fmt.Sprintf(" Author    %s", mediaNullUserID(media.AuthorID)),
		fmt.Sprintf(" Created   %s", media.DateCreated.String()),
		fmt.Sprintf(" Modified  %s", media.DateModified.String()),
		"",
		faint.Render(fmt.Sprintf(" ID  %s", media.MediaID)),
	}

	return strings.Join(lines, "\n")
}

func mediaNullStr(ns db.NullString) string {
	if ns.Valid {
		if ns.String == "" {
			return "(empty)"
		}
		return ns.String
	}
	return "(none)"
}

func mediaNullFloat(nf types.NullableFloat64) string {
	if nf.Valid {
		return fmt.Sprintf("%.2f", nf.Float64)
	}
	return "(none)"
}

func mediaNullUserID(nid types.NullableUserID) string {
	if nid.Valid {
		return string(nid.ID)
	}
	return "(none)"
}
