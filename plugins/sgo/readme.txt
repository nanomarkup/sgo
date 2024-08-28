package sgo // import "github.com/nanomarkup/sgo/plugins/sgo"
Package sgo implements a sgo plugin. It generates Go sources of an application.
CONSTANTS
const (
	AppName string = "sgo"
)
TYPES
type Builder interface {
	Build(appName string) error
	Clean(appName string) error
	SetLogger(logger Logger)
}
type Coder interface {
	Init(items map[string][][]string)
	Clean(appName string) error
	Generate(appName string) error
	SetLogger(logger Logger)
}
type Logger interface {
	Trace(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	IsTrace() bool
	IsDebug() bool
	IsInfo() bool
	IsWarn() bool
	IsError() bool
}
type Plugin struct {
	Coder     Coder
	Builder   Builder
	Handshake plugin.HandshakeConfig
	Logger    Logger
}
func (p *Plugin) Execute()
