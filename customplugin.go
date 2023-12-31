package main

import (
	"fmt"

	"github.com/thecsw/darkness/emilia/alpha/roxy"
	"github.com/thecsw/darkness/emilia/puck"
)

// To run, compile this file with `go build -buildmode=plugin -o custom.so customplugin.go`
// Then run darkness like normal with the appropriate stuff in the toml file (its in the example)

var PluginType roxy.PluginKind = roxy.MisaPlugin

// Contains the settings for darkness.toml. Needs to implement roxy.PluginConfigInterface
type Test struct {
	flag   bool
	submap map[string]int64
}

func (t *Test) Set(vals map[string]any) error {
	var invalid []string
	for key, val := range vals {
		if key != "flag" && key != "submap" {
			invalid = append(invalid, key)
			continue
		}
		var ok bool
		if key == "flag" {
			t.flag, ok = val.(bool)
			if !ok {
				return roxy.PluginError{Msg: fmt.Sprintf("%v for key %s is not a bool", val, key)}
			}
		}
		if key == "submap" {
			temp, ok := val.(map[string]interface{})
			if !ok {
				return roxy.PluginError{Msg: fmt.Sprintf("%v for key %s is not a map", val, key)}
			}
			t.submap = map[string]int64{}
			for k, v := range temp {
				if t.submap[k], ok = v.(int64); !ok {
					return roxy.PluginError{Msg: fmt.Sprintf("%v for key %s is not an int64", v, k)}
				}
			}
		}
	}
	if len(invalid) > 0 {
		return roxy.PluginError{Msg: fmt.Sprintf("Invalid keys: %s", invalid)}
	}
	return nil
}

func Init(vals map[string]any) (roxy.PluginConfigInterface, error) {
	test := &Test{}
	log := puck.NewLogger("Plugin", puck.InfoLevel)
	log.Info("Calling init!")
	err := test.Set(vals)
	return test, err
}

func Do(plugConf roxy.PluginConfigInterface, globConf interface{}, dryRun bool) error {
	log := puck.NewLogger("plugin", puck.InfoLevel)
	log.Info("zoinks!")
	conf := plugConf.(*Test)
	log.Infof("config: %v", conf)
	return nil
}
