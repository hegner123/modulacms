package config

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/compat"
)

// Color constants for the TUI palette. Bubble Tea v2 handles color
// downsampling automatically, so we specify only the truecolor hex value.
var (
	White       = lipgloss.Color("#FFFFFF")
	LightGray   = lipgloss.Color("#c0c0c0")
	Gray        = lipgloss.Color("#808080")
	Black       = lipgloss.Color("#000000")
	Purple      = lipgloss.Color("#6612e3")
	LightPurple = lipgloss.Color("#8347de")
	Emerald     = lipgloss.Color("#00CC66")
	Rose        = lipgloss.Color("#D90368")
	Yellow      = lipgloss.Color("#F1C40F")
	Orange      = lipgloss.Color("#F75C03")
	Blue        = lipgloss.Color("#5f5fff")
	Amber       = lipgloss.Color("#FFBF00")

	DefaultStyle Color = Color{
		Primary:       compat.AdaptiveColor{Light: Black, Dark: White},
		PrimaryBG:     compat.AdaptiveColor{Light: White, Dark: Black},
		Secondary:     compat.AdaptiveColor{Light: Gray, Dark: LightGray},
		SecondaryBG:   compat.AdaptiveColor{Light: White, Dark: Black},
		Tertiary:      compat.AdaptiveColor{Light: LightGray, Dark: Gray},
		TertiaryBG:    compat.AdaptiveColor{Light: Gray, Dark: Black},
		Accent:        compat.AdaptiveColor{Light: Blue, Dark: Blue},
		AccentBG:      compat.AdaptiveColor{Light: White, Dark: Blue},
		Accent2:       compat.AdaptiveColor{Light: Rose, Dark: Rose},
		Accent2BG:     compat.AdaptiveColor{Light: White, Dark: Black},
		Active:        compat.AdaptiveColor{Light: Black, Dark: Black},
		ActiveBG:      compat.AdaptiveColor{Light: Gray, Dark: LightGray},
		Status1:       compat.AdaptiveColor{Light: Black, Dark: White},
		Status1BG:     compat.AdaptiveColor{Light: LightGray, Dark: Black},
		Status2:       compat.AdaptiveColor{Light: Gray, Dark: Black},
		Status2BG:     compat.AdaptiveColor{Light: Black, Dark: Gray},
		Status3:       compat.AdaptiveColor{Light: LightPurple, Dark: LightPurple},
		Status3BG:     compat.AdaptiveColor{Light: Black, Dark: Black},
		PrimaryBorder: compat.AdaptiveColor{Light: Purple, Dark: Purple},
		AdminAccent:   compat.AdaptiveColor{Light: Amber, Dark: Amber},
		Warn:          compat.AdaptiveColor{Light: Orange, Dark: Orange},
		WarnBG:        compat.AdaptiveColor{Light: White, Dark: White},
	}
)

// Color defines the UI color palette for light and dark themes.
type Color struct {
	Primary       color.Color `json:"primary"`
	PrimaryBG     color.Color `json:"primary_background"`
	Secondary     color.Color `json:"secondary"`
	SecondaryBG   color.Color `json:"secondary_background"`
	Tertiary      color.Color `json:"tertiary"`
	TertiaryBG    color.Color `json:"tertiary_background"`
	Accent        color.Color `json:"accent"`
	AccentBG      color.Color `json:"accent_background"`
	Accent2       color.Color `json:"accent2"`
	Accent2BG     color.Color `json:"accent2_background"`
	Active        color.Color `json:"active"`
	ActiveBG      color.Color `json:"active_background"`
	Status1       color.Color `json:"status_1"`
	Status1BG     color.Color `json:"status_1_background"`
	Status2       color.Color `json:"status_2"`
	Status2BG     color.Color `json:"status_2_background"`
	Status3       color.Color `json:"status_3"`
	Status3BG     color.Color `json:"status_3_background"`
	PrimaryBorder color.Color `json:"primary_border"`
	AdminAccent   color.Color `json:"admin_accent"`
	Warn          color.Color `json:"warn"`
	WarnBG        color.Color `json:"warn_background"`
}

// TODO UI config loading
// TODO UI config default
