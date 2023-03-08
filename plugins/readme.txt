package plugins // import "github.com/sapplications/sgo/plugins"

Package plugins implements common objects for supporting plugins.

TYPES

type Builder interface {
	Build(app string) error
	Clean(app string, sources *map[string][][]string) error
	Generate(app string, sources *map[string][][]string) error
}

type BuilderPlugin struct {
	Impl Builder
}

func (BuilderPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error)

func (p *BuilderPlugin) Server(*plugin.MuxBroker) (interface{}, error)

