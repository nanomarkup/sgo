// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

// Package plugins implements common objects for supporting plugins.
package plugins

type Builder interface {
	Build(app string) error
	Clean(app string, sources *map[string]map[string]string) error
	Generate(app string, sources *map[string]map[string]string) error
}

type BuilderPlugin struct {
	Impl Builder
}
