// Copyright 2022 Vitalii Noha vitalii.noga@gmail.com. All rights reserved.

// Package app generates Go sources of an application.
package sgo

type Coder struct {
	Logger Logger
	items  map[string][][]string
}

type Builder struct {
	Logger Logger
}

type Logger interface {
	Trace(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	IsTrace() bool
	IsDebug() bool
	IsInfo() bool
	IsWarn() bool
	IsError() bool
}

const (
	// application
	GenNamePrefix   string = "Use"
	GenGroupPrefix  string = "Group"
	GenRefSufix     string = "Ref"
	GenAdapterSufix string = "Adapter"
	// notifications
	// errors
	AppIsMissingF                        string = "the selected \"%s\" application does not found"
	AppIsNotSpecified                    string = "the application is not specified"
	AppAttrIsEmptyF                      string = "the \"%s\" attribute is empty for the \"%s\" application"
	AppAttrIsMissingF                    string = "the \"%s\" attribute is not exist for the \"%s\" application"
	TypeIsMissingF                       string = "\"%s\" type does not found"
	TypeIsMissingFieldIdF                string = "\"%s\" type does not found (field Id)"
	TypeIsNotInterface                   string = "the receiver of \"%s\" type should be type of interface"
	TypeDoesNotSupportedF                string = "\"%s\" type of parameter does not supported"
	ItemIsMissingF                       string = "the %s item is not found"
	ItemIsIncorrect                      string = "cannot detect type of \"%s\" item"
	FieldIsMissingF                      string = "\"%s\" field of \"%s\" type does not exist"
	MethodIsMissingF                     string = "the \"%s\" method is missing in \"%s\""
	ParamsDoesNotSupportedF              string = "cannot resolve \"%s\" and \"%s\" parameters"
	BuilderFileDoesNotExistF             string = "\"%s\" does not exist. Please use a \"code\" command to create it"
	WrongNumberOfInputParamsF            string = "the number of input parameters are different for \"%s\" method of \"%s\" type and \"%s\" type"
	WrongNumberOfInputParamsForMethodsF  string = "the number of input parameters are different for \"%s\" and \"%s\" methods"
	WrongNumberOfOutputParamsF           string = "the number of output parameters are different for \"%s\" method of \"%s\" type and \"%s\" type"
	WrongNumberOfOutputParamsForMethodsF string = "the number of output parameters are different for \"%s\" and \"%s\" methods"
	ErrorOnGettingTypeDetails            string = "cannot collect type details"
	FuncBegTokenIsMissing                string = "incorrect syntax, the \"(\" is missing"
	FuncEndTokenIsMissing                string = "incorrect syntax, the \")\" is missing"
	GroupEndTokenIsMissing               string = "cannot get a group name, the \"]\" is missing"
)
