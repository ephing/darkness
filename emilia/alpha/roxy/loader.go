package roxy

import (
	"plugin"

	"github.com/BurntSushi/toml"
	"github.com/thecsw/darkness/yunyun"
)

type Provider struct {
	// Name is the name of the plugin
	Name string

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
}

// Just for organizing the information collected from the plugin
type pluginMembers struct {
	name       *string
	pluginkind *PluginKind
	init       func(map[string]any) (PluginConfigInterface, error)
	do         interface{}
}

/*
Verifies the contents of a plugin and its config
*/
func AlphaRegisterPlugin(path yunyun.FullPathFile, md toml.MetaData, prim toml.Primitive) (*Provider, error) {
	// Attempt to open shared library file
	plug, err := plugin.Open(string(path))
	if err != nil {
		return nil, err
	}

	plmem := &pluginMembers{}
	var ok bool

	// Get name of plugin
	symName, err := plug.Lookup("Name")
	if err != nil {
		return nil, err
	}
	plmem.name, ok = symName.(*string)
	if !ok {
		return nil, PluginError{
			Msg: `Invalid valid type for Name in plugin "` + string(path) + ". Expected: string",
		}
	}

	// get the plugin kind
	symType, err := plug.Lookup("PluginType")
	if err != nil {
		return nil, err
	}
	ptype, ok := symType.(*PluginKind)
	if !ok {
		pstype, ok := symType.(*string)
		if !ok {
			return nil, PluginError{
				Msg: "Invalid type for PluginType in plugin " + *plmem.name + ". Expected {roxy.PluginKind, string}",
			}
		}
		plmem.pluginkind = (*PluginKind)(pstype)
	} else {
		plmem.pluginkind = ptype
	}

	// get the init function
	symInit, err := plug.Lookup("Init")
	if err != nil {
		return nil, err
	}
	plmem.init, ok = symInit.(func(map[string]any) (PluginConfigInterface, error))
	if !ok {
		return nil, PluginError{
			Msg: "Invalid signature for Init of plugin " + *plmem.name +
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
		// see if its an html insertion function
		plmem.do, ok = symDo.(func(PluginConfigInterface, interface{}) yunyun.PageOption)
		if !ok {
			return nil, PluginError{
				Msg: "Invalid signature for Do function in plugin " + *plmem.name +
					". Expected: func(roxy.PluginConfigInterface, interface{}) yunyun.PageOption",
			}
		}
	default:
		return nil, PluginError{Msg: string(*plmem.pluginkind) + "s cannot be used in darkness.toml"}
	}

	return &Provider{
		Name: *plmem.name,
		Kind: *plmem.pluginkind,
		Data: tomlinit,
		Do:   plmem.do,
	}, err
}
