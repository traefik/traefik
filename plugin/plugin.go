package plugin

import (
	"fmt"
	"strings"
)

// Flaeg Parsing Hooks
// Plugin defines a plugin configuration in traefik
type Plugin struct {
	Path  string
	Type  string
	Order string
}

func (p *Plugin) Before() bool {
	return p.Order == PluginBefore || p.Order == PluginAround
}

func (p *Plugin) After() bool {
	return p.Order == PluginAfter || p.Order == PluginAround
}

func (p *Plugin) Around() bool {
	return p.Order == PluginAround
}

const (
	PluginGo     = "go"
	PluginNetRpc = "netrpc"
	PluginGrpc   = "grpc"

	PluginBefore = "before"
	PluginAfter  = "after"
	PluginAround = "around"
)

// Plugins defines a set of Plugin
type Plugins []Plugin

//Set Plugins from a string expression
func (p *Plugins) Set(str string) error {
	exps := strings.Split(str, ",")
	if len(exps) == 0 {
		return fmt.Errorf("bad Plugin format: %s", str)
	}
	for _, exp := range exps {
		parts := strings.Split(exp, "|")
		if len(parts) != 2 {
			return fmt.Errorf("bad Plugin definition format: %s", exp)
		}

		t := strings.Split(parts[0], ":")
		order := PluginAround
		switch len(t) {
		case 0:
			return fmt.Errorf("bad Plugin type:order format: %s", parts[0])
		case 2:
			order = t[1]
		}

		*p = append(*p, Plugin{Path: parts[1], Type: t[0], Order: order})
	}
	return nil
}

//Get returns Plugins instances
func (p *Plugins) Get() interface{} {
	return []Plugin(*p)
}

//String returns Plugins formated in string
func (p *Plugins) String() string {
	if len(*p) == 0 {
		return ""
	}
	var result []string
	for _, pp := range *p {
		result = append(result, pp.Type+"|"+pp.Path)
	}
	return strings.Join(result, ",")
}

//SetValue sets Plugins into the parser
func (p *Plugins) SetValue(val interface{}) {
	*p = Plugins(val.(Plugins))
}

// Type exports the Plugins type as a string
func (p *Plugins) Type() string {
	return "plugins"
}
