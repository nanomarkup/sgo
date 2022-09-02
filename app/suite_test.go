package app

import (
	"testing"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

type sgoSuite struct {
	items map[string]map[string]string
}

func (s *sgoSuite) copyItems() (r map[string]map[string]string) {
	r = map[string]map[string]string{}
	for k, v := range s.items {
		r[k] = v
	}
	return
}

var _ = check.Suite(&sgoSuite{
	items: map[string]map[string]string{
		"apps": {"test": ""},
		"test": {"entry": itemPath},
	},
})

const (
	itemPath string = "github.com/sapplications/sgo/test.Item1"
)
