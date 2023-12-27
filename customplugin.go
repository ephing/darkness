package main

import (
	"github.com/thecsw/darkness/emilia/alpha/roxy"
)

// To run, compile this file with `go build -buildmode=plugin -o custom.so customplugin.go`
// Then run darkness like normal with the appropriate stuff in the toml file (its in the example)

var Name string = "TestConfig"
var PluginType roxy.PluginType = roxy.AlphaPlugin

type Test struct {
	Head string `toml:"head"`
	Tail int64  `toml:"tail"`
}

func (t *Test) Set(vals map[string]any) error {
	for key := range vals {
		if key != "head" && key != "tail" {
			return roxy.PluginError{Msg: "Contains extra keys: " + key}
		}
	}

	head, ok := vals["head"]
	if !ok {
		// optionally just do nothing or set default value
		return roxy.PluginError{Msg: "Missing head"}
	}
	t.Head = head.(string)

	if tail, ok := vals["tail"]; ok {
		t.Tail, ok = tail.(int64)
		if !ok {
			return roxy.PluginError{Msg: "Invalid type for Test.tail"}
		}
	}

	return nil
}

func (t *Test) Get(key string) (any, error) {
	switch key {
	case "head":
		return t.Head, nil
	case "tail":
		return t.Tail, nil
	}
	return nil, roxy.PluginError{Msg: "Invalid key: " + key}
}

func Init(vals map[string]any) (roxy.TomlInitializer, error) {
	t := &Test{}
	err := t.Set(vals)
	return t, err
}
