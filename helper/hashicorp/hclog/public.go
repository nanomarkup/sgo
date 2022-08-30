// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

// Package helper provides methods for using "github.com/hashicorp/go-hclog" sources.
package helper

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
)

var LogLever = struct {
	NoLevel uint
	Trace   uint
	Debug   uint
	Info    uint
	Warn    uint
	Error   uint
	Off     uint
}{
	0,
	1,
	2,
	3,
	4,
	5,
	6,
}

func NewStdOut(name string, level uint) hclog.Logger {
	return hclog.New(&hclog.LoggerOptions{
		Name:   name,
		Level:  hclog.Level(level),
		Output: os.Stdout,
	})
}

func NewFileOut(name string, level uint) hclog.Logger {
	f, err := os.Create(fmt.Sprintf("%s.log", name))
	if err != nil {
		panic(err)
	}
	return hclog.New(&hclog.LoggerOptions{
		Name:   name,
		Level:  hclog.Level(level),
		Output: f,
	})
}
