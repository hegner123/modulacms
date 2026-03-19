package tui

import (
	"strings"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db/types"
)

// DialogAction represents the type of action this dialog performs
type DialogAction string

// Dialog action types.
const (
	DIALOGGENERIC              DialogAction = "generic"
	DIALOGDELETE               DialogAction = "delete"
	DIALOGACTIONCONFIRM        DialogAction = "action_confirm"
	DIALOGINITCONTENT          DialogAction = "init_content"
	DIALOGQUITCONFIRM          DialogAction = "quit_confirm"
	DIALOGDELETECONTENT        DialogAction = "delete_content"
	DIALOGDELETEDATATYPE       DialogAction = "delete_datatype"
	DIALOGDELETEFIELD          DialogAction = "delete_field"
	DIALOGDELETEROUTE          DialogAction = "delete_route"
	DIALOGDELETEMEDIA          DialogAction = "delete_media"
	DIALOGDELETEUSER           DialogAction = "delete_user"
	DIALOGDELETECONTENTFIELD   DialogAction = "delete_content_field"
	DIALOGDELETEADMINROUTE     DialogAction = "delete_admin_route"
	DIALOGDELETEADMINDATATYPE  DialogAction = "delete_admin_datatype"
	DIALOGDELETEADMINFIELD     DialogAction = "delete_admin_field"
	DIALOGBACKUPRESTORE        DialogAction = "backup_restore"
	DIALOGAPPROVEPLUGINROUTES  DialogAction = "approve_plugin_routes"
	DIALOGAPPROVEPLUGINSHOOKS  DialogAction = "approve_plugin_hooks"
	DIALOGQUICKSTART           DialogAction = "quickstart_confirm"
	DIALOGDELETEFIELDTYPE      DialogAction = "delete_field_type"
	DIALOGDELETEADMINFIELDTYPE DialogAction = "delete_admin_field_type"
	DIALOGDEPLOYPULL           DialogAction = "deploy_pull"
	DIALOGDEPLOYPUSH           DialogAction = "deploy_push"
	DIALOGPUBLISHCONTENT       DialogAction = "publish_content"
	DIALOGUNPUBLISHCONTENT     DialogAction = "unpublish_content"
	DIALOGRESTOREVERSION       DialogAction = "restore_version"
	DIALOGLOCALESELECT         DialogAction = "locale_select"
	DIALOGDELETEWEBHOOK        DialogAction = "delete_webhook"
	DIALOGPLUGINCONFIRM        DialogAction = "plugin_confirm"
	DIALOGDELETEMEDIAFOLDER    DialogAction = "delete_media_folder"
	DIALOGDELETEVALIDATION     DialogAction = "delete_validation"
	DIALOGDELETEADMINVALIDATION DialogAction = "delete_admin_validation"
	DIALOGDELETETOKEN          DialogAction = "delete_token"
)

// dialogBorderPadding accounts for border and padding in dialog width calculations.
const dialogBorderPadding = 6

// DialogModel represents a dialog that can be rendered on top of other content
type DialogModel struct {
	Title       string
	Message     string
	Width       int
	Height      int
	OkText      string
	CancelText  string
	ShowCancel  bool
	ReadyOK     bool
	Action      DialogAction
	ActionIndex int // Stores the pending action index for DIALOGACTIONCONFIRM
	Locales     []string
	help        help.Model
	borderStyle lipgloss.Style
	titleStyle  lipgloss.Style
	textStyle   lipgloss.Style
	buttonStyle lipgloss.Style
	focusIndex  int

	// Scroll state for dialogs whose message exceeds terminal height
	scroll ScrollState
}

// NewDialog creates a new dialog model with the given parameters
func NewDialog(title, message string, showCancel bool, action DialogAction) DialogModel {
	h := help.New()
	h.ShowAll = false

	return DialogModel{
		Title:       title,
		Message:     message,
		Width:       50,
		Height:      10,
		OkText:      "OK",
		CancelText:  "Cancel",
		ShowCancel:  showCancel,
		ReadyOK:     false,
		Action:      action,
		help:        h,
		borderStyle: lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(1, 2),
		titleStyle:  lipgloss.NewStyle().Bold(true).Foreground(config.DefaultStyle.Accent),
		textStyle:   lipgloss.NewStyle().PaddingTop(1).PaddingBottom(1),
		buttonStyle: lipgloss.NewStyle().Padding(0, 2),
		focusIndex:  0,
	}
}

// SetSize sets the dialog size
// SetSize sets the dialog size.
func (d *DialogModel) SetSize(width, height int) {
	d.Width = width
	d.Height = height
}

// SetButtons sets the dialog button text
// SetButtons sets the dialog button text.
func (d *DialogModel) SetButtons(okText, cancelText string) {
	d.OkText = okText
	d.CancelText = cancelText
}

// Update handles user input for the dialog
// Update handles user input for the dialog and returns the updated model and command.
func (d *DialogModel) Update(msg tea.Msg) (DialogModel, tea.Cmd) {
	switch d.Action {
	case DIALOGDELETE, DIALOGACTIONCONFIRM, DIALOGINITCONTENT, DIALOGQUITCONFIRM, DIALOGDELETECONTENT,
		DIALOGDELETEDATATYPE, DIALOGDELETEFIELD, DIALOGDELETEROUTE, DIALOGDELETEMEDIA, DIALOGDELETEUSER,
		DIALOGDELETECONTENTFIELD, DIALOGDELETEADMINROUTE, DIALOGDELETEADMINDATATYPE, DIALOGDELETEADMINFIELD,
		DIALOGBACKUPRESTORE, DIALOGAPPROVEPLUGINROUTES, DIALOGAPPROVEPLUGINSHOOKS,
		DIALOGQUICKSTART, DIALOGDELETEFIELDTYPE, DIALOGDELETEADMINFIELDTYPE,
		DIALOGDEPLOYPULL, DIALOGDEPLOYPUSH,
		DIALOGPUBLISHCONTENT, DIALOGUNPUBLISHCONTENT, DIALOGRESTOREVERSION,
		DIALOGDELETEWEBHOOK, DIALOGDELETEMEDIAFOLDER,
		DIALOGDELETEVALIDATION, DIALOGDELETEADMINVALIDATION,
		DIALOGDELETETOKEN:
		return d.ToggleControls(msg)
	case DIALOGLOCALESELECT:
		return d.LocaleSelectControls(msg)
	case DIALOGGENERIC:
		// Generic dialog dismisses on enter or esc
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "enter", "esc":
				return *d, func() tea.Msg { return DialogCancelMsg{} }
			}
		}
		return *d, nil
	default:
		return *d, nil
	}
}

// OverlayUpdate implements ModalOverlay for DialogModel.
func (d *DialogModel) OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

// OverlayView implements ModalOverlay for DialogModel.
func (d *DialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

// ToggleControls handles navigation and selection within the dialog.
func (d *DialogModel) ToggleControls(msg tea.Msg) (DialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab", "shift+tab", "h", "l", "j", "k", "left", "right", "up", "down":
			if d.ShowCancel {
				d.focusIndex = (d.focusIndex + 1) % 2
			}
			return *d, nil
		case "enter":
			if d.focusIndex == 0 {
				return *d, func() tea.Msg { return DialogAcceptMsg{Action: d.Action} }
			}
			return *d, func() tea.Msg { return DialogCancelMsg{} }
		case "esc":
			return *d, func() tea.Msg { return DialogCancelMsg{} }
		}
	}
	return *d, nil

}

// Render renders the dialog with its title, message, and buttons.
// Uses a pointer receiver so scrollableBody can persist offset changes.
func (d *DialogModel) Render(windowWidth, windowHeight int) string {
	contentWidth := d.Width
	innerW := contentWidth - dialogBorderPadding

	// --- Header ---
	header := d.titleStyle.Render(d.Title)

	// --- Message lines as scroll items ---
	messageText := d.textStyle.Width(innerW).Render(d.Message)
	messageLines := strings.Split(messageText, "\n")
	// Wrap each line as a single-item for scrollableBody
	items := make([]string, len(messageLines))
	for i, line := range messageLines {
		items[i] = line
	}

	// --- Footer (buttons) ---
	okButton := d.buttonStyle
	cancelButton := d.buttonStyle

	if d.focusIndex == 0 {
		okButton = okButton.
			Background(config.DefaultStyle.Accent).
			Foreground(config.DefaultStyle.Primary)
	} else {
		cancelButton = cancelButton.
			Background(config.DefaultStyle.Accent).
			Foreground(config.DefaultStyle.Primary)
	}

	okButtonText := okButton.Render(buttonLabel(d.OkText, d.focusIndex == 0))
	cancelButtonText := cancelButton.Render(buttonLabel(d.CancelText, d.focusIndex == 1))

	var footer string
	if d.ShowCancel {
		footer = lipgloss.JoinHorizontal(lipgloss.Center, okButtonText, "  ", cancelButtonText)
	} else {
		footer = lipgloss.JoinHorizontal(lipgloss.Center, okButtonText)
	}

	// --- Compute available body lines ---
	borderOverhead := 4
	headerH := lipgloss.Height(header) + 1
	footerH := lipgloss.Height(footer) + 1
	indicatorH := 2
	maxDialogH := windowHeight - 4
	maxBodyLines := maxDialogH - borderOverhead - headerH - footerH - indicatorH
	if maxBodyLines < 3 {
		maxBodyLines = 3
	}

	// Use ActionIndex as focus hint (tracks cursor in locale select; 0 for others → shows from top)
	visibleBody, topClip, bottomClip := d.scroll.scrollableBody(items, d.ActionIndex, maxBodyLines)

	// --- Assemble ---
	var content strings.Builder
	content.WriteString(header)
	content.WriteString("\n")

	if topClip {
		content.WriteString(scrollUpIndicator(innerW))
		content.WriteString("\n")
	}

	content.WriteString(visibleBody)
	content.WriteString("\n")

	if bottomClip {
		content.WriteString(scrollDownIndicator(innerW))
		content.WriteString("\n")
	}

	content.WriteString(footer)

	return d.borderStyle.Width(contentWidth).Render(content.String())
}

// Dialog-related messages

// DialogAcceptMsg signals that a dialog was accepted.
type DialogAcceptMsg struct {
	Action DialogAction
}

// DialogCancelMsg signals that a dialog was cancelled.
type DialogCancelMsg struct{}

// DialogReadyOK signals that a dialog is ready to accept input.
type DialogReadyOK struct{}

// ShowDialogMsg triggers showing a dialog.
type ShowDialogMsg struct {
	Title      string
	Message    string
	ShowCancel bool
}

// HandleShowDialog creates a command to show a generic dialog.
func HandleShowDialog(title, message string, showCancel bool) tea.Cmd {
	return func() tea.Msg {
		return ShowDialogMsg{
			Title:      title,
			Message:    message,
			ShowCancel: showCancel,
		}
	}
}

// ShowQuitConfirmDialogMsg triggers showing a quit confirmation dialog.
type ShowQuitConfirmDialogMsg struct{}

// ShowQuitConfirmDialogCmd creates a command to show a quit confirmation dialog.
func ShowQuitConfirmDialogCmd() tea.Cmd {
	return func() tea.Msg {
		return ShowQuitConfirmDialogMsg{}
	}
}

// ShowDeleteContentDialogMsg triggers showing a delete content confirmation dialog.
type ShowDeleteContentDialogMsg struct {
	ContentID   string
	ContentName string
	HasChildren bool
}

// ShowDeleteContentDialogCmd creates a command to show a delete content confirmation dialog.
func ShowDeleteContentDialogCmd(contentID, contentName string, hasChildren bool) tea.Cmd {
	return func() tea.Msg {
		return ShowDeleteContentDialogMsg{
			ContentID:   contentID,
			ContentName: contentName,
			HasChildren: hasChildren,
		}
	}
}

// DeleteContentRequestMsg triggers content deletion.
type DeleteContentRequestMsg struct {
	ContentID string
	RouteID   string
}

// DeleteContentCmd creates a command to delete content.
func DeleteContentCmd(contentID, routeID string) tea.Cmd {
	return func() tea.Msg {
		return DeleteContentRequestMsg{
			ContentID: contentID,
			RouteID:   routeID,
		}
	}
}

// ContentDeletedMsg is sent after content is successfully deleted.
type ContentDeletedMsg struct {
	ContentID types.ContentID
	RouteID   types.RouteID
	AdminMode bool
}

// filePickerChrome is the number of lines consumed by the overlay frame
// around the filepicker content: border (2) + padding (2) + title (1) +
// gap after title (1) + gap before hint (1) + hint (1) = 8.
const filePickerChrome = 8

// filePickerHeight returns the filepicker row count that fits inside the
// overlay for a terminal of the given height.
func filePickerHeight(termHeight int) int {
	h := termHeight - filePickerChrome
	if h < 4 {
		h = 4
	}
	return h
}

// FilePickerOverlay renders a file picker as a full-screen overlay.
func FilePickerOverlay(base string, fp filepicker.Model, width, height int) string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(config.DefaultStyle.Accent).
		Render("Select a file to upload")

	hint := lipgloss.NewStyle().
		Foreground(config.DefaultStyle.Secondary).
		Render("esc: cancel")

	// Content height = filepicker rows + title (1) + gaps (2) + hint (1).
	contentHeight := filePickerHeight(height) + 4

	pickerView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(config.DefaultStyle.Accent).
		Padding(1, 2).
		Width(width - 4).
		Height(contentHeight).
		Render(title + "\n\n" + fp.View() + "\n\n" + hint)

	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		pickerView,
		lipgloss.WithWhitespaceChars(" "),
	)
}

// LocaleSelectControls handles navigation within the locale selection dialog.
// Up/down cycles through locale options (tracked via ActionIndex),
// enter selects the current locale, esc cancels.
func (d *DialogModel) LocaleSelectControls(msg tea.Msg) (DialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if d.ActionIndex > 0 {
				d.ActionIndex--
				d.Message = buildLocaleDialogMessage(d.Locales, d.ActionIndex)
			}
			return *d, nil
		case "down", "j":
			if d.ActionIndex < len(d.Locales)-1 {
				d.ActionIndex++
				d.Message = buildLocaleDialogMessage(d.Locales, d.ActionIndex)
			}
			return *d, nil
		case "enter":
			return *d, func() tea.Msg { return DialogAcceptMsg{Action: d.Action} }
		case "esc":
			return *d, func() tea.Msg { return DialogCancelMsg{} }
		}
	}
	return *d, nil
}

// buildLocaleDialogMessage renders the locale list with a cursor indicator.
func buildLocaleDialogMessage(locales []string, cursor int) string {
	var b strings.Builder
	for i, loc := range locales {
		if i == cursor {
			b.WriteString("> " + loc + "\n")
		} else {
			b.WriteString("  " + loc + "\n")
		}
	}
	return b.String()
}

// ShowPublishDialogMsg triggers showing a publish/unpublish confirmation dialog.
type ShowPublishDialogMsg struct {
	ContentID   types.ContentID
	RouteID     types.RouteID
	ContentName string
	IsPublished bool
}

// ShowPublishDialogCmd creates a command to show the publish confirmation dialog.
func ShowPublishDialogCmd(contentID types.ContentID, routeID types.RouteID, contentName string, isPublished bool) tea.Cmd {
	return func() tea.Msg {
		return ShowPublishDialogMsg{
			ContentID:   contentID,
			RouteID:     routeID,
			ContentName: contentName,
			IsPublished: isPublished,
		}
	}
}
