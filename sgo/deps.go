package main

import (
	p1 "github.com/nanomarkup/sgo/plugins/sgo"
	p2 "github.com/nanomarkup/sgo"
	p3 "github.com/nanomarkup/sgo/helper/hashicorp/hclog"
	p4 "github.com/hashicorp/go-plugin"
)

func Execute() {
	app := UseSgoPlugin()
	app.Execute()
}

func UseSgoPlugin() p1.Plugin {
	v := p1.Plugin{}
	v.Coder = UseSgoCoderSgoCoderAdapterRef()
	v.Builder = UseSgoBuilderSgoBuilderAdapterRef()
	v.Handshake = UseGo_PluginHandshakeConfig()
	v.Logger = p3.NewFileOut("sgo", 1)
	return v
}

func UseSgoCoderRef() *p2.Coder {
	v := &p2.Coder{}
	return v
}

func UseSgoBuilderRef() *p2.Builder {
	v := &p2.Builder{}
	return v
}

func UseGo_PluginHandshakeConfig() p4.HandshakeConfig {
	v := p4.HandshakeConfig{}
	v.ProtocolVersion = 1
	v.MagicCookieKey = "SMART_PLUGIN"
	v.MagicCookieValue = "sbuilder"
	return v
}

type SgoBuilderSgoBuilderAdapter struct {
	p2.Builder
}

func (o *SgoBuilderSgoBuilderAdapter) SetLogger(a1 p1.Logger) {
	b1 := a1.(p2.Logger)
	o.Builder.SetLogger(b1)
}

func UseSgoBuilderSgoBuilderAdapterRef() *SgoBuilderSgoBuilderAdapter {
	v := &SgoBuilderSgoBuilderAdapter{}
	v.Builder = *UseSgoBuilderRef()
	return v
}

type SgoCoderSgoCoderAdapter struct {
	p2.Coder
}

func (o *SgoCoderSgoCoderAdapter) SetLogger(a1 p1.Logger) {
	b1 := a1.(p2.Logger)
	o.Coder.SetLogger(b1)
}

func UseSgoCoderSgoCoderAdapterRef() *SgoCoderSgoCoderAdapter {
	v := &SgoCoderSgoCoderAdapter{}
	v.Coder = *UseSgoCoderRef()
	return v
}

