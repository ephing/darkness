package roxy

import (
	"plugin"

	"github.com/BurntSushi/toml"
	"github.com/thecsw/darkness/yunyun"
)

type Provider struct {
	Name string
	Data TomlInitializer
}

func RegisterPlugin(path yunyun.FullPathFile, md toml.MetaData, prim toml.Primitive) (*Provider, error) {
	// Attempt to open shared library file
	plug, err := plugin.Open(string(path))
	if err != nil {
		return nil, err
	}

	// Get name of plugin
	symName, err := plug.Lookup("Name")
	if err != nil {
		return nil, err
	}

	// verify that this is an alpha plugin
	symType, err := plug.Lookup("PluginType")
	if err != nil {
		return nil, err
	}
	if *symType.(*PluginType) != AlphaPlugin {
		return nil, PluginError{"Not an Alpha Plugin"}
	}

	// get the init function
	symInit, err := plug.Lookup("Init")
	if err != nil {
		return nil, err
	}
	init := symInit.(func(map[string]any) (TomlInitializer, error))

	// decode the toml primitive data, ignore value types
	var keys map[string]any
	if err := md.PrimitiveDecode(prim, &keys); err != nil {
		return nil, err
	}
	// typify values using plugin, plus whatever else init might do
	tomlinit, err := init(keys)

	return &Provider{
		Name: *symName.(*string),
		Data: tomlinit,
	}, err
}
