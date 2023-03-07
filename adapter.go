// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package sgo

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (o *adapter) adapt(types []typeInfo, typeA string, fieldA string, typeB string, groupB string, ref bool) (string, error) {
	infoA := getType(types, typeA)
	if infoA == nil {
		return "", fmt.Errorf(TypeIsMissingF, typeA)
	}
	infoB := getType(types, typeB)
	if infoB == nil {
		return "", fmt.Errorf(TypeIsMissingF, typeB)
	}
	fieldId := ""
	for _, v := range infoA.Fields {
		if v.FieldName == fieldA {
			fieldId = v.Id
			break
		}
	}
	if fieldId == "" {
		return "", fmt.Errorf(FieldIsMissingF, fieldA, typeA)
	}
	fieldInfo := getType(types, fieldId)
	if fieldInfo == nil {
		return "", fmt.Errorf(TypeIsMissingF, fieldId)
	}
	// create a new struct
	name := ""
	nameA := fmt.Sprintf("%s%s", cases.Title(language.English, cases.NoLower).String(filepath.Base(fieldInfo.PkgPath)), fieldInfo.Name)
	nameB := fmt.Sprintf("%s%s", cases.Title(language.English, cases.NoLower).String(filepath.Base(infoB.PkgPath)), infoB.Name)
	if groupB == "" {
		name = fmt.Sprintf("%s%s%s", nameB, nameA, GenAdapterSufix)
	} else {
		name = fmt.Sprintf("%s%s%s%s%s", groupB, GenGroupPrefix, nameB, nameA, GenAdapterSufix)
	}
	funcName := GenNamePrefix + name
	if ref {
		funcName = funcName + GenRefSufix
	}
	// if the adapter exists then return it
	if o.code != nil && o.code[name] != nil {
		return funcName, nil
	}
	alias := string(appendImport(o.imports, infoB.PkgPath))
	code := []string{}
	code = append(code, fmt.Sprintf("type %s struct {\n", name))
	code = append(code, fmt.Sprintf("\t%s.%s\n}\n\n", alias, infoB.Name))
	// check methods
	var fA field
	var fB field
	var iA int
	var iB int
	var iP int
	var countA int
	var countB int
	var field string
	var fields []string
	var incompatible bool
	for _, v := range fieldInfo.Methods {
		found := false
		for _, x := range infoB.Methods {
			if x.Name == v.Name {
				// check input parameters
				iA = 0
				iB = 0
				found = true
				countA = len(v.In)
				countB = len(x.In)
				if countA > 0 && v.In[0].Id == "." && v.In[0].Kind == reflect.Ptr {
					iA++
				}
				if countB > 0 && x.In[0].Id == "." && x.In[0].Kind == reflect.Ptr {
					iB++
				}
				if (countA - iA) != (countB - iB) {
					return "", fmt.Errorf(WrongNumberOfInputParamsF, fieldA, typeA, typeB)
				}
				incompatible = false
				for i := iA; i < countA; i++ {
					fA = v.In[i]
					fB = x.In[iB]
					if fA.Kind != fB.Kind || fA.Id != fB.Id {
						incompatible = true
						break
					}
					iB++
				}
				// check output parameters
				if len(x.Out) != len(v.Out) {
					return "", fmt.Errorf(WrongNumberOfOutputParamsF, fieldA, typeA, typeB)
				}
				for i, p := range x.Out {
					fA = v.Out[i]
					if p.Kind != fA.Kind || p.Id != fA.Id {
						incompatible = true
						break
					}
				}
				if incompatible {
					// resolve incompatible method
					code = append(code, fmt.Sprintf("func (o *%s) %s(", name, v.Name))
					iA = 0
					iP = 1
					countA = len(v.In)
					if countA > 0 && v.In[0].Id == "." && v.In[0].Kind == reflect.Ptr {
						iA++
					}
					fields = []string{}
					for i := iA; i < countA; i++ {
						fA = v.In[i]
						alias = string(appendImport(o.imports, fA.PkgPath))
						field = fmt.Sprintf("%s.%s", alias, fA.TypeName)
						if field[0:1] == "." {
							field = field[1:]
						}
						fields = append(fields, fmt.Sprintf("a%d %s", iP, field))
						iP++
					}
					iP = 1
					code = append(code, strings.Join(fields, ", ")+")")
					fields = []string{}
					for i := range v.Out {
						fA = v.Out[i]
						alias = string(appendImport(o.imports, fA.PkgPath))
						field = fmt.Sprintf("%s.%s", alias, fA.TypeName)
						if field[0:1] == "." {
							field = field[1:]
						}
						fields = append(fields, fmt.Sprintf("r%d %s", iP, field))
						iP++
					}
					if len(fields) > 0 {
						code = append(code, fmt.Sprintf(" (%s)", strings.Join(fields, ", ")))
					}
					code = append(code, " {\n")
					rcode, err := o.resolveMethod(infoB.Name, x, v)
					if err != nil {
						return "", err
					}
					code = append(code, rcode...)
					code = append(code, "}\n\n")
				}
				break
			}
		}
		if !found {
			return "", fmt.Errorf(MethodIsMissingF, v.Name, infoB.Id)
		}
	}
	// generate the "Use" function
	if ref {
		code = append(code, fmt.Sprintf("func %s() *%s {\n", funcName, name))
		code = append(code, fmt.Sprintf("\tv := &%s{}\n", name))
	} else {
		code = append(code, fmt.Sprintf("func %s() %s {\n", funcName, name))
		code = append(code, fmt.Sprintf("\tv := %s{}\n", name))
	}
	if groupB == "" {
		code = append(code, fmt.Sprintf("\tv.%s = *%s%s%s()\n", infoB.Name, GenNamePrefix, nameB, GenRefSufix))
	} else {
		code = append(code, fmt.Sprintf("\tv.%s = *%s%s%s%s%s()\n", infoB.Name, GenNamePrefix, groupB, GenGroupPrefix, nameB, GenRefSufix))
	}
	code = append(code, "\treturn v\n")
	code = append(code, "}\n\n")
	// keep a new code
	if o.code == nil {
		o.code = map[string][]string{}
	}
	o.code[name] = append(o.code[name], code...)
	return funcName, nil
}

func (o *adapter) areTypesCompatible(types []typeInfo, typeA string, fieldA string, typeB string) (bool, error) {
	// get input types
	infoA := getType(types, typeA)
	if infoA == nil {
		return false, fmt.Errorf(TypeIsMissingF, typeA)
	}
	infoB := getType(types, typeB)
	if infoB == nil {
		return false, fmt.Errorf(TypeIsMissingF, typeB)
	}
	fieldId := ""
	var fieldOrigInfo field
	for _, v := range infoA.Fields {
		if v.FieldName == fieldA {
			fieldId = v.Id
			fieldOrigInfo = v
			break
		}
	}
	if fieldId == "" {
		return false, fmt.Errorf(FieldIsMissingF, fieldA, typeA)
	}
	fieldInfo := getType(types, fieldId)
	if fieldInfo == nil {
		if fieldOrigInfo.Id == "." && fieldOrigInfo.PkgPath == "" && fieldOrigInfo.TypeName == "" &&
			(fieldOrigInfo.Kind == reflect.Interface || fieldOrigInfo.Kind == reflect.Pointer) {
			// it is type of interface{}
			return true, nil
		} else {
			return false, fmt.Errorf(TypeIsMissingFieldIdF, fieldId)
		}
	}
	// check compatibility of input types
	if fieldInfo.Id == infoB.Id {
		return true, nil
	}
	if fieldInfo.Kind != reflect.Interface {
		return false, fmt.Errorf(TypeIsNotInterface, fieldInfo.Id)
	}
	// check methods
	var fA field
	var fB field
	var iA int
	var iB int
	var countA int
	var countB int
	for _, v := range fieldInfo.Methods {
		found := false
		for _, x := range infoB.Methods {
			if x.Name == v.Name {
				// check input parameters
				iA = 0
				iB = 0
				countA = len(v.In)
				countB = len(x.In)
				if countA > 0 && v.In[0].Id == "." && v.In[0].Kind == reflect.Ptr {
					iA++
				}
				if countB > 0 && x.In[0].Id == "." && x.In[0].Kind == reflect.Ptr {
					iB++
				}
				if (countA - iA) != (countB - iB) {
					return false, fmt.Errorf(WrongNumberOfInputParamsF, fieldA, typeA, typeB)
				}
				for i := iA; i < countA; i++ {
					fA = v.In[i]
					fB = x.In[iB]
					if fA.Kind != fB.Kind || fA.Id != fB.Id {
						return false, nil
					}
					iB++
				}
				// check output parameters
				if len(x.Out) != len(v.Out) {
					return false, fmt.Errorf(WrongNumberOfOutputParamsF, fieldA, typeA, typeB)
				}
				for i, p := range x.Out {
					fA = v.Out[i]
					if p.Kind != fA.Kind || p.Id != fA.Id {
						return false, nil
					}
				}
				found = true
				break
			}
		}
		if !found {
			return false, fmt.Errorf(MethodIsMissingF, v.Name, infoB.Id)
		}
	}
	return true, nil
}

func (o *adapter) resolveMethod(obj string, m1 method, m2 method) ([]string, error) {
	iA := 0
	iB := 0
	countA := len(m1.In)
	countB := len(m2.In)
	inCode := []string{}
	outCode := []string{}
	inputs := []string{}
	outputs := []string{}
	var fA field
	var fB field
	var name string
	if countA > 0 && m1.In[0].Id == "." && m1.In[0].Kind == reflect.Ptr {
		iA++
	}
	if countB > 0 && m2.In[0].Id == "." && m2.In[0].Kind == reflect.Ptr {
		iB++
	}
	if (countA - iA) != (countB - iB) {
		return nil, fmt.Errorf(WrongNumberOfInputParamsForMethodsF, m1.Name, m2.Name)
	}
	// process input parameters
	iP := 1
	for i := iA; i < countA; i++ {
		fA = m1.In[i]
		fB = m2.In[iB]
		name = "a" + strconv.Itoa(iP)
		name, code, err := o.resolveParameter(true, name, fB, "b"+strconv.Itoa(iP), fA)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, name)
		inCode = append(inCode, code...)
		iB++
		iP++
	}
	method := fmt.Sprintf("o.%s.%s(%s)", obj, m1.Name, strings.Join(inputs, ", "))
	// process output parameters
	if len(m2.Out) != len(m1.Out) {
		return nil, fmt.Errorf(WrongNumberOfOutputParamsForMethodsF, m1.Name, m2.Name)
	}
	if len(m1.Out) == 0 && len(m2.Out) == 0 {
		inCode = append(inCode, fmt.Sprintf("\t%s\n", method))
	} else if o.equals(m1.Out, m2.Out) {
		inCode = append(inCode, fmt.Sprintf("\treturn %s\n", method))
	} else {
		iP = 1
		for i, p := range m2.Out {
			fA = m1.Out[i]
			name = "v" + strconv.Itoa(iP)
			name, code, err := o.resolveParameter(false, name, fA, "r"+strconv.Itoa(iP), p)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, name)
			outCode = append(outCode, code...)
			iP++
		}
		method = fmt.Sprintf("\t%s := %s\n", strings.Join(outputs, ", "), method)
		inCode = append(inCode, method)
		inCode = append(inCode, outCode...)
		inCode = append(inCode, "\treturn\n")
	}
	return inCode, nil
}

func (o *adapter) equals(l1 []field, l2 []field) bool {
	var f2 field
	for i, f1 := range l1 {
		f2 = l2[i]
		if f1.Kind != f2.Kind || f1.Id != f2.Id {
			return false
		}
	}
	return true
}

func (o *adapter) resolveParameter(in bool, name1 string, f1 field, name2 string, f2 field) (string, []string, error) {
	if f1.Kind == f2.Kind && f1.Id == f2.Id {
		if in {
			return name1, nil, nil
		} else {
			return name2, nil, nil
		}
	} else if f1.Kind == reflect.Interface && f2.Kind == reflect.Interface {
		alias := string(appendImport(o.imports, f2.PkgPath))
		if in {
			return name2, []string{fmt.Sprintf("\t%s := %s.(%s.%s)\n", name2, name1, alias, f2.FieldName)}, nil
		} else {
			return name1, []string{fmt.Sprintf("\t%s = %s.(%s.%s)\n", name2, name1, alias, f2.FieldName)}, nil
		}
	}
	return "", nil, fmt.Errorf(ParamsDoesNotSupportedF, f1.TypeName, f2.TypeName)
}
