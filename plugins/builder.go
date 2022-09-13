// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package plugins

import (
	"encoding/gob"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// client's methods

func (c *builderClient) Build(app string) error {
	return c.client.Call("Plugin.Build", map[string]interface{}{
		"app": app,
	}, new(interface{}))
}

func (c *builderClient) Clean(app string, sources *map[string]map[string]string) error {
	return c.client.Call("Plugin.Clean", map[string]interface{}{
		"app":     app,
		"sources": sources,
	}, new(interface{}))
}

func (c *builderClient) Generate(app string, sources *map[string]map[string]string) error {
	return c.client.Call("Plugin.Generate", map[string]interface{}{
		"app":     app,
		"sources": sources,
	}, new(interface{}))
}

// server's methods

func (s *builderServer) Build(args map[string]interface{}, resp *interface{}) error {
	return s.Impl.Build(args["app"].(string))
}

func (s *builderServer) Clean(args map[string]interface{}, resp *interface{}) error {
	return s.Impl.Clean(args["app"].(string), args["sources"].(*map[string]map[string]string))
}

func (s *builderServer) Generate(args map[string]interface{}, resp *interface{}) error {
	return s.Impl.Generate(args["app"].(string), args["sources"].(*map[string]map[string]string))
}

// The implementation of plugin.Plugin so we can serve/consume this
//
// There are two methods: Server must return an RPC server for this plugin
// type. We construct a GreeterRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return GreeterRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.

func (p *BuilderPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	gob.Register(new(map[string]map[string]string))
	return &builderServer{Impl: p.Impl}, nil
}

func (BuilderPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	gob.Register(new(map[string]map[string]string))
	return &builderClient{client: c}, nil
}
