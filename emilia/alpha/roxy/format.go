package roxy

import "github.com/thecsw/darkness/yunyun"

// filters out non-Chiho plugins and formats the plugin for Chiho
func FormatForChiho(plugins []*Provider, conf interface{}) []yunyun.PageOption {
	var formatted []yunyun.PageOption
	for _, provider := range plugins {
		if provider.Kind == ChihoPlugin {
			Do := provider.Do.(func(PluginConfigInterface, interface{}) yunyun.PageOption)
			formatted = append(formatted, Do(provider.Data, conf))
		}
	}
	return formatted
}
