package sgo

import (
	"fmt"

	helper "github.com/sapplications/sgo/helper/hashicorp/hclog"
	"gopkg.in/check.v1"
)

func (s *sgoSuite) TestBuildEmpty(c *check.C) {
	b := Builder{
		Logger: helper.NewStdOut("sgo", helper.LogLever.Debug),
	}
	c.Assert(b.Build(s.name), check.ErrorMatches, fmt.Sprintf(BuilderFileDoesNotExistF, ".*"))
}
