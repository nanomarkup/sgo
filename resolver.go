// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package sgo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func (r *resolver) resolve(wd string) (items, []typeInfo, error) {
	id := ""
	kind := reflect.Interface
	list, err := r.getItems()
	if err != nil {
		return nil, nil, err
	}
	items := map[string]bool{}
	input := []typeInfo{}
	for _, x := range list {
		// the struct and interface types are supported only
		switch x.kind {
		case itemKind.Struct:
			kind = reflect.Struct
		default:
			continue
		}
		id = x.original
		// remove the group name from the original
		if x.group != "" {
			id = id[len(x.group)+2:]
		}
		// process only once as a simple type
		id = strings.TrimPrefix(id, "*")
		// do not process the same item again
		if _, found := items[id]; found {
			continue
		} else {
			items[id] = true
		}
		input = append(input, typeInfo{
			Id:      id,
			Kind:    kind,
			Name:    x.name,
			PkgPath: strings.TrimPrefix(x.path+x.pkg, "*"),
		})
	}
	all := []typeInfo{}
	if len(input) == 0 {
		return list, all, nil
	}
	for {
		curr, err := getTypeInfo(wd, input)
		if err != nil {
			return nil, nil, err
		}
		all = append(all, curr...)
		// get the next items to process
		input = []typeInfo{}
		for _, x := range curr {
			// process all fields
			if x.Fields != nil {
				for _, f := range x.Fields {
					// the struct and interface typers are supported only
					if (f.Kind != reflect.Struct && f.Kind != reflect.Interface) || f.Id == "." || f.PkgPath == "" {
						continue
					}
					// do not process the same item again
					if _, found := items[f.Id]; found {
						continue
					} else {
						items[f.Id] = true
					}
					input = append(input, typeInfo{
						Id:      f.Id,
						Kind:    f.Kind,
						Name:    f.TypeName,
						PkgPath: f.PkgPath,
					})
				}
			}
			// process all methods
			if x.Methods != nil {
				for _, m := range x.Methods {
					// input params
					for _, f := range m.In {
						// the struct and interface types are supported only
						if (f.Kind != reflect.Struct && f.Kind != reflect.Interface) || f.Id == "." || f.PkgPath == "" {
							continue
						}
						// do not process the same item again
						if _, found := items[f.Id]; found {
							continue
						} else {
							items[f.Id] = true
						}
						input = append(input, typeInfo{
							Id:      f.Id,
							Kind:    f.Kind,
							Name:    f.TypeName,
							PkgPath: f.PkgPath,
						})
					}
					// output params
					for _, f := range m.Out {
						// the struct and interface types are supported only
						if (f.Kind != reflect.Struct && f.Kind != reflect.Interface) || f.Id == "." || f.PkgPath == "" {
							continue
						}
						// do not process the same item again
						if _, found := items[f.Id]; found {
							continue
						} else {
							items[f.Id] = true
						}
						input = append(input, typeInfo{
							Id:      f.Id,
							Kind:    f.Kind,
							Name:    f.TypeName,
							PkgPath: f.PkgPath,
						})
					}
				}
			}
		}
		if len(input) == 0 {
			break
		}
	}
	return list, all, nil
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
	refType := false
	simpleItemName := ""
	if it.group == "" {
		simpleItemName = itemName
	} else {
		simpleItemName = itemName[len(it.group)+2:]
	}
	if simpleItemName[0] == '*' {
		refType = true
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
		id := ""
		name := it.original[strings.Index(it.original, "(")+1:]
		name = name[0:strings.Index(name, ")")]
		name = strings.Trim(name, " ")
		if name != "" {
			params := strings.Split(name, ",")
			for index, param := range params {
				id = strconv.Itoa(index)
				param = strings.Trim(param, " ")
				if strings.HasPrefix(param, "\"") {
					it.deps[id] = item{
						itemKind.String,
						"",
						"",
						"",
						"",
						param,
						false,
						make(items),
					}
				} else {
					refIt, err = r.getItem(param, list)
					if err != nil {
						return nil, err
					} else if refIt != nil {
						it.deps[id] = *refIt
					}
				}
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
	if refType {
		list[itemName] = it
	}
	return &it, nil
}
