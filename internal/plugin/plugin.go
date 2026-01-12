package plugin

import (
	lua "github.com/yuin/gopher-lua"
)

type Plugin struct {
	Path string
	Name string
}
type PluginRegister struct {
	State    *lua.LState
	Register []Plugin
}

func NewPluginRegister() *PluginRegister {
	o := lua.Options{}
	state := lua.NewState(o)
	return &PluginRegister{
		State: state,
	}

}

func (Pr *PluginRegister) RegisterPlugin(p Plugin) {
	pr := *Pr
	pr.Register = append(pr.Register, p)

}
