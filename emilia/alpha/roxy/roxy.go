package roxy

import "github.com/thecsw/darkness/yunyun"

type PluginKind string

const (
	ChihoPlugin      PluginKind = `chihoPlugin`
	MisaPlugin       PluginKind = `misaPlugin`
	HTMLExportPlugin PluginKind = `htmlExportPlugin`
)

/*
Plugin configs implement this interface.

Because structs are not valid symbols in go plugins, one cannot directly
import one using the go plugin `Lookup` function. Thus, access to the
plugin config has to pass through this interface.

Realistically, this interface will be implemented very similarly across all
plugins. However, unlike rust, there are not procedural macros in this language
that would allow me to auto-implement them, so that task is left to the plugin writer.
*/
type PluginConfigInterface interface {
	/*
		Sets the values in the config struct.
		Keys in the map should represent the struct members.
		In order to actually assign to the underlying struct, `Set` has to type check the values
	*/
	Set(map[string]any) error
}

type PluginError struct {
	Msg string
}

func (pe PluginError) Error() string {
	return pe.Msg
}

type Init = (func(map[string]any) (PluginConfigInterface, error))
type ChihoDo = (func(PluginConfigInterface, interface{}) yunyun.PageOption)
type MisaDo = (func(PluginConfigInterface, interface{}, bool) error)
type HTMLExportDo = (func(PluginConfigInterface, interface{}) string)
