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
	// parse item and add it to the list
	pkg := ""
	name := ""
	kind := itemKind.Struct
	group := ""
	path := ""
	pathSep := "/"
	nameSep := "."
	refType := false
	simpleItemName := itemName
	// extract group name if exists
	if strings.HasPrefix(itemName, "[") {
		if pos := strings.Index(itemName, "]"); pos > -1 {
			group = itemName[1:pos]
			simpleItemName = itemName[pos+1:]
		} else {
			return nil, fmt.Errorf(GroupEndTokenIsMissing)
		}
	}
	// process a simple item dependencies
	if simpleItemName[0] == '*' {
		refType = true
		simpleItemName = simpleItemName[1:]
	}
	// check item type and process it
	if strings.HasPrefix(itemName, "\"") {
		kind = itemKind.String
		name = itemName
	} else if _, err := strconv.ParseFloat(itemName, 64); err == nil {
		kind = itemKind.Number
		name = itemName
	} else {
		// get path
		var data []string
		if pos := strings.Index(itemName, "("); pos > -1 {
			kind = itemKind.Func
			data = strings.Split(itemName[:pos], pathSep)
		} else {
			data = strings.Split(itemName, pathSep)
		}
		dataLen := len(data)
		fullName := data[dataLen-1]
		if dataLen > 1 {
			data = data[:dataLen-1]
			path = strings.Join(data, pathSep) + pathSep
			if group != "" {
				path = path[len(group)+2:]
			}
		}
		// get pkg and item
		if fullName != "" {
			data = strings.Split(fullName, nameSep)
			dataLen = len(data)
			name = data[dataLen-1]
			if dataLen > 1 {
				pkg = data[0]
			}
		}
	}
	// create an item
	it := item{
		kind,
		name,
		group,
		pkg,
		path,
		itemName,
		make(items),
	}
	// process dependencies
	var refIt *item
	var err error
	deps := r.items[itemName]
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
		name = it.original[strings.Index(it.original, "(")+1:]
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
	if it.group == "" {
		list[simpleItemName] = it
	} else {
		list[fmt.Sprintf("[%s]%s", it.group, simpleItemName)] = it
	}
	// if the original item is ref item then add it too
	if refType {
		list[itemName] = it
	}
	return &it, nil
}
