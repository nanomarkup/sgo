package sgo

import (
	"fmt"
	"reflect"
	"sort"
)

type structBegGen struct {
	next structGenerator
}

type structEndGen struct {
	next structGenerator
}

type structCreateGen struct {
	next structGenerator
}

type structInitGen struct {
	next structGenerator
}

func (g *generator) createStruct(it item, types []typeInfo, imp imports, adapter *adapter) ([]string, error) {
	code := []string{}
	if err := g.structGenerator.execute(it, types, imp, &code, adapter); err != nil {
		return nil, err
	} else {
		return code, nil
	}
}

func (s *structBegGen) execute(it item, types []typeInfo, imp imports, code *[]string, adapter *adapter) error {
	funcName := getFuncName(&it, false)
	fullNameDefine := it.name
	if len(it.path) > 0 {
		if it.path[0] == '*' {
			alias := string(appendImport(imp, it.path[1:]+it.pkg))
			funcName = funcName + GenRefSufix
			fullNameDefine = fmt.Sprintf("*%s.%s", alias, it.name)
		} else {
			alias := string(appendImport(imp, it.path+it.pkg))
			fullNameDefine = fmt.Sprintf("%s.%s", alias, it.name)
		}
	}
	*code = append(*code, fmt.Sprintf("func %s() %s {\n", funcName, fullNameDefine))
	if s.next != nil {
		return s.next.execute(it, types, imp, code, adapter)
	} else {
		return nil
	}
}

func (s *structCreateGen) execute(it item, types []typeInfo, imp imports, code *[]string, adapter *adapter) error {
	returnName := it.name
	if len(it.path) > 0 {
		if it.ref {
			alias := string(appendImport(imp, it.path[1:]+it.pkg))
			returnName = fmt.Sprintf("&%s.%s", alias, it.name)
		} else {
			alias := string(appendImport(imp, it.path+it.pkg))
			returnName = fmt.Sprintf("%s.%s", alias, it.name)
		}
	}
	*code = append(*code, fmt.Sprintf("\tv := %s{}\n", returnName))
	if s.next != nil {
		return s.next.execute(it, types, imp, code, adapter)
	} else {
		return nil
	}
}

func (s *structInitGen) execute(it item, types []typeInfo, imp imports, code *[]string, adapter *adapter) error {
	for k, v := range it.deps {
		switch v.kind {
		case itemKind.Func:
			alias := string(appendImport(imp, v.path+v.pkg))
			field, err := getFieldInfo(types, it.original, k)
			if err != nil {
				return err
			}
			switch field.Kind {
			case reflect.Func:
				// if it is a reference to a func then just return it as is
				*code = append(*code, fmt.Sprintf("\tv.%s = %s.%s\n", k, alias, v.name))
			case reflect.Struct, reflect.Interface:
				// if it is a reference to a struct then perform it
				name := v.name + "("
				if len(v.deps) > 0 {
					keys := reflect.ValueOf(v.deps).MapKeys()
					keysOrder := func(i, j int) bool { return keys[i].String() < keys[j].String() }
					sort.Slice(keys, keysOrder)
					for i, n := range keys {
						d := v.deps[n.String()]
						// process all parameters IN PROGRESS
						parameter := ""
						switch d.kind {
						case itemKind.Func:
							parameter = d.name
						case itemKind.Struct:
							funcName := getFuncName(&d, len(d.path) > 0 && d.path[0] == '*')
							parameter = funcName + "()"
						case itemKind.String, itemKind.Number:
							parameter = d.original
						default:
							return fmt.Errorf(TypeDoesNotSupportedF, d.original)
						}

						if i == 0 {
							name = name + parameter
						} else {
							name = fmt.Sprintf("%s, %s", name, parameter)
						}
					}
				}
				name = name + ")"
				*code = append(*code, fmt.Sprintf("\tv.%s = %s.%s\n", k, alias, name))
			default:
				return fmt.Errorf(TypeDoesNotSupportedF, v.original)
			}
		case itemKind.Struct:
			typeId1 := it.original
			if typeId1[0] == '*' {
				typeId1 = typeId1[1:]
			}
			typeId2 := v.original
			if v.group != "" {
				typeId2 = typeId2[len(v.group)+2:]
			}
			if typeId2[0] == '*' {
				typeId2 = typeId2[1:]
			}
			supported, err := adapter.areTypesCompatible(types, typeId1, k, typeId2)
			if err != nil {
				return err
			}
			ref := len(v.path) > 0 && v.path[0] == '*'
			funcName := ""
			if supported {
				funcName = getFuncName(&v, ref)
				*code = append(*code, fmt.Sprintf("\tv.%s = %s()\n", k, funcName))
			} else {
				funcName, err = adapter.adapt(types, typeId1, k, typeId2, v.group, ref)
				if err != nil {
					return err
				}
				*code = append(*code, fmt.Sprintf("\tv.%s = %s()\n", k, funcName))
			}
		case itemKind.String, itemKind.Number:
			*code = append(*code, fmt.Sprintf("\tv.%s = %s\n", k, v.original))
		}
	}
	if s.next != nil {
		return s.next.execute(it, types, imp, code, adapter)
	} else {
		return nil
	}
}

func (s *structEndGen) execute(it item, types []typeInfo, imp imports, code *[]string, adapter *adapter) error {
	*code = append(*code, "\treturn v\n")
	*code = append(*code, "}\n\n")
	if s.next != nil {
		return s.next.execute(it, types, imp, code, adapter)
	} else {
		return nil
	}
}
