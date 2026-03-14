package tui

import (
	"github.com/hegner123/modulacms/internal/plugin"
	lua "github.com/yuin/gopher-lua"
)

// PluginScreenDef describes a plugin screen discovered from a manifest.
type PluginScreenDef struct {
	PluginName string
	ScreenName string
	Label      string
	Icon       string
	Hidden     bool
}

// NavigateToPluginScreenMsg requests navigation to a plugin's TUI screen.
// Handled in update.go to create a PluginTUIScreen and push it onto history.
type NavigateToPluginScreenMsg struct {
	PluginName string
	ScreenName string
	Params     map[string]string
}

// PluginDataMsg delivers the result of an async fetch/request action to the
// plugin screen. The bridge resumes the coroutine with a data event.
type PluginDataMsg struct {
	ID     string
	OK     bool
	Result lua.LValue
	Error  string
}

// PluginDialogResponseMsg delivers the result of a confirm action dialog.
type PluginDialogResponseMsg struct {
	Accepted bool
}

// PluginScreenInitMsg is sent to the PluginTUIScreen after construction to
// trigger coroutine initialization. Carries the bridge and init context.
type PluginScreenInitMsg struct {
	Bridge *plugin.CoroutineBridge
	L      *lua.LState
	Width  int
	Height int
	Params map[string]string
}

// PluginScreenErrorMsg signals that plugin screen initialization failed.
type PluginScreenErrorMsg struct {
	PluginName string
	ScreenName string
	Error      string
}

// ShowPluginConfirmDialogMsg requests a plugin confirm dialog to be shown.
// Handled by UpdateDialog to create a dialog with DIALOGPLUGINCONFIRM action.
type ShowPluginConfirmDialogMsg struct {
	Title   string
	Message string
}
