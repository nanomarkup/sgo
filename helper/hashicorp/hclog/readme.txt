package helper // import "github.com/nanomarkup/sgo/helper/hashicorp/hclog"
Package helper provides methods for using "github.com/hashicorp/go-hclog"
sources.
VARIABLES
var LogLever = struct {
	NoLevel uint
	Trace   uint
	Debug   uint
	Info    uint
	Warn    uint
	Error   uint
	Off     uint
}{
	0,
	1,
	2,
	3,
	4,
	5,
	6,
}
FUNCTIONS
func NewFileOut(name string, level uint) hclog.Logger
func NewStdOut(name string, level uint) hclog.Logger
