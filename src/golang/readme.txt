package golang // import "github.com/sapplications/sgo/src/golang"

Package golang generates Go sources of an application.

TYPES

type Builder struct {
	Logger Logger
	// Has unexported fields.
}

func (b *Builder) Build(application string) error

func (b *Builder) Clean(application string) error

func (b *Builder) Init(items map[string]map[string]string)

func (b *Builder) SetLogger(logger Logger)

type Coder struct {
	Logger Logger
	// Has unexported fields.
}

func (g *Coder) Clean(application string) error

func (g *Coder) Generate(application string) error

func (g *Coder) Init(items map[string]map[string]string)

func (g *Coder) SetLogger(logger Logger)

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

