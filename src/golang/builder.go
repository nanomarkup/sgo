// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package golang

import (
	"fmt"
	"os"
	"path/filepath"
)

func (b *Builder) Init(items map[string]map[string]string) {
	b.items = items
}

func (b *Builder) Build(application string) error {
	b.Logger.Info(fmt.Sprintf("building \"%s\" application", application))
	if err := checkApplication(application); err != nil {
		return err
	}
	// check the golang file with all dependencies is exist
	wd, _ := os.Getwd()
	folderPath := filepath.Join(wd, application)
	filePath := filepath.Join(folderPath, depsFileName)
	if _, err := os.Stat(filePath); err != nil {
		return fmt.Errorf("\"%s\" does not exist. Please use a \"generate\" command to create it.", filePath)
	}
	g := Coder{
		b.Logger,
		b.items,
	}
	// generate a golang app file if it is missing
	filePath = filepath.Join(folderPath, appFileName)
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			if err := g.generateAppFile(application); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	// build the application
	filePath = filepath.Join(folderPath, application+".exe")
	return goBuild(folderPath, filePath)
}

func (b *Builder) Clean(application string) error {
	b.Logger.Info(fmt.Sprintf("cleaning \"%s\" application", application))
	if err := checkApplication(application); err != nil {
		return err
	}
	// check the golang file with all dependencies is exist
	wd, _ := os.Getwd()
	folderPath := filepath.Join(wd, application)
	if _, err := os.Stat(folderPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return goClean(folderPath)
}

func (b *Builder) SetLogger(logger Logger) {
	b.Logger = logger
}
