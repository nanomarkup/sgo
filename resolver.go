// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package sgo

import (
	"fmt"
	"strconv"
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
	var deps map[string]string
	groupItemName := ""
	if it.group == "" {
		deps = r.items[simpleItemName]
	} else {
		groupItemName = fmt.Sprintf("[%s]%s", it.group, simpleItemName)
		deps = r.items[groupItemName]
	}
	for dep, res := range deps {
		refIt, err = r.getItem(res, list)
		if err != nil {
			return nil, err
		} else if refIt != nil {
			it.deps[dep] = *refIt
		}
	}
	// process the input parameters for functions
	if it.kind == itemKind.Func {
		params, err := getParser().parseFunc(it.original)
		if err != nil {
			return nil, err
		}
		id := ""
		for index, param := range params {
			id = strconv.Itoa(index)
			refIt, err = r.getItem(strings.Trim(param, " "), list)
			if err != nil {
				return nil, err
			} else if refIt != nil {
				it.deps[id] = *refIt
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
