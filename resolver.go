// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package sgo

import (
	"fmt"
	"strings"
)

func (r *resolver) resolve(wd string) (items, []typeInfo, error) {
	items, err := r.getItems()
	if err != nil {
		return nil, nil, err
	}
	info, err := getCompiler().getTypeInfo(items, wd)
	if err != nil {
		return nil, nil, err
	} else {
		return items, info, nil
	}
}

func (r *resolver) getItems() (list items, err error) {
	list = make(items)
	_, err = r.getItem(r.entryPoint, list)
	if err != nil {
		return nil, err
	} else {
		return list, nil
	}
}

func (r *resolver) getItem(itemName string, list items) (*item, error) {
	if it, found := list[itemName]; found {
		return &it, nil
	}
	it, err := getParser().parseItem(itemName)
	if err != nil {
		return nil, err
	}
	simpleItemName := ""
	if it.group == "" {
		simpleItemName = itemName
	} else {
		simpleItemName = itemName[len(it.group)+2:]
	}
	if simpleItemName[0] == '*' {
		simpleItemName = simpleItemName[1:]
	}
	// process dependencies
	var refIt *item
	//var err error
	var deps [][]string
	groupItemName := ""
	if it.group == "" {
		deps = r.items[simpleItemName]
	} else {
		groupItemName = fmt.Sprintf("[%s]%s", it.group, simpleItemName)
		deps = r.items[groupItemName]
	}
	var k, v string
	var l int
	for _, n := range deps {
		l = len(n)
		if l == 0 {
			continue
		}
		k = n[0]
		if l > 1 {
			v = n[1]
		} else {
			v = ""
		}
		refIt, err = r.getItem(v, list)
		if err != nil {
			return nil, err
		} else if refIt != nil {
			it.deps = append(it.deps, dep{k, refIt})
		}
	}
	// process the input parameters for functions
	if it.kind == itemKind.Func {
		params, err := getParser().parseFunc(it.original)
		if err != nil {
			return nil, err
		}
		for _, param := range params {
			param = strings.Trim(param, " ")
			refIt, err = r.getItem(param, list)
			if err != nil {
				return nil, err
			} else if refIt != nil {
				it.deps = append(it.deps, dep{param, refIt})
			}
		}
	}
	// add a simple item to the result set
	if groupItemName == "" {
		list[simpleItemName] = it
	} else {
		list[groupItemName] = it
	}
	// if the original item is ref item then add it too
	if it.ref {
		list[itemName] = it
	}
	return &it, nil
}
