// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

package sgo

import (
	"fmt"
	"os"
	"path/filepath"
)

func (b *Builder) Build(application string) error {
	b.Logger.Info(fmt.Sprintf("building \"%s\" application", application))
	if err := checkApplication(application); err != nil {
		return err
	}
	// check generated files exists
	wd, _ := os.Getwd()
	folderPath := filepath.Join(wd, application)
	filePath := filepath.Join(folderPath, depsFileName)
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(BuilderFileDoesNotExistF, filePath)
		} else {
			return err
		}
	}
	filePath = filepath.Join(folderPath, appFileName)
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(BuilderFileDoesNotExistF, filePath)
		} else {
			return err
		}
	}
	// build the application
	filePath = filepath.Join(folderPath, application+".exe")
	os.Chdir(folderPath)
	defer os.Chdir(wd)

	if _, err := goMod(folderPath, application); err != nil {
		return err
	}
	return goBuild(folderPath, filePath)
}

func (b *Builder) Clean(application string) error {
	b.Logger.Info(fmt.Sprintf("cleaning \"%s\" application", application))
	if err := checkApplication(application); err != nil {
		return err
	}
	// check the Go file with all dependencies is exist
	wd, _ := os.Getwd()
	folderPath := filepath.Join(wd, application)
	if _, err := os.Stat(folderPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	os.Chdir(folderPath)
	defer os.Chdir(wd)
	return goClean()
}

func (b *Builder) SetLogger(logger Logger) {
	b.Logger = logger
}
