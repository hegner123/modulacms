package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// MediaDimensionFormDialogModel represents a form dialog for media dimension CRUD.
type MediaDimensionFormDialogModel struct {
	dialogStyles

	Title    string
	Width    int
	Action   FormDialogAction
	EntityID string // MdID for edit

	LabelInput       textinput.Model
	WidthInput       textinput.Model
	HeightInput      textinput.Model
	AspectRatioInput textinput.Model

	// 0=Label, 1=Width, 2=Height, 3=AspectRatio, 4=Cancel, 5=Confirm
	focusIndex int
}

const (
	mdFocusLabel       = 0
	mdFocusWidth       = 1
	mdFocusHeight      = 2
	mdFocusAspect      = 3
	mdFocusCancel      = 4
	mdFocusConfirm     = 5
	mdMaxFocus         = 5
)

// NewMediaDimensionFormDialog creates a dialog for creating a new dimension.
func NewMediaDimensionFormDialog(title string) MediaDimensionFormDialogModel {
	labelInput := textinput.New()
	labelInput.Placeholder = "thumbnail"
	labelInput.CharLimit = 128
	labelInput.SetWidth(40)
	labelInput.Focus()

	widthInput := textinput.New()
	widthInput.Placeholder = "800"
	widthInput.CharLimit = 10
	widthInput.SetWidth(40)

	heightInput := textinput.New()
	heightInput.Placeholder = "600"
	heightInput.CharLimit = 10
	heightInput.SetWidth(40)

	aspectInput := textinput.New()
	aspectInput.Placeholder = "4:3 (optional)"
	aspectInput.CharLimit = 20
	aspectInput.SetWidth(40)

	return MediaDimensionFormDialogModel{
		dialogStyles:     newDialogStyles(),
		Title:            title,
		Width:            55,
		Action:           FORMDIALOGCREATEMEDIADIMENSION,
		LabelInput:       labelInput,
		WidthInput:       widthInput,
		HeightInput:      heightInput,
		AspectRatioInput: aspectInput,
		focusIndex:       mdFocusLabel,
	}
}

// NewEditMediaDimensionFormDialog creates a dialog pre-populated for editing.
func NewEditMediaDimensionFormDialog(title string, dim db.MediaDimensions) MediaDimensionFormDialogModel {
	labelInput := textinput.New()
	labelInput.Placeholder = "thumbnail"
	labelInput.CharLimit = 128
	labelInput.SetWidth(40)
	if dim.Label.Valid {
		labelInput.SetValue(dim.Label.String)
	}
	labelInput.Focus()

	widthInput := textinput.New()
	widthInput.Placeholder = "800"
	widthInput.CharLimit = 10
	widthInput.SetWidth(40)
	if dim.Width.Valid {
		widthInput.SetValue(dim.Width.String())
	}

	heightInput := textinput.New()
	heightInput.Placeholder = "600"
	heightInput.CharLimit = 10
	heightInput.SetWidth(40)
	if dim.Height.Valid {
		heightInput.SetValue(dim.Height.String())
	}

	aspectInput := textinput.New()
	aspectInput.Placeholder = "4:3 (optional)"
	aspectInput.CharLimit = 20
	aspectInput.SetWidth(40)
	if dim.AspectRatio.Valid {
		aspectInput.SetValue(dim.AspectRatio.String)
	}

	return MediaDimensionFormDialogModel{
		dialogStyles:     newDialogStyles(),
		Title:            title,
		Width:            55,
		Action:           FORMDIALOGEDITMEDIADIMENSION,
		EntityID:         dim.MdID,
		LabelInput:       labelInput,
		WidthInput:       widthInput,
		HeightInput:      heightInput,
		AspectRatioInput: aspectInput,
		focusIndex:       mdFocusLabel,
	}
}

func (d *MediaDimensionFormDialogModel) Update(msg tea.Msg) (MediaDimensionFormDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "tab", "down":
			d.mdFocusNext()
			return *d, nil
		case "shift+tab", "up":
			d.mdFocusPrev()
			return *d, nil
		case "left":
			if d.focusIndex == mdFocusConfirm {
				d.focusIndex = mdFocusCancel
				return *d, nil
			}
		case "right":
			if d.focusIndex == mdFocusCancel {
				d.focusIndex = mdFocusConfirm
				return *d, nil
			}
		case "enter":
			if d.focusIndex == mdFocusConfirm {
				return *d, func() tea.Msg {
					return MediaDimensionFormDialogAcceptMsg{
						Action:      d.Action,
						EntityID:    d.EntityID,
						Label:       d.LabelInput.Value(),
						Width:       d.WidthInput.Value(),
						Height:      d.HeightInput.Value(),
						AspectRatio: d.AspectRatioInput.Value(),
					}
				}
			}
			if d.focusIndex == mdFocusCancel {
				return *d, func() tea.Msg { return MediaDimensionFormDialogCancelMsg{} }
			}
			d.mdFocusNext()
			return *d, nil
		case "esc":
			return *d, func() tea.Msg { return MediaDimensionFormDialogCancelMsg{} }
		}
	}

	var cmd tea.Cmd
	switch d.focusIndex {
	case mdFocusLabel:
		d.LabelInput, cmd = d.LabelInput.Update(msg)
	case mdFocusWidth:
		d.WidthInput, cmd = d.WidthInput.Update(msg)
	case mdFocusHeight:
		d.HeightInput, cmd = d.HeightInput.Update(msg)
	case mdFocusAspect:
		d.AspectRatioInput, cmd = d.AspectRatioInput.Update(msg)
	}
	return *d, cmd
}

func (d *MediaDimensionFormDialogModel) mdFocusNext() {
	d.focusIndex = (d.focusIndex + 1) % (mdMaxFocus + 1)
	d.mdUpdateFocus()
}

func (d *MediaDimensionFormDialogModel) mdFocusPrev() {
	d.focusIndex = (d.focusIndex + mdMaxFocus) % (mdMaxFocus + 1)
	d.mdUpdateFocus()
}

func (d *MediaDimensionFormDialogModel) mdUpdateFocus() {
	d.LabelInput.Blur()
	d.WidthInput.Blur()
	d.HeightInput.Blur()
	d.AspectRatioInput.Blur()
	switch d.focusIndex {
	case mdFocusLabel:
		d.LabelInput.Focus()
	case mdFocusWidth:
		d.WidthInput.Focus()
	case mdFocusHeight:
		d.HeightInput.Focus()
	case mdFocusAspect:
		d.AspectRatioInput.Focus()
	}
}

func (d *MediaDimensionFormDialogModel) OverlayUpdate(msg tea.KeyPressMsg) (ModalOverlay, tea.Cmd) {
	updated, cmd := d.Update(msg)
	return &updated, cmd
}

func (d *MediaDimensionFormDialogModel) OverlayTick(msg tea.Msg) (ModalOverlay, tea.Cmd) {
	var cmd tea.Cmd
	switch d.focusIndex {
	case mdFocusLabel:
		d.LabelInput, cmd = d.LabelInput.Update(msg)
	case mdFocusWidth:
		d.WidthInput, cmd = d.WidthInput.Update(msg)
	case mdFocusHeight:
		d.HeightInput, cmd = d.HeightInput.Update(msg)
	case mdFocusAspect:
		d.AspectRatioInput, cmd = d.AspectRatioInput.Update(msg)
	}
	return d, cmd
}

func (d *MediaDimensionFormDialogModel) OverlayView(width, height int) string {
	return d.Render(width, height)
}

func (d MediaDimensionFormDialogModel) Render(windowWidth, windowHeight int) string {
	contentWidth := d.Width
	innerW := contentWidth - dialogBorderPadding

	var content strings.Builder
	content.WriteString(d.titleStyle.Render(d.Title))
	content.WriteString("\n\n")

	fields := []struct {
		label string
		input textinput.Model
		focus int
	}{
		{"Label", d.LabelInput, mdFocusLabel},
		{"Width (px)", d.WidthInput, mdFocusWidth},
		{"Height (px)", d.HeightInput, mdFocusHeight},
		{"Aspect Ratio", d.AspectRatioInput, mdFocusAspect},
	}

	for _, f := range fields {
		content.WriteString(d.labelStyle.Render(f.label))
		content.WriteString("\n")
		if d.focusIndex == f.focus {
			content.WriteString(d.focusedInputStyle.Width(innerW).Render(f.input.View()))
		} else {
			content.WriteString(d.inputStyle.Width(innerW).Render(f.input.View()))
		}
		content.WriteString("\n")
	}
	content.WriteString("\n")

	cancelStyle := d.cancelButtonStyle
	confirmStyle := d.confirmButtonStyle
	if d.focusIndex == mdFocusCancel {
		cancelStyle = cancelStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	if d.focusIndex == mdFocusConfirm {
		confirmStyle = confirmStyle.Background(config.DefaultStyle.Accent).Foreground(config.DefaultStyle.Primary)
	}
	cancelBtn := cancelStyle.Render(buttonLabel("Cancel", d.focusIndex == mdFocusCancel))
	confirmBtn := confirmStyle.Render(buttonLabel("Save", d.focusIndex == mdFocusConfirm))
	buttonBar := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "  ", confirmBtn)
	content.WriteString(buttonBar)

	return d.borderStyle.Width(contentWidth).Render(content.String())
}

// Messages

type MediaDimensionFormDialogAcceptMsg struct {
	Action      FormDialogAction
	EntityID    string
	Label       string
	Width       string
	Height      string
	AspectRatio string
}

type MediaDimensionFormDialogCancelMsg struct{}

type ShowMediaDimensionFormDialogMsg struct {
	Title string
}

type ShowEditMediaDimensionDialogMsg struct {
	Dimension db.MediaDimensions
}
