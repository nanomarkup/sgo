package sgo

import (
	"fmt"
	"strconv"
	"strings"
)

type itemParser interface {
	execute(string, *item) error
}

type refParser struct {
	next itemParser
}

type groupParser struct {
	next itemParser
}

type strParser struct {
	next itemParser
}

type intParser struct {
	next itemParser
}

type funcParser struct {
	next itemParser
}

type pathParser struct {
	next itemParser
}

func (p *parser) parseItem(input string) (item, error) {
	it := item{kind: itemKind.None, original: input, deps: make(items)}
	if err := p.itemParser.execute(input, &it); err != nil {
		return item{}, err
	}
	if it.kind == itemKind.None {
		return item{}, fmt.Errorf(ItemIsIncorrect, input)
	} else {
		return it, nil
	}
}

func (p *refParser) execute(input string, item *item) error {
	item.ref = input[0] == '*'
	// if item.ref {
	// 	input = input[1:]
	// }
	if p.next != nil {
		return p.next.execute(input, item)
	} else {
		return nil
	}
}

func (p *groupParser) execute(input string, item *item) error {
	if strings.HasPrefix(input, "[") {
		if pos := strings.Index(input, "]"); pos > -1 {
			item.group = input[1:pos]
			input = input[len(item.group)+2:]
		} else {
			return fmt.Errorf(GroupEndTokenIsMissing)
		}
	}
	if p.next != nil {
		return p.next.execute(input, item)
	} else {
		return nil
	}
}

func (p *strParser) execute(input string, item *item) error {
	if item.kind == itemKind.None && strings.HasPrefix(input, "\"") {
		item.kind = itemKind.String
		item.name = input
	}
	if p.next != nil {
		return p.next.execute(input, item)
	} else {
		return nil
	}
}

func (p *intParser) execute(input string, item *item) error {
	if item.kind == itemKind.None {
		if _, err := strconv.ParseFloat(input, 64); err == nil {
			item.kind = itemKind.Number
			item.name = input
		}
	}
	if p.next != nil {
		return p.next.execute(input, item)
	} else {
		return nil
	}
}

func (p *funcParser) execute(input string, item *item) error {
	if item.kind == itemKind.None {
		if pos := strings.Index(input, "("); pos > -1 {
			item.kind = itemKind.Func
		}
	}
	if p.next != nil {
		return p.next.execute(input, item)
	} else {
		return nil
	}
}

func (p *pathParser) execute(input string, item *item) error {
	if item.kind == itemKind.None || item.kind == itemKind.Func {
		var data []string
		pathSep := "/"
		nameSep := "."
		if item.kind == itemKind.None {
			item.kind = itemKind.Struct
			data = strings.Split(input, pathSep)
		} else {
			if pos := strings.Index(input, "("); pos > -1 {
				data = strings.Split(input[:pos], pathSep)
			}
		}
		// get path
		dataLen := len(data)
		fullName := data[dataLen-1]
		if dataLen > 1 {
			data = data[:dataLen-1]
			item.path = strings.Join(data, pathSep) + pathSep
		}
		// get pkg and item
		if fullName != "" {
			data = strings.Split(fullName, nameSep)
			dataLen = len(data)
			item.name = data[dataLen-1]
			if dataLen > 1 {
				item.pkg = data[0]
			}
		}
	}
	if p.next != nil {
		return p.next.execute(input, item)
	} else {
		return nil
	}
}
