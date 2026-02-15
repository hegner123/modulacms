package config

import "github.com/charmbracelet/lipgloss"

// Color constants and default UI style for TUI components.
var (
	White       = lipgloss.CompleteColor{TrueColor: "#FFFFFF", ANSI256: "15", ANSI: "15"}
	LightGray   = lipgloss.CompleteColor{TrueColor: "#c0c0c0", ANSI256: "254", ANSI: "7"}
	Gray        = lipgloss.CompleteColor{TrueColor: "#808080", ANSI256: "250", ANSI: "8"}
	Black       = lipgloss.CompleteColor{TrueColor: "#000000", ANSI256: "0", ANSI: "0"}
	Purple      = lipgloss.CompleteColor{TrueColor: "#6612e3", ANSI256: "129", ANSI: "5"}
	LightPurple = lipgloss.CompleteColor{TrueColor: "#8347de", ANSI256: "98", ANSI: "13"}
	Emerald     = lipgloss.CompleteColor{TrueColor: "#00CC66", ANSI256: "41", ANSI: "2"}
	Rose        = lipgloss.CompleteColor{TrueColor: "#D90368", ANSI256: "161", ANSI: "1"}
	Yellow      = lipgloss.CompleteColor{TrueColor: "#F1C40F", ANSI256: "220", ANSI: "11"}
	Orange      = lipgloss.CompleteColor{TrueColor: "#F75C03", ANSI256: "202", ANSI: "3"}
	Blue        = lipgloss.CompleteColor{TrueColor: "#5f5fff", ANSI256: "63", ANSI: "4"}

	DefaultStyle Color = Color{
		Primary: lipgloss.CompleteAdaptiveColor{
			Light: Black,
			Dark:  White,
		},
		PrimaryBG: lipgloss.CompleteAdaptiveColor{
			Light: White,
			Dark:  Black,
		},
		Secondary: lipgloss.CompleteAdaptiveColor{
			Light: Gray,
			Dark:  LightGray,
		},
		SecondaryBG: lipgloss.CompleteAdaptiveColor{
			Light: White,
			Dark:  Black,
		},
		Tertiary: lipgloss.CompleteAdaptiveColor{
			Light: LightGray,
			Dark:  Gray,
		},
		TertiaryBG: lipgloss.CompleteAdaptiveColor{
			Light: Gray,
			Dark:  Black,
		},
		Accent: lipgloss.CompleteAdaptiveColor{
			Light: Purple,
			Dark:  Purple,
		},
		AccentBG: lipgloss.CompleteAdaptiveColor{
			Light: White,
			Dark:  Blue,
		},
		Accent2: lipgloss.CompleteAdaptiveColor{
			Light: Rose,
			Dark:  Rose,
		},
		Accent2BG: lipgloss.CompleteAdaptiveColor{
			Light: White,
			Dark:  Black,
		},
		Active: lipgloss.CompleteAdaptiveColor{
			Light: Black,
			Dark:  Black,
		},
		ActiveBG: lipgloss.CompleteAdaptiveColor{
			Light: Gray,
			Dark:  LightGray,
		},
		Status1: lipgloss.CompleteAdaptiveColor{
			Light: Black,
			Dark:  White,
		},
		Status1BG: lipgloss.CompleteAdaptiveColor{
			Light: LightGray,
			Dark:  Black,
		},
		Status2: lipgloss.CompleteAdaptiveColor{
			Light: Gray,
			Dark:  Black,
		},
		Status2BG: lipgloss.CompleteAdaptiveColor{
			Light: Black,
			Dark:  Gray,
		},
		Status3: lipgloss.CompleteAdaptiveColor{
			Light: LightPurple,
			Dark:  LightPurple,
		},
		Status3BG: lipgloss.CompleteAdaptiveColor{
			Light: Black,
			Dark:  Black,
		},
		PrimaryBorder: lipgloss.CompleteAdaptiveColor{
			Light: Purple,
			Dark:  Purple,
		},
		Warn: lipgloss.CompleteAdaptiveColor{
			Light: Orange,
			Dark:  Orange,
		},
		WarnBG: lipgloss.CompleteAdaptiveColor{
			Light: White,
			Dark:  White,
		},
	}
)

// Color defines the UI color palette for light and dark themes.
type Color struct {
	Primary       lipgloss.CompleteAdaptiveColor `json:"primary"`
	PrimaryBG     lipgloss.CompleteAdaptiveColor `json:"primary_background"`
	Secondary     lipgloss.CompleteAdaptiveColor `json:"secondary"`
	SecondaryBG   lipgloss.CompleteAdaptiveColor `json:"secondary_backgroundg"`
	Tertiary      lipgloss.CompleteAdaptiveColor `json:"tertiary"`
	TertiaryBG    lipgloss.CompleteAdaptiveColor `json:"tertiary_background"`
	Accent        lipgloss.CompleteAdaptiveColor `json:"accent"`
	AccentBG      lipgloss.CompleteAdaptiveColor `json:"accent_background"`
	Accent2       lipgloss.CompleteAdaptiveColor `json:"accent2"`
	Accent2BG     lipgloss.CompleteAdaptiveColor `json:"accent2_background"`
	Active        lipgloss.CompleteAdaptiveColor `json:"active"`
	ActiveBG      lipgloss.CompleteAdaptiveColor `json:"active_background"`
	Status1       lipgloss.CompleteAdaptiveColor `json:"staus_1"`
	Status1BG     lipgloss.CompleteAdaptiveColor `json:"staus_1_background"`
	Status2       lipgloss.CompleteAdaptiveColor `json:"staus_2"`
	Status2BG     lipgloss.CompleteAdaptiveColor `json:"status_2_background"`
	Status3       lipgloss.CompleteAdaptiveColor `json:"staus_3"`
	Status3BG     lipgloss.CompleteAdaptiveColor `json:"satus_3_background"`
	PrimaryBorder lipgloss.CompleteAdaptiveColor `json:"primary_border"`
	Warn          lipgloss.CompleteAdaptiveColor `json:"warn"`
	WarnBG        lipgloss.CompleteAdaptiveColor `json:"warn_background"`
}

// TODO UI config loading
// TODO UI config default
