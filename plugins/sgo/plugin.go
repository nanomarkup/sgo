// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package sgo

import (
	"github.com/hashicorp/go-plugin"
	"github.com/nanomarkup/sgo/plugins"
)

func (p *Plugin) Execute() {
	builder := builder{
		coder:   p.Coder,
		builder: p.Builder,
	}
	builder.coder.SetLogger(p.Logger)
	builder.builder.SetLogger(p.Logger)
	// pluginMap is the map of plugins we can dispense.
	var pluginMap = map[string]plugin.Plugin{
		AppName: &plugins.BuilderPlugin{Impl: &builder},
	}
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: p.Handshake,
		Plugins:         pluginMap,
	})
}
