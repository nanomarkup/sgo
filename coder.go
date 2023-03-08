// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package sgo

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func (g *Coder) Init(items map[string][][]string) {
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
	// create a hidden folder as wd
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	wd = filepath.Join(wd, application, workingFolderName)
	if _, err = os.Stat(wd); err == nil {
		os.RemoveAll(wd)
	}
	if err = os.MkdirAll(wd, os.ModePerm); err != nil {
		return err
	}
	pWd, err := syscall.UTF16PtrFromString(wd)
	if err != nil {
		return err
	}
	if err = syscall.SetFileAttributes(pWd, syscall.FILE_ATTRIBUTE_HIDDEN); err != nil {
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
		for _, app := range apps {
			if app[0] == application {
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
				break
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
	found := false
	for _, app := range apps {
		if app[0] == application {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf(AppIsMissingF, application)
	}
	// read app details
	info, err := readItem(application, g.items)
	if err != nil {
		return "", err
	}
	// get entry point
	entry := ""
	found = false
	for _, i := range info {
		if i[0] == entryAttrName {
			if len(i) < 2 {
				return "", fmt.Errorf(AppAttrIsEmptyF, entryAttrName, application)
			}
			found = true
			entry = i[1]
			break
		}
	}
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
			funcName := fmt.Sprintf("\tapp := %s()\n", getFuncName(&entry, false))
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
	code2 := []string{}
	imports := imports{}
	adapter := adapter{}
	adapter.imports = imports
	// get all type of struct items to process
	its := map[string]bool{}
	g.getStructItems(entryPoint, list, its)
	// generate code for all type of struct items
	var err error
	gen := generator{}
	gen.structGenerator = &structBegGen{
		next: &structCreateGen{
			next: &structInitGen{
				next: &structEndGen{},
			},
		},
	}
	for i := range its {
		if it, found := list[i]; found {
			switch it.kind {
			case itemKind.Func:
				appendImport(imports, it.path+it.pkg)
			case itemKind.Struct:
				code2, err = gen.createStruct(it, types, imports, &adapter)
				if err != nil {
					return nil, nil, err
				}
				code = append(code, code2...)
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
			switch v.item.kind {
			case itemKind.Func:
				for _, d := range v.item.deps {
					if d.item.kind == itemKind.Struct {
						g.getStructItems(d.item.original, list, result)
					}
				}
			case itemKind.Struct:
				g.getStructItems(v.item.original, list, result)
			}
		}
	}
}
