// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package golang

import (
	"reflect"
	"strconv"
	"strings"
)

func (r *resolver) resolve() (items, []typeInfo, error) {
	id := ""
	kind := reflect.Interface
	list := r.getItems()
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
		id = strings.TrimPrefix(x.original, "*")
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
		curr, err := getTypeInfo(input)
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

func (r *resolver) getItems() (list items) {
	list = make(items)
	r.getItem(r.entryPoint, list)
	return list
}

func (r *resolver) getItem(itemName string, list items) *item {
	if it, found := list[itemName]; found {
		return &it
	}
	// parse item and add it to the list
	pkg := ""
	name := ""
	kind := itemKind.Struct
	path := ""
	pathSep := "/"
	nameSep := "."
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
		pkg,
		path,
		itemName,
		make(items),
	}
	// process a simple item dependencies
	simpleItemName := itemName
	if itemName[0] == '*' {
		simpleItemName = itemName[1:]
	}
	var refIt *item
	deps := r.items[simpleItemName]
	for dep, res := range deps {
		refIt = r.getItem(res, list)
		if refIt != nil {
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
						param,
						make(items),
					}
				} else {
					refIt = r.getItem(param, list)
					if refIt != nil {
						it.deps[id] = *refIt
					}
				}
			}
		}
	}
	// add simple and ref items to the result set and return it
	list[simpleItemName] = it
	if itemName[0] == '*' {
		list[itemName] = it
	}
	return &it
}
