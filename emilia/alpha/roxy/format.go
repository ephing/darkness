package roxy

import "github.com/thecsw/darkness/yunyun"

// filters out non-Chiho plugins and formats the plugin for Chiho
func FormatForChiho(plugins map[string]*Provider, conf interface{}) (formatted []yunyun.PageOption) {
	for _, provider := range plugins {
		if provider.Kind == ChihoPlugin {
			Do := provider.Do.(func(PluginConfigInterface, interface{}) yunyun.PageOption)
			formatted = append(formatted, Do(provider.Data, conf))
		}
	}
	return
}

type HTMLExportLocation = string

const (
	AuthorHeader HTMLExportLocation = `header`
)

// filters out non-HTMLExport plugins and plugins of the wrong location
func FormatForHTMLExport(plugins map[string]*Provider, location HTMLExportLocation) (formatted []*Provider) {
	for _, plgn := range plugins {
		if plgn.Kind != HTMLExportPlugin {
			continue
		}
		extras := plgn.Extra.(map[string]HTMLExportLocation)
		if _, ok := extras[location]; ok {
			formatted = append(formatted, plgn)
		}
	}
	return
}
