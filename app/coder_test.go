package app

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/sapplications/smod/lod"
	"gopkg.in/check.v1"
)

func (s *sgoSuite) TestCodeNumbers(c *check.C) {
	defer s.clean()
	items := s.copyItems()
	items[itemPath] = map[string]string{
		"Int1":   "5",
		"Float1": "5.02",
	}
	s.coder.Init(items)
	c.Assert(s.coder.Generate(s.name), check.IsNil)
	c.Assert(s.t.Run(fmt.Sprintf("%s-Build", getTestName(c)), func(t *testing.T) {
		if err := s.builder.Build(s.name); err != nil {
			t.Error(err)
		}
	}), check.Equals, true)
}

func (s *sgoSuite) TestCodeParameters(c *check.C) {
	defer s.clean()
	items := s.copyItems()
	items[itemPath] = map[string]string{
		"Field1V2": "github.com/sapplications/sgo/test.NewField1V2(\"Ariana\", \"Noha\")",
		"Field2":   "github.com/sapplications/sgo/test.NewField2(\"Vitalii\")",
		"Field3":   "github.com/sapplications/sgo/test.NewField3(github.com/sapplications/sgo/test.Field1)",
		"Logger":   "github.com/sapplications/sgo/helper/hashicorp/hclog.NewFileOut(\"sgo\", 3)",
	}
	s.coder.Init(items)
	c.Assert(s.coder.Generate(s.name), check.IsNil)
	c.Assert(s.t.Run(fmt.Sprintf("%s-Build", getTestName(c)), func(t *testing.T) {
		if err := s.builder.Build(s.name); err != nil {
			t.Error(err)
		}
	}), check.Equals, true)
}

func (s *sgoSuite) TestCodeRefs(c *check.C) {
	defer s.clean()
	items := s.copyItems()
	items[itemPath] = map[string]string{
		"Runner": "*github.com/sapplications/sgo/test.RunnerImpl",
	}
	s.coder.Init(items)
	c.Assert(s.coder.Generate(s.name), check.IsNil)
	c.Assert(s.t.Run(fmt.Sprintf("%s-Build", getTestName(c)), func(t *testing.T) {
		if err := s.builder.Build(s.name); err != nil {
			t.Error(err)
		}
	}), check.Equals, true)
}

func (s *sgoSuite) TestCodeFuncs(c *check.C) {
	defer s.clean()
	items := s.copyItems()
	items[itemPath] = map[string]string{
		"Hello":     "github.com/sapplications/sgo/test.Hello()",
		"EmptyFunc": "github.com/sapplications/sgo/test.EmptyFunc()",
	}
	s.coder.Init(items)
	c.Assert(s.coder.Generate(s.name), check.IsNil)
	c.Assert(s.t.Run(fmt.Sprintf("%s-Build", getTestName(c)), func(t *testing.T) {
		if err := s.builder.Build(s.name); err != nil {
			t.Error(err)
		}
	}), check.Equals, true)
}

func (s *sgoSuite) TestCodeCreators(c *check.C) {
	defer s.clean()
	items := s.copyItems()
	items[itemPath] = map[string]string{
		"Field1": "github.com/sapplications/sgo/test.NewField1()",
	}
	s.coder.Init(items)
	c.Assert(s.coder.Generate(s.name), check.IsNil)
	c.Assert(s.t.Run(fmt.Sprintf("%s-Build", getTestName(c)), func(t *testing.T) {
		if err := s.builder.Build(s.name); err != nil {
			t.Error(err)
		}
	}), check.Equals, true)
}

func (s *sgoSuite) TestCodeSgoUsingGoModules(c *check.C) {
	m := lod.Manager{}
	m.SetLogger(hclog.New(&hclog.LoggerOptions{
		Name:   "test",
		Level:  hclog.Trace,
		Output: os.Stdout,
	}))
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir(".\\..")
	r, e := m.ReadAll("sb")
	if e != nil {
		fmt.Println(e.Error())
		c.Error()
		return
	}
	c.Assert(r, check.NotNil)
	// create a temporary folder and use it as working folder for generating an application
	os.Chdir(c.MkDir())
	s.coder.Init(r.Items())
	c.Assert(s.coder.Generate("sgo"), check.IsNil)
}
