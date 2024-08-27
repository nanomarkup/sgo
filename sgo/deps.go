package main

import (
	p2 "github.com/hashicorp/go-plugin"
	p1 "github.com/nanomarkup/sgo"
	p4 "github.com/nanomarkup/sgo/helper/hashicorp/hclog"
	p3 "github.com/nanomarkup/sgo/plugins/sgo"
)

func Execute() {
	app := UseSgoPlugin()
	app.Execute()
}

func UseSgoCoderRef() *p1.Coder {
	v := &p1.Coder{}
	return v
}

func UseSgoBuilderRef() *p1.Builder {
	v := &p1.Builder{}
	return v
}

func UseGo_PluginHandshakeConfig() p2.HandshakeConfig {
	v := p2.HandshakeConfig{}
	v.ProtocolVersion = 1
	v.MagicCookieKey = "SMART_PLUGIN"
	v.MagicCookieValue = "sbuilder"
	return v
}

func UseSgoPlugin() p3.Plugin {
	v := p3.Plugin{}
	v.Coder = UseSgoCoderSgoCoderAdapterRef()
	v.Builder = UseSgoBuilderSgoBuilderAdapterRef()
	v.Handshake = UseGo_PluginHandshakeConfig()
	v.Logger = p4.NewFileOut("sgo", 1)
	return v
}

type SgoCoderSgoCoderAdapter struct {
	p1.Coder
}

func (o *SgoCoderSgoCoderAdapter) SetLogger(a1 p3.Logger) {
	b1 := a1.(p1.Logger)
	o.Coder.SetLogger(b1)
}

func UseSgoCoderSgoCoderAdapterRef() *SgoCoderSgoCoderAdapter {
	v := &SgoCoderSgoCoderAdapter{}
	v.Coder = *UseSgoCoderRef()
	return v
}

type SgoBuilderSgoBuilderAdapter struct {
	p1.Builder
}

func (o *SgoBuilderSgoBuilderAdapter) SetLogger(a1 p3.Logger) {
	b1 := a1.(p1.Logger)
	o.Builder.SetLogger(b1)
}

func UseSgoBuilderSgoBuilderAdapterRef() *SgoBuilderSgoBuilderAdapter {
	v := &SgoBuilderSgoBuilderAdapter{}
	v.Builder = *UseSgoBuilderRef()
	return v
}
