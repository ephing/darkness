package roxy

import (
	"plugin"

	"github.com/BurntSushi/toml"
	"github.com/thecsw/darkness/yunyun"
)

type Provider struct {
	// Kind defines where the plugin should have an affect (e.g. Chiho, Misa)
	Kind PluginKind

	// Data contains the config defined in darkness.toml
	Data PluginConfigInterface

	/*
		Do defines what the plugin do.
		Type is interface{} for generalization of plugin kinds, will need to type assert the actual
		function type in `format.go`. Type of function will be correct because it is checked
		when the plugin is registered
	*/
	Do interface{}

	// For extra info not used across all plugin kinds
	Extra interface{}
}

// Just for organizing the information collected from the plugin
type pluginMembers struct {
	pluginkind *PluginKind
	init       func(map[string]any) (PluginConfigInterface, error)
	do         interface{}
	extra      interface{}
}

// Get kind of plugin
func (pm *pluginMembers) getKind(plug *plugin.Plugin, name string) error {
	symType, err := plug.Lookup("PluginType")
	if err != nil {
		return err
	}
	ptype, ok := symType.(*PluginKind)
	if !ok {
		pstype, ok := symType.(*string)
		if !ok {
			return PluginError{
				Msg: "Invalid type for PluginType in plugin " + name + ". Expected {roxy.PluginKind, string}",
			}
		}
		pm.pluginkind = (*PluginKind)(pstype)
		return nil
	}
	pm.pluginkind = ptype
	return nil
}

// Verifies the contents of a plugin and its config
func RegisterPlugin(path yunyun.FullPathFile, name string, md toml.MetaData, prim toml.Primitive) (*Provider, error) {
	// Attempt to open shared library file
	plug, err := plugin.Open(string(path))
	if err != nil {
		return nil, err
	}

	plmem := &pluginMembers{}
	var ok bool

	if err := plmem.getKind(plug, name); err != nil {
		return nil, err
	}

	// get the init function
	symInit, err := plug.Lookup("Init")
	if err != nil {
		return nil, err
	}
	plmem.init, ok = symInit.(Init)
	if !ok {
		return nil, PluginError{
			Msg: "Invalid signature for Init of plugin " + name +
				". Expected: func(map[string]any) (roxy.PluginConfigInterface, error)",
		}
	}

	// decode the toml primitive data, ignore value types
	var keys map[string]any
	if err := md.PrimitiveDecode(prim, &keys); err != nil {
		return nil, err
	}
	// type check values using plugin, plus whatever else init might do
	tomlinit, err := plmem.init(keys)
	if err != nil {
		return nil, err
	}

	// Differing symbols based on plugin kind
	switch *plmem.pluginkind {
	case ChihoPlugin:
		// get the do function
		symDo, err := plug.Lookup("Do")
		if err != nil {
			return nil, err
		}
		// see if its a chiho do function
		plmem.do, ok = symDo.(ChihoDo)
		if !ok {
			return nil, PluginError{
				Msg: "Invalid signature for Do function in plugin " + name +
					". Expected: func(roxy.PluginConfigInterface, interface{}) yunyun.PageOption",
			}
		}
	case MisaPlugin:
		symDo, err := plug.Lookup("Do")
		if err != nil {
			return nil, err
		}
		// see if its a misado function
		plmem.do, ok = symDo.(MisaDo)
		if !ok {
			return nil, PluginError{
				Msg: "Invalid signature for Do function in plugin " + name +
					". Expected: func(roxy.PluginConfigInterface, interface{}, bool) error",
			}
		}
	default:
		return nil, PluginError{Msg: string(*plmem.pluginkind) + "s cannot be used in darkness.toml"}
	}

	return &Provider{
		Kind: *plmem.pluginkind,
		Data: tomlinit,
		Do:   plmem.do,
	}, err
}
