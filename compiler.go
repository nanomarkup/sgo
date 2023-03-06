package sgo

import (
	"reflect"
	"strings"
)

func (c *compiler) getTypeInfo(list items, wd string) ([]typeInfo, error) {
	id := ""
	kind := reflect.Interface
	done := map[string]bool{}
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
		if _, found := done[id]; found {
			continue
		} else {
			done[id] = true
		}
		input = append(input, typeInfo{
			Id:      id,
			Kind:    kind,
			Name:    x.name,
			PkgPath: strings.TrimPrefix(x.path+x.pkg, "*"),
		})
	}
	if len(input) == 0 {
		return []typeInfo{}, nil
	} else {
		return c.processTypes(input, done, wd)
	}
}

func (c *compiler) processTypes(list []typeInfo, done map[string]bool, wd string) ([]typeInfo, error) {
	curr, err := getTypeInfo(wd, list)
	if err != nil {
		return nil, err
	}
	// get next items to process
	next := []typeInfo{}
	for _, x := range curr {
		// process all fields
		if x.Fields != nil {
			for _, f := range x.Fields {
				// the struct and interface typers are supported only
				if (f.Kind != reflect.Struct && f.Kind != reflect.Interface) || f.Id == "." || f.PkgPath == "" {
					continue
				}
				// do not process the same item again
				if _, found := done[f.Id]; found {
					continue
				} else {
					done[f.Id] = true
				}
				next = append(next, typeInfo{
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
					if _, found := done[f.Id]; found {
						continue
					} else {
						done[f.Id] = true
					}
					next = append(next, typeInfo{
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
					if _, found := done[f.Id]; found {
						continue
					} else {
						done[f.Id] = true
					}
					next = append(next, typeInfo{
						Id:      f.Id,
						Kind:    f.Kind,
						Name:    f.TypeName,
						PkgPath: f.PkgPath,
					})
				}
			}
		}
	}
	if len(next) > 0 {
		// recursion...
		n, e := c.processTypes(next, done, wd)
		if e != nil {
			return nil, err
		}
		curr = append(curr, n...)
	}
	return curr, nil
}
