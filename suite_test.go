package sgo

import (
	"os"
	"strings"
	"testing"

	helper "github.com/nanomarkup/sgo/helper/hashicorp/hclog"
	"gopkg.in/check.v1"
)

func Test(t *testing.T) {
	suite.t = t
	check.TestingT(t)
}

type sgoSuite struct {
	t       *testing.T
	name    string
	items   map[string][][]string
	coder   Coder
	builder Builder
}

var logger = helper.NewStdOut("sgo", helper.LogLever.Debug)
var suite = sgoSuite{
	name: appName,
	items: map[string][][]string{
		"apps":  [][]string{{appName, ""}},
		appName: [][]string{{"entry", itemPath}},
	},
	coder:   Coder{Logger: logger},
	builder: Builder{Logger: logger},
}
var _ = check.Suite(&suite)

const (
	kind     string = "sb"
	appName  string = ".test"
	itemPath string = "github.com/nanomarkup/sgo/test.Item1"
)

func (s *sgoSuite) clean() {
	if _, err := os.Stat(s.name); err == nil {
		os.RemoveAll(s.name)
	}
}

func (s *sgoSuite) copyItems() (r map[string][][]string) {
	r = map[string][][]string{}
	for k, v := range s.items {
		r[k] = v
	}
	return
}

func getTestName(c *check.C) string {
	name := c.TestName()
	pos := strings.Index(name, ".")
	if pos > 0 {
		name = name[pos+1:]
	}
	return name
}
