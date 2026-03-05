package tui

import (
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/plugin"
)

// AppContext provides read-only access to shared application state for Screen implementations.
// Passed by value so screens cannot mutate the root Model directly.
type AppContext struct {
	DB               db.DbDriver
	Config           *config.Config
	Logger           Logger
	UserID           types.UserID
	Width            int
	Height           int
	ScreenMode       ScreenMode
	PanelFocus       FocusPanel
	PluginManager    *plugin.Manager
	ConfigManager    *config.Manager
	IsRemote         bool
	SSHFingerprint   string
	SSHKeyType       string
	SSHPublicKey     string
	ActiveLocale     string
	AccordionEnabled bool
}

// AppCtx builds an AppContext snapshot from the current Model state.
func (m Model) AppCtx() AppContext {
	return AppContext{
		DB:               m.DB,
		Config:           m.Config,
		Logger:           m.Logger,
		UserID:           m.UserID,
		Width:            m.Width,
		Height:           m.Height,
		ScreenMode:       m.ScreenMode,
		PanelFocus:       m.PanelFocus,
		PluginManager:    m.PluginManager,
		ConfigManager:    m.ConfigManager,
		IsRemote:         m.IsRemote,
		SSHFingerprint:   m.SSHFingerprint,
		SSHKeyType:       m.SSHKeyType,
		SSHPublicKey:     m.SSHPublicKey,
		ActiveLocale:     m.ActiveLocale,
		AccordionEnabled: m.AccordionEnabled,
	}
}
