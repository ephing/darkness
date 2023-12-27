package roxy

type PluginType string

const (
	AlphaPlugin PluginType = `alphaPlugin`
	MisaPlugin  PluginType = `misaPlugin`
)

type TomlInitializer interface {
	Set(map[string]any) error
	Get(string) (any, error)
}

type PluginError struct {
	Msg string
}

func (pe PluginError) Error() string {
	return pe.Msg
}
