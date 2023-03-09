package sgo

import (
	"fmt"
	"reflect"
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

func (s *structInitGen) genFunc(f *item) (string, error) {
	code := f.name + "("
	for i, n := range f.deps {
		d := n.item
		// process all parameters IN PROGRESS
		parameter := ""
		switch d.kind {
		case itemKind.Func:
			parameter = d.name
		case itemKind.Struct:
			funcName := getFuncName(d, len(d.path) > 0 && d.path[0] == '*')
			parameter = funcName + "()"
		case itemKind.String, itemKind.Number, itemKind.Boolean:
			parameter = d.original
		default:
			return "", fmt.Errorf(TypeDoesNotSupportedF, d.original)
		}

		if i == 0 {
			code = code + parameter
		} else {
			code = fmt.Sprintf("%s, %s", code, parameter)
		}
	}
	return code + ")", nil
}

func (s *structInitGen) execute(it item, types []typeInfo, imp imports, code *[]string, adapter *adapter) error {
	var err error
	var field *field
	for _, v := range it.deps {
		switch v.item.kind {
		case itemKind.Func:
			alias := string(appendImport(imp, v.item.path+v.item.pkg))
			if alias != "" {
				alias += "."
			}
			if v.name == "." {
				// execute the method
				f, e := s.genFunc(v.item)
				if e != nil {
					return e
				}
				*code = append(*code, fmt.Sprintf("\tv.%s\n", f))
			} else {
				if it.group == "" {
					field, err = getFieldInfo(types, it.original, v.name)
				} else {
					field, err = getFieldInfo(types, it.original[len(it.group)+2:], v.name)
				}
				if err != nil {
					return err
				}
				switch field.Kind {
				case reflect.Func:
					if v.item.exec {
						f, e := s.genFunc(v.item)
						if e != nil {
							return e
						}
						*code = append(*code, fmt.Sprintf("\tv.%s = %s%s\n", v.name, alias, f))
					} else {
						// it is a reference to a func then just return it as is
						*code = append(*code, fmt.Sprintf("\tv.%s = %s%s\n", v.name, alias, v.item.name))
					}
				case reflect.Struct, reflect.Interface:
					// if it is a reference to a struct then perform the function
					f, e := s.genFunc(v.item)
					if e != nil {
						return e
					}
					*code = append(*code, fmt.Sprintf("\tv.%s = %s%s\n", v.name, alias, f))
				default:
					return fmt.Errorf(TypeDoesNotSupportedF, v.item.original)
				}
			}
		case itemKind.Struct:
			typeId1 := it.original
			if typeId1[0] == '*' {
				typeId1 = typeId1[1:]
			}
			typeId2 := v.item.original
			if v.item.group != "" {
				typeId2 = typeId2[len(v.item.group)+2:]
			}
			if typeId2[0] == '*' {
				typeId2 = typeId2[1:]
			}
			supported, err := adapter.areTypesCompatible(types, typeId1, v.name, typeId2)
			if err != nil {
				return err
			}
			ref := len(v.item.path) > 0 && v.item.path[0] == '*'
			funcName := ""
			if supported {
				funcName = getFuncName(v.item, ref)
				*code = append(*code, fmt.Sprintf("\tv.%s = %s()\n", v.name, funcName))
			} else {
				funcName, err = adapter.adapt(types, typeId1, v.name, typeId2, v.item.group, ref)
				if err != nil {
					return err
				}
				*code = append(*code, fmt.Sprintf("\tv.%s = %s()\n", v.name, funcName))
			}
		case itemKind.String, itemKind.Number, itemKind.Boolean:
			*code = append(*code, fmt.Sprintf("\tv.%s = %s\n", v.name, v.item.original))
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
