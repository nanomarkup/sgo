// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package app

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (g *Coder) Init(items map[string]map[string]string) {
	g.items = items
}

func (g *Coder) Generate(application string) error {
	g.Logger.Info(fmt.Sprintf("generating \"%s\" application", application))
	if err := checkApplication(application); err != nil {
		return err
	}
	entry, err := g.entryPoint(application)
	if err != nil {
		return err
	}
	// create a temporary folder as wd
	wd, err := ioutil.TempDir("", "sc*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(wd)
	// generate a file with all dependencies
	err = g.generateDepsFile(application, entry, wd)
	if err != nil {
		return err
	}
	// generate an app file if it is missing
	pd, _ := os.Getwd()
	filePath := filepath.Join(pd, application, appFileName)
	if _, err := os.Stat(filePath); err != nil && os.IsNotExist(err) {
		if err := g.generateAppFile(application); err != nil {
			return err
		}
	}
	return err
}

func (g *Coder) Clean(application string) error {
	g.Logger.Info(fmt.Sprintf("cleaning \"%s\" application", application))
	if err := checkApplication(application); err != nil {
		return err
	}
	// get current application if it is missing
	if application == "" {
		return fmt.Errorf(AppIsNotSpecified)
	}
	if apps, err := readItem(appsItemName, g.items); err == nil {
		if _, found := apps[application]; found {
			if dir, err := os.Getwd(); err == nil {
				folderPath := filepath.Join(dir, application)
				// remove the apps file
				filePath := filepath.Join(folderPath, appFileName)
				if _, err := os.Stat(filePath); err == nil {
					os.Remove(filePath)
				}
				// remove the deps file
				filePath = filepath.Join(folderPath, depsFileName)
				if _, err := os.Stat(filePath); err == nil {
					os.Remove(filePath)
				}
				// remove the application folder if it is empty
				if empty, _ := isDirEmpty(folderPath); empty {
					os.Remove(folderPath)
				}
			}
		}
	}
	return nil
}

func (g *Coder) SetLogger(logger Logger) {
	g.Logger = logger
}

func (g *Coder) entryPoint(application string) (string, error) {
	// read the apps item
	apps, err := readItem(appsItemName, g.items)
	if err != nil {
		return "", err
	}
	// check the applicatin is exist
	if _, found := apps[application]; !found {
		return "", fmt.Errorf(AppIsMissingF, application)
	}
	// read app details
	info, err := readItem(application, g.items)
	if err != nil {
		return "", err
	}
	// get entry point
	entry, found := info[entryAttrName]
	if !found {
		return "", fmt.Errorf(AppAttrIsMissingF, entryAttrName, application)
	}
	return entry, nil
}

func (g *Coder) generateAppFile(application string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	filePath := filepath.Join(wd, application, appFileName)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()
	writer.WriteString("package main\n\n")
	writer.WriteString(fmt.Sprintf("const AppName = \"%s\"\n\n", application))
	writer.WriteString("func main() {\n")
	writer.WriteString("\tExecute()\n")
	writer.WriteString("}\n")
	return nil
}

func (g *Coder) generateDepsFile(application, entryPoint, wd string) error {
	// check and get info about all dependencies
	r := resolver{
		application,
		entryPoint,
		g.items,
	}
	list, types, err := r.resolve(wd)
	if err != nil {
		return err
	}
	code, imports, err := g.generateItems(entryPoint, list, types)
	if err != nil {
		return err
	}
	entry, found := list[entryPoint]
	if found && entry.kind == itemKind.String {
		imports["fmt"] = ""
	}
	// save dependencies to a file
	pd, _ := os.Getwd()
	root := filepath.Join(pd, application)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		os.Mkdir(root, os.ModePerm)
	}
	file, err := os.Create(filepath.Join(root, depsFileName))
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()
	writer.WriteString("package main\n\n")
	// write the import section
	if len(imports) > 0 {
		writer.WriteString("import (\n")
		for path, alias := range imports {
			if alias == "" {
				writer.WriteString(fmt.Sprintf("\t\"%s\"\n", path))
			} else {
				writer.WriteString(fmt.Sprintf("\t%s \"%s\"\n", alias, path))
			}
		}
		writer.WriteString(")\n\n")
	}
	// write entry point
	writer.WriteString("func Execute() {\n")
	if found {
		switch entry.kind {
		case itemKind.Func:
			writer.WriteString(fmt.Sprintf("\t%s.%s\n", entry.pkg, entry.name))
		case itemKind.Struct:
			funcName := fmt.Sprintf("\tapp := %s%s%s()\n", GenNamePrefix, cases.Title(language.English, cases.NoLower).String(entry.pkg), entry.name)
			funcName = strings.ReplaceAll(funcName, "-", "_")
			writer.WriteString(funcName)
			writer.WriteString("\tapp.Execute()\n")
		case itemKind.String:
			writer.WriteString(fmt.Sprintf("\tfmt.Println(%s)\n", entry.original))
		}
	}
	writer.WriteString("}\n\n")
	// write items
	if len(code) > 0 {
		for _, v := range code {
			writer.WriteString(v)
		}
	}
	return nil
}

func (g *Coder) generateItems(entryPoint string, list items, types []typeInfo) ([]string, imports, error) {
	code := []string{}
	imports := imports{}
	adapter := adapter{}
	adapter.imports = &imports
	// get all type of struct items to process
	its := map[string]bool{}
	g.getStructItems(entryPoint, list, its)
	// generate code for all type of struct items
	ref := false
	name := ""
	alias := ""
	typeId1 := ""
	typeId2 := ""
	funcName := ""
	parameter := ""
	fullNameDefine := ""
	fullNameReturn := ""
	var err error
	var field *field
	for i := range its {
		if it, found := list[i]; found {
			switch it.kind {
			case itemKind.Func:
				appendImport(imports, it.path+it.pkg)
			case itemKind.Struct:
				funcName = fmt.Sprintf("%s%s%s", GenNamePrefix, cases.Title(language.English, cases.NoLower).String(it.pkg), it.name)
				funcName = strings.ReplaceAll(funcName, "-", "_")
				fullNameDefine = it.name
				fullNameReturn = it.name
				if len(it.path) > 0 {
					if it.path[0] == '*' {
						alias = string(appendImport(imports, it.path[1:]+it.pkg))
						funcName = funcName + GenRefSufix
						fullNameDefine = fmt.Sprintf("*%s.%s", alias, it.name)
						fullNameReturn = fmt.Sprintf("&%s.%s", alias, it.name)
					} else {
						alias = string(appendImport(imports, it.path+it.pkg))
						fullNameDefine = fmt.Sprintf("%s.%s", alias, it.name)
						fullNameReturn = fullNameDefine
					}
				}
				// create a new item and initialize it
				code = append(code, fmt.Sprintf("func %s() %s {\n", funcName, fullNameDefine))
				if len(it.deps) == 0 {
					code = append(code, fmt.Sprintf("\treturn %s{}\n", fullNameReturn))
				} else {
					code = append(code, fmt.Sprintf("\tv := %s{}\n", fullNameReturn))
					for k, v := range it.deps {
						switch v.kind {
						case itemKind.Func:
							alias = string(appendImport(imports, v.path+v.pkg))
							field, err = g.getFieldInfo(types, it.original, k)
							if err != nil {
								return nil, nil, err
							}
							switch field.Kind {
							case reflect.Func:
								// if it is a reference to a func then just return it as is
								code = append(code, fmt.Sprintf("\tv.%s = %s.%s\n", k, alias, v.name))
							case reflect.Struct, reflect.Interface:
								// if it is a reference to a struct then perform it
								name = v.name + "("
								if len(v.deps) > 0 {
									keys := reflect.ValueOf(v.deps).MapKeys()
									keysOrder := func(i, j int) bool { return keys[i].String() < keys[j].String() }
									sort.Slice(keys, keysOrder)
									for i, n := range keys {
										d := v.deps[n.String()]
										// process all parameters IN PROGRESS
										parameter = ""
										switch d.kind {
										case itemKind.Func:
											parameter = d.name
										case itemKind.Struct:
											funcName = fmt.Sprintf("%s%s%s", GenNamePrefix, cases.Title(language.English, cases.NoLower).String(d.pkg), d.name)
											funcName = strings.ReplaceAll(funcName, "-", "_")
											if len(d.path) > 0 && d.path[0] == '*' {
												funcName = funcName + GenRefSufix
											}
											parameter = funcName + "()"
										case itemKind.String, itemKind.Number:
											parameter = d.original
										default:
											g.Logger.Error(fmt.Sprintf("\"%s\" type of parameter does not supported:", d.original))
											g.Logger.Error(fmt.Sprintf("\tkind=%d", d.kind))
											return nil, nil, fmt.Errorf(TypeDoesNotSupportedF, d.original)
										}

										if i == 0 {
											name = name + parameter
										} else {
											name = fmt.Sprintf("%s, %s", name, parameter)
										}
									}
								}
								name = name + ")"
								code = append(code, fmt.Sprintf("\tv.%s = %s.%s\n", k, alias, name))
							default:
								return nil, nil, fmt.Errorf(TypeDoesNotSupportedF, v.original)
							}
						case itemKind.Struct:
							typeId1 = it.original
							if typeId1[0] == '*' {
								typeId1 = typeId1[1:]
							}
							typeId2 = v.original
							if typeId2[0] == '*' {
								typeId2 = typeId2[1:]
							}
							supported, err := g.areTypesCompatible(types, typeId1, k, typeId2)
							if err != nil {
								return nil, nil, err
							}
							ref = len(v.path) > 0 && v.path[0] == '*'
							if supported {
								funcName = fmt.Sprintf("%s%s%s", GenNamePrefix, cases.Title(language.English, cases.NoLower).String(v.pkg), v.name)
								funcName = strings.ReplaceAll(funcName, "-", "_")
								if ref {
									funcName = funcName + GenRefSufix
								}
								code = append(code, fmt.Sprintf("\tv.%s = %s()\n", k, funcName))
							} else {
								funcName, err = adapter.adapt(types, typeId1, k, typeId2, ref)
								if err != nil {
									return nil, nil, err
								}
								code = append(code, fmt.Sprintf("\tv.%s = %s()\n", k, funcName))
							}
						case itemKind.String, itemKind.Number:
							code = append(code, fmt.Sprintf("\tv.%s = %s\n", k, v.original))
						}
					}
					code = append(code, "\treturn v\n")
				}
				code = append(code, "}\n\n")
			}
		}
	}
	// append adapters
	for _, value := range adapter.code {
		code = append(code, value...)
	}
	return code, imports, nil
}

func (g *Coder) getStructItems(original string, list items, result map[string]bool) {
	if result[original] {
		return
	}
	if it, found := list[original]; found {
		if it.kind == itemKind.Struct {
			result[original] = true
		}
		for _, v := range it.deps {
			switch v.kind {
			case itemKind.Func:
				for _, d := range v.deps {
					if d.kind == itemKind.Struct {
						g.getStructItems(d.original, list, result)
					}
				}
			case itemKind.Struct:
				g.getStructItems(v.original, list, result)
			}
		}
	}
}

func (g *Coder) getFieldInfo(types []typeInfo, item string, field string) (*field, error) {
	item = strings.TrimPrefix(item, "*")
	info := getType(types, item)
	if info == nil {
		return nil, fmt.Errorf(TypeIsMissingF, item)
	}
	for _, v := range info.Fields {
		if v.FieldName == field {
			return &v, nil
		}
	}
	return nil, fmt.Errorf(FieldIsMissingF, field, item)
}

func (g *Coder) areTypesCompatible(types []typeInfo, typeA string, fieldA string, typeB string) (bool, error) {
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
		if fieldOrigInfo.Id == "." && fieldOrigInfo.Kind == reflect.Interface && fieldOrigInfo.PkgPath == "" && fieldOrigInfo.TypeName == "" {
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
