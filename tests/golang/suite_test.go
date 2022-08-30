package golang

import (
	"testing"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

type GoSuite struct {
}

var _ = check.Suite(&GoSuite{})
