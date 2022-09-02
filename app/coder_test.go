package app

import (
	helper "github.com/sapplications/sgo/helper/hashicorp/hclog"
	"gopkg.in/check.v1"
)

func (s *sgoSuite) TestCodeNumbers(c *check.C) {
	g := Coder{
		Logger: helper.NewStdOut("sgo", helper.LogLever.Debug),
	}
	items := s.copyItems()
	items[itemPath] = map[string]string{
		"Int1":   "5",
		"Float1": "5.02",
	}
	g.Init(items)
	err := g.Generate("test")
	c.Assert(err, check.IsNil)
}

func (s *sgoSuite) TestCodeParameters(c *check.C) {
	g := Coder{
		Logger: helper.NewStdOut("sgo", helper.LogLever.Debug),
	}
	items := s.copyItems()
	items[itemPath] = map[string]string{
		"Field1V2": "github.com/sapplications/sgo/test.NewField1V2(\"Ariana\", \"Noha\")",
		"Field2":   "github.com/sapplications/sgo/test.NewField2(\"Vitalii\")",
		"Field3":   "github.com/sapplications/sgo/test.NewField3(github.com/sapplications/sgo/test.Field1)",
		"Logger":   "github.com/sapplications/sgo/helper/hashicorp/hclog.NewFileOut(\"sgo\", 3)",
	}
	g.Init(items)
	err := g.Generate("test")
	c.Assert(err, check.IsNil)
}

func (s *sgoSuite) TestCodeRefs(c *check.C) {
	g := Coder{
		Logger: helper.NewStdOut("sgo", helper.LogLever.Debug),
	}
	items := s.copyItems()
	items[itemPath] = map[string]string{
		"Runner": "*github.com/sapplications/sgo/test.RunnerImpl",
	}
	g.Init(items)
	err := g.Generate("test")
	c.Assert(err, check.IsNil)
}

func (s *sgoSuite) TestCodeFuncs(c *check.C) {
	g := Coder{
		Logger: helper.NewStdOut("sgo", helper.LogLever.Debug),
	}
	items := s.copyItems()
	items[itemPath] = map[string]string{
		"Hello":     "github.com/sapplications/sgo/test.Hello()",
		"EmptyFunc": "github.com/sapplications/sgo/test.EmptyFunc()",
	}
	g.Init(items)
	err := g.Generate("test")
	c.Assert(err, check.IsNil)
}

func (s *sgoSuite) TestCodeCreators(c *check.C) {
	g := Coder{
		Logger: helper.NewStdOut("sgo", helper.LogLever.Debug),
	}
	items := s.copyItems()
	items[itemPath] = map[string]string{
		"Field1": "github.com/sapplications/sgo/test.NewField1()",
	}
	g.Init(items)
	err := g.Generate("test")
	c.Assert(err, check.IsNil)
}
