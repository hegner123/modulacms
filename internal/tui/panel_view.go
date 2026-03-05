package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
)

// HeaderMode determines how the header is rendered based on terminal height.
type HeaderMode int

const (
	HeaderFull    HeaderMode = iota // ASCII art (tall terminals)
	HeaderCompact                   // Single line
	HeaderHidden                    // No header
)

// headerMode selects a header rendering mode based on available terminal height.
func (m Model) headerMode() HeaderMode {
	switch {
	case m.Height < 24:
		return HeaderHidden
	case m.Height < 40:
		return HeaderCompact
	default:
		return HeaderFull
	}
}

// BreadcrumbInfo holds contextual information displayed alongside the header.
type BreadcrumbInfo struct {
	PageName     string
	ItemCount    int
	SelectedItem string
	DBBadge      string
}

// breadcrumb builds a BreadcrumbInfo from the current model state.
func (m Model) breadcrumb() BreadcrumbInfo {
	info := BreadcrumbInfo{
		PageName: m.Page.Label,
	}
	if m.Config != nil {
		info.DBBadge = string(m.Config.Db_Driver)
	}
	return info
}

// renderBreadcrumbLine formats the breadcrumb as a single styled line.
// Empty or zero-valued fields are omitted.
func renderBreadcrumbLine(info BreadcrumbInfo) string {
	parts := []string{}
	if info.PageName != "" {
		parts = append(parts, info.PageName)
	}
	if info.ItemCount > 0 {
		parts = append(parts, fmt.Sprintf("%d items", info.ItemCount))
	}
	if info.SelectedItem != "" {
		parts = append(parts, info.SelectedItem)
	}

	line := strings.Join(parts, " | ")

	if info.DBBadge != "" {
		if line != "" {
			line += "  "
		}
		line += "[" + info.DBBadge + "]"
	}

	return line
}

// renderGutterStrip renders a minimal collapsed panel showing only the
// first character of the title vertically.
func renderGutterStrip(title string, height int, focused bool) string {
	borderColor := config.DefaultStyle.Tertiary
	if focused {
		borderColor = config.DefaultStyle.Accent
	}

	// Use the first character of the title as the gutter label.
	ch := " "
	if len(title) > 0 {
		ch = string([]rune(title)[0])
	}

	charStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(config.DefaultStyle.Accent)

	// Build vertical content: single character centered, rest empty.
	innerHeight := height - 3 // border top/bottom + title row
	if innerHeight < 1 {
		innerHeight = 1
	}
	lines := make([]string, innerHeight)
	mid := innerHeight / 2
	for i := range lines {
		if i == mid {
			lines[i] = charStyle.Render(ch)
		} else {
			lines[i] = " "
		}
	}

	// innerWidth is gutter width (4) minus border (2 chars)
	innerWidth := 2
	if innerWidth < 1 {
		innerWidth = 1
	}

	body := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(innerWidth).
		Height(innerHeight).
		Render(body)

	return box
}

// renderCMSPanelLayout renders the full panel layout: header, screen content, status bar.
func renderCMSPanelLayout(m Model) string {
	header := renderCMSHeader(m)
	statusBar := renderCMSPanelStatusBar(m)

	headerH := lipgloss.Height(header)
	statusH := lipgloss.Height(statusBar)
	bodyH := m.Height - headerH - statusH
	if bodyH < 1 {
		bodyH = 1
	}

	ctx := m.AppCtx()
	ctx.Height = bodyH
	screenContent := m.ActiveScreen.View(ctx)

	body := lipgloss.NewStyle().
		Width(m.Width).
		Height(bodyH).
		Render(screenContent)

	ui := lipgloss.JoinVertical(lipgloss.Left, header, body, statusBar)

	if m.FilePickerActive {
		return FilePickerOverlay(ui, m.FilePicker, m.Width, m.Height)
	}
	if m.ActiveOverlay != nil {
		return RenderOverlay(ui, m.ActiveOverlay, m.Width, m.Height)
	}
	return ui
}

// renderCMSHeader renders the top bar with the app title.
// The rendering adapts to terminal height via headerMode():
//   - HeaderFull: ASCII art title + breadcrumb line
//   - HeaderCompact: single styled line with version, page, and DB badge
//   - HeaderHidden: empty string (no header)
func renderCMSHeader(m Model) string {
	switch m.headerMode() {
	case HeaderHidden:
		return ""

	case HeaderCompact:
		info := m.breadcrumb()
		compactStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(config.DefaultStyle.Primary)

		parts := []string{"ModulaCMS v" + utility.Version}
		if info.PageName != "" {
			parts = append(parts, info.PageName)
		}
		if info.DBBadge != "" {
			parts = append(parts, info.DBBadge)
		}
		line := compactStyle.Render(strings.Join(parts, " | "))

		return lipgloss.NewStyle().
			Width(m.Width).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(config.DefaultStyle.Tertiary).
			Padding(0, 1).
			Render(line)

	default: // HeaderFull
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(config.DefaultStyle.Accent)

		title := titleStyle.Render(RenderTitle(m.Titles[m.TitleFont]))

		info := m.breadcrumb()
		crumbLine := renderBreadcrumbLine(info)
		var content string
		if crumbLine != "" {
			crumbStyle := lipgloss.NewStyle().
				Foreground(config.DefaultStyle.Secondary)
			content = title + "\n" + crumbStyle.Render(crumbLine)
		} else {
			content = title
		}

		return lipgloss.NewStyle().
			Width(m.Width).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(config.DefaultStyle.Tertiary).
			Padding(0, 1).
			Render(content)
	}
}

// statusBarHeight returns how many lines the statusbar should use based on
// terminal height: 2 lines for tall terminals, 1 for medium, 0 for tiny.
func (m Model) statusBarHeight() int {
	switch {
	case m.Height < 20:
		return 0
	case m.Height < 30:
		return 1
	default:
		return 2
	}
}

// renderCMSPanelStatusBar renders an adaptive status bar for the CMS layout.
//   - Height >= 30: 2-line bar (status badges + key hints)
//   - Height 20-29: 1-line condensed bar (6 prioritized key hints + badges)
//   - Height < 20:  hidden (empty string)
func renderCMSPanelStatusBar(m Model) string {
	switch m.statusBarHeight() {
	case 0:
		return ""
	case 1:
		return renderCondensedStatusBar(m)
	default:
		return renderFullStatusBar(m)
	}
}

// statusBarStyles returns the shared styles used by both status bar variants.
func statusBarStyles() (barStyle, keyStyle lipgloss.Style) {
	barFG := config.DefaultStyle.Primary
	barBG := config.DefaultStyle.PrimaryBG
	barStyle = lipgloss.NewStyle().Foreground(barFG).Background(barBG)
	keyStyle = barStyle.Bold(true)
	return
}

// padStatusLine fills a line to exactly the given width so the background
// color spans the full terminal width.
func padStatusLine(content string, width int, barStyle lipgloss.Style) string {
	w := lipgloss.Width(content)
	if w >= width {
		return content
	}
	return content + barStyle.Render(strings.Repeat(" ", width-w))
}

// collectKeyHints returns the key hints for the current page/screen.
func collectKeyHints(m Model) []KeyHint {
	if m.ActiveScreen != nil {
		if hinter, ok := m.ActiveScreen.(KeyHinter); ok {
			return hinter.KeyHints(m.Config.KeyBindings)
		}
	}
	return nil
}

// renderFullStatusBar renders the original 2-line status bar.
func renderFullStatusBar(m Model) string {
	barStyle, keyStyle := statusBarStyles()
	barBG := config.DefaultStyle.PrimaryBG

	padLine := func(content string) string {
		return padStatusLine(content, m.Width, barStyle)
	}

	// --- line 1: status badge + admin/client + mode + locale ---
	statusBadge := m.GetStatus()
	activeAccent := m.activeAccent()

	badgeStyle := lipgloss.NewStyle().
		Foreground(activeAccent).
		Background(barBG).
		Bold(true).
		Padding(0, 1)
	gap := barStyle.Render("  ")

	cmsLabel := "Client"
	if m.AdminMode {
		cmsLabel = "Admin"
	}
	cmsBadge := gap + badgeStyle.Render("["+cmsLabel+"]")
	modeBadge := gap + badgeStyle.Render("["+m.ScreenMode.String()+"]")

	var localeBadge string
	if m.Config.I18nEnabled() && m.ActiveLocale != "" {
		localeBadge = gap + badgeStyle.Render(strings.ToUpper(m.ActiveLocale))
	}

	var remoteBadge string
	if m.IsRemote && m.RemoteURL != "" {
		connStatus := m.GetRemoteStatus()
		remoteColor := lipgloss.Color("#7dc4e4")
		remoteLabel := "[remote: " + m.RemoteURL + "]"
		if connStatus == "disconnected" {
			remoteColor = lipgloss.Color("#ed8796")
			remoteLabel = "[remote: disconnected]"
		}
		remoteStyle := lipgloss.NewStyle().
			Foreground(remoteColor).
			Background(barBG).
			Bold(true).
			Padding(0, 1)
		remoteBadge = gap + remoteStyle.Render(remoteLabel)
	}

	var accordionBadge string
	if m.AccordionEnabled {
		accordionBadge = gap + badgeStyle.Render("[Accordion]")
	}

	line1 := statusBadge + cmsBadge + modeBadge + accordionBadge + localeBadge + remoteBadge

	// --- line 2: context-sensitive key hints ---
	hints := collectKeyHints(m)
	key := func(k, label string) string {
		return keyStyle.Render(k) + barStyle.Render(":"+label)
	}
	sep := barStyle.Render(" │ ")

	styledParts := make([]string, 0, len(hints))
	for _, h := range hints {
		styledParts = append(styledParts, key(h.Key, h.Label))
	}

	line2 := barStyle.Render(" ") + strings.Join(styledParts, sep)

	return padLine(line1) + "\n" + padLine(line2)
}

// renderCondensedStatusBar renders a single-line status bar with up to 6
// prioritized key hints plus compact badges, right-aligned.
func renderCondensedStatusBar(m Model) string {
	barStyle, keyStyle := statusBarStyles()
	barBG := config.DefaultStyle.PrimaryBG

	padLine := func(content string) string {
		return padStatusLine(content, m.Width, barStyle)
	}

	// Collect and limit hints to 6.
	hints := collectKeyHints(m)
	if len(hints) > 6 {
		hints = hints[:6]
	}

	key := func(k, label string) string {
		return keyStyle.Render(k) + barStyle.Render(":"+label)
	}

	styledParts := make([]string, 0, len(hints))
	for _, h := range hints {
		styledParts = append(styledParts, key(h.Key, h.Label))
	}
	hintStr := barStyle.Render(" ") + strings.Join(styledParts, barStyle.Render(" "))

	// Compact badges: [Page] [Mode] [locale]
	accentBadge := lipgloss.NewStyle().
		Foreground(m.activeAccent()).
		Background(barBG).
		Bold(true)

	cmsLabelC := "Client"
	if m.AdminMode {
		cmsLabelC = "Admin"
	}

	var badges []string
	badges = append(badges, accentBadge.Render("["+cmsLabelC+"]"))
	badges = append(badges, accentBadge.Render("["+m.Page.Label+"]"))
	badges = append(badges, accentBadge.Render("["+m.ScreenMode.String()+"]"))
	if m.Config.I18nEnabled() && m.ActiveLocale != "" {
		badges = append(badges, accentBadge.Render("["+strings.ToUpper(m.ActiveLocale)+"]"))
	}
	badgeStr := strings.Join(badges, barStyle.Render(" "))

	// Right-align badges: fill the gap between hints and badges.
	hintW := lipgloss.Width(hintStr)
	badgeW := lipgloss.Width(badgeStr)
	gap := m.Width - hintW - badgeW - 1 // -1 for trailing space
	if gap < 1 {
		gap = 1
	}
	line := hintStr + barStyle.Render(strings.Repeat(" ", gap)) + badgeStr

	return padLine(line)
}
