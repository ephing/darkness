package main

import (
	"fmt"

	"github.com/thecsw/darkness/emilia/alpha"
	"github.com/thecsw/darkness/emilia/alpha/roxy"
	"github.com/thecsw/darkness/emilia/puck"
	"github.com/thecsw/darkness/yunyun"
)

// To run, compile this file with `go build -buildmode=plugin -o custom.so customplugin.go`
// Then run darkness like normal with the appropriate stuff in the toml file (its in the example)

var Name string = "TestConfig"
var PluginType roxy.PluginKind = roxy.ChihoPlugin

// Contains the settings for darkness.toml. Needs to implement roxy.TomlInitializer
type Test struct {
	flag bool
}

/*
Realistically, this function should be the same in every plugin.
All it does is take in a string and return the value at the key
that string represents.

This is necessary because structs are not symbols, meaning I cannot
directly import the config struct from a go plugin. However, I wanted
to enforce some degree of type safety that a struct would. Thus, I used
*/
func (t *Test) Get(key string) (any, error) {
	switch key {
	case "flag":
		return t.flag, nil
	}
	return nil, roxy.PluginError{Msg: "Invalid key: " + key}
}

func (t *Test) Set(vals map[string]any) error {
	var invalid []string
	for key, val := range vals {
		if key != "flag" {
			invalid = append(invalid, key)
			continue
		}
		if key == "flag" {
			var ok bool
			t.flag, ok = val.(bool)
			if !ok {
				return roxy.PluginError{Msg: "Value for key " + key + " is not a bool"}
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

func Do(plugConf roxy.PluginConfigInterface, globConf interface{}) yunyun.PageOption {
	conf := globConf.(*alpha.DarknessConfig)
	f, _ := plugConf.Get("flag")
	flag := f.(bool)
	return func(*yunyun.Page) {
		conf.Runtime.Logger.Warnf("Flag: %t", flag)
		conf.Runtime.Logger.Warnf("Title: %s", conf.Title)
	}
}
